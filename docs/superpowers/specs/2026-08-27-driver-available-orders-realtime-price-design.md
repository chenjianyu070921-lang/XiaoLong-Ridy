# Driver Available Orders And Realtime Price Design

## Goal

Add two driver-facing capabilities:

- A temporary dispatch hall / available order list based on the driver's latest reported position.
- A realtime trip price query that reuses the existing pricesvc estimation logic while keeping pricing authority in ordersvc.

## Scope

- Keep driver identity from JWT. The client must not pass `driverId`.
- Keep location dependency on the existing driver location report path and Redis `driver:pos:<driver_id>` snapshot.
- Keep the existing `POST /api/driver/v1/orders/available` route, but change its behavior from global wait-accept list to location-based temporary dispatch hall.
- Add `POST /api/driver/v1/orders/realtime-price`.
- Add an ordersvc RPC for realtime trip pricing because ordersvc owns order status and driver ownership checks.

## Temporary Dispatch Hall Design

`api/driver` reads the current driver's latest Redis position from `driver:pos:<driver_id>` and verifies the driver is online through `driver:online`. If the driver is offline or has no position snapshot, the endpoint returns an empty page.

When the driver is online and has a position, `api/driver` calls `ordersvc.ListOrders(status=WAIT_ACCEPT)` to fetch wait-accept orders, calls `ordersvc.GetOrder` for each listed order to obtain pickup coordinates, computes distance from the driver to the pickup point in memory, filters by radius, sorts nearest first, and applies pagination to the filtered list.

This is an explicit technical debt item. The business target is directed dispatch. After dispatchsvc exposes a query over `dispatch_record(status=1)` for the current driver, this implementation should be replaced. The current whole-list fetch plus in-memory filter only fits training/demo data volume.

## Realtime Price Design

`api/driver` authenticates the driver and forwards `orderId`, `distanceM`, and `durationS` to ordersvc. `ordersvc` loads the order, verifies it belongs to the current driver, verifies status is `ON_TRIP`, then calls `pricesvc.EstimatePrice` using the order's `userId`, `cityCode`, `carType`, and the request distance/duration.

The response returns the total and price detail from pricesvc. It does not persist price data, change order status, or create payment. `FinishTrip` remains the only final settlement path.

## Errors

- Invalid driver id, order id, distance, duration, page, or page size returns the existing invalid-parameter error.
- Available orders returns an empty list when Redis is unavailable, the driver is offline, or no position snapshot exists.
- Realtime price returns forbidden when the order does not belong to the current driver.
- Realtime price returns order-status-not-allowed when the order is not `ON_TRIP`.
- pricesvc errors are returned to the caller for realtime price so the app can retry.

## Testing

- Add api/driver logic tests proving available orders require online position, filter by distance, and sort nearest first.
- Add ordersvc logic tests proving realtime price checks driver ownership, checks `ON_TRIP`, forwards distance/duration/city/car type to pricesvc, and maps price detail.
- Add api/driver tests proving realtime price forwards the authenticated driver id to ordersvc.
- Run scoped Go tests for `api/driver` and `rpc/ordersvc`.
