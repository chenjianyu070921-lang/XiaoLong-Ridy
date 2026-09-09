package adminservicelogic

import (
	"encoding/json"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const adminTimeLayout = "2006-01-02 15:04:05"

// exportFilters 是后台导出任务支持的白名单筛选字段集合。
// 所有字段最终以 SQL 占位参数传入，禁止将客户端 JSON 原样拼接到 SQL。
type exportFilters struct {
	Keyword     string
	Status      int32
	AuditStatus int32
	UserID      int64
	DriverID    int64
	AdminID     int64
	Module      string
	Action      string
	TargetType  string
	TargetID    int64
	StartTime   string
	EndTime     string
}

// parseExportFilters 解析并校验指定导出类型的筛选 JSON。
func parseExportFilters(exportType, raw string) (exportFilters, error) {
	filters := exportFilters{}
	if strings.TrimSpace(raw) == "" {
		return filters, nil
	}
	var values map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return filters, status.Error(codes.InvalidArgument, "导出筛选条件必须是合法JSON")
	}
	allowed := exportFilterFields(exportType)
	for key := range values {
		if !allowed[key] {
			return filters, status.Errorf(codes.InvalidArgument, "导出类型不支持筛选字段%s", key)
		}
	}
	if err := decodeExportFilter(values, "keyword", &filters.Keyword); err != nil {
		return filters, err
	}
	if err := decodeExportFilter(values, "status", &filters.Status); err != nil {
		return filters, err
	}
	if err := decodeExportFilter(values, "audit_status", &filters.AuditStatus); err != nil {
		return filters, err
	}
	if err := decodeExportFilter(values, "user_id", &filters.UserID); err != nil {
		return filters, err
	}
	if err := decodeExportFilter(values, "driver_id", &filters.DriverID); err != nil {
		return filters, err
	}
	if err := decodeExportFilter(values, "admin_id", &filters.AdminID); err != nil {
		return filters, err
	}
	if err := decodeExportFilter(values, "module", &filters.Module); err != nil {
		return filters, err
	}
	if err := decodeExportFilter(values, "action", &filters.Action); err != nil {
		return filters, err
	}
	if err := decodeExportFilter(values, "target_type", &filters.TargetType); err != nil {
		return filters, err
	}
	if err := decodeExportFilter(values, "target_id", &filters.TargetID); err != nil {
		return filters, err
	}
	if err := decodeExportFilter(values, "start_time", &filters.StartTime); err != nil {
		return filters, err
	}
	if err := decodeExportFilter(values, "end_time", &filters.EndTime); err != nil {
		return filters, err
	}
	if err := validateExportTimeRange(filters.StartTime, filters.EndTime); err != nil {
		return filters, err
	}
	return filters, nil
}

// exportFilterFields 返回每种导出类型允许的筛选字段，防止跨资源条件混用。
func exportFilterFields(exportType string) map[string]bool {
	commonTime := map[string]bool{"start_time": true, "end_time": true}
	switch exportType {
	case "users":
		return map[string]bool{"keyword": true, "status": true, "start_time": true, "end_time": true}
	case "drivers":
		return map[string]bool{"keyword": true, "audit_status": true, "start_time": true, "end_time": true}
	case "orders":
		return map[string]bool{"keyword": true, "status": true, "user_id": true, "driver_id": true, "start_time": true, "end_time": true}
	case "operation_logs":
		return map[string]bool{"admin_id": true, "module": true, "action": true, "target_type": true, "target_id": true, "start_time": true, "end_time": true}
	case "statistics":
		return commonTime
	default:
		return map[string]bool{}
	}
}

// decodeExportFilter 仅在字段存在时将 JSON 值解码到目标变量。
func decodeExportFilter(values map[string]json.RawMessage, key string, target any) error {
	raw, ok := values[key]
	if !ok {
		return nil
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return status.Errorf(codes.InvalidArgument, "导出筛选字段%s格式不合法", key)
	}
	return nil
}

// validateExportTimeRange 统一校验导出时间格式和先后关系。
func validateExportTimeRange(startText, endText string) error {
	if startText == "" && endText == "" {
		return nil
	}
	var start, end time.Time
	var err error
	if startText != "" {
		start, err = time.ParseInLocation(adminTimeLayout, startText, time.Local)
		if err != nil {
			return status.Error(codes.InvalidArgument, "导出开始时间格式不合法")
		}
	}
	if endText != "" {
		end, err = time.ParseInLocation(adminTimeLayout, endText, time.Local)
		if err != nil {
			return status.Error(codes.InvalidArgument, "导出结束时间格式不合法")
		}
	}
	if !start.IsZero() && !end.IsZero() && !end.After(start) {
		return status.Error(codes.InvalidArgument, "导出结束时间必须晚于开始时间")
	}
	return nil
}

// exportWhere 使用固定列名构造导出筛选条件，values 均通过占位参数绑定。
func exportWhere(base string, filters exportFilters, fields map[string]string) (string, []any) {
	parts := []string{base}
	args := make([]any, 0)
	appendText := func(value, field string) {
		if value != "" && field != "" {
			parts = append(parts, field+" = ?")
			args = append(args, value)
		}
	}
	appendID := func(value int64, field string) {
		if value > 0 && field != "" {
			parts = append(parts, field+" = ?")
			args = append(args, value)
		}
	}
	appendStatus := func(value int32, field string) {
		if value > 0 && field != "" {
			parts = append(parts, field+" = ?")
			args = append(args, value)
		}
	}
	if field := fields["keyword"]; field != "" && filters.Keyword != "" {
		parts = append(parts, field+" LIKE ?")
		args = append(args, "%"+filters.Keyword+"%")
	}
	appendStatus(filters.Status, fields["status"])
	appendStatus(filters.AuditStatus, fields["audit_status"])
	appendID(filters.UserID, fields["user_id"])
	appendID(filters.DriverID, fields["driver_id"])
	appendID(filters.AdminID, fields["admin_id"])
	appendText(filters.Module, fields["module"])
	appendText(filters.Action, fields["action"])
	appendText(filters.TargetType, fields["target_type"])
	appendID(filters.TargetID, fields["target_id"])
	if filters.StartTime != "" && fields["created_at"] != "" {
		parts = append(parts, fields["created_at"]+" >= ?")
		args = append(args, filters.StartTime)
	}
	if filters.EndTime != "" && fields["created_at"] != "" {
		parts = append(parts, fields["created_at"]+" <= ?")
		args = append(args, filters.EndTime)
	}
	return "WHERE " + strings.Join(parts, " AND "), args
}
