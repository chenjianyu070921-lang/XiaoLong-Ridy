package main
import ("fmt";"context";"github.com/redis/go-redis/v9")
func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr:"115.191.16.159:6379", Password:"4ay1nkal3u8ed77y", DB:0})
	rdb.SRem(ctx, "driver:online", "999", "8").Result()
	rdb.ZRem(ctx, "driver:geo:default", "999", "8").Result()
	rdb.Del(ctx, "driver:pos:8").Result()
	fmt.Println("????")
	on, _ := rdb.SMembers(ctx, "driver:online").Result()
	g, _ := rdb.ZRange(ctx, "driver:geo:default", 0, -1).Result()
	fmt.Println("online:", on, "geo:", g)
}
