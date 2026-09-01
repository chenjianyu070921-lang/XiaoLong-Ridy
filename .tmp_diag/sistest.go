package main
import ("fmt";"context";"github.com/redis/go-redis/v9")
func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr:"115.191.16.159:6379", Password:"4ay1nkal3u8ed77y", DB:0})
	// ???????????????
	sism, err := rdb.SIsMember(ctx, "driver:online", "8").Result()
	fmt.Println("SIsMember driver:online 8 ->", sism, "err:", err)
	scard, _ := rdb.SCard(ctx, "driver:online").Result()
	fmt.Println("SCard driver:online ->", scard)
	smem, _ := rdb.SMembers(ctx, "driver:online").Result()
	fmt.Println("SMembers driver:online ->", smem)
	// OnlineStore ? hash
	hs, _ := rdb.HGetAll(ctx, "driver:online:8").Result()
	fmt.Println("HGetAll driver:online:8 ->", hs)
	// ???????
	rdb.SAdd(ctx, "driver:online", "777").Result()
	s2, _ := rdb.SIsMember(ctx, "driver:online", "777").Result()
	sm2, _ := rdb.SMembers(ctx, "driver:online").Result()
	fmt.Println("SAdd 777 ? SIsMember ->", s2, " SMembers ->", sm2)
}
