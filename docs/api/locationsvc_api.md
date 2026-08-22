# locationsvc 位置服务接口文档

> 服务职责：地图与位置相关能力，包括逆地址解析、POI 搜索、路径规划、司机位置上报与附近司机查询。依赖 Redis（GEO + 缓存）、MySQL（POI 缓存库）以及高德地图 API。

## 1. 服务信息

| 项 | 值 |
| --- | --- |
| 服务名 | `locationsvc.rpc` |
| 协议 | gRPC（zrpc） |
| 监听地址 | `0.0.0.0:9001` |
| 注册中心 | etcd（`127.0.0.1:2379`，Key=`locationsvc.rpc`） |
| 外部依赖 | 高德地图 API（`MapService.ApiKey`，provider=`amap`，base=`https://restapi.amap.com/v3`） |
| 存储 | Redis（GEO Key `driver:geo` 用于附近司机查询）、MySQL（`xiaolongridy` 库的 `poi` 缓存表、运行期 `driver_location` 表） |

服务由 go-zero zrpc 暴露，调用方通过 proto 生成的 `locationsvcclient` 包或 grpc 直连接入。

## 2. 枚举与关键数据结构

### 2.1 OnlineStatus（司机听单状态，用于 `ReportLocation.online_status`）

| 值 | 说明 | 
| --- | --- |
| `0` | 离线 |
| `1` | 在线（听单中） |
| `2` | 行程中 |

> 说明：proto 中该字段为 `int32`，并非强枚举类型，调用方按上述约定的整数传值。
下面，
### 2.2 ReverseGeocodeResp（逆地址解析返回）

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `province` | string | 省 |
| `city` | string | 市（直辖市时高德 `city` 为空，已用 `province` 兜底，如"北京市"） |
| `district` | string | 区 |
| `address` | string | 详细地址 |
| `poi_name` | string | 附近 POI 名称（无则空串） |

### 2.3 POIItem（POI 条目）

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `name` | string | POI 名称 |
| `address` | string | 地址 |
| `lat` | double | 纬度 |
| `lng` | double | 经度 |
| `category` | string | 分类 |
| `distance` | int32 | 距离中心点的距离（米） |

### 2.4 RoutePlanResp（路径规划返回）

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `distance` | int32 | 距离（米） |
| `duration` | int32 | 预计时间（秒） |
| `polyline` | string | 路线点串（经纬度编码串，用于地图绘制） |
| `origin_lat` / `origin_lng` | double | 起点经纬度（原样回显） |
| `dest_lat` / `dest_lng` | double | 终点经纬度（原样回显） |

### 2.5 NearbyDriver（附近司机）

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `driver_id` | int64 | 司机 ID |
| `lng` | double | 经度 |
| `lat` | double | 纬度 |
| `distance` | double | 距查询中心点的距离（米，升序排列） |

## 3. 接口列表

服务共 5 个 RPC：`ReverseGeocode`、`POISearch`、`RoutePlan`、`ReportLocation`、`NearbyDrivers`。

---

### 3.1 ReverseGeocode — 逆地址解析（经纬度 → 地址）

将经纬度解析为结构化地址，调用高德 `regeo` 接口返回真实数据。

**请求 `ReverseGeocodeReq`**

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `lat` | double | 是 | 纬度 |
| `lng` | double | 是 | 经度 |

**响应 `ReverseGeocodeResp`**：见 2.2 节。

**说明**：仅支持高德服务范围内坐标（中国境内）；境外坐标 `address` 为空，接口返回错误 `无法解析该坐标对应的地址（可能不在高德服务范围内）`。

---

### 3.2 POISearch — POI 搜索

按关键词在指定中心点附近搜索 POI。先查本地 `poi` 缓存库，未命中再调用高德 POI 搜索，结果回写缓存（下次同关键词直接命中本地库，不再调用高德）。

**请求 `POISearchReq`**

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `keyword` | string | 是 | 搜索关键词 |
| `lat` | double | 是 | 中心纬度 |
| `lng` | double | 是 | 中心经度 |
| `radius` | int32 | 否 | 搜索半径（米），透传高德 |
| `page` | int32 | 否 | 页码 |
| `size` | int32 | 否 | 每页条数 |

**响应 `POISearchResp`**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `items` | repeated POIItem | POI 列表（见 2.3 节） |
| `total` | int32 | 总条数（高德返回） |

---

### 3.3 RoutePlan — 路径规划

计算两点间驾车路线，调用高德 `direction/driving` 接口返回真实可行驶路线。

**请求 `RoutePlanReq`**

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `origin_lat` | double | 是 | 起点纬度 |
| `origin_lng` | double | 是 | 起点经度 |
| `destination_lat` | double | 是 | 终点纬度 |
| `destination_lng` | double | 是 | 终点经度 |

**响应 `RoutePlanResp`**：见 2.4 节。

---

### 3.4 ReportLocation — 司机位置上报

司机实时位置上报：同时落库 `driver_location` 表（`driver_id` 唯一，冲突则更新最新位置）并写入 Redis GEO（供 `NearbyDrivers` 查询）。

**请求 `ReportLocationReq`**

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `driver_id` | int64 | 是 | 司机 ID（≤0 报错） |
| `lng` | double | 是 | 经度 |
| `lat` | double | 是 | 纬度 |
| `heading` | int32 | 否 | 行驶方向（0-359 度） |
| `speed_kmh` | double | 否 | 当前速度（km/h） |
| `online_status` | int32 | 否 | 听单状态，见 2.1 节 |
| `order_id` | int64 | 否 | 关联订单 ID，无则 0 |

**响应 `ReportLocationResp`**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `success` | bool | 上报是否成功 |

**说明**：经纬度越界（`lat` 超出 [-90,90] 或 `lng` 超出 [-180,180]）返回错误。

---

### 3.5 NearbyDrivers — 附近司机查询

基于 Redis GEO 半径搜索附近司机，按距离升序返回。

**请求 `NearbyDriversReq`**

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `lng` | double | 是 | 中心经度 |
| `lat` | double | 是 | 中心纬度 |
| `radius` | double | 是 | 搜索半径（米，必须 >0） |
| `limit` | int32 | 否 | 返回条数上限；默认 20，最大 100 |

**响应 `NearbyDriversResp`**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `drivers` | repeated NearbyDriver | 司机列表，按 `distance` 升序（见 2.5 节） |

**说明**：仅返回已通过 `ReportLocation` 写入 Redis GEO 的司机；未上报过位置的司机不会出现在结果中（返回空列表是正常的）。

---

## 4. 错误约定

| 错误 | 触发场景 |
| --- | --- |
| `driver_id 非法` | ReportLocation 的 `driver_id ≤ 0` |
| `经纬度非法` | 经纬度超出合法范围（ReportLocation / NearbyDrivers） |
| `radius 必须大于 0` | NearbyDrivers 的 `radius ≤ 0` |
| `无法解析该坐标对应的地址…` | ReverseGeocode 坐标不在高德服务范围内 |
| 高德 API 错误 | ReverseGeocode / POISearch / RoutePlan 调用高德失败，透传底层错误 |

> 错误通过 gRPC status 返回，调用方应使用 `status.FromError` 解析。

## 5. 调用方接入指引

1. 在调用方服务配置中声明 `locationsvc` 的 zrpc 客户端（target 指向 `127.0.0.1:9001` 或 etcd 服务发现名 `locationsvc.rpc`）。
2. 引入 `rpc/locationsvc/locationsvc` 包生成的 `locationsvcclient`，用 `NewLocationService(zrpc.MustNewClient(c))` 获取 client。
3. 调用示例（Go）：

```go
client := locationsvcclient.NewLocationService(zrpc.MustNewClient(c))

// 附近司机查询
resp, err := client.NearbyDrivers(ctx, &locationsvc.NearbyDriversReq{
    Lng:    116.397,
    Lat:    39.908,
    Radius: 3000,
    Limit:  10,
})

// 司机位置上报（需先上报，查询才有结果）
_, _ = client.ReportLocation(ctx, &locationsvc.ReportLocationReq{
    DriverId:    1001,
    Lng:         116.398,
    Lat:         39.909,
    OnlineStatus: 1,
})
```

## 6. 注意事项

- `NearbyDrivers` 的结果完全依赖 `ReportLocation` 写入的 Redis GEO 数据；调用查询前必须先有司机上报位置，否则返回空列表。
- 高德 `ApiKey` 配置在 `etc/locationsvc.yaml` 的 `MapService.ApiKey`，缺失或越权会导致 3.1/3.2/3.3 调用失败。
- `POISearch` 首次请求命中高德后会写回本地 `poi` 缓存表，后续相同关键词直接走缓存；如需刷新缓存需清理对应数据。
- **待优化项（当前未实现）**：`driver:geo` 位置 Key 的过期策略、热点城市分桶优化，目前所有司机共用同一 GEO Key，海量司机场景下查询性能需进一步治理。
