package main
import ("fmt";"time";"github.com/IBM/sarama")
func main() {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_1_0_0
	pc, err := sarama.NewConsumer([]string{"115.191.16.159:9092"}, cfg)
	if err != nil { fmt.Println("new consumer err:", err); return }
	defer pc.Close()
	parts, err := pc.Partitions("order.created")
	if err != nil { fmt.Println("partitions err:", err); return }
	fmt.Println("order.created partitions:", parts)
	for _, p := range parts {
		cp, err := pc.ConsumePartition("order.created", p, sarama.OffsetOldest)
		if err != nil { fmt.Printf("consume partition %d err: %v\n", p, err); continue }
		select {
		case m := <-cp.Messages():
			fmt.Printf("partition %d got: %s\n", p, string(m.Value))
		case <-time.After(5*time.Second):
			fmt.Printf("partition %d no msg in 5s\n", p)
		}
		cp.Close()
	}
}
