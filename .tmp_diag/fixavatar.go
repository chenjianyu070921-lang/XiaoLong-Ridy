package main
import ("database/sql";"fmt";_ "github.com/go-sql-driver/mysql")
func main() {
	db, _ := sql.Open("mysql", "root:4ay1nkal3u8ed77y@tcp(115.191.16.159:3306)/xiaolong_ridy?charset=utf8mb4")
	defer db.Close()
	_, err := db.Exec("UPDATE driver SET avatar_url='' WHERE id=8")
	if err != nil { fmt.Println("????:", err) } else { fmt.Println("??8?????") }
}
