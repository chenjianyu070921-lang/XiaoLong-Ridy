-- =============================================================
-- 模块三：管理后台（api/admin）
-- 表清单：admin_user、admin_operation_log、coupon、user_coupon、promotion_activity、blacklist
-- =============================================================

-- 后台管理员表：保存后台登录账号、角色和状态。
-- 必须有这张表：管理后台需要独立的账号体系，区分超级管理员和普通运营人员。
CREATE TABLE IF NOT EXISTS `admin_user` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '管理员ID',
  `username` VARCHAR(50) NOT NULL COMMENT '登录用户名',
  `password_hash` VARCHAR(255) NOT NULL COMMENT '密码哈希',
  `real_name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '姓名',
  `role` TINYINT NOT NULL DEFAULT 2 COMMENT '角色：1超级管理员 2运营 3客服',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1正常 2禁用',
  `last_login_at` DATETIME DEFAULT NULL COMMENT '最后登录时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台管理员表';

-- 操作日志表：记录后台管理员的敏感操作。
-- 必须有这张表：审核、封禁、改派、调价等操作需要可追溯，出现纠纷时能定位是谁在什么时间做了什么。
CREATE TABLE IF NOT EXISTS `admin_operation_log` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `admin_id` BIGINT UNSIGNED NOT NULL COMMENT '管理员ID',
  `module` VARCHAR(50) NOT NULL COMMENT '操作模块：driver/order/coupon/risk等',
  `action` VARCHAR(50) NOT NULL COMMENT '操作动作：audit/ban/change等',
  `target_type` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '操作对象类型',
  `target_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作对象ID',
  `detail` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '操作内容/备注',
  `ip` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '操作IP',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
  PRIMARY KEY (`id`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_target` (`target_type`, `target_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台操作日志表';

-- 优惠券模板表：保存平台创建的优惠券批次规则。
-- 必须有这张表：优惠券由后台配置，乘客端和计价模块都按模板发放和抵扣。
CREATE TABLE IF NOT EXISTS `coupon` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '优惠券模板ID',
  `name` VARCHAR(50) NOT NULL COMMENT '优惠券名称',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT '类型：1满减 2折扣 3立减',
  `face_value` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '减免/立减金额',
  `discount` DECIMAL(3,2) NOT NULL DEFAULT 1.00 COMMENT '折扣率，如0.80表示8折',
  `threshold_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '使用门槛金额，0表示无门槛',
  `total_count` INT NOT NULL DEFAULT 0 COMMENT '发行总量，0表示不限',
  `received_count` INT NOT NULL DEFAULT 0 COMMENT '已领取数量',
  `per_user_limit` INT NOT NULL DEFAULT 1 COMMENT '每人限领数量',
  `valid_start_at` DATETIME NOT NULL COMMENT '生效时间',
  `valid_end_at` DATETIME NOT NULL COMMENT '失效时间',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 2停用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优惠券模板表';

-- 用户优惠券表：保存每个用户领取到的优惠券实例。
-- 必须有这张表：优惠券是用户维度的资产，使用状态和抵扣订单必须精确记录。
CREATE TABLE IF NOT EXISTS `user_coupon` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户券ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `coupon_id` BIGINT UNSIGNED NOT NULL COMMENT '优惠券模板ID',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '核销订单ID，未使用为0',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1未使用 2已使用 3已过期',
  `received_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '领取时间',
  `used_at` DATETIME DEFAULT NULL COMMENT '使用时间',
  `expire_at` DATETIME NOT NULL COMMENT '过期时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_status` (`user_id`, `status`),
  KEY `idx_coupon_id` (`coupon_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户优惠券表';

-- 营销活动表：保存新人礼、邀请有礼、限时活动等配置。
-- 必须有这张表：拉新和补贴活动需要独立配置生效时间、奖励规则和状态。
CREATE TABLE IF NOT EXISTS `promotion_activity` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '活动ID',
  `name` VARCHAR(100) NOT NULL COMMENT '活动名称',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT '类型：1新人礼 2邀请有礼 3限时活动',
  `config` TEXT NOT NULL COMMENT '活动配置JSON（奖励内容/参与条件）',
  `start_at` DATETIME NOT NULL COMMENT '开始时间',
  `end_at` DATETIME NOT NULL COMMENT '结束时间',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1未开始 2进行中 3已结束',
  `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建管理员ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='营销活动表';

-- 黑名单表：保存用户/司机的风控黑名单记录。
-- 必须有这张表：刷单、恶意取消、违规司机需要被封禁和留痕，登录/下单/接单前都要校验。
CREATE TABLE IF NOT EXISTS `blacklist` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '黑名单ID',
  `target_type` VARCHAR(20) NOT NULL COMMENT '对象类型：user/driver/device',
  `target_id` BIGINT UNSIGNED NOT NULL COMMENT '对象ID',
  `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '拉黑原因',
  `operator_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作管理员ID',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1生效 2已解除',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_target` (`target_type`, `target_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='风控黑名单表';
