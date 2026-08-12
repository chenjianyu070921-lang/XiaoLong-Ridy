package server

import (
	"context"

	"XiaoLong-Ridy/rpc/pushesvc/internal/logic"
	"XiaoLong-Ridy/rpc/pushesvc/internal/svc"
	"XiaoLong-Ridy/rpc/pushesvc/pushesvc"
)

type PushServiceServer struct {
	svcCtx *svc.ServiceContext
	pushesvc.UnimplementedPushServiceServer
}

func NewPushServiceServer(svcCtx *svc.ServiceContext) *PushServiceServer {
	return &PushServiceServer{
		svcCtx: svcCtx,
	}
}

func (s *PushServiceServer) SendNotice(ctx context.Context, in *pushesvc.SendNoticeReq) (*pushesvc.SendNoticeResp, error) {
	l := logic.NewSendNoticeLogic(ctx, s.svcCtx)
	return l.SendNotice(in)
}

func (s *PushServiceServer) SendPush(ctx context.Context, in *pushesvc.SendPushReq) (*pushesvc.SendPushResp, error) {
	l := logic.NewSendPushLogic(ctx, s.svcCtx)
	return l.SendPush(in)
}

func (s *PushServiceServer) SendSMS(ctx context.Context, in *pushesvc.SendSMSReq) (*pushesvc.SendSMSResp, error) {
	l := logic.NewSendSMSLogic(ctx, s.svcCtx)
	return l.SendSMS(in)
}
