package core

import (
	"aggregation-core/env"
	"aggregation-core/logging"
	"aggregation-core/types"
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var (
	log           = logging.GetLogger()
	connections   = make([]*websocket.Conn, 0)
	connectionsMu sync.Mutex
)

func GetGreenLightDuration(message types.KafkaMessage) float32 {
	return float32(getTotalVehiclesCount(message)) * env.VEHICLE_TIME_COEFFICIENT
}

func getTotalVehiclesCount(message types.KafkaMessage) int {
	return message.CarsCount + message.TrucksCount + message.BusCount + message.MotorcyclesCount
}

func AddConnection(conn *websocket.Conn) {
	connectionsMu.Lock()
	defer connectionsMu.Unlock()
	connections = append(connections, conn)
}

func RemoveConnection(conn *websocket.Conn) {
	connectionsMu.Lock()
	defer connectionsMu.Unlock()
	for i, c := range connections {
		if c == conn {
			connections = append(connections[:i], connections[i+1:]...)
			break
		}
	}
}

func BroadcastMessage(message interface{}) {
	var wg sync.WaitGroup
	connectionsMu.Lock()
	defer connectionsMu.Unlock()
	for _, conn := range connections {
		wg.Add(1)
		go func(conn *websocket.Conn) {
			defer wg.Done()
			if err := conn.WriteJSON(message); err != nil {
				log.Error("failed to write message", zap.Error(err))
			}
		}(conn)
	}
	wg.Wait()
}
