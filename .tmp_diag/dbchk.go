package main
import ("database/sql";"fmt";_ "github.com/go-sql-driver/mysql")
func main() {
	db, _ := sql.Open("mysql", "root:4ay1nkal3u8ed77y@tcp(115.191.16.159:3306)/xiaolong_ridy?charset=utf8mb4")
	defer db.Close()
	q := func(label, query string) {
		rows, err := db.Query(query)
		if err != nil { fmt.Printf("[%s] err: %v\n", label, err); return }
		defer rows.Close()
		cols, _ := rows.Columns()
		for rows.Next() {
			vals := make([]any, len(cols)); ptrs := make([]any, len(cols))
			for i := range vals { ptrs[i] = &vals[i] }
			rows.Scan(ptrs...)
			for i := range vals { fmt.Printf("%v ", vals[i]) }
			fmt.Println()
		}
	}
	fmt.Println("[driver ??8] id phone real_name status avatar_url")
	q("d8", "SELECT id, phone, real_name, status, COALESCE(avatar_url,'') FROM driver WHERE id=8 OR phone='19397622796'")
	fmt.Println("[driver ??]")
	q("all", "SELECT id, phone, real_name, status FROM driver ORDER BY id")
	fmt.Println("[vehicle ??]")
	q("v", "SELECT id, driver_id, plate_no, status FROM driver_vehicle ORDER BY id")
	fmt.Println("[cert ??]")
	q("c", "SELECT id, driver_id, vehicle_id, audit_status FROM driver_certification ORDER BY id")
}
