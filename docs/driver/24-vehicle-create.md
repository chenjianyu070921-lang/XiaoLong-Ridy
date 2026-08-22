# 24. Submit Vehicle

## Endpoint

`POST /api/driver/v1/vehicles`

Requires `Authorization: Bearer <JWT>`.

## Request

```json
{
  "plateNo": "粤B12345",
  "brand": "BYD",
  "model": "Han",
  "color": "black",
  "vehicleType": 1,
  "registrationDate": 1700000000,
  "insuranceNo": "INS-001",
  "insuranceExpireAt": 1800000000
}
```

`driverId` is always taken from the JWT and must not be supplied by the client.

## Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 77,
    "status": "VEHICLE_STATUS_PENDING",
    "createdAt": 1700000000
  },
  "timestamp": 1700000000,
  "traceId": "trace_xxx"
}
```
