-- 修复历史版本钱包流水唯一索引，允许同一用户多次充值和提现。
SET @wallet_index_exists := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'user_wallet_transaction'
    AND index_name = 'uk_wallet_order_type'
);
SET @wallet_drop_sql := IF(
  @wallet_index_exists > 0,
  'ALTER TABLE `user_wallet_transaction` DROP INDEX `uk_wallet_order_type`',
  'SELECT 1'
);
PREPARE wallet_drop_stmt FROM @wallet_drop_sql;
EXECUTE wallet_drop_stmt;
DEALLOCATE PREPARE wallet_drop_stmt;
