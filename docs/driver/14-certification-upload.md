# 司机资质上传接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/drivers/certification/upload` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `driversvc.UploadCertification` |

## 2. 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `vehicleId` | int64 | 是 | 车辆 ID |
| `idCardFront` | string | 否 | 身份证正面 Base64 |
| `idCardBack` | string | 否 | 身份证反面 Base64 |
| `driverLicense` | string | 否 | 驾驶证 Base64 |
| `vehicleLicense` | string | 否 | 行驶证 Base64 |

至少上传一张图片。图片可带 `data:image/...;base64,` 前缀，也可只传纯 Base64。

## 3. 请求示例

```json
{
  "vehicleId": 1001,
  "idCardFront": "<base64>",
  "idCardBack": "<base64>",
  "driverLicense": "<base64>",
  "vehicleLicense": "<base64>"
}
```

## 4. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 资质记录 ID |
| `certification` | object | 资质详情 |
| `certification.auditStatus` | int | 审核状态 |
| `certification.auditRemark` | string | 审核备注 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-CERT-UP-E01 | 未登录 | HTTP 401 |
| DRIVER-CERT-UP-E02 | 全部图片为空 | HTTP 400 |
| DRIVER-CERT-UP-E03 | Base64 非法 | HTTP 400 或下游错误 |
| DRIVER-CERT-UP-E04 | 图片超过大小上限 | HTTP 400 或下游错误 |

## 6. 处理链路

`api/driver -> CertificationLogic.UploadCertification -> driversvc.UploadCertification -> MinIO/driver_certification`。
