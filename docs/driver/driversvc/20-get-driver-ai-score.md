# GetDriverAiScore RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `GetDriverAiScore(GetDriverAiScoreRequest) returns (GetDriverAiScoreResponse)` |
| 对应 HTTP | `GET /api/driver/v1/drivers/ai-score?id=` |
| 当前状态 | 已实现 |
| 业务逻辑 | `GetDriverAiScoreLogic.GetDriverAiScore` |

## 2. 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `driver_id` | int64 | 是 | 司机 ID |

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `driver_id` | int64 | 司机 ID |
| `ai_score` | double | 综合推荐分 |
| `level` | int32 | 推荐等级 |
| `factors` | repeated AiScoreFactor | 影响因子 |
| `degraded` | bool | 是否降级 |
| `degrade_reason` | string | 降级原因 |

## 4. 处理链路

`api/driver.GetDriverAiScore -> driversvc.GetDriverAiScore`。HTTP 文档见 [../09-driver-ai-score.md](../09-driver-ai-score.md)。
