# 验证码登录接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/auth/login-by-sms` |
| 是否登录 | 否 |
| 当前状态 | 已实现 |
| 下游 RPC | `driversvc.LoginBySMS` |

## 2. 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `phone` | string | 是 | 中国大陆手机号 |
| `code` | string | 是 | 登录验证码 |

## 3. 请求示例

```bash
curl -X POST http://127.0.0.1:8082/api/driver/v1/auth/login-by-sms \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800000000","code":"123456"}'
```

## 4. 响应字段

同 [密码登录接口](02-auth-login-by-password.md)。

## 5. 异常用例

| 用例编号 | 场景 | 请求体 | 预期 |
| --- | --- | --- | --- |
| DRIVER-AUTH-SMSLOGIN-E01 | 验证码错误 | `{"phone":"13800000000","code":"000000"}` | HTTP 400 / `41001` |
| DRIVER-AUTH-SMSLOGIN-E02 | 验证码过期 | 超过有效期后登录 | HTTP 400 / `41001` |
| DRIVER-AUTH-SMSLOGIN-E03 | 账号不可用 | 状态冻结或注销 | HTTP 403 / `40301` |

## 6. 处理链路

`api/driver -> CodeCache.Verify -> driversvc.LoginBySMS`。验证码只在 API 层校验，RPC 层负责按手机号签发登录态。
