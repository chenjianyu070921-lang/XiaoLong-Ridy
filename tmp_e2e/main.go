package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"XiaoLong-Ridy/common/jwtx"

	_ "github.com/go-sql-driver/mysql"
)

const dsn = "root:4ay1nkal3u8ed77y@tcp(115.191.16.159:3306)/xiaolong_ridy?charset=utf8mb4&parseTime=true&loc=Local"
const chatBase = "http://127.0.0.1:18090/api/chat/v1"
const signKey = "driversvc-local-dev-key"

func main() {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("DB_OPEN_ERR", err)
		return
	}
	defer db.Close()

	// 1) 发现订单表，找一条活跃订单（status 2/3/4 且有司机）
	tables := []string{}
	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema='xiaolong_ridy' AND table_name LIKE '%order%'")
	if err != nil {
		fmt.Println("TABLES_ERR", err)
		return
	}
	for rows.Next() {
		var t string
		rows.Scan(&t)
		tables = append(tables, t)
	}
	rows.Close()
	fmt.Println("order tables:", tables)

	var orderID, driverID int64
	var status int
	found := false
	for _, t := range tables {
		if o, d, s, ok := findActiveOrder(db, t); ok {
			orderID, driverID, status, found = o, d, s, true
			fmt.Printf("FOUND order in %s: orderId=%d driverId=%d status=%d\n", t, orderID, driverID, status)
			break
		}
	}
	if !found {
		fmt.Println("NO_ACTIVE_ORDER")
		return
	}

	// 2) 签司机 JWT（driverId = 该订单的司机，满足 chatsvc 权限校验）
	tok, err := jwtx.SignAccountToken(jwtx.AccountTokenPayload{
		AccountID:     uint64(driverID),
		AccountType:   "driver",
		AccountStatus: 1,
		Phone:         "13800000000",
		Role:          "driver",
		Issuer:        "driversvc",
		TTL:           2 * time.Hour,
	}, signKey)
	if err != nil {
		fmt.Println("SIGN_ERR", err)
		return
	}
	fmt.Println("JWT signed, len=", len(tok))
	auth := "Bearer " + tok
	client := &http.Client{Timeout: 10 * time.Second}

	// 3) 会话
	convID := doGetConv(client, auth, orderID)
	if convID == 0 {
		return
	}
	fmt.Println(">>> conversationId =", convID)

	// 4) 发送文本（测敏感词 + 幂等之外的正常路径）
	doPost(client, chatBase+"/send", auth, map[string]interface{}{
		"conversationId": convID,
		"msgType":        1,
		"content":        "您好，我马上到上车点",
		"clientMsgId":    fmt.Sprintf("e2e-%d-1", time.Now().UnixNano()),
	})

	// 5) 发送敏感词（应被拦截 46000）
	doPost(client, chatBase+"/send", auth, map[string]interface{}{
		"conversationId": convID,
		"msgType":        1,
		"content":        "加我微信 xxx",
		"clientMsgId":    fmt.Sprintf("e2e-%d-2", time.Now().UnixNano()),
	})

	// 6) 列表
	doGet(client, chatBase+"/messages?conversationId="+strconv.FormatInt(convID, 10), auth)

	// 7) 标记已读
	doPost(client, chatBase+"/read", auth, map[string]interface{}{"conversationId": convID})
}

func findActiveOrder(db *sql.DB, t string) (orderID, driverID int64, status int, ok bool) {
	q := fmt.Sprintf("SELECT * FROM `%s` WHERE status IN (2,3,4) AND driver_id>0 LIMIT 1", t)
	rows, err := db.Query(q)
	if err != nil {
		return
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return
	}
	idxOrder, idxDriver, idxStatus := -1, -1, -1
	for i, c := range cols {
		cl := strings.ToLower(c)
		if cl == "order_id" || cl == "id" {
			idxOrder = i
		}
		if cl == "driver_id" {
			idxDriver = i
		}
		if cl == "status" {
			idxStatus = i
		}
	}
	if idxOrder < 0 || idxDriver < 0 || idxStatus < 0 {
		return
	}
	if !rows.Next() {
		return
	}
	vals := make([]sql.NullString, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	rows.Scan(ptrs...)
	orderID, _ = strconv.ParseInt(vals[idxOrder].String, 10, 64)
	driverID, _ = strconv.ParseInt(vals[idxDriver].String, 10, 64)
	status, _ = strconv.Atoi(vals[idxStatus].String)
	ok = true
	return
}

func doGet(client *http.Client, url, auth string) (convID int64) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", auth)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("GET_ERR", err)
		return
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	fmt.Println("GET", url, "->", resp.StatusCode, string(raw))
	var m map[string]interface{}
	json.Unmarshal(raw, &m)
	if d, ok := m["data"].(map[string]interface{}); ok {
		if v, ok := d["conversationId"].(float64); ok {
			convID = int64(v)
		}
	}
	return
}

func doGetConv(client *http.Client, auth string, orderID int64) int64 {
	return doGet(client, chatBase+"/conversation?orderId="+strconv.FormatInt(orderID, 10), auth)
}

func doPost(client *http.Client, url, auth string, body interface{}) {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, strings.NewReader(string(b)))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("POST_ERR", err)
		return
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	fmt.Println("POST", url, "->", resp.StatusCode, string(raw))
}
