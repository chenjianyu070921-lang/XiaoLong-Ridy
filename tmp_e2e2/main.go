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

const chatBase = "http://127.0.0.1:18090/api/chat/v1"

// 已由上轮验证可知 orderId=4 的订单：driverId=1, passengerId=128
const orderID int64 = 4

func signJWT(typ string, id uint64, key string) string {
	tok, err := jwtx.SignAccountToken(jwtx.AccountTokenPayload{
		AccountID:     id,
		AccountType:   typ,
		AccountStatus: 1,
		Phone:         "13800000000",
		Role:          typ,
		Issuer:        typ + "svc",
		TTL:           2 * time.Hour,
	}, key)
	if err != nil {
		panic(err)
	}
	return tok
}

func post(client *http.Client, url, auth string, body interface{}) (map[string]interface{}, int) {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, strings.NewReader(string(b)))
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("POST_ERR", err)
		return nil, -1
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	fmt.Println("POST", url, "->", resp.StatusCode, string(raw))
	var m map[string]interface{}
	json.Unmarshal(raw, &m)
	return m, resp.StatusCode
}

func get(client *http.Client, url, auth string) (map[string]interface{}, int) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", auth)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("GET_ERR", err)
		return nil, -1
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	fmt.Println("GET", url, "->", resp.StatusCode, string(raw))
	var m map[string]interface{}
	json.Unmarshal(raw, &m)
	return m, resp.StatusCode
}

func main() {
	client := &http.Client{Timeout: 10 * time.Second}

	// A) 乘客端直连（local-development-signing-key）：应能与同一订单建会话，peer=司机
	passTok := signJWT("passenger", 128, "local-development-signing-key")
	pconv, _ := get(client, chatBase+"/conversation?orderId="+strconv.FormatInt(orderID, 10), "Bearer "+passTok)
	cid := int64(pconv["data"].(map[string]interface{})["conversationId"].(float64))
	fmt.Printf("[A] 乘客会话 conversationId=%d (应为 1，同一订单复用)\n", cid)

	// A2) 乘客发消息
	post(client, chatBase+"/send", "Bearer "+passTok, map[string]interface{}{
		"conversationId": cid, "msgType": 1, "content": "师傅我到 B 口了",
		"clientMsgId": fmt.Sprintf("p-e2e-%d", time.Now().UnixNano()),
	})

	// B) 越权：用不属于该订单的司机(999)访问 → 应 40301 非订单参与方
	wrongTok := signJWT("driver", 999, "driversvc-local-dev-key")
	_, code := get(client, chatBase+"/conversation?orderId="+strconv.FormatInt(orderID, 10), "Bearer "+wrongTok)
	fmt.Printf("[B] 越权司机访问 -> HTTP %d (期望 403/40301)\n", code)

	// C) 幂等：司机用相同 clientMsgId 连发两次 → 应返回同一 messageId，不重复入库
	drvTok := signJWT("driver", 1, "driversvc-local-dev-key")
	same := fmt.Sprintf("idem-%d", time.Now().UnixNano())
	r1, _ := post(client, chatBase+"/send", "Bearer "+drvTok, map[string]interface{}{
		"conversationId": cid, "msgType": 1, "content": "幂等测试", "clientMsgId": same,
	})
	time.Sleep(200 * time.Millisecond)
	r2, _ := post(client, chatBase+"/send", "Bearer "+drvTok, map[string]interface{}{
		"conversationId": cid, "msgType": 1, "content": "幂等测试", "clientMsgId": same,
	})
	m1 := r1["data"].(map[string]interface{})["messageId"]
	m2 := r2["data"].(map[string]interface{})["messageId"]
	fmt.Printf("[C] 幂等：第一次 messageId=%v 第二次 messageId=%v (期望相同)\n", m1, m2)

	// D) 读回消息总数，确认没有重复行（应含 3 条：司机、乘客、司机幂等1次）
	msgs, _ := get(client, chatBase+"/messages?conversationId="+strconv.FormatInt(cid, 10)+"&limit=50", "Bearer "+drvTok)
	arr := msgs["data"].(map[string]interface{})["messages"].([]interface{})
	fmt.Printf("[D] 会话内消息数=%d (期望 3：司机/乘客/司机幂等)\n", len(arr))

	_ = sql.DB{}
}
