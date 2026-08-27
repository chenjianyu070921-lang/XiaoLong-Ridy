-- 优化 user_coupon 查询性能：创建订单时根据用户+状态查可用券并排序。
-- 原查询 WHERE user_id=? AND status=? ORDER BY received_at DESC, id DESC
-- 仅 idx_user_status(user_id, status) 无法覆盖 ORDER BY received_at，导致 filesort。
-- 新增联合索引 (user_id, status, received_at, id)，使过滤与排序全部走索引。

ALTER TABLE `user_coupon`
  ADD INDEX `idx_user_status_received` (`user_id`, `status`, `received_at`, `id`);
