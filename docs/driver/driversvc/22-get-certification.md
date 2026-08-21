# GetCertification RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `GetCertification(GetCertificationRequest) returns (GetCertificationResponse)` |
| 对应 HTTP | `GET /api/driver/v1/drivers/certification` |
| 当前状态 | 已实现 |
| 业务逻辑 | `GetCertificationLogic.GetCertification` |

## 2. 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `driver_id` | int64 | 是 | 司机 ID，由 HTTP JWT 注入 |

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `certification` | CertificationInfo | 资质详情，无记录时为空 |
| `found` | bool | 是否查到记录 |

## 4. 处理链路

`api/driver.GetCertification -> driversvc.GetCertification`。HTTP 文档见 [../15-certification-get.md](../15-certification-get.md)。
