package main
import ("fmt";"github.com/segmentio/kafka-go";"context")
func main() {
	c, err := kafka.DialContext(context.Background(), "tcp", "115.191.16.159:9092")
	if err != nil { fmt.Println("dial err:", err); return }
	defer c.Close()
	parts, err := c.ReadPartitions("orderclient.created")
	if err != nil { fmt.Println("read partitions err:", err); return }
	fmt.Println("topic orderclient.created partitions:", len(parts))
}
