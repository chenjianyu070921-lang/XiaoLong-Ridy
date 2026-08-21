# CreateDriver RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `CreateDriver(CreateDriverRequest) returns (CreateDriverResponse)` |
| 对应 HTTP | `POST /api/driver/v1/drivers` |
| 当前状态 | 已实现 |
| 业务逻辑 | `CreateDriverLogic.CreateDriver` |

## 2. 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `phone` | string | 是 | 手机号 |
| `password_hash` | string | 是 | bcrypt 哈希；HTTP API 会先把明文 `password` 转为哈希 |
| `real_name` | string | 是 | 真实姓名 |
| `id_card_no` | string | 是 | 身份证号 |
| `driver_license_no` | string | 是 | 驾驶证号 |
| `avatar_url` | string | 否 | 头像地址 |

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 司机 ID |
| `status` | DriverStatus | 初始为 `DRIVER_STATUS_PENDING` |
| `created_at` | int64 | Unix 秒 |

## 4. 处理链路

`api/driver.CreateDriver -> bcrypt -> driversvc.CreateDriver -> driver`。司机端 HTTP 文档见 [../05-driver-create.md](../05-driver-create.md)。
