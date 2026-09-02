import { wgs84ToGcj02, normalizeBrowserLocationForAmap } from '../src/utils/geo.js'

function assertNear(actual, expected, tolerance, message) {
  if (Math.abs(actual - expected) > tolerance) {
    throw new Error(`${message}: expected ${expected}, got ${actual}`)
  }
}

function assertEqual(actual, expected, message) {
  if (actual !== expected) {
    throw new Error(`${message}: expected ${expected}, got ${actual}`)
  }
}

const suqianVocationalCollegeWgs84 = { longitude: 118.29846, latitude: 33.94003 }
const suqianGcj02 = wgs84ToGcj02(suqianVocationalCollegeWgs84.longitude, suqianVocationalCollegeWgs84.latitude)

assertNear(suqianGcj02.longitude, 118.30399, 0.00001, '宿迁 WGS84 longitude should convert to GCJ-02')
assertNear(suqianGcj02.latitude, 33.93871, 0.00001, '宿迁 WGS84 latitude should convert to GCJ-02')

const normalized = normalizeBrowserLocationForAmap(suqianVocationalCollegeWgs84)
assertNear(normalized.longitude, suqianGcj02.longitude, 0.000001, 'normalized longitude should use GCJ-02')
assertNear(normalized.latitude, suqianGcj02.latitude, 0.000001, 'normalized latitude should use GCJ-02')

const paris = wgs84ToGcj02(2.3522, 48.8566)
assertEqual(paris.longitude, 2.3522, 'longitude outside China should remain unchanged')
assertEqual(paris.latitude, 48.8566, 'latitude outside China should remain unchanged')

console.log('driver geo coordinate conversion checks passed')
