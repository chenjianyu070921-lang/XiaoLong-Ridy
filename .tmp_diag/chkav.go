package main
import ("fmt";"github.com/redis/go-redis/v9";"context")
func main() {
	rdb := redis.NewClient(&redis.Options{Addr: "115.191.16.159:6379", Password: "4ay1nkal3u8ed77y"})
	m, _ := rdb.SMembers(context.Background(), "driver:available:8").Result()
	online, _ := rdb.SIsMember(context.Background(), "driver:online", "8").Result()
	fmt.Println("available:", m, "online:", online)
}
