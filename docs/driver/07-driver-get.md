# 查询司机详情接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `GET` |
| 请求路径 | `/api/driver/v1/drivers/get` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `driversvc.GetDriver` |

## 2. 查询参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 司机 ID |

## 3. 请求示例

```bash
curl "http://127.0.0.1:8082/api/driver/v1/drivers/get?id=25" \
  -H "Authorization: Bearer $TOKEN"
```

## 4. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `driver.id` | int64 | 司机 ID |
| `driver.phone` | string | 脱敏手机号 |
| `driver.idCardNo` | string | 脱敏身份证号 |
| `driver.status` | string | 司机状态 |
| `driver.onlineStatus` | int | 在线状态 |
| `driver.createdAt` | int64 | 创建时间 |
| `driver.updatedAt` | int64 | 更新时间 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-GET-E01 | 缺少 `id` | HTTP 400 |
| DRIVER-GET-E02 | `id` 非数字或小于等于 0 | HTTP 400 |
| DRIVER-GET-E03 | 未登录 | HTTP 401 |

## 6. 处理链路

`api/driver -> DriverLogic.GetDriver -> driversvc.GetDriver`。响应不返回密码哈希。
