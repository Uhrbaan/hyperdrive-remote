package path

import (
	"container/ring"
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"slices"
	"strconv"
	"strings"
	"sync"
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

var (
	mu            sync.Mutex
	target        = "23"
	intersections = []int{4, 1, 5, 2, 9, 6, 12, 18, 19, 3}
	curves        = []int{13, 16, 15, 14}
	history       = ring.New(3) // The shortest loop is 6 elements long
)

func sign(a int) int {
	if a < 0 {
		return -1
	} else {
		return 1
	}
}

func trackShape(track int) string {
	switch true {
	case slices.Contains(curves, track):
		return "curve"
	case slices.Contains(intersections, track):
		return "intersection"
	default:
		return "straight"
	}
}

func position(track, lane int) string {
	shape := trackShape(track)
	var suffix string

	switch shape {
	case "curve", "straight":
		if lane < 9 {
			suffix = "bottom"
		} else {
			suffix = "top"
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

	return fmt.Sprintf("%02d.%s.%s", track, shape, suffix)
}

func VehicleTracking(client mqtt.Client, track graph.Graph[string, string]) {
	vehicleIdCh := make(chan string)
	client.Subscribe(vehicleIDTopic, 1, func(c mqtt.Client, m mqtt.Message) {
		var data vehicleIdPayload
		json.Unmarshal(m.Payload(), &data)
		vehicleIdCh <- data.ID
	})
	id := <-vehicleIdCh

	trackUpdate := make(chan trackPayload)
	client.Subscribe(fmt.Sprintf(vehicleTrackTopicFormat, id), 1, func(c mqtt.Client, m mqtt.Message) {
		var data []trackPayload

		err := json.Unmarshal(m.Payload(), &data)
		if err != nil {
			log.Println("Could not unmarshal data.")
			fmt.Println(string(m.Payload()))
			return
		}
		fmt.Println(string(m.Payload()))

		trackUpdate <- data[0]
	})

	client.Subscribe(vehicleTargetTopic, 1, func(c mqtt.Client, m mqtt.Message) {
		var data tilePayload
		err := json.Unmarshal(m.Payload(), &data)
		if err != nil {
			log.Println("Could not unmarshal message:", string(m.Payload()))
			return
		}

		// sanitize: we only get a number, and we need to map it to a string with the correct suffix if necessary.
		n := data.ID
		suffix := ""
		if slices.Contains(intersections, n) {
			// if the selected element is an intersection, add a -c suffix (the car shall stop on the straight path).
			suffix = "-c"
		}

		if n == 0 || n == 17 {
			// we cannot stop on the crossing. This is invalid.
			log.Println("It is not allowed to stop on the crossing. Setting it to the default value 15.")
			n = 15
		}

		mu.Lock()
		target = fmt.Sprintf("%02d%s", n, suffix)
		mu.Unlock()
	})

	adjacency, err := track.AdjacencyMap()
	if err != nil {
		log.Fatal("Unable to generate the Adjacency map:", err)
	}

	var duration time.Duration = 1500 * time.Millisecond
	timer := time.NewTicker(duration)

	for {
		select {
		case track := <-trackUpdate:
			fmt.Println("Current track history is:")
			history.Do(func(a any) {
				if a == nil {
					return
				}
				fmt.Println("\t" + a.(string))
			})

			node := position(track.Value.TrackID, track.Value.TrackLocation)
			history = history.Next()
			history.Value = node

			payload, _ := json.Marshal(tilePayload{
				ID: track.Value.TrackID,
			})
			client.Publish(vehiclePositionTopic, 1, false, payload)

			timer.Reset(duration)

		case <-timer.C:
			fmt.Println("Current track history is:")
			history.Do(func(a any) {
				if a == nil {
					return
				}
				fmt.Println("\t" + a.(string))
			})

			if history.Value == nil {
				continue
			}

			current := history.Value.(string)
			neighbours := adjacency[current]

			fmt.Println("Current neighbours of", current, "are:")
			for k := range neighbours {
				fmt.Println("\t", k)
			}

			// for the prediction, we assume the car did not change lanes, so we remove any keys with the ID of the current one
			for k := range neighbours {
				if k[:2] == current[:2] {
					fmt.Println("Removing key", k, "since we are not turning.")
					delete(neighbours, k)
				}
			}

			history.Do(func(a any) {
				if a == nil {
					return
				}
				// remove neighbours we've already been to
				delete(neighbours, a.(string))
			})

			if len(neighbours) == 0 {
				continue
			}

			// finally, we sort the remaining neighbours by the probability that it did not get detected. Usually, the left and right sections of intersections tend to not get detected, as well as inner curves
			neighboursList := slices.SortedFunc(maps.Keys(neighbours), func(a, b string) int {
				probability := func(s string) int {
					switch true {
					case strings.Contains(s, "curve.inner"):
						return 1
					case strings.Contains(s, "intersection.low"):
						return 2
					case strings.Contains(s, "intersection.high"):
						return 2
					case strings.Contains(s, "intersection.bottom"):
						return 3
					case strings.Contains(s, "curve.outer"):
						return 4
					case strings.Contains(s, "straight"):
						return 5
					default:
						return 0
					}
				}

				A := probability(a)
				B := probability(b)

				return A - B
			})

			// now, the most probable neighbour is the first in the list (hopefully). We add it to the history and send a message.
			fmt.Println("Probable neighbours of", current, "are:")
			for _, k := range neighboursList {
				fmt.Println("\t", k)
			}

			fmt.Println("Current history with current", current, "are:")
			history.Do(func(a any) {
				if a != nil {
					fmt.Println("\t" + a.(string))
				}
			})

			id, _ := strconv.Atoi(neighboursList[0][:2])
			payload, _ := json.Marshal(tilePayload{
				ID: id,
			})
			client.Publish(vehiclePositionTopic, 1, false, payload)

			history = history.Next()
			history.Value = neighboursList[0]
		}
	}
}
