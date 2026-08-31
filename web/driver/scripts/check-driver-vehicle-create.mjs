import { strict as assert } from 'node:assert'
import { resolveCreatedVehicle, resolveDriverVehicleId } from '../src/utils/vehicle.js'

const payload = {
  plateNo: '粤B12345',
  brand: 'BYD',
  model: 'Han',
  color: 'black',
  vehicleType: 1,
  registrationDate: 1700000000,
  insuranceNo: 'INS-1',
  insuranceExpireAt: 1800000000
}

const created = {
  id: 44,
  status: 'VEHICLE_STATUS_PENDING',
  createdAt: 1788180000
}

let requestedId = 0
const vehicle = await resolveCreatedVehicle(payload, created, async (id) => {
  requestedId = id
  return {
    vehicle: {
      ...payload,
      id,
      driverId: 11,
      status: 'VEHICLE_STATUS_PENDING',
      createdAt: 1788180000,
      updatedAt: 1788180000
    }
  }
})

assert.equal(requestedId, 44, 'vehicle detail must be loaded with the created id')
assert.equal(vehicle.id, 44, 'resolved vehicle must keep the created id')
assert.equal(vehicle.plateNo, '粤B12345', 'resolved vehicle must include the plate number for the UI')
assert.equal(vehicle.brand, 'BYD', 'resolved vehicle must include brand for the UI')
assert.equal(vehicle.model, 'Han', 'resolved vehicle must include model for the UI')

const fallback = await resolveCreatedVehicle(payload, created, async () => {
  throw new Error('detail unavailable')
})

assert.equal(fallback.id, 44, 'fallback vehicle must keep the created id')
assert.equal(fallback.plateNo, '粤B12345', 'fallback vehicle must keep submitted plate number')
assert.equal(fallback.brand, 'BYD', 'fallback vehicle must keep submitted brand')
assert.equal(fallback.status, 'VEHICLE_STATUS_PENDING', 'fallback vehicle must keep backend status')

assert.equal(resolveDriverVehicleId({ vehicleId: 88 }, null), 88, 'driver profile vehicleId must be usable without local cache')
assert.equal(resolveDriverVehicleId({ vehicleId: 0 }, { id: 44 }), 44, 'cached vehicle id must be used when profile has no vehicleId')
assert.equal(resolveDriverVehicleId({}, null), 0, 'missing vehicle id must resolve to zero')

console.log('driver vehicle create checks passed')
