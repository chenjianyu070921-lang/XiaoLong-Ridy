# pushesvc 消息推送服务接口文档

> 服务职责：统一的消息触达能力，包含站内信、App 推送、短信三种通道。App 推送与短信通过 Provider 抽象层对接真实第三方网关（极光/个推/小米、阿里云/腾讯云等），未配置真实通道时走本地 noop 模式（开发/演示用）。所有外发结果均回写 `push_log` 表，站内信写入 `notices` 表。

## 1. 服务信息

| 项 | 值 |
| --- | --- |
| 服务名 | `pushesvc.rpc` |
| 协议 | gRPC（zrpc） |
| 监听地址 | `0.0.0.0:9002` |
| 注册中心 | etcd（`127.0.0.1:2379`，Key=`pushesvc.rpc`） |
| 存储 | MySQL（`xiaolongridy` 库的 `notices`、`push_log` 表）、Redis（配置依赖） |
| 外部通道 | 可配置 App 推送网关（jpush / xiaomi / getui）与短信网关（aliyun / tencent），未配置则 noop |

服务由 go-zero zrpc 暴露，调用方通过 proto 生成的 `pushesvcclient` 包或 grpc 直连接入。

## 2. 枚举与关键数据结构

### 2.1 BizType（业务类型，用于站内信与短信）

| 字段 | 值 | 说明 |
| --- | --- | --- |
| 站内信 `biz_type` | `1` / `2` / `3` | 1=订单通知，2=系统通知，3=活动通知 |
| 短信 `biz_type` | `1` / `2` / `3` | 1=验证码，2=通知，3=营销 |

> 说明：proto 中 `biz_type` 为 `int32`，非强枚举，调用方按上述整数传值。

### 2.2 外部通道配置（来自 `etc/pushesvc.yaml`）

| 配置项 | 说明 |
| --- | --- |
| `SMS.Provider` | 短信通道：`""` 或 `mock` → 本地 noop；`aliyun` / `tencent` → 真实网关 |
| `SMS.AccessKey` / `SMS.SecretKey` / `SMS.SignName` | 短信网关鉴权信息（noop 时不校验） |
| `Push.Provider` | 推送通道：`""` 或 `mock` → 本地 noop；`jpush` / `xiaomi` / `getui` → 真实网关 |
| `Push.AppKey` / `Push.MasterSecret` | 推送网关鉴权信息（noop 时不校验） |

网关地址映射：`aliyun→https://dysmsapi.aliyuncs.com`、`tencent→https://sms.tencentcloudapi.com`、`jpush→https://api.jpush.cn/v3/push`、`xiaomi→https://api.xmpush.xiaomi.com/v3/message/regid`、`getui→https://restapi.getui.com/v2/push`。

### 2.3 落库表结构（隐含）

| 表 | 关键字段 | 写入时机 |
| --- | --- | --- |
| `notices` | `user_id`、`title`、`content`、`biz_type`、`id` | `SendNotice` 成功写入，响应返回 `notice_id` |
| `push_log` | `user_id`、`push_type`(1=App推送/2=短信)、`title`、`content`、`target`、`result`(1成功/0失败)、`error_msg` | `SendPush` / `SendSMS` 每次调用均写入，记录真实通道结果 |

## 3. 接口列表

服务共 3 个 RPC：`SendNotice`、`SendPush`、`SendSMS`。

---

### 3.1 SendNotice — 发送站内信

向指定用户写入一条站内信，落库 `notices` 表并返回主键。

**请求 `SendNoticeReq`**

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `user_id` | int64 | 是 | 接收用户 ID（≤0 报错） |
| `title` | string | 是 | 标题（与 content 均不可为空） |
| `content` | string | 是 | 内容 |
| `biz_type` | int32 | 否 | 业务类型，见 2.1 节 |

**响应 `SendNoticeResp`**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `notice_id` | int64 | 新写入站内信的主键 ID |

**说明**：纯数据库写入，不调用外部通道。

---

### 3.2 SendPush — 发送 App 推送

向指定用户设备发送 App 推送，调用真实推送通道（由 `Push.Provider` 决定），失败自动重试一次，结果回写 `push_log`。

**请求 `SendPushReq`**

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `user_id` | int64 | 是 | 接收用户 ID（≤0 报错） |
| `title` | string | 否* | 标题（与 `body` 不能同时为空） |
| `body` | string | 否* | 内容（与 `title` 不能同时为空） |
| `extras` | string | 否 | 扩展字段（透传通道） |
| `device_type` | string | 否 | 设备类型：`ios` / `android` |

**响应 `SendPushResp`**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `success` | bool | 通道最终是否成功（含重试） |

**说明**：
- 通道调用至多 2 次（首次 + 失败重试 1 次）。
- 无论成功失败都会写 `push_log`（`result=1` 成功 / `0` 失败，`error_msg` 记录失败原因）。
- `noop` 模式下（未配置 `Push.Provider`）模拟成功，返回 `success=true`，`push_log.result=1`。
- 配置了真实网关但鉴权/参数错误时，通道返回非 2xx，记录 `result=0`，本次接口仍返回 `success=false`（不抛 gRPC 错误）。

---

### 3.3 SendSMS — 发送短信

向指定手机号发送短信，调用真实短信通道（由 `SMS.Provider` 决定），失败自动重试一次，结果回写 `push_log`。

**请求 `SendSMSReq`**

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `phone` | string | 是 | 手机号（空报错） |
| `content` | string | 是 | 短信内容（空报错） |
| `biz_type` | int32 | 否 | 业务类型，见 2.1 节 |

**响应 `SendSMSResp`**

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `success` | bool | 通道最终是否成功（含重试） |

**说明**：
- 通道调用至多 2 次（首次 + 失败重试 1 次）。
- 无论成功失败都会写 `push_log`（`push_type=2`，`user_id=0`，`target=手机号`，`result=1` 成功 / `0` 失败，`error_msg` 记录失败原因）。
- `noop` 模式下（未配置 `SMS.Provider`）模拟成功，返回 `success=true`，`push_log.result=1`。
- 配置了真实网关但鉴权/参数错误时，记录 `result=0`，接口返回 `success=false`（不抛 gRPC 错误）。

---

## 4. 错误约定

| 错误 | 触发场景 |
| --- | --- |
| `user_id 非法` | SendNotice / SendPush 的 `user_id ≤ 0` |
| `推送标题和内容不能都为空` | SendPush 的 `title` 与 `body` 同时为空 |
| `phone 不能为空` | SendSMS 的 `phone` 为空 |
| `短信内容不能为空` | SendSMS 的 `content` 为空 |
| 数据库写入错误 | `notices` / `push_log` 落库异常时透传底层错误 |
| 通道失败（不抛错） | SendPush / SendSMS 通道最终失败：不返回 gRPC 错误，仅 `success=false` 并写 `push_log.result=0` |

> 参数校验类错误通过 gRPC status 返回；通道发送失败被吞掉并以 `success=false` 体现，调用方需据此判断触达结果。

## 5. 调用方接入指引

1. 在调用方服务配置中声明 `pushesvc` 的 zrpc 客户端（target 指向 `127.0.0.1:9002` 或 etcd 服务发现名 `pushesvc.rpc`）。
2. 引入 `rpc/pushesvc/pushesvc` 包生成的 `pushesvcclient`，用 `NewPushService(zrpc.MustNewClient(c))` 获取 client。
3. 调用示例（Go）：

```go
client := pushesvcclient.NewPushService(zrpc.MustNewClient(c))

// 站内信
resp, err := client.SendNotice(ctx, &pushesvc.SendNoticeReq{
    UserId:  1001,
    Title:   "订单通知",
    Content: "您有一笔新订单",
    BizType: 1,
})

// App 推送
_, _ = client.SendPush(ctx, &pushesvc.SendPushReq{
    UserId:     1001,
    Title:      "优惠提醒",
    Body:       "您有一张优惠券即将过期",
    DeviceType: "android",
})

// 短信
_, _ = client.SendSMS(ctx, &pushesvc.SendSMSReq{
    Phone:   "13800000000",
    Content: "您的验证码是 123456",
    BizType: 1,
})
```

## 6. 注意事项

- **通道模式**：默认 `SMS.Provider` / `Push.Provider` 为空，走 `noop` 模式，本地可直接跑通且 `push_log` 记录为成功。要对接真实网关，需在 `etc/pushesvc.yaml` 填写对应 `Provider` 及鉴权信息。
- **重试与幂等**：`SendPush` / `SendSMS` 通道失败会重试 1 次；当前未实现去重/幂等键，重复调用会产生多条 `push_log` 记录与多次真实外发，调用方应避免短时间重复请求。
- **结果可观测性**：所有外发结果以 `push_log` 为准，建议上游通过 `push_log.result` / `error_msg` 做触达对账，而非仅依赖接口 `success` 字段。
- **当前接入状态**：`pushesvc` 自身已实现并可直接调用；上游（如 `ordersvc` / `usersvc`）通过 Kafka/gRPC 真正触发本服务的调用链路尚未串联，需在其他模块补充调用方。
