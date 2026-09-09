-- 订单表补充城市编码字段：用于派单 GEO 按真实城市检索、计价按真实城市规则重算。
-- 本脚本需要兼容 MySQL 5.7；该版本不支持 ADD COLUMN IF NOT EXISTS，
-- 因此通过 information_schema 判断字段是否存在，保证脚本重复执行不会报错。
SET @ride_order_city_code_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'ride_order'
    AND COLUMN_NAME = 'city_code'
);
SET @ride_order_city_code_sql := IF(
  @ride_order_city_code_exists = 0,
  CONCAT(
    'ALTER TABLE `ride_order` ADD COLUMN `city_code` VARCHAR(16) NOT NULL DEFAULT ',
    CHAR(39), CHAR(39),
    ' COMMENT ', CHAR(39), '城市编码，空表示默认城市', CHAR(39),
    ' AFTER `car_type`'
  ),
  'SELECT 1'
);
PREPARE ride_order_city_code_stmt FROM @ride_order_city_code_sql;
EXECUTE ride_order_city_code_stmt;
DEALLOCATE PREPARE ride_order_city_code_stmt;
