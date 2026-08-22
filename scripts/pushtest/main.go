package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"XiaoLong-Ridy/rpc/pushesvc/pushesvc"
)

// 测试客户端：直接调用 pushesvc gRPC 三个接口，验证落库。
// 运行：go run ./scripts/pushtest  （需 pushesvc 已在 127.0.0.1:9002 运行）
func main() {
	addr := "127.0.0.1:9002"
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("连接失败:", err)
		os.Exit(1)
	}
	defer conn.Close()

	c := pushesvc.NewPushServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log := func(name string, resp interface{}, err error) {
		if err != nil {
			fmt.Printf("[失败] %s: %v\n", name, err)
		} else {
			fmt.Printf("[成功] %s: %+v\n", name, resp)
		}
	}

	// 站内信：真实落库 notices 表
	n, ne := c.SendNotice(ctx, &pushesvc.SendNoticeReq{
		UserId: 1001, Title: "订单通知", Content: "您的订单已接单", BizType: 1,
	})
	log("SendNotice", n, ne)

	// App 推送：noop 模拟成功，落 push_log（含 extras）
	p, pe := c.SendPush(ctx, &pushesvc.SendPushReq{
		UserId: 1001, Title: "新订单", Body: "附近有新的订单", DeviceType: "android", Extras: `{"orderId":888}`,
	})
	log("SendPush", p, pe)

	// 短信：noop 模拟成功，落 push_log（含 biz_type）
	s, se := c.SendSMS(ctx, &pushesvc.SendSMSReq{
		Phone: "13800000001", Content: "您的验证码是1234", BizType: 1,
	})
	log("SendSMS", s, se)

	fmt.Println("\n已发送三类消息，请到 MySQL 核查 notices / push_log 表。")
}
