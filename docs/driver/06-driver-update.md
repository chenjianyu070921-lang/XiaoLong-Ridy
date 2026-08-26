# 更新司机接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/drivers/update` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `driversvc.UpdateDriver` |

## 2. 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 司机 ID |
| `phone` | string | 否 | 手机号 |
| `password` | string | 否 | 明文新密码，API 层转 bcrypt |
| `realName` | string | 否 | 真实姓名 |
| `idCardNo` | string | 否 | 身份证号 |
| `driverLicenseNo` | string | 否 | 驾驶证号 |
| `avatarUrl` | string | 否 | 头像地址 |
| `status` | string | 否 | `DRIVER_STATUS_PENDING/NORMAL/FROZEN/CANCELLED` |

## 3. 请求示例

```json
{"id":25,"password":"NewDriver@123","realName":"张三","status":"DRIVER_STATUS_NORMAL"}
```

## 4. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 司机 ID |
| `status` | string | 更新后的状态 |
| `updatedAt` | int64 | 更新时间，Unix 秒 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-UPDATE-E01 | `id<=0` | HTTP 400 |
| DRIVER-UPDATE-E02 | 状态枚举非法 | HTTP 400 |
| DRIVER-UPDATE-E03 | 未登录 | HTTP 401 |

## 6. 处理链路

`api/driver -> DriverLogic.UpdateDriver -> driversvc.UpdateDriver`。RPC optional 字段为 `nil` 时不更新。
