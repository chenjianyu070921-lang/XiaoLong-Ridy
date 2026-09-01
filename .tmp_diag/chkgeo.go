package main
import ("fmt";"context";"github.com/redis/go-redis/v9")
func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr:"115.191.16.159:6379", Password:"4ay1nkal3u8ed77y", DB:0})
	keys, err := rdb.Keys(ctx, "driver:geo*").Result()
	if err != nil { fmt.Println("keys err:", err); return }
	fmt.Println("driver:geo* keys:", keys)
	for _, k := range keys {
		members, err := rdb.ZRange(ctx, k, 0, -1).Result()
		if err != nil { fmt.Println("zrange", k, err); continue }
		fmt.Printf("  %s members=%v\n", k, members)
	}
	// ????
	online, _ := rdb.SMembers(ctx, "driver:online").Result()
	fmt.Println("driver:online:", online)
	// ??8 pos
	pos, err := rdb.Get(ctx, "driver:pos:8").Result()
	fmt.Println("driver:pos:8:", pos, err)
}
