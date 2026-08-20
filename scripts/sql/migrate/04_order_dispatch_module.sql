-- =============================================================
-- 模块四：订单与派单调度（rpc/ordersvc + rpc/dispatchsvc + orderclient-event-consumer）
-- 表清单：ride_order、order_status_log、dispatch_record
-- =============================================================

-- 订单主表：保存每一笔打车订单的完整信息。
-- 必须有这张表：订单是整个系统的核心业务实体，乘客端、司机端、派单、计价、支付都围绕订单 ID 协作。
CREATE TABLE IF NOT EXISTS `ride_order` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `order_no` VARCHAR(32) NOT NULL COMMENT '订单号（对外展示）',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '乘客用户ID',
  `driver_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '司机ID，未接单为0',
  `car_type` TINYINT NOT NULL DEFAULT 1 COMMENT '车型：1特惠快车 2快车 3拼车',
  `from_address` VARCHAR(255) NOT NULL COMMENT '起点地址',
  `from_longitude` DECIMAL(10,6) NOT NULL COMMENT '起点经度',
  `from_latitude` DECIMAL(10,6) NOT NULL COMMENT '起点纬度',
  `to_address` VARCHAR(255) NOT NULL COMMENT '终点地址',
  `to_longitude` DECIMAL(10,6) NOT NULL COMMENT '终点经度',
  `to_latitude` DECIMAL(10,6) NOT NULL COMMENT '终点纬度',
  `estimated_distance_m` INT NOT NULL DEFAULT 0 COMMENT '预估距离（米）',
  `estimated_duration_s` INT NOT NULL DEFAULT 0 COMMENT '预估时长（秒）',
  `estimated_price` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '预估价格',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待接单 2已接单 3行程中 4待支付 5已完成 6已取消',
  `cancel_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '取消原因',
  `cancel_by` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '取消方：user/driver/system/admin',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '下单时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_user_status` (`user_id`, `status`),
  KEY `idx_driver_status` (`driver_id`, `status`),
  KEY `idx_status_created` (`status`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='打车订单主表';

-- 订单状态日志表：记录订单每一次状态流转。
-- 必须有这张表：订单状态机需要完整留痕，方便排查异常订单、统计取消原因和支撑客服处理。
CREATE TABLE IF NOT EXISTS `order_status_log` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  `from_status` TINYINT NOT NULL COMMENT '原状态',
  `to_status` TINYINT NOT NULL COMMENT '新状态',
  `operator_type` VARCHAR(20) NOT NULL DEFAULT 'system' COMMENT '操作方：user/driver/system/admin',
  `operator_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作方ID',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注/原因',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单状态日志表';

-- 派单记录表：保存每笔订单向候选司机的派单/抢单过程。
-- 必须有这张表：派单引擎需要记录派了哪些司机、谁接受谁拒绝，用于匹配算法调优和问题排查。
CREATE TABLE IF NOT EXISTS `dispatch_record` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '派单记录ID',
  `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  `driver_id` BIGINT UNSIGNED NOT NULL COMMENT '候选司机ID',
  `dispatch_type` TINYINT NOT NULL DEFAULT 1 COMMENT '派单方式：1自动派单 2抢单 3改派',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1派单中 2已接受 3已拒绝 4超时 5已取消',
  `match_score` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '匹配分（距离/评分/顺路度加权）',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '派单时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_driver_status` (`driver_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='派单记录表';
