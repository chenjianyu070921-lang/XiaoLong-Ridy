package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"XiaoLong-Ridy/rpc/usersvc/client"
)

// main 是 usersvc 的开发期启动入口。
// 当前 HTTP 网关通过本地 client 访问 usersvc；接入 gRPC 后可将此处替换为 RPC Server。
func main() {
	_ = client.NewLocalClient("local-development-signing-key", func(phone, code string) {
		log.Printf("本地短信验证码：phone=%s code=%s", phone, code)
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	log.Println("usersvc started, waiting for gRPC transport integration...")
	<-stop
	log.Println("usersvc stopped")
}
