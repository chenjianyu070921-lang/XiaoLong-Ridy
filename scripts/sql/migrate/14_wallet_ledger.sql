-- 钱包账本：记录充值、提现、优惠券抵扣和订单支付等所有资金变动。
CREATE TABLE IF NOT EXISTS `user_wallet` (
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `balance` DECIMAL(12,2) NOT NULL DEFAULT 0 COMMENT '可用余额（元）',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户钱包余额';

CREATE TABLE IF NOT EXISTS `user_wallet_transaction` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '流水ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `type` VARCHAR(20) NOT NULL COMMENT '分类：coupon/recharge/withdraw/order',
  `amount` DECIMAL(12,2) NOT NULL COMMENT '变动金额，收入为正、支出为负',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联订单ID',
  `title` VARCHAR(100) NOT NULL COMMENT '流水标题',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发生时间',
  PRIMARY KEY (`id`), KEY `idx_wallet_user_time` (`user_id`,`created_at`), KEY `idx_wallet_order` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户钱包分类流水';

-- 防止订单支付或回调重试造成重复记账。
ALTER TABLE `user_wallet_transaction` ADD UNIQUE KEY `uk_wallet_order_type` (`user_id`,`order_id`,`type`);
