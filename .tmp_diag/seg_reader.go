package main
import ("fmt";"context";"time";"github.com/segmentio/kafka-go")
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"115.191.16.159:9092"},
		GroupID: "diag-seg-group2",
		Topic: "order.created",
		StartOffset: kafka.FirstOffset,
	})
	defer r.Close()
	for i := 0; i < 3; i++ {
		msg, err := r.FetchMessage(ctx)
		if err != nil { fmt.Println("fetch err:", err); return }
		fmt.Printf("GOT part=%d offset=%d val=%s\n", msg.Partition, msg.Offset, string(msg.Value))
		r.CommitMessages(ctx, msg)
	}
}
