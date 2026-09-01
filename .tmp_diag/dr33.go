package main
import ("database/sql";"fmt";_ "github.com/go-sql-driver/mysql")
func main() {
	db, _ := sql.Open("mysql", "root:4ay1nkal3u8ed77y@tcp(115.191.16.159:3306)/xiaolong_ridy?charset=utf8mb4")
	defer db.Close()
	rows, err := db.Query("SELECT id, order_id, driver_id, status, created_at FROM dispatch_record WHERE order_id=33")
	if err != nil { fmt.Println("err:", err); return }
	defer rows.Close()
	cols, _ := rows.Columns()
	for rows.Next() { vals := make([]any, len(cols)); ptrs := make([]any, len(cols)); for i := range vals { ptrs[i]=&vals[i] }; rows.Scan(ptrs...); for i := range vals { fmt.Printf("%v ", vals[i]) }; fmt.Println() }
}
