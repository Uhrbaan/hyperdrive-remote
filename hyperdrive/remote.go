package hyperdrive

/*
Remote(controller):

It is a separate process.
The Anki service cans subscribe to any of its branches and will mirror it.
*/

import (
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Remote struct {
	Client mqtt.Client
}

// Data definitions of hyperdrive objects

type Intent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type LanePayload struct {
	Velocity         float32 `json:"velocity"`         // {0...1000} # Default: 0
	Acceleration     float32 `json:"acceleration"`     // {0...1000} # Default: 0
	Offset           float32 `json:"offset"`           // {-100.0...100.0} # Default: 0.0
	OffsetFromCenter float32 `json:"offsetFromCenter"` // {-100.0...100.0} # Default: 0.0
}

type SubscriptionPayload struct {
	Topic     string `json:"topic"`     // {topic-filter} # Default: null
	Subscribe bool   `json:"subscribe"` // {true|false} # Default: false
}

func InitializeRemote(client mqtt.Client, vehicleDiscoverTopic string, vehicleSubscriptionTopicFormat string) ([]string, error) {
	// start By discovering available vehicles
	vehicleMap, err := Discover(client, vehicleDiscoverTopic)
	if err != nil {
		return nil, err
	}
	log.Println(vehicleMap)

	var vehicleList []string
	for vehicle := range vehicleMap {
		vehicleList = append(vehicleList, vehicle)
	}
	return vehicleList, nil
}
