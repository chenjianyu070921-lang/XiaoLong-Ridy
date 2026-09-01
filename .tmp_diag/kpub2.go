package main
import ("fmt";"github.com/segmentio/kafka-go";"context";"time")
func main() {
	w := &kafka.Writer{Addr: kafka.TCP("115.191.16.159:9092"), Topic: "order.created", Balancer: &kafka.LeastBytes{}}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err := w.WriteMessages(ctx, kafka.Message{Key: []byte("33"), Value: []byte(`{"order_id":33,"order_no":"DD2026083122252817","from_longitude":118.283231,"from_latitude":34.020904,"car_type":1,"city_code":"SH"}`)})
	if err != nil { fmt.Println("WRITE ERR:", err) } else { fmt.Println("????") }
}
