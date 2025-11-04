package main

import (
	// "hyperdrive/remote/hyperdrive"
	"encoding/json"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	rpiIp    = "10.42.0.1"
	mqttPort = ":1883"
	// docsPort = ":18443"
)

func main() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(rpiIp + mqttPort)
	opts.SetClientID(uuid.NewString())

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("Could not establish connection with MQTT server: ", token.Error())
	}

	type intent struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}

	type discoverPayload struct {
		Topic     string `json:"topic"`
		Subscribe bool   `json:"subscribe"`
	}

	msg := intent{
		Type: "discoverSubscription",
		Payload: discoverPayload{
			Topic:     "RemoteControl/+/E/hosts/discover",
			Subscribe: true,
		},
	}

	b, err := json.Marshal(msg)
	if err != nil {
		log.Fatal("failed to marshal intent: ", err)
	}

	pubTopic := "Anki/Hosts/U/I"
	p := client.Publish(pubTopic, 0, false, b)
	if ok := p.WaitTimeout(5 * time.Second); !ok {
		log.Println("publish did not complete within timeout, continuing")
	}
	if p.Error() != nil {
		log.Fatal("publish error: ", p.Error())
	}

	log.Printf("Published intent to %s: %s\n", pubTopic, string(b))

	// Give broker a moment then disconnect
	time.Sleep(250 * time.Millisecond)
	client.Disconnect(250)
}
