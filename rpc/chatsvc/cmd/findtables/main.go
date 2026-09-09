package main

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(mysql.Open("root:4ay1nkal3u8ed77y@tcp(115.191.16.159:3306)/xiaolong_ridy?charset=utf8mb4"), &gorm.Config{})
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	var tables []string
	if err := db.Raw("SHOW TABLES").Scan(&tables).Error; err != nil {
		fmt.Println("err:", err)
		return
	}
	for _, t := range tables {
		fmt.Println(t)
	}
}