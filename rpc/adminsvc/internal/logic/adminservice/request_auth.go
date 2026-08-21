package adminservicelogic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const adminTokenMetadataKey = "x-admin-token"

// ValidateAdminTokenFromContext 从 gRPC metadata 读取管理员 token，并回查 Redis 会话和管理员账号状态。
// HTTP 网关和 RPC 服务端均不信任调用方自报的管理员身份，最终以会话和 admin_user 查询结果为准。
func ValidateAdminTokenFromContext(ctx context.Context, svcCtx *svc.ServiceContext) (*adminRow, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "缺少管理员会话")
	}
	values := md.Get(adminTokenMetadataKey)
	if len(values) != 1 || strings.TrimSpace(values[0]) == "" {
		return nil, status.Error(codes.Unauthenticated, "缺少管理员会话")
	}
	return validateSession(ctx, svcCtx, values[0])
}
