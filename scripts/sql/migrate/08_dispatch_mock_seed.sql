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

DELETE FROM `dispatch_record` WHERE `order_id` IN (5001, 5002, 5003, 5004, 5005);
DELETE FROM `order_status_log` WHERE `order_id` IN (5001, 5002, 5003, 5004, 5005);
DELETE FROM `ride_order` WHERE `id` IN (5001, 5002, 5003, 5004, 5005);

INSERT INTO `ride_order`
  (`id`, `order_no`, `user_id`, `driver_id`, `car_type`, `from_address`, `from_longitude`, `from_latitude`, `to_address`, `to_longitude`, `to_latitude`, `estimated_distance_m`, `estimated_duration_s`, `estimated_price`, `status`, `cancel_reason`, `cancel_by`, `created_at`, `updated_at`)
VALUES
  (5001, 'MOCK-NEARBY-5001', 7001, 0, 1, '国贸桥西', 116.470800, 39.900200, '三里屯太古里', 116.473800, 39.932200, 5200, 780, 18.00, 1, '', '', NOW(), NOW()),
  (5002, 'MOCK-NEARBY-5002', 7002, 0, 1, '建国门外大街', 116.471500, 39.901000, '北京站', 116.427300, 39.904300, 6100, 840, 20.00, 1, '', '', NOW(), NOW()),
  (5003, 'MOCK-NEARBY-5003', 7003, 0, 1, '朝阳门', 116.472200, 39.899700, '东直门', 116.442100, 39.947100, 4300, 660, 15.50, 1, '', '', NOW(), NOW()),
  (5004, 'MOCK-NEARBY-5004', 7004, 0, 1, '双井', 116.470300, 39.902100, '十里河', 116.465200, 39.873400, 3800, 600, 14.20, 1, '', '', NOW(), NOW()),
  (5005, 'MOCK-NEARBY-5005', 7005, 0, 1, 'CBD', 116.469900, 39.901400, '望京SOHO', 116.481900, 39.990200, 9500, 1020, 26.80, 1, '', '', NOW(), NOW());

INSERT INTO `order_status_log`
  (`order_id`, `from_status`, `to_status`, `operator_type`, `operator_id`, `remark`, `created_at`)
VALUES
  (5001, 0, 1, 'user', 7001, 'mock create order', NOW()),
  (5002, 0, 1, 'user', 7002, 'mock create order', NOW()),
  (5003, 0, 1, 'user', 7003, 'mock create order', NOW()),
  (5004, 0, 1, 'user', 7004, 'mock create order', NOW()),
  (5005, 0, 1, 'user', 7005, 'mock create order', NOW());

INSERT INTO `dispatch_record`
  (`order_id`, `driver_id`, `dispatch_type`, `status`, `match_score`, `remark`, `reject_reason`, `created_at`, `updated_at`)
VALUES
  (5001, 9001, 1, 1, 98.50, 'mock dispatch seed', '', NOW(), NOW()),
  (5002, 9001, 1, 1, 97.80, 'mock dispatch seed', '', NOW(), NOW()),
  (5003, 9001, 1, 1, 97.20, 'mock dispatch seed', '', NOW(), NOW()),
  (5004, 9001, 1, 1, 96.60, 'mock dispatch seed', '', NOW(), NOW()),
  (5005, 9001, 1, 1, 95.90, 'mock dispatch seed', '', NOW(), NOW());
