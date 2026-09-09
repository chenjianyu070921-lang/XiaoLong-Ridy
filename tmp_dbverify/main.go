package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "root:4ay1nkal3u8ed77y@tcp(115.191.16.159:3306)/xiaolong_ridy?charset=utf8mb4&parseTime=true&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("OPEN_ERR", err)
		os.Exit(1)
	}
	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema='xiaolong_ridy' AND table_name IN ('im_conversation','im_message')")
	if err != nil {
		fmt.Println("QUERY_ERR", err)
		os.Exit(1)
	}
	defer rows.Close()
	fmt.Println("EXISTS:")
	for rows.Next() {
		var n string
		rows.Scan(&n)
		fmt.Println(" -", n)
	}
}
