# DeleteDriver RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `DeleteDriver(DeleteDriverRequest) returns (DeleteDriverResponse)` |
| 对应 HTTP | `POST /api/driver/v1/drivers/delete?id=` |
| 当前状态 | 已实现 |
| 业务逻辑 | `DeleteDriverLogic.DeleteDriver` |

## 2. 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 司机 ID |

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 司机 ID |
| `success` | bool | 是否删除成功 |

## 4. 处理链路

`api/driver.DeleteDriver -> driversvc.DeleteDriver -> driver`。当前为软删除。HTTP 文档见 [../08-driver-delete.md](../08-driver-delete.md)。
