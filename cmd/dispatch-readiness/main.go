// Command dispatch-readiness 只读诊断指定司机当前是否具备「接到派单」的条件。
//
// 用法（Windows PowerShell，仓库根目录执行）：
//
//	go run ./cmd/dispatch-readiness -driver-id 12
//	go run ./cmd/dispatch-readiness -driver-id 12 -lng 118.31 -lat 33.96
//
// 连接信息默认从 rpc/dispatchsvc/etc/dispatchsvc.yaml 读取（myredis / mysql），
// 可用 -redis-addr / -redis-pass / -redis-db / -mysql-dsn 覆盖，便于连远端环境。
// 全程只读（PING / GET / HGETALL / SISMEMBER / GEOPOS / GEOSEARCH / PUBSUB CHANNELS），不写入任何数据。
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"XiaoLong-Ridy/common/constants"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
)

// dispatchRadiusLevels 与派单引擎保持一致的半径梯度（米）。
// 来源：rpc/dispatchsvc/internal/engine/geo_dispatch_engine.go
var dispatchRadiusLevels = []float64{3000, 5000, 10000, 20000}

// dispatchMaxCandidates 单次 GEO 搜索取回的候选数，自检时放宽以便观察排名。
const dispatchMaxCandidates = 50

// driverStatusNormal 司机账号正常态；来源 rpc/driversvc/proto/driversvc.proto。
const driverStatusNormal = 2

// 在线状态语义；来源 rpc/driversvc/internal/onlinestore/online_store.go。
const (
	onlineStatusOffline = 0
	onlineStatusOnline  = 1
	onlineStatusOnTrip  = 2
)

// 检测结果状态：FAIL 为阻断项（必然接不到派单），WARN 为降级提示，INFO 为参考信息。
const (
	stateOK   = "OK"
	stateFAIL = "FAIL"
	stateWARN = "WARN"
	stateINFO = "INFO"
	stateSKIP = "SKIP"
)

type svcConfig struct {
	Mysql struct {
		DSN string `yaml:"dsn"`
	} `yaml:"mysql"`
	MyRedis struct {
		Host string `yaml:"host"`
		Pass string `yaml:"pass"`
		DB   int    `yaml:"db"`
	} `yaml:"myredis"`
}

type checkResult struct {
	name   string
	state  string
	detail string
	fix    string
}

func main() {
	driverID := flag.Int64("driver-id", 0, "待检测的司机 ID（必填）")
	configPath := flag.String("config", "rpc/dispatchsvc/etc/dispatchsvc.yaml", "服务配置文件路径，用于读取 Redis/MySQL 连接信息")
	redisAddr := flag.String("redis-addr", "", "Redis 地址，覆盖配置文件")
	redisPass := flag.String("redis-pass", "", "Redis 密码，覆盖配置文件")
	redisDB := flag.Int("redis-db", -1, "Redis DB，覆盖配置文件")
	mysqlDSN := flag.String("mysql-dsn", "", "MySQL DSN，覆盖配置文件；留空则跳过账号状态检查")
	lng := flag.Float64("lng", 0, "模拟上车点经度，用于验证能否被派单引擎搜到；留空则使用司机当前坐标")
	lat := flag.Float64("lat", 0, "模拟上车点纬度")
	timeout := flag.Duration("timeout", 10*time.Second, "单次检查超时时间")
	flag.Parse()

	if *driverID <= 0 {
		fmt.Fprintln(os.Stderr, "错误：必须使用 -driver-id 指定司机 ID")
		flag.Usage()
		os.Exit(2)
	}

	cfg, cfgErr := loadConfig(*configPath)
	if cfgErr != nil {
		fmt.Fprintf(os.Stderr, "警告：读取配置 %s 失败（%v），连接信息需通过命令行参数提供\n", *configPath, cfgErr)
	}

	addr := firstNonEmpty(*redisAddr, cfg.MyRedis.Host)
	pass := firstNonEmpty(*redisPass, cfg.MyRedis.Pass)
	db := cfg.MyRedis.DB
	if *redisDB >= 0 {
		db = *redisDB
	}
	if addr == "" {
		fmt.Fprintln(os.Stderr, "错误：未能确定 Redis 地址，请使用 -redis-addr 指定")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	rdb := redis.NewClient(&redis.Options{Addr: addr, Password: pass, DB: db})
	defer rdb.Close()

	fmt.Printf("司机派单可接性自检\n")
	fmt.Printf("  司机 ID   : %d\n", *driverID)
	fmt.Printf("  Redis     : %s (db=%d)\n", addr, db)
	if *lng != 0 || *lat != 0 {
		fmt.Printf("  模拟上车点: %.6f, %.6f\n", *lng, *lat)
	}
	fmt.Println(strings.Repeat("-", 78))

	if err := rdb.Ping(ctx).Err(); err != nil {
		fmt.Fprintf(os.Stderr, "错误：Redis 连接失败：%v\n", err)
		os.Exit(2)
	}

	dsn := firstNonEmpty(*mysqlDSN, cfg.Mysql.DSN)
	results := runChecks(ctx, rdb, *driverID, dsn, *lng, *lat)

	blocked := 0
	for _, r := range results {
		fmt.Printf("[%-4s] %-12s %s\n", r.state, r.name, r.detail)
		if r.state == stateFAIL {
			blocked++
		}
	}

	fmt.Println(strings.Repeat("-", 78))
	if blocked > 0 {
		fmt.Printf("结论：当前接不到派单，存在 %d 项阻断。修复动作：\n", blocked)
		idx := 0
		for _, r := range results {
			if r.state == stateFAIL && r.fix != "" {
				idx++
				fmt.Printf("  %d. [%s] %s\n", idx, r.name, r.fix)
			}
		}
		os.Exit(1)
	}
	fmt.Println("结论：全部闸门通过，该司机当前可被派单引擎选中并收到推送。")
}

// loadConfig 解析服务配置文件，用于复用仓库已有的 Redis/MySQL 连接信息。
func loadConfig(path string) (*svcConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return &svcConfig{}, err
	}
	var cfg svcConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return &svcConfig{}, err
	}
	return &cfg, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// runChecks 按派单链路的真实顺序逐项检查司机接单前置条件。
func runChecks(ctx context.Context, rdb *redis.Client, driverID int64, dsn string, lng, lat float64) []checkResult {
	member := strconv.FormatInt(driverID, 10)
	results := []checkResult{checkAccountStatus(ctx, dsn, driverID)}

	// 闸门 1：司机在线状态（Redis hash，心跳保活，TTL 3 分钟）。
	onlineHash := fmt.Sprintf("%s:%d", constants.RedisDriverOnline, driverID)
	vals, err := rdb.HGetAll(ctx, onlineHash).Result()
	ttl, ttlErr := rdb.TTL(ctx, onlineHash).Result()
	switch {
	case err != nil:
		results = append(results, checkResult{"在线状态", stateFAIL, "读取失败：" + err.Error(), "检查 Redis 连通性与权限"})
	case len(vals) == 0:
		results = append(results, checkResult{"在线状态", stateFAIL, "driver:online:<id> 不存在，司机从未上线或心跳已过期",
			"调用 POST /api/driver/v1/drivers/online 上线，并保持心跳（TTL 3 分钟内必须续期）"})
	default:
		status, _ := strconv.Atoi(vals["online_status"])
		detail := fmt.Sprintf("online_status=%d(%s) device=%s 位置=%s,%s", status, onlineStatusText(status),
			orDefault(vals["device_id"], "-"), orDefault(vals["longitude"], "-"), orDefault(vals["latitude"], "-"))
		if ttlErr == nil && ttl > 0 {
			detail += fmt.Sprintf(" TTL=%.0fs", ttl.Seconds())
		}
		switch status {
		case onlineStatusOnline:
			results = append(results, checkResult{"在线状态", stateOK, detail, ""})
		case onlineStatusOnTrip:
			results = append(results, checkResult{"在线状态", stateFAIL, detail,
				"司机处于行程中，需先结束当前行程（driver:busy 会随订单完成/取消移除）"})
		default:
			results = append(results, checkResult{"在线状态", stateFAIL, detail,
				"司机未在线，调用 POST /api/driver/v1/drivers/online 上线"})
		}
	}

	// 闸门 2：是否在派单池集合 driver:online 内。
	inPool, err := rdb.SIsMember(ctx, constants.RedisDriverOnline, member).Result()
	switch {
	case err != nil:
		results = append(results, checkResult{"派单池", stateFAIL, "读取失败：" + err.Error(), "检查 Redis 连通性"})
	case !inPool:
		results = append(results, checkResult{"派单池", stateFAIL, fmt.Sprintf("不在集合 %s 内，派单引擎 filterAvailable 会直接过滤", constants.RedisDriverOnline),
			"重新上线（上线会写 driver:online 集合）；若已上线仍不在，检查上报端 Redis 兼容层是否丢弃了 SADD"})
	default:
		results = append(results, checkResult{"派单池", stateOK, fmt.Sprintf("已在集合 %s 内", constants.RedisDriverOnline), ""})
	}

	// 闸门 3：GEO 坐标是否存在于 driver:geo:default。
	geoKey := fmt.Sprintf(constants.RedisDriverGeo, "default")
	pos, err := rdb.GeoPos(ctx, geoKey, member).Result()
	driverLng, driverLat := 0.0, 0.0
	hasGeo := false
	if err == nil && len(pos) > 0 && pos[0] != nil && (pos[0].Longitude != 0 || pos[0].Latitude != 0) {
		hasGeo = true
		driverLng, driverLat = pos[0].Longitude, pos[0].Latitude
	}
	if !hasGeo {
		results = append(results, checkResult{"GEO 位置", stateFAIL, fmt.Sprintf("在 %s 中没有坐标", geoKey),
			"上线并上报位置（POST 位置上报接口），坐标由 driversvc 写入 driver:geo:default"})
	} else {
		results = append(results, checkResult{"GEO 位置", stateOK,
			fmt.Sprintf("%s 坐标 %.6f, %.6f", geoKey, driverLng, driverLat), ""})
	}

	// 位置缓存新鲜度：driver:pos:<id> TTL 2 分钟，过期不影响 GEO，但影响司机端订单展示。
	posKey := fmt.Sprintf(constants.RedisDriverPos, driverID)
	if posTTL, err := rdb.TTL(ctx, posKey).Result(); err == nil && posTTL > 0 {
		results = append(results, checkResult{"位置缓存", stateOK, fmt.Sprintf("%s TTL=%.0fs", posKey, posTTL.Seconds()), ""})
	} else {
		results = append(results, checkResult{"位置缓存", stateWARN, fmt.Sprintf("%s 已过期或不存在（不影响派单，影响司机端订单列表）", posKey),
			"保持位置上报频率高于 2 分钟一次"})
	}

	// 闸门 4：是否处于忙碌集合 driver:busy。
	busy, err := rdb.SIsMember(ctx, constants.RedisDriverBusy, member).Result()
	switch {
	case err != nil:
		results = append(results, checkResult{"忙碌状态", stateFAIL, "读取失败：" + err.Error(), "检查 Redis 连通性"})
	case busy:
		results = append(results, checkResult{"忙碌状态", stateFAIL, fmt.Sprintf("在集合 %s 内，派单引擎会跳过忙碌司机", constants.RedisDriverBusy),
			"完成/取消当前订单让 ordersvc 移除 busy；异常残留可让司机重新上线（上线时会 SREM 清理）"})
	default:
		results = append(results, checkResult{"忙碌状态", stateOK, "未处于忙碌状态", ""})
	}

	// 闸门 5：能否被派单引擎按半径梯度搜到（复现 geoDispatchEngine.FindCandidates 的搜索行为）。
	results = append(results, checkReachable(ctx, rdb, driverID, member, geoKey, lng, lat, driverLng, driverLat, hasGeo))

	// 推送通道：driver:push:<id> 是否有订阅者（driver-api 的 WebSocket 连接）。
	pushChannel := fmt.Sprintf(constants.RedisDriverPush, driverID)
	if channels, err := rdb.PubSubChannels(ctx, pushChannel).Result(); err == nil && len(channels) > 0 {
		results = append(results, checkResult{"推送通道", stateOK, fmt.Sprintf("%s 有 %d 个订阅者，WS 在线", pushChannel, len(channels)), ""})
	} else {
		results = append(results, checkResult{"推送通道", stateWARN, fmt.Sprintf("%s 无订阅者，司机端 WebSocket 未连接", pushChannel),
			"让司机端保持 WebSocket 连接；未连接时仍可轮询 /api/driver/v1/orders/available 兜底"})
	}

	// 待接派单：driver:available:<id> 非空说明已有派单在等待接单（TTL 90s）。
	availableKey := fmt.Sprintf(constants.RedisDriverAvailable, driverID)
	if pending, err := rdb.SMembers(ctx, availableKey).Result(); err == nil && len(pending) > 0 {
		if availTTL, _ := rdb.TTL(ctx, availableKey).Result(); availTTL > 0 {
			results = append(results, checkResult{"待接派单", stateINFO,
				fmt.Sprintf("%s 有 %d 个待接订单 %v，剩余 %.0fs", availableKey, len(pending), pending, availTTL.Seconds()), ""})
		} else {
			results = append(results, checkResult{"待接派单", stateINFO,
				fmt.Sprintf("%s 有 %d 个待接订单 %v", availableKey, len(pending), pending), ""})
		}
	} else {
		results = append(results, checkResult{"待接派单", stateINFO, fmt.Sprintf("%s 为空，当前没有派给该司机的订单", availableKey), ""})
	}

	return results
}

// checkReachable 用派单引擎相同的 GEO 搜索方式验证司机能否被选中，并输出命中的半径等级。
func checkReachable(ctx context.Context, rdb *redis.Client, driverID int64, member, geoKey string, lng, lat, driverLng, driverLat float64, hasGeo bool) checkResult {
	if !hasGeo {
		return checkResult{"GEO 可达", stateFAIL, "无坐标，无法参与 GEO 搜索", "先补齐 GEO 坐标"}
	}
	// 未指定模拟上车点时，用司机自身坐标做自搜，验证 GEO 桶与可用性。
	if lng == 0 && lat == 0 {
		lng, lat = driverLng, driverLat
	}
	memberID, _ := strconv.ParseInt(member, 10, 64)
	for _, radius := range dispatchRadiusLevels {
		locs, err := rdb.GeoSearchLocation(ctx, geoKey, &redis.GeoSearchLocationQuery{
			GeoSearchQuery: redis.GeoSearchQuery{
				Longitude:  lng,
				Latitude:   lat,
				Radius:     radius,
				RadiusUnit: "m",
				Sort:       "ASC",
				Count:      dispatchMaxCandidates,
			},
			WithDist: true,
		}).Result()
		if err != nil && err != redis.Nil {
			return checkResult{"GEO 可达", stateFAIL, "GEO 搜索失败：" + err.Error(), "检查 Redis 版本是否支持 GEOSEARCH"}
		}
		rank := 0
		for i, loc := range locs {
			if id, err := strconv.ParseInt(loc.Name, 10, 64); err == nil && id == memberID {
				rank = i + 1
				break
			}
		}
		if rank > 0 {
			return checkResult{"GEO 可达", stateOK, fmt.Sprintf("距上车点 %.0fm 半径内命中，排名 %d/%d", radius, rank, len(locs)), ""}
		}
	}
	return checkResult{"GEO 可达", stateFAIL, "在 3/5/10/20km 半径内均未被搜到（可能超出最大半径，或 GEO 桶与派单引擎配置不一致）",
		"确认司机坐标与上车点距离在 20km 内；确认 driversvc 写入的 GEO 桶（默认 driver:geo:default）与 dispatchsvc 读取的桶一致"}
}

// checkAccountStatus 校验司机账号是否为正常态（driversvc 上线时的前置校验）。
func checkAccountStatus(ctx context.Context, dsn string, driverID int64) checkResult {
	if dsn == "" {
		return checkResult{"账号状态", stateSKIP, "未提供 MySQL DSN，跳过", "传入 -mysql-dsn 可校验账号状态"}
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return checkResult{"账号状态", stateWARN, "打开数据库失败：" + err.Error(), "检查 DSN 与网络可达性"}
	}
	defer db.Close()

	var (
		id           uint64
		phone        string
		status       int
		onlineStatus int
	)
	err = db.QueryRowContext(ctx,
		"SELECT id, phone, status, online_status FROM driver WHERE id = ? AND deleted_at IS NULL", driverID,
	).Scan(&id, &phone, &status, &onlineStatus)
	switch {
	case err == sql.ErrNoRows:
		return checkResult{"账号状态", stateFAIL, fmt.Sprintf("driver 表中不存在 id=%d 的司机", driverID), "确认司机 ID 是否正确"}
	case err != nil:
		return checkResult{"账号状态", stateWARN, "查询失败：" + err.Error(), "检查 DSN 与网络可达性"}
	}
	if status != driverStatusNormal {
		return checkResult{"账号状态", stateFAIL,
			fmt.Sprintf("status=%d(%s)，上线接口会返回 PermissionDenied", status, driverStatusText(status)),
			"将 driver.status 置为 2（NORMAL），否则 SetDriverOnline 直接拒绝"}
	}
	return checkResult{"账号状态", stateOK, fmt.Sprintf("status=2(NORMAL) phone=%s db_online_status=%d", maskPhone(phone), onlineStatus), ""}
}

func driverStatusText(status int) string {
	switch status {
	case 0:
		return "UNSPECIFIED"
	case 1:
		return "PENDING"
	case 2:
		return "NORMAL"
	case 3:
		return "FROZEN"
	case 4:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

func onlineStatusText(status int) string {
	switch status {
	case onlineStatusOnline:
		return "在线"
	case onlineStatusOnTrip:
		return "行程中"
	default:
		return "离线"
	}
}

func maskPhone(phone string) string {
	if len(phone) <= 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

func orDefault(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}
