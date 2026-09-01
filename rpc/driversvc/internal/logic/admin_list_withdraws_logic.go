package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/rpc/driversvc/internal/repository"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AdminListWithdrawsLogic 封装管理后台查询提现申请的业务逻辑。
type AdminListWithdrawsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewAdminListWithdrawsLogic 创建管理后台提现列表逻辑处理器。
func NewAdminListWithdrawsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListWithdrawsLogic {
	return &AdminListWithdrawsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AdminListWithdraws 按状态、司机 ID 和关键词（提现单号/收款人/收款账户）分页查询提现记录。
func (l *AdminListWithdrawsLogic) AdminListWithdraws(in *proto.AdminListWithdrawsRequest) (*proto.ListWithdrawsResponse, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if l.svcCtx == nil || l.svcCtx.DriverWithdrawRepository == nil {
		return nil, status.Error(codes.FailedPrecondition, "driver withdraw repository not ready")
	}
	filter := repository.AdminWithdrawFilter{
		Page:     in.GetPage(),
		PageSize: in.GetPageSize(),
		Status:   in.GetStatus(),
		DriverID: uint64(in.GetDriverId()),
		Keyword:  strings.TrimSpace(in.GetKeyword()),
	}
	if in.Status != nil && in.GetStatus() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid withdraw status filter")
	}
	records, total, err := l.svcCtx.DriverWithdrawRepository.AdminList(l.ctx, filter)
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
