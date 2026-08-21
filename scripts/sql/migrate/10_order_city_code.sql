-- 订单表补充城市编码字段：用于派单 GEO 按真实城市检索、计价按真实城市规则重算。
-- 执行环境：线上库 xiaolong_ridy（115.191.16.159:3306）。
ALTER TABLE `ride_order`
  ADD COLUMN `city_code` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '城市编码，空表示默认城市' AFTER `car_type`;
