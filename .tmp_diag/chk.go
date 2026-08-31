package main
import ("database/sql";"fmt";_ "github.com/go-sql-driver/mysql")
func main() {
	db, _ := sql.Open("mysql", "root:4ay1nkal3u8ed77y@tcp(115.191.16.159:3306)/xiaolong_ridy?charset=utf8mb4")
	defer db.Close()
	var id int; var phone, name, avatar string; var status, svc int
	db.QueryRow("SELECT id, phone, name, status, service_score, COALESCE(avatar_url,'') FROM driver WHERE id=8").Scan(&id,&phone,&name,&status,&svc,&avatar)
	fmt.Printf("??: id=%d phone=%s name=%s status=%d service_score=%d avatar='%s'\n", id, phone, name, status, svc, avatar)
	var cid int; var audit int
	db.QueryRow("SELECT id, audit_status FROM driver_certification WHERE driver_id=8").Scan(&cid,&audit)
	fmt.Printf("??: id=%d audit_status=%d\n", cid, audit)
	var vid int; var plate string; var vstatus int
	db.QueryRow("SELECT id, plate, status FROM driver_vehicle WHERE driver_id=8").Scan(&vid,&plate,&vstatus)
	fmt.Printf("??: id=%d plate=%s status=%d\n", vid, plate, vstatus)
}
