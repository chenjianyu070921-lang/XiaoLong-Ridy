order-goctl:
	goctl rpc protoc rpc/ordersvc/proto/ordersvc.proto \
    --go_out=rpc/ordersvc/proto \
    --go-grpc_out=rpc/ordersvc/proto \
    --zrpc_out=rpc/ordersvc

