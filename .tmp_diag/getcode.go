package main
import ("fmt";"os";"github.com/redis/go-redis/v9";"context")
func main() {
	rdb := redis.NewClient(&redis.Options{Addr: "115.191.16.159:6379", Password: "4ay1nkal3u8ed77y"})
	code, err := rdb.Get(context.Background(), "driver:sms:code:19397622796").Result()
	if err != nil { fmt.Println("ERR", err); os.Exit(1) }
	fmt.Print(code)
}
