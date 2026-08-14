-- =============================================================
-- 模块三：管理后台运营业务补充表
-- 适用范围：api/admin
-- 说明：
-- 1. 本脚本只定义管理后台运营侧补充表，不直接修改线上业务数据。
-- 2. 03_admin_module.sql 已包含 admin_operation_log、coupon、user_coupon、promotion_activity、blacklist 基础表。
-- 3. 本脚本补充投诉工单、工单流转、风控命中、营销发布/发券任务等运营闭环表。
-- =============================================================

-- 投诉/申诉工单主表：统一承载用户投诉、订单申诉、司机申诉等后台处理任务。
CREATE TABLE IF NOT EXISTS `admin_complaint_work_order` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '工单ID',
  `work_order_no` VARCHAR(32) NOT NULL COMMENT '工单编号',
  `work_order_type` TINYINT NOT NULL COMMENT '工单类型：1用户投诉 2订单申诉 3司机处罚申诉',
  `source_type` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '来源类型：user/driver/order/system',
  `source_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '来源业务ID',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联订单ID',
  `user_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联乘客ID',
  `driver_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联司机ID',
  `title` VARCHAR(100) NOT NULL COMMENT '工单标题',
  `content` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '投诉或申诉内容',
  `priority` TINYINT NOT NULL DEFAULT 2 COMMENT '优先级：1低 2中 3高 4紧急',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待处理 2跟进中 3仲裁完成 4已结案 5已关闭',
  `assignee_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '当前处理人管理员ID',
  `arbitration_result` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '仲裁结果',
  `remark` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '后台备注',
  `version` INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '乐观锁版本号，用于避免多人同时修改覆盖',
  `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建人管理员ID，系统创建时为0',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `closed_at` DATETIME DEFAULT NULL COMMENT '结案或关闭时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_work_order_no` (`work_order_no`),
  KEY `idx_status_priority` (`status`, `priority`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_driver_id` (`driver_id`),
  KEY `idx_assignee_status` (`assignee_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台投诉申诉工单主表';

-- 工单流转记录表：记录每次状态变化、分配、备注和仲裁动作。
CREATE TABLE IF NOT EXISTS `admin_work_order_flow` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '流转记录ID',
  `work_order_id` BIGINT UNSIGNED NOT NULL COMMENT '工单ID',
  `from_status` TINYINT NOT NULL DEFAULT 0 COMMENT '变更前状态',
  `to_status` TINYINT NOT NULL DEFAULT 0 COMMENT '变更后状态',
  `action` VARCHAR(50) NOT NULL COMMENT '动作：assign/follow/arbitrate/close/reopen',
  `operator_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作管理员ID',
  `content` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '处理内容或备注',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_work_order_id` (`work_order_id`),
  KEY `idx_operator_id` (`operator_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台工单流转记录表';

-- 工单证据表：保存轨迹截图、录音、聊天记录、支付凭证等证据索引。
CREATE TABLE IF NOT EXISTS `admin_work_order_evidence` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '证据ID',
  `work_order_id` BIGINT UNSIGNED NOT NULL COMMENT '工单ID',
  `evidence_type` VARCHAR(30) NOT NULL COMMENT '证据类型：track/audio/chat/payment/image/text',
  `evidence_url` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '证据文件或资源地址',
  `content` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '文本证据内容',
  `uploaded_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '上传人管理员ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_work_order_id` (`work_order_id`),
  KEY `idx_type_created` (`evidence_type`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台工单证据表';

-- 优惠券发布记录表：记录优惠券模板发布到计价/营销服务的版本和结果。
CREATE TABLE IF NOT EXISTS `admin_coupon_publish_record` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '发布记录ID',
  `coupon_id` BIGINT UNSIGNED NOT NULL COMMENT '优惠券模板ID',
  `publish_version` VARCHAR(32) NOT NULL COMMENT '发布版本号',
  `publish_scope` VARCHAR(20) NOT NULL DEFAULT 'draft' COMMENT '发布范围：draft/gray/full/rollback',
  `target_config` TEXT NOT NULL COMMENT '目标人群、城市、灰度比例等JSON配置',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待发布 2发布成功 3发布失败 4已回滚',
  `failure_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '失败原因',
  `operator_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作管理员ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_coupon_id` (`coupon_id`),
  KEY `idx_status_created` (`status`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台优惠券发布记录表';

-- 发券任务表：记录批量发券任务，实际发券可由 MQ/Job 异步执行。
CREATE TABLE IF NOT EXISTS `admin_coupon_issue_task` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '发券任务ID',
  `task_no` VARCHAR(32) NOT NULL COMMENT '任务编号',
  `coupon_id` BIGINT UNSIGNED NOT NULL COMMENT '优惠券模板ID',
  `target_type` VARCHAR(20) NOT NULL COMMENT '目标类型：user/batch/crowd',
  `target_config` TEXT NOT NULL COMMENT '目标用户或人群配置JSON',
  `total_count` INT NOT NULL DEFAULT 0 COMMENT '计划发放数量',
  `success_count` INT NOT NULL DEFAULT 0 COMMENT '成功数量',
  `fail_count` INT NOT NULL DEFAULT 0 COMMENT '失败数量',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待执行 2执行中 3成功 4部分失败 5失败',
  `failure_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '失败原因',
  `operator_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作管理员ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task_no` (`task_no`),
  KEY `idx_coupon_id` (`coupon_id`),
  KEY `idx_status_created` (`status`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台发券任务表';

-- 风控黑名单命中记录表：记录用户、司机、设备等命中黑名单或风控规则的过程。
CREATE TABLE IF NOT EXISTS `risk_blacklist_hit_record` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '命中记录ID',
  `blacklist_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联 blacklist ID',
  `target_type` VARCHAR(20) NOT NULL COMMENT '目标类型：user/driver/device/phone',
  `target_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '目标ID',
  `scene` VARCHAR(30) NOT NULL COMMENT '命中场景：login/order/dispatch/pay/refund',
  `risk_level` TINYINT NOT NULL DEFAULT 1 COMMENT '风险等级：1低 2中 3高',
  `hit_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '命中原因',
  `request_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '请求链路ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_target_scene` (`target_type`, `target_id`, `scene`),
  KEY `idx_blacklist_id` (`blacklist_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='风控黑名单命中记录表';
