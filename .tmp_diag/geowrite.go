package main
import ("fmt";"context";"time";"github.com/redis/go-redis/v9")
func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr:"115.191.16.159:6379", Password:"4ay1nkal3u8ed77y", DB:0})
	// ? syncDispatchDriverOnline ???GeoAdd + SAdd + HSet + Expire
	_, err := rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.GeoAdd(ctx, "driver:geo:default", &redis.GeoLocation{Name:"8", Longitude:118.283231, Latitude:34.020904})
		pipe.SAdd(ctx, "driver:online", "8")
		pipe.HSet(ctx, "driver:pos:8", map[string]interface{}{"driver_id":"8","longitude":"118.283231","latitude":"34.020904","report_time":fmt.Sprint(time.Now().Unix())})
		pipe.Expire(ctx, "driver:pos:8", 2*time.Minute)
		return nil
	})
	fmt.Println("pipeline err:", err)
	time.Sleep(500*time.Millisecond)
	g, _ := rdb.ZRange(ctx, "driver:geo:default", 0, -1).Result()
	on, _ := rdb.SMembers(ctx, "driver:online").Result()
	p, _ := rdb.HGetAll(ctx, "driver:pos:8").Result()
	fmt.Println("geo:", g, "online:", on, "pos:", p)
}
