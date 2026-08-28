package adminservicelogic

import (
	"context"
	"strconv"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	locationsvc "XiaoLong-Ridy/rpc/locationsvc/locationsvc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetOrderTrackLogic 处理管理后台订单轨迹查询。
type GetOrderTrackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetOrderTrackLogic 创建订单轨迹查询逻辑对象。
func NewGetOrderTrackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderTrackLogic {
	return &GetOrderTrackLogic{ctx: ctx, svcCtx: svcCtx}
}

// GetOrderTrack 通过 locationsvc 查询订单轨迹点，adminsvc 不直接读取位置服务表。
// 该接口只读，用于后台客服查看行程轨迹、处理绕路投诉和工单证据核验。
func (l *GetOrderTrackLogic) GetOrderTrack(in *adminsvc.OrderTrackRequest) (*adminsvc.OrderTrackResponse, error) {
	if err := requireAdminRoles(l.ctx, l.svcCtx, 1, 2, 3); err != nil {
		return nil, err
	}
	if in.GetOrderId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "订单ID不能为空")
	}
	if l.svcCtx == nil || l.svcCtx.LocationSvc == nil {
		return nil, status.Error(codes.Unavailable, "位置服务未启动")
	}
	resp, err := l.svcCtx.LocationSvc.GetOrderTrack(l.ctx, &locationsvc.GetOrderTrackReq{
		OrderId:   in.GetOrderId(),
		StartTime: in.GetStartTime(),
		EndTime:   in.GetEndTime(),
		Limit:     in.GetLimit(),
	})
	if err != nil {
		return nil, err
	}
	points := make([]*adminsvc.OrderTrackPoint, 0, len(resp.GetPoints()))
	for _, point := range resp.GetPoints() {
		points = append(points, &adminsvc.OrderTrackPoint{
			Id:         point.GetId(),
			OrderId:    point.GetOrderId(),
			DriverId:   point.GetDriverId(),
			Longitude:  formatFloat(point.GetLng()),
			Latitude:   formatFloat(point.GetLat()),
			SpeedKmh:   formatFloat(point.GetSpeedKmh()),
			Direction:  point.GetDirection(),
			RecordedAt: unixText(point.GetRecordedAt()),
		})
	}
	return &adminsvc.OrderTrackResponse{Points: points}, nil
}

// formatFloat 将位置服务 double 转换为后台接口字符串，保持经纬度和速度展示稳定。
func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
