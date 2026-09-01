package main
import ("fmt";"context";"time";"net";"github.com/segmentio/kafka-go")
func main() {
	broker := "115.191.16.159:9092"
	topics := []string{"order.created","dispatch.new","order.paid","order.refunded","order.status.changed","order.canceled"}
	for _, t := range topics {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		conn, err := kafka.DialContext(ctx, "tcp", broker)
		if err != nil { fmt.Println("dial", t, err); cancel(); continue }
		ctl, err := conn.Controller()
		if err != nil { fmt.Println("controller", t, err); conn.Close(); cancel(); continue }
		ctlAddr := net.JoinHostPort(ctl.Host, fmt.Sprint(ctl.Port))
		ctlConn, err := kafka.DialContext(ctx, "tcp", ctlAddr)
		if err != nil { fmt.Println("dial ctl", t, err); conn.Close(); cancel(); continue }
		err = ctlConn.CreateTopics(kafka.TopicConfig{Topic: t, NumPartitions: 1, ReplicationFactor: 1})
		if err != nil { fmt.Println("create", t, "err:", err) } else { fmt.Println("create", t, "OK") }
		ctlConn.Close(); conn.Close(); cancel()
	}
}
