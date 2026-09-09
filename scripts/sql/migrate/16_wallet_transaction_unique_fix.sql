-- 修复钱包流水历史遗留的错误唯一索引。
-- 充值和提现使用 order_id=0，同一用户允许产生多条相同类型流水。
SET @idx_exists := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'user_wallet_transaction'
    AND index_name = 'uk_wallet_order_type'
);
SET @drop_sql := IF(@idx_exists > 0,
  'ALTER TABLE `user_wallet_transaction` DROP INDEX `uk_wallet_order_type`',
  'SELECT 1');
PREPARE stmt FROM @drop_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
