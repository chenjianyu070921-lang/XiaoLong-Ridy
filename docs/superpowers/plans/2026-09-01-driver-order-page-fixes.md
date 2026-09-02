# Driver Order Page Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move nearby order browsing to the order tab, restore working accept/reject actions, remove popup warnings, and restart services with current backend routes.

**Architecture:** Keep the home page focused on status, map, and current dispatch operation. Put nearby order list state in `DriverHome.vue` and render it through `DriverOrdersPanel.vue`, including a compact 5-item page and a H5 popup with 10-item pages. Keep accept/reject routed through the existing `handleOrderAction(action, order)` function.

**Tech Stack:** Vue 3, Vant 4, Vite, Go driver API, existing `check-driver-web.mjs` structure checks.

## Global Constraints

- No standalone `/workbench` route.
- Homepage remains the driver workbench first screen.
- Avatar upload route is `/api/driver/v1/drivers/avatar/upload`.
- Nearby order actions must pass the full order object to `handleOrderAction`.
- Vant popup `teleport` must not receive boolean `false`.

---

### Task 1: Structure Checks

**Files:**
- Modify: `web/driver/scripts/check-driver-web.mjs`

**Interfaces:**
- Consumes: `DriverHome.vue`, `DriverOrdersPanel.vue`, `DriverProfileEdit.vue`
- Produces: failing checks for nearby order migration, popup teleport, and accept/reject wiring

- [ ] Add assertions requiring `nearbyOrders`, `nearbyOrderPageSize = 5`, `nearbyOrderExpandedPageSize = 10`, `nearbyOrderPopupVisible`, `openNearbyOrderPopup`, and `@order-action`.
- [ ] Add assertions rejecting `:teleport="false"`.
- [ ] Run `npm run test:driver-web` and verify it fails before implementation.

### Task 2: Order Tab Nearby Orders

**Files:**
- Modify: `web/driver/src/views/DriverHome.vue`
- Modify: `web/driver/src/components/driver-home/DriverOrdersPanel.vue`

**Interfaces:**
- Consumes: `listAvailableOrders(data, config)`, `handleOrderAction(action, order)`
- Produces: nearby order list with 5-item page and expanded 10-item H5 popup

- [ ] In `DriverHome.vue`, add nearby order state separate from historical order state.
- [ ] Fetch nearby orders with `{ page, pageSize }`, using page size 5 for inline and 10 for popup.
- [ ] Pass nearby order props and events into `DriverOrdersPanel.vue`.
- [ ] In `DriverOrdersPanel.vue`, render a nearby order section when the order tab is active, with refresh, expand, pagination, accept, and reject buttons.

### Task 3: Warnings And Services

**Files:**
- Modify: `web/driver/src/views/DriverHome.vue`
- Build/restart: `api/driver`, `web/driver`

**Interfaces:**
- Consumes: Vant popup teleport contract and Go driver API route table
- Produces: no boolean teleport warning, current backend route available, restarted frontend

- [ ] Replace `:teleport="false"` with a valid selector string.
- [ ] Build and restart `api-driver` so avatar upload and profile update routes exist in the running binary.
- [ ] Restart driver frontend.
- [ ] Run `npm run test:driver-web`, `npm run build`, and `go test ./api/driver -count=1`.
