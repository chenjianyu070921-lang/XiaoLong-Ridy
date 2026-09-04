package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListWithdrawsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListWithdrawsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListWithdrawsLogic {
	return &ListWithdrawsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListWithdraws 查询司机提现记录，按申请时间倒序分页返回。
func (l *ListWithdrawsLogic) ListWithdraws(in *proto.ListWithdrawsRequest) (*proto.ListWithdrawsResponse, error) {
	if in == nil || in.DriverId <= 0 {
		return nil, errors.New("请求参数不能为空")
	}
	if l.svcCtx == nil || l.svcCtx.DriverWithdrawRepository == nil {
		return nil, errors.New("driver withdraw repository not ready")
	}
	records, total, err := l.svcCtx.DriverWithdrawRepository.ListByDriver(l.ctx, uint64(in.DriverId), in.Page, in.PageSize)
	if err != nil {
		return nil, err
	}
	resp := &proto.ListWithdrawsResponse{
		Total: total,
	}
	for _, record := range records {
		resp.Records = append(resp.Records, toWithdrawRecord(record))
	}
	return resp, nil
}

func toWithdrawRecord(w *model.DriverWithdraw) *proto.WithdrawRecord {
	if w == nil {
		return nil
	}
	record := &proto.WithdrawRecord{
		Id:         int64(w.Id),
		DriverId:   int64(w.DriverId),
		WithdrawNo: w.WithdrawNo,
		Amount:     w.Amount,
		PayeeName:  w.PayeeName,
		PayAccount: w.PayAccount,
		Status:     int32(w.Status),
		Remark:     w.Remark,
		CreatedAt:  w.CreatedAt.Unix(),
	}
	if w.AppliedAt != nil {
		record.AppliedAt = w.AppliedAt.Unix()
	}
	if w.PaidAt != nil {
		record.PaidAt = w.PaidAt.Unix()
	}
	return record
}
