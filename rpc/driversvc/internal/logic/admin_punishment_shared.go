package logic

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	__proto "XiaoLong-Ridy/rpc/driversvc/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// applyAdminPunishment 在司机域事务内执行处罚动作，并按 event_id/action_type 幂等。
func applyAdminPunishment(ctx context.Context, svcCtx *svc.ServiceContext, in *__proto.AdminPunishmentRequest, reverse bool) (*__proto.CommonResponse, error) {
	if in == nil || in.GetDriverId() <= 0 || strings.TrimSpace(in.GetEventId()) == "" || strings.TrimSpace(in.GetPunishmentNo()) == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id、punishment_no和event_id不能为空")
	}
	if svcCtx == nil || svcCtx.DB == nil || svcCtx.DriverRepository == nil || svcCtx.DriverPunishmentRepository == nil {
		return nil, status.Error(codes.FailedPrecondition, "driver dependencies not ready")
	}
	var actions []string
	if err := json.Unmarshal([]byte(in.GetActions()), &actions); err != nil {
		return nil, status.Error(codes.InvalidArgument, "actions格式不合法")
	}
	if _, err := svcCtx.DriverRepository.GetByID(ctx, uint64(in.GetDriverId())); err != nil {
		return nil, status.Error(codes.NotFound, "driver not found")
	}
	driverID := uint64(in.GetDriverId())
	err := svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		punishmentRepo := svcCtx.DriverPunishmentRepository
		for _, action := range actions {
			exists, err := punishmentRepo.EffectExists(ctx, tx, in.GetEventId(), action)
			if err != nil {
				return err
			}
			if exists {
				continue
			}
			switch action {
			case "freeze":
				value := int8(__proto.DriverStatus_DRIVER_STATUS_FROZEN)
				if reverse {
					value = int8(__proto.DriverStatus_DRIVER_STATUS_NORMAL)
				}
				if err := punishmentRepo.UpdateStatus(ctx, tx, driverID, value, time.Now()); err != nil {
					return err
				}
			case "no_dispatch":
				value := int8(0)
				if reverse {
					value = 1
				}
				if err := punishmentRepo.UpdateOnlineStatus(ctx, tx, driverID, value, time.Now()); err != nil {
					return err
				}
				if err := punishmentRepo.UpdateLocationOnlineStatus(ctx, tx, driverID, value); err != nil {
					return err
				}
			case "deduct_score":
				delta := float64(in.GetScoreDelta())
				if reverse {
					delta = -delta
				}
				if err := punishmentRepo.AddScore(ctx, tx, driverID, delta, time.Now()); err != nil {
					return err
				}
			case "downgrade":
				delta := int8(in.GetPriorityWeightDelta())
				if reverse {
					delta = -delta
				}
				if err := punishmentRepo.AddLevel(ctx, tx, driverID, delta, time.Now()); err != nil {
					return err
				}
			case "fine":
			default:
				return status.Error(codes.InvalidArgument, "处罚动作不合法")
			}
			effect := &model.DriverPunishmentEffect{
				EventId:             in.GetEventId(),
				PunishmentNo:        in.GetPunishmentNo(),
				DriverId:            driverID,
				ActionType:          action,
				ScoreDelta:          in.GetScoreDelta(),
				PriorityWeightDelta: in.GetPriorityWeightDelta(),
			}
			if err := punishmentRepo.RecordEffect(ctx, tx, effect); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &__proto.CommonResponse{Message: "ok"}, nil
}
