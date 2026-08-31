package main
import ("fmt";"context";"github.com/redis/go-redis/v9")
func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr:"115.191.16.159:6379", Password:"4ay1nkal3u8ed77y", DB:0})
	on, _ := rdb.SMembers(ctx, "driver:online").Result()
	g, _ := rdb.ZRange(ctx, "driver:geo:default", 0, -1).Result()
	pos, _ := rdb.HGetAll(ctx, "driver:pos:8").Result()
	av, _ := rdb.SMembers(ctx, "driver:available:8").Result()
	fmt.Println("online:", on)
	fmt.Println("geo:", g)
	fmt.Println("pos8:", pos)
	fmt.Println("available8:", av)
	// ? GEOSEARCH ???????? 3000m
	loc, _ := rdb.GeoSearchLocation(ctx, "driver:geo:default", &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{Longitude:118.283231, Latitude:34.020904, Radius:3000, RadiusUnit:"m", Sort:"ASC", Count:10}, WithDist:true}).Result()
	fmt.Println("GEOSEARCH 3000m:", loc)
}
