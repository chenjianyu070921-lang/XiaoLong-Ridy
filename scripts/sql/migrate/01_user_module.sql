-- =============================================================
-- 模块一：乘客端 / 用户服务（rpc/usersvc + api/passenger）
-- 表清单：user、user_address
-- =============================================================

-- 用户表：保存乘客账号、登录凭证和基本资料。
-- 必须有这张表：所有乘客业务（下单、订单、优惠券、行程记录）都以用户 ID 为身份基础。
CREATE TABLE IF NOT EXISTS `user` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `phone` VARCHAR(20) NOT NULL COMMENT '手机号（登录账号）',
  `password_hash` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '密码哈希，验证码/第三方登录可为空',
  `nickname` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '昵称',
  `avatar_url` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '头像地址',
  `gender` TINYINT NOT NULL DEFAULT 0 COMMENT '性别：0未知 1男 2女',
  `real_name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '实名认证姓名',
  `id_card_no` VARCHAR(30) NOT NULL DEFAULT '' COMMENT '实名认证身份证号',
  `register_source` VARCHAR(20) NOT NULL DEFAULT 'phone' COMMENT '注册来源：phone/wechat/alipay',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '账号状态：1正常 2冻结',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='乘客用户表';

-- 常用地址表：保存乘客保存的常用地址。
-- 必须有这张表：叫车时支持“家/公司/常用地址”一键选择，减少重复输入，也用于地址搜索联想。
CREATE TABLE IF NOT EXISTS `user_address` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '地址ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `contact_name` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '联系人姓名',
  `contact_phone` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '联系人电话',
  `tag` VARCHAR(20) NOT NULL DEFAULT 'other' COMMENT '地址标签：home/work/other',
  `address` VARCHAR(255) NOT NULL COMMENT '详细地址',
  `longitude` DECIMAL(10,6) NOT NULL COMMENT '经度',
  `latitude` DECIMAL(10,6) NOT NULL COMMENT '纬度',
  `is_default` TINYINT NOT NULL DEFAULT 0 COMMENT '是否默认地址：0否 1是',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_user_tag` (`user_id`, `tag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='乘客常用地址表';
