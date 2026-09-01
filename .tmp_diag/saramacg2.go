package main
import ("fmt";"context";"time";"github.com/IBM/sarama")
type gh struct{}
func (gh) Setup(sarama.ConsumerGroupSession) error { return nil }
func (gh) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (gh) ConsumeClaim(s sarama.ConsumerGroupSession, c sarama.ConsumerGroupClaim) error { return nil }
func main() {
	for _, v := range []sarama.KafkaVersion{sarama.V0_10_2_0, sarama.V0_11_0_0, sarama.V2_0_0_0, sarama.V2_8_0_0} {
		cfg := sarama.NewConfig()
		cfg.Version = v
		cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
		cfg.Consumer.Return.Errors = true
		cg, err := sarama.NewConsumerGroup([]string{"115.191.16.159:9092"}, "sarama-diag-v", cfg)
		if err != nil { fmt.Printf("V=%s new err=%v\n", v, err); continue }
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		err = cg.Consume(ctx, []string{"order.created"}, &gh{})
		cancel()
		cg.Close()
		if err != nil { fmt.Printf("V=%s consume err=%v\n", v, err) } else { fmt.Printf("V=%s consume nil (OK)\n", v) }
	}
}
