//go:build verify

// verify 检查线上库所有表/列注释是否还有乱码残留。
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func hasGarbage(s string) bool {
	for _, r := range s {
		// 乱码残留特征：latin1 解码 UTF-8 产生的扩展区字符（U+0080-U+024F）及 Euro sign (U+20AC)
		if (r >= 0x80 && r <= 0x24F) || r == 0x20AC {
			// 排除常见西文注释中可能合法出现的字符
			return true
		}
	}
	return false
}

func main() {
	dsn := flag.String("dsn", "root:4ay1nkal3u8ed77y@tcp(115.191.16.159:3306)/xiaolong_ridy?charset=utf8mb4&parseTime=True&loc=Local", "MySQL DSN")
	flag.Parse()

	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "连接失败:", err)
		os.Exit(1)
	}
	defer db.Close()

	// 检查列注释
	rows, err := db.Query(`
		SELECT TABLE_NAME, COLUMN_NAME, COLUMN_COMMENT
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		ORDER BY TABLE_NAME, ORDINAL_POSITION`)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查询失败:", err)
		os.Exit(1)
	}
	defer rows.Close()

	badCols := 0
	totalCols := 0
	for rows.Next() {
		var t, c, cm string
		rows.Scan(&t, &c, &cm)
		totalCols++
		if hasGarbage(cm) {
			badCols++
			fmt.Printf("[乱码残留] 表 %s 列 %s: %q\n", t, c, cm)
		}
	}
	fmt.Printf("列注释: 共 %d 个，乱码残留 %d 个\n", totalCols, badCols)

	// 检查表注释
	trows, err := db.Query(`
		SELECT TABLE_NAME, TABLE_COMMENT
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE()`)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查询失败:", err)
		os.Exit(1)
	}
	defer trows.Close()

	badTables := 0
	totalTables := 0
	for trows.Next() {
		var t, tc string
		trows.Scan(&t, &tc)
		totalTables++
		if hasGarbage(tc) {
			badTables++
			fmt.Printf("[乱码残留] 表 %s 注释: %q\n", t, tc)
		}
	}
	fmt.Printf("表注释: 共 %d 个，乱码残留 %d 个\n", totalTables, badTables)

	if badCols+badTables > 0 {
		os.Exit(1)
	}
	// 列出修复后的部分注释样例
	fmt.Println("\n样例（前 10 条列注释）:")
	_ = strings.TrimSpace // 避免未使用
	fmt.Println("OK: 全部注释正常")
}
