# 删除司机接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/drivers/delete` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `driversvc.DeleteDriver` |

## 2. 查询参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 司机 ID |

## 3. 请求示例

```bash
curl -X POST "http://127.0.0.1:8082/api/driver/v1/drivers/delete?id=25" \
  -H "Authorization: Bearer $TOKEN"
```

## 4. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 司机 ID |
| `success` | bool | 是否删除成功 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-DELETE-E01 | 缺少 `id` | HTTP 400 |
| DRIVER-DELETE-E02 | 未登录 | HTTP 401 |
| DRIVER-DELETE-E03 | driversvc 不可用 | HTTP 502 |

## 6. 处理链路

`api/driver -> DriverLogic.DeleteDriver -> driversvc.DeleteDriver`。当前为软删除司机账号。
