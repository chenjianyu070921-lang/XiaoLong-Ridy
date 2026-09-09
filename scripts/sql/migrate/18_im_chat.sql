-- 司乘聊天（IM）模块建表
-- 说明：按「司乘聊天IM模块统一实施方案 V1.0」+ 项目实际微调
--   1) order_id 使用 BIGINT，对齐 ordersvc.order_id（与订单表 JOIN / 索引起效）；
--      另增 order_no VARCHAR(64) 存业务单号供前端显示。
--   2) 会话状态枚举采用方案B：1进行中 / 2已归档 / 3已关闭；
--      与订单状态映射（详见 chatsvc 逻辑）：
--        ACCEPTED(2)/ON_TRIP(3)/WAIT_PAY(4) → 1 进行中（可收发）
--        COMPLETED(5)                       → 2 已归档（只读）
--        CANCELLED(6)/REFUNDED(7)           → 3 已关闭
--   3) im_message 增加 client_msg_id 唯一键，保证发送幂等（同一 ID 重复提交只落一条）。

CREATE TABLE IF NOT EXISTS im_conversation (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  order_id        BIGINT       NOT NULL COMMENT '订单ID（对齐 ordersvc order_id）',
  order_no        VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '订单业务号（显示用）',
  driver_id       BIGINT       NOT NULL COMMENT '司机ID',
  passenger_id    BIGINT       NOT NULL COMMENT '乘客ID',
  status          TINYINT      NOT NULL DEFAULT 1 COMMENT '1进行中 2已归档 3已关闭',
  last_msg        VARCHAR(500) NOT NULL DEFAULT '' COMMENT '最后一条消息摘要',
  last_msg_at     DATETIME     NULL COMMENT '最后一条消息时间',
  unread_driver   INT          NOT NULL DEFAULT 0 COMMENT '司机未读数',
  unread_passenger INT         NOT NULL DEFAULT 0 COMMENT '乘客未读数',
  create_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order (order_id),
  KEY idx_driver (driver_id),
  KEY idx_passenger (passenger_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='司乘聊天会话';

CREATE TABLE IF NOT EXISTS im_message (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  conversation_id BIGINT       NOT NULL COMMENT '会话ID',
  order_id        BIGINT       NOT NULL COMMENT '订单ID（对齐订单表）',
  sender_type     TINYINT      NOT NULL COMMENT '发送方：1司机 2乘客',
  sender_id       BIGINT       NOT NULL COMMENT '发送人ID',
  msg_type        TINYINT      NOT NULL DEFAULT 1 COMMENT '1文本 2快捷短语',
  content         VARCHAR(500) NOT NULL COMMENT '消息内容',
  client_msg_id   VARCHAR(64)  NOT NULL COMMENT '客户端幂等ID',
  create_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_conversation (conversation_id, id),
  UNIQUE KEY uk_client_msg (client_msg_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='司乘聊天消息';
