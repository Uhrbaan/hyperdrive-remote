package hyperdrive

/*
Remote(controller):

It is a separate process.
The Anki service cans subscribe to any of its branches and will mirror it.
*/

import (
	"fmt"
	"log"
	"strconv"

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

const (
	ankiVehicleSubscriptionTopic = "Anki/Vehicles/U/%s/I"
)

func StartRemote(address string, port int, uuid string) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(address + ":" + strconv.Itoa(port))
	opts.SetClientID(uuid)

	// Connect to the broker and initialize the client
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("Could not establish connection with MQTT server: ", token.Error())
	}
	log.Println("Connected to mosquitto broker.")

	// start By discovering available vehicles
	vehicleMap, err := discover(client)
	if err != nil {
		log.Fatal("Could not discover vehicles:", err)
	}
	log.Println(vehicleMap)

	var vehicleList []string
	for key := range vehicleMap {
		vehicleList = append(vehicleList, key)
	}

	// Initialize all the subscriptions because fuch it
	for _, vehicle := range vehicleList {
		SyncSubscription(client, "connectSubscription", fmt.Sprintf(ankiVehicleSubscriptionTopic, vehicle), fmt.Sprintf(ConnectTopic, vehicle), true)
		SyncSubscription(client, "speedSubscription", fmt.Sprintf(ankiVehicleSubscriptionTopic, vehicle), fmt.Sprintf(SpeedTopic, vehicle), true)
		SyncSubscription(client, "lightsSubscription", fmt.Sprintf(ankiVehicleSubscriptionTopic, vehicle), fmt.Sprintf(LightsTopic, vehicle), true)
	}

	App(client, vehicleList)
}
