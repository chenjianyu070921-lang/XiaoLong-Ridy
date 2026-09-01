package main
import ("fmt";"context";"time";"github.com/IBM/sarama")
type gh struct{}
func (gh) Setup(sarama.ConsumerGroupSession) error { fmt.Println("setup ok"); return nil }
func (gh) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (gh) ConsumeClaim(s sarama.ConsumerGroupSession, c sarama.ConsumerGroupClaim) error {
	for msg := range c.Messages() { fmt.Printf("got msg topic=%s val=%s\n", msg.Topic, string(msg.Value)); s.MarkMessage(msg,"") }
	return nil
}
func main() {
	cfg := sarama.NewConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Return.Errors = true
	cg, err := sarama.NewConsumerGroup([]string{"115.191.16.159:9092"}, "sarama-diag-group", cfg)
	if err != nil { fmt.Println("new cg err:", err); return }
	defer cg.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	err = cg.Consume(ctx, []string{"order.created"}, &gh{})
	if err != nil { fmt.Println("consume err:", err) } else { fmt.Println("consume returned nil") }
}
