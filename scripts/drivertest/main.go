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

// 测试客户端：直接调用 driversvc gRPC 接口，插入示例数据并保留在库中，方便核查。
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

	// ---- 示例数据：3 个司机 + 各自车辆/认证/服务分/提现，全部入库保留 ----

	drivers := []proto.CreateDriverRequest{
		{Phone: "13800000001", PasswordHash: "e10adc3949ba59abbe56e057f20f883e", RealName: "王伟", IdCardNo: "110101199001011234", DriverLicenseNo: "DL10000001", AvatarUrl: "http://cdn.xxx/avatar/1.png"},
		{Phone: "13800000002", PasswordHash: "e10adc3949ba59abbe56e057f20f883e", RealName: "李娜", IdCardNo: "110101199203054321", DriverLicenseNo: "DL10000002", AvatarUrl: "http://cdn.xxx/avatar/2.png"},
		{Phone: "13800000003", PasswordHash: "e10adc3949ba59abbe56e057f20f883e", RealName: "张强", IdCardNo: "110101198807189876", DriverLicenseNo: "DL10000003", AvatarUrl: "http://cdn.xxx/avatar/3.png"},
	}

	vehicles := []proto.CreateVehicleRequest{
		{DriverId: 0, PlateNo: "京A12345", Brand: "比亚迪", Model: "汉EV", Color: "白", VehicleType: 2, RegistrationDate: "2023-05-10", InsuranceNo: "INS20230001", InsuranceExpireAt: "2025-05-09"},
		{DriverId: 0, PlateNo: "沪B67890", Brand: "特斯拉", Model: "Model 3", Color: "黑", VehicleType: 2, RegistrationDate: "2022-11-20", InsuranceNo: "INS20220002", InsuranceExpireAt: "2024-11-19"},
		{DriverId: 0, PlateNo: "粤C13579", Brand: "小鹏", Model: "P7", Color: "银", VehicleType: 1, RegistrationDate: "2024-03-15", InsuranceNo: "INS20240003", InsuranceExpireAt: "2026-03-14"},
	}

	certs := []proto.CreateCertificationRequest{
		{DriverId: 0, VehicleId: 0, IdCardFrontUrl: "http://cdn.xxx/idcard/1_front.png", IdCardBackUrl: "http://cdn.xxx/idcard/1_back.png", DriverLicenseUrl: "http://cdn.xxx/dl/1.png", VehicleLicenseUrl: "http://cdn.xxx/vl/1.png"},
		{DriverId: 0, VehicleId: 0, IdCardFrontUrl: "http://cdn.xxx/idcard/2_front.png", IdCardBackUrl: "http://cdn.xxx/idcard/2_back.png", DriverLicenseUrl: "http://cdn.xxx/dl/2.png", VehicleLicenseUrl: "http://cdn.xxx/vl/2.png"},
		{DriverId: 0, VehicleId: 0, IdCardFrontUrl: "http://cdn.xxx/idcard/3_front.png", IdCardBackUrl: "http://cdn.xxx/idcard/3_back.png", DriverLicenseUrl: "http://cdn.xxx/dl/3.png", VehicleLicenseUrl: "http://cdn.xxx/vl/3.png"},
	}

	scores := []proto.CreateScoreRequest{
		{DriverId: 0, Score: 96.5, Level: 5, MonthOrders: 210, MonthCancelRate: 1.2, MonthComplaintCount: 0},
		{DriverId: 0, Score: 88.0, Level: 4, MonthOrders: 156, MonthCancelRate: 3.4, MonthComplaintCount: 2},
		{DriverId: 0, Score: 75.5, Level: 3, MonthOrders: 98, MonthCancelRate: 6.1, MonthComplaintCount: 5},
	}

	withdraws := []proto.CreateWithdrawRequest{
		{DriverId: 0, WithdrawNo: "WD20260801001", Amount: 800.0, PayeeName: "王伟", PayAccount: "6222021001000000001"},
		{DriverId: 0, WithdrawNo: "WD20260802002", Amount: 1500.0, PayeeName: "李娜", PayAccount: "6222021001000000002"},
		{DriverId: 0, WithdrawNo: "WD20260803003", Amount: 500.0, PayeeName: "张强", PayAccount: "6222021001000000003"},
	}

	// ---- 插入司机 ----
	driverIDs := make([]int64, 0, len(drivers))
	for i, d := range drivers {
		r, e := c.CreateDriver(ctx, &d)
		log(fmt.Sprintf("CreateDriver-%d", i+1), r, e)
		if r != nil {
			driverIDs = append(driverIDs, r.Id)
		}
	}

	// ---- 插入车辆（关联司机）----
	vehicleIDs := make([]int64, 0, len(vehicles))
	for i, v := range vehicles {
		v.DriverId = driverIDs[i]
		r, e := c.CreateVehicle(ctx, &v)
		log(fmt.Sprintf("CreateVehicle-%d", i+1), r, e)
		if r != nil {
			vehicleIDs = append(vehicleIDs, r.Id)
		}
	}

	// ---- 插入认证（关联司机+车辆）----
	certIDs := make([]int64, 0, len(certs))
	for i, ct := range certs {
		ct.DriverId = driverIDs[i]
		ct.VehicleId = vehicleIDs[i]
		r, e := c.CreateCertification(ctx, &ct)
		log(fmt.Sprintf("CreateCertification-%d", i+1), r, e)
		if r != nil {
			certIDs = append(certIDs, r.Id)
		}
		// 审核通过
		if r != nil {
			u, ue := c.UpdateCertification(ctx, &proto.UpdateCertificationRequest{
				Id: r.Id, AuditStatus: int32Ptr(2), AuditedBy: int64Ptr(100), AuditedAt: int64Ptr(time.Now().Unix()),
			})
			log(fmt.Sprintf("UpdateCertification-审核通过-%d", i+1), u, ue)
		}
	}

	// ---- 插入服务分（关联司机）----
	scoreIDs := make([]int64, 0, len(scores))
	for i, s := range scores {
		s.DriverId = driverIDs[i]
		r, e := c.CreateScore(ctx, &s)
		log(fmt.Sprintf("CreateScore-%d", i+1), r, e)
		if r != nil {
			scoreIDs = append(scoreIDs, r.Id)
		}
	}

	// ---- 插入提现（关联司机）----
	withdrawIDs := make([]int64, 0, len(withdraws))
	for i, w := range withdraws {
		w.DriverId = driverIDs[i]
		r, e := c.CreateWithdraw(ctx, &w)
		log(fmt.Sprintf("CreateWithdraw-%d", i+1), r, e)
		if r != nil {
			withdrawIDs = append(withdrawIDs, r.Id)
		}
		// 打款成功
		if r != nil {
			u, ue := c.UpdateWithdraw(ctx, &proto.UpdateWithdrawRequest{
				Id: r.Id, Status: int32Ptr(2), PaidAt: int64Ptr(time.Now().Unix()),
			})
			log(fmt.Sprintf("UpdateWithdraw-打款成功-%d", i+1), u, ue)
		}
	}

	// ---- 列表校验（确认数据已入库）----
	ld, lde := c.ListDrivers(ctx, &proto.ListDriversRequest{Page: 1, PageSize: 10})
	log("ListDrivers", ld, lde)
	lv, lve := c.ListVehicles(ctx, &proto.ListVehiclesRequest{DriverId: driverIDs[0]})
	log("ListVehicles", lv, lve)
	lc, lce := c.ListCertifications(ctx, &proto.ListCertificationsRequest{DriverId: driverIDs[0]})
	log("ListCertifications", lc, lce)
	ls, lse := c.ListScores(ctx, &proto.ListScoresRequest{DriverId: driverIDs[0]})
	log("ListScores", ls, lse)
	lw, lwe := c.ListWithdraws(ctx, &proto.ListWithdrawsRequest{DriverId: driverIDs[0]})
	log("ListWithdraws", lw, lwe)

	fmt.Println("\n示例数据已全部写入 driver 库，未删除，可直接在 MySQL 中核查。")
}

func strPtr(s string) *string      { return &s }
func int32Ptr(i int32) *int32      { return &i }
func int64Ptr(i int64) *int64      { return &i }
func float64Ptr(f float64) *float64 { return &f }
