package main
import ("fmt";"context";"github.com/redis/go-redis/v9")
func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr:"115.191.16.159:6379", Password:"4ay1nkal3u8ed77y", DB:0})
	on, _ := rdb.SMembers(ctx, "driver:online").Result()
	g, _ := rdb.ZRange(ctx, "driver:geo:default", 0, -1).Result()
	fmt.Println("driver:online:", on)
	fmt.Println("driver:geo:default:", g)
}
