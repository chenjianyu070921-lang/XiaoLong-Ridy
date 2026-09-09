package logic

import (
	"context"

	"XiaoLong-Ridy/api/admin/internal/svc"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
)

// CapacityLogic 负责后台运力地图 HTTP 请求到 adminsvc RPC 的参数转换。
type CapacityLogic struct {
	ctx *svc.ServiceContext
}

// NewCapacityLogic 创建运力地图逻辑对象。
func NewCapacityLogic(ctx *svc.ServiceContext) *CapacityLogic { return &CapacityLogic{ctx: ctx} }

// Map 查询实时运力地图快照。
func (l *CapacityLogic) Map(ctx context.Context, statusCode, onlineStatus, limit int32) (*adminclient.CapacityMapResponse, error) {
	return l.ctx.AdminSvc.GetCapacityMap(ctx, &adminclient.CapacityMapRequest{Status: statusCode, OnlineStatus: onlineStatus, Limit: limit})
}
