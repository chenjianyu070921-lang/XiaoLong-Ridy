package ws

import (
	"fmt"
	"net/http"
	"time"

	"XiaoLong-Ridy/common/constants"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// LocationWSHandler 实时位置推送：订阅某司机位置，周期从 Redis GEO 读取并推送给 WebSocket 客户端。
// 这样前端（乘客端/后台）可实时看到司机位置变化，补齐“实时推送”能力。
func LocationWSHandler(rdb *redis.Redis) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		driverID := r.URL.Query().Get("driverId")
		if driverID == "" {
			http.Error(w, "driverId required", http.StatusBadRequest)
			return
		}
		city := r.URL.Query().Get("city")
		geoKey := constants.DriverGeoKeyOf(city)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				pos, err := rdb.GeoPos(geoKey, driverID)
				if err != nil || len(pos) == 0 {
					continue
				}
				p := pos[0]
				msg := fmt.Sprintf(`{"driverId":%q,"city":%q,"lng":%v,"lat":%v}`, driverID, city, p.Longitude, p.Latitude)
				if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
					return
				}
			}
		}
	}
}
