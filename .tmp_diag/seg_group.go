package main
import ("fmt";"context";"time";"github.com/segmentio/kafka-go")
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	g := kafka.NewReaderGroup(kafka.ReaderGroupConfig{
		Brokers: []string{"115.191.16.159:9092"},
		GroupID: "diag-seg-group",
		Topics:  []string{"order.created"},
		CommitInterval: time.Second,
		StartOffset: kafka.FirstOffset,
		SessionTimeout: 30*time.Second,
		HeartbeatInterval: 5*time.Second,
	})
	gen, err := g.Next(ctx)
	if err != nil { fmt.Println("Next err:", err); return }
	defer gen.Close()
	for i := 0; i < 3; i++ {
		msg, err := gen.Fetch(ctx, 2*time.Second)
		if err != nil { fmt.Println("fetch err:", err); break }
		fmt.Printf("GOT part=%d offset=%d val=%s\n", msg.Partition, msg.Offset, string(msg.Value))
	}
}
