# 司机 AI 推荐分接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `GET` |
| 请求路径 | `/api/driver/v1/drivers/ai-score` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `driversvc.GetDriverAiScore` |

## 2. 查询参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 司机 ID |

## 3. 请求示例

```bash
curl "http://127.0.0.1:8082/api/driver/v1/drivers/ai-score?id=25" \
  -H "Authorization: Bearer $TOKEN"
```

## 4. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `driverId` | int64 | 司机 ID |
| `aiScore` | float64 | 综合推荐分 |
| `level` | int32 | 推荐等级 |
| `factors` | array | 影响因子列表 |
| `degraded` | bool | 是否降级 |
| `degradeReason` | string | 降级原因 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-AISCORE-E01 | 缺少 `id` | HTTP 400 |
| DRIVER-AISCORE-E02 | 未登录 | HTTP 401 |
| DRIVER-AISCORE-E03 | 下游推荐数据不可用 | `degraded=true` 或错误透传 |

## 6. 处理链路

`api/driver -> DriverLogic.GetDriverAiScore -> driversvc.GetDriverAiScore`。
