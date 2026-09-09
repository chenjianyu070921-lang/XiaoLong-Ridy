package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Order struct {
	ID          int64
	DriverID    int64
	PassengerID int64
	Status      int
}

// 手签 HS256（与 go-zero rest/jwt.NewJwtToken 签出的 token 等价，
// apichat 的 JwtAuth 中间件用同密钥 HMAC-SHA256 校验，能通过）。
func signHS256(secret string, payload map[string]any) string {
	hb, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	pb, _ := json.Marshal(payload)
	h64 := base64.RawURLEncoding.EncodeToString(hb)
	p64 := base64.RawURLEncoding.EncodeToString(pb)
	signing := h64 + "." + p64
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signing))
	return signing + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func main() {
	dsn := "root:4ay1nkal3u8ed77y@tcp(115.191.16.159:3306)/xiaolong_ridy?charset=utf8mb4&parseTime=true"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("db open err:", err)
		return
	}
	var cols []struct {
		Field string
		Type  string
	}
	if err := db.Raw("SHOW COLUMNS FROM ride_order").Scan(&cols).Error; err != nil {
		fmt.Println("cols err:", err)
		return
	}
	fmt.Println("=== ride_order columns ===")
	for _, c := range cols {
		fmt.Printf("  %-28s %s\n", c.Field, c.Type)
	}

	var rows []Order
	if err := db.Raw("SELECT id, driver_id, status FROM ride_order WHERE id IN (4,75) ORDER BY id").Scan(&rows).Error; err != nil {
		fmt.Println("db query err:", err)
		return
	}
	fmt.Println("=== orders (MySQL) ===")
	for _, o := range rows {
		fmt.Printf("  id=%d  driver_id=%d  passenger_id=%d  status=%d\n", o.ID, o.DriverID, o.PassengerID, o.Status)
	}
	real := map[int64]int64{}
	for _, o := range rows {
		real[o.ID] = o.DriverID
	}

	sign := func(driverID int64) string {
		return signHS256("local-development-signing-key", map[string]any{
			"sub":           "driver_" + strconv.FormatInt(driverID, 10),
			"accountId":     uint64(driverID),
			"accountType":   "driver",
			"accountStatus": 1,
			"iss":           "chatdiag",
			"iat":           time.Now().Unix(),
			"exp":           time.Now().Add(2 * time.Hour).Unix(),
		})
	}

	call := func(label string, token string, orderID int64) {
		url := fmt.Sprintf("http://127.0.0.1:18090/api/chat/v1/conversation?orderId=%d", orderID)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("\n=== %s ===\n  http err: %v\n", label, err)
			return
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		var pretty interface{}
		_ = json.Unmarshal(b, &pretty)
		pb, _ := json.MarshalIndent(pretty, "  ", "  ")
		fmt.Printf("\n=== %s ===\n  HTTP %d\n  body: %s\n", label, resp.StatusCode, string(pb))
	}

	call("[A] order=75 caller_id=999 (司机端实际场景)", sign(999), 75)
	if d, ok := real[75]; ok {
		call(fmt.Sprintf("[B] order=75 caller_id=%d (订单真实司机)", d), sign(d), 75)
	}
	if d, ok := real[4]; ok {
		call(fmt.Sprintf("[C] order=4  caller_id=%d (订单真实司机)", d), sign(d), 4)
	}
	call("[D] order=75 caller_id=0 (越权边界)", sign(0), 75)
}