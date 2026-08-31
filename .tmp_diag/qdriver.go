package main
import ("database/sql";"fmt";_ "github.com/go-sql-driver/mysql")
func main() {
	db, _ := sql.Open("mysql", "root:4ay1nkal3u8ed77y@tcp(115.191.16.159:3306)/xiaolong_ridy?charset=utf8mb4")
	defer db.Close()
	// ??B12345 ?????
	rows, _ := db.Query("SELECT v.driver_id, v.plate_no, v.status FROM driver_vehicle v WHERE v.plate_no LIKE '%B12345%'")
	found := false
	for rows.Next() { var did int64; var pno string; var st int; rows.Scan(&did, &pno, &st); found = true; fmt.Printf("??: driver_id=%d plate=%s status=%d\n", did, pno, st);
		// ??????
		var ph, name, avatar string; var dstatus int
		db.QueryRow("SELECT phone, real_name, avatar_url, status FROM driver WHERE id=?", did).Scan(&ph, &name, &avatar, &dstatus)
		fmt.Printf("  ??: phone=%s name=%s avatar=[%s] status=%d\n", ph, name, avatar, dstatus)
		var audit int; var remark, front string
		err := db.QueryRow("SELECT audit_status, audit_remark, id_card_front_url FROM driver_certification WHERE driver_id=?", did).Scan(&audit, &remark, &front)
		if err != nil { fmt.Println("  ??: ???", err) } else { fmt.Printf("  ??: audit=%d remark=[%s] front=[%s]\n", audit, remark, front) }
	}
	if !found { fmt.Println("????B12345 ??") }
}
