package main
import ("fmt";"github.com/IBM/sarama")
func main() {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_1_0_0
	c, err := sarama.NewClient([]string{"115.191.16.159:9092"}, cfg)
	if err != nil { fmt.Println("client err:", err); return }
	defer c.Close()
	b := c.Brokers()[0]
	if err := b.Open(cfg); err != nil { fmt.Println("open err:", err); return }
	av, err := b.ApiVersions(nil)
	if err != nil { fmt.Println("apiVersions err:", err); return }
	apivs := av.ApiVersions
	fmt.Println("broker ApiVersions ??:", len(apivs))
	for _, a := range apivs {
		switch a.ApiKey { case 0: fmt.Printf("Produce max=%d\n", a.MaxVersion); case 8: fmt.Printf("OffsetFetch max=%d\n", a.MaxVersion); case 11: fmt.Printf("OffsetCommit max=%d\n", a.MaxVersion); case 15: fmt.Printf("FindCoordinator max=%d\n", a.MaxVersion); case 19: fmt.Printf("JoinGroup max=%d\n", a.MaxVersion); case 18: fmt.Printf("SyncGroup max=%d\n", a.MaxVersion) }
	}
}
