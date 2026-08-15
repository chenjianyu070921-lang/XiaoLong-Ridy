# dispatchsvc 最小派单联调记录

日期：2026-08-15
负责人：成员4（订单与派单调度）

## 1. 涉及服务

- dispatchsvc：新增派单 RPC，gRPC 端口 8083
- ordersvc：创建订单后自动调用 dispatchsvc.DispatchOrder
- MySQL：xiaolong_ridy

## 2. 数据库准备

执行迁移脚本：

```sql
-- 先执行基础迁移
scripts/sql/migrate/01_user_module.sql
scripts/sql/migrate/02_driver_module.sql
scripts/sql/migrate/04_order_dispatch_module.sql
scripts/sql/migrate/06_location_module.sql
-- 再执行派单 mock 种子数据
scripts/sql/migrate/08_dispatch_mock_seed.sql
```

## 3. 启动顺序

```bash
cd rpc/dispatchsvc && go run . -f etc/dispatchsvc.yaml
cd rpc/ordersvc && go run . -f etc/ordersvc.yaml
```

## 4. 联调步骤

1. 通过 Apipost 或 grpcurl 调用 `ordersvc.Order.CreateOrder`，入参包含起终点经纬度和 car_type。
2. 创建订单成功后，ordersvc 会调用 `dispatchsvc.Dispatch.DispatchOrder`。
3. 查询数据库确认派单记录：

```sql
SELECT id, order_id, driver_id, dispatch_type, status, match_score, remark
FROM dispatch_record
WHERE order_id = ?;
```

预期结果：返回 3 条记录，driver_id 为 9001/9002/9003，dispatch_type=1（自动派单），status=1（派单中）。

4. 调用 `dispatchsvc.Dispatch.ListDispatchRecords`，按 order_id 分页查询，确认 total 与记录一致。

## 5. 验证命令

```bash
go build ./rpc/dispatchsvc/... ./rpc/ordersvc/...
go test ./rpc/dispatchsvc/... ./rpc/ordersvc/...
```

## 6. 遗留问题

- P0 阶段为 mock 直派，未接入真实司机位置匹配。
- 司机接单后 dispatch_record.status 仍为派单中，P2 接入 order-event-consumer 后更新。
