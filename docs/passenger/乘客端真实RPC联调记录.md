# 乘客端真实 RPC 联调记录

## 联调范围

本次联调覆盖乘客端 API 到 `usersvc`、`ordersvc`、`pricesvc` 的真实 gRPC 客户端注入。乘客端默认使用真实 gRPC 地址，不再在 RPC 地址缺省时静默回退 `LocalClient`。如需无下游依赖的本地测试，必须显式配置 `PASSENGER_CLIENT_MODE=local`。

## 生产环境密钥注入

生产环境不得在 `api/passenger/etc/passenger.yaml` 写入 JWT、七牛 AccessKey 或 SecretKey。请使用环境变量注入，JWT 密钥可用以下命令生成 32 字节随机值：

```bash
export PASSENGER_TOKEN_SIGNING_KEY="$(openssl rand -base64 32)"
export PASSENGER_QINIU_ACCESS_KEY="<七牛 AccessKey>"
export PASSENGER_QINIU_SECRET_KEY="<七牛 SecretKey>"
```

`PASSENGER_TOKEN_SIGNING_KEY` 优先于配置文件；未设置时兼容读取 `JWT_SIGNING_KEY`。密钥应由部署平台 Secret/Environment 管理，禁止提交到 Git。

## 环境变量

```powershell
$env:PASSENGER_HTTP_ADDR=':8091'
$env:PASSENGER_TOKEN_SIGNING_KEY='local-development-signing-key'
$env:PASSENGER_USERSVC_ADDR='127.0.0.1:50052'
$env:PASSENGER_ORDERSVC_ADDR='127.0.0.1:50051'
$env:PASSENGER_PRICESVC_ADDR='127.0.0.1:50053'
$env:PASSENGER_CLIENT_MODE='grpc'
$env:PASSENGER_PRICE_CITY_CODE='110000'
```

说明：

- `PASSENGER_CLIENT_MODE` 默认 `grpc`，生产和联调均使用真实 gRPC 客户端。
- `PASSENGER_CLIENT_MODE=local` 只用于单元测试或临时本地演示，不能作为验收默认配置。
- `PASSENGER_USERSVC_ADDR` 默认 `127.0.0.1:50052`。
- `PASSENGER_ORDERSVC_ADDR` 默认 `127.0.0.1:50051`。
- `PASSENGER_PRICESVC_ADDR` 默认 `127.0.0.1:50053`。
- `PASSENGER_PRICE_CITY_CODE` 默认 `110000`，真实 `pricesvc` 计价规则查询依赖该城市编码。

## 启动顺序

```powershell
$env:GOROOT='D:\Go\go1.26.5'
$env:GOCACHE='D:\gowork\XiaoLong-Ridy\.gocache'
& 'D:\Go\go1.26.5\bin\go.exe' run .\rpc\usersvc -f .\rpc\usersvc\etc\usersvc.yaml
& 'D:\Go\go1.26.5\bin\go.exe' run .\rpc\ordersvc -f .\rpc\ordersvc\etc\ordersvc.yaml
& 'D:\Go\go1.26.5\bin\go.exe' run .\rpc\pricesvc -f .\rpc\pricesvc\etc\pricesvc.yaml
& 'D:\Go\go1.26.5\bin\go.exe' run .\api\passenger
```

注意：`ordersvc` 和 `pricesvc` 启动依赖各自 `etc/*.yaml` 中的 MySQL 可访问。若数据库不可用，乘客端能启动但创建订单会在下游 RPC 返回错误。

## Apipost 核心接口联调记录

### 1. 发送短信验证码

- 方法：`POST`
- 地址：`http://127.0.0.1:8091/api/passenger/v1/auth/send-sms-code`
- 请求体：

```json
{
  "phone": "13800138000"
}
```

- 期望：响应 `code=0`，`data.success=true`。
- 验证码来源：当前 `usersvc` 仍使用内存短信服务，验证码会打印在 `usersvc` 控制台日志。

### 2. 短信登录

- 方法：`POST`
- 地址：`http://127.0.0.1:8091/api/passenger/v1/auth/login-by-sms`
- 请求体：

```json
{
  "phone": "13800138000",
  "code": "{{smsCode}}"
}
```

- 期望：响应 `code=0`，返回 `data.token`、`data.refreshToken` 和 `data.user.userId`。
- Apipost 变量：将 `data.token` 保存为环境变量 `passengerToken`。

### 3. 创建订单

- 方法：`POST`
- 地址：`http://127.0.0.1:8091/api/passenger/v1/orders/create`
- Header：`Authorization: Bearer {{passengerToken}}`
- 请求体：

```json
{
  "carType": 1,
  "fromAddress": "北京市朝阳区建国路88号",
  "fromLongitude": 116.47319,
  "fromLatitude": 39.9096,
  "toAddress": "北京市海淀区中关村大街27号",
  "toLongitude": 116.31683,
  "toLatitude": 39.98472
}
```

- 期望：响应 `code=0`，返回 `data.orderId`、`data.orderNo`、`data.estimatedPriceCents`、`data.status`。
- 链路：`passenger api -> pricesvc.EstimatePrice -> ordersvc.CreateOrder`，全程通过真实 gRPC client。

## 自动化验证

```powershell
$env:GOROOT='D:\Go\go1.26.5'
$env:GOCACHE=(Join-Path $env:TEMP 'codex-go-cache-XiaoLong-Ridy')
& 'D:\Go\go1.26.5\bin\go.exe' test ./api/passenger/...
& 'D:\Go\go1.26.5\bin\go.exe' test ./...
```

当前自动化验证覆盖：

- `usersvc` 错误变量导出单元测试。
- passenger 环境变量配置读取。
- passenger 默认使用真实 gRPC adapter。
- passenger 显式 `PASSENGER_CLIENT_MODE=local` 时才使用 LocalClient。
- passenger 到 `pricesvc` 的坐标请求到里程/时长/城市编码转换。
- passenger 下单携带优惠券时调用 `pricesvc.CalculateDiscount` 计算抵扣和应付金额。
- passenger HTTP 发短信、短信登录、创建订单核心流程。
- passenger 个人中心通过 `usersvc.GetProfile` 查询用户资料。
- passenger 实名提交通过 `usersvc.SubmitRealName` 更新用户实名字段。
- 下单、地址、订单列表参数校验。
- passenger 统一错误码映射。

## P2 边界说明

优惠券使用已在下单链路接入 `pricesvc.CalculateDiscount`，本期采用“请求携带优惠券参数，价格服务计算抵扣”的轻量模式，与 `pricesvc.proto` 中“优惠券信息本期作为入参传入，不落库”的设计保持一致。

个人中心和实名已补齐 `usersvc` 的 proto、repository、logic、server、LocalClient 链路，并由 passenger API 接入：

- `POST /api/passenger/v1/profile/me`：携带 `Authorization: Bearer {token}`，请求体 `{}`，返回当前乘客基础资料。
- `POST /api/passenger/v1/profile/real-name`：携带 `realName` 和 `idCardNo`，调用 `usersvc.SubmitRealName` 保存实名字段并返回更新后的认证状态。

用户优惠券列表本期未扩展为独立查询接口；当前只覆盖“下单时使用优惠券”的核心链路。
