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
const cachedDriver = readJSON("driverProfile") || null;

const state = {
  token: localStorage.getItem("driverToken") || "",
  driver: cachedDriver,
  onlineStatus: Number(cachedDriver?.onlineStatus ?? 0),
  tripPhase: localStorage.getItem("driverTripPhase") || "idle",
  currentOrderId: localStorage.getItem("driverCurrentOrderId") || "",
  orderPage: 1,
  orderPageSize: 8,
  orderStatus: 0,
  orderTotal: 0,
};

// 心跳定时器与间隔：用 var 声明避免 TDZ（防止函数先于 let 初始化被调用时报错）。
var heartbeatTimer = null;
var HEARTBEAT_INTERVAL = 15000;

buttons.forEach((button) => {
  button.addEventListener("click", () => {
    const target = button.dataset.authTarget;

    buttons.forEach((item) => item.classList.toggle("is-active", item === button));
    forms.forEach((form) => {
      form.classList.toggle("is-active", form.dataset.authForm === target);
    });

    title.textContent = target === "register" ? "司机注册" : "司机登录";
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

      if (form.dataset.authForm === "login") {
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

document.querySelectorAll("[data-close-edit]").forEach((button) => {
  button.addEventListener("click", closeEditModal);
});

editModal.addEventListener("click", (event) => {
  if (event.target === editModal) {
    closeEditModal();
  }
});

document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && !editModal.hidden) {
    closeEditModal();
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
  state.onlineStatus = 0;
  state.tripPhase = "idle";
  state.currentOrderId = "";
  localStorage.removeItem("driverToken");
  localStorage.removeItem("driverProfile");
  localStorage.removeItem("driverOnlineStatus");
  localStorage.removeItem("driverTripPhase");
  localStorage.removeItem("driverCurrentOrderId");
  dashboardView.hidden = true;
  authView.hidden = false;
  menuPanel.hidden = true;
}

document.querySelector("[data-refresh]").addEventListener("click", () => {
  loadDashboardData();
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

if (state.token && state.driver?.id) {
  showDashboard();
  loadDashboardData();
  if (state.onlineStatus === 1) {
    startHeartbeat();
  }
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
  } catch (error) {
    setDashboardMessage("主页数据加载失败", "error");
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
      localStorage.setItem("driverTripPhase", state.tripPhase);
      localStorage.removeItem("driverCurrentOrderId");
      stopHeartbeat();
    } else {
      startHeartbeat();
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
    "finish-trip": {
      path: "/driver/orders/finish-trip",
      phase: "idle",
      message: "行程已结束，当前空闲",
      payload: {
        orderId: orderID,
        actualDistanceM: 0,
        actualDurationS: 0,
        actualPriceCents: 0,
      },
    },
  }[action];

  if (!config) {
    return;
  }

  setOrderButtonsDisabled(true);
  setDashboardMessage("订单操作提交中...");
  try {
    await requestJSON(config.path, {
      method: "POST",
      token: state.token,
      body: JSON.stringify(config.payload),
    });
    state.currentOrderId = config.phase === "idle" ? "" : String(orderID);
    state.tripPhase = config.phase;
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
    } else {
      localStorage.removeItem("driverCurrentOrderId");
    }
    renderStatus();
    renderCoreArea();
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
  renderStatus();
  renderCoreArea();
}

function showPanel(target) {
  panels.forEach((panel) => {
    panel.hidden = panel.dataset.panel !== target;
  });
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

function renderCoreArea() {
  const corePill = document.querySelector("[data-core-pill]");
  const tripEmpty = document.querySelector("[data-trip-empty]");
  const idleBadge = document.querySelector("[data-idle-badge]");
  const idleTitle = document.querySelector("[data-idle-title]");
  const idleDesc = document.querySelector("[data-idle-desc]");
  const fromAddress = document.querySelector("[data-from-address]");
  const toAddress = document.querySelector("[data-to-address]");

  const hasTrip = state.tripPhase === "pickup" || state.tripPhase === "trip";
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
  fromAddress.textContent = hasTrip ? "等待订单详情" : "--";
  toAddress.textContent = hasTrip ? "等待订单详情" : "--";
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
    return [
      orderActionButton("accept", orderId, "接单", "primary-button"),
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

function formatDriverStatus(status) {
  const map = {
    DRIVER_STATUS_PENDING: "待审核",
    DRIVER_STATUS_NORMAL: "正常",
    DRIVER_STATUS_FROZEN: "冻结",
    DRIVER_STATUS_CANCELLED: "注销",
  };
  return map[status] || status || "--";
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

function readJSON(key) {
  try {
    return JSON.parse(localStorage.getItem(key));
  } catch {
    return null;
  }
}
