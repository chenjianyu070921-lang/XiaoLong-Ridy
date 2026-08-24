const authView = document.querySelector("[data-view='auth']");
const dashboardView = document.querySelector("[data-view='dashboard']");
const title = document.querySelector("#auth-title");
const buttons = document.querySelectorAll("[data-auth-target]");
const forms = document.querySelectorAll("[data-auth-form]");
const message = document.querySelector("[data-form-message]");
const dashboardMessage = document.querySelector("[data-dashboard-message]");
const menuButton = document.querySelector("[data-menu-button]");
const menuPanel = document.querySelector("[data-menu-panel]");
const panelButtons = document.querySelectorAll("[data-panel-target]");
const panels = document.querySelectorAll("[data-panel]");
const editModal = document.querySelector("[data-edit-modal]");
const editMessage = document.querySelector("[data-edit-message]");
const orderList = document.querySelector("[data-order-list]");
const orderListEmpty = document.querySelector("[data-order-list-empty]");
const orderStatus = document.querySelector("[data-order-status]");
const orderSummary = document.querySelector("[data-order-summary]");
const orderPageLabel = document.querySelector("[data-order-page]");
const orderPrevious = document.querySelector("[data-order-prev]");
const orderNext = document.querySelector("[data-order-next]");
const vehicleForm = document.querySelector("[data-vehicle-form]");
const vehicleMessage = document.querySelector("[data-vehicle-message]");
const certificationForm = document.querySelector("[data-certification-form]");
const certificationMessage = document.querySelector("[data-certification-message]");
const finishModal = document.querySelector("[data-finish-modal]");
const finishMessage = document.querySelector("[data-finish-message]");
const cachedDriver = readJSON("driverProfile") || null;
const cachedVehicle = readJSON("driverVehicle") || null;
const cachedCertification = readJSON("driverCertification") || null;
const cachedCurrentOrder = readJSON("driverCurrentOrder") || null;

const state = {
  token: localStorage.getItem("driverToken") || "",
  driver: cachedDriver,
  vehicle: cachedVehicle,
  vehicleId: Number(localStorage.getItem("driverVehicleId") || cachedVehicle?.id || 0),
  certification: cachedCertification,
  onlineStatus: Number(cachedDriver?.onlineStatus ?? 0),
  tripPhase: localStorage.getItem("driverTripPhase") || "idle",
  currentOrderId: localStorage.getItem("driverCurrentOrderId") || "",
  finishOrderId: 0,
  currentOrder: cachedCurrentOrder,
  orders: [],
  orderPage: 1,
  orderPageSize: 8,
  orderStatus: 0,
  orderTotal: 0,
};

// 心跳定时器与间隔：用 var 声明避免 TDZ（防止函数先于 let 初始化被调用时报错）。
var heartbeatTimer = null;
var HEARTBEAT_INTERVAL = 15000;

// 位置上报定时器与间隔：上线后定时上报经纬度，用 var 声明避免 TDZ。
var locationTimer = null;
var LOCATION_INTERVAL = 10000;
var lastLatitude = null;
var lastLongitude = null;

var tripRefreshTimer = null;
var TRIP_REFRESH_INTERVAL = 5000;
var tripRefreshInFlight = false;

buttons.forEach((button) => {
  button.addEventListener("click", () => {
    const target = button.dataset.authTarget;

    buttons.forEach((item) => item.classList.toggle("is-active", item === button));
    forms.forEach((form) => {
      form.classList.toggle("is-active", form.dataset.authForm === target);
    });

    title.textContent = {
      register: "司机注册",
      "login-sms": "验证码登录",
      login: "司机登录",
    }[target] || "司机登录";
    setAuthMessage("");
  });
});

forms.forEach((form) => {
  form.addEventListener("submit", async (event) => {
    event.preventDefault();

    const submitButton = form.querySelector("button[type='submit']");
    const payload = Object.fromEntries(new FormData(form).entries());

    submitButton.disabled = true;
    setAuthMessage("提交中...");

    try {
      const result = await requestJSON(form.action, {
        method: "POST",
        body: JSON.stringify(compactPayload(payload)),
      });

      if (form.dataset.authForm === "login" || form.dataset.authForm === "login-sms") {
        state.token = result.data.token;
        state.driver = result.data.driver;
        localStorage.setItem("driverToken", state.token);
        localStorage.setItem("driverProfile", JSON.stringify(state.driver));
        setAuthMessage("登录成功", "success");
        showDashboard();
        await loadDashboardData();
      } else {
        setAuthMessage("注册成功，请等待资质审核", "success");
      }
    } catch (error) {
      setAuthMessage(error.message || "服务暂不可用", "error");
    } finally {
      submitButton.disabled = false;
    }
  });
});

document.querySelector("[data-send-sms-code]").addEventListener("click", async () => {
  const phone = document.querySelector("[data-sms-phone]").value.trim();
  if (!phone) {
    setAuthMessage("请输入手机号", "error");
    return;
  }
  const button = document.querySelector("[data-send-sms-code]");
  button.disabled = true;
  setAuthMessage("正在发送验证码...");
  try {
    await requestJSON("/driver/sms-code", {
      method: "POST",
      body: JSON.stringify({ phone }),
    });
    setAuthMessage("验证码已发送", "success");
  } catch (error) {
    setAuthMessage(error.message || "验证码发送失败", "error");
  } finally {
    window.setTimeout(() => {
      button.disabled = false;
    }, 3000);
  }
});

menuButton.addEventListener("click", () => {
  menuPanel.hidden = !menuPanel.hidden;
});

panelButtons.forEach((button) => {
  button.addEventListener("click", () => {
    showPanel(button.dataset.panelTarget);
    menuPanel.hidden = true;
  });
});

document.querySelector("[data-open-edit]").addEventListener("click", () => {
  openEditModal();
  menuPanel.hidden = true;
});

document.querySelector("[data-open-dispatches]").addEventListener("click", () => {
  orderStatus.value = "1";
  state.orderStatus = 1;
  loadOrders(1);
  document.querySelector("[data-order-panel]").scrollIntoView({ behavior: "smooth", block: "start" });
  menuPanel.hidden = true;
});

document.querySelectorAll("[data-close-edit]").forEach((button) => {
  button.addEventListener("click", closeEditModal);
});

document.querySelectorAll("[data-close-finish]").forEach((button) => {
  button.addEventListener("click", closeFinishModal);
});

editModal.addEventListener("click", (event) => {
  if (event.target === editModal) {
    closeEditModal();
  }
});

finishModal.addEventListener("click", (event) => {
  if (event.target === finishModal) {
    closeFinishModal();
  }
});

document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && !editModal.hidden) {
    closeEditModal();
  }
  if (event.key === "Escape" && !finishModal.hidden) {
    closeFinishModal();
  }
});

document.querySelector("[data-logout]").addEventListener("click", () => {
  forceLogout();
});

// forceLogout 清空本地登录态并回到登录页，供登出按钮与心跳被踢/凭证失效时复用。
function forceLogout() {
  stopHeartbeat();
  state.token = "";
  state.driver = null;
  state.vehicle = null;
  state.vehicleId = 0;
  state.certification = null;
  state.currentOrder = null;
  state.orders = [];
  state.onlineStatus = 0;
  state.tripPhase = "idle";
  state.currentOrderId = "";
  localStorage.removeItem("driverToken");
  localStorage.removeItem("driverProfile");
  localStorage.removeItem("driverOnlineStatus");
  localStorage.removeItem("driverVehicle");
  localStorage.removeItem("driverVehicleId");
  localStorage.removeItem("driverCertification");
  localStorage.removeItem("driverCurrentOrder");
  localStorage.removeItem("driverTripPhase");
  localStorage.removeItem("driverCurrentOrderId");
  stopTripRealtime();
  dashboardView.hidden = true;
  authView.hidden = false;
  menuPanel.hidden = true;
}

document.querySelector("[data-refresh]").addEventListener("click", () => {
  loadDashboardData();
});

document.querySelector("[data-refresh-vehicle]").addEventListener("click", () => {
  loadVehicle();
});

document.querySelector("[data-refresh-certification]").addEventListener("click", () => {
  loadCertification();
});

document.querySelector("[data-refresh-orders]").addEventListener("click", () => {
  loadOrders(1);
});

orderStatus.addEventListener("change", () => {
  state.orderStatus = Number(orderStatus.value || 0);
  loadOrders(1);
});

orderPrevious.addEventListener("click", () => {
  if (state.orderPage > 1) {
    loadOrders(state.orderPage - 1);
  }
});

orderNext.addEventListener("click", () => {
  if (state.orderPage * state.orderPageSize < state.orderTotal) {
    loadOrders(state.orderPage + 1);
  }
});

orderList.addEventListener("click", (event) => {
  const button = event.target.closest("[data-order-action]");
  if (button) {
    if (button.dataset.orderAction === "finish-trip") {
      openFinishModal(button.dataset.orderId);
      return;
    }
    handleOrderAction(button.dataset.orderAction, button.dataset.orderId);
  }
});

document.querySelector("[data-online]").addEventListener("click", async () => {
  await setWorkStatus("/driver/online", 1, "已上线，当前空闲中");
});

document.querySelector("[data-offline]").addEventListener("click", async () => {
  await setWorkStatus("/driver/offline", 0, "已下线休息");
});

document.querySelector("[data-edit-form]").addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!state.driver?.id) {
    setEditMessage("司机信息缺失，请重新登录", "error");
    return;
  }

  const form = event.currentTarget;
  const payload = compactPayload(Object.fromEntries(new FormData(form).entries()));
  payload.id = state.driver.id;
  const submitButton = form.querySelector("button[type='submit']");
  submitButton.disabled = true;
  setEditMessage("正在保存...");

  try {
    await requestJSON(form.action, {
      method: "POST",
      token: state.token,
      body: JSON.stringify(payload),
    });
    setEditMessage("司机信息已更新", "success");
    await loadDashboardData();
    window.setTimeout(closeEditModal, 500);
  } catch (error) {
    setEditMessage(error.message || "保存失败", "error");
  } finally {
    submitButton.disabled = false;
  }
});

vehicleForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const submitButton = vehicleForm.querySelector("button[type='submit']");
  submitButton.disabled = true;
  setVehicleMessage("正在提交车辆信息...");

  try {
    const payload = buildVehiclePayload(Object.fromEntries(new FormData(vehicleForm).entries()));
    const result = await requestJSON(vehicleForm.action, {
      method: "POST",
      token: state.token,
      body: JSON.stringify(payload),
    });
    state.vehicleId = Number(result.data.id || 0);
    localStorage.setItem("driverVehicleId", String(state.vehicleId));
    setVehicleMessage("车辆信息提交成功，等待审核", "success");
    await loadVehicle();
  } catch (error) {
    setVehicleMessage(error.message || "车辆信息提交失败", "error");
  } finally {
    submitButton.disabled = false;
  }
});

certificationForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const submitButton = certificationForm.querySelector("button[type='submit']");
  submitButton.disabled = true;
  setCertificationMessage("正在上传资质...");

  try {
    const payload = await buildCertificationPayload(certificationForm);
    await requestJSON(certificationForm.action, {
      method: "POST",
      token: state.token,
      body: JSON.stringify(payload),
    });
    setCertificationMessage("资质上传成功，等待后台审核", "success");
    await loadCertification();
  } catch (error) {
    setCertificationMessage(error.message || "资质上传失败", "error");
  } finally {
    submitButton.disabled = false;
  }
});

document.querySelector("[data-finish-form]").addEventListener("submit", async (event) => {
  event.preventDefault();
  await submitFinishTrip(event.currentTarget);
});

if (state.token && state.driver?.id) {
  showDashboard();
  loadDashboardData();
  if (state.onlineStatus === 1) {
    startHeartbeat();
    startLocationReporting();
    startTripRealtime();
  }
  connectPushChannel();
} else {
  authView.hidden = false;
  dashboardView.hidden = true;
}

async function loadDashboardData() {
  if (!state.token || !state.driver?.id) {
    return;
  }

  try {
    const [profileResult, scoreResult, ordersResult] = await Promise.allSettled([
      requestJSON(`/driver/me?driverId=${encodeURIComponent(state.driver.id)}`, {
        method: "GET",
        token: state.token,
      }),
      requestJSON(`/driver/ai-score?driverId=${encodeURIComponent(state.driver.id)}`, {
        method: "GET",
        token: state.token,
      }),
      requestOrders(state.orderPage),
    ]);

    if (profileResult.status === "fulfilled") {
      const driver = profileResult.value.data.driver;
      state.driver = { ...state.driver, ...driver };
      state.onlineStatus = Number(driver.onlineStatus ?? 0);
      localStorage.setItem("driverProfile", JSON.stringify(state.driver));
      localStorage.setItem("driverOnlineStatus", String(state.onlineStatus));
      renderDriver();
    } else {
      renderDriver();
      setDashboardMessage(profileResult.reason.message || "个人信息加载失败", "error");
    }

    if (scoreResult.status === "fulfilled") {
      const score = scoreResult.value.data.aiScore;
      document.querySelector("[data-service-score]").textContent = Number(score || 0).toFixed(1);
    } else {
      document.querySelector("[data-service-score]").textContent = "--";
    }

    if (ordersResult.status === "fulfilled") {
      renderOrders(ordersResult.value.data);
    } else {
      renderOrders({ list: [], total: 0, page: state.orderPage, pageSize: state.orderPageSize });
      setDashboardMessage(ordersResult.reason.message || "订单列表加载失败", "error");
    }

    renderStatus();
    renderVehicle();
    renderCertification();
    if (state.onlineStatus > 0 || state.tripPhase !== "idle") {
      startTripRealtime();
    }
  } catch (error) {
    setDashboardMessage("主页数据加载失败", "error");
  }
}

async function loadVehicle() {
  if (!state.token) {
    return;
  }
  if (!state.vehicleId) {
    renderVehicle();
    setVehicleMessage("请先提交车辆信息", "error");
    return;
  }

  setVehicleMessage("正在查询车辆信息...");
  try {
    const result = await requestJSON(`/driver/vehicles/get?id=${encodeURIComponent(state.vehicleId)}`, {
      method: "GET",
      token: state.token,
    });
    state.vehicle = result.data.vehicle;
    localStorage.setItem("driverVehicle", JSON.stringify(state.vehicle));
    renderVehicle();
    setVehicleMessage("车辆信息已刷新", "success");
  } catch (error) {
    setVehicleMessage(error.message || "车辆信息查询失败", "error");
  }
}

async function loadCertification() {
  if (!state.token) {
    return;
  }

  setCertificationMessage("正在查询资质...");
  try {
    const result = await requestJSON("/driver/certification", {
      method: "GET",
      token: state.token,
    });
    state.certification = result.data.found ? result.data.certification : null;
    if (state.certification) {
      localStorage.setItem("driverCertification", JSON.stringify(state.certification));
    } else {
      localStorage.removeItem("driverCertification");
    }
    renderCertification();
    setCertificationMessage(state.certification ? "资质信息已刷新" : "暂无资质记录", state.certification ? "success" : "");
  } catch (error) {
    setCertificationMessage(error.message || "资质查询失败", "error");
  }
}

async function loadOrders(page = state.orderPage) {
  if (!state.token) {
    return;
  }

  setOrderListLoading(true);
  try {
    const result = await requestOrders(page);
    renderOrders(result.data);
  } catch (error) {
    renderOrders({ list: [], total: 0, page, pageSize: state.orderPageSize });
    setDashboardMessage(error.message || "订单列表加载失败", "error");
  } finally {
    setOrderListLoading(false);
  }
}

async function requestOrders(page) {
  if (state.orderStatus === 1) {
    const dispatches = await requestDispatches(page);
    return normalizeDispatchResult(dispatches.data);
  }

  if (state.orderStatus !== 0) {
    return requestOrderList(page, state.orderStatus);
  }

  const [dispatches, orders] = await Promise.all([
    requestDispatches(page),
    requestOrderList(page, 0),
  ]);
  return mergeOrderResults(dispatches.data, orders.data);
}

function requestOrderList(page, status) {
  return requestJSON("/driver/orders", {
    method: "POST",
    token: state.token,
    body: JSON.stringify({
      page,
      pageSize: state.orderPageSize,
      status,
    }),
  });
}

function requestDispatches(page) {
  return requestJSON("/driver/dispatches", {
    method: "POST",
    token: state.token,
    body: JSON.stringify({
      page,
      pageSize: state.orderPageSize,
      status: 1,
    }),
  });
}

function normalizeDispatchResult(data = {}) {
  return {
    data: {
      list: normalizeDispatchOrders(data),
      total: Number(data.total || 0),
      page: Number(data.page || state.orderPage || 1),
      pageSize: Number(data.pageSize || state.orderPageSize),
    },
  };
}

function mergeOrderResults(dispatchData = {}, orderData = {}) {
  const dispatchOrders = normalizeDispatchOrders(dispatchData);
  const dispatchOrderIds = new Set(dispatchOrders.map((order) => Number(order.orderId || 0)));
  const orders = Array.isArray(orderData.list)
    ? orderData.list.filter((order) => !dispatchOrderIds.has(Number(order.orderId || 0)))
    : [];
  return {
    data: {
      list: [...dispatchOrders, ...orders],
      total: Number(dispatchData.total || 0) + Number(orderData.total || 0),
      page: Number(orderData.page || dispatchData.page || state.orderPage || 1),
      pageSize: Number(orderData.pageSize || dispatchData.pageSize || state.orderPageSize),
    },
  };
}

function normalizeDispatchOrders(data = {}) {
  return Array.isArray(data.list)
    ? data.list.map((item) => ({
        ...(item.order || {}),
        source: "dispatch",
        dispatchStatus: item.dispatch?.status || 0,
        dispatchId: item.dispatch?.id || 0,
        matchScore: item.dispatch?.matchScore || 0,
      }))
    : [];
}

async function setWorkStatus(path, fallbackStatus, successText) {
  setStatusButtonsDisabled(true);
  setDashboardMessage("状态切换中...");
  try {
    const result = await requestJSON(path, {
      method: "POST",
      token: state.token,
      body: "{}",
    });
    state.onlineStatus = Number(result.data.onlineStatus ?? fallbackStatus);
    localStorage.setItem("driverOnlineStatus", String(state.onlineStatus));
    persistOnlineStatus();
    if (state.onlineStatus === 0) {
      state.tripPhase = "idle";
      state.currentOrderId = "";
      state.currentOrder = null;
      localStorage.setItem("driverTripPhase", state.tripPhase);
      localStorage.removeItem("driverCurrentOrderId");
      localStorage.removeItem("driverCurrentOrder");
      stopHeartbeat();
      stopLocationReporting();
      stopTripRealtime();
    } else {
      startHeartbeat();
      startLocationReporting();
      startTripRealtime();
      connectPushChannel();
    }
    renderStatus();
    renderCoreArea();
    setDashboardMessage(successText, "success");
  } catch (error) {
    setDashboardMessage(error.message || "状态切换失败", "error");
  } finally {
    renderStatus();
    renderCoreArea();
  }
}

// ---- 心跳保活：上线后定时上报，凭证失效或被踢则退出登录 ----
// heartbeatTimer / HEARTBEAT_INTERVAL 已在顶部用 var 声明，避免 TDZ。

function getDeviceId() {
  let deviceId = localStorage.getItem("driverDeviceId");
  if (!deviceId) {
    deviceId = "web-" + Math.random().toString(36).slice(2, 10);
    localStorage.setItem("driverDeviceId", deviceId);
  }
  return deviceId;
}

function startHeartbeat() {
  if (heartbeatTimer) return;
  sendHeartbeat();
  heartbeatTimer = window.setInterval(sendHeartbeat, HEARTBEAT_INTERVAL);
}

function stopHeartbeat() {
  if (heartbeatTimer) {
    window.clearInterval(heartbeatTimer);
    heartbeatTimer = null;
  }
}

// ---- 位置上报：上线后定时上报经纬度，供派单引擎就近匹配 ----
function startLocationReporting() {
  if (locationTimer) return;
  // 浏览器定位（若用户授权）作为默认坐标来源；上报前至少填充一次有效坐标。
  if (navigator.geolocation) {
    navigator.geolocation.watchPosition(
      (position) => {
        lastLatitude = position.coords.latitude;
        lastLongitude = position.coords.longitude;
        sendLocation();
      },
      () => {},
      { enableHighAccuracy: true, maximumAge: 5000 }
    );
  }
  sendLocation();
  locationTimer = window.setInterval(sendLocation, LOCATION_INTERVAL);
}

function stopLocationReporting() {
  if (locationTimer) {
    window.clearInterval(locationTimer);
    locationTimer = null;
  }
}

async function sendLocation() {
  if (!state.token || !state.driver?.id) {
    stopLocationReporting();
    return;
  }
  if (lastLatitude === null || lastLongitude === null) {
    return;
  }
  try {
    await requestJSON("/driver/location/report", {
      method: "POST",
      token: state.token,
      body: JSON.stringify({
        deviceId: getDeviceId(),
        longitude: lastLongitude,
        latitude: lastLatitude,
      }),
    });
  } catch (error) {
    // 位置上报失败不影响主流程，仅静默重试。
  }
}

// ---- 行程实时更新：优先吃 WebSocket 推送；没有推送时用轻量轮询兜底 ----
function startTripRealtime() {
  if (tripRefreshTimer) return;
  refreshRealtimeTrip();
  tripRefreshTimer = window.setInterval(refreshRealtimeTrip, TRIP_REFRESH_INTERVAL);
}

function stopTripRealtime() {
  if (tripRefreshTimer) {
    window.clearInterval(tripRefreshTimer);
    tripRefreshTimer = null;
  }
  tripRefreshInFlight = false;
}

async function refreshRealtimeTrip() {
  if (!state.token || tripRefreshInFlight) {
    return;
  }
  tripRefreshInFlight = true;
  try {
    const result = await requestRealtimeOrders();
    renderOrders(result.data);
  } catch (error) {
    // 实时刷新失败不打断司机操作，下一轮继续尝试。
  } finally {
    tripRefreshInFlight = false;
  }
}

async function requestRealtimeOrders() {
  const [dispatches, orders] = await Promise.all([
    requestDispatches(1),
    requestOrderList(1, 0),
  ]);
  return mergeOrderResults(dispatches.data, orders.data);
}

// ---- 推送通道（WebSocket）：当前项目未提供司机 WebSocket 端点，仅保留可配置连接骨架 ----
// 当部署注入了 window.DRIVER_WS_URL 时建立连接，收到派单/订单推送后触发对应刷新；
// 未配置时不连接，派单与订单仍通过轮询获取（见 requestDispatches / requestOrderList）。
var pushSocket = null;

function connectPushChannel() {
  const wsURL = window.DRIVER_WS_URL || "";
  if (!wsURL || !state.token) {
    return;
  }
  if (pushSocket && pushSocket.readyState <= 1) {
    return;
  }
  try {
    pushSocket = new WebSocket(wsURL);
  } catch (error) {
    return;
  }
  pushSocket.onopen = () => {
    try {
      pushSocket.send(JSON.stringify({ type: "auth", token: state.token }));
    } catch (error) {}
  };
  pushSocket.onmessage = (event) => {
    let payload = {};
    try {
      payload = JSON.parse(event.data || "{}");
    } catch (error) {
      return;
    }
    // 收到派单或订单相关推送，刷新对应列表。
    if (payload.type === "dispatch" || payload.type === "order") {
      refreshRealtimeTrip();
    }
  };
  pushSocket.onclose = () => {
    // 登录态仍在则延迟重连，最多避免风暴。
    if (state.token) {
      window.setTimeout(connectPushChannel, 5000);
    }
  };
  pushSocket.onerror = () => {
    if (pushSocket) {
      pushSocket.close();
    }
  };
}

async function sendHeartbeat() {
  if (!state.token) {
    stopHeartbeat();
    return;
  }
  try {
    const result = await requestJSON("/driver/heartbeat", {
      method: "POST",
      token: state.token,
      body: JSON.stringify({
        deviceId: getDeviceId(),
        longitude: 0,
        latitude: 0,
      }),
    });
    if (result.data && result.data.kicked) {
      forceLogout();
      setDashboardMessage("账号已在其他设备登录，已退出", "error");
    }
  } catch (error) {
    // 凭证过期/未授权（401）统一退出登录
    if (String(error.message || "").includes("401") || String(error.message || "").includes("登录")) {
      forceLogout();
    }
  }
}

async function handleOrderAction(action, orderId) {
  if (action === "finish-trip") {
    openFinishModal(orderId);
    return;
  }
  const orderID = Number(orderId || state.currentOrderId);
  if (!orderID || orderID <= 0) {
    setDashboardMessage("订单信息无效，请刷新订单列表", "error");
    return;
  }

  const config = {
    accept: {
      path: "/driver/orders/accept",
      phase: "pickup",
      message: "接单成功，状态已切换为正在接驾",
      payload: { orderId: orderID },
    },
    reject: {
      path: "/driver/orders/reject",
      phase: "idle",
      message: "已拒绝派单",
      payload: { orderId: orderID, reason: "司机主动拒单" },
    },
    "confirm-arrive": {
      path: "/driver/orders/confirm-arrive",
      phase: "pickup",
      message: "已确认到达上车点",
      payload: { orderId: orderID },
    },
    "start-trip": {
      path: "/driver/orders/start-trip",
      phase: "trip",
      message: "行程已开始",
      payload: { orderId: orderID },
    },
  }[action];

  if (!config) {
    return;
  }

  setOrderButtonsDisabled(true);
  setDashboardMessage("订单操作提交中...");
  try {
    const cachedOrder = findOrderById(orderID);
    await requestJSON(config.path, {
      method: "POST",
      token: state.token,
      body: JSON.stringify(config.payload),
    });
    state.currentOrderId = config.phase === "idle" ? "" : String(orderID);
    state.tripPhase = config.phase;
    state.currentOrder = config.phase === "idle" ? null : cachedOrder;
    if (config.phase === "trip") {
      state.onlineStatus = 2;
      localStorage.setItem("driverOnlineStatus", "2");
      persistOnlineStatus();
    }
    if (config.phase === "idle" && state.onlineStatus === 2) {
      state.onlineStatus = 1;
      localStorage.setItem("driverOnlineStatus", "1");
      persistOnlineStatus();
    }
    localStorage.setItem("driverTripPhase", state.tripPhase);
    if (state.currentOrderId) {
      localStorage.setItem("driverCurrentOrderId", state.currentOrderId);
      if (state.currentOrder) {
        localStorage.setItem("driverCurrentOrder", JSON.stringify(state.currentOrder));
      }
    } else {
      localStorage.removeItem("driverCurrentOrderId");
      localStorage.removeItem("driverCurrentOrder");
    }
    renderStatus();
    renderCoreArea();
    if (config.phase !== "idle") {
      startTripRealtime();
    }
    await loadOrders(state.orderPage);
    setDashboardMessage(config.message, "success");
  } catch (error) {
    setDashboardMessage(error.message || "订单操作失败", "error");
  } finally {
    setOrderButtonsDisabled(false);
  }
}

function showDashboard() {
  authView.hidden = true;
  dashboardView.hidden = false;
  renderDriver();
  renderVehicle();
  renderCertification();
  renderStatus();
  renderCoreArea();
}

function showPanel(target) {
  panels.forEach((panel) => {
    panel.hidden = panel.dataset.panel !== target;
  });
  if (target === "certification") {
    renderCertification();
  }
}

function openEditModal() {
  renderDriver();
  setEditMessage("");
  editModal.hidden = false;
  document.querySelector("[data-edit-real-name]").focus();
}

function closeEditModal() {
  editModal.hidden = true;
  setEditMessage("");
}

function openFinishModal(orderId) {
  const orderID = Number(orderId || state.currentOrderId || 0);
  if (!orderID) {
    setDashboardMessage("订单信息无效，请刷新订单列表", "error");
    return;
  }
  state.finishOrderId = orderID;
  setFinishMessage("");
  finishModal.hidden = false;
  document.querySelector("[data-finish-form] input[name='actualDistanceM']").focus();
}

function closeFinishModal() {
  finishModal.hidden = true;
  state.finishOrderId = 0;
  setFinishMessage("");
  document.querySelector("[data-finish-form]").reset();
}

async function submitFinishTrip(form) {
  const orderID = Number(state.finishOrderId || state.currentOrderId || 0);
  if (!orderID) {
    setFinishMessage("订单信息无效，请刷新订单列表", "error");
    return;
  }
  const payload = Object.fromEntries(new FormData(form).entries());
  const body = {
    orderId: orderID,
    actualDistanceM: Number(payload.actualDistanceM || 0),
    actualDurationS: Number(payload.actualDurationS || 0),
    actualPriceCents: Number(payload.actualPriceCents || 0),
  };
  if (body.actualDistanceM < 0 || body.actualDurationS < 0 || body.actualPriceCents < 0) {
    setFinishMessage("行程数据不能为负数", "error");
    return;
  }

  const submitButton = form.querySelector("button[type='submit']");
  submitButton.disabled = true;
  setFinishMessage("正在结束行程...");
  try {
    await requestJSON(form.action, {
      method: "POST",
      token: state.token,
      body: JSON.stringify(body),
    });
    state.currentOrderId = "";
    state.currentOrder = null;
    state.tripPhase = "idle";
    if (state.onlineStatus === 2) {
      state.onlineStatus = 1;
      localStorage.setItem("driverOnlineStatus", "1");
      persistOnlineStatus();
    }
    localStorage.setItem("driverTripPhase", state.tripPhase);
    localStorage.removeItem("driverCurrentOrderId");
    localStorage.removeItem("driverCurrentOrder");
    closeFinishModal();
    renderStatus();
    renderCoreArea();
    await loadOrders(state.orderPage);
    setDashboardMessage("行程已结束，当前空闲", "success");
  } catch (error) {
    setFinishMessage(error.message || "结束行程失败", "error");
  } finally {
    submitButton.disabled = false;
  }
}

function renderDriver() {
  const driver = state.driver || {};
  const name = driver.realName || "司机";
  const phone = driver.phone || "--";
  const avatar = driver.avatarUrl || "";

  document.querySelector("[data-driver-name]").textContent = name;
  document.querySelector("[data-driver-phone]").textContent = phone;
  document.querySelector("[data-driver-initial]").textContent = name.slice(0, 1) || "司";
  document.querySelector("[data-profile-phone]").textContent = phone;
  document.querySelector("[data-profile-license]").textContent = driver.driverLicenseNo || "--";
  document.querySelector("[data-profile-status]").textContent = formatDriverStatus(driver.status);
  document.querySelector("[data-edit-real-name]").value = driver.realName || "";
  document.querySelector("[data-edit-avatar-url]").value = avatar;
  document.querySelector("[data-edit-license]").value = driver.driverLicenseNo || "";

  const avatarImg = document.querySelector("[data-driver-avatar]");
  const initial = document.querySelector("[data-driver-initial]");
  if (avatar) {
    avatarImg.src = avatar;
    avatarImg.hidden = false;
    initial.hidden = true;
  } else {
    avatarImg.hidden = true;
    initial.hidden = false;
  }
}

function renderVehicle() {
  const vehicle = state.vehicle || {};
  document.querySelector("[data-vehicle-id]").textContent = vehicle.id || state.vehicleId || "--";
  document.querySelector("[data-vehicle-plate]").textContent = vehicle.plateNo || "--";
  document.querySelector("[data-vehicle-model]").textContent = [vehicle.brand, vehicle.model].filter(Boolean).join(" ") || "--";
  document.querySelector("[data-vehicle-status]").textContent = formatVehicleStatus(vehicle.status);
  renderCertification();
}

function renderCertification() {
  const certification = state.certification || {};
  const vehicleID = certification.vehicleId || state.vehicleId || "";
  document.querySelector("[data-certification-id]").textContent = certification.id || "--";
  document.querySelector("[data-certification-vehicle]").textContent = vehicleID || "--";
  document.querySelector("[data-certification-status]").textContent = formatCertificationStatus(certification.auditStatus);
  document.querySelector("[data-certification-remark]").textContent = certification.auditRemark || "--";
  document.querySelector("[data-certification-vehicle-id]").value = vehicleID || "";
}

function renderStatus() {
  const current = document.querySelector("[data-current-status]");
  const dot = document.querySelector("[data-status-dot]");
  const cards = document.querySelectorAll("[data-status-card]");
  const idleState = document.querySelector("[data-idle-state]");
  const pickupState = document.querySelector("[data-pickup-state]");
  const tripState = document.querySelector("[data-trip-state]");
  const onlineButton = document.querySelector("[data-online]");
  const offlineButton = document.querySelector("[data-offline]");

  let statusKey = "offline";
  let label = "下线休息";
  if (state.tripPhase === "pickup") {
    statusKey = "pickup";
    label = "正在接驾";
  } else if (state.tripPhase === "trip" || state.onlineStatus === 2) {
    statusKey = "trip";
    label = "行程进行中";
  } else if (state.onlineStatus === 1) {
    statusKey = "idle";
    label = "空闲中";
  }

  current.textContent = label;
  dot.dataset.status = statusKey;
  cards.forEach((card) => {
    card.classList.toggle("is-active", card.dataset.statusCard === statusKey);
  });
  idleState.textContent = state.onlineStatus === 1 && state.tripPhase === "idle" ? "可接单" : "未接单";
  pickupState.textContent = state.tripPhase === "pickup" ? "进行中" : "空闲";
  tripState.textContent = state.tripPhase === "trip" || state.onlineStatus === 2 ? "进行中" : "空闲";
  onlineButton.classList.toggle("is-current", state.onlineStatus === 1);
  offlineButton.classList.toggle("is-current", state.onlineStatus === 0);
  onlineButton.disabled = state.onlineStatus === 1;
  offlineButton.disabled = state.onlineStatus === 0;
}

function setStatusButtonsDisabled(disabled) {
  document.querySelector("[data-online]").disabled = disabled;
  document.querySelector("[data-offline]").disabled = disabled;
}

function syncCurrentTripFromOrders(list) {
  const orders = Array.isArray(list) ? list : [];
  const currentID = Number(state.currentOrderId || 0);
  let activeOrder = currentID ? orders.find((order) => Number(order.orderId || 0) === currentID) : null;
  if (!activeOrder) {
    activeOrder = orders.find((order) => [2, 3].includes(Number(order.status || 0))) || null;
  }

  if (!activeOrder) {
    if (state.tripPhase !== "idle" && currentID) {
      renderCoreArea();
    }
    return;
  }

  const phase = phaseFromOrderStatus(activeOrder.status);
  if (phase === "idle") {
    state.currentOrder = null;
    state.currentOrderId = "";
    state.tripPhase = "idle";
    if (state.onlineStatus === 2) {
      state.onlineStatus = 1;
      localStorage.setItem("driverOnlineStatus", "1");
      persistOnlineStatus();
    }
    localStorage.removeItem("driverCurrentOrder");
    localStorage.removeItem("driverCurrentOrderId");
  } else {
    state.currentOrder = activeOrder;
    state.currentOrderId = String(activeOrder.orderId || "");
    state.tripPhase = phase;
    if (phase === "trip") {
      state.onlineStatus = 2;
      localStorage.setItem("driverOnlineStatus", "2");
      persistOnlineStatus();
    }
    localStorage.setItem("driverTripPhase", state.tripPhase);
    localStorage.setItem("driverCurrentOrderId", state.currentOrderId);
    localStorage.setItem("driverCurrentOrder", JSON.stringify(state.currentOrder));
  }
  renderStatus();
  renderCoreArea();
}

function phaseFromOrderStatus(status) {
  const value = Number(status || 0);
  if (value === 2) {
    return "pickup";
  }
  if (value === 3) {
    return "trip";
  }
  return "idle";
}

function findOrderById(orderID) {
  return state.orders.find((order) => Number(order.orderId || 0) === Number(orderID || 0)) || null;
}

function renderCoreArea() {
  const corePill = document.querySelector("[data-core-pill]");
  const tripEmpty = document.querySelector("[data-trip-empty]");
  const idleBadge = document.querySelector("[data-idle-badge]");
  const idleTitle = document.querySelector("[data-idle-title]");
  const idleDesc = document.querySelector("[data-idle-desc]");
  const fromAddress = document.querySelector("[data-from-address]");
  const toAddress = document.querySelector("[data-to-address]");
  const orderNo = document.querySelector("[data-current-order-no]");
  const orderStatusText = document.querySelector("[data-current-order-status]");
  const refreshTime = document.querySelector("[data-trip-refresh-time]");

  const hasTrip = state.tripPhase === "pickup" || state.tripPhase === "trip";
  const currentOrder = state.currentOrder || {};
  corePill.textContent = hasTrip ? (state.tripPhase === "pickup" ? "正在接驾" : "行程进行中") : "空闲";
  corePill.dataset.phase = state.tripPhase;
  tripEmpty.hidden = hasTrip;
  if (!hasTrip) {
    idleBadge.textContent = state.onlineStatus === 1 ? "空闲状态" : "休息状态";
    idleTitle.textContent = state.onlineStatus === 1 ? "没接单，等待派单" : "没接单，已下线";
    idleDesc.textContent = state.onlineStatus === 1
      ? "当前没有接驾或行程任务，系统可继续派发新订单。"
      : "当前没有接驾或行程任务，点击上线接单后进入等待派单。";
  }
  fromAddress.textContent = hasTrip ? (currentOrder.fromAddress || "等待订单详情") : "--";
  toAddress.textContent = hasTrip ? (currentOrder.toAddress || "等待订单详情") : "--";
  orderNo.textContent = hasTrip ? (currentOrder.orderNo || currentOrder.orderId || "--") : "--";
  orderStatusText.textContent = hasTrip ? formatOrderStatus(currentOrder.status) : "--";
  refreshTime.textContent = new Date().toLocaleTimeString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function setOrderButtonsDisabled(disabled) {
  document.querySelectorAll("[data-order-action]").forEach((button) => {
    button.disabled = disabled;
  });
}

function persistOnlineStatus() {
  if (!state.driver) {
    return;
  }
  state.driver = { ...state.driver, onlineStatus: state.onlineStatus };
  localStorage.setItem("driverProfile", JSON.stringify(state.driver));
}

function setOrderListLoading(loading) {
  document.querySelector("[data-refresh-orders]").disabled = loading;
  orderStatus.disabled = loading;
  orderPrevious.disabled = loading;
  orderNext.disabled = loading;
  if (loading) {
    orderList.innerHTML = '<div class="order-list-loading">正在加载订单...</div>';
    orderListEmpty.hidden = true;
  }
}

function renderOrders(data = {}) {
  const list = Array.isArray(data.list) ? data.list : [];
  const total = Number(data.total || 0);
  const page = Number(data.page || state.orderPage || 1);
  const pageSize = Number(data.pageSize || state.orderPageSize);

  state.orderPage = page;
  state.orderPageSize = pageSize;
  state.orderTotal = total;
  state.orders = list;
  syncCurrentTripFromOrders(list);
  orderStatus.value = String(state.orderStatus);
  orderList.innerHTML = list.map(renderOrderItem).join("");
  orderListEmpty.hidden = list.length !== 0;
  orderSummary.textContent = total ? `共 ${total} 条记录` : "暂无订单";
  orderPageLabel.textContent = `${page} / ${Math.max(1, Math.ceil(total / pageSize))}`;
  orderPrevious.disabled = page <= 1;
  orderNext.disabled = page * pageSize >= total;
}

function renderOrderItem(item) {
  const order = item || {};
  const orderId = Number(order.orderId || 0);
  const actions = renderOrderActions(order, orderId);
  const statusText = order.source === "dispatch" ? formatDispatchStatus(order.dispatchStatus) : formatOrderStatus(order.status);
  return `
    <article class="order-item">
      <div class="order-item-main">
        <div class="order-item-heading">
          <strong>${escapeHTML(order.orderNo || `订单 ${orderId || "--"}`)}</strong>
          <span class="status-tag" data-order-status="${order.status || 0}" data-dispatch-status="${order.dispatchStatus || 0}">${statusText}</span>
        </div>
        <div class="order-route">
          <span>${escapeHTML(order.fromAddress || "--")}</span>
          <span class="route-arrow">→</span>
          <span>${escapeHTML(order.toAddress || "--")}</span>
        </div>
        <div class="order-meta">
          <span>${statusText}</span>
          <span>${formatPrice(order.estimatedPriceCents)}</span>
          <span>${formatTimestamp(order.createdAt)}</span>
        </div>
      </div>
      <div class="order-item-actions">${actions}</div>
    </article>
  `;
}

function renderOrderActions(order, orderId) {
  if (!orderId) {
    return "";
  }
  if (order.source === "dispatch" || order.status === 1) {
    const acceptLabel = order.source === "dispatch" ? "抢单" : "接单";
    return [
      orderActionButton("accept", orderId, acceptLabel, "primary-button"),
      orderActionButton("reject", orderId, "拒单", "secondary-button"),
    ].join("");
  }
  if (order.status === 2) {
    return [
      orderActionButton("confirm-arrive", orderId, "确认到达", "secondary-button"),
      orderActionButton("start-trip", orderId, "开始行程", "secondary-button"),
    ].join("");
  }
  if (order.status === 3) {
    return orderActionButton("finish-trip", orderId, "结束行程", "primary-button");
  }
  return "";
}

function orderActionButton(action, orderId, label, className) {
  return `<button class="${className} compact order-row-button" type="button" data-order-action="${action}" data-order-id="${orderId}">${label}</button>`;
}

function formatDispatchStatus(status) {
  return {
    1: "待处理派单",
    2: "已接受派单",
    3: "已拒绝派单",
    4: "派单超时",
    5: "派单取消",
  }[status] || "派单状态未知";
}

function formatOrderStatus(status) {
  return {
    1: "订单待接单",
    2: "已接单",
    3: "行程中",
    4: "待支付",
    5: "已完成",
    6: "已取消",
  }[status] || "状态未知";
}

function formatPrice(cents) {
  const value = Number(cents);
  return Number.isFinite(value) ? `预估 ¥${(value / 100).toFixed(2)}` : "价格待定";
}

function formatTimestamp(timestamp) {
  const value = Number(timestamp);
  if (!value) {
    return "--";
  }
  return new Date(value * 1000).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function escapeHTML(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

async function requestJSON(url, options = {}) {
  const headers = {
    Accept: "application/json",
    ...(options.headers || {}),
  };
  if (options.body) {
    headers["Content-Type"] = "application/json";
  }
  if (options.token) {
    headers.Authorization = `Bearer ${options.token}`;
  }

  const response = await fetch(url, {
    method: options.method || "GET",
    headers,
    body: options.body,
  });
  const text = await response.text();
  const result = text ? JSON.parse(text) : {};

  if (!response.ok || result.code !== 0) {
    throw new Error(result.message || "请求失败");
  }
  return result;
}

function compactPayload(payload) {
  return Object.fromEntries(
    Object.entries(payload).filter(([, value]) => String(value || "").trim() !== "")
  );
}

function buildVehiclePayload(formData) {
  const payload = compactPayload(formData);
  payload.vehicleType = Number(payload.vehicleType || 0);
  if (payload.registrationDate) {
    payload.registrationDate = dateToUnixSeconds(payload.registrationDate);
  }
  if (payload.insuranceExpireAt) {
    payload.insuranceExpireAt = dateToUnixSeconds(payload.insuranceExpireAt);
  }
  return compactPayload(payload);
}

async function buildCertificationPayload(form) {
  const formData = new FormData(form);
  const payload = {
    vehicleId: Number(formData.get("vehicleId") || 0),
  };
  const fileFields = ["idCardFront", "idCardBack", "driverLicense", "vehicleLicense"];
  for (const field of fileFields) {
    const file = formData.get(field);
    if (file && file.size > 0) {
      payload[field] = await fileToBase64(file);
    }
  }
  if (!payload.vehicleId) {
    throw new Error("请先填写车辆ID");
  }
  if (!fileFields.some((field) => payload[field])) {
    throw new Error("请至少上传一张资质图片");
  }
  return payload;
}

function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result || ""));
    reader.onerror = () => reject(new Error("图片读取失败"));
    reader.readAsDataURL(file);
  });
}

function dateToUnixSeconds(value) {
  const time = new Date(`${value}T00:00:00`).getTime();
  return Number.isFinite(time) ? Math.floor(time / 1000) : 0;
}

function formatDriverStatus(status) {
  const map = {
    DRIVER_STATUS_PENDING: "待审核",
    DRIVER_STATUS_NORMAL: "正常",
    DRIVER_STATUS_FROZEN: "冻结",
    DRIVER_STATUS_CANCELLED: "注销",
  };
  return map[status] || status || "--";
}

function formatVehicleStatus(status) {
  const map = {
    VEHICLE_STATUS_PENDING: "待审核",
    VEHICLE_STATUS_NORMAL: "正常",
    VEHICLE_STATUS_DISABLED: "禁用",
  };
  return map[status] || status || "--";
}

function formatCertificationStatus(status) {
  const map = {
    1: "待审核",
    2: "已通过",
    3: "已驳回",
  };
  return map[Number(status || 0)] || "--";
}

function setAuthMessage(text, type) {
  message.textContent = text;
  message.classList.toggle("is-success", type === "success");
  message.classList.toggle("is-error", type === "error");
}

function setDashboardMessage(text, type) {
  dashboardMessage.textContent = text;
  dashboardMessage.classList.toggle("is-success", type === "success");
  dashboardMessage.classList.toggle("is-error", type === "error");
}

function setEditMessage(text, type) {
  editMessage.textContent = text;
  editMessage.classList.toggle("is-success", type === "success");
  editMessage.classList.toggle("is-error", type === "error");
}

function setVehicleMessage(text, type) {
  vehicleMessage.textContent = text;
  vehicleMessage.classList.toggle("is-success", type === "success");
  vehicleMessage.classList.toggle("is-error", type === "error");
}

function setCertificationMessage(text, type) {
  certificationMessage.textContent = text;
  certificationMessage.classList.toggle("is-success", type === "success");
  certificationMessage.classList.toggle("is-error", type === "error");
}

function setFinishMessage(text, type) {
  finishMessage.textContent = text;
  finishMessage.classList.toggle("is-success", type === "success");
  finishMessage.classList.toggle("is-error", type === "error");
}

function readJSON(key) {
  try {
    return JSON.parse(localStorage.getItem(key));
  } catch {
    return null;
  }
}
