# 发送登录验证码接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/auth/send-sms-code` |
| 是否登录 | 否 |
| 当前状态 | 已实现 |
| 业务逻辑 | `AuthLogic.SendSMSCode` |

## 2. 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `phone` | string | 是 | 中国大陆手机号，格式 `1[3-9]xxxxxxxxx` |

## 3. 请求示例

```bash
curl -X POST http://127.0.0.1:8082/api/driver/v1/auth/send-sms-code \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800000000"}'
```

## 4. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `success` | bool | 是否发送成功 |
| `expireIn` | int | 验证码有效期，单位秒 |

## 5. 异常用例

| 用例编号 | 场景 | 请求体 | 预期 |
| --- | --- | --- | --- |
| DRIVER-AUTH-SMS-E01 | 手机号为空 | `{}` | HTTP 400 |
| DRIVER-AUTH-SMS-E02 | 手机号格式错误 | `{"phone":"abc"}` | HTTP 400 |
| DRIVER-AUTH-SMS-E03 | 请求方法错误 | GET | HTTP 405 |

## 6. 处理链路

`api/driver -> AuthLogic.SendSMSCode -> CodeCache`。联调阶段验证码写入服务端日志，不调用真实短信通道。
