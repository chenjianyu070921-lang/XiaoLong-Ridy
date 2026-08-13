-- 司机实时位置表（模块四使用）
CREATE TABLE IF NOT EXISTS driver_location (
    id          BIGINT   NOT NULL AUTO_INCREMENT COMMENT '主键',
    driver_id   BIGINT   NOT NULL                COMMENT '司机ID',
    lat         DOUBLE   NOT NULL                COMMENT '纬度',
    lng         DOUBLE   NOT NULL                COMMENT '经度',
    speed       DOUBLE   DEFAULT 0               COMMENT '速度 km/h',
    direction   DOUBLE   DEFAULT 0               COMMENT '方向角',
    order_id    BIGINT   DEFAULT 0               COMMENT '关联订单ID, 0=空闲',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '上报时间',
    PRIMARY KEY (id),
    INDEX idx_driver_id (driver_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='司机实时位置表';
