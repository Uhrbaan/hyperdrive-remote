package main

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
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
	ID string `json:"id"`
}

const (
	vehicleTrackTopicFormat = "Anki/Vehicles/U/%s/E/track"
	vehiclePositionTopic    = "/hobHq10yb9dKwxrdfhtT/vehicle/position"
)

func VehicleTracking(client mqtt.Client, id string, track graph.Graph[string, Node]) {
	trackUpdate := make(chan trackPayload)
	client.Subscribe(fmt.Sprintf(vehicleTrackTopicFormat, id), 1, func(c mqtt.Client, m mqtt.Message) {
		var data trackPayload

		err := json.Unmarshal(m.Payload(), &data)
		if err != nil {
			log.Fatal("Could not unmarshal data.")
		}

		trackUpdate <- data
	})

	previousTrack := ""
	adjacency, _ := track.AdjacencyMap()
	fmt.Println(adjacency)

	for {
		// 0. Wait for an update in the track ID
		track := <-trackUpdate
		number := track.Value.TrackID
		lane := track.Value.TrackLocation
		suffix := ""

		// add a suffix if the current track is an intersection track.
		if slices.Contains([]int{4, 1, 5, 2, 9, 6, 12, 18, 19, 3}, number) {
			switch true {
			case lane <= 4 && lane >= 1:
				suffix = "-a"
			case lane <= 8 && lane >= 5:
				suffix = "-b"
			case lane <= 16 && lane >= 9:
				suffix = "-c"
			}
		}

		nodeID := fmt.Sprintf("%02d%s", number, suffix)

		// 1. take all possible neighbours of the current (latest) track ID
		neighbours := []string{}
		for k := range adjacency[nodeID] {
			if previousTrack != k { // current can not be previous
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

		// update previous track
		previousTrack = nodeID
		time.Sleep(1 * time.Second)
	}
}
