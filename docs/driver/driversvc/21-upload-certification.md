# UploadCertification RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `UploadCertification(UploadCertificationRequest) returns (UploadCertificationResponse)` |
| 对应 HTTP | `POST /api/driver/v1/drivers/certification/upload` |
| 当前状态 | 已实现 |
| 业务逻辑 | `UploadCertificationLogic.UploadCertification` |

## 2. 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `driver_id` | int64 | 是 | 司机 ID，由 HTTP JWT 注入 |
| `vehicle_id` | int64 | 是 | 车辆 ID |
| `id_card_front` | string | 否 | 身份证正面 Base64 |
| `id_card_back` | string | 否 | 身份证反面 Base64 |
| `driver_license` | string | 否 | 驾驶证 Base64 |
| `vehicle_license` | string | 否 | 行驶证 Base64 |

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 资质记录 ID |
| `certification` | CertificationInfo | 资质详情 |

## 4. 处理链路

`api/driver.UploadCertification -> driversvc.UploadCertification -> MinIO/driver_certification`。HTTP 文档见 [../14-certification-upload.md](../14-certification-upload.md)。
