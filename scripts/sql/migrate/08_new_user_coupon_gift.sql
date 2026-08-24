-- 新用户注册礼包优惠券模板种子数据。
-- 领取接口固定使用 9001-9004，重复执行不会重复插入。
INSERT INTO coupon (id, name, type, face_value, discount, threshold_amount, total_count, per_user_limit, valid_start_at, valid_end_at, status)
VALUES
  (9001, '新人首单立减20元', 3, 20.00, 1.00, 25.00, 0, 1, NOW(), DATE_ADD(NOW(), INTERVAL 90 DAY), 2),
  (9002, '新人第二单立减8元', 3, 8.00, 1.00, 20.00, 0, 1, NOW(), DATE_ADD(NOW(), INTERVAL 90 DAY), 2),
  (9003, '新人第三单立减5元', 3, 5.00, 1.00, 20.00, 0, 1, NOW(), DATE_ADD(NOW(), INTERVAL 90 DAY), 2),
  (9004, '夜间出行立减5元', 3, 5.00, 1.00, 15.00, 0, 1, NOW(), DATE_ADD(NOW(), INTERVAL 90 DAY), 2)
ON DUPLICATE KEY UPDATE
  name = VALUES(name), type = VALUES(type), face_value = VALUES(face_value), threshold_amount = VALUES(threshold_amount), status = VALUES(status);


