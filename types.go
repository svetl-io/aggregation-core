package main

type KafkaMessage struct {
	CarsCount        int    `json:"cars_count"`
	TrucksCount      int    `json:"trucks_count"`
	BusCount         int    `json:"bus_count"`
	MotorcyclesCount int    `json:"motorcycles_count"`
	DateTime         string `json:"datetime"`
}
