-- 司机回家目的地 / 顺路回家模式设置表（driversvc）。
-- 司机设置回家目的地并开启回家模式后，派单与大厅推单只推送路线顺路的订单。

CREATE TABLE IF NOT EXISTS `driver_home_destination` (
  `id`                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `driver_id`         BIGINT UNSIGNED NOT NULL                COMMENT '司机 ID',
  `home_addr`         VARCHAR(255) NOT NULL DEFAULT ''        COMMENT '回家目的地文字地址',
  `home_lng`          DOUBLE       NOT NULL DEFAULT 0         COMMENT '回家目的地经度',
  `home_lat`          DOUBLE       NOT NULL DEFAULT 0         COMMENT '回家目的地纬度',
  `is_home_mode_open` TINYINT(1)  NOT NULL DEFAULT 0         COMMENT '是否开启回家顺路模式：0关 1开',
  `max_detour_ratio`  DOUBLE       NOT NULL DEFAULT 0.2       COMMENT '最大绕路比例：顺路判定允许相比直达回家多出的路程占比',
  `create_at`         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_at`         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_driver_id` (`driver_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '司机回家目的地与顺路回家模式设置';
