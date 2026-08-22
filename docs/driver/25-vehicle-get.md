# 25. Get Vehicle

## Endpoint

`GET /api/driver/v1/vehicles/get?id=77`

Requires `Authorization: Bearer <JWT>`.

The API only returns vehicles that belong to the current driver in the JWT.

## Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "vehicle": {
      "id": 77,
      "driverId": 25,
      "plateNo": "粤B12345",
      "brand": "BYD",
      "model": "Han",
      "color": "black",
      "vehicleType": 1,
      "registrationDate": 1700000000,
      "insuranceNo": "INS-001",
      "insuranceExpireAt": 1800000000,
      "status": "VEHICLE_STATUS_PENDING",
      "createdAt": 1700000000,
      "updatedAt": 1700000000
    }
  },
  "timestamp": 1700000000,
  "traceId": "trace_xxx"
}
```
