# GetDriver RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `GetDriver(GetDriverRequest) returns (GetDriverResponse)` |
| 对应 HTTP | `GET /api/driver/v1/drivers/get?id=` |
| 当前状态 | 已实现 |
| 业务逻辑 | `GetDriverLogic.GetDriver` |

## 2. 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 司机 ID |

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `driver` | Driver | 司机详情；RPC 包含完整手机号/身份证/密码哈希 |

## 4. API 对齐

HTTP API 会对 `phone`、`idCardNo` 脱敏，并不返回 `password_hash`。HTTP 文档见 [../07-driver-get.md](../07-driver-get.md)。
