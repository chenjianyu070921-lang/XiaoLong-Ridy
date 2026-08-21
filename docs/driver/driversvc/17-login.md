# Login RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `Login(LoginRequest) returns (LoginResponse)` |
| 对应 HTTP | `POST /api/driver/v1/auth/login-by-password` |
| 当前状态 | 已实现 |
| 业务逻辑 | `LoginLogic.Login` |

## 2. 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `phone` | string | 是 | 手机号 |
| `password` | string | 是 | 明文密码 |

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `token` | string | JWT |
| `expire_in` | int64 | 有效期，单位秒 |
| `driver` | Driver | 司机摘要 |

## 4. 处理链路

`api/driver.LoginByPassword -> driversvc.Login`。HTTP 文档见 [../02-auth-login-by-password.md](../02-auth-login-by-password.md)。
