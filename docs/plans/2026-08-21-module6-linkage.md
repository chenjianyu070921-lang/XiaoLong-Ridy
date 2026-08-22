# 模块六链路打通 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 修复同行评审指出的“业务链路断点”——位置流无生产者/无消费者、WebSocket 不存在、GEO key 硬编码不一致，使司机位置从上报→流→消费→在线状态→实时推送整链路可运行。

**Architecture:** 在 locationsvc ReportLocation 写入 Redis GEO 的同时，向 Redis Stream `driver:location:stream` 发布位置事件；新建 location-consumer 用 Redis XReadGroup 消费该 Stream，维护 `driver:online` 在线集合并更新 DB online_status；api/driver 提供 HTTP 上报入口与 WebSocket 实时位置推送；公共常量统一收敛到 common/constants。

**Tech Stack:** Go / go-zero / go-redis / Redis Stream / gorilla/websocket（若可装）

---

### Task 1: 抽取公共常量并统一引用
**Files:**
- Modify: `common/constants/constants.go`（新增 `DriverGeoKey` / `DriverOnlineKey` / `LocationStreamKey`）
- Modify: `rpc/locationsvc/internal/svc/servicecontext.go`（改为引用 `constants.DriverGeoKey`）
- Modify: `job/internal/task/task.go`（本地 `driverGeoKey` 改为引用 `constants.DriverGeoKey`）

**Step:** 直接替换常量定义，保持字符串值不变，编译确认无差异。

---

### Task 2: locationsvc 发布位置事件（生产者）
**Files:** Modify `rpc/locationsvc/internal/logic/reportlocationlogic.go`

**Step:** 在 GeoAdd 之后 `Redis.XAdd(ctx, &redis.XAddArgs{Stream: constants.LocationStreamKey, Values: ...})`，字段 driver_id/lng/lat/online_status/ts。

---

### Task 3: location-consumer 真正消费（打通链路）
**Files:**
- Modify: `mq-consumer/location-consumer/internal/config/config.go`（Kafka → RedisConf）
- Modify: `mq-consumer/location-consumer/internal/handler/locationhandler.go`（保留消息结构）
- 重写: `mq-consumer/location-consumer/locationconsumer.go`（XReadGroup 消费 Stream，维护 driver:online，更新 online_status）
- Modify: `mq-consumer/location-consumer/etc/location-consumer.yaml`（RedisConf 替换 Kafka）

**Step:** 启动消费者组读取 `driver:location:stream`，解析字段，ZADD/SADD 维护在线集合，必要时更新 DB。

---

### Task 4: api/driver 增加司机位置上报入口（生产者前端入口）
**Files:** Modify `api/driver/driver.go` + 新增 `api/driver/internal/handler/report_location.go`
**Step:** 增加 `/api/driver/v1/drivers/report-location`，内部调用 locationsvc ReportLocation。

---

### Task 5: api/driver 增加 WebSocket 实时位置推送
**Files:** 新增 `api/driver/internal/ws/ws.go` + 修改 `api/driver/driver.go`
**Step:** 用 gorilla/websocket（或裸实现）提供 `/ws/location?driverId=`，周期从 Redis GEO 读取该司机位置推送客户端。

---

### Task 6: 编译验证 + 单测
**Step:** `go build ./...`；为 consumer 解析函数与 locationsvc 关键逻辑补单测；实际启动 locationsvc + location-consumer，上报后验证 driver:online 与 DB online_status 更新。
