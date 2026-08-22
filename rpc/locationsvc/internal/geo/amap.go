package geo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"XiaoLong-Ridy/rpc/locationsvc/internal/config"
	"XiaoLong-Ridy/rpc/locationsvc/internal/model"
)

const (
	defaultRetryCount = 3
	defaultCacheTTL   = 5 * time.Minute
)

type cacheItem struct {
	data      []byte
	expiresAt time.Time
}

// Client 高德地图客户端
type Client struct {
	apiKey  string
	baseURL string
	httpCli *http.Client
	mu      sync.RWMutex
	cache   map[string]cacheItem
}

// NewClient 创建高德地图客户端
func NewClient(c config.MapServiceConfig) *Client {
	return &Client{
		apiKey:  c.ApiKey,
		baseURL: c.BaseUrl,
		httpCli: &http.Client{Timeout: 5 * time.Second},
		cache:   make(map[string]cacheItem),
	}
}

// doWithRetry 对高德 HTTP 请求做指数退避重试，提高网络抖动场景下的成功率。
func (c *Client) doWithRetry(reqURL string) ([]byte, error) {
	var lastErr error
	for i := 0; i < defaultRetryCount; i++ {
		if i > 0 {
			time.Sleep(time.Duration(i) * time.Second)
		}
		httpResp, err := c.httpCli.Get(reqURL)
		if err != nil {
			lastErr = err
			continue
		}
		body, err := io.ReadAll(httpResp.Body)
		httpResp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		return body, nil
	}
	return nil, fmt.Errorf("请求高德API失败（已重试%d次）: %w", defaultRetryCount, lastErr)
}

func (c *Client) getCache(key string) []byte {
	c.mu.RLock()
	item, ok := c.cache[key]
	c.mu.RUnlock()
	if !ok {
		return nil
	}
	if time.Now().After(item.expiresAt) {
		c.mu.Lock()
		delete(c.cache, key)
		c.mu.Unlock()
		return nil
	}
	return item.data
}

func (c *Client) setCache(key string, data []byte, ttl time.Duration) {
	if ttl <= 0 {
		ttl = defaultCacheTTL
	}
	c.mu.Lock()
	c.cache[key] = cacheItem{data: data, expiresAt: time.Now().Add(ttl)}
	c.mu.Unlock()
}

// AmapPoiResponse 高德周边搜索接口响应结构
type AmapPoiResponse struct {
	Status string `json:"status"` // 1=成功
	Info   string `json:"info"`   // 提示信息
	Count  string `json:"count"`  // 结果总数（高德返回的是字符串）
	Pois   []struct {
		Name     string `json:"name"`
		Address  string `json:"address"`
		Location string `json:"location"` // "经度,纬度"
		Type     string `json:"type"`     // 分类，如"商务住宅;楼宇"
		Distance string `json:"distance"` // 距离（米）
	} `json:"pois"`
}

// SearchPoi 调用高德"周边搜索"（place/around）接口，结果做 5 分钟内存缓存。
func (c *Client) SearchPoi(keyword string, lat, lng float64, radius, page, size int32) (*AmapPoiResponse, error) {
	if radius <= 0 {
		radius = 5000
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}

	params := url.Values{}
	params.Set("key", c.apiKey)
	params.Set("keywords", keyword)
	params.Set("location", fmt.Sprintf("%.6f,%.6f", lng, lat))
	params.Set("radius", strconv.Itoa(int(radius)))
	params.Set("page", strconv.Itoa(int(page)))
	params.Set("offset", strconv.Itoa(int(size)))
	params.Set("extensions", "base")
	params.Set("output", "json")

	reqURL := fmt.Sprintf("%s/place/around?%s", c.baseURL, params.Encode())
	cacheKey := fmt.Sprintf("poi:%s:%.6f:%.6f:%d:%d:%d", keyword, lat, lng, radius, page, size)

	body := c.getCache(cacheKey)
	if body == nil {
		var err error
		body, err = c.doWithRetry(reqURL)
		if err != nil {
			return nil, err
		}
		c.setCache(cacheKey, body, defaultCacheTTL)
	}

	var result AmapPoiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析高德响应失败: %w", err)
	}
	if result.Status != "1" {
		return nil, fmt.Errorf("高德API返回错误: %s", result.Info)
	}
	return &result, nil
}

// ToPoiModels 把高德返回的结果转成本地 POI 模型
func (r *AmapPoiResponse) ToPoiModels() []model.Poi {
	var list []model.Poi
	for _, p := range r.Pois {
		lng, lat, err := parseLocation(p.Location)
		if err != nil {
			continue
		}
		distance, _ := strconv.Atoi(p.Distance)
		list = append(list, model.Poi{
			Name:      p.Name,
			Address:   p.Address,
			Longitude: lng,
			Latitude:  lat,
			Category:  p.Type,
			Distance:  distance,
			Source:    "amap",
		})
	}
	return list
}

// Total 返回结果总数（高德返回字符串，转成 int32）
func (r *AmapPoiResponse) Total() int32 {
	n, _ := strconv.Atoi(r.Count)
	return int32(n)
}

// AmapRegeoResponse 高德逆地理编码（regeo）响应结构
type AmapRegeoResponse struct {
	Status string `json:"status"`
	Info   string `json:"info"`
	Regeocode struct {
		FormattedAddress json.RawMessage `json:"formatted_address"`
		AddressComponent struct {
			Province string          `json:"province"`
			City     json.RawMessage `json:"city"`
			District string          `json:"district"`
			Township string          `json:"township"`
		} `json:"addressComponent"`
		Pois []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"pois"`
	} `json:"regeocode"`
}

// ReverseGeocode 调用高德"逆地理编码"接口，带重试 + 内存缓存降级。
func (c *Client) ReverseGeocode(lat, lng float64) (*AmapRegeoResponse, error) {
	params := url.Values{}
	params.Set("key", c.apiKey)
	params.Set("location", fmt.Sprintf("%.6f,%.6f", lng, lat))
	params.Set("extensions", "base")
	params.Set("output", "json")

	reqURL := fmt.Sprintf("%s/geocode/regeo?%s", c.baseURL, params.Encode())
	cacheKey := fmt.Sprintf("regeo:%.6f:%.6f", lat, lng)

	body := c.getCache(cacheKey)
	if body == nil {
		var err error
		body, err = c.doWithRetry(reqURL)
		if err != nil {
			return nil, err
		}
		c.setCache(cacheKey, body, defaultCacheTTL)
	}

	var result AmapRegeoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析高德逆地理编码响应失败: %w", err)
	}
	if result.Status != "1" {
		return nil, fmt.Errorf("高德API返回错误: %s", result.Info)
	}
	return &result, nil
}

// AddressStr 获取结构化地址字符串
func (r *AmapRegeoResponse) AddressStr() string {
	raw := r.Regeocode.FormattedAddress
	if len(raw) == 0 {
		return ""
	}
	var addr string
	if err := json.Unmarshal(raw, &addr); err != nil {
		return ""
	}
	return addr
}

// CityStr 兼容直辖市
func (r *AmapRegeoResponse) CityStr() string {
	raw := r.Regeocode.AddressComponent.City
	if len(raw) == 0 {
		return ""
	}
	var city string
	if err := json.Unmarshal(raw, &city); err != nil {
		return ""
	}
	return city
}

// AmapRouteResponse 高德驾车路径规划响应结构
type AmapRouteResponse struct {
	Status string `json:"status"`
	Info   string `json:"info"`
	Route  struct {
		Origin      string `json:"origin"`
		Destination string `json:"destination"`
		Paths       []struct {
			Distance string `json:"distance"`
			Duration string `json:"duration"`
			Strategy string `json:"strategy"`
			Steps    []struct {
				Instruction string `json:"instruction"`
				Polyline    string `json:"polyline"`
			} `json:"steps"`
		} `json:"paths"`
	} `json:"route"`
}

// RoutePlan 调用高德"驾车路径规划"接口，带重试 + 内存缓存降级。
func (c *Client) RoutePlan(originLat, originLng, destLat, destLng float64) (*AmapRouteResponse, error) {
	params := url.Values{}
	params.Set("key", c.apiKey)
	params.Set("origin", fmt.Sprintf("%.6f,%.6f", originLng, originLat))
	params.Set("destination", fmt.Sprintf("%.6f,%.6f", destLng, destLat))
	params.Set("extensions", "all")
	params.Set("output", "json")

	reqURL := fmt.Sprintf("%s/direction/driving?%s", c.baseURL, params.Encode())
	cacheKey := fmt.Sprintf("route:%.6f:%.6f:%.6f:%.6f", originLat, originLng, destLat, destLng)

	body := c.getCache(cacheKey)
	if body == nil {
		var err error
		body, err = c.doWithRetry(reqURL)
		if err != nil {
			return nil, err
		}
		c.setCache(cacheKey, body, defaultCacheTTL)
	}

	var result AmapRouteResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析高德驾车路径规划响应失败: %w", err)
	}
	if result.Status != "1" {
		return nil, fmt.Errorf("高德API返回错误: %s", result.Info)
	}
	if len(result.Route.Paths) == 0 {
		return nil, fmt.Errorf("高德未返回可行路线")
	}
	return &result, nil
}

// Distance 返回第一条规划路线的距离（米）
func (r *AmapRouteResponse) Distance() int32 {
	if len(r.Route.Paths) == 0 {
		return 0
	}
	n, _ := strconv.Atoi(r.Route.Paths[0].Distance)
	return int32(n)
}

// Duration 返回第一条规划路线的预计时间（秒）
func (r *AmapRouteResponse) Duration() int32 {
	if len(r.Route.Paths) == 0 {
		return 0
	}
	n, _ := strconv.Atoi(r.Route.Paths[0].Duration)
	return int32(n)
}

// Polyline 把第一条路线各 step 的路线点串拼成完整的一条
func (r *AmapRouteResponse) Polyline() string {
	if len(r.Route.Paths) == 0 {
		return ""
	}
	var parts []string
	for _, step := range r.Route.Paths[0].Steps {
		if p := strings.Trim(step.Polyline, ";"); p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, ";")
}

// parseLocation 解析高德坐标串 "经度,纬度"
func parseLocation(loc string) (lng, lat float64, err error) {
	parts := strings.Split(loc, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("坐标格式错误: %s", loc)
	}
	lng, err = strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, err
	}
	lat, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, 0, err
	}
	return lng, lat, nil
}
