# LoginBySMS RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `LoginBySMS(LoginBySMSRequest) returns (LoginResponse)` |
| 对应 HTTP | `POST /api/driver/v1/auth/login-by-sms` |
| 当前状态 | 已实现 |
| 业务逻辑 | `LoginLogic.LoginBySMS` |

## 2. 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `phone` | string | 是 | 手机号 |

## 3. 响应字段

同 [Login](17-login.md)。验证码由 HTTP API 层校验，不传入 RPC。

## 4. 处理链路

`api/driver.CodeCache.Verify -> driversvc.LoginBySMS`。HTTP 文档见 [../03-auth-login-by-sms.md](../03-auth-login-by-sms.md)。
