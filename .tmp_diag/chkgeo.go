package main
import ("fmt";"github.com/redis/go-redis/v9";"context")
func main() {
	rdb := redis.NewClient(&redis.Options{Addr: "115.191.16.159:6379", Password: "4ay1nkal3u8ed77y"})
	m, _ := rdb.SIsMember(context.Background(), "driver:online", "8").Result()
	geo, _ := rdb.GeoPos(context.Background(), "driver:geo", "8").Result()
	fmt.Println("driver:online ?8:", m, "| GEO:", geo)
}
