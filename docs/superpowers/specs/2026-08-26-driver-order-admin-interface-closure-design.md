# Driver Order/Admin Interface Closure Design

## Goal

Close the driver-facing integration gaps for admin driver records, dispatch retrieval, and driver order detail access without changing framework or non-driver business logic.

## Scope

- Add real admin HTTP list/detail interfaces for driver base records so the admin web `/drivers` page no longer uses an empty placeholder.
- Keep driver certification upload/review interfaces unchanged; they already use the `driver_certification` workflow.
- Fix driver order detail authorization so the current driver can view owned current/history orders, and can also view unassigned wait-accept orders from the dispatch/available order flows.
- Keep another driver's accepted/in-trip/history order protected.
- Keep the driver web order and dispatch calls on `/api/driver/v1`, with normalized order IDs and action availability.

## API Contract

Admin:

- `GET /admin/v1/drivers?page=1&page_size=20&keyword=&status=0`
- `GET /admin/v1/drivers/{id}`

Driver:

- `POST /api/driver/v1/orders/detail` with body `{"orderId":1001}`
- `POST /api/driver/v1/orders/dispatches` with body `{"page":1,"pageSize":8,"status":0}`
- `POST /api/driver/v1/orders/list` with body `{"page":1,"pageSize":8,"status":0}`
- `POST /api/driver/v1/orders/available` with body `{"page":1,"pageSize":8}`

## Data Flow

- Admin HTTP calls `adminsvc` for driver base records through new admin RPC methods backed by driver tables.
- Driver order detail continues to read `ordersvc.GetOrder`.
- Authorization is enforced in the driver API layer using the token-derived driver ID.
- Dispatch records continue to read dispatchsvc records and enrich each record with `ordersvc.GetOrder`.

## Error Handling

- Invalid IDs return existing parameter errors.
- Driver order detail returns forbidden when an order belongs to another driver.
- Unassigned wait-accept orders are allowed so drivers can inspect orders before accepting.

## Tests

- Add driver order logic tests for wait-accept detail access and other-driver protection.
- Add admin route tests for `/admin/v1/drivers` and `/admin/v1/drivers/{id}`.
- Run impacted Go packages.
