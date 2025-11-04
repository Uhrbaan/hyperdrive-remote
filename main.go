package main

// Remote-Control for: Discovery (true/false)
// Remote-Control for: Connect of a/all/specific car (true/false)
// Remote-Control for: Driving the car (Speed)
// Remote-Control for: Lane-change (Steering)
// Remote-Control for: Lights

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	rpiIp    = "10.42.0.1"
	mqttPort = ":1883"
	// docsPort = ":18443"
)

type Subscription struct {
	Topic     string `json:"topic"`
	Subscribe bool   `json:"subscribe"`
}

func main() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(rpiIp + mqttPort)
	opts.SetClientID(uuid.NewString())

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("Could not establish connection with MQTT server: ", token.Error())
	}

	payload, _ := json.Marshal(Subscription{
		Topic:     "RemoteControl/+/E/hosts/discover",
		Subscribe: true,
	})

	if token := client.Publish("Anki/Hosts/U/I", 1, false, payload); token.Wait() && token.Error() != nil {
		log.Fatal("Failed to publish to Anki/Hosts/U/I : ", token.Error())
	}
	log.Println("Managed to publish", string(payload))

	topics := []string{
		"RemoteControl/+/E/vehicles/connect/#",
		"RemoteControl/+/E/vehicles/lights/#",
		"RemoteControl/+/E/vehicles/laneChange/#",
		"RemoteControl/+/E/vehicles/speed/#",
		"RemoteControl/+/E/vehicles/panic/#",
		"RemoteControl/+/E/vehicles/panic/#",
	}

	for _, t := range topics {
		payload, _ := json.Marshal(Subscription{
			Topic:     t,
			Subscribe: true,
		})
		if token := client.Publish("Anki/Hosts/U/I", 1, false, payload); token.Wait() && token.Error() != nil {
			log.Fatal("Failed to publish to Anki/Hosts/U/I : ", token.Error())
		}
		log.Println("Managed to publish", string(payload), "on", t)
	}

	if token := client.Subscribe("Anki/Hosts/U/hyperdrive/E/vehicle/discovered/#", 1, func(client mqtt.Client, message mqtt.Message) {
		msg := message.Payload()
		log.Println(string(msg), "on", message.Topic())
	}); token.Wait() && token.Error() != nil {
		log.Fatal("Failed to subscribe to discovered vehicles.", token.Error())
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // block until quit signal
	log.Println("Recieved quit signal. Exiting.")
}
