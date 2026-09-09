package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "root:4ay1nkal3u8ed77y@tcp(115.191.16.159:3306)/xiaolong_ridy?charset=utf8mb4&parseTime=true&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("OPEN_ERR", err)
		return
	}
	defer db.Close()
	// 清理本次 E2E 在 orderId=4 会话(conversation_id=1)下写入的测试数据
	r1, err := db.Exec("DELETE FROM im_message WHERE conversation_id = 1")
	if err != nil {
		fmt.Println("DEL_MSG_ERR", err)
		return
	}
	n1, _ := r1.RowsAffected()
	r2, err := db.Exec("DELETE FROM im_conversation WHERE id = 1")
	if err != nil {
		fmt.Println("DEL_CONV_ERR", err)
		return
	}
	n2, _ := r2.RowsAffected()
	fmt.Printf("CLEANED im_message=%d rows, im_conversation=%d rows\n", n1, n2)
}
