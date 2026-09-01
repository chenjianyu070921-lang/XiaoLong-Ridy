package main
import ("database/sql";"fmt";_ "github.com/go-sql-driver/mysql")
func main() {
	db, _ := sql.Open("mysql", "root:4ay1nkal3u8ed77y@tcp(115.191.16.159:3306)/xiaolong_ridy?charset=utf8mb4")
	defer db.Close()
	show := func(label, query string) {
		rows, err := db.Query(query)
		if err != nil { fmt.Printf("[%s] err: %v\n", label, err); return }
		defer rows.Close()
		cols, _ := rows.Columns()
		for rows.Next() { vals := make([]any, len(cols)); ptrs := make([]any, len(cols)); for i := range vals { ptrs[i]=&vals[i] }; rows.Scan(ptrs...); for i := range vals { fmt.Printf("%v ", vals[i]) }; fmt.Println() }
	}
	fmt.Println("== ride_order ??(?15) ==")
	show("o", "SELECT id, order_no, user_id, driver_id, status, from_longitude, from_latitude, created_at FROM ride_order ORDER BY id DESC LIMIT 15")
	fmt.Println("== ????? ==")
	show("s", "SELECT status, COUNT(*) FROM ride_order GROUP BY status")
}
