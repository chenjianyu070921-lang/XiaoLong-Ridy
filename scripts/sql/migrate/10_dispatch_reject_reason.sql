-- Add a dedicated driver reject reason column for existing dispatch_record tables.
ALTER TABLE `dispatch_record`
  ADD COLUMN `reject_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '拒单原因' AFTER `remark`;