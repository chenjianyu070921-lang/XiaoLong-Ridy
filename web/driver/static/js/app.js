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

const state = {
  token: localStorage.getItem("driverToken") || "",
  driver: readJSON("driverProfile") || null,
  onlineStatus: Number(localStorage.getItem("driverOnlineStatus") || 0),
  tripPhase: localStorage.getItem("driverTripPhase") || "idle",
  currentOrderId: localStorage.getItem("driverCurrentOrderId") || "",
};

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

document.querySelector("[data-logout]").addEventListener("click", () => {
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
});

document.querySelector("[data-refresh]").addEventListener("click", () => {
  loadDashboardData();
});

document.querySelector("[data-online]").addEventListener("click", async () => {
  await setWorkStatus("/driver/online", 1, "已上线，当前空闲中");
});

document.querySelector("[data-offline]").addEventListener("click", async () => {
  await setWorkStatus("/driver/offline", 0, "已下线休息");
});

document.querySelectorAll("[data-order-action]").forEach((button) => {
  button.addEventListener("click", () => {
    handleOrderAction(button.dataset.orderAction);
  });
});

document.querySelector("[data-edit-form]").addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!state.driver?.id) {
    setDashboardMessage("司机信息缺失，请重新登录", "error");
    return;
  }

  const form = event.currentTarget;
  const payload = compactPayload(Object.fromEntries(new FormData(form).entries()));
  payload.id = state.driver.id;

  try {
    await requestJSON(form.action, {
      method: "POST",
      token: state.token,
      body: JSON.stringify(payload),
    });
    setDashboardMessage("司机信息已更新", "success");
    showPanel("profile");
    await loadDashboardData();
  } catch (error) {
    setDashboardMessage(error.message || "保存失败", "error");
  }
});

if (state.token && state.driver?.id) {
  showDashboard();
  loadDashboardData();
} else {
  authView.hidden = false;
  dashboardView.hidden = true;
}

async function loadDashboardData() {
  if (!state.token || !state.driver?.id) {
    return;
  }

  try {
    const [profileResult, scoreResult] = await Promise.allSettled([
      requestJSON(`/driver/me?driverId=${encodeURIComponent(state.driver.id)}`, {
        method: "GET",
        token: state.token,
      }),
      requestJSON(`/driver/ai-score?driverId=${encodeURIComponent(state.driver.id)}`, {
        method: "GET",
        token: state.token,
      }),
    ]);

    if (profileResult.status === "fulfilled") {
      const driver = profileResult.value.data.driver;
      state.driver = { ...state.driver, ...driver };
      localStorage.setItem("driverProfile", JSON.stringify(state.driver));
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

    renderStatus();
  } catch (error) {
    setDashboardMessage("主页数据加载失败", "error");
  }
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
    if (state.onlineStatus === 0) {
      state.tripPhase = "idle";
      state.currentOrderId = "";
      localStorage.setItem("driverTripPhase", state.tripPhase);
      localStorage.removeItem("driverCurrentOrderId");
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

async function handleOrderAction(action) {
  const orderInput = document.querySelector("[data-order-id]");
  const orderID = Number(orderInput.value || state.currentOrderId);
  if (!orderID || orderID <= 0) {
    setDashboardMessage("请先输入订单ID", "error");
    orderInput.focus();
    return;
  }

  const config = {
    accept: {
      path: "/driver/orders/accept",
      phase: "pickup",
      message: "接单成功，状态已切换为正在接驾",
      payload: { orderId: orderID },
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
    }
    if (config.phase === "idle" && state.onlineStatus === 2) {
      state.onlineStatus = 1;
      localStorage.setItem("driverOnlineStatus", "1");
    }
    localStorage.setItem("driverTripPhase", state.tripPhase);
    if (state.currentOrderId) {
      localStorage.setItem("driverCurrentOrderId", state.currentOrderId);
    } else {
      localStorage.removeItem("driverCurrentOrderId");
      orderInput.value = "";
    }
    renderStatus();
    renderCoreArea();
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

function renderDriver() {
  const driver = state.driver || {};
  const name = driver.realName || "司机";
  const phone = driver.phone || "--";
  const avatar = driver.avatarUrl || "";

  document.querySelector("[data-driver-name]").textContent = name;
  document.querySelector("[data-driver-phone]").textContent = phone;
  document.querySelector("[data-driver-initial]").textContent = name.slice(0, 1) || "司";
  document.querySelector("[data-profile-id]").textContent = driver.id || "--";
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
  const orderInput = document.querySelector("[data-order-id]");
  const corePill = document.querySelector("[data-core-pill]");
  const tripEmpty = document.querySelector("[data-trip-empty]");
  const idleBadge = document.querySelector("[data-idle-badge]");
  const idleTitle = document.querySelector("[data-idle-title]");
  const idleDesc = document.querySelector("[data-idle-desc]");
  const fromAddress = document.querySelector("[data-from-address]");
  const toAddress = document.querySelector("[data-to-address]");

  if (state.currentOrderId && !orderInput.value) {
    orderInput.value = state.currentOrderId;
  }

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

function readJSON(key) {
  try {
    return JSON.parse(localStorage.getItem(key));
  } catch {
    return null;
  }
}
