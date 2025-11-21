package hyperdrive

/*
Remote(controller):

It is a separate process.
The Anki service cans subscribe to any of its branches and will mirror it.
*/

import (
	"fmt"
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

// type LanePayload struct {
// 	Velocity         float32 `json:"velocity"`         // {0...1000} # Default: 0
// 	Acceleration     float32 `json:"acceleration"`     // {0...1000} # Default: 0
// 	Offset           float32 `json:"offset"`           // {-100.0...100.0} # Default: 0.0
// 	OffsetFromCenter float32 `json:"offsetFromCenter"` // {-100.0...100.0} # Default: 0.0
// }

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
		err1 := SyncSubscription(client, "connectSubscription", fmt.Sprintf(vehicleSubscriptionTopicFormat, vehicle), fmt.Sprintf(ConnectTopic, vehicle), true)
		err2 := SyncSubscription(client, "speedSubscription", fmt.Sprintf(vehicleSubscriptionTopicFormat, vehicle), fmt.Sprintf(SpeedTopic, vehicle), true)
		err3 := SyncSubscription(client, "lightsSubscription", fmt.Sprintf(vehicleSubscriptionTopicFormat, vehicle), fmt.Sprintf(LightsTopic, vehicle), true)
		err4 := SyncSubscription(client, "laneSubscription", fmt.Sprintf(vehicleSubscriptionTopicFormat, vehicle), fmt.Sprintf(LaneTopic, vehicle), true)
		err5 := SyncSubscription(client, "cancelLaneSubscription", fmt.Sprintf(vehicleSubscriptionTopicFormat, vehicle), fmt.Sprintf(CancelLaneTopic, vehicle), true)

		// Only add the vehicles to the list if all the subscriptions could be sent.
		if err1 == nil && err2 == nil && err3 == nil && err4 == nil && err5 == nil {
			vehicleList = append(vehicleList, vehicle)
		}
	}
	return vehicleList, nil
}
