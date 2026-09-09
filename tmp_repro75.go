package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"XiaoLong-Ridy/common/jwtx"
)

func main() {
	tok, _ := jwtx.SignAccountToken(jwtx.AccountTokenPayload{
		AccountID: 999, AccountType: "driver", AccountStatus: 1,
		Phone: "13900000008", Role: "driver", Issuer: "driversvc", TTL: time.Hour,
	}, "driversvc-local-dev-key")
	u := "http://127.0.0.1:18090/api/chat/v1/conversation?" + url.Values{"orderId": {"75"}}.Encode()
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("DO_ERR", err)
		return
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	fmt.Println("HTTP", resp.StatusCode)
	fmt.Println("BODY", string(b))
	var m map[string]any
	if json.Unmarshal(b, &m) == nil {
		fmt.Printf("PARSED code=%v message=%q\n", m["code"], m["message"])
	}
}