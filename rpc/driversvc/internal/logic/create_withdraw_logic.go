package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// withdrawStatusPending 提现申请中（司机侧发起后的初始状态）。
const withdrawStatusPending int8 = 1

type CreateWithdrawLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateWithdrawLogic {
	return &CreateWithdrawLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateWithdraw 创建提现申请：校验金额与收款账户，生成提现单号并落库，初始状态为申请中。
// 注：实际打款由 adminsvc 审核后执行，本 RPC 仅负责司机侧发起与记录。
func (l *CreateWithdrawLogic) CreateWithdraw(in *proto.CreateWithdrawRequest) (*proto.CreateWithdrawResponse, error) {
	if in == nil || in.DriverId <= 0 {
		return nil, errors.New("请求参数不能为空")
	}
	if err := validateWithdrawAmount(in.Amount); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.PayeeName) == "" {
		return nil, errors.New("收款人姓名不能为空")
	}
	if strings.TrimSpace(in.PayAccount) == "" {
		return nil, errors.New("收款账户不能为空")
	}

	now := time.Now()
	withdraw := &model.DriverWithdraw{
		DriverId:   uint64(in.DriverId),
		WithdrawNo: generateWithdrawNo(now),
		Amount:     in.Amount,
		PayeeName:  strings.TrimSpace(in.PayeeName),
		PayAccount: strings.TrimSpace(in.PayAccount),
		Status:     withdrawStatusPending,
		AppliedAt:  &now,
	}
	if err := l.svcCtx.DriverWithdrawRepository.Create(l.ctx, withdraw); err != nil {
		return nil, err
	}
	return &proto.CreateWithdrawResponse{
		Id:         int64(withdraw.Id),
		WithdrawNo: withdraw.WithdrawNo,
		Status:     int32(withdraw.Status),
		CreatedAt:  withdraw.CreatedAt.Unix(),
	}, nil
}

// generateWithdrawNo 生成提现单号：WD + 时间(秒) + 4位随机数。
func generateWithdrawNo(now time.Time) string {
	return fmt.Sprintf("WD%010d%04d", now.Unix(), now.Nanosecond()%10000)
}
