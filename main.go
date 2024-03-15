package main

import (
	"aggregation-core/env"
	"aggregation-core/logging"
	"encoding/json"
	"os"
	"os/signal"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
)

var (
	sourceBroker = env.SOURCE_KAFKA_BROKER
	topicsString = env.TOPICS
	log          = logging.GetLogger()
)

func main() {
	defer func() { log.Info("bye") }()

	topics := strings.Split(topicsString, ",")

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": sourceBroker,
		"group.id":          "traffic-data",
		"auto.offset.reset": "latest",
	})

	if err != nil {
		log.Error("failed to create consumer", zap.Error(err))
		return
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for _, topic := range topics {
		err := consumer.Subscribe(topic, nil)

		if err != nil {
			log.Error("failed to subscribe to topic", zap.String("topic", topic), zap.Error(err))
			return
		}
	}

	running := true
	for running {
		select {
		case <-signals:
			running = false
		default:
			ev := consumer.Poll(100)

			switch e := ev.(type) {
			case *kafka.Message:
				var message KafkaMessage
				err := json.Unmarshal(e.Value, &message)

				if err != nil {
					log.Error("failed to unmarshal message", zap.Error(err))
					continue
				}

				log.Info("received message", zap.Any("message", message))
			}
		}
	}
}
