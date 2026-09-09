package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"XiaoLong-Ridy/api/chat/internal/config"
	"XiaoLong-Ridy/api/chat/internal/handler"
	"XiaoLong-Ridy/api/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
)

func main() {
	configPath := flag.String("f", "etc/chat.yaml", "chat api config path")
	flag.Parse()

	var cfg config.Config
	// 复用 go-zero 配置加载器：YAML→JSON 后按 json 标签（大小写不敏感）解析，
	// 与 chatsvc 等其它服务保持一致，能正确解析 zrpc.RpcClientConf 等嵌套配置。
	conf.MustLoad(*configPath, &cfg)

	signingKeys := cfg.SigningKeys
	if env := os.Getenv("CHAT_SIGNING_KEYS"); env != "" {
		signingKeys = strings.Split(env, ",")
	}
	if len(signingKeys) == 0 {
		log.Fatal("未配置任何 JWT 签名密钥")
	}

	svcCtx := svc.NewServiceContext(cfg.ChatRpc, signingKeys)

	root := recoverMiddleware(withCORS(handler.NewRouter(svcCtx)))
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           root,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("chat api 监听 %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("chat api 启动失败: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
	log.Println("chat api 已关闭")
}

// withCORS 允许跨域（司机端 web 经 vite 代理访问）。
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// recoverMiddleware 兜底 panic，避免单请求崩溃拖垮进程。
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"code": 50000, "message": "服务内部错误", "data": nil,
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
