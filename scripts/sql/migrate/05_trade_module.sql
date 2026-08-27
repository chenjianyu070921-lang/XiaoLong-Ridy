-- =============================================================
-- 模块五：计价与支付（rpc/pricesvc + rpc/paysvc）
-- 表清单：price_rule、order_price、payment、settlement
-- =============================================================

-- 计价规则表：保存不同车型/时段的计价参数。
-- 必须有这张表：起步价、里程费、时长费、夜间费都依赖规则配置，后台可调整且不能写死在代码里。
CREATE TABLE IF NOT EXISTS `price_rule` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '规则ID',
  `name` VARCHAR(50) NOT NULL COMMENT '规则名称',
  `city_code` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '城市编码，空表示全局',
  `car_type` TINYINT NOT NULL DEFAULT 1 COMMENT '车型：1特惠快车 2快车 3拼车',
  `base_price` DECIMAL(10,2) NOT NULL COMMENT '起步价',
  `base_distance_km` DECIMAL(6,2) NOT NULL DEFAULT 0.00 COMMENT '起步包含里程（公里）',
  `per_km_price` DECIMAL(6,2) NOT NULL COMMENT '每公里价格',
  `per_minute_price` DECIMAL(6,2) NOT NULL COMMENT '每分钟时长费',
  `night_start_time` TIME DEFAULT NULL COMMENT '夜间费开始时间',
  `night_end_time` TIME DEFAULT NULL COMMENT '夜间费结束时间',
  `night_surcharge` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '夜间附加费（元/单）',
  `dynamic_max_factor` DECIMAL(3,2) NOT NULL DEFAULT 1.00 COMMENT '动态调价最大倍数',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 2停用',
  `effective_at` DATETIME NOT NULL COMMENT '生效时间',
  `expire_at` DATETIME DEFAULT NULL COMMENT '失效时间，NULL表示长期有效',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_car_status` (`car_type`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='计价规则表';

-- 订单价格表：保存订单的预估价格和最终实付价格明细。
-- 必须有这张表：一口价、优惠抵扣、实付金额都需要按订单固化明细，避免后续规则调整影响历史订单。
CREATE TABLE IF NOT EXISTS `order_price` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '价格记录ID',
  `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  `price_rule_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '使用的计价规则ID',
  `estimated_price` DECIMAL(10,2) NOT NULL COMMENT '预估价格',
  `actual_price` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '实际总价',
  `base_fee` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '起步价费用',
  `distance_fee` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '里程费用',
  `time_fee` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '时长费用',
  `night_fee` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '夜间附加费',
  `dynamic_fee` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '动态调价费用',
  `discount_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '优惠券抵扣金额',
  `platform_subsidy` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '平台补贴金额',
  `payable_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '乘客实付金额',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1预估 2已确认',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_id` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单价格明细表';

-- 支付单表：保存每笔订单的支付/退款记录。
-- 必须有这张表：支付是资金行为，需要记录支付渠道、金额、状态和第三方单号，供回调、查询和退款使用。
CREATE TABLE IF NOT EXISTS `payment` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '支付单ID',
  `payment_no` VARCHAR(32) NOT NULL COMMENT '平台支付单号',
  `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '支付用户ID',
  `amount` BIGINT NOT NULL COMMENT '支付金额（分）',
  `channel` VARCHAR(20) NOT NULL DEFAULT 'wechat' COMMENT '支付渠道：wechat/alipay/balance',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待支付 2支付成功 3支付失败 4已退款',
  `transaction_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '第三方支付流水号',
  `refund_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '已退款金额（分）',
  `event_sent` TINYINT NOT NULL DEFAULT 0 COMMENT '支付成功事件是否已发送：0-未发送 1-已发送',
  `paid_at` DATETIME DEFAULT NULL COMMENT '支付成功时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_payment_no` (`payment_no`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_user_status` (`user_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付单表';

-- 结算表：保存每笔订单的司机收入、平台抽成和结算状态。
-- 必须有这张表：司机端收入、后台对账、财务结算都依赖每笔订单的结算快照。
CREATE TABLE IF NOT EXISTS `settlement` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '结算ID',
  `settlement_no` VARCHAR(32) NOT NULL COMMENT '结算单号',
  `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  `driver_id` BIGINT UNSIGNED NOT NULL COMMENT '司机ID',
  `total_amount` DECIMAL(10,2) NOT NULL COMMENT '订单实际总金额',
  `platform_commission_rate` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '平台抽成比例（%）',
  `platform_commission` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '平台抽成金额',
  `driver_income` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '司机收入',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待结算 2已结算',
  `settled_at` DATETIME DEFAULT NULL COMMENT '结算时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_settlement_no` (`settlement_no`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_driver_status` (`driver_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单结算表';

-- =============================================================
-- 初始数据：计价规则 seed 已拆分至 07_price_rule_seed.sql
-- =============================================================
