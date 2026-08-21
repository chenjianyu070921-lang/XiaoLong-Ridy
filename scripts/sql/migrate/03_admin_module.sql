-- =============================================================
-- Module 03: admin backend (api/admin)
-- Tables: admin_user, admin_operation_log, coupon, user_coupon,
--         promotion_activity, blacklist
-- =============================================================

CREATE TABLE IF NOT EXISTS `admin_user` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'admin id',
  `username` VARCHAR(50) NOT NULL COMMENT 'login username',
  `password_hash` VARCHAR(255) NOT NULL COMMENT 'password hash',
  `real_name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT 'real name',
  `role` TINYINT NOT NULL DEFAULT 2 COMMENT 'role: 1 super admin, 2 ops, 3 customer service',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT 'status: 1 normal, 2 disabled',
  `last_login_at` DATETIME DEFAULT NULL COMMENT 'last login time',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created time',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated time',
  `deleted_at` DATETIME DEFAULT NULL COMMENT 'soft delete time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='admin user table';

CREATE TABLE IF NOT EXISTS `admin_operation_log` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'log id',
  `admin_id` BIGINT UNSIGNED NOT NULL COMMENT 'admin id',
  `module` VARCHAR(50) NOT NULL COMMENT 'module: driver/order/coupon/risk',
  `action` VARCHAR(50) NOT NULL COMMENT 'action: audit/ban/change',
  `target_type` VARCHAR(20) NOT NULL DEFAULT '' COMMENT 'target type',
  `target_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'target id',
  `detail` VARCHAR(500) NOT NULL DEFAULT '' COMMENT 'operation detail',
  `ip` VARCHAR(50) NOT NULL DEFAULT '' COMMENT 'operation ip',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'operation time',
  PRIMARY KEY (`id`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_target` (`target_type`, `target_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='admin operation log table';

CREATE TABLE IF NOT EXISTS `coupon` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'coupon template id',
  `name` VARCHAR(50) NOT NULL COMMENT 'coupon name',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT 'type: 1 threshold, 2 discount, 3 instant',
  `face_value` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT 'face value',
  `discount` DECIMAL(3,2) NOT NULL DEFAULT 1.00 COMMENT 'discount rate',
  `threshold_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT 'threshold amount',
  `total_count` INT NOT NULL DEFAULT 0 COMMENT 'total count, 0 unlimited',
  `received_count` INT NOT NULL DEFAULT 0 COMMENT 'received count',
  `per_user_limit` INT NOT NULL DEFAULT 1 COMMENT 'per user limit',
  `valid_start_at` DATETIME NOT NULL COMMENT 'valid start time',
  `valid_end_at` DATETIME NOT NULL COMMENT 'valid end time',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT 'status: 1 draft, 2 enabled, 3 disabled',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created time',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated time',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='coupon template table';

CREATE TABLE IF NOT EXISTS `user_coupon` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'user coupon id',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT 'user id',
  `coupon_id` BIGINT UNSIGNED NOT NULL COMMENT 'coupon template id',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'used order id, 0 unused',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT 'status: 1 unused, 2 used, 3 expired',
  `received_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'received time',
  `used_at` DATETIME DEFAULT NULL COMMENT 'used time',
  `expire_at` DATETIME NOT NULL COMMENT 'expire time',
  PRIMARY KEY (`id`),
  KEY `idx_user_status` (`user_id`, `status`),
  KEY `idx_coupon_id` (`coupon_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='user coupon table';

CREATE TABLE IF NOT EXISTS `promotion_activity` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'activity id',
  `name` VARCHAR(100) NOT NULL COMMENT 'activity name',
  `type` TINYINT NOT NULL DEFAULT 1 COMMENT 'type: 1 new user, 2 invite, 3 limited time',
  `config` TEXT NOT NULL COMMENT 'activity config json',
  `start_at` DATETIME NOT NULL COMMENT 'start time',
  `end_at` DATETIME NOT NULL COMMENT 'end time',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT 'status: 1 not started, 2 running, 3 ended',
  `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'creator admin id',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created time',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated time',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='promotion activity table';

CREATE TABLE IF NOT EXISTS `blacklist` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'blacklist id',
  `target_type` VARCHAR(20) NOT NULL COMMENT 'target type: user/driver/device',
  `target_id` BIGINT UNSIGNED NOT NULL COMMENT 'target id',
  `reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'blacklist reason',
  `operator_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'operator admin id',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT 'status: 1 active, 2 released',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created time',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated time',
  PRIMARY KEY (`id`),
  KEY `idx_target` (`target_type`, `target_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='risk blacklist table';
