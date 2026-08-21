# UpdateDriver RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `UpdateDriver(UpdateDriverRequest) returns (UpdateDriverResponse)` |
| 对应 HTTP | `POST /api/driver/v1/drivers/update` |
| 当前状态 | 已实现 |
| 业务逻辑 | `UpdateDriverLogic.UpdateDriver` |

## 2. 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 司机 ID |
| `phone` | optional string | 否 | 手机号 |
| `password_hash` | optional string | 否 | 新密码 bcrypt 哈希 |
| `real_name` | optional string | 否 | 真实姓名 |
| `id_card_no` | optional string | 否 | 身份证号 |
| `driver_license_no` | optional string | 否 | 驾驶证号 |
| `avatar_url` | optional string | 否 | 头像地址 |
| `status` | optional DriverStatus | 否 | 司机状态 |

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 司机 ID |
| `status` | DriverStatus | 更新后状态 |
| `updated_at` | int64 | Unix 秒 |

## 4. 处理链路

`api/driver.UpdateDriver -> driversvc.UpdateDriver -> driver`。optional 字段为 nil 时不更新。HTTP 文档见 [../06-driver-update.md](../06-driver-update.md)。
