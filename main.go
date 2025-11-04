package main

// Remote-Control for: Discovery (true/false) ✅
// Remote-Control for: Connect of a/all/specific car (true/false)
// Remote-Control for: Driving the car (Speed)
// Remote-Control for: Lane-change (Steering)
// Remote-Control for: Lights

import (
	"bufio"
	"fmt"
	"hyperdrive/remote/remote"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	rpiIp    = "10.42.0.1"
	mqttPort = ":1883"
)

func main() {
	// Configure the mosquitto client
	opts := mqtt.NewClientOptions()
	opts.AddBroker(rpiIp + mqttPort)
	opts.SetClientID(uuid.NewString())

	// Connect to the broker and initialize the client
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("Could not establish connection with MQTT server: ", token.Error())
	}

	// Disconnect cleanly from the broker when closing the program
	defer client.Disconnect(0)

	// Create a new Remote object and sync the discover event with the rest of the system
	remote := remote.Remote{Client: client}
	remote.SyncDiscoverWith("Anki/Hosts/U/hyperdrive/I", true)

	// Run function in background that listens on the console for commands
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Println("Please enter a command of the shape 'name=value' and press Enter.")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSpace(text)
			values := strings.Split(text, "=")

			switch values[0] {
			case "discover":
				discover := false
				if values[1] == "true" {
					discover = true
				}
				remote.Discover(discover)
			}
		}
	}()

	// Block the main function from finishing until it recieves an interrupt (ctrl-C)
	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
