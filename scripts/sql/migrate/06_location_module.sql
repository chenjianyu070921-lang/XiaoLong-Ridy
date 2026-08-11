-- =============================================================
-- 模块六：位置服务与基础设施（locationsvc + pushesvc + location-consumer + job + common）
-- 表清单：driver_location、ride_track_point、geofence、push_message
-- =============================================================

-- 司机实时位置表：保存司机当前最新位置和在线状态。
-- 必须有这张表：派单引擎需要按实时位置找附近司机，乘客端也要展示司机接驾位置。
CREATE TABLE IF NOT EXISTS `driver_location` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `driver_id` BIGINT UNSIGNED NOT NULL COMMENT '司机ID',
  `longitude` DECIMAL(10,6) NOT NULL COMMENT '经度',
  `latitude` DECIMAL(10,6) NOT NULL COMMENT '纬度',
  `heading` SMALLINT NOT NULL DEFAULT 0 COMMENT '行驶方向（0-359度）',
  `speed_kmh` DECIMAL(5,1) NOT NULL DEFAULT 0.0 COMMENT '当前速度（km/h）',
  `online_status` TINYINT NOT NULL DEFAULT 0 COMMENT '听单状态：0离线 1在线 2行程中',
  `report_time` DATETIME NOT NULL COMMENT '位置上报时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '写入时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_driver_id` (`driver_id`),
  KEY `idx_online_status` (`online_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='司机实时位置表';

-- 行程轨迹点表：保存订单行程中的轨迹点。
-- 必须有这张表：行程结束按轨迹校验里程、乘客端展示路线回放、客服处理绕路投诉都要用到。
CREATE TABLE IF NOT EXISTS `ride_track_point` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '轨迹点ID',
  `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  `driver_id` BIGINT UNSIGNED NOT NULL COMMENT '司机ID',
  `longitude` DECIMAL(10,6) NOT NULL COMMENT '经度',
  `latitude` DECIMAL(10,6) NOT NULL COMMENT '纬度',
  `speed_kmh` DECIMAL(5,1) NOT NULL DEFAULT 0.0 COMMENT '速度（km/h）',
  `direction` SMALLINT NOT NULL DEFAULT 0 COMMENT '方向（0-359度）',
  `recorded_at` DATETIME NOT NULL COMMENT '轨迹时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '写入时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_time` (`order_id`, `recorded_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='行程轨迹点表';

-- 电子围栏表：配置运营区域、禁运区域。
-- 必须有这张表：区域运力调度、禁运区校验、运营范围限制都需要按区域配置判断。
CREATE TABLE IF NOT EXISTS `geofence` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '围栏ID',
  `name` VARCHAR(100) NOT NULL COMMENT '围栏名称',
  `area_type` TINYINT NOT NULL DEFAULT 1 COMMENT '类型：1运营区 2禁运区 3热区',
  `center_longitude` DECIMAL(10,6) NOT NULL COMMENT '中心经度',
  `center_latitude` DECIMAL(10,6) NOT NULL COMMENT '中心纬度',
  `radius_m` INT NOT NULL DEFAULT 1000 COMMENT '半径（米）',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1启用 2停用',
  `remark` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_type_status` (`area_type`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='电子围栏表';

-- 消息推送表：保存站内信、App 推送、短信等消息记录。
-- 必须有这张表：订单状态、验证码、活动通知都要统一记录发送状态，失败时支持重试和排查。
CREATE TABLE IF NOT EXISTS `push_message` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '消息ID',
  `biz_type` VARCHAR(30) NOT NULL COMMENT '业务类型：order/activity/system/verify_code',
  `target_type` VARCHAR(20) NOT NULL DEFAULT 'user' COMMENT '接收方类型：user/driver',
  `target_id` BIGINT UNSIGNED NOT NULL COMMENT '接收方ID',
  `order_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联订单ID，无则0',
  `title` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '消息标题',
  `content` VARCHAR(500) NOT NULL COMMENT '消息内容',
  `channel` VARCHAR(20) NOT NULL DEFAULT 'app' COMMENT '渠道：app/sms/ws',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待发送 2已发送 3发送失败',
  `send_at` DATETIME DEFAULT NULL COMMENT '发送时间',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_target_status` (`target_type`, `target_id`, `status`),
  KEY `idx_biz_status` (`biz_type`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息推送记录表';
