# 密码登录接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/auth/login-by-password` |
| 是否登录 | 否 |
| 当前状态 | 已实现 |
| 下游 RPC | `driversvc.Login` |

## 2. 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `phone` | string | 是 | 中国大陆手机号 |
| `password` | string | 是 | 明文密码，长度 `8~72` 字节 |

## 3. 请求示例

```bash
curl -X POST http://127.0.0.1:8082/api/driver/v1/auth/login-by-password \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800000000","password":"Driver@123"}'
```

## 4. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `token` | string | 司机 JWT |
| `expireIn` | int64 | Token 有效期，单位秒 |
| `driver.id` | int64 | 司机 ID |
| `driver.phone` | string | 脱敏手机号 |
| `driver.status` | string | 司机状态 |

## 5. 异常用例

| 用例编号 | 场景 | 请求体 | 预期 |
| --- | --- | --- | --- |
| DRIVER-AUTH-PWD-E01 | 手机号格式错误 | `{"phone":"abc","password":"Driver@123"}` | HTTP 401 / `40102` |
| DRIVER-AUTH-PWD-E02 | 密码长度不足 | `{"phone":"13800000000","password":"123"}` | HTTP 401 / `40102` |
| DRIVER-AUTH-PWD-E03 | 账号被冻结或注销 | 正确账号密码但状态不可用 | HTTP 403 / `40301` |
| DRIVER-AUTH-PWD-E04 | driversvc 不可用 | gRPC 连接失败 | HTTP 502 / `50001` |

## 6. 处理链路

`api/driver -> AuthLogic.LoginByPassword -> driversvc.Login`。账号存在性、状态和密码由 `driversvc` 最终校验。
