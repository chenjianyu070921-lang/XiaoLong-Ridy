-- =============================================================
-- 模块五 P0 ③：支付对账/分账相关表
-- 表：payment_reconcile_diff、payment_channel_reconcile_log
-- 说明：settlement 表追加 auto_settled、settled_job_run_id 字段
-- =============================================================

-- 支付渠道对账差异表：平台支付单 vs 渠道流水不一致时记录，供人工/自动处理。
CREATE TABLE IF NOT EXISTS `payment_reconcile_diff` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '差异ID',
  `payment_no` VARCHAR(32) NOT NULL COMMENT '平台支付单号',
  `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  `run_id` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '触发的对账 job run_id',
  `diff_type` TINYINT NOT NULL COMMENT '差异类型：1平台有渠道无 2平台无渠道有 3金额不一致 4状态不一致',
  `platform_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '平台金额（分）',
  `channel_amount` BIGINT NOT NULL DEFAULT 0 COMMENT '渠道金额（分）',
  `platform_status` TINYINT NOT NULL DEFAULT 0 COMMENT '平台支付单状态',
  `channel_status` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '渠道侧状态',
  `channel_tx_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '渠道流水号',
  `detected_at` DATETIME NOT NULL COMMENT '差异检测时间',
  `resolved_at` DATETIME DEFAULT NULL COMMENT '差异处理时间',
  `resolved_by` VARCHAR(32) DEFAULT NULL COMMENT '处理人/系统',
  `remark` VARCHAR(255) DEFAULT NULL COMMENT '备注',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_payment_no` (`payment_no`),
  KEY `idx_diff_type` (`diff_type`),
  KEY `idx_resolved_at` (`resolved_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付渠道对账差异表';

-- 支付对账执行日志：每次对账 job 执行留痕。
CREATE TABLE IF NOT EXISTS `payment_channel_reconcile_log` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `run_id` VARCHAR(32) NOT NULL COMMENT '本次执行 UUID',
  `started_at` DATETIME NOT NULL COMMENT '开始时间',
  `finished_at` DATETIME DEFAULT NULL COMMENT '结束时间',
  `scanned_count` INT NOT NULL DEFAULT 0 COMMENT '扫描支付单数',
  `diff_count` INT NOT NULL DEFAULT 0 COMMENT '发现差异数',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1执行中 2成功 3失败',
  `error_message` VARCHAR(512) DEFAULT NULL COMMENT '失败原因',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_run_id` (`run_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付对账执行日志';

-- 结算表追加自动结算字段（幂等保障：同一订单自动结算只写一次）
ALTER TABLE `settlement`
  ADD COLUMN `auto_settled` TINYINT NOT NULL DEFAULT 0 COMMENT '是否自动结算：0否 1是' AFTER `status`,
  ADD COLUMN `settled_job_run_id` VARCHAR(32) DEFAULT NULL COMMENT '触发的自动结算 job run_id' AFTER `auto_settled`;
