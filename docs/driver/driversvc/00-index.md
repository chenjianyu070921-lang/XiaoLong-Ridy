# driversvc RPC 接口文档目录

> 版本: v1.0  
> 更新时间: 2026-08-21  
> Proto: `rpc/driversvc/proto/driversvc.proto`  
> 服务名: `driversvc.Driversvc`

## 通用说明

| 项 | 说明 |
| --- | --- |
| RPC 命名 | 与 `ordersvc` 一致，`service` 方法清单在 proto 前部集中声明 |
| 时间字段 | `created_at`、`updated_at`、`report_time`、`server_time` 均为 Unix 秒 |
| 司机身份 | 司机端 HTTP API 不接收客户端 `driverId`，统一由 JWT 注入到 RPC 请求 |
| 状态值 | `online_status`: 0离线、1在线、2行程中 |

## HTTP API 对齐表

| driversvc RPC | 司机端 HTTP API | 文档 |
| --- | --- | --- |
| `CreateDriver` | `POST /api/driver/v1/drivers` | [01-create-driver.md](01-create-driver.md) |
| `RegisterDriver` | `POST /api/driver/v1/drivers/register` | [02-register-driver.md](02-register-driver.md) |
| `UpdateDriver` | `POST /api/driver/v1/drivers/update` | [03-update-driver.md](03-update-driver.md) |
| `DeleteDriver` | `POST /api/driver/v1/drivers/delete` | [04-delete-driver.md](04-delete-driver.md) |
| `GetDriver` | `GET /api/driver/v1/drivers/get` | [05-get-driver.md](05-get-driver.md) |
| `GetDriverByPhone` | 登录逻辑内部使用 | [06-get-driver-by-phone.md](06-get-driver-by-phone.md) |
| `SetDriverOnline` | `POST /api/driver/v1/drivers/online` | [07-set-driver-online.md](07-set-driver-online.md) |
| `SetDriverOffline` | `POST /api/driver/v1/drivers/offline` | [08-set-driver-offline.md](08-set-driver-offline.md) |
| `ReportLocation` | `POST /api/driver/v1/drivers/location/report` | [09-report-location.md](09-report-location.md) |
| `SetDriverServiceStatus` | 订单行程逻辑内部使用 | [10-set-driver-service-status.md](10-set-driver-service-status.md) |
| `Heartbeat` | `POST /api/driver/v1/drivers/heartbeat` | [11-heartbeat.md](11-heartbeat.md) |
| `CreateVehicle` | 无司机端 HTTP 路由 | [12-create-vehicle.md](12-create-vehicle.md) |
| `UpdateVehicle` | 无司机端 HTTP 路由 | [13-update-vehicle.md](13-update-vehicle.md) |
| `DeleteVehicle` | 无司机端 HTTP 路由 | [14-delete-vehicle.md](14-delete-vehicle.md) |
| `GetVehicle` | 无司机端 HTTP 路由 | [15-get-vehicle.md](15-get-vehicle.md) |
| `ListDrivers` | 无司机端 HTTP 路由 | [16-list-drivers.md](16-list-drivers.md) |
| `Login` | `POST /api/driver/v1/auth/login-by-password` | [17-login.md](17-login.md) |
| `LoginBySMS` | `POST /api/driver/v1/auth/login-by-sms` | [18-login-by-sms.md](18-login-by-sms.md) |
| `ListNearbyDrivers` | 派单引擎内部使用 | [19-list-nearby-drivers.md](19-list-nearby-drivers.md) |
| `GetDriverAiScore` | `GET /api/driver/v1/drivers/ai-score` | [20-get-driver-ai-score.md](20-get-driver-ai-score.md) |
| `UploadCertification` | `POST /api/driver/v1/drivers/certification/upload` | [21-upload-certification.md](21-upload-certification.md) |
| `GetCertification` | `GET /api/driver/v1/drivers/certification` | [22-get-certification.md](22-get-certification.md) |
| `ApproveCertification` | 后台审核链路使用 | [23-approve-certification.md](23-approve-certification.md) |
| `RejectCertification` | 后台审核链路使用 | [24-reject-certification.md](24-reject-certification.md) |
