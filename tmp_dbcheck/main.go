package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("USAGE: dbcheck <dsn>")
		os.Exit(2)
	}
	dsn := os.Args[1]
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("OPEN_ERR", err)
		os.Exit(1)
	}
	if err := db.Ping(); err != nil {
		fmt.Println("PING_ERR", err)
		os.Exit(1)
	}
	fmt.Println("PING_OK")

	data, err := os.ReadFile("scripts/sql/migrate/18_im_chat.sql")
	if err != nil {
		fmt.Println("READ_ERR", err)
		os.Exit(1)
	}
	for _, raw := range strings.Split(string(data), ";") {
		// 去掉每段开头的注释行，再判断是否还有实际语句
		var b strings.Builder
		for _, line := range strings.Split(raw, "\n") {
			t := strings.TrimSpace(line)
			if strings.HasPrefix(t, "--") || t == "" {
				continue
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
		s := strings.TrimSpace(b.String())
		if s == "" {
			continue
		}
		if _, err := db.Exec(s); err != nil {
			fmt.Println("EXEC_ERR", err, "\nSQL:", s)
			os.Exit(1)
		}
		fmt.Println("OK:", firstLine(s))
	}
	fmt.Println("ALL_DONE")
}

func firstLine(s string) string {
	for _, l := range strings.Split(s, "\n") {
		l = strings.TrimSpace(l)
		if l != "" {
			return l
		}
	}
	return ""
}
