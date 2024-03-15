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

	"net/http"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gammazero/workerpool"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var (
	sourceBroker              = env.SOURCE_KAFKA_BROKER
	topicsString              = env.TOPICS
	log                       = logging.GetLogger()
	poolSize                  = env.POOL_SIZE
	greenLightDurationChannel = make(chan types.TrafficLightInfo)
	upgrader                  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	signals = make(chan os.Signal, 1)
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
			greenLightDuration := <-greenLightDurationChannel
			core.BroadcastMessage(greenLightDuration)
		}
	}
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
					greenLightDuration := core.GetGreenLightDuration(message)
					log.Debug("green light duration", zap.Any("duration", greenLightDuration))
					trafficLightInfo := types.TrafficLightInfo{
						TrafficLightID:     message.TrafficLightId,
						GreenLightDuration: greenLightDuration,
					}
					greenLightDurationChannel <- trafficLightInfo // traffic light info to ws channel
				})
			}
		}
	}
}
