package main

// Remote-Control for: Discovery (true/false) ✅
// Remote-Control for: Connect of a/all/specific car (true/false)
// Remote-Control for: Driving the car (Speed)
// Remote-Control for: Lane-change (Steering)
// Remote-Control for: Lights

import (
	"encoding/json"
	"hyperdrive/remote/remote"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	rpiIp    = "10.42.0.1"
	mqttPort = 1883
)

const defaultVehiclesDiscoverTopic = "Anki/Hosts/U/hyperdrive/E/vehicle/discovered/#"

func main() {
	go remote.StartRemote(rpiIp, mqttPort, uuid.NewString())

	// Make a new client to send the necessary topic, since it is decoupled
	opts := mqtt.NewClientOptions()
	opts.AddBroker(rpiIp + ":" + strconv.Itoa(mqttPort))
	opts.SetClientID(uuid.NewString())

	// Connect to the broker and initialize the client
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("Could not establish connection with MQTT server: ", token.Error())
	}
	log.Println("Connected to mosquitto broker.")

	time.Sleep(1 * time.Second) // waiting a second else it's too fast for the remote process

	// Sending the required topic to the remote.
	payload, _ := json.Marshal(remote.DiscoverVehiclesTopic{Topic: defaultVehiclesDiscoverTopic})
	client.Publish(remote.ListenDiscoverTopicTopic, 1, false, payload)
	log.Println()

	// Run function in background that listens on the console for commands
	// go func() {
	// 	reader := bufio.NewReader(os.Stdin)
	// 	for {
	// 		fmt.Println("Please enter a command of the shape 'name=value' and press Enter.")
	// 		text, _ := reader.ReadString('\n')
	// 		text = strings.TrimSpace(text)
	// 		values := strings.Split(text, "=")

	// 		switch values[0] {
	// 		case "discover":
	// 			discover := false
	// 			if values[1] == "true" {
	// 				discover = true
	// 			}
	// 			go r.Discover(discover)
	// 			data, _ := json.Marshal(remote.DiscoverVehiclesTopic{Topic: defaultVehiclesDiscoverTopic})
	// 			client.Publish(remote.ListenDiscoverTopicTopic, 1, false, data)
	// 		}
	// 	}
	// }()

	// Block execution until somebody types ctrl-c.
	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
