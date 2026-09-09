# 管理后台列表高级筛选设计

## 目标与范围

仅修改 `web/admin`，为既有通用工作区的列表页补齐后端已开放的字段化高级筛选，并为风控命中记录提供可查询的筛选控件。

不修改 `api/admin`、`rpc/adminsvc`、数据库结构、迁移脚本、业务数据和任何接口出入参。

## 方案

在 `console/index.vue` 内定义按业务 `kind` 索引的筛选配置表。每项配置描述请求字段名、控件类型、标签、占位文案和固定选项；模板按配置循环渲染，不再为每个列表继续增加平行的条件分支。

筛选状态仍复用当前 `filters` 响应式对象。提交查询时重置为第一页并将非空字段、分页参数透传至既有 API 封装；重置时清空当前页面配置中的所有字段，不影响其他列表页面的默认参数。

时间范围使用一个 `datetimerange` 控件，并在请求前展开为接口已有的 `start_time`、`end_time`，值格式保持 `YYYY-MM-DD HH:mm:ss`。ID 字段使用数字输入框，防止将非数值传给 `int64` 查询参数。

## 字段映射

| 页面 | 新增筛选字段 |
| --- | --- |
| 用户管理 | `start_time`、`end_time` |
| 司机审核 | `start_time`、`end_time` |
| 订单管理 | `user_id`、`driver_id`、`start_time`、`end_time` |
| 优惠券模板 | `type`、`status`、`start_time`、`end_time` |
| 发券任务 | `coupon_id`、`status`、`start_time`、`end_time` |
| 计价规则 | `city_code`、`car_type`、`status` |
| 工单 | `status`、`assignee_id`、`work_order_type` |
| 操作日志 | `admin_id`、`module`、`action`、`target_type`、`target_id`、`start_time`、`end_time` |
| 风控命中记录 | `target_type`、`target_id`、`scene`、`risk_level` |

所有枚举值仅使用现有接口文档或当前服务端实现明确支持的值。操作日志的 `module`、`action` 保持自由文本输入，避免前端枚举遗漏后端审计动作；城市编码同样保持文本输入。

## 错误处理与验收

筛选请求继续由统一 Axios 客户端处理鉴权和错误提示；前端不将空字符串作为有效 ID 传递。进入每个列表、执行查询、重置、分页和每页条数变更均复用已有列表加载逻辑。

验收标准：

1. 所有字段化控件与上表一致，点击查询后的 URL 请求参数符合后端接口定义。
2. 日期范围展开为 `start_time`、`end_time`，重置后不保留旧筛选值。
3. 风控命中记录可按场景、风险等级、目标对象类型与目标 ID 查询。
4. `npm run build` 和 `git diff --check -- web/admin` 成功。
