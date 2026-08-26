// fix_table_comment 修复线上数据库所有表/列的中文注释（乱码恢复）。
// 用法: go run scripts/fix_table_comment/main.go [--dsn "user:pass@tcp(host:port)/db?charset=utf8mb4"] [--dry-run]
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// tableComments: 表名 -> 表注释
var tableComments = map[string]string{
	"user":                          "乘客用户表",
	"user_address":                  "乘客常用地址表",
	"driver":                        "司机账号表",
	"driver_vehicle":                "司机车辆表",
	"driver_certification":          "司机资质认证表",
	"driver_score":                  "司机服务分表",
	"driver_withdraw":               "司机提现表",
	"admin_user":                    "后台管理员表",
	"admin_operation_log":           "后台操作日志表",
	"coupon":                        "优惠券模板表",
	"user_coupon":                   "用户优惠券表",
	"promotion_activity":            "营销活动表",
	"blacklist":                     "风控黑名单表",
	"ride_order":                    "打车订单主表",
	"order_status_log":              "订单状态日志表",
	"dispatch_record":               "派单记录表",
	"price_rule":                    "计价规则表",
	"order_price":                   "订单价格明细表",
	"payment":                       "支付单表",
	"settlement":                    "订单结算表",
	"driver_location":               "司机实时位置表",
	"ride_track_point":              "行程轨迹点表",
	"geofence":                      "电子围栏表",
	"push_message":                  "消息推送记录表",
	"admin_complaint_work_order":    "后台投诉申诉工单主表",
	"admin_work_order_flow":         "后台工单流转记录表",
	"admin_work_order_evidence":     "后台工单证据表",
	"admin_coupon_publish_record":   "后台优惠券发布记录表",
	"admin_coupon_issue_task":       "后台发券任务表",
	"risk_blacklist_hit_record":     "风控黑名单命中记录表",
}

// columnComments: 表名 -> 列名 -> 列注释
var columnComments = map[string]map[string]string{
	"user": {
		"id":              "用户ID",
		"phone":           "手机号（登录账号）",
		"password_hash":   "密码哈希，验证码/第三方登录可为空",
		"nickname":        "昵称",
		"avatar_url":      "头像地址",
		"gender":          "性别：0未知 1男 2女",
		"real_name":       "实名认证姓名",
		"id_card_no":      "实名认证身份证号",
		"register_source": "注册来源：phone/wechat/alipay",
		"status":          "账号状态：1正常 2冻结",
		"created_at":      "创建时间",
		"updated_at":      "更新时间",
		"deleted_at":      "软删除时间",
	},
	"user_address": {
		"id":            "地址ID",
		"user_id":       "用户ID",
		"contact_name":  "联系人姓名",
		"contact_phone": "联系人电话",
		"tag":           "地址标签：home/work/other",
		"address":       "详细地址",
		"longitude":     "经度",
		"latitude":      "纬度",
		"is_default":    "是否默认地址：0否 1是",
		"created_at":    "创建时间",
		"updated_at":    "更新时间",
		"deleted_at":    "软删除时间",
	},
	"driver": {
		"id":                "司机ID",
		"phone":             "手机号（登录账号）",
		"password_hash":     "密码哈希",
		"real_name":         "真实姓名",
		"id_card_no":        "身份证号",
		"driver_license_no": "驾驶证号",
		"avatar_url":        "头像地址",
		"status":            "账号状态：1待审核 2正常 3冻结 4注销",
		"created_at":        "创建时间",
		"updated_at":        "更新时间",
		"deleted_at":        "软删除时间",
	},
	"driver_vehicle": {
		"id":                  "车辆ID",
		"driver_id":           "司机ID",
		"plate_no":            "车牌号",
		"brand":               "品牌",
		"model":               "车型",
		"color":               "车身颜色",
		"vehicle_type":        "车辆类型：1特惠快车 2快车 3拼车",
		"registration_date":   "注册日期",
		"insurance_no":        "保险单号",
		"insurance_expire_at": "保险到期日",
		"status":              "状态：1待审核 2正常 3禁用",
		"created_at":          "创建时间",
		"updated_at":          "更新时间",
	},
	"driver_certification": {
		"id":                  "认证ID",
		"driver_id":           "司机ID",
		"vehicle_id":          "车辆ID",
		"id_card_front_url":   "身份证人像面",
		"id_card_back_url":    "身份证国徽面",
		"driver_license_url":  "驾驶证照片",
		"vehicle_license_url": "行驶证照片",
		"audit_status":        "审核状态：1待审核 2通过 3驳回",
		"audit_remark":        "驳回原因/审核备注",
		"audited_by":          "审核人（后台管理员ID）",
		"audited_at":          "审核时间",
		"created_at":          "创建时间",
		"updated_at":          "更新时间",
	},
	"driver_score": {
		"id":                    "主键ID",
		"driver_id":             "司机ID",
		"score":                 "服务分",
		"level":                 "司机等级：1-5",
		"month_orders":          "当月完单数",
		"month_cancel_rate":     "当月取消率（%）",
		"month_complaint_count": "当月投诉数",
		"updated_at":            "更新时间",
	},
	"driver_withdraw": {
		"id":           "提现ID",
		"driver_id":    "司机ID",
		"withdraw_no":  "提现单号",
		"amount":       "提现金额",
		"payee_name":   "收款人姓名",
		"pay_account":  "收款账号",
		"status":       "状态：1申请中 2打款成功 3打款失败",
		"remark":       "失败原因/备注",
		"applied_at":   "申请时间",
		"paid_at":      "打款时间",
		"created_at":   "创建时间",
	},
	"admin_user": {
		"id":            "管理员ID",
		"username":      "登录用户名",
		"password_hash": "密码哈希",
		"real_name":     "真实姓名",
		"role":          "角色：1超级管理员 2运营 3客服",
		"status":        "状态：1正常 2禁用",
		"last_login_at": "最后登录时间",
		"created_at":    "创建时间",
		"updated_at":    "更新时间",
		"deleted_at":    "软删除时间",
	},
	"admin_operation_log": {
		"id":          "日志ID",
		"admin_id":    "管理员ID",
		"module":      "模块：driver/orderclient/coupon/risk",
		"action":      "动作：audit/ban/change",
		"target_type": "操作对象类型",
		"target_id":   "操作对象ID",
		"detail":      "操作详情",
		"ip":          "操作IP",
		"created_at":  "操作时间",
	},
	"coupon": {
		"id":               "优惠券模板ID",
		"name":             "优惠券名称",
		"type":             "类型：1满减 2折扣 3立减",
		"face_value":       "面值",
		"discount":         "折扣率",
		"threshold_amount": "使用门槛金额",
		"total_count":      "发放总量，0表示不限",
		"received_count":   "已领取数量",
		"per_user_limit":   "每人限领数量",
		"valid_start_at":   "生效开始时间",
		"valid_end_at":     "生效结束时间",
		"status":           "状态：1启用 2停用",
		"created_at":       "创建时间",
		"updated_at":       "更新时间",
	},
	"user_coupon": {
		"id":          "用户优惠券ID",
		"user_id":     "用户ID",
		"coupon_id":   "优惠券模板ID",
		"order_id":    "使用订单ID，0表示未使用",
		"status":      "状态：1未使用 2已使用 3已过期",
		"received_at": "领取时间",
		"used_at":     "使用时间",
		"expire_at":   "过期时间",
	},
	"promotion_activity": {
		"id":         "活动ID",
		"name":       "活动名称",
		"type":       "类型：1新人 2邀请 3限时",
		"config":     "活动配置JSON",
		"start_at":   "开始时间",
		"end_at":     "结束时间",
		"status":     "状态：1未开始 2进行中 3已结束",
		"created_by": "创建人管理员ID",
		"created_at": "创建时间",
		"updated_at": "更新时间",
	},
	"blacklist": {
		"id":          "黑名单ID",
		"target_type": "目标类型：user/driver/device",
		"target_id":   "目标ID",
		"reason":      "拉黑原因",
		"operator_id": "操作人管理员ID",
		"status":      "状态：1生效中 2已解除",
		"created_at":  "创建时间",
		"updated_at":  "更新时间",
	},
	"ride_order": {
		"id":                  "订单ID",
		"order_no":            "订单号（对外展示）",
		"user_id":             "乘客用户ID",
		"driver_id":           "司机ID，未接单为0",
		"car_type":            "车型：1特惠快车 2快车 3拼车",
		"from_address":        "起点地址",
		"from_longitude":      "起点经度",
		"from_latitude":       "起点纬度",
		"to_address":          "终点地址",
		"to_longitude":        "终点经度",
		"to_latitude":         "终点纬度",
		"estimated_distance_m": "预估距离（米）",
		"estimated_duration_s": "预估时长（秒）",
		"estimated_price":      "预估价格",
		"status":               "状态：1待接单 2已接单 3行程中 4待支付 5已完成 6已取消",
		"cancel_reason":        "取消原因",
		"cancel_by":            "取消方：user/driver/system/admin",
		"created_at":           "下单时间",
		"updated_at":           "更新时间",
		"deleted_at":           "软删除时间",
	},
	"order_status_log": {
		"id":            "日志ID",
		"order_id":      "订单ID",
		"from_status":   "原状态",
		"to_status":     "新状态",
		"operator_type": "操作方：user/driver/system/admin",
		"operator_id":   "操作方ID",
		"remark":        "备注/原因",
		"created_at":    "记录时间",
	},
	"dispatch_record": {
		"id":           "派单记录ID",
		"order_id":     "订单ID",
		"driver_id":    "候选司机ID",
		"dispatch_type": "派单方式：1自动派单 2抢单 3改派",
		"status":       "状态：1派单中 2已接受 3已拒绝 4超时 5已取消",
		"match_score":  "匹配分（距离/评分/顺路度加权）",
		"remark":       "备注",
		"created_at":   "派单时间",
		"updated_at":   "更新时间",
	},
	"price_rule": {
		"id":                   "规则ID",
		"name":                 "规则名称",
		"city_code":            "城市编码，空表示全局",
		"car_type":             "车型：1特惠快车 2快车 3拼车",
		"base_price":           "起步价",
		"base_distance_km":     "起步包含里程（公里）",
		"per_km_price":         "每公里价格",
		"per_minute_price":     "每分钟时长费",
		"night_start_time":     "夜间费开始时间",
		"night_end_time":       "夜间费结束时间",
		"night_surcharge":      "夜间附加费（元/单）",
		"dynamic_max_factor":   "动态调价最大倍数",
		"status":               "状态：1启用 2停用",
		"effective_at":         "生效时间",
		"expire_at":            "失效时间，NULL表示长期有效",
		"created_at":           "创建时间",
		"updated_at":           "更新时间",
	},
	"order_price": {
		"id":                "价格记录ID",
		"order_id":          "订单ID",
		"price_rule_id":     "使用的计价规则ID",
		"estimated_price":   "预估价格",
		"actual_price":      "实际总价",
		"base_fee":          "起步价费用",
		"distance_fee":      "里程费用",
		"time_fee":          "时长费用",
		"night_fee":         "夜间附加费",
		"dynamic_fee":       "动态调价费用",
		"discount_amount":   "优惠券抵扣金额",
		"platform_subsidy":  "平台补贴金额",
		"payable_amount":    "乘客实付金额",
		"status":            "状态：1预估 2已确认",
		"created_at":        "创建时间",
		"updated_at":        "更新时间",
	},
	"payment": {
		"id":             "支付单ID",
		"payment_no":     "平台支付单号",
		"order_id":       "订单ID",
		"user_id":        "支付用户ID",
		"amount":         "支付金额",
		"channel":        "支付渠道：wechat/alipay/balance",
		"status":         "状态：1待支付 2支付成功 3支付失败 4已退款",
		"transaction_id": "第三方支付流水号",
		"refund_amount":  "已退款金额",
		"paid_at":        "支付成功时间",
		"created_at":     "创建时间",
		"updated_at":     "更新时间",
	},
	"settlement": {
		"id":                       "结算ID",
		"settlement_no":            "结算单号",
		"order_id":                 "订单ID",
		"driver_id":                "司机ID",
		"total_amount":             "订单实际总金额",
		"platform_commission_rate": "平台抽成比例（%）",
		"platform_commission":      "平台抽成金额",
		"driver_income":            "司机收入",
		"status":                   "状态：1待结算 2已结算",
		"settled_at":               "结算时间",
		"created_at":               "创建时间",
	},
	"driver_location": {
		"id":            "主键ID",
		"driver_id":     "司机ID",
		"longitude":     "经度",
		"latitude":      "纬度",
		"heading":       "行驶方向（0-359度）",
		"speed_kmh":     "当前速度（km/h）",
		"online_status": "听单状态：0离线 1在线 2行程中",
		"report_time":   "位置上报时间",
		"created_at":    "写入时间",
	},
	"ride_track_point": {
		"id":          "轨迹点ID",
		"order_id":    "订单ID",
		"driver_id":   "司机ID",
		"longitude":   "经度",
		"latitude":    "纬度",
		"speed_kmh":   "速度（km/h）",
		"direction":   "方向（0-359度）",
		"recorded_at": "轨迹时间",
		"created_at":  "写入时间",
	},
	"geofence": {
		"id":               "围栏ID",
		"name":             "围栏名称",
		"area_type":        "类型：1运营区 2禁运区 3热区",
		"center_longitude": "中心经度",
		"center_latitude":  "中心纬度",
		"radius_m":         "半径（米）",
		"status":           "状态：1启用 2停用",
		"remark":           "备注",
		"created_at":       "创建时间",
		"updated_at":       "更新时间",
	},
	"push_message": {
		"id":          "消息ID",
		"biz_type":    "业务类型：orderclient/activity/system/verify_code",
		"target_type": "接收方类型：user/driver",
		"target_id":   "接收方ID",
		"order_id":    "关联订单ID，无则0",
		"title":       "消息标题",
		"content":     "消息内容",
		"channel":     "渠道：app/sms/ws",
		"status":      "状态：1待发送 2已发送 3发送失败",
		"send_at":     "发送时间",
		"created_at":  "创建时间",
	},
	"admin_complaint_work_order": {
		"id":                  "工单ID",
		"work_order_no":       "工单编号",
		"work_order_type":     "工单类型：1用户投诉 2订单申诉 3司机处罚申诉",
		"source_type":         "来源类型：user/driver/orderclient/system",
		"source_id":           "来源业务ID",
		"order_id":            "关联订单ID",
		"user_id":             "关联乘客ID",
		"driver_id":           "关联司机ID",
		"title":               "工单标题",
		"content":             "投诉或申诉内容",
		"priority":            "优先级：1低 2中 3高 4紧急",
		"status":              "状态：1待处理 2跟进中 3仲裁完成 4已结案 5已关闭",
		"assignee_id":         "当前处理人管理员ID",
		"arbitration_result":  "仲裁结果",
		"remark":              "后台备注",
		"version":             "乐观锁版本号，用于避免多人同时修改覆盖",
		"created_by":          "创建人管理员ID，系统创建时为0",
		"created_at":          "创建时间",
		"updated_at":          "更新时间",
		"closed_at":           "结案或关闭时间",
	},
	"admin_work_order_flow": {
		"id":            "流转记录ID",
		"work_order_id": "工单ID",
		"from_status":   "变更前状态",
		"to_status":     "变更后状态",
		"action":        "动作：assign/follow/arbitrate/close/reopen",
		"operator_id":   "操作管理员ID",
		"content":       "处理内容或备注",
		"created_at":    "创建时间",
	},
	"admin_work_order_evidence": {
		"id":            "证据ID",
		"work_order_id": "工单ID",
		"evidence_type": "证据类型：track/audio/chat/payment/image/text",
		"evidence_url":  "证据文件或资源地址",
		"content":       "文本证据内容",
		"uploaded_by":   "上传人管理员ID",
		"created_at":    "创建时间",
	},
	"admin_coupon_publish_record": {
		"id":              "发布记录ID",
		"coupon_id":       "优惠券模板ID",
		"publish_version": "发布版本号",
		"publish_scope":   "发布范围：draft/gray/full/rollback",
		"target_config":   "目标人群、城市、灰度比例等JSON配置",
		"status":          "状态：1待发布 2发布成功 3发布失败 4已回滚",
		"failure_reason":  "失败原因",
		"operator_id":     "操作管理员ID",
		"created_at":      "创建时间",
		"updated_at":      "更新时间",
	},
	"admin_coupon_issue_task": {
		"id":             "发券任务ID",
		"task_no":        "任务编号",
		"coupon_id":      "优惠券模板ID",
		"target_type":    "目标类型：user/batch/crowd",
		"target_config":  "目标用户或人群配置JSON",
		"total_count":    "计划发放数量",
		"success_count":  "成功数量",
		"fail_count":     "失败数量",
		"status":         "状态：1待执行 2执行中 3成功 4部分失败 5失败",
		"failure_reason": "失败原因",
		"operator_id":    "操作管理员ID",
		"created_at":     "创建时间",
		"updated_at":     "更新时间",
	},
	"risk_blacklist_hit_record": {
		"id":           "命中记录ID",
		"blacklist_id": "关联 blacklist ID",
		"target_type":  "目标类型：user/driver/device/phone",
		"target_id":    "目标ID",
		"scene":        "命中场景：login/orderclient/dispatch/pay/refund",
		"risk_level":   "风险等级：1低 2中 3高",
		"hit_reason":   "命中原因",
		"request_id":   "请求链路ID",
		"created_at":   "创建时间",
	},
}

type colInfo struct {
	Table     string
	Column    string
	ColType   string
	Nullable  string
	Default   sql.NullString
	Extra     string
	Charset   sql.NullString
	Collation sql.NullString
}

func isNumericType(ct string) bool {
	ct = strings.ToLower(ct)
	for _, p := range []string{"tinyint", "smallint", "mediumint", "bigint", "int", "decimal", "numeric", "float", "double", "bit", "year"} {
		if strings.HasPrefix(ct, p) {
			return true
		}
	}
	return false
}

func escapeSQL(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// buildModify 构造 ALTER TABLE ... MODIFY COLUMN ... COMMENT '...'，完整保留列属性
func buildModify(c colInfo, comment string) string {
	s := "ALTER TABLE `" + c.Table + "` MODIFY COLUMN `" + c.Column + "` " + c.ColType
	if c.Charset.Valid && c.Charset.String != "" {
		s += " CHARACTER SET " + c.Charset.String
		if c.Collation.Valid && c.Collation.String != "" {
			s += " COLLATE " + c.Collation.String
		}
	}
	if c.Nullable == "NO" {
		s += " NOT NULL"
	} else {
		s += " NULL"
	}
	extra := strings.ToLower(c.Extra)
	// 默认值
	if c.Default.Valid {
		if strings.EqualFold(c.Default.String, "CURRENT_TIMESTAMP") {
			s += " DEFAULT CURRENT_TIMESTAMP"
		} else if isNumericType(c.ColType) {
			s += " DEFAULT " + c.Default.String
		} else {
			s += " DEFAULT '" + escapeSQL(c.Default.String) + "'"
		}
	}
	if strings.Contains(extra, "auto_increment") {
		s += " AUTO_INCREMENT"
	}
	if strings.Contains(extra, "on update current_timestamp") {
		s += " ON UPDATE CURRENT_TIMESTAMP"
	}
	s += " COMMENT '" + escapeSQL(comment) + "'"
	return s
}

func main() {
	dsn := flag.String("dsn", "root:4ay1nkal3u8ed77y@tcp(115.191.16.159:3306)/xiaolong_ridy?charset=utf8mb4&parseTime=True&loc=Local", "MySQL DSN")
	dry := flag.Bool("dry-run", false, "仅打印将要执行的 SQL，不实际执行")
	flag.Parse()

	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "连接失败:", err)
		os.Exit(1)
	}
	defer db.Close()
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(5)

	// 收集目标表
	tableSet := map[string]bool{}
	for t := range tableComments {
		tableSet[t] = true
	}

	// 读取线上列定义
	rows, err := db.Query(`
		SELECT TABLE_NAME, COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT, EXTRA,
		       CHARACTER_SET_NAME, COLLATION_NAME
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		ORDER BY TABLE_NAME, ORDINAL_POSITION`)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查询列定义失败:", err)
		os.Exit(1)
	}
	var cols []colInfo
	for rows.Next() {
		var c colInfo
		var nullable string
		var def, cs, coll sql.NullString
		var extra string
		if err := rows.Scan(&c.Table, &c.Column, &c.ColType, &nullable, &def, &extra, &cs, &coll); err != nil {
			fmt.Fprintln(os.Stderr, "扫描失败:", err)
			os.Exit(1)
		}
		c.Nullable, c.Default, c.Extra, c.Charset, c.Collation = nullable, def, extra, cs, coll
		cols = append(cols, c)
	}
	rows.Close()

	// 统计
	var stmts []string
	tableOrder := make([]string, 0, len(tableComments))
	for t := range tableComments {
		tableOrder = append(tableOrder, t)
	}
	sort.Strings(tableOrder)

	foundTables := map[string]bool{}
	missingCols := 0
	missingTables := 0

	for _, t := range tableOrder {
		colMap := columnComments[t]
		if !tableSet[t] {
			continue
		}
		tFound := false
		for _, c := range cols {
			if c.Table != t {
				continue
			}
			tFound = true
			comment, ok := colMap[c.Column]
			if !ok {
				// 仓库无此列定义，跳过不修改
				continue
			}
			stmts = append(stmts, buildModify(c, comment))
		}
		if tFound {
			foundTables[t] = true
			stmts = append(stmts, "ALTER TABLE `"+t+"` COMMENT = '"+escapeSQL(tableComments[t])+"'")
		}
	}

	for _, t := range tableOrder {
		if !foundTables[t] {
			fmt.Printf("[SKIP] 表 %s 在线上库不存在\n", t)
			missingTables++
		} else {
			colMap := columnComments[t]
			exist := map[string]bool{}
			for _, c := range cols {
				if c.Table == t {
					exist[c.Column] = true
				}
			}
			for cname := range colMap {
				if !exist[cname] {
					fmt.Printf("[WARN] 表 %s 缺少仓库定义列 %s（线上无此列，跳过）\n", t, cname)
					missingCols++
				}
			}
		}
	}

	fmt.Printf("待执行语句: %d 条（表注释 %d 张 + 列注释 %d 条），跳过缺失表 %d，线上缺少列 %d\n",
		len(stmts), len(foundTables), len(stmts)-len(foundTables), missingTables, missingCols)

	if *dry {
		for _, s := range stmts {
			fmt.Println(s)
		}
		return
	}

	// 执行
	success, failed := 0, 0
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			failed++
			fmt.Printf("[FAIL] %s\n   原因: %v\n", s, err)
		} else {
			success++
		}
	}
	fmt.Printf("执行完成: 成功 %d, 失败 %d\n", success, failed)
	if failed > 0 {
		os.Exit(2)
	}
}
