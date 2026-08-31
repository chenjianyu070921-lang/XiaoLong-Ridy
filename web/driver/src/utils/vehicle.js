function unwrapVehicleResponse(value) {
  return value?.vehicle || value || null
}

function hasVehicleDetails(vehicle) {
  return !!(vehicle && (vehicle.plateNo || vehicle.brand || vehicle.model))
}

export async function resolveCreatedVehicle(payload = {}, created = {}, loadVehicleById) {
  const createdVehicle = unwrapVehicleResponse(created) || {}
  const vehicleId = Number(createdVehicle.id || created?.id || 0)
  const fallback = {
    ...payload,
    ...createdVehicle,
    ...(vehicleId > 0 ? { id: vehicleId } : {})
  }

  if (hasVehicleDetails(createdVehicle)) {
    return fallback
  }

  if (vehicleId > 0 && typeof loadVehicleById === 'function') {
    try {
      const detailVehicle = unwrapVehicleResponse(await loadVehicleById(vehicleId))
      if (detailVehicle && typeof detailVehicle === 'object') {
        return {
          ...fallback,
          ...detailVehicle,
          id: Number(detailVehicle.id || vehicleId)
        }
      }
    } catch {
      return fallback
    }
  }

  return fallback
}

export function resolveDriverVehicleId(driver = {}, vehicle = null) {
  return Number(driver?.vehicleId || driver?.vehicleID || vehicle?.id || 0)
}
