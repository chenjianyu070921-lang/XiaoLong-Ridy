package main
import ("fmt";"context";"time";"github.com/IBM/sarama")
type gh struct{}
func (gh) Setup(sarama.ConsumerGroupSession) error { fmt.Println("SETUP ok"); return nil }
func (gh) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (gh) ConsumeClaim(s sarama.ConsumerGroupSession, c sarama.ConsumerGroupClaim) error {
	for msg := range c.Messages() { fmt.Printf("GOT topic=%s part=%d offset=%d val=%s\n", msg.Topic, msg.Partition, msg.Offset, string(msg.Value)); s.MarkMessage(msg, "") }
	return nil
}
func main() {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_1_0_0
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Return.Errors = true
	cfg.Consumer.Group.Rebalance.Timeout = 30 * time.Second
	cfg.Metadata.Retry.Max = 5
	cg, err := sarama.NewConsumerGroup([]string{"115.191.16.159:9092"}, "diag-grp-long", cfg)
	if err != nil { fmt.Println("new cg err:", err); return }
	defer cg.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()
	err = cg.Consume(ctx, []string{"order.created"}, &gh{})
	if err != nil { fmt.Println("consume err:", err) } else { fmt.Println("consume nil") }
}
