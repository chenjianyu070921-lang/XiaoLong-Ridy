-- 司机业务迁移：平台提现密码 + 银行卡表
ALTER TABLE `driver` ADD COLUMN `withdraw_password_hash` varchar(255) NOT NULL DEFAULT '' COMMENT '平台提现密码bcrypt' AFTER `id_card_no`;

CREATE TABLE IF NOT EXISTS `driver_bank_card` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `driver_id` bigint unsigned NOT NULL COMMENT '司机ID',
  `bank_name` varchar(50) NOT NULL COMMENT '开户行',
  `card_no` varchar(255) NOT NULL COMMENT '卡号AES密文',
  `card_no_hash` char(64) NOT NULL COMMENT '卡号SHA256(查重/唯一)',
  `holder_name` varchar(50) NOT NULL COMMENT '持卡人姓名(须与司机实名一致)',
  `holder_id_card` varchar(30) NOT NULL COMMENT '持卡人身份证(须与司机实名一致)',
  `reserved_phone` varchar(20) NOT NULL COMMENT '银行预留手机号',
  `withdraw_password_hash` varchar(255) NOT NULL DEFAULT '' COMMENT '司机级平台提现密码bcrypt(冗余)',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1正常 0禁用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_driver_card_hash` (`driver_id`,`card_no_hash`),
  KEY `idx_driver` (`driver_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='司机银行卡表';
