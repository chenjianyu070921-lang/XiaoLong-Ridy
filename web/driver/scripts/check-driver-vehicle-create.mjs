import { strict as assert } from 'node:assert'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
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

const root = resolve(import.meta.dirname, '..')
const useDriverAssets = readFileSync(resolve(root, 'src/composables/useDriverAssets.js'), 'utf8')
const driverFormat = readFileSync(resolve(root, 'src/utils/driver-format.js'), 'utf8')
const vehiclePanel = readFileSync(resolve(root, 'src/components/driver-home/DriverVehiclePanel.vue'), 'utf8')

assert.match(useDriverAssets, /await\s+driverStore\.refreshProfile\(\{\s*silentError:\s*true\s*\}\)/, 'vehicle page must refresh driver profile before loading vehicle id')
assert.match(useDriverAssets, /driverStore\.vehicleId[\s\S]*updateVehicle\(\{\s*\.\.\.payload,\s*id:\s*driverStore\.vehicleId\s*\}\)/, 'existing vehicle submit must update instead of creating a duplicate vehicle')
assert.match(driverFormat, /export function unixSecondsToDateInput/, 'vehicle API timestamps must be formatted for date inputs')
assert.match(useDriverAssets, /unixSecondsToDateInput\(driverStore\.vehicle\.registrationDate\)/, 'vehicle form must convert registration timestamp to date input value')
assert.match(useDriverAssets, /unixSecondsToDateInput\(driverStore\.vehicle\.insuranceExpireAt\)/, 'vehicle form must convert insurance timestamp to date input value')
assert.match(vehiclePanel, /driverStore\.vehicle\?\.id\s*\|\|\s*'--'/, 'vehicle panel must render the actual cached vehicle id')
assert.match(vehiclePanel, /emit\('submit-vehicle',?\)/, 'vehicle panel must keep the create-or-update submit action')
assert.match(vehiclePanel, />提交车辆<\/button>/, 'vehicle panel primary submit button must keep the current submit copy')
assert.match(vehiclePanel, /emit\('submit-vehicle-update',?\)[\s\S]*>更新<\/button>/, 'vehicle panel must keep the separate explicit update action')

console.log('driver vehicle create checks passed')
