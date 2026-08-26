CREATE TABLE IF NOT EXISTS `order_review` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '评价ID',
  `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '乘客用户ID',
  `driver_id` BIGINT UNSIGNED NOT NULL COMMENT '司机ID',
  `rating` TINYINT NOT NULL COMMENT '评分，1至5',
  `comment` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '评价内容',
  `tags` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '评价标签',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_id` (`order_id`),
  KEY `idx_driver_created` (`driver_id`, `created_at`),
  KEY `idx_user_created` (`user_id`, `created_at`),
  CONSTRAINT `chk_order_review_rating` CHECK (`rating` BETWEEN 1 AND 5)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='乘客订单评价';
