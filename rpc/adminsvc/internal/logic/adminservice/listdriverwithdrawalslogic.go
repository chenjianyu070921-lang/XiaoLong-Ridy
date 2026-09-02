package adminservicelogic

import (
	"context"
	"strconv"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	driverproto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListDriverWithdrawalsLogic 处理管理后台查询司机提现申请列表 RPC。
type ListDriverWithdrawalsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewListDriverWithdrawalsLogic 创建司机提现列表逻辑对象。
func NewListDriverWithdrawalsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDriverWithdrawalsLogic {
	return &ListDriverWithdrawalsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListDriverWithdrawals 透传 driversvc.AdminListWithdraws 并将金额转为字符串、
// 时间戳转为东八区可读时间，避免管理后台前端再做单位换算。
func (l *ListDriverWithdrawalsLogic) ListDriverWithdrawals(in *adminsvc.DriverWithdrawListRequest) (*adminsvc.DriverWithdrawListResponse, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "请求不能为空")
	}
	if l.svcCtx == nil || l.svcCtx.DriverSvc == nil {
		return nil, status.Error(codes.FailedPrecondition, "driver service is not running or downstream RPC is disabled")
	}
	req := &driverproto.AdminListWithdrawsRequest{
		Page:     in.GetPage(),
		PageSize: in.GetPageSize(),
		DriverId: in.GetDriverId(),
		Keyword:  in.GetKeyword(),
	}
	if status := in.GetStatus(); status > 0 {
		req.Status = &status
	}
	resp, err := l.svcCtx.DriverSvc.AdminListWithdraws(l.ctx, req)
	if err != nil {
		return nil, err
	}
	list := make([]*adminsvc.DriverWithdraw, 0, len(resp.GetRecords()))
	for _, r := range resp.GetRecords() {
		list = append(list, &adminsvc.DriverWithdraw{
			Id:         r.GetId(),
			DriverId:   r.GetDriverId(),
			WithdrawNo: r.GetWithdrawNo(),
			Amount:     strconv.FormatFloat(r.GetAmount(), 'f', 2, 64),
			PayeeName:  r.GetPayeeName(),
			PayAccount: r.GetPayAccount(),
			Status:     r.GetStatus(),
			Remark:     r.GetRemark(),
			AppliedAt:  formatUnixSecond(r.GetAppliedAt()),
			PaidAt:     formatUnixSecond(r.GetPaidAt()),
			CreatedAt:  formatUnixSecond(r.GetCreatedAt()),
		})
	}
	return &adminsvc.DriverWithdrawListResponse{
		List:     list,
		Total:    resp.GetTotal(),
		Page:     in.GetPage(),
		PageSize: in.GetPageSize(),
	}, nil
}
