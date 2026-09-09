-- =============================================================
-- 模块二：司机端（api/driver）
-- 表清单：driver、driver_vehicle、driver_certification、driver_score、driver_withdraw
-- =============================================================

-- 司机表：保存司机账号、实名信息和账号状态。
-- 必须有这张表：司机入驻、登录、接单、结算都以司机 ID 为核心身份。
CREATE TABLE IF NOT EXISTS `driver` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '司机ID',
  `phone` VARCHAR(20) NOT NULL COMMENT '手机号（登录账号）',
  `password_hash` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '密码哈希',
  `real_name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '真实姓名',
  `id_card_no` VARCHAR(30) NOT NULL DEFAULT '' COMMENT '身份证号',
  `driver_license_no` VARCHAR(30) NOT NULL DEFAULT '' COMMENT '驾驶证号',
  `avatar_url` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '头像地址',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '账号状态：1待审核 2正常 3冻结 4注销',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '软删除时间',
  `online_status` TINYINT NOT NULL DEFAULT 0 COMMENT 'driver online status: 0 offline, 1 online, 2 on trip',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_phone` (`phone`),
  UNIQUE KEY `uk_driver_license_no` (`driver_license_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='司机账号表';

-- 车辆表：保存司机绑定的车辆信息。
-- 必须有这张表：一个司机可能换车，车辆资质需要独立维护，行程和接驾展示也需要车型车牌信息。
CREATE TABLE IF NOT EXISTS `driver_vehicle` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '车辆ID',
  `driver_id` BIGINT UNSIGNED NOT NULL COMMENT '司机ID',
  `plate_no` VARCHAR(20) NOT NULL COMMENT '车牌号',
  `brand` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '品牌',
  `model` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '车型',
  `color` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '车身颜色',
  `vehicle_type` TINYINT NOT NULL DEFAULT 1 COMMENT '车辆类型：1特惠快车 2快车 3拼车',
  `registration_date` DATE DEFAULT NULL COMMENT '注册日期',
  `insurance_no` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '保险单号',
  `insurance_expire_at` DATE DEFAULT NULL COMMENT '保险到期日',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待审核 2正常 3禁用',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_plate_no` (`plate_no`),
  KEY `idx_driver_id` (`driver_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='司机车辆表';

-- 司机认证表：保存司机/车辆资料的上传与审核记录。
-- 必须有这张表：司机入驻必须经过资质审核，审核状态需要留痕，后台也要按状态查询待审核列表。
CREATE TABLE IF NOT EXISTS `driver_certification` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '认证ID',
  `driver_id` BIGINT UNSIGNED NOT NULL COMMENT '司机ID',
  `vehicle_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '车辆ID',
  `id_card_front_url` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '身份证人像面',
  `id_card_back_url` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '身份证国徽面',
  `driver_license_url` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '驾驶证照片',
  `vehicle_license_url` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '行驶证照片',
  `audit_status` TINYINT NOT NULL DEFAULT 1 COMMENT '审核状态：1待审核 2通过 3驳回',
  `audit_remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '驳回原因/审核备注',
  `audited_by` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '审核人（后台管理员ID）',
  `audited_at` DATETIME DEFAULT NULL COMMENT '审核时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_driver_status` (`driver_id`, `audit_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='司机资质认证表';

-- 司机评分表：保存司机的服务分、等级和运营指标。
-- 必须有这张表：派单引擎需要按评分/完单率排序，司机端也要展示服务分和等级。
CREATE TABLE IF NOT EXISTS `driver_score` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `driver_id` BIGINT UNSIGNED NOT NULL COMMENT '司机ID',
  `score` DECIMAL(5,2) NOT NULL DEFAULT 100.00 COMMENT '服务分',
  `level` TINYINT NOT NULL DEFAULT 1 COMMENT '司机等级：1-5',
  `month_orders` INT NOT NULL DEFAULT 0 COMMENT '当月完单数',
  `month_cancel_rate` DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '当月取消率（%）',
  `month_complaint_count` INT NOT NULL DEFAULT 0 COMMENT '当月投诉数',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_driver_id` (`driver_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='司机服务分表';

-- 司机提现表：保存司机的提现申请和打款结果。
-- 必须有这张表：司机端收入管理需要查看提现记录，财务需要按状态对账和打款。
CREATE TABLE IF NOT EXISTS `driver_withdraw` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '提现ID', 
  `driver_id` BIGINT UNSIGNED NOT NULL COMMENT '司机ID',
  `withdraw_no` VARCHAR(32) NOT NULL COMMENT '提现单号',
  `amount` DECIMAL(10,2) NOT NULL COMMENT '提现金额',
  `payee_name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '收款人姓名',
  `pay_account` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '收款账号',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1申请中 2打款成功 3打款失败',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '失败原因/备注',
  `applied_at` DATETIME DEFAULT NULL COMMENT '申请时间',
  `paid_at` DATETIME DEFAULT NULL COMMENT '打款时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_withdraw_no` (`withdraw_no`),
  KEY `idx_driver_status` (`driver_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='司机提现表';
