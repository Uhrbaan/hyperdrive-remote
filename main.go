package main

// Remote-Control for: Discovery (true/false)
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
	"slices"
	"strings"
	"syscall"

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
	// opts.AddBroker("tcp://localhost:1883")
	opts.SetClientID(uuid.NewString())

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("Could not establish connection with MQTT server: ", token.Error())
	}

	remote := remote.Remote{Client: client}
	remote.SyncWith("Anki/Hosts/U/hyperdrive/I", true)

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Println("Please enter [true|false].")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSpace(text)
			fmt.Println("User entered:", text+".")
			discover := false
			if slices.Contains([]string{"true", "True", "t", "T"}, text) {
				discover = true
			}
			remote.Discover(discover)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // block until quit signal
	log.Println("Recieved quit signal. Exiting.")
}
