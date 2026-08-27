# Driver Listen Preference

## Endpoints

- `POST /api/driver/v1/drivers/listen-preference`
- `GET /api/driver/v1/drivers/listen-preference`

Both endpoints require `Authorization: Bearer <JWT>` and use the current driver from the token.

## POST Body

```json
{
  "acceptRealtime": true,
  "acceptReservation": true
}
```

At least one of `acceptRealtime` and `acceptReservation` must be `true`.

## Response Data

```json
{
  "driverId": 25,
  "acceptRealtime": true,
  "acceptReservation": true,
  "updatedAt": 1787731200
}
```

## Online Shortcut

`POST /api/driver/v1/drivers/online` also accepts optional `acceptRealtime` and `acceptReservation` fields. When present, driversvc stores the preference and syncs Redis dispatch preference sets while marking the driver online.