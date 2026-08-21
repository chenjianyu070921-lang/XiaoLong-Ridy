# RejectCertification RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `RejectCertification(AuditCertificationRequest) returns (CommonResponse)` |
| 对应 HTTP | 无司机端 HTTP 路由，后台审核链路使用 |
| 当前状态 | 已实现 |
| 业务逻辑 | `RejectCertificationLogic.RejectCertification` |

## 2. 请求字段

同 [ApproveCertification](23-approve-certification.md)，其中 `remark` 必填，用于记录驳回原因。

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `message` | string | 成功提示 |

## 4. API 对齐

该 RPC 不属于司机端 HTTP API；由后台审核链路调用，只更新资质审核状态和备注。
