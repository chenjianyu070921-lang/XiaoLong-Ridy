-- =============================================================
-- 模块一：乘客端评价（api/passenger）
-- 表清单：order_review
-- =============================================================

-- 订单评价表：保存乘客对已完成订单和司机服务的评分、标签与文字评价。
-- 必须有这张表：乘客端评价接口依赖该表落库；order_id 唯一索引用于防止同一订单重复评价。
CREATE TABLE IF NOT EXISTS `order_review` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '评价ID',
  `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '乘客用户ID',
  `driver_id` BIGINT UNSIGNED NOT NULL COMMENT '司机ID',
  `rating` TINYINT NOT NULL COMMENT '评分：1-5',
  `comment` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '乘客文字评价',
  `tags` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '评价标签，多个标签用逗号分隔',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '评价创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '评价更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_id` (`order_id`),
  KEY `idx_user_created` (`user_id`, `created_at`),
  KEY `idx_driver_created` (`driver_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='乘客订单评价表';
