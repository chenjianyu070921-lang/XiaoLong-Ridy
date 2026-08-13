-- 站内信表
CREATE TABLE IF NOT EXISTS notices (
    id         BIGINT        NOT NULL AUTO_INCREMENT  COMMENT '主键ID',
    user_id    BIGINT        NOT NULL                 COMMENT '用户ID',
    title      VARCHAR(200)  NOT NULL                 COMMENT '标题',
    content    TEXT          NOT NULL                 COMMENT '内容',
    biz_type   TINYINT       NOT NULL DEFAULT 1       COMMENT '业务类型: 1=订单通知 2=系统通知 3=活动通知',
    is_read    TINYINT       NOT NULL DEFAULT 0       COMMENT '是否已读: 0=未读 1=已读',
    created_at DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (id),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='站内信表';
