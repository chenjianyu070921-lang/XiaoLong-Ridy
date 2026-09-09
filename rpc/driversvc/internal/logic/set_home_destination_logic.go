package logic

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// defaultMaxDetourRatio 顺路判定默认允许的最大绕路比例（相对司机直达回家的路程）。
const defaultMaxDetourRatio = 0.2

type SetHomeDestinationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetHomeDestinationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetHomeDestinationLogic {
	return &SetHomeDestinationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SetHomeDestination 保存（或更新）司机回家目的地；open=true 时同时开启回家顺路模式。
// 注意：设置目的地不等于自动改听单状态，只是开启订单顺路过滤。
func (l *SetHomeDestinationLogic) SetHomeDestination(in *proto.SetHomeDestinationRequest) (*proto.HomeDestinationResponse, error) {
	if in == nil || in.DriverId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "driver_id required")
	}
	if l.svcCtx == nil || l.svcCtx.DriverHomeDestinationRepository == nil {
		return nil, status.Error(codes.Internal, "home destination repository not ready")
	}
	addr := strings.TrimSpace(in.HomeAddr)
	if addr == "" || !validLongitudeLatitude(in.HomeLng, in.HomeLat) {
		return nil, status.Error(codes.InvalidArgument, "home address or coordinate invalid")
	}
	ratio := in.MaxDetourRatio
	if ratio <= 0 {
		ratio = defaultMaxDetourRatio
	}
	openValue := int8(0)
	if in.Open {
		openValue = 1
	}

	setting := &model.DriverHomeDestination{
		DriverID:       uint64(in.DriverId),
		HomeAddr:       addr,
		HomeLng:        in.HomeLng,
		HomeLat:        in.HomeLat,
		IsHomeModeOpen: openValue,
		MaxDetourRatio: ratio,
	}
	if err := l.svcCtx.DriverHomeDestinationRepository.Upsert(l.ctx, setting); err != nil {
		l.Errorf("save home destination failed: driverId=%d err=%v", in.DriverId, err)
		return nil, status.Error(codes.Internal, "save home destination failed")
	}
	syncHomeDestinationCache(l.ctx, l.svcCtx, setting)
	return homeDestinationResponse(setting), nil
}

// syncHomeDestinationCache 将回家目的地写入 Redis，供派单引擎与司机端大厅做顺路过滤时快速读取。
// 缓存失败只记录日志：过滤层读不到会退化为"不过滤"，不会阻断司机听单。
func syncHomeDestinationCache(ctx context.Context, svcCtx *svc.ServiceContext, setting *model.DriverHomeDestination) {
	if svcCtx == nil || svcCtx.RedisClient == nil || setting == nil || setting.DriverID == 0 {
		return
	}
	key := fmt.Sprintf(constants.RedisDriverHome, setting.DriverID)
	fields := map[string]interface{}{
		"lng":              strconv.FormatFloat(setting.HomeLng, 'f', 6, 64),
		"lat":              strconv.FormatFloat(setting.HomeLat, 'f', 6, 64),
		"addr":             setting.HomeAddr,
		"open":             strconv.Itoa(int(setting.IsHomeModeOpen)),
		"max_detour_ratio": strconv.FormatFloat(setting.MaxDetourRatio, 'f', 3, 64),
		"update_at":        strconv.FormatInt(time.Now().Unix(), 10),
	}
	if err := svcCtx.RedisClient.HSet(ctx, key, fields).Err(); err != nil {
		logx.WithContext(ctx).Errorf("sync home destination cache failed: driverId=%d err=%v", setting.DriverID, err)
	}
}

// homeDestinationResponse 将设置转换为 RPC 响应。
func homeDestinationResponse(setting *model.DriverHomeDestination) *proto.HomeDestinationResponse {
	if setting == nil {
		return &proto.HomeDestinationResponse{HasSetting: false}
	}
	ratio := setting.MaxDetourRatio
	if ratio <= 0 {
		ratio = defaultMaxDetourRatio
	}
	return &proto.HomeDestinationResponse{
		HasSetting:      true,
		DriverId:        int64(setting.DriverID),
		HomeAddr:        setting.HomeAddr,
		HomeLng:         setting.HomeLng,
		HomeLat:         setting.HomeLat,
		IsHomeModeOpen:  setting.IsHomeModeOpen == 1,
		MaxDetourRatio:  ratio,
	}
}
