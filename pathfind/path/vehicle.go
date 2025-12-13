package path

import (
	"encoding/json"
	"fmt"
	"hyperdrive/remote/hyperdrive"
	"log"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dominikbraun/graph"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	rootTopic               = "/hobHq10yb9dKwxrdfhtT" + hyperdrive.UserSuffix
	vehicleTrackTopicFormat = "Anki/Vehicles/U/%s/E/track"
	vehiclePositionTopic    = rootTopic + "/vehicle/position"
	vehiclePredictionTopic  = rootTopic + "/vehicle/prediction"
	vehicleInstructionTopic = rootTopic + "/vehicle/instruction"
)

var trackTypes = map[int]string{
	13: "curve", 16: "curve", 15: "curve", 14: "curve",
	4: "intersection", 1: "intersection", 5: "intersection", 2: "intersection",
	9: "intersection", 6: "intersection", 12: "intersection", 18: "intersection", 19: "intersection", 3: "intersection",
}

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

type instructionPayload struct {
	Id         string `json:"id"`
	LaneChange string `json:"lane_change"`
	Forward    bool   `json:"forward"`
}

func getTrackShape(trackID int) string {
	if shape, exists := trackTypes[trackID]; exists {
		return shape
	}
	return "straight"
}

func calculatePositionNode(trackID, lane int) string {
	shape := getTrackShape(trackID)
	var suffix string

	switch shape {
	case "curve", "straight":
		suffix = "top"
		if lane < 9 {
			suffix = "bottom"
		}
	case "intersection":
		if lane < 5 {
			suffix = "low"
		} else if lane >= 5 && lane < 9 {
			suffix = "high"
		} else {
			suffix = "bottom"
		}
	}
	return fmt.Sprintf("%02d.%s.%s", trackID, shape, suffix)
}

func getPredictionProbability(nodeID string) int {
	switch {
	case strings.Contains(nodeID, "curve.inner"):
		return 1
	case strings.Contains(nodeID, "intersection.low"), strings.Contains(nodeID, "intersection.high"):
		return 2
	case strings.Contains(nodeID, "intersection.bottom"):
		return 3
	case strings.Contains(nodeID, "curve.outer"):
		return 4
	case strings.Contains(nodeID, "straight"):
		return 5
	default:
		return 0
	}
}

func VehicleTracking(client mqtt.Client, trackGraph graph.Graph[string, string]) {
	vehicleID := waitForVehicleID(client)
	log.Printf("Starting tracking for Vehicle ID: %s", vehicleID)

	trackCh := make(chan trackPayload)
	client.Subscribe(fmt.Sprintf(vehicleTrackTopicFormat, vehicleID), 1, func(c mqtt.Client, m mqtt.Message) {
		var data []trackPayload
		if err := json.Unmarshal(m.Payload(), &data); err != nil {
			log.Printf("Error unmarshalling track: %v", err)
			return
		}
		if len(data) > 0 {
			trackCh <- data[0]
		}
	})

	nextStepCh := make(chan string)
	client.Subscribe(nextStepTopic, 1, func(c mqtt.Client, m mqtt.Message) {
		var data nextStepPayload
		if err := json.Unmarshal(m.Payload(), &data); err != nil {
			log.Printf("Error unmarshalling next step: %v", err)
			return
		}
		nextStepCh <- data.NextStep
	})

	adjacency, err := trackGraph.AdjacencyMap()
	if err != nil {
		log.Fatal("Unable to generate Adjacency map:", err)
	}

	history := make([]string, 0, 4)
	predictionTimeout := 1000 * time.Millisecond
	timer := time.NewTicker(predictionTimeout)
	defer timer.Stop()

	updateHistory := func(newNode string) {
		if len(history) > 0 && history[len(history)-1] == newNode {
			return
		}
		if len(history) >= 4 {
			history = history[1:]
		}
		history = append(history, newNode)
	}

	for {
		var currentPositionNode string
		stateUpdated := false

		select {
		case trackData := <-trackCh: // Getting data from vehicle
			if trackData.Value.TrackID == 0 {
				log.Println("Received invalid ID of 0")
				continue
			}

			currentPositionNode = calculatePositionNode(trackData.Value.TrackID, trackData.Value.TrackLocation)
			updateHistory(currentPositionNode)

			fmt.Printf("Track Update. History: %v\n", history)

			sendJSON(client, vehiclePositionTopic, tilePayload{ID: trackData.Value.TrackID})
			timer.Reset(predictionTimeout)
			stateUpdated = true

		case <-timer.C: // prediction on timeout
			if len(history) == 0 {
				continue
			}
			fmt.Println("------ PREDICTION --------")

			currentNode := history[len(history)-1]
			predictedNode, ok := predictNextNode(currentNode, history, adjacency)

			if !ok {
				log.Println("Could not predict next node")
				continue
			}

			// Extract ID from string (e.g., "02.curve" -> 2)
			predictedID, _ := strconv.Atoi(predictedNode[:2])

			fmt.Printf("Predicted: %s. New History: %v\n", predictedNode, history)
			sendJSON(client, vehiclePredictionTopic, tilePayload{ID: predictedID})

			updateHistory(predictedNode)
			currentPositionNode = predictedNode
			stateUpdated = true

		case nextStep := <-nextStepCh:
			// if ids are the same and changing from bottom -> top or the opposite
			if nextStep[:2] != currentPositionNode[:2] {
				continue
			}

			instruction := instructionPayload{Id: vehicleID}
			// turn right
			if strings.Contains(nextStep, "top") && strings.Contains(currentPositionNode, "bottom") {
				instruction.LaneChange = "rigth"
			} else {
				instruction.LaneChange = "left"
			}

			// TODO: implement calculation to know which direction we should drive in.
			instruction.Forward = true

			sendJSON(client, vehicleInstructionTopic, instruction)
		}

		if stateUpdated && currentPositionNode != "" {
			sendJSON(client, vehiclePositionTopic, positionPayload{ID: currentPositionNode})
		}
	}
}

// waitForVehicleID blocks until a vehicle ID is received.
func waitForVehicleID(client mqtt.Client) string {
	ch := make(chan string)
	token := client.Subscribe(vehicleIDTopic, 1, func(c mqtt.Client, m mqtt.Message) {
		var data vehicleIdPayload
		if err := json.Unmarshal(m.Payload(), &data); err == nil {
			ch <- data.ID
		}
	})
	token.Wait()

	defer client.Unsubscribe(vehicleIDTopic)
	return <-ch
}

// predictNextNode calculates the most likely next node based on history and graph.
func predictNextNode(current string, history []string, adjacency map[string]map[string]graph.Edge[string]) (string, bool) {
	neighbours, exists := adjacency[current]
	if !exists || len(neighbours) == 0 {
		return "", false
	}

	// Filter candidates
	candidates := make([]string, 0, len(neighbours))
	currentPrefix := current[:2] // Assumes ID format "XX....."

	for k := range neighbours {
		// not the same tile id (don't change lanes)
		if len(k) >= 2 && k[:2] == currentPrefix {
			continue
		}
		// don't go back
		if slices.Contains(history, k) {
			continue
		}
		candidates = append(candidates, k)
	}

	if len(candidates) == 0 {
		return "", false
	}

	// Sort by probability (Ascending order based on original code logic)
	sort.Slice(candidates, func(i, j int) bool {
		return getPredictionProbability(candidates[i]) < getPredictionProbability(candidates[j])
	})

	return candidates[0], true
}

func sendJSON(client mqtt.Client, topic string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal payload for %s: %v", topic, err)
		return
	}
	client.Publish(topic, 1, false, data)
}
