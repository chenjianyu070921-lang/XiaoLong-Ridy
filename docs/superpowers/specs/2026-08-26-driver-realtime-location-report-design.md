# Driver Realtime Location Report Design

## Goal

Complete the driver realtime location reporting path so a driver location report updates the latest location table and Redis location data for dispatch, trip navigation, and a future dispatch hall.

## Scope

- Keep the existing HTTP endpoint: `POST /api/driver/v1/drivers/location/report`.
- Keep driver identity from JWT; clients must not pass `driverId`.
- Keep `driversvc.ReportLocation` as the reusable core path so other interfaces can call the same behavior without duplicating HTTP logic.
- Do not change bottom-layer/framework code.
- Do not change non-driver business behavior.

## Design

`api/driver` validates the request, takes the driver id from JWT, and calls `driversvc.ReportLocation`. When `locationsvc` is configured, it continues to dual-write heading/speed/order location data there. When `orderId` is present, the API continues writing trip trajectory points through its existing repository.

`driversvc.ReportLocation` remains the source of truth for online heartbeat and latest driver location used by dispatch. It validates driver id, device id, and coordinates, runs the existing `OnlineStore.Heartbeat` device binding check, writes `driver_location` through the existing repository when the device is not kicked, updates `driver.online_status`, and syncs Redis.

Redis sync must include:

- GEO position under `driver:geo:<city>` for dispatch candidate lookup.
- Online membership under `driver:online`.
- Latest position snapshot under `driver:pos:<driver_id>` for direct lookup by future dispatch hall/navigation paths.

## Reusable Interface

The reusable unit is in driversvc logic, not in the HTTP handler:

- `syncDispatchDriverOnline(ctx, svcCtx, driverID, longitude, latitude)` updates Redis online/GEO and latest position snapshot.
- `syncDispatchDriverOffline(ctx, svcCtx, driverID)` removes online/GEO/position snapshot.

Other driver-side RPC logic can call these helpers directly without going through `api/driver`.

## Errors

- Invalid driver id, empty device id, or invalid coordinates return existing driversvc validation errors.
- Device mismatch returns `kicked=true` and does not update `driver_location` or dispatch Redis position data.
- Redis or database failures are returned to the caller; the caller can retry the heartbeat.

## Testing

- Add driversvc logic tests proving online sync writes `driver:pos:<driver_id>` and removes it on offline.
- Add driversvc logic tests proving `ReportLocation` preserves saved listen preference when refreshing Redis.
- Keep existing API tests proving current driver id and location are forwarded.
- Run scoped tests for `api/driver`, `rpc/driversvc`, and `rpc/dispatchsvc` because dispatch consumes the Redis keys.