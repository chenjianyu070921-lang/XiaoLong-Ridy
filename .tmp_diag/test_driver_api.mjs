const base = process.env.DRIVER_API_BASE || "http://127.0.0.1:8082/api/driver/v1";
const password = process.env.DRIVER_TEST_PASSWORD || "Driver@123";
const fixedPhone = process.env.DRIVER_TEST_PHONE || "";
const fixedToken = process.env.DRIVER_TEST_TOKEN || "";
const fixedDriverId = Number(process.env.DRIVER_TEST_DRIVER_ID || 0);

function uniquePhone() {
  const suffix = Date.now().toString().slice(-8);
  return `139${suffix}`;
}

async function call(name, method, path, body, token) {
  const headers = { "content-type": "application/json" };
  if (token) headers.authorization = `Bearer ${token}`;
  const started = Date.now();
  let status = 0;
  let text = "";
  let json = null;
  let ok = false;
  try {
    const resp = await fetch(`${base}${path}`, {
      method,
      headers,
      body: body == null ? undefined : JSON.stringify(body),
    });
    status = resp.status;
    text = await resp.text();
    try {
      json = text ? JSON.parse(text) : null;
    } catch {
      json = null;
    }
    ok = resp.ok && (!json || json.code === 0 || json.code === 200);
  } catch (err) {
    text = err?.message || String(err);
  }
  return {
    name,
    method,
    path,
    status,
    code: json?.code ?? null,
    message: json?.message ?? text.slice(0, 200),
    ok,
    ms: Date.now() - started,
    data: json?.data ?? null,
  };
}

function printResult(result) {
  const state = result.ok ? "PASS" : "FAIL";
  const expected = result.expectedError ? " expected-error" : "";
  console.log(`${state}${expected} ${result.method} ${result.path} [${result.name}] http=${result.status} code=${result.code} ${result.ms}ms`);
  if (!result.ok) {
    console.log(`  message: ${result.message}`);
  }
}

const results = [];

async function main() {
  const phone = fixedPhone || uniquePhone();
  const registerBody = {
    phone,
    password,
    realName: "ApiTestDriver",
    idCardNo: `11010119900101${Date.now().toString().slice(-4)}`,
    driverLicenseNo: `DL${Date.now()}`,
    avatarUrl: "",
  };

  const setupSteps = fixedToken && fixedDriverId
    ? []
    : fixedPhone
    ? [["login password", "POST", "/auth/login-by-password", { phone, password }]]
    : [
        ["send sms", "POST", "/auth/send-sms-code", { phone }],
        ["register", "POST", "/drivers/register", registerBody],
        ["login password", "POST", "/auth/login-by-password", { phone, password }],
      ];

  for (const step of setupSteps) {
    const result = await call(...step);
    results.push(result);
    printResult(result);
  }

  const login = results.find((r) => r.name === "login password");
  const token = fixedToken || login?.data?.token;
  const driverId = fixedDriverId || login?.data?.driver?.id;
  if (!token || !driverId) {
    console.log("SUMMARY " + JSON.stringify({ phone, driverId: driverId || 0, token: false, results }, null, 2));
    process.exitCode = 1;
    return;
  }

  const deviceId = `api-test-${driverId}`;
  const validPos = { deviceId, longitude: 116.397428, latitude: 39.90923 };
  const safeSteps = [
    ["get driver", "GET", "/drivers/get", null],
    ["ai score", "GET", "/drivers/ai-score", null],
    ["certification get", "GET", "/drivers/certification", null],
    ["online reject zero coord", "POST", "/drivers/online", { deviceId, longitude: 0, latitude: 0 }],
    ["online valid coord", "POST", "/drivers/online", validPos],
    ["location reject zero coord", "POST", "/drivers/location/report", { deviceId, longitude: 0, latitude: 0 }],
    ["location valid coord", "POST", "/drivers/location/report", { ...validPos, heading: 90, speedKmh: 12.5 }],
    ["heartbeat", "POST", "/drivers/heartbeat", { ...validPos }],
    ["nearby drivers", "POST", "/drivers/nearby", { longitude: validPos.longitude, latitude: validPos.latitude, radiusMeters: 3000, limit: 5 }],
    ["available orders", "POST", "/orders/available", { page: 1, pageSize: 10 }],
    ["my order list", "POST", "/orders/list", { page: 1, pageSize: 10, status: 0 }],
    ["my dispatches", "POST", "/orders/dispatches", { page: 1, pageSize: 10, status: 0 }],
    ["income summary", "GET", "/income/summary", null],
    ["income today", "GET", "/income/today", null],
    ["income week", "GET", "/income/week", null],
    ["income bills", "POST", "/income/bills", { page: 1, pageSize: 10 }],
    ["withdraws list", "POST", "/withdraws/list", { page: 1, pageSize: 10 }],
    ["reviews list", "POST", "/reviews/list", { page: 1, pageSize: 10 }],
  ];
  if (process.env.DRIVER_TEST_OFFLINE === "1") {
    safeSteps.push(["offline", "POST", "/drivers/offline", { ...validPos }]);
  }

  for (const step of safeSteps) {
    const result = await call(...step, token);
    if (result.name.includes("reject zero coord")) {
      result.expectedError = true;
      result.ok = result.status === 400 && result.code !== 0;
    }
    results.push(result);
    printResult(result);
  }

  const available = results.find((r) => r.name === "available orders");
  const firstOrderId = available?.data?.list?.[0]?.orderId || 0;
  if (firstOrderId > 0) {
    for (const step of [
      ["order detail", "POST", "/orders/detail", { orderId: firstOrderId }],
      ["order trajectory", "POST", "/orders/trajectory", { orderId: firstOrderId }],
    ]) {
      const result = await call(...step, token);
      results.push(result);
      printResult(result);
    }
  } else {
    console.log("SKIP POST /orders/detail [order detail] no available order");
    console.log("SKIP POST /orders/trajectory [order trajectory] no available order");
  }

  const failed = results.filter((r) => !r.ok);
  console.log("SUMMARY " + JSON.stringify({
    base,
    phone,
    driverId,
    pass: results.length - failed.length,
    fail: failed.length,
    failed: failed.map((r) => ({ name: r.name, method: r.method, path: r.path, status: r.status, code: r.code, message: r.message })),
    availableOrders: available?.data?.list?.length ?? null,
  }, null, 2));
  if (failed.length) process.exitCode = 1;
}

main();
