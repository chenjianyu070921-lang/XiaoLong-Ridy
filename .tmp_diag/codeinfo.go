package main
import ("fmt";"github.com/redis/go-redis/v9";"context")
func main() {
	rdb := redis.NewClient(&redis.Options{Addr: "115.191.16.159:6379", Password: "4ay1nkal3u8ed77y"})
	ttl, _ := rdb.TTL(context.Background(), "driver:sms:code:19397622796").Result()
	code, _ := rdb.Get(context.Background(), "driver:sms:code:19397622796").Result()
	fmt.Println("code=", code, "ttl=", ttl)
}
