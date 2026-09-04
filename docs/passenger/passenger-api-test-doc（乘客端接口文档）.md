# 乘客端 API 接口测试文档

> **版本**: v1.1  
> **更新时间**: 2026-08-31  
> **服务地址**: `http://localhost:8091`  
> **Content-Type**: `application/json; charset=utf-8`

> **路径说明**：当前网关统一使用 `/api/passenger/v1` 前缀。早期示例中的 `/api/v1` 旧路径已废弃，请以 `api/passenger/passenger.api` 和当前路由为准。

---

## 目录

1. [通用说明](#1-通用说明)
2. [统一响应格式](#2-统一响应格式)
3. [错误码清单](#3-错误码清单)
4. [认证模块](#4-认证模块)
5. [个人中心模块](#5-个人中心模块)
6. [地址管理模块](#6-地址管理模块)
7. [订单模块](#7-订单模块)
8. [优惠券模块](#8-优惠券模块)
9. [价格预估模块](#9-价格预估模块)

---

## 1. 通用说明

### 1.1 认证方式

| 接口类型 | Authorization | 说明 |
|---------|--------------|------|
| 公开接口 | 不需要 | 发送验证码、登录、刷新令牌 |
| 需登录 | `Bearer {token}` | 个人中心、地址、订单、优惠券 |
| 需实名 | `Bearer {token}` + realNameStatus=verified | 提交实名后 |

### 1.2 测试前置条件

```bash
# 获取 Token（先调用登录接口）
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 通用请求头
HEADER_AUTH="Authorization: Bearer $TOKEN"
HEADER_JSON="Content-Type: application/json"
```

---

## 2. 统一响应格式

### 成功响应
```json
{
  "code": 0,
  "message": "success",
  "data": {"..."},
  "timestamp": 1724150400,
  "traceId": "trace_a1b2c3d4"
}
```

### 错误响应
```json
{
  "code": 40000,
  "message": "请求参数不合法",
  "data": null,
  "timestamp": 1724150400,
  "traceId": "trace_e5f6g7h8"
}
```

---

## 3. 错误码清单

| 错误码 | HTTP状态码 | 含义 | 触发场景 |
|--------|-----------|------|---------|
| 0 | 200 | 成功 | 正常返回 |
| 40000 | 400 | 请求参数不合法 | 必填字段缺失/格式错误/业务校验失败 |
| 40001 | 400 | 手机号格式错误 | phone 不是11位数字 |
| 40101 | 401 | Token已过期 | JWT过期 |
| 40102 | 401 | Token无效 | 篡改/伪造/空Token |
| 40301 | 403 | 账号已冻结 | 用户被禁用 |
| 40303 | 403 | 无权访问该资源 | 操作他人资源 |
| 40401 | 404 | 地址不存在 | 删除/查询不存在的地址ID |
| 40402 | 404 | 用户不存在 | 查询已注销用户 |
| 40500 | 405 | 方法不允许 | POST用GET访问 |
| 41001 | 400 | 验证码错误 | code输入错误 |
| 41002 | 400 | 验证码已过期 | 超过5分钟未使用 |
| 41003 | 429 | 验证码发送过于频繁 | 60秒内重复发送 |
| 41011 | 400 | 电话格式错误(地址) | 地址联系电话格式非法 |
| 41014 | 400 | 经纬度错误 | 经度[-180,180]或纬度[-90,90]越界 |
| 42001 | 400 | 实名认证失败 | 姓名与身份证号不一致/身份证号无效/腾讯云服务异常 |
| 50000 | 500 | 服务器内部错误 | 未预期的异常 |
| 50001 | 502 | 下游服务不可用 | gRPC连接失败/超时 |
| 40002 | 400 | 密码格式不合法 | 密码不满足 8-64 位且同时包含字母和数字 |
| 40103 | 401 | 手机号或密码错误 | 账号不存在、未设置密码或密码校验失败 |

---

## 4. 认证模块

### 4.1 发送短信验证码

**`POST /api/v1/auth/sms-code`**

#### 正常用例

| 用例编号 | 场景 | 请求体 | 预期响应 |
|---------|------|--------|---------|
| AUTH-SMS-001 | 正常发送验证码 | `{"phone":"13800138000"}` | `code:0, data:{success:true, expireIn:300}` |

**请求示例**
```bash
curl -X POST http://localhost:8091/api/v1/auth/sms-code \
  -H "$HEADER_JSON" \
  -d '{"phone":"13800138000"}'
```

**预期响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "success": true,
    "expireIn": 300
  },
  "timestamp": 1724150400,
  "traceId": "trace_xxx"
}
```

#### 异常用例（参数拦截）

| 用例编号 | 场景 | 请求体 | 预期code | 预期HTTP |
|---------|------|--------|----------|----------|
| AUTH-SMS-E01 | phone为空 | `{}` | 40000 | 400 |
| AUTH-SMS-E02 | phone为null | `{"phone":null}` | 40000 | 400 |
| AUTH-SMS-E03 | phone非数字 | `{"phone":"abc"}` | 40001 | 400 |
| AUTH-SMS-E04 | phone不足11位 | `{"phone":"138"}` | 40001 | 400 |
| AUTH-SMS-E05 | phone超过11位 | `{"phone":"138001380000"}` | 40001 | 400 |
| AUTH-SMS-E06 | phone含特殊字符 | `{"phone":"138-0013-8000"}` | 40001 | 400 |
| AUTH-SMS-E07 | JSON格式错误 | `not json` | 40000 | 400 |
| AUTH-SMS-E08 | 额外字段 | `{"phone":"13800138000","extra":"x"}` | 0 | 200 (忽略额外字段) |

#### 多分支用例（业务逻辑）

| 用例编号 | 场景 | 前置条件 | 预期结果 |
|---------|------|---------|---------|
| AUTH-SMS-B01 | 同一手机号60秒内重复发送 | 已在60秒内发过 | `code:41003, HTTP:429` |
| AUTH-SMS-B02 | 不同手机号连续发送 | 无 | 均成功，各自独立冷却 |
| AUTH-SMS-B03 | 虚拟号码段(170/171) | 使用170开头号码 | 根据业务规则：允许/拒绝 |

#### 边界值用例

| 用例编号 | 场景 | 请求体 | 预期结果 |
|---------|------|--------|---------|
| AUTH-SMS-BD01 | phone最小长度1位 | `{"phone":"1"}` | 40001 |
| AUTH-SMS-BD02 | phone最大长度20位 | `{"phone":"13800138000123456789"}` | 40001 |
| AUTH-SMS-BD03 | phone刚好11位正确 | `{"phone":"13800138000"}` | 0 (成功) |
| AUTH-SMS-BD04 | phone全为0 | `{"phone":"00000000000"}` | 40001 或 0 (视业务) |

---

### 4.2 短信验证码登录

**`POST /api/v1/auth/login`**

#### 正常用例

| 用例编号 | 场景 | 请求体 | 预期响应 |
|---------|------|--------|---------|
| AUTH-LOGIN-001 | 新用户首次登录 | `{"phone":"13800138000","code":"123456"}` | `isNewUser:true, token非空` |
| AUTH-LOGIN-002 | 老用户正常登录 | `{"phone":"13800138001","code":"654321"}` | `isNewUser:false, token非空` |

**请求示例**
```bash
curl -X POST http://localhost:8091/api/v1/auth/login \
  -H "$HEADER_JSON" \
  -d '{"phone":"13800138000","code":"123456"}'
```

**预期响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGci...",
    "refreshToken": "eyJhbGci...",
    "isNewUser": true,
    "user": {
      "userId": 1,
      "phone": "138****8000",
      "nickname": "乘客1",
      "avatarUrl": "",
      "realNameStatus": "unverified"
    }
  },
  "timestamp": 1724150400,
  "traceId": "trace_xxx"
}
```

#### 异常用例（参数拦截）

| 用例编号 | 场景 | 请求体 | 预期code | 预期HTTP |
|---------|------|--------|----------|----------|
| AUTH-LG-E01 | phone为空 | `{"code":"123456"}` | 40000/40001 | 400 |
| AUTH-LG-E02 | code为空 | `{"phone":"13800138000"}` | 40000 | 400 |
| AUTH-LG-E03 | phone格式错误 | `{"phone":"abc","code":"123"}` | 40001 | 400 |
| AUTH-LG-E04 | code长度不为6 | `{"phone":"13800138000","code":"123"}` | 40000 | 400 |
| AUTH-LG-E05 | code含非数字 | `{"phone":"13800138000","code":"abcdef"}` | 40000 | 400 |
| AUTH-LG-E06 | 全部字段为空 | `{}` | 40000 | 400 |

#### 多分支用例（业务逻辑）

| 用例编号 | 场景 | 前置条件 | 预期结果 |
|---------|------|---------|---------|
| AUTH-LG-B01 | 验证码错误 | 已发送正确码123456，输入111111 | `code:41001` |
| AUTH-LG-B02 | 验证码已过期 | 发送后等待5分钟+再提交 | `code:41002` |
| AUTH-LG-B03 | 验证码已被消耗 | 同一验证码第二次使用 | `code:41001` 或 41002 |
| AUTH-LG-B04 | 冻结账号登录 | 用户状态为frozen | `code:40301, HTTP:403` |
| AUTH-LG-B05 | 注销账号登录 | 用户状态为deleted | `code:40402, HTTP:404` |

#### 安全性用例

| 用例编号 | 场景 | 请求方式 | 预期结果 |
|---------|------|---------|---------|
| AUTH-LG-S01 | GET方法访问 | GET /api/v1/auth/login | `code:40500, HTTP:405` |
| AUTH-LG-S02 | SQL注入尝试 | `{"phone":"138'; DROP TABLE--","code":"123456"}` | 40001 (被拦截) |
| AUTH-LG-S03 | XSS尝试 | `{"phone":"<script>alert(1)</script>","code":"123456"}` | 40001 (被拦截) |

---

### 4.3 手机号密码登录

**`POST /api/passenger/v1/auth/login-by-password`**（无需登录）

密码需为 8-64 位且同时包含字母和数字；账号需先通过短信登录后在个人中心设置密码。

```bash
curl -X POST http://localhost:8091/api/passenger/v1/auth/login-by-password \
  -H "$HEADER_JSON" \
  -d '{"phone":"13800138000","password":"Pass1234"}'
```

响应结构与短信登录一致，返回 `token`、`refreshToken` 和 `user`。账号不存在、未设置密码或密码错误时返回 HTTP 401，客户端可切换验证码登录。

### 4.4 刷新令牌

**`POST /api/v1/auth/refresh-token`**

#### 正常用例

| 用例编号 | 场景 | 请求体 | 预期响应 |
|---------|------|--------|---------|
| AUTH-REF-001 | 正常刷新 | `{"refreshToken":"eyJ..."}` | 返回新token+新refreshToken |

#### 异常用例

| 用例编号 | 场景 | 请求体 | 预期code | 预期HTTP |
|---------|------|--------|----------|----------|
| AUTH-REF-E01 | refreshToken为空 | `{}` | 40000 | 400 |
| AUTH-REF-E02 | refreshToken无效 | `{"refreshToken":"invalid"}` | 40102 | 401 |
| AUTH-REF-E03 | refreshToken已过期 | 过期的JWT | 40101 | 401 |
| AUTH-REF-E04 | 使用accessToken刷新 | 传入token而非refreshToken | 40102 | 401 |

---

### 4.4 登出

**`POST /api/v1/auth/logout`**

#### 正常用例

| 用例编号 | 场景 | 请求头 | 预期响应 |
|---------|------|--------|---------|
| AUTH-LOGOUT-001 | 正常登出 | `Authorization: Bearer {valid_token}` | `code:0, data:{success:true}` |

#### 异常用例

| 用例编号 | 场景 | 请求头 | 预期code | 预期HTTP |
|---------|------|--------|----------|----------|
| AUTH-LOGOUT-E01 | 无Token | 不带Authorization | 40102 | 401 |
| AUTH-LOGOUT-E02 | Token无效 | `Bearer invalid_token` | 40102 | 401 |
| AUTH-LOGOUT-E03 | Token已注销 | 重复登出 | 40102 | 401 (已在黑名单) |

---

## 5. 个人中心模块

### 5.1 获取个人信息

**`GET /api/v1/profile`** (需登录)

#### 正常用例

| 用例编号 | 场景 | 预期响应 |
|---------|------|---------|
| PROFILE-GET-001 | 已登录用户查看资料 | 返回userId/phone/nickname/avatarUrl/realNameStatus |

**请求示例**
```bash
curl -X GET http://localhost:8091/api/v1/profile \
  -H "$HEADER_AUTH"
```

**预期响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user": {
      "userId": 1,
      "phone": "138****8000",
      "nickname": "乘客1",
      "avatarUrl": "https://cdn.example.com/avatar/1.png",
      "realNameStatus": "verified"
    }
  },
  "timestamp": 1724150400,
  "traceId": "trace_xxx"
}
```

#### 异常用例（鉴权拦截）

| 用例编号 | 场景 | 请求头 | 预期code | 预期HTTP |
|---------|------|--------|----------|----------|
| PROFILE-GET-E01 | 无Authorization头 | 不带 | 40102 | 401 |
| PROFILE-GET-E02 | 空Bearer | `Authorization: Bearer` | 40102 | 401 |
| PROFILE-GET-E03 | Bearer格式错误 | `Authorization: Basic xxx` | 40102 | 401 |
| PROFILE-GET-E04 | 篡改Token | 修改JWT payload | 40102 | 401 |
| PROFILE-GET-E05 | 过期Token | exp已过期的JWT | 40101 | 401 |
| PROFILE-GET-E06 | 司机Token访问 | accountType=driver的JWT | 40102 | 401 |

---

### 5.2 提交实名认证

**`POST /api/v1/profile/realname`** (需登录)

> **说明**: 本接口调用**腾讯云市场身份证实名认证 API** 进行实时二要素核验（姓名+身份证号），验证通过后立即更新用户实名状态。

**后端处理流程**:
```
1. 参数校验（非空、格式）
2. 调用腾讯云市场 API: POST https://ap-beijing.cloudmarket-apigw.com/service-18c38npd/idcard/VerifyIdcardv2
   - 请求体: {"cardNo": "身份证号", "realName": "姓名"}
   - 认证方式: HMAC-SHA1 签名
3. 根据返回结果:
   - isMatch=true → 更新数据库，realNameStatus=verified
   - isMatch=false → 返回错误
4. 返回更新后的用户信息
```

#### 正常用例

| 用例编号 | 场景 | 请求体 | 预期响应 |
|---------|------|--------|---------|
| PROFILE-RN-001 | 正确姓名+身份证号 | `{"realName":"张三","idCardNo":"110101199001011234"}` | realNameStatus变为verified |

**请求示例**
```bash
curl -X POST http://localhost:8091/api/v1/profile/realname \
  -H "$HEADER_AUTH" \
  -H "$HEADER_JSON" \
  -d '{"realName":"张三","idCardNo":"110101199001011234"}'
```

**预期响应（成功）**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user": {
      "userId": 1,
      "phone": "138****8000",
      "nickname": "乘客1",
      "avatarUrl": "",
      "realNameStatus": "verified",
      "realName": "张三",
      "idCardNo": "110***********1234"
    }
  },
  "timestamp": 1724150400,
  "traceId": "trace_xxx"
}
```

**预期响应（认证失败）**
```json
{
  "code": 42001,
  "message": "实名认证失败：姓名与身份证号不一致",
  "data": null,
  "timestamp": 1724150400,
  "traceId": "trace_xxx"
}
```

#### 异常用例（参数拦截）

| 用例编号 | 场景 | 请求体 | 预期code | 预期HTTP |
|---------|------|--------|----------|----------|
| PROFILE-RN-E01 | realName为空 | `{"idCardNo":"110101199001011234"}` | 40000 | 400 |
| PROFILE-RN-E02 | idCardNo为空 | `{"realName":"张三"}` | 40000 | 400 |
| PROFILE-RN-E03 | realName含数字 | `{"realName":"张三123","idCardNo":"110101199001011234"}` | 40000 | 400 |
| PROFILE-RN-E04 | idCardNo格式错误(17位) | `{"realName":"张三","idCardNo":"11010119900101123"}` | 40000 | 400 |
| PROFILE-RN-E05 | idCardNo格式错误(19位) | `{"realName":"张三","idCardNo":"1101011990010112345"}` | 40000 | 400 |
| PROFILE-RN-E06 | idCardNo校验位错误 | `{"realName":"张三","idCardNo":"110101199001011235"}` | 40000 | 400 |
| PROFILE-RN-E07 | idCardNo含字母X(合法) | `{"realName":"张三","idCardNo":"11010119900101123X"}` | 42001 或 0 (视API) |

#### 多分支用例（业务逻辑）

| 用例编号 | 场景 | 前置条件 | 预期结果 |
|---------|------|---------|---------|
| PROFILE-RN-B01 | 姓名与身份证号不匹配 | 真实身份证号+错误姓名 | `code:42001, message:"实名认证失败"` |
| PROFILE-RN-B02 | 身份证号不存在 | 编造的假身份证号 | `code:42001, message:"实名认证失败"` |
| PROFILE-RN-B03 | 已认证用户再次提交 | 已是verified状态 | 成功（覆盖更新或忽略） |
| PROFILE-RN-B04 | 未登录提交 | 无Token | 40102 |
| PROFILE-RN-B05 | 腾讯云服务不可用 | API超时/网络异常 | 50001 或 50000 |
| PROFILE-RN-B06 | 未配置腾讯云密钥 | 开发环境未配置 | 成功（跳过核验，直接保存） |

#### 安全性用例

| 用例编号 | 场景 | 请求体 | 预期结果 |
|---------|------|--------|---------|
| PROFILE-RN-S01 | SQL注入尝试 | `{"realName":"张'; DROP TABLE--","idCardNo":"110101199001011234"}` | 40000 (被参数校验拦截) |
| PROFILE-RN-S02 | XSS尝试 | `{"realName":"<script>alert(1)</script>","idCardNo":"110101199001011234"}` | 40000 (被参数校验拦截) |
| PROFILE-RN-S03 | 特殊字符姓名 | `{"realName":"张-三·测试","idCardNo":"110101199001011234"}` | 视业务规则允许/拒绝 |

---

## 6. 地址管理模块

### 6.1 新增常用地址

**`POST /api/v1/addresses`** (需登录)

#### 正常用例

| 用例编号 | 场景 | 请求体 | 预期响应 |
|---------|------|--------|---------|
| ADDR-CREATE-001 | 新增完整地址 | 见下方 | 返回地址ID > 0 |

**请求示例**
```bash
curl -X POST http://localhost:8091/api/v1/addresses \
  -H "$HEADER_AUTH" \
  -H "$HEADER_JSON" \
  -d '{
    "contactName": "张三",
    "contactPhone": "13900139000",
    "tag": "家",
    "address": "北京市朝阳区xxx街道xxx号",
    "longitude": 116.397128,
    "latitude": 39.916527,
    "isDefault": true,
    "sort": 1
  }'
```

**预期响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "contactName": "张三",
    "contactPhone": "13900139000",
    "tag": "家",
    "address": "北京市朝阳区xxx街道xxx号",
    "longitude": 116.397128,
    "latitude": 39.916527,
    "isDefault": true,
    "sort": 1
  },
  "timestamp": 1724150400,
  "traceId": "trace_xxx"
}
```

#### 异常用例（参数拦截）

| 用例编号 | 场景 | 请求体 | 预期code | 预期HTTP |
|---------|------|--------|----------|----------|
| ADDR-CR-E01 | contactName为空 | 缺少contactName | 40000 | 400 |
| ADDR-CR-E02 | contactPhone为空 | 缺少contactPhone | 40000 | 400 |
| ADDR-CR-E03 | contactPhone格式错误 | `"contactPhone":"abc"` | 41011 | 400 |
| ADDR-CR-E04 | address为空 | 缺少address | 40000 | 400 |
| ADDR-CR-E05 | longitude越界(>180) | `"longitude":200` | 41014 | 400 |
| ADDR-CR-E06 | longitude越界(<-180) | `"longitude":-200` | 41014 | 400 |
| ADDR-CR-E07 | latitude越界(>90) | `"latitude":100` | 41014 | 400 |
| ADDR-CR-E08 | latitude越界(<-90) | `"latitude":-100` | 41014 | 400 |
| ADDR-CR-E09 | 全部必填字段为空 | `{}` | 40000 | 400 |

#### 边界值用例

| 用例编号 | 场景 | 请求体 | 预期结果 |
|---------|------|--------|---------|
| ADDR-CR-BD01 | longitude刚好180 | `"longitude":180` | 0 (成功，国际日期变更线) |
| ADDR-CR-BD02 | longitude刚好-180 | `"longitude":-180` | 0 (成功) |
| ADDR-CR-BD03 | latitude刚好90 | `"latitude":90` | 0 (成功，北极点) |
| ADDR-CR-BD04 | latitude刚好-90 | `"latitude":-90` | 0 (成功，南极点) |
| ADDR-CR-BD05 | 经纬度均为0 | `"longitude":0,"latitude":0` | 0 (成功，本初子午线/赤道) |
| ADDR-CR-BD06 | contactName超长(100字) | 100个汉字 | 0 或 40000 (视限制) |
| ADDR-CR-BD07 | address超长(500字) | 500个字符 | 0 或 40000 (视限制) |

---

### 6.2 查询地址列表

**`GET /api/v1/addresses`** (需登录)

#### 正常用例

| 用例编号 | 场景 | 预期响应 |
|---------|------|---------|
| ADDR-LIST-001 | 有地址的用户 | 返回地址数组，按sort排序 |
| ADDR-LIST-002 | 无地址的用户 | 返回空数组 `[]` |

#### 异常用例

| 用例编号 | 场景 | 预期code | 预期HTTP |
|---------|------|----------|----------|
| ADDR-LIST-E01 | 未登录 | 40102 | 401 |

---

### 6.3 更新地址

**`PUT /api/v1/addresses`** (需登录)

#### 正常用例

| 用例编号 | 场景 | 请求体 | 预期响应 |
|---------|------|--------|---------|
| ADDR-UPDATE-001 | 更新已有地址 | `{"id":1,"contactName":"李四",...}` | 更新成功 |

#### 异常用例

| 用例编号 | 场景 | 请求体 | 预期code | 预期HTTP |
|---------|------|--------|----------|----------|
| ADDR-UP-E01 | id为0 | `"id":0,...` | 40000 | 400 |
| ADDR-UP-E02 | id不存在 | `"id":99999,...` | 40401 | 404 |
| ADDR-UP-E03 | id为负数 | `"id":-1,...` | 40000 | 400 |
| ADDR-UP-E04 | 修改他人地址 | 其他用户的地址ID | 40303 | 403 |
| ADDR-UP-E05 | 参数校验同新增 | 同6.1异常用例 | 对应code | 400 |

---

### 6.4 删除地址

**`DELETE /api/v1/addresses`** (需登录)

#### 正常用例

| 用例编号 | 场景 | 请求体 | 预期响应 |
|---------|------|--------|---------|
| ADDR-DEL-001 | 删除已有地址 | `{"id":1}` | `success:true` |

#### 异常用例

| 用例编号 | 场景 | 请求体 | 预期code | 预期HTTP |
|---------|------|--------|----------|----------|
| ADDR-DEL-E01 | id为0 | `{"id":0}` | 40000 | 400 |
| ADDR-DEL-E02 | id不存在 | `{"id":99999}` | 40401 | 404 |
| ADDR-DEL-E03 | 删除他人地址 | 其他用户的地址ID | 40303 | 403 |
| ADDR-DEL-E04 | 重复删除 | 已删除的ID | 40401 | 404 |

---

## 7. 订单模块

### 7.1 创建订单

**`POST /api/v1/orders`** (需登录)

#### 正常用例

| 用例编号 | 场景 | 请求体 | 预期响应 |
|---------|------|--------|---------|
| ORDER-CREATE-001 | 创建快车订单 | 见下方 | 返回orderId/orderNo/status=WAITING |

**请求示例**
```bash
curl -X POST http://localhost:8091/api/v1/orders \
  -H "$HEADER_AUTH" \
  -H "$HEADER_JSON" \
  -d '{
    "carType": 1,
    "fromAddress": "北京市朝阳区望京SOHO",
    "fromLongitude": 116.479,
    "fromLatitude": 39.998,
    "toAddress": "北京首都国际机场T3",
    "toLongitude": 116.612,
    "toLatitude": 40.079
  }'
```

**预期响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "orderId": 5001,
    "orderNo": "RO202608200001",
    "estimatedPriceCents": 3500,
    "originalPriceCents": 3500,
    "discountAmountCents": 0,
    "payableAmountCents": 3500,
    "userCouponId": 0,
    "status": 1,
    "createdAt": 1724150400
  },
  "timestamp": 1724150400,
  "traceId": "trace_xxx"
}
```

#### 异常用例（参数拦截）

| 用例编号 | 场景 | 请求体 | 预期code | 预期HTTP |
|---------|------|--------|----------|----------|
| ORDER-CR-E01 | carType为0/负数 | `"carType":0` | 40000 | 400 |
| ORDER-CR-E02 | fromAddress为空 | 缺少fromAddress | 40000 | 400 |
| ORDER-CR-E03 | fromLongitude/fromLatitude缺失 | 缺少经纬度 | 40000/41014 | 400 |
| ORDER-CR-E04 | toAddress为空 | 缺少toAddress | 40000 | 400 |
| ORDER-CR-E05 | toLongitude/toLatitude缺失 | 缺少目的经纬度 | 40000/41014 | 400 |
| ORDER-CR-E06 | 起终点相同 | from==to完全一致 | 40000 | 400 |
| ORDER-CR-E07 | 起终点距离过近(<100m) | 起终点相距50米 | 40000 | 400 |
| ORDER-CR-E08 | carType不支持 | `"carType":99` | 40000 | 400 |
| ORDER-CR-E09 | 全部字段为空 | `{}` | 40000 | 400 |

#### 多分支用例（业务逻辑）

| 用例编号 | 场景 | 前置条件 | 预期结果 |
|---------|------|---------|---------|
| ORDER-CR-B01 | 使用优惠券下单 | userCouponId有效 | discountAmountCents>0 |
| ORDER-CR-B02 | 优惠券不可用 | 已过期/不满足门槛 | 40000 或忽略券 |
| ORDER-CR-B03 | 用户有进行中订单 | 存在status<4的订单 | 40000 (请先完成当前行程) |
| ORDER-CR-B04 | 下游ordersvc不可用 | gRPC断开 | 50001 | 502 |
| ORDER-CR-B05 | 未登录创建 | 无Token | 40102 | 401 |

#### 优惠券相关分支

| 用例编号 | 场景 | 请求体附加字段 | 预期结果 |
|---------|------|---------------|---------|
| ORDER-CR-CP01 | 有效满减券 | couponId/couponType/couponThresholdCents等 | 抵扣成功 |
| ORDER-CR-CP02 | 金额不足门槛 | 订单金额<couponThresholdCents | 40000 或不抵扣 |
| ORDER-CR-CP03 | 券不属于当前用户 | 他人优惠券ID | 40303 或 40000 |
| ORDER-CR-CP04 | 券已使用 | status=used | 40000 |

---

### 7.2 订单列表查询

**`GET /api/v1/orders`** (需登录)

#### 正常用例

| 用例编号 | 场景 | 查询参数 | 预期响应 |
|---------|------|---------|---------|
| ORDER-LIST-001 | 查询全部订单 | 无参数或status=0 | 返回分页订单列表 |
| ORDER-LIST-002 | 按状态筛选 | `?status=1` | 只返回待接单订单 |
| ORDER-LIST-003 | 分页查询 | `?page=1&pageSize=10` | 返回第1页10条 |

#### 异常用例

| 用例编号 | 场景 | 查询参数 | 预期code | 预期HTTP |
|---------|------|---------|----------|----------|
| ORDER-LST-E01 | status为负数 | `?status=-1` | 40000 | 400 |
| ORDER-LST-E02 | status超出范围 | `?status=99` | 40000 | 400 |
| ORDER-LST-E03 | page为0 | `?page=0` | 40000 | 400 |
| ORDER-LST-E04 | page为负数 | `?page=-1` | 40000 | 400 |
| ORDER-LST-E05 | pageSize为0 | `?pageSize=0` | 40000 | 400 |
| ORDER-LST-E06 | pageSize超限 | `?pageSize=1000` | 40000 或自动截断 |
| ORDER-LST-E07 | 未登录 | 无Token | 40102 | 401 |

---

### 7.3 订单详情查询

**`GET /api/v1/orders/{orderId}`** (需登录)

#### 正常用例

| 用例编号 | 场景 | 预期响应 |
|---------|------|---------|
| ORDER-GET-001 | 查询自己的订单 | 返回完整订单详情 |
| ORDER-GET-002 | 查询不存在的orderId | 40401 | 404 |

#### 异常用例

| 用例编号 | 场景 | 预期code | 预期HTTP |
|---------|------|----------|----------|
| ORDER-GET-E01 | orderId为0 | 40000 | 400 |
| ORDER-GET-E02 | orderId为负数 | 40000 | 400 |
| ORDER-GET-E03 | orderId非数字 | 40000 | 400 (路由不匹配则404) |
| ORDER-GET-E04 | 查询他人订单 | 40303 | 403 |
| ORDER-GET-E05 | 未登录 | 40102 | 401 |

---

### 7.4 取消订单

**`POST /api/v1/orders/cancel`** (需登录)

#### 正常用例

| 用例编号 | 场景 | 请求体 | 预期响应 |
|---------|------|--------|---------|
| ORDER-CANCEL-001 | 待接单时取消 | `{"orderId":5001,"reason":"不想坐了"}` | status变为CANCELLED |

#### 异常用例（参数拦截）

| 用例编号 | 场景 | 请求体 | 预期code | 预期HTTP |
|---------|------|--------|----------|----------|
| ORDER-CNCL-E01 | orderId为0 | `{"orderId":0}` | 40000 | 400 |
| ORDER-CNCL-E02 | reason为空 | `{"orderId":5001}` | 40000 或 0 (视是否必填) |
| ORDER-CNCL-E03 | 订单不存在 | `{"orderId":99999}` | 40401/40402 | 404 |

#### 多分支用例（状态机拦截）

| 用例编号 | 场景 | 当前订单状态 | 预期结果 |
|---------|------|-------------|---------|
| ORDER-CNCL-B01 | 待接单(WAITING)取消 | status=1 | 成功取消 |
| ORDER-CNCL-B02 | 已接单(ACCEPTED)取消 | status=2 | 成功(可能扣费) |
| ORDER-CNCL-B03 | 行程中(IN_PROGRESS)取消 | status=3 | 40000 (行程中不可取消) |
| ORDER-CNCL-B04 | 已完成(COMPLETED)取消 | status=4 | 40000 (已完成不可取消) |
| ORDER-CNCL-B05 | 已取消(CANCELLED)重复取消 | status=5 | 40000 (已取消) |
| ORDER-CNCL-B06 | 支付中(PAYING)取消 | status=6 | 40000 或允许 |
| ORDER-CNCL-B07 | 已支付待评价(PAID)取消 | status=7 | 40000 (不可取消) |
| ORDER-CNCL-B08 | 已退款(REFUNDED)取消 | status=8 | 40000 (已结束) |

---

### 7.5 发起支付

**`POST /api/v1/orders/pay`** (需登录)

#### 正常用例

| 用例编号 | 场景 | 请求体 | 预期响应 |
|---------|------|--------|---------|
| ORDER-PAY-001 | 微信支付 | `{"orderId":5001,"channel":1}` | 返回paymentId/payParams |
| ORDER-PAY-002 | 支付宝支付 | `{"orderId":5001,"channel":2}` | 返回paymentId/payParams |

#### 异常用例（参数拦截）

| 用例编号 | 场景 | 请求体 | 预期code | 预期HTTP |
|---------|------|--------|----------|----------|
| ORDER-PAY-E01 | orderId为0 | `{"orderId":0,"channel":1}` | 40000 | 400 |
| ORDER-PAY-E02 | channel为0 | `{"orderId":5001,"channel":0}` | 40000 | 400 |
| ORDER-PAY-E03 | channel不支持 | `{"orderId":5001,"channel":99}` | 40000 | 400 |
| ORDER-PAY-E04 | 订单不存在 | `{"orderId":99999,"channel":1}` | 40401 | 404 |

#### 多分支用例（状态机拦截）

| 用例编号 | 场景 | 当前订单状态 | 预期结果 |
|---------|------|-------------|---------|
| ORDER-PAY-B01 | 待支付(PAID_PENDING)支付 | status=7 | 成功发起支付 |
| ORDER-PAY-B02 | 待接单(WAITING)支付 | status=1 | 40000 (订单不可支付) |
| ORDER-PAY-B03 | 行程中(IN_PROGRESS)支付 | status=3 | 40000 (订单不可支付) |
| ORDER-PAY-B04 | 已完成(COMPLETED)支付 | status=4 | 40000 (订单不可支付) |
| ORDER-PAY-B05 | 已取消(CANCELLED)支付 | status=5 | 40000 (订单不可支付) |
| ORDER-PAY-B06 | 已支付(PAID)重复支付 | status=8 | 40000 (已支付) |
| ORDER-PAY-B07 | paysvc不可用 | gRPC断开 | 50001 | 502 |

---

## 8. 优惠券模块

### 8.1 领取优惠券

**`POST /api/v1/coupons/claim`** (需登录)

#### 正常用例

| 用例编号 | 场景 | 请求体 | 预期响应 |
|---------|------|--------|---------|
| COUPON-CLAIM-001 | 领取可用优惠券 | `{"couponId":1}` | 返回userCoupon信息，status=UNUSED |

#### 异常用例（参数拦截）

| 用例编号 | 场景 | 请求体 | 预期code | 预期HTTP |
|---------|------|--------|----------|----------|
| COUPON-CLM-E01 | couponId为0 | `{"couponId":0}` | 40000 | 400 |
| COUPON-CLM-E02 | couponId为负数 | `{"couponId":-1}` | 40000 | 400 |
| COUPON-CLM-E03 | couponId缺失 | `{}` | 40000 | 400 |

#### 多分支用例（业务逻辑）

| 用例编号 | 场景 | 前置条件 | 预期结果 |
|---------|------|---------|---------|
| COUPON-CLM-B01 | 优惠券不存在 | couponId=99999 | 40401 (优惠券不存在) |
| COUPON-CLM-B02 | 优惠券已下架 | status=inactive | 40000 (优惠券不可领取) |
| COUPON-CLM-B03 | 已领完(库存0) | 领取次数达上限 | 40000 (领取次数已达上限) |
| COUPON-CLM-B04 | 重复领取 | 已领过同款券 | 40000 (不可重复领取) |
| COUPON-CLM-B05 | 未登录领取 | 无Token | 40102 | 401 |
| COUPON-CLM-B06 | usersvc不可用 | gRPC断开 | 50001 | 502 |

---

### 8.2 我的优惠券列表

**`GET /api/v1/coupons`** (需登录)

#### 正常用例

| 用例编号 | 场景 | 查询参数 | 预期响应 |
|---------|------|---------|---------|
| COUPON-LIST-001 | 查询全部 | 无参数或status=0 | 返回全部优惠券 |
| COUPON-LIST-002 | 查询未使用 | `?status=1` | 只返回UNUSED状态的券 |
| COUPON-LIST-003 | 查询已使用 | `?status=2` | 只返回USED状态的券 |
| COUPON-LIST-004 | 查询已过期 | `?status=3` | 只返回EXPIRED状态的券 |

#### 异常用例

| 用例编号 | 场景 | 查询参数 | 预期code | 预期HTTP |
|---------|------|---------|----------|----------|
| COUPON-LST-E01 | status为负数 | `?status=-1` | 40000 | 400 |
| COUPON-LST-E02 | status超出范围 | `?status=5` | 40000 | 400 |
| COUPON-LST-E03 | 未登录 | 无Token | 40102 | 401 |

#### 状态值对照表

| status值 | 含义 |
|---------|------|
| 0 | 全部 |
| 1 | 未使用(UNUSED) |
| 2 | 已使用(USED) |
| 3 | 已过期(EXPIRED) |
| 4 | 已锁定(LOCKED) |

---

## 9. 价格预估模块

### 9.1 预估价格

**`POST /api/v1/prices/estimate`** (需登录)

#### 正常用例

| 用例编号 | 场景 | 请求体 | 预期响应 |
|---------|------|--------|---------|
| PRICE-EST-001 | 快车预估 | 见下方 | 返回预估价/原价/明细 |

**请求示例**
```bash
curl -X POST http://localhost:8091/api/v1/prices/estimate \
  -H "$HEADER_AUTH" \
  -H "$HEADER_JSON" \
  -d '{
    "carType": 1,
    "fromLongitude": 116.479,
    "fromLatitude": 39.998,
    "toLongitude": 116.612,
    "toLatitude": 40.079
  }'
```

**预期响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "estimatedPriceCents": 3500,
    "originalPriceCents": 3500,
    "discountAmountCents": 0,
    "estimatedDistanceM": 15000,
    "estimatedDurationS": 1200,
    "priceBreakdown": [
      {"name": "起步价", "amountCents": 1300},
      {"name": "里程费(12km)", "amountCents": 1800},
      {"name": "时长费(20min)", "amountCents": 400}
    ]
  },
  "timestamp": 1724150400,
  "traceId": "trace_xxx"
}
```

#### 异常用例（参数拦截）

| 用例编号 | 场景 | 请求体 | 预期code | 预期HTTP |
|---------|------|--------|----------|----------|
| PRICE-EST-E01 | carType为0 | `"carType":0,...` | 40000 | 400 |
| PRICE-EST-E02 | 起点经纬度缺失 | 缺少fromLongitude | 40000/41014 | 400 |
| PRICE-EST-E03 | 终点经纬度缺失 | 缺少toLongitude | 40000/41014 | 400 |
| PRICE-EST-E04 | 起点经度越界 | `"fromLongitude":200` | 41014 | 400 |
| PRICE-EST-E05 | 终点纬度越界 | `"toLatitude":100` | 41014 | 400 |
| PRICE-EST-E06 | 起终点相同 | from==to | 40000 | 400 |

#### 多分支用例

| 用例编号 | 场景 | 前置条件 | 预期结果 |
|---------|------|---------|---------|
| PRICE-EST-B01 | 夜间时段(23:00-06:00) | 当前时间为凌晨 | 包含夜间附加费 |
| PRICE-EST-B02 | 高峰时段动态调价 | 配置了动态倍率 | priceMultiplier>1.0 |
| PRICE-EST-B03 | pricesvc不可用 | gRPC断开 | 50001 | 502 |
| PRICE-EST-B04 | 未登录预估 | 无Token | 40102 | 401 |

---

## 附录 A: 接口汇总速查表

| 模块 | 方法 | 路径 | 认证 | 用例数(正/异/多分支) |
|-----|------|------|------|---------------------|
| 认证 | POST | /api/v1/auth/sms-code | 否 | 1/8/3/4 |
| 认证 | POST | /api/v1/auth/login | 否 | 2/6/5/3 |
| 认证 | POST | /api/v1/auth/refresh-token | 否 | 1/4/0 |
| 认证 | POST | /api/v1/auth/logout | 是 | 1/3/0 |
| 个人中心 | GET | /api/v1/profile | 是 | 1/6/0 |
| 个人中心 | POST | /api/v1/profile/realname | 是 | 1/7/6/3 |
| 地址 | POST | /api/v1/addresses | 是 | 1/9/7 |
| 地址 | GET | /api/v1/addresses | 是 | 2/1/0 |
| 地址 | PUT | /api/v1/addresses | 是 | 1/5/0 |
| 地址 | DELETE | /api/v1/addresses | 是 | 1/4/0 |
| 订单 | POST | /api/v1/orders | 是 | 1/9/5/4 |
| 订单 | GET | /api/v1/orders | 是 | 3/7/0 |
| 订单 | GET | /api/v1/orders/:id | 是 | 1/5/0 |
| 订单 | POST | /api/v1/orders/cancel | 是 | 1/3/8 |
| 订单 | POST | /api/v1/orders/pay | 是 | 2/4/7 |
| 优惠券 | POST | /api/v1/coupons/claim | 是 | 1/3/6 |
| 优惠券 | GET | /api/v1/coupons | 是 | 4/3/0 |
| 价格 | POST | /api/v1/prices/estimate | 是 | 1/6/4 |

**总计**: 18 个接口 | ~27 个正常用例 | ~95 个异常用例 | ~56 个多分支用例 | ~16 个边界值用例

---

## 附录 B: 测试执行检查清单

### P0 - 必须通过（阻塞上线）

- [ ] 所有公开接口无需Token可访问
- [ ] 所有需登录接口无Token返回401
- [ ] 无效/过期Token返回对应错误码
- [ ] 参数缺失/格式错误返回40000
- [ ] 手机号格式校验（11位数字）
- [ ] 经纬度范围校验（经度[-180,180]，纬度[-90,90]）
- [ ] 订单状态机流转校验（不可逆操作被拦截）

### P1 - 应当通过（影响体验）

- [ ] 验证码冷却时间60秒生效
- [ ] 验证码5分钟过期
- [ ] 优惠券领取次数限制
- [ ] 他人资源访问返回403
- [ ] 下游服务不可用时返回502而非500

### P2 - 建议通过（优化项）

- [ ] SQL注入/XSS攻击被参数校验拦截
- [ ] 额外JSON字段被忽略不报错
- [ ] 边界值（180/-180/90/-90）处理正确
- [ ] 超长字符串有合理截断或拒绝
