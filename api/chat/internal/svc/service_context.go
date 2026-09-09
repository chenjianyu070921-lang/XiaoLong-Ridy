package svc

import (
	chatproto "XiaoLong-Ridy/rpc/chatsvc/proto"

	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 承载 api/chat 的依赖：chatsvc gRPC 客户端 + JWT 密钥列表。
type ServiceContext struct {
	ChatClient  chatproto.ChatClient
	SigningKeys []string
}

func NewServiceContext(chatRpc zrpc.RpcClientConf, signingKeys []string) *ServiceContext {
	conn := zrpc.MustNewClient(chatRpc).Conn()
	return &ServiceContext{
		ChatClient:  chatproto.NewChatClient(conn),
		SigningKeys: signingKeys,
	}
}
