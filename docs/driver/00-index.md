# 司机端专属接口文档目录

> 版本: v1.0  
> 更新时间: 2026-08-21  
> 服务地址: `http://127.0.0.1:8082`  
> 请求前缀: `/api/driver/v1`

## 通用说明

| 项 | 说明 |
| --- | --- |
| Content-Type | `application/json; charset=utf-8` |
| 认证方式 | 需登录接口携带 `Authorization: Bearer <JWT>` |
| 统一成功响应 | `{"code":0,"message":"success","data":{},"timestamp":0,"traceId":""}` |
| 参数错误 | HTTP `400`，业务码通常为 `40000` 或 `50000` |
| 未登录/凭证错误 | HTTP `401`，业务码 `40102` |
| 账号冻结/注销 | HTTP `403`，业务码 `40301` |
| 下游不可用 | HTTP `502`，业务码 `50001` |

## 接口清单

| 模块 | 接口 | 文档 |
| --- | --- | --- |
| 认证 | `POST /auth/send-sms-code` | [01-auth-send-sms-code.md](01-auth-send-sms-code.md) |
| 认证 | `POST /auth/login-by-password` | [02-auth-login-by-password.md](02-auth-login-by-password.md) |
| 认证 | `POST /auth/login-by-sms` | [03-auth-login-by-sms.md](03-auth-login-by-sms.md) |
| 司机账号 | `POST /drivers/register` | [04-driver-register.md](04-driver-register.md) |
| 司机账号 | `POST /drivers` | [05-driver-create.md](05-driver-create.md) |
| 司机账号 | `POST /drivers/update` | [06-driver-update.md](06-driver-update.md) |
| 司机账号 | `GET /drivers/get` | [07-driver-get.md](07-driver-get.md) |
| 司机账号 | `POST /drivers/delete` | [08-driver-delete.md](08-driver-delete.md) |
| 司机账号 | `GET /drivers/ai-score` | [09-driver-ai-score.md](09-driver-ai-score.md) |
| 在线状态 | `POST /drivers/online` | [10-driver-online.md](10-driver-online.md) |
| 在线状态 | `POST /drivers/offline` | [11-driver-offline.md](11-driver-offline.md) |
| 在线状态 | `POST /drivers/heartbeat` | [12-driver-heartbeat.md](12-driver-heartbeat.md) |
| 在线状态 | `POST /drivers/location/report` | [13-driver-location-report.md](13-driver-location-report.md) |
| 资质 | `POST /drivers/certification/upload` | [14-certification-upload.md](14-certification-upload.md) |
| 资质 | `GET /drivers/certification` | [15-certification-get.md](15-certification-get.md) |
| 订单行程 | `POST /orders/accept` | [16-order-accept.md](16-order-accept.md) |
| 订单行程 | `POST /orders/reject` | [17-order-reject.md](17-order-reject.md) |
| 订单行程 | `POST /orders/dispatches` | [18-order-dispatches.md](18-order-dispatches.md) |
| 订单行程 | `POST /orders/start-trip` | [19-order-start-trip.md](19-order-start-trip.md) |
| 订单行程 | `POST /orders/confirm-arrive` | [20-order-confirm-arrive.md](20-order-confirm-arrive.md) |
| 订单行程 | `POST /orders/finish-trip` | [21-order-finish-trip.md](21-order-finish-trip.md) |
| 智能助手 | `POST /agent/chat` | [22-agent-chat.md](22-agent-chat.md) |

## 业务边界

- 司机身份统一从 JWT 获取，订单/位置/资质接口不得信任客户端传入的 `driverId`。
- 订单状态推进由 `ordersvc` 负责最终校验；派单归属与拒单状态由 `dispatchsvc` 负责最终校验。
- 司机账号、在线状态、位置、资质和推荐分由 `driversvc` 承载。

## driversvc RPC 文档

`driversvc` RPC 专属文档放在 [driversvc/00-index.md](driversvc/00-index.md)。其中每个 RPC 单独成文，并标明对应的司机端 HTTP API；没有司机端 HTTP 路由的 RPC 会标为内部/后台链路。
