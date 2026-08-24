package main

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

//go:embed templates/* static/css/* static/js/*
var webFS embed.FS

func main() {
	r := gin.Default()
	apiBase := driverAPIBase()
	httpClient := &http.Client{Timeout: 8 * time.Second}

	tpl := template.Must(template.ParseFS(webFS, "templates/*.html"))
	r.SetHTMLTemplate(tpl)

	staticFS, err := fs.Sub(webFS, "static")
	if err != nil {
		log.Fatalf("driver web static fs: %v", err)
	}
	r.StaticFS("/static", http.FS(staticFS))

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":       "Driver Console",
			"driverWSURL": driverWSURL(apiBase),
		})
	})

	r.POST("/driver/sms-code", proxyDriverAPI(httpClient, apiBase, http.MethodPost, "/api/driver/v1/auth/send-sms-code", false))
	r.POST("/driver/login", proxyDriverAPI(httpClient, apiBase, http.MethodPost, "/api/driver/v1/auth/login-by-password", false))
	r.POST("/driver/login-sms", proxyDriverAPI(httpClient, apiBase, http.MethodPost, "/api/driver/v1/auth/login-by-sms", false))
	r.POST("/driver/register", proxyDriverAPI(httpClient, apiBase, http.MethodPost, "/api/driver/v1/drivers/register", false))
	r.GET("/driver/me", proxyDriverAPI(httpClient, apiBase, http.MethodGet, "/api/driver/v1/drivers/get", true))
	r.GET("/driver/ai-score", proxyDriverAPI(httpClient, apiBase, http.MethodGet, "/api/driver/v1/drivers/ai-score", true))
	r.POST("/driver/online", proxyDriverAPI(httpClient, apiBase, http.MethodPost, "/api/driver/v1/drivers/online", true))
	r.POST("/driver/offline", proxyDriverAPI(httpClient, apiBase, http.MethodPost, "/api/driver/v1/drivers/offline", true))
	r.POST("/driver/heartbeat", proxyDriverAPI(httpClient, apiBase, http.MethodPost, "/api/driver/v1/drivers/heartbeat", true))
	r.POST("/driver/location/report", proxyDriverAPI(httpClient, apiBase, http.MethodPost, "/api/driver/v1/drivers/location/report", true))
	r.POST("/driver/update", proxyDriverAPI(httpClient, apiBase, http.MethodPost, "/api/driver/v1/drivers/update", true))
	r.POST("/driver/vehicles", proxyDriverAPI(httpClient, apiBase, http.MethodPost, "/api/driver/v1/vehicles", true))
	r.GET("/driver/vehicles/get", proxyDriverAPI(httpClient, apiBase, http.MethodGet, "/api/driver/v1/vehicles/get", true))
	r.POST("/driver/certification/upload", proxyDriverAPI(httpClient, apiBase, http.MethodPost, "/api/driver/v1/drivers/certification/upload", true))
	r.GET("/driver/certification", proxyDriverAPI(httpClient, apiBase, http.MethodGet, "/api/driver/v1/drivers/certification", true))
	r.POST("/driver/dispatches", proxyDriverAPI(httpClient, apiBase, http.MethodPost, "/api/driver/v1/orders/dispatches", true))
	r.POST("/driver/orders", proxyDriverAPI(httpClient, apiBase, http.MethodPost, "/api/driver/v1/orders/list", true))
	r.POST("/driver/orders/accept", proxyDriverAPI(httpClient, apiBase, http.MethodPost, "/api/driver/v1/orders/accept", true))
	r.POST("/driver/orders/reject", proxyDriverAPI(httpClient, apiBase, http.MethodPost, "/api/driver/v1/orders/reject", true))
	r.POST("/driver/orders/confirm-arrive", proxyDriverAPI(httpClient, apiBase, http.MethodPost, "/api/driver/v1/orders/confirm-arrive", true))
	r.POST("/driver/orders/start-trip", proxyDriverAPI(httpClient, apiBase, http.MethodPost, "/api/driver/v1/orders/start-trip", true))
	r.POST("/driver/orders/finish-trip", proxyDriverAPI(httpClient, apiBase, http.MethodPost, "/api/driver/v1/orders/finish-trip", true))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	webAddr := strings.TrimSpace(os.Getenv("DRIVER_WEB_ADDR"))
	if webAddr == "" {
		webAddr = ":8080"
	}
	if err := r.Run(webAddr); err != nil {
		log.Fatalf("driver web server exited: %v", err)
	}
}

func driverAPIBase() string {
	if value := strings.TrimSpace(os.Getenv("DRIVER_API_BASE")); value != "" {
		return strings.TrimRight(value, "/")
	}
	return "http://127.0.0.1:8082"
}

func driverWSURL(apiBase string) string {
	if value := strings.TrimSpace(os.Getenv("DRIVER_WS_URL")); value != "" {
		return value
	}
	target, err := url.Parse(strings.TrimRight(apiBase, "/") + "/api/driver/v1/ws")
	if err != nil {
		return ""
	}
	switch target.Scheme {
	case "https":
		target.Scheme = "wss"
	default:
		target.Scheme = "ws"
	}
	return target.String()
}

func proxyDriverAPI(client *http.Client, apiBase, method, path string, requireAuth bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetURL, err := buildTargetURL(apiBase, path, c.Request.URL.Query())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "message": "代理地址创建失败"})
			return
		}

		var body io.Reader = http.NoBody
		if method != http.MethodGet {
			rawBody, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 50000, "message": "请求体读取失败"})
				return
			}
			body = bytes.NewReader(rawBody)
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), method, targetURL, body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "message": "代理请求创建失败"})
			return
		}
		req.Header.Set("Accept", "application/json")
		if method != http.MethodGet {
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
		}
		if requireAuth {
			auth := strings.TrimSpace(c.GetHeader("Authorization"))
			if auth == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"code": 40102, "message": "登录凭证无效"})
				return
			}
			req.Header.Set("Authorization", auth)
		}

		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"code": 50000, "message": "driver-api 暂不可用"})
			return
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"code": 50000, "message": "driver-api 响应读取失败"})
			return
		}
		contentType := resp.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/json; charset=utf-8"
		}
		c.Data(resp.StatusCode, contentType, respBody)
	}
}

func buildTargetURL(apiBase, path string, query url.Values) (string, error) {
	target, err := url.Parse(apiBase + path)
	if err != nil {
		return "", err
	}
	values := target.Query()
	if id := strings.TrimSpace(query.Get("driverId")); id != "" {
		values.Set("id", id)
	}
	for key, list := range query {
		if key == "driverId" {
			continue
		}
		for _, value := range list {
			values.Add(key, value)
		}
	}
	target.RawQuery = values.Encode()
	return target.String(), nil
}
