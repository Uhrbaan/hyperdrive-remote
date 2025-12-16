package path

import (
	"encoding/json"
	"fmt"
	"hyperdrive/remote/pathfind/instruct"
	"hyperdrive/remote/pathfind/util"
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
	vehicleTrackTopicFormat      = "Anki/Vehicles/U/%s/E/track"
	vehicleAbsolutePositionTopic = util.RootTopic + "/vehicle/absolute-position"
	vehiclePredictionTopic       = util.RootTopic + "/vehicle/prediction"
	vehiclePositionTopic         = util.RootTopic + "/vehicle/position"
)

var TrackTypes = map[int]string{
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

func getTrackShape(trackID int) string {
	if shape, exists := TrackTypes[trackID]; exists {
		return shape
	}
	return "straight"
}

func calculatePositionNode(trackID, lane int) string {
	shape := getTrackShape(trackID)
	var suffix string

	switch shape {
	case "curve":
		suffix = "outer"
		if lane < 9 {
			suffix = "inner"
		}
	case "straight":
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
	vehicleID := util.WaitForVehicleID(client)
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
		log.Println("[Vehicle] The next step is", data.NextStep)
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

		select {
		case trackData := <-trackCh: // Getting data from vehicle
			if trackData.Value.TrackID == 0 {
				log.Println("Received invalid ID of 0")
				continue
			}

			currentPositionNode := calculatePositionNode(trackData.Value.TrackID, trackData.Value.TrackLocation)
			updateHistory(currentPositionNode)

			fmt.Printf("Track Update. History: %v\n", history)

			util.SendJSON(client, vehicleAbsolutePositionTopic, tilePayload{ID: trackData.Value.TrackID})
			util.SendJSON(client, vehiclePositionTopic, positionPayload{history[len(history)-1]})
			timer.Reset(predictionTimeout)

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

			updateHistory(predictedNode)
			util.SendJSON(client, vehiclePredictionTopic, tilePayload{ID: predictedID})
			util.SendJSON(client, vehiclePositionTopic, positionPayload{history[len(history)-1]})

		case nextStep := <-nextStepCh:
			// fmt.Println("[Vehicle] nextStep:", nextStep, "currentPositionNode:", history[len(history)-1])
			currentNode := history[len(history)-1]
			if nextStep == "" || history[len(history)-1] == "" {
				continue
			}

			instruction := instruct.LaneChangeMessage{}
			// only turn on straight lines
			if nextStep[:2] == currentNode[:2] {
				// turn right
				if strings.Contains(nextStep, "top") && strings.Contains(currentNode, "bottom") {
					instruction.LaneChange = "right"
				} else if strings.Contains(currentNode, "top") && strings.Contains(nextStep, "bottom") {
					instruction.LaneChange = "left"
				} else {
					instruction.LaneChange = ""
				}
			}

			// if slices.Contains(history, nextStep) {
			// 	// if the next step is in the history, it means we have to drive back.
			// 	instruction.Forward = false
			// } else {
			instruction.Forward = true
			// }

			util.SendJSON(client, instruct.InstructionTopic, instruction)
			log.Println("[Vehicle] To go to next step, going:", instruction)
		}
	}
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
