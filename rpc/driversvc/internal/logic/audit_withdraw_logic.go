package logic

import (
	"context"
	"errors"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// 提现状态：与 02_driver_module.sql 中 driver_withdraw.status 枚举保持一致。
const (
	withdrawStatusPaid    int32 = 2 // 打款成功
	withdrawStatusFailed  int32 = 3 // 打款失败
)

// AuditWithdrawLogic 封装管理后台审核提现申请的业务逻辑。
type AuditWithdrawLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewAuditWithdrawLogic 创建管理后台提现审核逻辑处理器。
func NewAuditWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuditWithdrawLogic {
	return &AuditWithdrawLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AuditWithdraw 审核申请中的提现：approve=true 置为打款成功并写打款时间，
// approve=false 置为打款失败且备注必填。只有申请中（状态 1）的记录允许审核，
// 返回 driver_id 供调用方（adminsvc）做司机端通知与审计联动。
func (l *AuditWithdrawLogic) AuditWithdraw(in *proto.AuditWithdrawRequest) (*proto.AuditWithdrawResponse, error) {
	if in == nil || in.GetWithdrawId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "withdraw id is required")
	}
	if in.GetOperatorId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "operator id is required")
	}
	remark := strings.TrimSpace(in.GetRemark())
	if !in.GetApprove() && remark == "" {
		return nil, status.Error(codes.InvalidArgument, "reject remark is required")
	}
	if l.svcCtx == nil || l.svcCtx.DriverWithdrawRepository == nil {
		return nil, status.Error(codes.FailedPrecondition, "driver withdraw repository not ready")
	}

	record, err := l.svcCtx.DriverWithdrawRepository.GetByID(l.ctx, uint64(in.GetWithdrawId()))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "withdraw record not found")
		}
		return nil, err
	}
	if record == nil || record.Status != withdrawStatusPending {
		return nil, status.Error(codes.FailedPrecondition, "only pending withdraw can be audited")
	}

	target := withdrawStatusPaid
	var paidAt *time.Time
	if in.GetApprove() {
		now := time.Now()
		paidAt = &now
	} else {
		target = withdrawStatusFailed
	}
	if err := l.svcCtx.DriverWithdrawRepository.Audit(l.ctx, uint64(in.GetWithdrawId()), target, remark, paidAt); err != nil {
		return nil, err
	}
	return &proto.AuditWithdrawResponse{DriverId: int64(record.DriverId)}, nil
}
