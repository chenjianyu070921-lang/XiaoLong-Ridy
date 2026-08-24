-- 修复历史新人券模板状态，并为历史账号补齐缺失的用户券。
-- 脚本可重复执行：每个用户每个新人券模板最多保留一条 user_coupon 记录。
UPDATE coupon
SET status = 2,
    valid_start_at = CASE WHEN valid_start_at IS NULL THEN NOW() ELSE valid_start_at END,
    valid_end_at = CASE WHEN valid_end_at IS NULL OR valid_end_at <= NOW() THEN DATE_ADD(NOW(), INTERVAL 90 DAY) ELSE valid_end_at END
WHERE id IN (9001, 9002, 9003, 9004);

-- 将新人券模板发放到历史用户的 user_coupon 表，避免老账号进入确认行程页时显示“无可用优惠券”。
INSERT INTO user_coupon (user_id, coupon_id, order_id, status, received_at, expire_at)
SELECT u.id, c.id, 0, 1, NOW(), c.valid_end_at
FROM user u
JOIN coupon c ON c.id IN (9001, 9002, 9003, 9004) AND c.status = 2
WHERE NOT EXISTS (
    SELECT 1
    FROM user_coupon uc
    WHERE uc.user_id = u.id AND uc.coupon_id = c.id
);

-- 重新按实际插入数量刷新模板领取数，保持后台统计一致。
UPDATE coupon c
SET received_count = (
    SELECT COUNT(*) FROM user_coupon uc WHERE uc.coupon_id = c.id
)
WHERE c.id IN (9001, 9002, 9003, 9004);


