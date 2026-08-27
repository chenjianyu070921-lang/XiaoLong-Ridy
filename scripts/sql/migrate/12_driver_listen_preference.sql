-- Add driver listen preference for realtime/reservation dispatch filtering.
CREATE TABLE IF NOT EXISTS `driver_listen_preference` (
  `driver_id` BIGINT UNSIGNED NOT NULL COMMENT 'driver id',
  `accept_realtime` TINYINT(1) NOT NULL DEFAULT 1 COMMENT 'accept realtime orders: 0 no, 1 yes',
  `accept_reservation` TINYINT(1) NOT NULL DEFAULT 1 COMMENT 'accept reservation orders: 0 no, 1 yes',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created time',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated time',
  PRIMARY KEY (`driver_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='driver listen preference';