package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dominikbraun/graph"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type trackPayload struct {
	Timestamp uint64 `json:"timestamp"`
	Value     struct {
		TrackID       int    `json:"trackID"`
		TrackLocation int    `json:"trackLocation"`
		Direction     string `json:"direction"`
	} `json:"value"`
}

type positionPayload struct {
	ID int `json:"id"`
}

const (
	vehicleTrackTopicFormat = "Anki/Vehicles/U/%s/E/track"
	vehiclePositionTopic    = "/hobHq10yb9dKwxrdfhtT/vehicle/position"
)

func VehicleTracking(client mqtt.Client, id string, track graph.Graph[int, int]) {
	trackUpdate := make(chan trackPayload)
	client.Subscribe(fmt.Sprintf(vehicleTrackTopicFormat, id), 1, func(c mqtt.Client, m mqtt.Message) {
		var data trackPayload

		err := json.Unmarshal(m.Payload(), &data)
		if err != nil {
			log.Fatal("Could not unmarshal data.")
		}

		trackUpdate <- data
	})

	adjacency, _ := track.AdjacencyMap()
	fmt.Println(adjacency)

	for {
		// 0. Wait for an update in the track ID
		track := <-trackUpdate

		// 1. take all possible neighbours of the current (latest) track ID
		neighbours := []int{}
		for k := range adjacency[path[1]] {
			if path[0] != k { // current can not be previous
				neighbours = append(neighbours, k)
			}
		}
		fmt.Printf("The possible neighbours are: %v\n", neighbours)

		// 2. if there is only one neighbour, then we are coming from that one
		if len(neighbours) == 1 {
			payload, _ := json.Marshal(positionPayload{
				ID: neighbours[0],
			})
			client.Publish(vehiclePositionTopic, 1, false, payload)
		}

		// 3. If we have more elements, then we need to use the rest of the information

		time.Sleep(1 * time.Second)
	}
}
