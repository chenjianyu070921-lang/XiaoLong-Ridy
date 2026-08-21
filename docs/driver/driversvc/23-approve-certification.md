# ApproveCertification RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `ApproveCertification(AuditCertificationRequest) returns (CommonResponse)` |
| 对应 HTTP | 无司机端 HTTP 路由，后台审核链路使用 |
| 当前状态 | 已实现 |
| 业务逻辑 | `ApproveCertificationLogic.ApproveCertification` |

## 2. 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `certification_id` | int64 | 是 | 资质记录 ID |
| `remark` | string | 否 | 审核备注 |
| `operator_id` | int64 | 是 | 操作人 ID |
| `ip` | string | 否 | 操作来源 IP |

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `message` | string | 成功提示 |

## 4. API 对齐

该 RPC 不属于司机端 HTTP API；由后台审核链路调用，并同步推进资质、司机、车辆状态。
