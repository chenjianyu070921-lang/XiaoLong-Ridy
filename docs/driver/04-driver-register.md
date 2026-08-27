# 司机自注册接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/drivers/register` |
| 是否登录 | 否 |
| 当前状态 | 已实现 |
| 下游 RPC | `driversvc.RegisterDriver` |

## 2. 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `phone` | string | 是 | 中国大陆手机号 |
| `password` | string | 是 | 明文密码，服务端生成 bcrypt 哈希 |
| `realName` | string | 是 | 真实姓名 |
| `idCardNo` | string | 是 | 18 位身份证号 |
| `driverLicenseNo` | string | 是 | 驾驶证号 |
| `avatarUrl` | string | 否 | 头像地址 |

## 3. 请求示例

```json
{
  "phone": "13800000000",
  "password": "Driver@123",
  "realName": "张三",
  "idCardNo": "110101199001011234",
  "driverLicenseNo": "DL10000001",
  "avatarUrl": ""
}
```

## 4. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 新司机 ID |
| `status` | string | 初始状态，通常为 `DRIVER_STATUS_PENDING` |
| `createdAt` | int64 | 创建时间，Unix 秒 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-REGISTER-E01 | 手机号格式错误 | HTTP 400 |
| DRIVER-REGISTER-E02 | 密码长度不合法 | HTTP 400 |
| DRIVER-REGISTER-E03 | 身份证号格式错误 | HTTP 400 |
| DRIVER-REGISTER-E04 | 手机号或驾驶证号重复 | HTTP 400 或 gRPC 错误透传 |

## 6. 处理链路

`api/driver -> DriverLogic.RegisterDriver -> bcrypt -> driversvc.RegisterDriver -> driver`。
