package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"XiaoLong-Ridy/rpc/driversvc/proto"
)

// 测试客户端：直接调用 driversvc gRPC 司机接口，插入示例数据并保留在库中，方便核查。
// 运行：go run ./scripts/drivertest  （需 driversvc 已在 127.0.0.1:8080 运行）
func main() {
	addr := "127.0.0.1:8080"
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("连接失败:", err)
		os.Exit(1)
	}
	defer conn.Close()

	c := proto.NewDriversvcClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log := func(name string, resp interface{}, err error) {
		if err != nil {
			fmt.Printf("[失败] %s: %v\n", name, err)
		} else {
			fmt.Printf("[成功] %s: %+v\n", name, resp)
		}
	}

	drivers := []proto.CreateDriverRequest{
		{Phone: "13800000001", PasswordHash: "e10adc3949ba59abbe56e057f20f883e", RealName: "王伟", IdCardNo: "110101199001011234", DriverLicenseNo: "DL10000001", AvatarUrl: "http://cdn.xxx/avatar/1.png"},
		{Phone: "13800000002", PasswordHash: "e10adc3949ba59abbe56e057f20f883e", RealName: "李娜", IdCardNo: "110101199203054321", DriverLicenseNo: "DL10000002", AvatarUrl: "http://cdn.xxx/avatar/2.png"},
		{Phone: "13800000003", PasswordHash: "e10adc3949ba59abbe56e057f20f883e", RealName: "张强", IdCardNo: "110101198807189876", DriverLicenseNo: "DL10000003", AvatarUrl: "http://cdn.xxx/avatar/3.png"},
	}

	driverIDs := make([]int64, 0, len(drivers))
	for i, d := range drivers {
		r, e := c.CreateDriver(ctx, &d)
		log(fmt.Sprintf("CreateDriver-%d", i+1), r, e)
		if r != nil {
			driverIDs = append(driverIDs, r.Id)
		}
	}

	for i, id := range driverIDs {
		u, ue := c.UpdateDriver(ctx, &proto.UpdateDriverRequest{
			Id:     id,
			Status: proto.DriverStatus_DRIVER_STATUS_NORMAL.Enum(),
		})
		log(fmt.Sprintf("UpdateDriver-置正常-%d", i+1), u, ue)
	}

	for i, id := range driverIDs {
		g, ge := c.GetDriver(ctx, &proto.GetDriverRequest{Id: id})
		log(fmt.Sprintf("GetDriver-%d", i+1), g, ge)
	}

	fmt.Println("\n示例司机数据已全部写入 driver 库，未删除，可直接在 MySQL 中核查。")
}
