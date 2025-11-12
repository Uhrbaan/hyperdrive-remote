package main

// Remote-Control for: Discovery (true/false) ✅
// Remote-Control for: Connect of a/all/specific car (true/false) ✅ (kinda...)
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
	rpiIp = "10.42.0.1"
	// rpiIp    = "test.mosquitto.org"
	mqttPort = 1883
)

const (
	vehiclesDiscoverTopic        = "Anki/Hosts/U/hyperdrive/E/vehicle/discovered/#"
	ankiVehicleSubscriptionTopic = "Anki/Vehicles/U/%s/I"
)

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

	// Sending the discover subscription. Waiting to make sure that is
	remote.SyncSubscription(client, "discoverSubscription", "Anki/Hosts/U/I", remote.DiscoverTopic, true)
	log.Println("Sent discover subscription.")
	time.Sleep(2 * time.Second)

	// Sending the required topic to the remote.
	payload, _ := json.Marshal(remote.DiscoverVehiclesTopic{Topic: vehiclesDiscoverTopic})
	client.Publish(remote.ListenDiscoverTopicTopic, 1, false, payload)
	log.Println()

	// // Enableing the subscriptions
	// for vehicle := range vehicleListCh {
	// 	remote.SyncSubscription(client, "connectSubscription", fmt.Sprintf(ankiVehicleSubscriptionTopic, vehicle), fmt.Sprintf(remote.ConnectTopic, vehicle), true)
	// 	remote.SyncSubscription(client, "speedSubscription", fmt.Sprintf(ankiVehicleSubscriptionTopic, vehicle), fmt.Sprintf(remote.SpeedTopic, vehicle), true)
	// }

	// // Tell remote process that all the subscriptions went through
	// waitForSubscriptions <- true

	// Block execution until somebody types ctrl-c.
	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
