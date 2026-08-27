# Driver Listen Preference Design

## Goal

Add driver listen preferences for realtime and reservation orders. Drivers can configure which order types they accept, the setting is persisted, and dispatch only selects drivers whose preference matches the order type.

## Scope

This design is limited to driver-side business and dispatch candidate filtering:

- `api/driver` exposes HTTP request/response fields and routes for the driver app.
- `rpc/driversvc` owns the persisted driver preference and synchronizes online drivers to dispatch Redis state.
- `rpc/dispatchsvc` filters candidate drivers by the order type carried in dispatch requests.
- SQL migrations add a small preference table and order type fields needed by dispatch.

No framework, generated base plumbing, or unrelated non-driver business behavior is changed except generated files produced from edited proto definitions.

## Order Type Model

Use an explicit order type instead of reusing `car_type`.

- `ORDER_TYPE_REALTIME = 1`
- `ORDER_TYPE_RESERVATION = 2`

Default behavior is backward-compatible: when the caller does not provide an order type, dispatch treats it as realtime. Existing drivers without saved preference accept both realtime and reservation orders.

## Preference Model

Persist preferences in `driver_listen_preference`:

- `driver_id BIGINT UNSIGNED PRIMARY KEY`
- `accept_realtime TINYINT(1) NOT NULL DEFAULT 1`
- `accept_reservation TINYINT(1) NOT NULL DEFAULT 1`
- `created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP`
- `updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP`

Validation requires at least one of `acceptRealtime` or `acceptReservation` to be true.

## API And RPC

Driver HTTP:

- Add `POST /api/driver/v1/drivers/listen-preference` to update the current driver's preference.
- Add `GET /api/driver/v1/drivers/listen-preference` to read the current driver's preference.
- Extend `POST /api/driver/v1/drivers/online` to optionally carry `acceptRealtime` and `acceptReservation`; omitted values use the saved preference/default both.

Driver RPC:

- Add `SetDriverListenPreference(SetDriverListenPreferenceRequest) returns (DriverListenPreferenceResponse)`.
- Add `GetDriverListenPreference(GetDriverListenPreferenceRequest) returns (DriverListenPreferenceResponse)`.
- Extend `SetDriverOnlineRequest` with optional preference booleans.

Dispatch RPC:

- Extend `DispatchOrderRequest` with `order_type`.
- Dispatch engine receives the order type and filters Redis GEO candidates by each driver's preference.

## Redis Synchronization

When a driver goes online, `driversvc` writes existing online state and also writes preference membership:

- `driver:pref:realtime` contains drivers accepting realtime orders.
- `driver:pref:reservation` contains drivers accepting reservation orders.

When a driver goes offline, `driversvc` removes the driver from both preference sets. Updating preference while online refreshes the two sets. Dispatch filtering intersects candidates with the set matching `order_type`.

Fail-open rule: if Redis is unavailable in driversvc, persistence still succeeds but online synchronization returns the Redis error only during online/update paths that already depend on Redis. If dispatchsvc has no preference checker configured, it preserves existing behavior and only applies online/busy filtering.

## Proto Generation

Fix the previously wrong root-relative command by using the service-local script:

```powershell
powershell -ExecutionPolicy Bypass -File rpc/driversvc/scripts/regenerate_proto.ps1
```

The script changes directory to `rpc/driversvc` and runs:

```powershell
goctl rpc protoc proto/driversvc.proto --go_out=proto --go-grpc_out=proto --zrpc_out=. --style=go_zero
```

For dispatchsvc, use the same service-local pattern if generated wrappers are needed. Do not run the old root command with `rpc/driversvc/proto/driversvc.proto` plus output paths rooted again under `rpc/driversvc`.

## Testing

- API tests cover updating/getting preferences and online preference forwarding.
- driversvc logic tests cover default both, validation, persistence, and Redis preference set sync.
- dispatchsvc engine tests cover filtering realtime vs reservation candidates.
- Generated proto code is verified by compiling `api/driver/...`, `rpc/driversvc/...`, and `rpc/dispatchsvc/...`.