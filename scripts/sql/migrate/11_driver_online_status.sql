-- Add driver.online_status for existing databases.
SET @driver_online_status_sql = (
  SELECT IF(
    COUNT(*) = 0,
    'ALTER TABLE `driver` ADD COLUMN `online_status` TINYINT NOT NULL DEFAULT 0 COMMENT ''driver online status: 0 offline, 1 online, 2 on trip'' AFTER `status`',
    'SELECT 1'
  )
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'driver'
    AND COLUMN_NAME = 'online_status'
);

PREPARE driver_online_status_stmt FROM @driver_online_status_sql;
EXECUTE driver_online_status_stmt;
DEALLOCATE PREPARE driver_online_status_stmt;
