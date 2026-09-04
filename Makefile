# Go 编译并行数（Windows 上默认值太高容易导致 "Insufficient system resources" 错误）
export GOFLAGS=-p=2

# RPC 服务 goctl 生成命令
order-goctl:
	goctl rpc protoc rpc/ordersvc/proto/ordersvc.proto \
	--go_out=rpc/ordersvc/proto \
	--go-grpc_out=rpc/ordersvc/proto \
	--zrpc_out=rpc/ordersvc \
	--style=go_zero

dispatchsvc-goctl:
	goctl rpc protoc rpc/dispatchsvc/proto/dispatchsvc.proto \
	--go_out=rpc/dispatchsvc/proto \
	--go-grpc_out=rpc/dispatchsvc/proto \
	--zrpc_out=rpc/dispatchsvc \
	--style=go_zero

locationsvc-goctl:
	goctl rpc protoc rpc/locationsvc/proto/locationsvc.proto \
	--go_out=rpc/locationsvc/proto \
	--go-grpc_out=rpc/locationsvc/proto \
	--zrpc_out=rpc/locationsvc \
	--style=go_zero

paysvc-goctl:
	goctl rpc protoc rpc/paysvc/proto/paysvc.proto \
	--go_out=rpc/paysvc/proto \
	--go-grpc_out=rpc/paysvc/proto \
	--zrpc_out=rpc/paysvc \
	--style=go_zero

pricesvc-goctl:
	goctl rpc protoc rpc/pricesvc/proto/pricesvc.proto \
	--go_out=rpc/pricesvc/proto \
	--go-grpc_out=rpc/pricesvc/proto \
	--zrpc_out=rpc/pricesvc \
	--style=go_zero

adminsvc-goctl:
	powershell -ExecutionPolicy Bypass -File scripts/admin-test/regenerate_adminsvc_proto.ps1

pushesvc-goctl:
	goctl rpc protoc rpc/pushesvc/proto/pushesvc.proto \
	--go_out=rpc/pushesvc/proto \
	--go-grpc_out=rpc/pushesvc/proto \
	--zrpc_out=rpc/pushesvc \
	--style=go_zero

usersvc-goctl:
	goctl rpc protoc rpc/usersvc/proto/usersvc.proto \
	--go_out=rpc/usersvc/proto \
	--go-grpc_out=rpc/usersvc/proto \
	--zrpc_out=rpc/usersvc \
	--style=go_zero
# RPC 服务 main 启动命令

run-ordersvc:
	go run .\rpc\ordersvc\ordersvc.go -f .\rpc\ordersvc\etc\ordersvc.yaml

run-usersvc:
	go run .\rpc\usersvc\main.go -f .\rpc\usersvc\etc\usersvc.yaml

run-adminsvc:
	go run .\rpc\adminsvc\admin.go -f .\rpc\adminsvc\etc\admin.yaml

run-driversvc:
	go run .\rpc\driversvc\driversvc.go -f .\rpc\driversvc\etc\driversvc.yaml

run-dispatchsvc:
	go run .\rpc\dispatchsvc\dispatchsvc.go -f .\rpc\dispatchsvc\etc\dispatchsvc.yaml

run-locationsvc:
	go run .\rpc\locationsvc\locationsvc.go -f .\rpc\locationsvc\etc\locationsvc.yaml

run-pricesvc:
	go run .\rpc\pricesvc\pricesvc.go -f .\rpc\pricesvc\etc\pricesvc.yaml

run-paysvc:
	go run .\rpc\paysvc\paysvc.go -f .\rpc\paysvc\etc\paysvc.yaml

run-pushesvc:
	go run .\rpc\pushesvc\pushesvc.go -f .\rpc\pushesvc\etc\pushesvc.yaml

# API 服务 main 启动命令

run-passenger-api:
	powershell -Command "cd api/passenger; go run ."

run-driver-api:
	powershell -Command "cd api/driver; go run ."

run-admin-api:
	powershell -Command "cd api/admin; go run ."

# 前端启动命令

install-web:
	powershell -Command "cd web/user; npm install"
	powershell -Command "cd web/admin; npm install"

run-user-web:
	powershell -Command "cd web/user; npm run dev"

run-admin-web:
	powershell -Command "cd web/admin; npm run dev"

run-driver-web:
	powershell -Command "cd web/driver; npm run dev"
# 消息消费者与定时任务启动命令

run-order-event-consumer:
	go run .\mq-consumer\order-event-consumer\main.go -f .\mq-consumer\order-event-consumer\etc\order-event-consumer.yaml

run-location-consumer:
	go run .\mq-consumer\location-consumer\locationconsumer.go -f .\mq-consumer\location-consumer\etc\location-consumer.yaml

run-job:
	go run .\job\job.go -f .\job\etc\job.yaml

.PHONY: run-ordersvc run-usersvc run-adminsvc run-driversvc run-dispatchsvc run-locationsvc run-pricesvc run-paysvc run-pushesvc run-passenger-api run-driver-api run-admin-api run-order-event-consumer run-location-consumer run-job run-user-web run-admin-web install-web

