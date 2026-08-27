# Driver Location Report

## Endpoint

`POST /api/driver/v1/drivers/location/report`

Requires `Authorization: Bearer <JWT>`. The driver id always comes from JWT.

## Body

```json
{
  "deviceId": "dev-001",
  "longitude": 116.397,
  "latitude": 39.908,
  "heading": 90,
  "speedKmh": 36.5,
  "orderId": 1001
}
```

`deviceId`, `longitude`, and `latitude` are required. `heading`, `speedKmh`, and `orderId` are optional extension fields for locationsvc and trip trajectory.

## Behavior

The reusable core is `driversvc.ReportLocation`, so other driver-side entry points can call the same path. It refreshes online heartbeat/device binding, upserts `driver_location`, updates `driver.online_status`, writes Redis GEO `driver:geo:<city>`, writes online set `driver:online`, and writes latest snapshot `driver:pos:<driver_id>`.

A kicked old device returns `kicked=true` and does not update location storage. Location heartbeats preserve the driver's saved listen preference.