package server

import (
	"context"

	adminservicelogic "XiaoLong-Ridy/rpc/adminsvc/internal/logic/adminservice"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const adminTokenMetadataKey = "x-admin-token"

const (
	adminRoleSuper = int32(1)
	adminRoleOps   = int32(2)
	adminRoleCS    = int32(3)
)

// NewAuthorizationInterceptor 创建 adminsvc 的服务端授权拦截器。
// 除登录、首次注册和会话校验外，所有后台 RPC 都必须携带由 HTTP 网关透传的管理员 token。
func NewAuthorizationInterceptor(svcCtx *svc.ServiceContext) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if isPublicAdminMethod(info.FullMethod) {
			return handler(ctx, req)
		}
		admin, err := adminservicelogic.ValidateAdminTokenFromContext(ctx, svcCtx)
		if err != nil {
			return nil, err
		}
		if !requestAdminIDMatches(req, admin.ID) {
			return nil, status.Error(codes.PermissionDenied, "请求操作者与管理员会话不一致")
		}
		if !roleAllowed(info.FullMethod, admin.Role) {
			return nil, status.Error(codes.PermissionDenied, "当前管理员无权执行该操作")
		}
		return handler(ctx, req)
	}
}

// requestAdminIDMatches 验证写请求中的 admin_id 与已认证会话一致。
// 使用 protobuf 反射覆盖所有已定义 admin_id 字段的请求，避免新增接口遗漏身份一致性校验。
func requestAdminIDMatches(req any, authenticatedAdminID int64) bool {
	message, ok := req.(interface{ ProtoReflect() protoreflect.Message })
	if !ok {
		return true
	}
	field := message.ProtoReflect().Descriptor().Fields().ByName("admin_id")
	if field == nil {
		return true
	}
	return message.ProtoReflect().Get(field).Int() == authenticatedAdminID
}

// isPublicAdminMethod 返回无需拦截器会话校验的方法。
// Register 仍会在业务逻辑中按数据库中的真实操作者角色完成授权，不能仅信任请求参数。
func isPublicAdminMethod(method string) bool {
	switch method {
	case "/adminsvc.AdminService/Login", "/adminsvc.AdminService/Register", "/adminsvc.AdminService/ValidateSession":
		return true
	default:
		return false
	}
}

// roleAllowed 根据已确认的最小权限矩阵判断角色是否能调用指定 RPC。
func roleAllowed(method string, role int32) bool {
	if role == adminRoleSuper {
		return true
	}
	if role != adminRoleOps && role != adminRoleCS {
		return false
	}
	if role == adminRoleCS {
		switch method {
		case "/adminsvc.AdminService/Logout", "/adminsvc.AdminService/Me", "/adminsvc.AdminService/Menus",
			"/adminsvc.AdminService/ListOperationLogs", "/adminsvc.AdminService/ListUsers", "/adminsvc.AdminService/GetUser",
			"/adminsvc.AdminService/ListDriverCertifications", "/adminsvc.AdminService/GetDriverCertification",
			"/adminsvc.AdminService/ListOrders", "/adminsvc.AdminService/GetOrder", "/adminsvc.AdminService/ListAbnormalOrders",
			"/adminsvc.AdminService/CancelOrder":
			return true
		default:
			return false
		}
	}

	// 运营拥有日常运营查询和配置编辑权限；资金、风控、价格生效及活动正式发布收归超管。
	switch method {
	case "/adminsvc.AdminService/DisableCoupon", "/adminsvc.AdminService/IssueCoupon",
		"/adminsvc.AdminService/CreatePriceRule", "/adminsvc.AdminService/UpdatePriceRule",
		"/adminsvc.AdminService/EnablePriceRule", "/adminsvc.AdminService/DisablePriceRule",
		"/adminsvc.AdminService/PublishPromotionActivity", "/adminsvc.AdminService/RollbackPromotionActivity",
		"/adminsvc.AdminService/AddBlacklist", "/adminsvc.AdminService/ReleaseBlacklist":
		return false
	default:
		return true
	}
}
