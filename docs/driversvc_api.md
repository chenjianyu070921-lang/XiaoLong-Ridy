# driversvc 司机服务接口文档

> 服务职责：司机主表（`driver` 库）的增删改查，仅负责司机账号数据，不含订单、审核流程等业务编排。

## 1. 服务信息

| 项 | 值 |
| --- | --- |
| 服务名 | `driversvc.rpc` |
| 协议 | gRPC（zrpc） |
| 监听地址 | `0.0.0.0:8080` |
| 数据库 | `driver` 库（dsn 见 `etc/driversvc.yaml`） |
| 删除语义 | 软删除（GORM `deleted_at`，不物理删除） |

服务由 go-zero zrpc 暴露，调用方通过 proto 生成的 `driversvcclient` 包或 grpc 直连接入。

## 2. 数据结构

### DriverStatus 枚举

| 枚举名 | 值 | 说明 |
| --- | --- | --- |
| `DRIVER_STATUS_UNSPECIFIED` | 0 | 未指定（默认值，不应作为有效状态） |
| `DRIVER_STATUS_PENDING` | 1 | 待审核（新建账号初始状态） |
| `DRIVER_STATUS_NORMAL` | 2 | 正常 |
| `DRIVER_STATUS_FROZEN` | 3 | 已冻结 |
| `DRIVER_STATUS_CANCELLED` | 4 | 已注销 |

### Driver（司机信息）

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 主键 |
| `phone` | string | 手机号 |
| `password_hash` | string | 密码哈希（注意：明文密码不在接口流转） |
| `real_name` | string | 真实姓名 |
| `id_card_no` | string | 身份证号 |
| `driver_license_no` | string | 驾驶证号 |
| `avatar_url` | string | 头像地址 |
| `status` | DriverStatus | 账号状态 |
| `created_at` | int64 | 创建时间（Unix 秒） |
| `updated_at` | int64 | 更新时间（Unix 秒） |

## 3. 接口列表

服务共 4 个 RPC：`CreateDriver`、`UpdateDriver`、`GetDriver`、`DeleteDriver`。

---

### 3.1 CreateDriver — 创建司机

新建司机账号，状态固定初始化为 `PENDING`（待审核）。

**请求 `CreateDriverRequest`**

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `phone` | string | 是 | 手机号 |
| `password_hash` | string | 是 | 密码哈希 |
| `real_name` | string | 否 | 真实姓名 |
| `id_card_no` | string | 否 | 身份证号 |
| `driver_license_no` | string | 否 | 驾驶证号 |
| `avatar_url` | string | 否 | 头像地址 |

**响应 `CreateDriverResponse`**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 新创建的司机 ID |
| `status` | DriverStatus | 始终返回 `PENDING` |
| `created_at` | int64 | 创建时间（Unix 秒） |

---

### 3.2 UpdateDriver — 更新司机

按 `id` 定位司机，**仅更新请求中显式传入的字段**（proto3 `optional` 字段，未传为 `nil` 时不更新）。

**请求 `UpdateDriverRequest`**

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 司机 ID |
| `phone` | optional string | 否 | 传入则更新手机号 |
| `password_hash` | optional string | 否 | 传入则更新密码哈希 |
| `real_name` | optional string | 否 | 传入则更新真实姓名 |
| `id_card_no` | optional string | 否 | 传入则更新身份证号 |
| `driver_license_no` | optional string | 否 | 传入则更新驾驶证号 |
| `avatar_url` | optional string | 否 | 传入则更新头像地址 |
| `status` | optional DriverStatus | 否 | 传入则更新状态（如审批通过置 `NORMAL`，冻结置 `FROZEN`） |

**响应 `UpdateDriverResponse`**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 司机 ID |
| `status` | DriverStatus | 更新后的状态 |
| `updated_at` | int64 | 更新时间（Unix 秒） |

**说明**：更新前会先校验司机存在性；若司机不存在（含已被软删）返回 `driver not found` 错误。

---

### 3.3 GetDriver — 查询司机

按 `id` 返回司机完整信息。

**请求 `GetDriverRequest`**

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 司机 ID |

**响应 `GetDriverResponse`**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `driver` | Driver | 司机完整信息（见第 2 节） |

**说明**：软删除的司机不可见，查询会返回 `driver not found`。

---

### 3.4 DeleteDriver — 删除司机

按 `id` 软删除司机（设置 `deleted_at`，数据仍保留在库中）。

**请求 `DeleteDriverRequest`**

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 司机 ID |

**响应 `DeleteDriverResponse`**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 司机 ID |
| `success` | bool | 删除是否成功 |

**说明**：删除前会先校验司机存在性；不存在返回 `driver not found`。软删除后，该司机在 `GetDriver`/`UpdateDriver` 中均不可见。

---

## 4. 错误约定

| 错误 | 触发场景 |
| --- | --- |
| `driver not found` | 对不存在或已被软删除的司机执行 `Get`/`Update`/`Delete` |
| 数据库错误 | GORM 写入/查询异常时透传底层错误 |

> 错误通过 gRPC status 返回，调用方应使用 `status.FromError` 解析。

## 5. 调用方接入指引

1. 在调用方服务配置中声明 `driversvc` 的 zrpc 客户端（target 指向 `127.0.0.1:8080` 或 etcd 服务发现名 `driversvc.rpc`）。
2. 引入 `rpc/driversvc/proto` 包生成的 `driversvcclient`，用 `NewDriversvc(zrpc.MustNewClient(c))` 获取 client。
3. 调用示例（Go）：

```go
client := driversvcclient.NewDriversvc(zrpc.MustNewClient(c))
resp, err := client.CreateDriver(ctx, &driversvc.CreateDriverRequest{
    Phone:        "13800000000",
    PasswordHash: "...",
})
```

> 注意：`password_hash` 由调用方在上游完成哈希后再传入，本服务不负责密码加密。

## 6. 注意事项

- 本服务仅管理司机主表数据，认证、审核流程、订单关联等业务逻辑由上游（如 admin / order）负责。
- `status` 的初始值与流转由调用方控制：创建固定 `PENDING`，后续由 `UpdateDriver` 修改（如 admin 审批通过置 `NORMAL`）。
- 所有时间戳以 Unix 秒（int64）在网络传输，由 GORM `parseTime=True` 映射数据库 `DATETIME`。
