package adminservicelogic

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const couponTimeLayout = "2006-01-02 15:04:05"

// ListCouponsLogic 处理优惠券模板列表查询。
type ListCouponsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListCouponsLogic 创建优惠券列表逻辑。
func NewListCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCouponsLogic {
	return &ListCouponsLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListCoupons 查询优惠券模板列表。
func (l *ListCouponsLogic) ListCoupons(in *adminsvc.CouponListRequest) (*adminsvc.CouponListResponse, error) {
	where, args := buildCouponWhere(in)
	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, `SELECT COUNT(1) FROM coupon `+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	limit := normalizePageSize(in.GetPageSize())
	queryArgs := append(args, limit, offset(in.GetPage(), in.GetPageSize()))
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, `
		SELECT id, name, type, CAST(face_value AS CHAR), CAST(discount AS CHAR),
		       CAST(threshold_amount AS CHAR), total_count, received_count, per_user_limit,
		       valid_start_at, valid_end_at, status, created_at, updated_at
		FROM coupon `+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*adminsvc.Coupon, 0)
	for rows.Next() {
		item, err := scanCouponRow(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return &adminsvc.CouponListResponse{List: list, Total: total, Page: normalizePage(in.GetPage()), PageSize: limit}, rows.Err()
}

// CreateCouponLogic 处理优惠券模板新增。
type CreateCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateCouponLogic 创建优惠券新增逻辑。
func NewCreateCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCouponLogic {
	return &CreateCouponLogic{ctx: ctx, svcCtx: svcCtx}
}

// CreateCoupon 新增优惠券模板。
func (l *CreateCouponLogic) CreateCoupon(in *adminsvc.CouponRequest) (*adminsvc.CommonResponse, error) {
	if err := validateCouponRequest(in); err != nil {
		return nil, err
	}
	tx, err := l.svcCtx.MySQL.BeginTx(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(l.ctx, `
		INSERT INTO coupon (
			name, type, face_value, discount, threshold_amount, total_count,
			received_count, per_user_limit, valid_start_at, valid_end_at, status
		) VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?, ?, ?)
	`, in.GetName(), in.GetType(), in.GetFaceValue(), in.GetDiscount(), in.GetThresholdAmount(),
		in.GetTotalCount(), in.GetPerUserLimit(), in.GetValidStartAt(), in.GetValidEndAt(), in.GetStatus())
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	if err := createOperationLogTx(l.ctx, tx, in.GetAdminId(), "coupon", "create", "coupon", id, fmt.Sprintf("创建优惠券模板：%s", in.GetName()), in.GetIp()); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// UpdateCouponLogic 处理优惠券模板编辑。
type UpdateCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewUpdateCouponLogic 创建优惠券编辑逻辑。
func NewUpdateCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCouponLogic {
	return &UpdateCouponLogic{ctx: ctx, svcCtx: svcCtx}
}

// UpdateCoupon 更新优惠券模板。
func (l *UpdateCouponLogic) UpdateCoupon(in *adminsvc.CouponRequest) (*adminsvc.CommonResponse, error) {
	if in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "优惠券ID不能为空")
	}
	if err := validateCouponRequest(in); err != nil {
		return nil, err
	}
	tx, err := l.svcCtx.MySQL.BeginTx(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(l.ctx, `
		UPDATE coupon
		SET name = ?, type = ?, face_value = ?, discount = ?, threshold_amount = ?,
		    total_count = ?, per_user_limit = ?, valid_start_at = ?, valid_end_at = ?, status = ?
		WHERE id = ?
	`, in.GetName(), in.GetType(), in.GetFaceValue(), in.GetDiscount(), in.GetThresholdAmount(),
		in.GetTotalCount(), in.GetPerUserLimit(), in.GetValidStartAt(), in.GetValidEndAt(), in.GetStatus(), in.GetId())
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, status.Error(codes.NotFound, "优惠券不存在")
	}
	if err := createOperationLogTx(l.ctx, tx, in.GetAdminId(), "coupon", "update", "coupon", in.GetId(), fmt.Sprintf("编辑优惠券模板：%s", in.GetName()), in.GetIp()); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// DisableCouponLogic 处理优惠券模板下架。
type DisableCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewDisableCouponLogic 创建优惠券下架逻辑。
func NewDisableCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DisableCouponLogic {
	return &DisableCouponLogic{ctx: ctx, svcCtx: svcCtx}
}

// DisableCoupon 将优惠券模板状态置为停用。
func (l *DisableCouponLogic) DisableCoupon(in *adminsvc.CouponRequest) (*adminsvc.CommonResponse, error) {
	if in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "优惠券ID不能为空")
	}
	tx, err := l.svcCtx.MySQL.BeginTx(l.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(l.ctx, `
		UPDATE coupon
		SET status = 3, updated_at = ?
		WHERE id = ? AND status <> 3
	`, time.Now(), in.GetId())
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, status.Error(codes.NotFound, "优惠券不存在或已停用")
	}
	if err := createOperationLogTx(l.ctx, tx, in.GetAdminId(), "coupon", "disable", "coupon", in.GetId(), fmt.Sprintf("下架优惠券模板：%s", in.GetName()), in.GetIp()); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// validateCouponRequest 校验优惠券模板参数。
func validateCouponRequest(in *adminsvc.CouponRequest) error {
	if strings.TrimSpace(in.GetName()) == "" {
		return status.Error(codes.InvalidArgument, "优惠券名称不能为空")
	}
	if in.GetType() < 1 || in.GetType() > 3 {
		return status.Error(codes.InvalidArgument, "优惠券类型不合法")
	}
	if in.GetStatus() < 1 || in.GetStatus() > 3 {
		return status.Error(codes.InvalidArgument, "优惠券状态不合法")
	}
	if in.GetPerUserLimit() <= 0 || in.GetTotalCount() < 0 {
		return status.Error(codes.InvalidArgument, "优惠券数量参数不合法")
	}
	if _, err := strconv.ParseFloat(strings.TrimSpace(in.GetFaceValue()), 64); err != nil {
		return status.Error(codes.InvalidArgument, "面值格式不合法")
	}
	if _, err := strconv.ParseFloat(strings.TrimSpace(in.GetDiscount()), 64); err != nil {
		return status.Error(codes.InvalidArgument, "折扣格式不合法")
	}
	if _, err := strconv.ParseFloat(strings.TrimSpace(in.GetThresholdAmount()), 64); err != nil {
		return status.Error(codes.InvalidArgument, "门槛金额格式不合法")
	}
	start, err := time.ParseInLocation(couponTimeLayout, in.GetValidStartAt(), time.Local)
	if err != nil {
		return status.Error(codes.InvalidArgument, "有效期开始时间格式不合法")
	}
	end, err := time.ParseInLocation(couponTimeLayout, in.GetValidEndAt(), time.Local)
	if err != nil {
		return status.Error(codes.InvalidArgument, "有效期结束时间格式不合法")
	}
	if !end.After(start) {
		return status.Error(codes.InvalidArgument, "有效期结束时间必须晚于开始时间")
	}
	return nil
}

// buildCouponWhere 组装优惠券查询条件。
func buildCouponWhere(in *adminsvc.CouponListRequest) (string, []any) {
	parts := []string{"1=1"}
	args := make([]any, 0)
	if in.GetKeyword() != "" {
		parts = append(parts, "name LIKE ?")
		args = append(args, "%"+in.GetKeyword()+"%")
	}
	if in.GetType() > 0 {
		parts = append(parts, "type = ?")
		args = append(args, in.GetType())
	}
	if in.GetStatus() > 0 {
		parts = append(parts, "status = ?")
		args = append(args, in.GetStatus())
	}
	if in.GetStartTime() != "" {
		parts = append(parts, "created_at >= ?")
		args = append(args, in.GetStartTime())
	}
	if in.GetEndTime() != "" {
		parts = append(parts, "created_at <= ?")
		args = append(args, in.GetEndTime())
	}
	return "WHERE " + strings.Join(parts, " AND "), args
}

// scanCouponRow 将查询结果转成 protobuf 优惠券对象。
func scanCouponRow(rows *sql.Rows) (*adminsvc.Coupon, error) {
	var item adminsvc.Coupon
	var validStartAt, validEndAt, createdAt, updatedAt sql.NullTime
	if err := rows.Scan(&item.Id, &item.Name, &item.Type, &item.FaceValue, &item.Discount, &item.ThresholdAmount,
		&item.TotalCount, &item.ReceivedCount, &item.PerUserLimit, &validStartAt, &validEndAt, &item.Status,
		&createdAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("scan coupon row: %w", err)
	}
	item.ValidStartAt = formatNullTime(validStartAt)
	item.ValidEndAt = formatNullTime(validEndAt)
	item.CreatedAt = formatNullTime(createdAt)
	item.UpdatedAt = formatNullTime(updatedAt)
	return &item, nil
}
