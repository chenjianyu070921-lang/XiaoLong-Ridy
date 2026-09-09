package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "root:4ay1nkal3u8ed77y@tcp(115.191.16.159:3306)/xiaolong_ridy?charset=utf8mb4&parseTime=true&loc=Local"
	db, _ := sql.Open("mysql", dsn)
	defer db.Close()

	// 列名
	rows, err := db.Query("SELECT COLUMN_NAME FROM information_schema.columns WHERE table_schema='xiaolong_ridy' AND table_name='ride_order' ORDER BY ORDINAL_POSITION")
	if err != nil {
		fmt.Println("SCHEMA_ERR", err)
		return
	}
	defer rows.Close()
	fmt.Println("ride_order columns:")
	for rows.Next() {
		var c string
		rows.Scan(&c)
		fmt.Println("  ", c)
	}

	// 按常见 id 列名尝试 orderId=75
	for _, col := range []string{"orderId", "order_id", "id"} {
		q := fmt.Sprintf("SELECT * FROM ride_order WHERE %s = 75 LIMIT 1", col)
		r, err := db.Query(q)
		if err != nil {
			fmt.Printf("try %s: ERR %v\n", col, err)
			continue
		}
		cols, _ := r.Columns()
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		found := false
		for r.Next() {
			r.Scan(ptrs...)
			found = true
			fmt.Printf("FOUND by %s=75:\n", col)
			for i, c := range cols {
				fmt.Printf("    %s = %v\n", c, vals[i])
			}
		}
		r.Close()
		if found {
			break
		}
		fmt.Printf("try %s=75: not found\n", col)
	}
}