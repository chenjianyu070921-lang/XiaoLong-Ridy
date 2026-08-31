import crypto from "node:crypto";
import fs from "node:fs";
import net from "node:net";
import { execFileSync } from "node:child_process";

const base = process.env.DRIVER_API_BASE || "http://127.0.0.1:8082/api/driver/v1";
const mysqlExe = process.env.MYSQL_EXE || "D:\\phpstudy_pro\\Extensions\\MySQL5.7.26\\bin\\mysql.exe";
const mysqlArgs = [
  "--ssl-mode=DISABLED",
  "-h",
  process.env.MYSQL_HOST || "115.191.16.159",
  "-P",
  process.env.MYSQL_PORT || "3306",
  "-uroot",
  `-p${process.env.MYSQL_PASSWORD || "4ay1nkal3u8ed77y"}`,
  process.env.MYSQL_DATABASE || "xiaolong_ridy",
  "--batch",
  "--raw",
  "--skip-column-names",
];
const driverPhone = process.env.DRIVER_TEST_PHONE || "13900009991";
const driverPassword = process.env.DRIVER_TEST_PASSWORD || "Driver@123";
const driverID = Number(process.env.DRIVER_TEST_DRIVER_ID || 3);
const signingKey = process.env.DRIVER_SIGNING_KEY || "local-development-signing-key";
const onePixelPng =
  "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=";

const created = {
  orders: [],
  dispatches: [],
  withdraws: [],
  certifications: [],
  vehicles: [],
  registeredDrivers: [],
};
const results = [];

function signDriverToken(accountID = driverID, accountStatus = 2) {
  const now = Math.floor(Date.now() / 1000);
  const enc = (value) => Buffer.from(JSON.stringify(value)).toString("base64url");
  const header = { alg: "HS256", typ: "JWT" };
  const payload = {
    sub: `driver_${accountID}`,
    accountId: accountID,
    accountType: "driver",
    accountStatus,
    phone: "139****9991",
    role: "driver",
    iat: now,
    exp: now + 7200,
    iss: "driversvc",
  };
  const unsigned = `${enc(header)}.${enc(payload)}`;
  const signature = crypto.createHmac("sha256", signingKey).update(unsigned).digest("base64url");
  return `${unsigned}.${signature}`;
}

function sqlQuote(value) {
  return `'${String(value).replaceAll("\\", "\\\\").replaceAll("'", "''")}'`;
}

function mysql(sql) {
  return execFileSync(mysqlExe, [...mysqlArgs, "-e", sql], {
    encoding: "utf8",
    stdio: ["ignore", "pipe", "pipe"],
  }).trim();
}

function insertOrder(label, lon = 116.378, lat = 39.865) {
  const orderNo = `T${Date.now()}${label}`.slice(0, 31);
  const sql = `
    INSERT INTO ride_order
      (order_no,user_id,driver_id,car_type,city_code,from_address,from_longitude,from_latitude,to_address,to_longitude,to_latitude,estimated_distance_m,estimated_duration_s,estimated_price,status)
    VALUES
      (${sqlQuote(orderNo)},1,0,1,'110000',${sqlQuote(`driver-api-test-${label}`)},${lon},${lat},${sqlQuote(`driver-api-test-to-${label}`)},${lon + 0.01},${lat + 0.01},2500,600,34.40,1);
    SELECT LAST_INSERT_ID();
  `;
  const id = Number(mysql(sql).split(/\s+/).pop());
  created.orders.push(id);
  return id;
}

function insertDispatch(orderID) {
  const sql = `
    INSERT INTO dispatch_record (order_id,driver_id,dispatch_type,status,match_score,remark,reject_reason)
    VALUES (${orderID},${driverID},1,1,95.5,'driver api test','');
    SELECT LAST_INSERT_ID();
  `;
  const id = Number(mysql(sql).split(/\s+/).pop());
  created.dispatches.push(id);
  return id;
}

async function call(name, method, path, body, token, options = {}) {
  const headers = { "content-type": "application/json" };
  if (token) headers.authorization = `Bearer ${token}`;
  const started = Date.now();
  let status = 0;
  let text = "";
  let json = null;
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
  } catch (err) {
    text = err?.message || String(err);
  }
  const expectStatus = options.expectStatus || 200;
  const expectCode = options.expectCode ?? 0;
  const ok = status === expectStatus && (expectCode === undefined || (json?.code ?? null) === expectCode);
  const result = {
    name,
    method,
    path,
    status,
    code: json?.code ?? null,
    message: json?.message || text.slice(0, 200),
    ok,
    ms: Date.now() - started,
    data: json?.data ?? null,
    expected: `${expectStatus}/${expectCode}`,
  };
  results.push(result);
  print(result);
  return result;
}

async function callStatic(name, path) {
  const started = Date.now();
  const resp = await fetch(`http://127.0.0.1:8082${path}`);
  const ok = resp.status === 200;
  const result = {
    name,
    method: "GET",
    path,
    status: resp.status,
    code: null,
    message: ok ? "static file served" : await resp.text(),
    ok,
    ms: Date.now() - started,
    expected: "200/static",
  };
  results.push(result);
  print(result);
  return result;
}

function print(result) {
  const state = result.ok ? "PASS" : "FAIL";
  console.log(`${state} ${result.method} ${result.path} [${result.name}] http=${result.status} code=${result.code} ${result.ms}ms`);
  if (!result.ok) console.log(`  expected=${result.expected} message=${result.message}`);
}

function parseSMSCode(phone) {
  const logPath = ".run/logs/api-driver.err.log";
  const deadline = Date.now() + 3000;
  const pattern = new RegExp(`phone=${phone} code=(\\d{6})`, "g");
  while (Date.now() < deadline) {
    const text = fs.existsSync(logPath) ? fs.readFileSync(logPath, "utf8") : "";
    const matches = [...text.matchAll(pattern)];
    if (matches.length) return matches.at(-1)[1];
    Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, 100);
  }
  return "";
}

function wsCheck(token) {
  return new Promise((resolve) => {
    const url = new URL(base.replace("http:", "ws:") + `/ws?token=${encodeURIComponent(token)}`);
    const key = crypto.randomBytes(16).toString("base64");
    const socket = net.createConnection({ host: url.hostname, port: Number(url.port || 80) });
    let data = "";
    const started = Date.now();
    const done = (ok, status = 0, message = "") => {
      socket.destroy();
      const result = {
        name: "websocket",
        method: "GET",
        path: "/ws",
        status,
        code: null,
        message,
        ok,
        ms: Date.now() - started,
        expected: "101 websocket + connected/auth message",
      };
      results.push(result);
      print(result);
      resolve(result);
    };
    socket.setTimeout(5000, () => done(false, 0, "websocket timeout"));
    socket.on("error", (err) => done(false, 0, err.message));
    socket.on("connect", () => {
      socket.write(
        [
          `GET ${url.pathname}${url.search} HTTP/1.1`,
          `Host: ${url.host}`,
          "Upgrade: websocket",
          "Connection: Upgrade",
          `Sec-WebSocket-Key: ${key}`,
          "Sec-WebSocket-Version: 13",
          "",
          "",
        ].join("\r\n"),
      );
    });
    socket.on("data", (chunk) => {
      data += chunk.toString("binary");
      const statusLine = data.split("\r\n", 1)[0] || "";
      const status = Number(statusLine.match(/HTTP\/1\.1\s+(\d+)/)?.[1] || 0);
      if (status && status !== 101) done(false, status, statusLine);
      if (status === 101 && data.includes("connected")) done(true, 101, "connected");
      if (status === 101 && data.includes("auth_failed")) done(false, 101, "auth_failed");
    });
  });
}

async function main() {
  const validToken = signDriverToken();
  const now = Date.now();
  const registerPhone = `139${String(now).slice(-8)}`;
  const registerBody = {
    phone: registerPhone,
    password: driverPassword,
    realName: "ApiCheckDriver",
    idCardNo: `11010119900101${String(now).slice(-4)}`,
    driverLicenseNo: `DL${now}`,
    avatarUrl: "",
  };

  await call("send sms", "POST", "/auth/send-sms-code", { phone: driverPhone });
  const smsCode = parseSMSCode(driverPhone);
  if (smsCode) {
    await call("login by sms", "POST", "/auth/login-by-sms", { phone: driverPhone, code: smsCode });
  } else {
    results.push({
      name: "login by sms",
      method: "POST",
      path: "/auth/login-by-sms",
      status: 0,
      code: null,
      message: "sms code not found in api-driver log",
      ok: false,
      ms: 0,
      expected: "200/0",
    });
    print(results.at(-1));
  }
  await call("login by password", "POST", "/auth/login-by-password", { phone: driverPhone, password: driverPassword });
  const registerResp = await call("register", "POST", "/drivers/register", registerBody);
  if (registerResp.data?.id) created.registeredDrivers.push(registerResp.data.id);

  const driver = await call("get driver", "GET", "/drivers/get", null, validToken);
  const originalAvatar = driver.data?.driver?.avatarUrl || "";
  await call("update driver", "POST", "/drivers/update", { id: driverID, avatarUrl: `api-check-${now}` }, validToken);
  await call("ai score", "GET", "/drivers/ai-score", null, validToken);

  const vehicle = await call(
    "create vehicle",
    "POST",
    "/vehicles",
    {
      plateNo: `粤B${String(now).slice(-5)}`,
      brand: "ApiBrand",
      model: "ApiModel",
      color: "White",
      vehicleType: 1,
      insuranceNo: `INS${now}`,
    },
    validToken,
  );
  const vehicleID = vehicle.data?.id;
  if (vehicleID) {
    created.vehicles.push(vehicleID);
    await call("get vehicle", "GET", `/vehicles/get?id=${vehicleID}`, null, validToken);
    await call("update vehicle", "POST", "/vehicles/update", { id: vehicleID, color: "Black", model: "ApiModelUpdated" }, validToken);
    const cert = await call(
      "upload certification",
      "POST",
      "/drivers/certification/upload",
      {
        vehicleId: vehicleID,
        idCardFront: onePixelPng,
        idCardBack: onePixelPng,
        driverLicense: onePixelPng,
        vehicleLicense: onePixelPng,
      },
      validToken,
    );
    if (cert.data?.id) created.certifications.push(cert.data.id);
    const fileURL = cert.data?.certification?.idCardFrontUrl;
    if (fileURL) await callStatic("certification static file", fileURL);
  }
  await call("get certification", "GET", "/drivers/certification", null, validToken);

  const pos = { deviceId: `api-check-${driverID}`, longitude: 116.378, latitude: 39.865 };
  await call("online rejects zero", "POST", "/drivers/online", { deviceId: pos.deviceId, longitude: 0, latitude: 0 }, validToken, {
    expectStatus: 400,
    expectCode: 50000,
  });
  await call("online", "POST", "/drivers/online", pos, validToken);
  await call("heartbeat", "POST", "/drivers/heartbeat", pos, validToken);
  await call("location rejects zero", "POST", "/drivers/location/report", { deviceId: pos.deviceId, longitude: 0, latitude: 0 }, validToken, {
    expectStatus: 400,
    expectCode: 50000,
  });
  await call("location report", "POST", "/drivers/location/report", { ...pos, heading: 88, speedKmh: 20 }, validToken);
  await call("nearby drivers", "POST", "/drivers/nearby", { longitude: pos.longitude, latitude: pos.latitude, radiusMeters: 3000, limit: 5 }, validToken);

  const rejectOrderID = insertOrder("R");
  insertDispatch(rejectOrderID);
  const flowOrderID = insertOrder("F");
  insertDispatch(flowOrderID);
  const cancelOrderID = insertOrder("C");
  insertDispatch(cancelOrderID);

  await call("available orders", "POST", "/orders/available", { page: 1, pageSize: 10 }, validToken);
  await call("order detail wait accept", "POST", "/orders/detail", { orderId: flowOrderID }, validToken);
  await call("reject order", "POST", "/orders/reject", { orderId: rejectOrderID, reason: "api check reject" }, validToken);

  await call("accept order", "POST", "/orders/accept", { orderId: flowOrderID }, validToken);
  await call("order detail accepted", "POST", "/orders/detail", { orderId: flowOrderID }, validToken);
  await call("confirm arrive", "POST", "/orders/confirm-arrive", { orderId: flowOrderID }, validToken);
  await call("start trip", "POST", "/orders/start-trip", { orderId: flowOrderID }, validToken);
  await call("location trajectory point", "POST", "/drivers/location/report", { ...pos, orderId: flowOrderID, heading: 90, speedKmh: 35 }, validToken);
  await call("order trajectory", "POST", "/orders/trajectory", { orderId: flowOrderID }, validToken);
  await call("finish trip", "POST", "/orders/finish-trip", { orderId: flowOrderID, actualDistanceM: 2500, actualDurationS: 600 }, validToken);

  await call("accept cancel order", "POST", "/orders/accept", { orderId: cancelOrderID }, validToken);
  await call("cancel order", "POST", "/orders/cancel", { orderId: cancelOrderID, reason: "api check cancel" }, validToken);

  await call("my order list", "POST", "/orders/list", { page: 1, pageSize: 10, status: 0 }, validToken);
  await call("my dispatches", "POST", "/orders/dispatches", { page: 1, pageSize: 10, status: 0 }, validToken);
  await call("income summary", "GET", "/income/summary", null, validToken);
  await call("income today", "GET", "/income/today", null, validToken);
  await call("income week", "GET", "/income/week", null, validToken);
  await call("income bills", "POST", "/income/bills", { page: 1, pageSize: 10 }, validToken);

  const withdraw = await call("create withdraw", "POST", "/withdraws", { amount: 0.01, payeeName: "Api Check", payAccount: `acct-${now}` }, validToken);
  if (withdraw.data?.id) created.withdraws.push(withdraw.data.id);
  await call("withdraws list", "POST", "/withdraws/list", { page: 1, pageSize: 10 }, validToken);
  await call("reviews list", "POST", "/reviews/list", { page: 1, pageSize: 10 }, validToken);
  await call("agent chat", "POST", "/agent/chat", { question: "driver api health check" }, validToken);
  await wsCheck(validToken);
  await call("offline", "POST", "/drivers/offline", pos, validToken);
  await call("online restore", "POST", "/drivers/online", pos, validToken);

  if (vehicleID) {
    await call("delete vehicle", "POST", "/vehicles/delete", { id: vehicleID }, validToken);
  }
  await call("restore driver avatar", "POST", "/drivers/update", { id: driverID, avatarUrl: originalAvatar }, validToken);
}

function cleanup() {
  const statements = [];
  if (created.dispatches.length) statements.push(`DELETE FROM dispatch_record WHERE id IN (${created.dispatches.join(",")})`);
  if (created.orders.length) statements.push(`UPDATE ride_order SET deleted_at = NOW() WHERE id IN (${created.orders.join(",")})`);
  if (created.withdraws.length) statements.push(`DELETE FROM driver_withdraw WHERE id IN (${created.withdraws.join(",")})`);
  if (created.certifications.length) statements.push(`DELETE FROM driver_certification WHERE id IN (${created.certifications.join(",")})`);
  if (created.vehicles.length) statements.push(`DELETE FROM driver_vehicle WHERE id IN (${created.vehicles.join(",")})`);
  if (created.registeredDrivers.length) statements.push(`UPDATE driver SET deleted_at = NOW() WHERE id IN (${created.registeredDrivers.join(",")})`);
  if (!statements.length) return;
  if (created.orders.length) statements.unshift(`DELETE FROM order_status_log WHERE order_id IN (${created.orders.join(",")})`);
  try {
    mysql(`${statements.join(";")};`);
  } catch (err) {
    console.log(`CLEANUP_FAIL ${err.message}`);
  }
}

try {
  await main();
} finally {
  cleanup();
  const failed = results.filter((r) => !r.ok);
  console.log(
    "SUMMARY " +
      JSON.stringify(
        {
          base,
          driverID,
          total: results.length,
          pass: results.length - failed.length,
          fail: failed.length,
          failed: failed.map((r) => ({
            name: r.name,
            method: r.method,
            path: r.path,
            status: r.status,
            code: r.code,
            message: r.message,
            expected: r.expected,
          })),
          created,
        },
        null,
        2,
      ),
  );
  if (failed.length) process.exitCode = 1;
}
