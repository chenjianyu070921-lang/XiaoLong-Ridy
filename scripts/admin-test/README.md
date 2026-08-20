# 管理后台接口自动化测试脚本

适用范围：`api/admin`（模块三：管理后台）。测试逻辑、请求体和期望结果均对齐 `docs/api/管理后台接口文档.md`、`docs/admin/管理后台P0说明文档.md` 与当前路由实现。

## 环境要求

1. 本地 Redis 已启动：`127.0.0.1:6379`。
2. 远程 MySQL `xiaolong_ridy` 可达（配置见 `api/admin/etc/admin.json`）。
3. Go 工具链可用。
4. 如 `rpc/adminsvc` 使用默认 `etc/admin.yaml`（带 Etcd 注册），需要本地 etcd；本脚本会自动生成一份不含 Etcd 的 `.gotmp/adminsvc-admin-test.yaml` 用于联调。

> 已知限制（2026-08-18）：
> - `rpc/ordersvc` 当前因配置结构 `redis` 键与 go-zero v1.7.2 `RpcServerConf.Redis` 冲突无法启动（模块四问题），脚本用占位 gRPC（15051）顶替，`POST /orders/{id}/cancel` 无法端到端验证。
> - 远程库未应用 `07_admin_operation_business.sql`，`admin_coupon_issue_task`、`risk_blacklist_hit_record` 缺失，发券任务/发券/风控命中记录接口返回 500。
> - 详见 `docs/admin/管理后台全量接口测试报告-2026-08-18.md`。

## 使用步骤

```powershell
# 1. 启动服务栈（ordersvc:50051 -> driversvc:8080 -> adminsvc:8084 -> api/admin:8083）
.\scripts\admin-test\start_admin_stack.ps1

# 2. 运行接口测试（只读 + 错误路径用例）
.\scripts\admin-test\admin_api_test.ps1

# 3. 运行包含正向写操作用例的完整测试（会创建并回收 AUTOTEST_* 测试数据）
.\scripts\admin-test\admin_api_test.ps1 -WriteOps

# 4. 测试完成后停止服务栈
.\scripts\admin-test\start_admin_stack.ps1 -Stop
```

## 自定义参数

```powershell
.\scripts\admin-test\admin_api_test.ps1 `
  -BaseUrl "http://127.0.0.1:8083" `
  -Username "admin" `
  -Password "123456" `
  -WriteOps `
  -ReportPath "C:\tmp\admin-report.json"
```

## 测试覆盖

- 鉴权：注册（非法请求体）、登录（空/错误/正确）、退出、当前账号、菜单、未带 token 的 401、方法不允许 405。
- 操作日志：列表 + 筛选。
- 用户管理：列表、筛选、详情 404、冻结/解冻 404、非法路径 ID。
- 司机审核：列表、筛选、详情 404、通过/驳回 404。
- 订单管理：列表、筛选、详情 404、异常订单列表、后台取消 404。
- 优惠券：列表、新增（非法体）、编辑 404、下架 404、发券 404/非法配置、发券任务列表；`-WriteOps` 下覆盖 新增→发券→编辑→下架 正向链路。
- 营销活动：列表、新增（非法体）、编辑 404、发布/回滚 404；`-WriteOps` 下覆盖 新增→编辑→发布→回滚。
- 数据统计：总览、订单统计、优惠券统计。
- 导出任务：列表、新增（非法体）；`-WriteOps` 下覆盖正向创建。
- 风控：黑名单列表、新增（非法体）、解除 404、命中记录；`-WriteOps` 下覆盖 新增→解除。

## 输出

- 控制台逐条 PASS/FAIL。
- JSON 报告默认写到 `.gotmp/admin-test-logs/admin-api-report-<时间戳>.json`，可通过 `-ReportPath` 覆盖。

> 说明：正向写操作会向测试库写入 `AUTOTEST_*` 优惠券/活动/黑名单记录，并在用例内完成下架/解除回收；发券用例会向 `user_coupon` 写入一条 `user_id=999999` 的测试记录，用于验证发券链路。
