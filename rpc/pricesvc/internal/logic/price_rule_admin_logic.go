package logic

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/pricesvc/internal/model"
	"XiaoLong-Ridy/rpc/pricesvc/internal/repository"
	"XiaoLong-Ridy/rpc/pricesvc/internal/svc"
	"XiaoLong-Ridy/rpc/pricesvc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

const (
	priceRuleTimeLayout = "15:04:05"
	priceRuleDateLayout = "2006-01-02 15:04:05"
)

// PriceRuleAdminLogic 负责计价规则管理 RPC。
type PriceRuleAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewPriceRuleAdminLogic 创建计价规则管理逻辑对象。
func NewPriceRuleAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PriceRuleAdminLogic {
	return &PriceRuleAdminLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListPriceRules 查询计价规则列表。
func (l *PriceRuleAdminLogic) ListPriceRules(in *proto.PriceRuleListRequest) (*proto.PriceRuleListResponse, error) {
	limit := normalizePageSize(int(in.GetPageSize()))
	list, total, err := repository.NewPriceRuleRepo(l.svcCtx.DB).List(
		l.ctx,
		in.GetKeyword(),
		in.GetCityCode(),
		in.GetCarType(),
		in.GetStatus(),
		limit,
		offset(int(in.GetPage()), limit),
	)
	if err != nil {
		return nil, err
	}
	items := make([]*proto.PriceRule, 0, len(list))
	for i := range list {
		items = append(items, priceRuleModelToPB(&list[i]))
	}
	return &proto.PriceRuleListResponse{
		List:     items,
		Total:    total,
		Page:     int32(normalizePage(int(in.GetPage()))),
		PageSize: int32(limit),
	}, nil
}

// GetPriceRule 查询计价规则详情。
func (l *PriceRuleAdminLogic) GetPriceRule(in *proto.PriceRuleDetailRequest) (*proto.PriceRule, error) {
	if in.GetId() <= 0 {
		return nil, invalidPriceRuleArgument("price rule id不能为空")
	}
	rule, err := repository.NewPriceRuleRepo(l.svcCtx.DB).GetByID(l.ctx, in.GetId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, priceRuleNotFound()
		}
		return nil, err
	}
	return priceRuleModelToPB(rule), nil
}

// CreatePriceRule 新增计价规则。
func (l *PriceRuleAdminLogic) CreatePriceRule(in *proto.PriceRuleRequest) (*proto.CommonResponse, error) {
	rule, err := priceRuleRequestToModel(in, true)
	if err != nil {
		return nil, invalidPriceRuleArgument(err.Error())
	}
	if err := repository.NewPriceRuleRepo(l.svcCtx.DB).Create(l.ctx, rule); err != nil {
		return nil, err
	}
	return &proto.CommonResponse{Message: "ok"}, nil
}

// UpdatePriceRule 更新计价规则。
func (l *PriceRuleAdminLogic) UpdatePriceRule(in *proto.PriceRuleRequest) (*proto.CommonResponse, error) {
	rule, err := priceRuleRequestToModel(in, false)
	if err != nil {
		return nil, invalidPriceRuleArgument(err.Error())
	}
	if rule.Id <= 0 {
		return nil, invalidPriceRuleArgument("price rule id不能为空")
	}
	if err := repository.NewPriceRuleRepo(l.svcCtx.DB).Update(l.ctx, rule); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, priceRuleNotFound()
		}
		return nil, err
	}
	return &proto.CommonResponse{Message: "ok"}, nil
}

// SetPriceRuleStatus 更新计价规则状态。
func (l *PriceRuleAdminLogic) SetPriceRuleStatus(in *proto.PriceRuleStatusRequest) (*proto.CommonResponse, error) {
	if in.GetId() <= 0 {
		return nil, invalidPriceRuleArgument("price rule id不能为空")
	}
	if in.GetStatus() != 1 && in.GetStatus() != 2 {
		return nil, invalidPriceRuleArgument("price rule status不合法")
	}
	if err := repository.NewPriceRuleRepo(l.svcCtx.DB).UpdateStatus(l.ctx, in.GetId(), in.GetStatus()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, priceRuleNotFound()
		}
		return nil, err
	}
	return &proto.CommonResponse{Message: "ok"}, nil
}

// invalidPriceRuleArgument 将计价规则参数错误转换为标准 gRPC InvalidArgument。
func invalidPriceRuleArgument(message string) error {
	return status.Error(codes.InvalidArgument, message)
}

// priceRuleNotFound 将计价规则不存在转换为标准 gRPC NotFound。
func priceRuleNotFound() error {
	return status.Error(codes.NotFound, "计价规则不存在")
}

// priceRuleRequestToModel 将 RPC 请求转换为计价规则模型。
func priceRuleRequestToModel(in *proto.PriceRuleRequest, isCreate bool) (*model.PriceRule, error) {
	if in == nil {
		return nil, fmt.Errorf("price rule request不能为空")
	}
	if strings.TrimSpace(in.GetName()) == "" {
		return nil, fmt.Errorf("price rule name不能为空")
	}
	if strings.TrimSpace(in.GetCityCode()) == "" {
		in.CityCode = ""
	}
	if in.GetCarType() <= 0 {
		return nil, fmt.Errorf("price rule car_type不能为空")
	}
	if in.GetStatus() != 1 && in.GetStatus() != 2 {
		return nil, fmt.Errorf("price rule status不合法")
	}
	basePrice, err := parseDecimal(in.GetBasePrice())
	if err != nil {
		return nil, fmt.Errorf("base_price无效: %w", err)
	}
	baseDistanceKm, err := parseDecimal(in.GetBaseDistanceKm())
	if err != nil {
		return nil, fmt.Errorf("base_distance_km无效: %w", err)
	}
	perKmPrice, err := parseDecimal(in.GetPerKmPrice())
	if err != nil {
		return nil, fmt.Errorf("per_km_price无效: %w", err)
	}
	perMinutePrice, err := parseDecimal(in.GetPerMinutePrice())
	if err != nil {
		return nil, fmt.Errorf("per_minute_price无效: %w", err)
	}
	nightSurcharge, err := parseDecimal(in.GetNightSurcharge())
	if err != nil {
		return nil, fmt.Errorf("night_surcharge无效: %w", err)
	}
	dynamicMaxFactor, err := parseDecimal(in.GetDynamicMaxFactor())
	if err != nil {
		return nil, fmt.Errorf("dynamic_max_factor无效: %w", err)
	}
	effectiveAt, err := parseRequiredTime(in.GetEffectiveAt())
	if err != nil {
		return nil, fmt.Errorf("effective_at无效: %w", err)
	}
	expireAt, err := parseOptionalDateTime(in.GetExpireAt())
	if err != nil {
		return nil, fmt.Errorf("expire_at无效: %w", err)
	}
	if expireAt != nil && expireAt.Before(effectiveAt) {
		return nil, fmt.Errorf("expire_at不能早于effective_at")
	}
	nightStartTime, err := parseOptionalClock(in.GetNightStartTime())
	if err != nil {
		return nil, fmt.Errorf("night_start_time无效: %w", err)
	}
	nightEndTime, err := parseOptionalClock(in.GetNightEndTime())
	if err != nil {
		return nil, fmt.Errorf("night_end_time无效: %w", err)
	}
	if (nightStartTime == nil) != (nightEndTime == nil) {
		return nil, fmt.Errorf("夜间起止时间必须同时存在或同时为空")
	}
	rule := &model.PriceRule{
		Id:               uint64(in.GetId()),
		Name:             strings.TrimSpace(in.GetName()),
		CityCode:         strings.TrimSpace(in.GetCityCode()),
		CarType:          int8(in.GetCarType()),
		BasePrice:        basePrice,
		BaseDistanceKm:   baseDistanceKm,
		PerKmPrice:       perKmPrice,
		PerMinutePrice:   perMinutePrice,
		NightStartTime:   nightStartTime,
		NightEndTime:     nightEndTime,
		NightSurcharge:   nightSurcharge,
		DynamicMaxFactor: dynamicMaxFactor,
		Status:           int8(in.GetStatus()),
		EffectiveAt:      effectiveAt,
		ExpireAt:         expireAt,
	}
	if !isCreate && rule.Id == 0 {
		return nil, fmt.Errorf("price rule id不能为空")
	}
	return rule, nil
}

// priceRuleModelToPB 将模型转换为 RPC 对象。
func priceRuleModelToPB(rule *model.PriceRule) *proto.PriceRule {
	if rule == nil {
		return nil
	}
	return &proto.PriceRule{
		Id:               int64(rule.Id),
		Name:             rule.Name,
		CityCode:         rule.CityCode,
		CarType:          int32(rule.CarType),
		BasePrice:        formatFloat(rule.BasePrice),
		BaseDistanceKm:   formatFloat(rule.BaseDistanceKm),
		PerKmPrice:       formatFloat(rule.PerKmPrice),
		PerMinutePrice:   formatFloat(rule.PerMinutePrice),
		NightStartTime:   derefString(rule.NightStartTime),
		NightEndTime:     derefString(rule.NightEndTime),
		NightSurcharge:   formatFloat(rule.NightSurcharge),
		DynamicMaxFactor: formatFloat(rule.DynamicMaxFactor),
		Status:           int32(rule.Status),
		EffectiveAt:      rule.EffectiveAt.Format(priceRuleDateLayout),
		ExpireAt:         formatOptionalTime(rule.ExpireAt),
		CreatedAt:        rule.CreatedAt.Format(priceRuleDateLayout),
		UpdatedAt:        rule.UpdatedAt.Format(priceRuleDateLayout),
	}
}

// parseDecimal 解析保留两位小数的字符串。
func parseDecimal(v string) (float64, error) {
	if strings.TrimSpace(v) == "" {
		return 0, fmt.Errorf("不能为空")
	}
	return strconv.ParseFloat(v, 64)
}

// normalizePage 统一修正后台计价规则分页页码，避免非法页码传入仓储层。
func normalizePage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

// normalizePageSize 统一限制后台计价规则分页大小，避免一次性读取过多数据。
func normalizePageSize(pageSize int) int {
	if pageSize <= 0 {
		return 20
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}

// offset 根据页码和分页大小计算数据库查询偏移量。
func offset(page, pageSize int) int {
	return (normalizePage(page) - 1) * normalizePageSize(pageSize)
}

// parseRequiredTime 解析必填日期时间。
func parseRequiredTime(v string) (time.Time, error) {
	return time.ParseInLocation(priceRuleDateLayout, strings.TrimSpace(v), time.Local)
}

// parseOptionalDateTime 解析可空日期时间。
func parseOptionalDateTime(v string) (*time.Time, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, nil
	}
	t, err := time.ParseInLocation(priceRuleDateLayout, v, time.Local)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// parseOptionalClock 解析可空时分秒时间。
func parseOptionalClock(v string) (*string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, nil
	}
	if _, err := time.ParseInLocation(priceRuleTimeLayout, v, time.Local); err != nil {
		return nil, err
	}
	return &v, nil
}

// formatOptionalTime 将可空时间格式化为字符串。
func formatOptionalTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(priceRuleDateLayout)
}

// derefString 读取可空字符串。
func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// formatFloat 将数值格式化为两位小数字符串。
func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}
