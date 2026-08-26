# 司机资质查询接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `GET` |
| 请求路径 | `/api/driver/v1/drivers/certification` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `driversvc.GetCertification` |

## 2. 请求参数

无。司机 ID 从 JWT 获取。

## 3. 请求示例

```bash
curl http://127.0.0.1:8082/api/driver/v1/drivers/certification \
  -H "Authorization: Bearer $TOKEN"
```

## 4. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `certification` | object/null | 资质记录 |
| `found` | bool | 是否存在资质记录 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-CERT-GET-E01 | 未登录 | HTTP 401 |
| DRIVER-CERT-GET-E02 | 无资质记录 | HTTP 200，`found=false` |
| DRIVER-CERT-GET-E03 | driversvc 不可用 | HTTP 502 |

## 6. 处理链路

`api/driver -> CertificationLogic.GetCertification -> driversvc.GetCertification`。
