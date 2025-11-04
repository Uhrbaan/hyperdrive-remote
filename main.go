package main

import (
	"log"

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
}
