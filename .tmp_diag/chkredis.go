package main
import ("fmt";"context";"time";"github.com/redis/go-redis/v9")
func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr:"115.191.16.159:6379", Password:"4ay1nkal3u8ed77y", DB:0})
	m, _ := rdb.HGetAll(ctx, "driver:pos:8").Result()
	fmt.Println("driver:pos:8 hash:", m)
	fmt.Println("now unix:", time.Now().Unix())
	on, _ := rdb.SMembers(ctx, "driver:online").Result()
	fmt.Println("driver:online:", on)
	g, _ := rdb.ZRange(ctx, "driver:geo:default", 0, -1).Result()
	fmt.Println("driver:geo:default:", g)
	busy, _ := rdb.SMembers(ctx, "driver:busy").Result()
	fmt.Println("driver:busy:", busy)
}
