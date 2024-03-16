package main

import (
	"aggregation-core/core"
	"aggregation-core/env"
	"aggregation-core/logging"
	"aggregation-core/types"
	"encoding/json"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"net/http"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gammazero/workerpool"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var (
	sourceBroker             = env.SOURCE_KAFKA_BROKER
	topicsString             = env.TOPICS
	log                      = logging.GetLogger()
	poolSize                 = env.POOL_SIZE
	trafficLightEventChannel = make(chan types.TrafficLightEvent)
	upgrader                 = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	signals           = make(chan os.Signal, 1)
	trafficLightState = make(map[int]bool) // Key: traffic light ID, Value: true for green, false for red
	stateMutex        = sync.RWMutex{}
	pauseConsumption  bool
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("failed to upgrade connection", zap.Error(err))
		return
	}
	defer conn.Close()

	core.AddConnection(conn)
	defer core.RemoveConnection(conn)

	for {
		select {
		case <-signals:
			return // stop on interrupt
		default:
			trafficLightEvent := <-trafficLightEventChannel
			core.BroadcastMessage(trafficLightEvent)
		}
	}
}

func sendGreenLightState(trafficLightId int) {
	stateMutex.RLock()
	defer stateMutex.RUnlock()
	if !trafficLightState[trafficLightId] {
		trafficLightEventChannel <- core.GreenLightState(trafficLightId)
		trafficLightState[trafficLightId] = true
	}
}

func sendRedLightState(trafficLightId int) {
	stateMutex.RLock()
	defer stateMutex.RUnlock()
	trafficLightEventChannel <- core.RedLightState(trafficLightId)
	trafficLightState[trafficLightId] = false
}

func pauseConsumer() {
	pauseConsumption = true
}

func resumeConsumer() {
	pauseConsumption = false
}

func main() {
	defer func() { log.Info("bye") }()

	topics := strings.Split(topicsString, ",")

	http.HandleFunc("/ws", wsHandler)

	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Error("failed to start http server", zap.Error(err))
		}
		log.Info("http server started")
	}()

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": sourceBroker,
		"group.id":          "traffic-data",
		"auto.offset.reset": "latest",
	})

	if err != nil {
		log.Error("failed to create consumer", zap.Error(err))
		return
	}

	signal.Notify(signals, os.Interrupt)

	for _, topic := range topics {
		err := consumer.Subscribe(topic, nil)

		if err != nil {
			log.Error("failed to subscribe to topic", zap.String("topic", topic), zap.Error(err))
			return
		}
	}

	wp := workerpool.New(poolSize)

	running := true
	for running {
		select {
		case <-signals:
			running = false
		default:
			if pauseConsumption {
				time.Sleep(100 * time.Millisecond) // Sleep briefly to avoid busy-waiting
				continue
			}

			ev := consumer.Poll(100)

			switch e := ev.(type) {
			case *kafka.Message:
				var message types.KafkaMessage
				err := json.Unmarshal(e.Value, &message)

				if err != nil {
					log.Error("failed to unmarshal message", zap.Error(err))
					continue
				}

				log.Info("received message", zap.Any("message", message))

				wp.Submit(func() {
					greenLightDuration, redLightDuration := core.GetTrafficLightDuration(message)
					log.Debug("green light duration", zap.Float32("duration", greenLightDuration))
					log.Debug("red light duration", zap.Float32("duration", redLightDuration))

					sendGreenLightState(message.TrafficLightId)

					pauseConsumer()
					time.Sleep(time.Duration(int(greenLightDuration)) * time.Second)
					resumeConsumer()

					sendRedLightState(message.TrafficLightId)

					pauseConsumer()
					time.Sleep(time.Duration(int(redLightDuration)) * time.Second)
					resumeConsumer()
				})
			}
		}
	}
}
