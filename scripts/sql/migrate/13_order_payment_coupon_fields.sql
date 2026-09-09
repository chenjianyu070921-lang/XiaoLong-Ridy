-- 订单表补充优惠券与支付退款金额字段：用于下单锁券、支付确认、订单退款和后台强制退款。
-- 本脚本兼容 MySQL 5.7；通过 information_schema 判断字段是否存在，确保重复执行不会报错。

SET @ride_order_coupon_id_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'ride_order'
    AND COLUMN_NAME = 'coupon_id'
);
SET @ride_order_coupon_id_sql := IF(
  @ride_order_coupon_id_exists = 0,
  'ALTER TABLE `ride_order` ADD COLUMN `coupon_id` BIGINT NOT NULL DEFAULT 0 COMMENT ''锁定的优惠券ID，0表示未使用优惠券'' AFTER `estimated_price`',
  'SELECT 1'
);
PREPARE ride_order_coupon_id_stmt FROM @ride_order_coupon_id_sql;
EXECUTE ride_order_coupon_id_stmt;
DEALLOCATE PREPARE ride_order_coupon_id_stmt;

SET @ride_order_discount_cents_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'ride_order'
    AND COLUMN_NAME = 'discount_cents'
);
SET @ride_order_discount_cents_sql := IF(
  @ride_order_discount_cents_exists = 0,
  'ALTER TABLE `ride_order` ADD COLUMN `discount_cents` BIGINT NOT NULL DEFAULT 0 COMMENT ''优惠金额，单位分'' AFTER `coupon_id`',
  'SELECT 1'
);
PREPARE ride_order_discount_cents_stmt FROM @ride_order_discount_cents_sql;
EXECUTE ride_order_discount_cents_stmt;
DEALLOCATE PREPARE ride_order_discount_cents_stmt;

SET @ride_order_payable_cents_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'ride_order'
    AND COLUMN_NAME = 'payable_cents'
);
SET @ride_order_payable_cents_sql := IF(
  @ride_order_payable_cents_exists = 0,
  'ALTER TABLE `ride_order` ADD COLUMN `payable_cents` BIGINT NOT NULL DEFAULT 0 COMMENT ''订单应付金额，单位分'' AFTER `discount_cents`',
  'SELECT 1'
);
PREPARE ride_order_payable_cents_stmt FROM @ride_order_payable_cents_sql;
EXECUTE ride_order_payable_cents_stmt;
DEALLOCATE PREPARE ride_order_payable_cents_stmt;

SET @ride_order_paid_cents_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'ride_order'
    AND COLUMN_NAME = 'paid_cents'
);
SET @ride_order_paid_cents_sql := IF(
  @ride_order_paid_cents_exists = 0,
  'ALTER TABLE `ride_order` ADD COLUMN `paid_cents` BIGINT NOT NULL DEFAULT 0 COMMENT ''已支付金额，单位分'' AFTER `payable_cents`',
  'SELECT 1'
);
PREPARE ride_order_paid_cents_stmt FROM @ride_order_paid_cents_sql;
EXECUTE ride_order_paid_cents_stmt;
DEALLOCATE PREPARE ride_order_paid_cents_stmt;

SET @ride_order_refund_cents_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'ride_order'
    AND COLUMN_NAME = 'refund_cents'
);
SET @ride_order_refund_cents_sql := IF(
  @ride_order_refund_cents_exists = 0,
  'ALTER TABLE `ride_order` ADD COLUMN `refund_cents` BIGINT NOT NULL DEFAULT 0 COMMENT ''已退款金额，单位分'' AFTER `paid_cents`',
  'SELECT 1'
);
PREPARE ride_order_refund_cents_stmt FROM @ride_order_refund_cents_sql;
EXECUTE ride_order_refund_cents_stmt;
DEALLOCATE PREPARE ride_order_refund_cents_stmt;

SET @ride_order_actual_price_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'ride_order'
    AND COLUMN_NAME = 'actual_price'
);
SET @ride_order_actual_price_sql := IF(
  @ride_order_actual_price_exists = 0,
  'ALTER TABLE `ride_order` ADD COLUMN `actual_price` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT ''实际价格'' AFTER `refund_cents`',
  'SELECT 1'
);
PREPARE ride_order_actual_price_stmt FROM @ride_order_actual_price_sql;
EXECUTE ride_order_actual_price_stmt;
DEALLOCATE PREPARE ride_order_actual_price_stmt;
