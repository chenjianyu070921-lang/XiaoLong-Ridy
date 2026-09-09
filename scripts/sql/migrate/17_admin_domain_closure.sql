-- =============================================================
-- 模块三：管理后台第一期业务闭环补充表
-- 适用范围：rpc/adminsvc、job、下游事件消费者。
-- 执行约束：
-- 1. 本文件仅定义新增表和索引，不修改已有业务表或业务数据。
-- 2. 代码不会在启动时自动执行本文件，必须由数据库发布流程显式执行。
-- 3. 所有跨服务写操作均先在 adminsvc 本地事务中写入业务状态、审计日志和领域 outbox。
-- =============================================================

-- 司机处罚规则表：定义后台可配置的处罚动作、金额、分值和版本。
CREATE TABLE IF NOT EXISTS `admin_driver_punishment_rule` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '处罚规则ID',
  `name` VARCHAR(100) NOT NULL COMMENT '规则名称',
  `violation_type` VARCHAR(50) NOT NULL COMMENT '违规类型，例如拒载、绕路、恶意取消',
  `actions` JSON NOT NULL COMMENT '处罚动作JSON，例如禁接单、冻结、扣分、罚款、降权',
  `penalty_cents` BIGINT NOT NULL DEFAULT 0 COMMENT '罚款金额，单位分；仅记录待结算金额，不直接扣余额',
  `score_delta` INT NOT NULL DEFAULT 0 COMMENT '司机服务分变更值，扣分使用负数',
  `priority_weight_delta` INT NOT NULL DEFAULT 0 COMMENT '派单权重变更值，降权使用负数',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 2停用',
  `version` INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '规则版本，用于发布和审计追踪',
  `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建管理员ID',
  `updated_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '最后更新管理员ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_violation_status` (`violation_type`, `status`),
  KEY `idx_status_updated` (`status`, `updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理后台司机处罚规则表';

-- 司机处罚单表：保存后台处罚请求及其最终执行状态，不复制 driversvc 的司机主数据。
CREATE TABLE IF NOT EXISTS `admin_driver_punishment` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '处罚单ID',
  `punishment_no` VARCHAR(64) NOT NULL COMMENT '处罚单编号',
  `driver_id` BIGINT UNSIGNED NOT NULL COMMENT '司机ID',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联订单ID，无订单来源时为0',
  `rule_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '处罚规则ID，人工处罚时可为0',
  `actions` JSON NOT NULL COMMENT '处罚动作JSON',
  `reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '处罚原因',
  `penalty_cents` BIGINT NOT NULL DEFAULT 0 COMMENT '待结算罚款金额，单位分',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态：pending/processing/effective/failed/cancelled',
  `request_id` VARCHAR(64) NOT NULL COMMENT '后台请求幂等号',
  `failure_reason` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '最近失败原因',
  `created_by` BIGINT UNSIGNED NOT NULL COMMENT '创建管理员ID',
  `cancelled_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '撤销管理员ID',
  `cancelled_at` DATETIME DEFAULT NULL COMMENT '撤销时间',
  `effective_at` DATETIME DEFAULT NULL COMMENT '生效时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_punishment_no` (`punishment_no`),
  UNIQUE KEY `uk_punishment_request` (`driver_id`, `request_id`),
  KEY `idx_driver_status` (`driver_id`, `status`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_status_created` (`status`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理后台司机处罚单表';

-- 司机处罚申诉表：保存司机的申诉材料索引和后台复核结果。
CREATE TABLE IF NOT EXISTS `admin_punishment_appeal` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '处罚申诉ID',
  `appeal_no` VARCHAR(64) NOT NULL COMMENT '申诉编号',
  `punishment_id` BIGINT UNSIGNED NOT NULL COMMENT '处罚单ID',
  `driver_id` BIGINT UNSIGNED NOT NULL COMMENT '司机ID',
  `content` VARCHAR(2000) NOT NULL DEFAULT '' COMMENT '申诉说明',
  `evidence_config` JSON NULL COMMENT '证据索引JSON，仅保存资源标识或地址',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态：pending/reviewing/upheld/revoked/rejected',
  `review_result` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '复核结论',
  `reviewed_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '复核管理员ID',
  `reviewed_at` DATETIME DEFAULT NULL COMMENT '复核时间',
  `request_id` VARCHAR(64) NOT NULL COMMENT '申诉请求幂等号',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_appeal_no` (`appeal_no`),
  UNIQUE KEY `uk_appeal_request` (`punishment_id`, `request_id`),
  KEY `idx_punishment_status` (`punishment_id`, `status`),
  KEY `idx_driver_created` (`driver_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理后台司机处罚申诉表';

-- 退款补偿任务表：接管退款失败、超时后的持久化补偿状态，不依赖单实例 Redis 队列。
CREATE TABLE IF NOT EXISTS `admin_refund_compensation_task` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '退款补偿任务ID',
  `task_no` VARCHAR(64) NOT NULL COMMENT '退款补偿任务编号',
  `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  `refund_no` VARCHAR(64) NOT NULL COMMENT '退款单号',
  `refund_cents` BIGINT NOT NULL COMMENT '退款金额，单位分',
  `reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '退款原因',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态：pending/processing/retrying/success/manual_review/failed',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '已重试次数',
  `max_retry` INT NOT NULL DEFAULT 5 COMMENT '最大自动重试次数',
  `next_retry_at` DATETIME DEFAULT NULL COMMENT '下次重试时间',
  `last_response` VARCHAR(2000) NOT NULL DEFAULT '' COMMENT '支付通道或下游最近响应摘要',
  `failure_reason` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '最近失败原因',
  `request_id` VARCHAR(64) NOT NULL COMMENT '后台请求幂等号',
  `created_by` BIGINT UNSIGNED NOT NULL COMMENT '创建管理员ID',
  `handled_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '人工处理管理员ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_refund_task_no` (`task_no`),
  UNIQUE KEY `uk_refund_no` (`refund_no`),
  KEY `idx_status_retry` (`status`, `next_retry_at`),
  KEY `idx_order_id` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理后台退款补偿任务表';

-- 异步发券批次表：记录批量发券的分片进度和结果，实际用户券写入由 usersvc 执行。
CREATE TABLE IF NOT EXISTS `admin_coupon_issue_batch` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '发券批次ID',
  `batch_no` VARCHAR(64) NOT NULL COMMENT '发券批次编号',
  `coupon_id` BIGINT UNSIGNED NOT NULL COMMENT '优惠券模板ID',
  `target_type` VARCHAR(32) NOT NULL COMMENT '目标类型，例如explicit_users/crowd',
  `target_config` JSON NOT NULL COMMENT '目标配置JSON',
  `total_count` INT NOT NULL DEFAULT 0 COMMENT '计划发放总数',
  `success_count` INT NOT NULL DEFAULT 0 COMMENT '成功数量',
  `fail_count` INT NOT NULL DEFAULT 0 COMMENT '失败数量',
  `processing_count` INT NOT NULL DEFAULT 0 COMMENT '处理中数量',
  `cursor_offset` INT NOT NULL DEFAULT 0 COMMENT '已完成分片游标',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态：pending/processing/partial_success/success/failed/cancelled',
  `request_id` VARCHAR(64) NOT NULL COMMENT '后台请求幂等号',
  `failure_reason` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '最近失败原因',
  `created_by` BIGINT UNSIGNED NOT NULL COMMENT '创建管理员ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_coupon_batch_no` (`batch_no`),
  UNIQUE KEY `uk_coupon_batch_request` (`coupon_id`, `request_id`),
  KEY `idx_status_created` (`status`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理后台异步发券批次表';

-- 活动发布任务表：记录真实冲突检查、规则版本、发布与回滚的最终状态。
CREATE TABLE IF NOT EXISTS `admin_promotion_publish_task` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '活动发布任务ID',
  `task_no` VARCHAR(64) NOT NULL COMMENT '活动发布任务编号',
  `promotion_id` BIGINT UNSIGNED NOT NULL COMMENT '营销活动ID',
  `action` VARCHAR(32) NOT NULL COMMENT '动作：publish/rollback',
  `publish_scope` VARCHAR(32) NOT NULL DEFAULT 'full' COMMENT '范围：full/gray',
  `target_config` JSON NULL COMMENT '灰度目标配置JSON',
  `rule_version` VARCHAR(64) NOT NULL COMMENT '待发布或回滚的规则版本',
  `conflict_result` JSON NULL COMMENT 'pricesvc 真实冲突检测结果',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态：pending/checking/publishing/success/retrying/failed',
  `request_id` VARCHAR(64) NOT NULL COMMENT '后台请求幂等号',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  `next_retry_at` DATETIME DEFAULT NULL COMMENT '下次重试时间',
  `failure_reason` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '最近失败原因',
  `created_by` BIGINT UNSIGNED NOT NULL COMMENT '创建管理员ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_promotion_task_no` (`task_no`),
  UNIQUE KEY `uk_promotion_request` (`promotion_id`, `action`, `request_id`),
  KEY `idx_status_retry` (`status`, `next_retry_at`),
  KEY `idx_promotion_created` (`promotion_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理后台活动发布任务表';

-- 统一领域 outbox：保证后台业务提交成功后跨服务事件可可靠投递。
CREATE TABLE IF NOT EXISTS `admin_domain_outbox` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '领域事件主键',
  `event_no` VARCHAR(64) NOT NULL COMMENT '全局唯一事件编号',
  `event_type` VARCHAR(100) NOT NULL COMMENT '事件类型',
  `aggregate_type` VARCHAR(64) NOT NULL COMMENT '聚合对象类型',
  `aggregate_id` BIGINT UNSIGNED NOT NULL COMMENT '聚合对象ID',
  `request_id` VARCHAR(64) NOT NULL COMMENT '业务请求幂等号',
  `payload` JSON NOT NULL COMMENT '脱敏事件载荷JSON',
  `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态：pending/running/success/failed',
  `retry_count` INT NOT NULL DEFAULT 0 COMMENT '已投递次数',
  `max_retry` INT NOT NULL DEFAULT 5 COMMENT '最大投递次数',
  `next_retry_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '下次投递时间',
  `lease_owner` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '当前任务租约持有者',
  `lease_expires_at` DATETIME DEFAULT NULL COMMENT '租约过期时间',
  `failure_reason` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '最近投递失败原因',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_domain_event_no` (`event_no`),
  UNIQUE KEY `uk_domain_outbox_request` (`event_type`, `aggregate_type`, `aggregate_id`, `request_id`),
  KEY `idx_domain_outbox_dispatch` (`status`, `next_retry_at`),
  KEY `idx_domain_outbox_lease` (`status`, `lease_expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理后台领域可靠事件表';

-- 司机域处罚效果幂等表：driversvc 消费处罚事件时使用。
CREATE TABLE IF NOT EXISTS `driver_punishment_effect` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `event_id` VARCHAR(128) NOT NULL,
  `punishment_no` VARCHAR(64) NOT NULL,
  `driver_id` BIGINT UNSIGNED NOT NULL,
  `action_type` VARCHAR(32) NOT NULL,
  `score_delta` INT NOT NULL DEFAULT 0,
  `priority_weight_delta` INT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_driver_punishment_event` (`event_id`, `action_type`),
  KEY `idx_driver_punishment_driver` (`driver_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='司机处罚效果幂等表';
