package main
import ("fmt";"github.com/IBM/sarama")
func main() {
	brokers := []string{"115.191.16.159:9092"}
	for _, v := range []sarama.KafkaVersion{
		sarama.V0_10_2_0, sarama.V2_1_0_0, sarama.V2_6_0_0, sarama.V3_6_0_0,
	} {
		cfg := sarama.NewConfig()
		cfg.Version = v
		client, err := sarama.NewClient(brokers, cfg)
		if err != nil { fmt.Printf("Version=%s err=%v\n", v, err); continue }
		topics, err := client.Topics()
		if err != nil { fmt.Printf("Version=%s Topics err=%v\n", v, err); client.Close(); continue }
		fmt.Printf("Version=%s OK topics=%v\n", v, topics)
		client.Close()
	}
}
