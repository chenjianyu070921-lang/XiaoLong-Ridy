-- 推送日志表（可选）
CREATE TABLE IF NOT EXISTS push_log (
    id          BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    user_id     BIGINT       NOT NULL                COMMENT '用户ID',
    push_type   TINYINT      NOT NULL                COMMENT '类型: 1=App推送 2=短信',
    title       VARCHAR(200) DEFAULT ''              COMMENT '标题',
    content     TEXT         NOT NULL                COMMENT '内容',
    target      VARCHAR(100) DEFAULT ''              COMMENT '目标(设备token/手机号)',
    result      TINYINT      DEFAULT 0               COMMENT '结果: 0=失败 1=成功',
    error_msg   VARCHAR(500) DEFAULT ''              COMMENT '失败原因',
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='推送日志表';
