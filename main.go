package main

// Remote-Control for: Discovery (true/false) ✅
// Remote-Control for: Connect of a/all/specific car (true/false) ✅ (kinda...)
// Remote-Control for: Driving the car (Speed)
// Remote-Control for: Lane-change (Steering)
// Remote-Control for: Lights

import (
	"log"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	rpiIp = "10.42.0.1"
	// rpiIp    = "test.mosquitto.org"
	mqttPort = 1883
)

func main() {
	// Make a new client to send the necessary topic, since it is decoupled
	opts := mqtt.NewClientOptions()
	opts.AddBroker(rpiIp + ":" + strconv.Itoa(mqttPort))
	opts.SetClientID(uuid.NewString())

	// Connect to the broker and initialize the client
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("Could not establish connection with MQTT server: ", token.Error())
	}
	log.Println("Connected to mosquitto broker on", rpiIp+":"+strconv.Itoa(mqttPort))

}
