package main
import ("database/sql";"fmt";_ "github.com/go-sql-driver/mysql")
func main() {
	db, _ := sql.Open("mysql", "root:4ay1nkal3u8ed77y@tcp(115.191.16.159:3306)/xiaolong_ridy?charset=utf8mb4")
	defer db.Close()
	for _, t := range []string{"driver","driver_vehicle","driver_certification"} {
		fmt.Printf("== %s ==", t)
		rows, _ := db.Query("SHOW COLUMNS FROM "+t)
		for rows.Next() { var f, ty, nl, k, d, ex string; rows.Scan(&f,&ty,&nl,&k,&d,&ex); fmt.Printf(" %s(%s)", f, ty) }
		fmt.Println()
	}
}
