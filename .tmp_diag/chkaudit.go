package main
import ("database/sql";"fmt";_ "github.com/go-sql-driver/mysql")
func main() {
	db, _ := sql.Open("mysql", "root:4ay1nkal3u8ed77y@tcp(115.191.16.159:3306)/xiaolong_ridy?charset=utf8mb4")
	defer db.Close()
	var audit, ds int
	db.QueryRow("SELECT audit_status FROM driver_certification WHERE driver_id=8").Scan(&audit)
	db.QueryRow("SELECT status FROM driver WHERE id=8").Scan(&ds)
	fmt.Printf("?? audit_status=%d (2=??) | ?? status=%d (2=??)\n", audit, ds)
}
