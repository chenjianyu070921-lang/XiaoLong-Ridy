# RegisterDriver RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `RegisterDriver(CreateDriverRequest) returns (CreateDriverResponse)` |
| 对应 HTTP | `POST /api/driver/v1/drivers/register` |
| 当前状态 | 已实现 |
| 业务逻辑 | `CreateDriverLogic.CreateDriver` |

## 2. 请求/响应字段

同 [CreateDriver](01-create-driver.md)。HTTP API 仍传明文 `password`，API 层生成 `password_hash` 后调用本 RPC。

## 3. 处理链路

`api/driver.RegisterDriver -> bcrypt -> driversvc.RegisterDriver -> driver`。司机端 HTTP 文档见 [../04-driver-register.md](../04-driver-register.md)。
