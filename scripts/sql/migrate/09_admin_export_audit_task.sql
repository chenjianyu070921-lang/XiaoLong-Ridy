-- 管理后台导出任务与审计补偿表。
-- 注意：本脚本只作为数据库迁移定义提交，需由数据库负责人按发布流程执行，代码侧不会自动执行迁移。

CREATE TABLE IF NOT EXISTS admin_export_task (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '自增主键',
    task_no VARCHAR(64) NOT NULL COMMENT '导出任务编号',
    export_type VARCHAR(64) NOT NULL COMMENT '导出类型：users/orders/drivers/operation_logs/statistics',
    filters JSON NULL COMMENT '导出筛选条件 JSON',
    status VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '任务状态：pending/running/success/failed/canceled',
    file_path VARCHAR(512) NOT NULL DEFAULT '' COMMENT '服务本地文件路径',
    file_url VARCHAR(512) NOT NULL DEFAULT '' COMMENT '下载 URL 或对象存储地址',
    failure_reason VARCHAR(512) NOT NULL DEFAULT '' COMMENT '失败原因',
    admin_id BIGINT NOT NULL COMMENT '创建任务的管理员 ID',
    ip VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建任务的客户端 IP',
    retry_count INT NOT NULL DEFAULT 0 COMMENT '重试次数',
    canceled_at DATETIME NULL COMMENT '取消时间',
    expires_at DATETIME NULL COMMENT '文件过期时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_admin_export_task_no (task_no),
    KEY idx_admin_export_task_type_status (export_type, status),
    KEY idx_admin_export_task_admin_created (admin_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理后台导出任务表';

CREATE TABLE IF NOT EXISTS admin_audit_outbox (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '自增主键',
    event_no VARCHAR(64) NOT NULL COMMENT '审计补偿事件编号',
    module VARCHAR(64) NOT NULL COMMENT '业务模块',
    action VARCHAR(64) NOT NULL COMMENT '业务动作',
    target_type VARCHAR(64) NOT NULL COMMENT '审计对象类型',
    target_id BIGINT NOT NULL DEFAULT 0 COMMENT '审计对象 ID',
    admin_id BIGINT NOT NULL COMMENT '操作管理员 ID',
    detail VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '补偿审计详情',
    ip VARCHAR(64) NOT NULL DEFAULT '' COMMENT '客户端 IP',
    status VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '补偿状态：pending/running/success/failed',
    retry_count INT NOT NULL DEFAULT 0 COMMENT '补偿重试次数',
    failure_reason VARCHAR(512) NOT NULL DEFAULT '' COMMENT '最近一次失败原因',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_admin_audit_outbox_event_no (event_no),
    KEY idx_admin_audit_outbox_status (status, created_at),
    KEY idx_admin_audit_outbox_target (target_type, target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理后台审计补偿任务表';
