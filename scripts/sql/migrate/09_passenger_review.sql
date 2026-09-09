-- 乘客评价储存表：直接接收并存储乘客对司机的原始评价，
-- 作为自动评价 Agent（agent/driver）按司机手机号查询评价进行打分的数据源。
-- 对应模型：agent/driver.PassengerReview；仓储：agent/driver.GormReviewStore。

CREATE TABLE IF NOT EXISTS `passenger_review` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `driver_phone`    VARCHAR(20)  NOT NULL                COMMENT '司机手机号（Agent 打分查询键）',
  `order_id`        VARCHAR(64)  NOT NULL DEFAULT ''     COMMENT '关联订单号，可为空',
  `passenger_phone` VARCHAR(20)  NOT NULL DEFAULT ''     COMMENT '乘客手机号，可脱敏存储',
  `rating`          TINYINT      NOT NULL DEFAULT 0      COMMENT '乘客星级 0-5，0 表示纯文字评价',
  `comment`         VARCHAR(500) NOT NULL DEFAULT ''     COMMENT '乘客文字评价（打分主要依据）',
  `tags`            VARCHAR(255) NOT NULL DEFAULT ''     COMMENT '乘客勾选标签，逗号分隔（原始留存）',
  `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '评价时间',
  PRIMARY KEY (`id`),
  KEY `idx_driver_phone_created` (`driver_phone`, `created_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '乘客评价储存表（自动评价 Agent 数据源）';
