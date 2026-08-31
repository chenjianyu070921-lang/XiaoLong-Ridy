package main
import ("database/sql";"fmt";_ "github.com/go-sql-driver/mysql")
func main() {
	db, _ := sql.Open("mysql", "root:4ay1nkal3u8ed77y@tcp(115.191.16.159:3306)/xiaolong_ridy?charset=utf8mb4")
	defer db.Close()
	db.Exec("UPDATE driver_certification SET audit_status=1 WHERE driver_id=8")
	var st int; db.QueryRow("SELECT audit_status FROM driver_certification WHERE driver_id=8").Scan(&st)
	fmt.Println("?? audit_status =", st, "(1=???)")
}
