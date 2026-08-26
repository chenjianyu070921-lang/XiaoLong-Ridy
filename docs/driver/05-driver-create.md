# 创建司机接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/drivers` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `driversvc.CreateDriver` |

## 2. 请求参数

同 [司机自注册接口](04-driver-register.md)。当前 HTTP 路由要求有效司机 JWT；后台角色隔离由上层权限系统负责。

## 3. 请求示例

```bash
curl -X POST http://127.0.0.1:8082/api/driver/v1/drivers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800000000","password":"Driver@123","realName":"张三","idCardNo":"110101199001011234","driverLicenseNo":"DL10000001"}'
```

## 4. 响应字段

同 [司机自注册接口](04-driver-register.md)。

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-CREATE-E01 | 未携带 Token | HTTP 401 / `40102` |
| DRIVER-CREATE-E02 | 参数格式错误 | HTTP 400 |
| DRIVER-CREATE-E03 | driversvc 不可用 | HTTP 502 / `50001` |

## 6. 处理链路

`api/driver -> RequireAuth -> DriverLogic.CreateDriver -> driversvc.CreateDriver`。
