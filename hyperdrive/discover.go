package hyperdrive

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	DiscoverTopic            = "RemoteControl/U/E/hosts/discover"
	ListenDiscoverTopicTopic = "RemoteControl/Hosts/I"
)

type DiscoverPayload struct {
	Value bool `json:"value"`
}

type DiscoverVehiclesTopic struct {
	Topic string `json:"topic"`
}

type Vehicle struct {
	ID    string
	Model string
}

type rawVehicleDataValue struct {
	Model string `json:"value"`
	Rssi  int    `json:"rssi"`
}
type rawVehicleData struct {
	Timestamp int                 `json:"timestamp"`
	Value     rawVehicleDataValue `json:"value"`
}

func Discover(client mqtt.Client, vehicleDiscoverTopic string) (map[string]Vehicle, error) {
	// Now, send a discovery signal
	vehicleList, err := discoverVehicles(client, vehicleDiscoverTopic, true)
	if err != nil {
		println("Could not discover vehicles:", err)
		return nil, err
	}
	return vehicleList, nil
}

func discoverVehicles(client mqtt.Client, vehicleListTopic string, value bool) (map[string]Vehicle, error) {
	// Send the "discover" payload
	payload, err := json.Marshal(DiscoverPayload{
		Value: value,
	})

	if err != nil {
		return nil, err
	}

	if token := client.Publish(DiscoverTopic, 1, false, payload); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	log.Println("Sent discovery on", DiscoverTopic)

	// now, create temporary subscription to look for vehicles.
	vehicleData := make(chan Vehicle, 1)
	client.Subscribe(vehicleListTopic, 1, func(client mqtt.Client, msg mqtt.Message) {
		topicBits := strings.Split(msg.Topic(), "/")
		id := strings.TrimSpace(topicBits[len(topicBits)-1])
		if err != nil {
			log.Println("Could not read the vehicle ID:", err)
			return
		}

		// currently doesn't read the vehicle model properly for some reason.
		var data [1]rawVehicleData
		err = json.Unmarshal(msg.Payload(), &data)
		if err != nil {
			log.Println("Could not ger raw vehicle data:", err)
			return
		}

		log.Println(data)

		vehicleData <- Vehicle{
			ID:    id,
			Model: data[0].Value.Model,
		}
	})
	defer client.Unsubscribe(vehicleListTopic) // we don't need it anymore after

	// var vehicleList map[string]string
	log.Println("Listing vehicles.")
	rawVehicleList := listUniqueVehicles(vehicleData)

	return rawVehicleList, nil
}

func listUniqueVehicles(vehicleData chan Vehicle) map[string]Vehicle {
	list := make(map[string]Vehicle)
	timeout := time.After(2 * time.Second)

	for {
		select {
		case <-timeout:
			log.Println("Timeout")
			return list
		case s := <-vehicleData:
			log.Println("Vehicle found:", s)
			list[s.ID] = s
		}
	}
}
