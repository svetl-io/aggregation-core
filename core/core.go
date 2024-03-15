package core

import (
	"aggregation-core/env"
	"aggregation-core/types"
)

func GetGreenLightDuration(message types.KafkaMessage) float32 {
	return float32(getTotalVehiclesCount(message)) * env.VEHICLE_TIME_COEFFICIENT
}

func getTotalVehiclesCount(message types.KafkaMessage) int {
	return message.CarsCount + message.TrucksCount + message.BusCount + message.MotorcyclesCount
}

func SendGreenLightDurationToWS(greenLightDuration float32) {
	return // TODO
}
