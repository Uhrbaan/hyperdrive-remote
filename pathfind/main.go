package main

import (
	"log"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	rpiIp    = "10.42.0.1"
	mqttPort = 1883
	// rpiIp                = "test.mosquitto.org"
)

func main() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(rpiIp + ":" + strconv.Itoa(mqttPort))
	opts.SetClientID(uuid.NewString())

	// Connect to the broker and initialize the client
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("Could not establish connection with MQTT server: ", token.Error())
	}
	log.Println("Connected to mosquitto broker on", rpiIp+":"+strconv.Itoa(mqttPort))

	g := buildTrackGraph()
	go VehicleTracking(client, "8EF0310F", g)

	UI(client)
}
