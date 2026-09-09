# AI 运营助手流式输出（SSE）设计

日期：2026-09-07
状态：已确认，实现中

## 1. 背景与目标

- 现状：`AskAiAgent` 为同步 unary 调用，方舟生成结构化 JSON 实测约 **43 秒**；前端 `AiAgentDrawer.vue` 在此期间消息区完全空白（仅发送按钮转圈），演示体验差。
- 目标：**3~5 秒出首字**，结论逐字渲染；数据证据先于结论展示；模型不可用/超时仍保留降级语义。

## 2. 架构与数据流

```
前端 EventSource
  → GET /admin/v1/ai-agent/ask/stream（SSE）
网关 admin-api（net/http，手写 SSE，逐帧 Flush）
  → gRPC: AdminSvc.AskAiAgentStream(AiAskRequest) returns (stream AiAnswerChunk)
adminsvc engine
  ├─ ① facts    ：查库（约 1s）立即推「数据证据」→ 前端秒出卡片
  ├─ ② delta    ：方舟 stream=true，边收边推原始文本增量 → 前端逐字显示「结论」
  ├─ ③ final    ：生成完毕后做防幻觉校验，推结构化结果（priorities/actions/source_mode）
  └─ ④ fallback ：超时/校验失败时推模板答案（沿用既有降级语义）
```

保留原 unary `AskAiAgent` 不动，兼容现有调用与测试。

## 3. 协议

```proto
rpc AskAiAgentStream(AiAskRequest) returns (stream AiAnswerChunk);

message AiAnswerChunk {
  string type = 1;              // facts | delta | final | fallback | error
  string text = 2;              // delta 时的增量文本
  AiAnswerResponse answer = 3;  // facts/final/fallback 的结构化内容
  string conversation_id = 4;
  string trace_id = 5;
}
```

SSE 帧格式：`data: {"type":"delta","text":"..."}\n\n`；结束帧 `data: {"type":"final",...}`。

## 4. 改动清单

| # | 位置 | 改动 |
|---|---|---|
| 1 | `rpc/adminsvc/admin.proto` | 新增流式 RPC 与 `AiAnswerChunk` |
| 2 | pb 重新生成 | 临时目录生成，**只拷 `.pb.go`**，不覆盖 `internal/` |
| 3 | `rpc/adminsvc/internal/aiagent/ark.go` | 新增 `GenerateStream`（`stream=true`，解析 SSE `data:` 帧） |
| 4 | `rpc/adminsvc/internal/aiagent/agent.go` | 新增 `AskStream`：facts → delta → 校验 final / fallback |
| 5 | `rpc/adminsvc/internal/logic/adminservice/ai_agent_logic.go` | 实现 `AskAiAgentStream`，转发 emit |
| 6 | `api/admin/internal/handler/router_ai.go` | 新增 SSE 路由（走 `authRequired`，`text/event-stream` + Flush） |
| 7 | `web/admin/src/components/AiAgentDrawer.vue`、`api/modules` | EventSource 订阅；facts 渲染证据、delta 逐字、final 补齐；关闭时 abort |

## 5. 风险与回退

- **风险**：`goctl rpc protoc` 带 `--zrpc_out` 会生成整套服务骨架，可能覆盖 `internal/`；且 `Makefile` 引用的 `scripts/admin-test/regenerate_adminsvc_proto.ps1` 实际不存在。
  - 规避：在临时目录生成，仅拷贝 `adminsvc/*.pb.go` 到现有包目录；操作前备份。
- **回退方案**：若 goctl 生成与现有代码不兼容（编译失败），降级为**方案 B**——新增轻量 `facts` 接口，前端先渲染证据卡片再异步等结论，不改 proto 流式。
- 网关 HTTP `WriteTimeout=60s` 需覆盖流式总时长；若模型更慢需同步放宽。

## 6. 验收

- 现有 `aiagent` 与网关单测不回归（`go test ./rpc/adminsvc/... ./api/admin/...`）。
- `curl -N` 能按 `facts → delta… → final` 顺序收到事件。
- 前端：提问后约 1 秒出现数据证据卡片，随后结论逐字增长，结束后补齐优先对象/建议动作。
- 模型超时/校验失败时收到 `fallback`，`source_mode=template_fallback`，行为与改动前一致。
