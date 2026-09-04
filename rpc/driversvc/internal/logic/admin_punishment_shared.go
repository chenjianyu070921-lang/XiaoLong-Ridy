package logic

import (
	"context"
	"encoding/json"
	"strings"
	"time"

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
	if svcCtx == nil || svcCtx.DB == nil || svcCtx.DriverRepository == nil {
		return nil, status.Error(codes.FailedPrecondition, "driver dependencies not ready")
	}
	var actions []string
	if err := json.Unmarshal([]byte(in.GetActions()), &actions); err != nil {
		return nil, status.Error(codes.InvalidArgument, "actions格式不合法")
	}
	if _, err := svcCtx.DriverRepository.GetByID(ctx, uint64(in.GetDriverId())); err != nil {
		return nil, status.Error(codes.NotFound, "driver not found")
	}
	err := svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, action := range actions {
			var exists int64
			if err := tx.Table("driver_punishment_effect").Where("event_id=? AND action_type=?", in.GetEventId(), action).Count(&exists).Error; err != nil {
				return err
			}
			if exists > 0 {
				continue
			}
			switch action {
			case "freeze":
				value := __proto.DriverStatus_DRIVER_STATUS_FROZEN
				if reverse {
					value = __proto.DriverStatus_DRIVER_STATUS_NORMAL
				}
				if err := tx.Table("driver").Where("id=?", in.GetDriverId()).Updates(map[string]interface{}{"status": value, "updated_at": time.Now()}).Error; err != nil {
					return err
				}
			case "no_dispatch":
				value := 0
				if reverse {
					value = 1
				}
				if err := tx.Table("driver").Where("id=?", in.GetDriverId()).Updates(map[string]interface{}{"online_status": value, "updated_at": time.Now()}).Error; err != nil {
					return err
				}
				if err := tx.Table("driver_location").Where("driver_id=?", in.GetDriverId()).Updates(map[string]interface{}{"online_status": value}).Error; err != nil {
					return err
				}
			case "deduct_score":
				delta := float64(in.GetScoreDelta())
				if reverse {
					delta = -delta
				}
				if err := tx.Exec("UPDATE driver_score SET score=GREATEST(0, score + ?), updated_at=? WHERE driver_id=?", delta, time.Now(), in.GetDriverId()).Error; err != nil {
					return err
				}
			case "downgrade":
				delta := in.GetPriorityWeightDelta()
				if reverse {
					delta = -delta
				}
				if err := tx.Exec("UPDATE driver_score SET level=GREATEST(1, level + ?), updated_at=? WHERE driver_id=?", delta, time.Now(), in.GetDriverId()).Error; err != nil {
					return err
				}
			case "fine":
			default:
				return status.Error(codes.InvalidArgument, "处罚动作不合法")
			}
			if err := tx.Exec("INSERT INTO driver_punishment_effect(event_id,punishment_no,driver_id,action_type,score_delta,priority_weight_delta) VALUES (?,?,?,?,?,?)", in.GetEventId(), in.GetPunishmentNo(), in.GetDriverId(), action, in.GetScoreDelta(), in.GetPriorityWeightDelta()).Error; err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
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
