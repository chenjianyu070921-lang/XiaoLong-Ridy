-- =============================================================
-- 模块四：派单 mock 种子数据（P0 联调用）
-- 说明：dispatch_record 表已由 04_order_dispatch_module.sql 创建，
--       本脚本仅为 mock 直派提供 3 个在线司机和实时位置。
-- =============================================================

INSERT IGNORE INTO `driver` (`id`, `phone`, `password_hash`, `real_name`, `id_card_no`, `driver_license_no`, `avatar_url`, `status`, `created_at`, `updated_at`) VALUES
  (9001, '13900009001', '', '测试司机A', '110101199001010011', 'A1101019900101', '', 2, NOW(), NOW()),
  (9002, '13900009002', '', '测试司机B', '110101199001010022', 'A1101019900102', '', 2, NOW(), NOW()),
  (9003, '13900009003', '', '测试司机C', '110101199001010033', 'A1101019900103', '', 2, NOW(), NOW());

INSERT IGNORE INTO `driver_location` (`driver_id`, `longitude`, `latitude`, `heading`, `speed_kmh`, `online_status`, `report_time`, `created_at`) VALUES
  (9001, 116.470000, 39.900000, 0, 0.0, 1, NOW(), NOW()),
  (9002, 116.480000, 39.910000, 0, 0.0, 1, NOW(), NOW()),
  (9003, 116.460000, 39.890000, 0, 0.0, 1, NOW(), NOW());
