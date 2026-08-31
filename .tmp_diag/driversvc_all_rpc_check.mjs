import fs from "node:fs";
import path from "node:path";
import { execFileSync } from "node:child_process";

const grpcurl = process.env.GRPCURL_EXE || "D:\\gocode\\bin\\grpcurl.exe";
const target = process.env.DRIVERSVC_GRPC_ADDR || "127.0.0.1:50055";
const protoDir = process.env.DRIVERSVC_PROTO_DIR || "rpc\\driversvc\\proto";
const protoFile = process.env.DRIVERSVC_PROTO_FILE || "driversvc.proto";
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

const driverID = Number(process.env.DRIVER_TEST_DRIVER_ID || 3);
const driverPhone = process.env.DRIVER_TEST_PHONE || "13900009991";
const driverPassword = process.env.DRIVER_TEST_PASSWORD || "Driver@123";
const deviceID = process.env.DRIVER_TEST_DEVICE_ID || `rpc-check-${driverID}`;
const pos = {
  longitude: Number(process.env.DRIVER_TEST_LON || 116.378),
  latitude: Number(process.env.DRIVER_TEST_LAT || 39.865),
};
const onePixelPng =
  "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=";

const created = {
  drivers: [],
  vehicles: [],
  certifications: [],
  withdraws: [],
};
const results = [];

function sqlQuote(value) {
  return `'${String(value).replaceAll("\\", "\\\\").replaceAll("'", "''")}'`;
}

function mysql(sql) {
  return execFileSync(mysqlExe, [...mysqlArgs, "-e", sql], {
    encoding: "utf8",
    stdio: ["ignore", "pipe", "pipe"],
  }).trim();
}

function parseID(value) {
  const n = Number(value);
  return Number.isFinite(n) ? n : 0;
}

function intField(value, defaultValue = 0) {
  if (value === undefined || value === null) return defaultValue;
  const n = Number(value);
  return Number.isFinite(n) ? n : defaultValue;
}

function mkDriver(prefix) {
  const seed = `${Date.now()}${Math.floor(Math.random() * 10000)}`;
  return {
    phone: `139${seed.slice(-8)}`,
    passwordHash: driverPassword,
    realName: `${prefix}Driver`,
    idCardNo: `110101199001${seed.slice(-6)}`,
    driverLicenseNo: `DL${prefix}${seed}`.slice(0, 32),
    avatarUrl: "",
  };
}

function mkVehicle(driverId, prefix) {
  const seed = `${Date.now()}${Math.floor(Math.random() * 10000)}`;
  return {
    driverId,
    plateNo: `粤B${seed.slice(-5)}`,
    brand: `${prefix}Brand`,
    model: `${prefix}Model`,
    color: "White",
    vehicleType: 1,
    insuranceNo: `INS${prefix}${seed}`.slice(0, 32),
  };
}

function grpc(method, body, { record = true, expectError = false, assert = null, name = method } = {}) {
  const started = Date.now();
  let stdout = "";
  let stderr = "";
  let exitCode = 0;
  let json = null;
  try {
    stdout = execFileSync(
      grpcurl,
      ["-plaintext", "-import-path", protoDir, "-proto", protoFile, "-d", "@", target, `driversvc.DriverService/${method}`],
      {
        input: JSON.stringify(body ?? {}),
        encoding: "utf8",
        stdio: ["pipe", "pipe", "pipe"],
        timeout: 15000,
      },
    );
    if (stdout.trim()) json = JSON.parse(stdout);
  } catch (err) {
    exitCode = err.status || 1;
    stdout = err.stdout?.toString?.() || "";
    stderr = err.stderr?.toString?.() || err.message || String(err);
  }

  let ok = expectError ? exitCode !== 0 : exitCode === 0;
  let message = expectError ? stderr.trim() || stdout.trim() : stderr.trim();
  if (ok && assert) {
    try {
      assert(json);
    } catch (err) {
      ok = false;
      message = err.message;
    }
  }

  const result = {
    name,
    method,
    ok,
    exitCode,
    ms: Date.now() - started,
    message,
    response: json,
  };
  if (record) {
    results.push(result);
    print(result);
  }
  if (!ok) {
    const error = new Error(`${method} failed: ${message || stdout}`);
    error.result = result;
    throw error;
  }
  return json;
}

function print(result) {
  const state = result.ok ? "PASS" : "FAIL";
  console.log(`${state} RPC ${result.method} [${result.name}] exit=${result.exitCode} ${result.ms}ms`);
  if (!result.ok) console.log(`  message=${result.message}`);
}

function setupDriver(prefix) {
  const resp = grpc("CreateDriver", mkDriver(prefix), { record: false });
  const id = parseID(resp?.id);
  if (!id) throw new Error(`setup driver failed: ${JSON.stringify(resp)}`);
  created.drivers.push(id);
  return id;
}

function setupVehicle(driverId, prefix) {
  const resp = grpc("CreateVehicle", mkVehicle(driverId, prefix), { record: false });
  const id = parseID(resp?.id);
  if (!id) throw new Error(`setup vehicle failed: ${JSON.stringify(resp)}`);
  created.vehicles.push(id);
  return id;
}

function setupCertification(driverId, vehicleId) {
  const urlPrefix = `/rpc-check/drivers/${driverId}/${Date.now()}`;
  const sql = `
    INSERT INTO driver_certification
      (driver_id,vehicle_id,id_card_front_url,id_card_back_url,driver_license_url,vehicle_license_url,audit_status,audit_remark,audited_by)
    VALUES
      (${driverId},${vehicleId},${sqlQuote(`${urlPrefix}/front.png`)},${sqlQuote(`${urlPrefix}/back.png`)},${sqlQuote(`${urlPrefix}/driver.png`)},${sqlQuote(`${urlPrefix}/vehicle.png`)},1,'',0);
    SELECT LAST_INSERT_ID();
  `;
  const id = Number(mysql(sql).split(/\s+/).pop());
  if (!id) throw new Error("setup certification failed: missing inserted id");
  created.certifications.push(id);
  return id;
}

async function main() {
  const createResp = grpc("CreateDriver", mkDriver("Create"), {
    assert: (resp) => {
      if (!parseID(resp?.id)) throw new Error("missing created driver id");
    },
  });
  const createDriverID = parseID(createResp.id);
  created.drivers.push(createDriverID);

  const registerResp = grpc("RegisterDriver", mkDriver("Register"), {
    assert: (resp) => {
      if (!parseID(resp?.id)) throw new Error("missing registered driver id");
    },
  });
  const registerDriverID = parseID(registerResp.id);
  created.drivers.push(registerDriverID);

  grpc("UpdateDriver", { id: createDriverID, avatarUrl: `rpc-check-${Date.now()}` }, {
    assert: (resp) => {
      if (parseID(resp?.id) !== createDriverID) throw new Error("updated driver id mismatch");
    },
  });
  grpc("GetDriver", { id: driverID }, {
    assert: (resp) => {
      if (parseID(resp?.driver?.id) !== driverID) throw new Error("driver id mismatch");
    },
  });
  grpc("GetDriverByPhone", { phone: driverPhone }, {
    assert: (resp) => {
      if (resp?.driver?.phone !== driverPhone) throw new Error("driver phone mismatch");
    },
  });
  grpc("ListDrivers", { page: 1, pageSize: 10, keyword: driverPhone }, {
    assert: (resp) => {
      if (parseID(resp?.total) < 1) throw new Error("expected at least one driver");
    },
  });
  grpc("Login", { phone: driverPhone, password: driverPassword }, {
    assert: (resp) => {
      if (!resp?.token) throw new Error("missing login token");
    },
  });
  grpc("LoginBySms", { phone: driverPhone }, {
    assert: (resp) => {
      if (!resp?.token) throw new Error("missing sms login token");
    },
  });

  grpc("SetDriverOnline", { driverId: driverID, deviceId: deviceID, ...pos }, {
    assert: (resp) => {
      if (parseID(resp?.driverId) !== driverID || Number(resp?.onlineStatus) !== 1) throw new Error("online response mismatch");
    },
  });
  grpc("SetDriverOnline", { driverId: driverID, deviceId: `${deviceID}-zero`, longitude: 0, latitude: 0 }, {
    name: "SetDriverOnline rejects zero coordinates",
    expectError: true,
  });
  grpc("ReportLocation", { driverId: driverID, deviceId: deviceID, longitude: pos.longitude + 0.001, latitude: pos.latitude + 0.001 }, {
    assert: (resp) => {
      if (parseID(resp?.driverId) !== driverID || Number(resp?.onlineStatus) !== 1) throw new Error("report location response mismatch");
    },
  });
  grpc("ReportLocation", { driverId: driverID, deviceId: deviceID, longitude: 0, latitude: 0 }, {
    name: "ReportLocation rejects zero coordinates",
    expectError: true,
  });
  grpc("Heartbeat", { driverId: driverID, deviceId: deviceID, longitude: pos.longitude + 0.002, latitude: pos.latitude + 0.002 }, {
    assert: (resp) => {
      if (Number(resp?.onlineStatus) !== 1 || resp?.kicked) throw new Error("heartbeat response mismatch");
    },
  });
  grpc("Heartbeat", { driverId: driverID, deviceId: deviceID, longitude: 0, latitude: 0 }, {
    name: "Heartbeat rejects zero coordinates",
    expectError: true,
  });
  grpc("SetDriverServiceStatus", { driverId: driverID, onlineStatus: 2 }, {
    assert: (resp) => {
      if (Number(resp?.onlineStatus) !== 2) throw new Error("service status response mismatch");
    },
  });
  grpc("SetDriverServiceStatus", { driverId: driverID, onlineStatus: 1 }, { record: false });
  grpc("ListNearbyDrivers", { longitude: pos.longitude, latitude: pos.latitude, radiusMeters: 3000, limit: 5 }, {
    assert: (resp) => {
      if (!Array.isArray(resp?.drivers)) throw new Error("nearby drivers response is not a list");
    },
  });
  grpc("SetDriverOffline", { driverId: driverID, deviceId: deviceID, ...pos }, {
    assert: (resp) => {
      if (parseID(resp?.driverId) !== driverID || intField(resp?.onlineStatus, 0) !== 0) throw new Error("offline response mismatch");
    },
  });
  grpc("SetDriverOnline", { driverId: driverID, deviceId: deviceID, ...pos }, { record: false });

  const vehicleResp = grpc("CreateVehicle", mkVehicle(createDriverID, "Vehicle"), {
    assert: (resp) => {
      if (!parseID(resp?.id)) throw new Error("missing vehicle id");
    },
  });
  const vehicleID = parseID(vehicleResp.id);
  created.vehicles.push(vehicleID);
  grpc("GetVehicle", { id: vehicleID }, {
    assert: (resp) => {
      if (parseID(resp?.vehicle?.id) !== vehicleID) throw new Error("vehicle id mismatch");
    },
  });
  grpc("UpdateVehicle", { id: vehicleID, driverId: createDriverID, color: "Black", model: "RpcModelUpdated" }, {
    assert: (resp) => {
      if (parseID(resp?.id) !== vehicleID) throw new Error("updated vehicle id mismatch");
    },
  });

  const certDriverID = setupDriver("Cert");
  const certVehicleID = setupVehicle(certDriverID, "Cert");
  const certResp = grpc("UploadCertification", {
    driverId: certDriverID,
    vehicleId: certVehicleID,
    idCardFront: onePixelPng,
    idCardBack: onePixelPng,
    driverLicense: onePixelPng,
    vehicleLicense: onePixelPng,
  }, {
    assert: (resp) => {
      if (!parseID(resp?.id)) throw new Error("missing certification id");
    },
  });
  const certID = parseID(certResp.id);
  created.certifications.push(certID);
  grpc("GetCertification", { driverId: certDriverID }, {
    assert: (resp) => {
      if (!resp?.found || parseID(resp?.certification?.id) !== certID) throw new Error("certification not found");
    },
  });

  const approveDriverID = setupDriver("Approve");
  const approveVehicleID = setupVehicle(approveDriverID, "Approve");
  const approveCertID = setupCertification(approveDriverID, approveVehicleID);
  grpc("ApproveCertification", { certificationId: approveCertID, remark: "rpc check approve", operatorId: 1, ip: "127.0.0.1" }, {
    assert: (resp) => {
      if (resp?.message !== "ok") throw new Error("approve response mismatch");
    },
  });

  const rejectDriverID = setupDriver("Reject");
  const rejectVehicleID = setupVehicle(rejectDriverID, "Reject");
  const rejectCertID = setupCertification(rejectDriverID, rejectVehicleID);
  grpc("RejectCertification", { certificationId: rejectCertID, remark: "rpc check reject", operatorId: 1, ip: "127.0.0.1" }, {
    assert: (resp) => {
      if (resp?.message !== "ok") throw new Error("reject response mismatch");
    },
  });

  const withdrawResp = grpc("CreateWithdraw", { driverId: driverID, amount: 0.01, payeeName: "Rpc Check", payAccount: `acct-${Date.now()}` }, {
    assert: (resp) => {
      if (!parseID(resp?.id)) throw new Error("missing withdraw id");
    },
  });
  created.withdraws.push(parseID(withdrawResp.id));
  grpc("ListWithdraws", { driverId: driverID, page: 1, pageSize: 10 }, {
    assert: (resp) => {
      if (!Array.isArray(resp?.records)) throw new Error("withdraw records response is not a list");
    },
  });
  grpc("GetDriverAiScore", { driverId: driverID }, {
    assert: (resp) => {
      if (parseID(resp?.driverId) !== driverID) throw new Error("ai score driver id mismatch");
    },
  });

  grpc("FreezeDriver", { driverId: createDriverID, reason: "rpc check freeze", operatorId: 1, ip: "127.0.0.1" }, {
    assert: (resp) => {
      if (resp?.message !== "ok") throw new Error("freeze response mismatch");
    },
  });
  grpc("DeleteVehicle", { id: vehicleID, driverId: createDriverID }, {
    assert: (resp) => {
      if (parseID(resp?.id) !== vehicleID || !resp?.success) throw new Error("delete vehicle response mismatch");
    },
  });
  grpc("DeleteDriver", { id: registerDriverID }, {
    assert: (resp) => {
      if (parseID(resp?.id) !== registerDriverID || !resp?.success) throw new Error("delete driver response mismatch");
    },
  });
}

function cleanup() {
  try {
    grpc("SetDriverOnline", { driverId: driverID, deviceId: deviceID, ...pos }, { record: false });
  } catch (err) {
    console.log(`RESTORE_FAIL SetDriverOnline ${err.message}`);
  }

  const statements = [];
  if (created.withdraws.length) statements.push(`DELETE FROM driver_withdraw WHERE id IN (${created.withdraws.join(",")})`);
  if (created.certifications.length) statements.push(`DELETE FROM driver_certification WHERE id IN (${created.certifications.join(",")})`);
  if (created.vehicles.length) statements.push(`DELETE FROM driver_vehicle WHERE id IN (${created.vehicles.join(",")})`);
  if (created.drivers.length) {
    statements.push(`DELETE FROM driver_location WHERE driver_id IN (${created.drivers.join(",")})`);
    statements.push(`UPDATE driver SET deleted_at = NOW(), online_status = 0 WHERE id IN (${created.drivers.join(",")})`);
  }
  if (statements.length) {
    try {
      mysql(`${statements.join(";")};`);
    } catch (err) {
      console.log(`CLEANUP_FAIL ${err.message}`);
    }
  }

  for (const id of created.drivers) {
    const dir = path.resolve(".run", "certifications", "drivers", String(id));
    const root = path.resolve(".run", "certifications", "drivers");
    if (dir.startsWith(root + path.sep) && fs.existsSync(dir)) {
      fs.rmSync(dir, { recursive: true, force: true });
    }
  }
}

try {
  await main();
} catch (err) {
  if (err.result && !results.includes(err.result)) {
    results.push(err.result);
    print(err.result);
  } else {
    console.log(`FATAL ${err.message}`);
  }
} finally {
  cleanup();
  const failed = results.filter((r) => !r.ok);
  const methods = [...new Set(results.map((r) => r.method))];
  console.log(
    "SUMMARY " +
      JSON.stringify(
        {
          target,
          driverID,
          methods: methods.length,
          total: results.length,
          pass: results.length - failed.length,
          fail: failed.length,
          failed: failed.map((r) => ({
            name: r.name,
            method: r.method,
            exitCode: r.exitCode,
            message: r.message,
          })),
          created,
        },
        null,
        2,
      ),
  );
  if (failed.length) process.exitCode = 1;
}
