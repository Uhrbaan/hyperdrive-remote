package util

import (
	"encoding/json"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	RootTopic      = "/hobHq10yb9dKwxrdfhtT"
	VehicleIDTopic = "/hobHq10yb9dKwxrdfhtT/vehicle/id"
)

type VehicleIdPayload struct {
	ID string `json:"id"`
}

func WaitForVehicleID(client mqtt.Client) string {
	ch := make(chan string)
	client.Subscribe(VehicleIDTopic, 1, func(c mqtt.Client, m mqtt.Message) {
		var data VehicleIdPayload
		if err := json.Unmarshal(m.Payload(), &data); err == nil {
			ch <- data.ID
		}
	})

	defer client.Unsubscribe(VehicleIDTopic)
	return <-ch
}

func SendJSON(client mqtt.Client, topic string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal payload for %s: %v", topic, err)
		return
	}
	client.Publish(topic, 1, false, data)
}
