import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const homePath = resolve(root, 'src/views/DriverHome.vue')
const source = readFileSync(homePath, 'utf8')
const inlinePanelCount = (source.match(/v-show="activeTab ===/g) || []).length
const workStatusCardMatches = [...source.matchAll(/<section\s+([^>]*class="work-status-card"[^>]*)>/g)]

if (inlinePanelCount > 1) {
  console.error(`DriverHome.vue still owns ${inlinePanelCount} tab panels; move panels into focused components.`)
  process.exit(1)
}

if (workStatusCardMatches.length !== 1) {
  console.error(`Expected exactly one work-status-card, found ${workStatusCardMatches.length}.`)
  process.exit(1)
}

const workStatusAttrs = workStatusCardMatches[0][1]
if (!/v-if="activeTab === 0"/.test(workStatusAttrs)) {
  console.error('work-status-card must only render on the home tab.')
  process.exit(1)
}

console.log('DriverHome.vue tab panels are delegated to focused components.')
