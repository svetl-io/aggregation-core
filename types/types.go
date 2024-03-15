package types

type KafkaMessage struct {
	CarsCount        int    `json:"cars_count"`
	TrucksCount      int    `json:"trucks_count"`
	BusCount         int    `json:"bus_count"`
	MotorcyclesCount int    `json:"motorcycles_count"`
	DateTime         string `json:"datetime"`
	TrafficLightId   int    `json:"traffic_light_id"`
}

type TrafficLightInfo struct {
	TrafficLightID     int     `json:"traffic_light_id"`
	GreenLightDuration float32 `json:"green_light_duration"`
}
