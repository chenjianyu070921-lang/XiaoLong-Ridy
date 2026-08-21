# ListOrderStatusLogs 查询订单状态流水

## 接口概述
- **服务**：ordersvc（Order 服务）
- **方法**：`ListOrderStatusLogs(ListOrderStatusLogsRequest) returns (ListOrderStatusLogsResponse)`
- **功能**：分页查询指定订单的状态变更流水（含创建、接单、行程、取消、支付等各节点记录）。
- **调用方**：乘客端（C）、司机端（B）、管理后台（D）

## 请求参数（ListOrderStatusLogsRequest）
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_id | int64 | 是 | 订单 ID，必须 > 0 |
| page | int32 | 否 | 页码，默认 1 |
| page_size | int32 | 否 | 每页条数，默认 20，上限 100 |

## 响应参数（ListOrderStatusLogsResponse）
| 字段 | 类型 | 说明 |
|------|------|------|
| list | repeated OrderStatusLog | 状态流水列表 |
| total | int64 | 总记录数 |
| page | int32 | 实际页码 |
| page_size | int32 | 实际每页条数 |

### OrderStatusLog 结构
| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 流水 ID |
| order_id | int64 | 订单 ID |
| from_status | int32 | 变更前状态（OrderStatus 枚举值） |
| to_status | int32 | 变更后状态（OrderStatus 枚举值） |
| operator_type | string | 操作方：user / driver / system / admin |
| operator_id | int64 | 操作人 ID（system 时为 0） |
| remark | string | 备注说明 |
| created_at | int64 | 记录时间，Unix 秒 |

## 分页规则
- page <= 0 时归一化为 1；page_size <= 0 时归一化为 20；page_size > 100 时截断为 100。

## 错误码
| 错误 | 触发条件 |
|------|----------|
| invalid order params | order_id<=0 |

## 调用示例
**请求**
```json
{
  "order_id": 50001,
  "page": 1,
  "page_size": 20
}
```
**响应**
```json
{
  "list": [
    {
      "id": 9001,
      "order_id": 50001,
      "from_status": 0,
      "to_status": 1,
      "operator_type": "user",
      "operator_id": 1001,
      "remark": "创建订单",
      "created_at": 1789600000
    },
    {
      "id": 9002,
      "order_id": 50001,
      "from_status": 1,
      "to_status": 2,
      "operator_type": "driver",
      "operator_id": 2001,
      "remark": "司机接单",
      "created_at": 1789600200
    }
  ],
  "total": 2,
  "page": 1,
  "page_size": 20
}
```
