# GetDriverByPhone RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `GetDriverByPhone(GetDriverByPhoneRequest) returns (GetDriverByPhoneResponse)` |
| 对应 HTTP | 无直接路由，登录逻辑内部使用 |
| 当前状态 | 已实现 |
| 业务逻辑 | `GetDriverByPhoneLogic.GetDriverByPhone` |

## 2. 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `phone` | string | 是 | 手机号 |

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `driver` | Driver | 司机详情，登录链路会读取账号状态和密码哈希 |

## 4. 处理链路

`api/driver AuthLogic -> driversvc.GetDriverByPhone`。该 RPC 不对外暴露为司机端 HTTP 路由。
