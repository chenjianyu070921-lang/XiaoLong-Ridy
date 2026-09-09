# 司机智能助手对话接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/agent/chat` |
| 是否登录 | 是，或携带内部服务 Token |
| 当前状态 | 已实现 |
| 业务逻辑 | `AgentChatHandler` |

## 2. 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `question` | string | 是 | 司机输入的问题 |

## 3. 请求示例

```json
{"question":"我现在有哪些待接派单？"}
```

## 4. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `answer` | string | 助手回答 |
| `loopCount` | int | 工具调用轮数 |
| `observations` | array | 工具观察结果 |
| `mode` | string | 运行模式 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-AGENT-E01 | 未登录且无内部服务 Token | HTTP 401 |
| DRIVER-AGENT-E02 | 请求方法错误 | HTTP 405 |
| DRIVER-AGENT-E03 | 请求体不是 JSON | HTTP 400 |

## 6. 处理链路

`api/driver -> RequireDriverOrInternalService -> AgentChatHandler`。该接口允许司机 JWT 或受信内部服务 Token 访问。
