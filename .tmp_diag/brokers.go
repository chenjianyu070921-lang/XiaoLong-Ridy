package main
import ("fmt";"github.com/IBM/sarama")
func main() {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_1_0_0
	client, err := sarama.NewClient([]string{"115.191.16.159:9092"}, cfg)
	if err != nil { fmt.Println("client err:", err); return }
	defer client.Close()
	bs := client.Brokers()
	fmt.Printf("broker ??: %d\n", len(bs))
	for _, b := range bs {
		conn, _ := b.Connected()
		fmt.Printf("  broker id=%d addr=%s connected=%v\n", b.ID(), b.Addr(), conn)
	}
	// ? client ? group coordinator
	coord, err := client.Coordinator("diag-grp")
	if err != nil { fmt.Println("coordinator err:", err) } else { fmt.Printf("coordinator addr=%s id=%d\n", coord.Addr(), coord.ID()) }
}
