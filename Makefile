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

