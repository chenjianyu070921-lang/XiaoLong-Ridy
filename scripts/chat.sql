-- 司乘聊天（IM）模块统一实施方案 V1.0 建表脚本
-- 数据库：xiaolong_ridy（与现有业务库一致）
-- 说明：本服务也支持启动时 AutoMigrate 自动建表，此 SQL 作为线上迁移与字段口径的权威来源。

CREATE TABLE IF NOT EXISTS `im_conversation` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_id`          VARCHAR(64)    NOT NULL COMMENT '订单号，唯一',
  `driver_id`         BIGINT         NOT NULL DEFAULT 0 COMMENT '司机ID',
  `passenger_id`      BIGINT         NOT NULL DEFAULT 0 COMMENT '乘客ID',
  `status`            TINYINT        NOT NULL DEFAULT 1 COMMENT '1进行中 2已归档 3已关闭',
  `last_msg`          VARCHAR(500)   NOT NULL DEFAULT '' COMMENT '最后一条消息摘要',
  `last_msg_at`       DATETIME      NULL COMMENT '最后一条消息时间',
  `unread_driver`     INT            NOT NULL DEFAULT 0 COMMENT '司机未读数',
  `unread_passenger`  INT            NOT NULL DEFAULT 0 COMMENT '乘客未读数',
  `create_at`         DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='司乘聊天会话';

CREATE TABLE IF NOT EXISTS `im_message` (
  `id`               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `conversation_id`  BIGINT         NOT NULL COMMENT '会话ID',
  `order_id`         VARCHAR(64)    NOT NULL COMMENT '订单号',
  `sender_type`      TINYINT        NOT NULL COMMENT '发送方：1司机 2乘客',
  `sender_id`        BIGINT         NOT NULL COMMENT '发送人ID',
  `msg_type`         TINYINT        NOT NULL DEFAULT 1 COMMENT '1文本 2快捷短语',
  `content`          VARCHAR(500)   NOT NULL COMMENT '消息内容',
  `client_msg_id`    VARCHAR(64)    NOT NULL COMMENT '客户端幂等ID',
  `create_at`        DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_conversation` (`conversation_id`, `id`),
  UNIQUE KEY `uk_client_msg` (`client_msg_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='司乘聊天消息';
