package main
import ("fmt";"context";"github.com/redis/go-redis/v9")
func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr:"115.191.16.159:6379", Password:"4ay1nkal3u8ed77y", DB:0})
	// 1. ????
	for _, k := range []string{"driver:online","driver:geo:default","driver:pos:8","driver:online:8"} {
		t, err := rdb.Type(ctx, k).Result()
		fmt.Printf("TYPE %s = %s (err=%v)\n", k, t, err)
	}
	// 2. ?????????????? key?
	n1, e1 := rdb.SAdd(ctx, "diag:set:test", "x").Result()
	fmt.Printf("SADD diag:set:test -> %d err=%v\n", n1, e1)
	n2, e2 := rdb.GeoAdd(ctx, "diag:geo:test", &redis.GeoLocation{Name:"1", Longitude:118, Latitude:34}).Result()
	fmt.Printf("GEOADD diag:geo:test -> %d err=%v\n", n2, e2)
	n3, e3 := rdb.HSet(ctx, "diag:hash:test", "a", "b").Result()
	fmt.Printf("HSET diag:hash:test -> %v err=%v\n", n3, e3)
	// 3. ?????? driver:online / driver:geo:default
	n4, e4 := rdb.SAdd(ctx, "driver:online", "999").Result()
	fmt.Printf("SADD driver:online 999 -> %d err=%v\n", n4, e4)
	n5, e5 := rdb.GeoAdd(ctx, "driver:geo:default", &redis.GeoLocation{Name:"999", Longitude:118, Latitude:34}).Result()
	fmt.Printf("GEOADD driver:geo:default 999 -> %d err=%v\n", n5, e5)
}
