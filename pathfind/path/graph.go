package path

import (
	"encoding/json"
	"fmt"
	"hyperdrive/remote/pathfind/util"
	"log"
	"os"

	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goccy/go-yaml"
)

type TrackConfig struct {
	Shapes map[string]ShapeDefinition `yaml:"shapes"`
	Edges  []EdgePair                 `yaml:"edges"`
}

type EdgePair struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

// ShapeDefinition holds the lane segments for a particular shape type.
type ShapeDefinition struct {
	Lanes []LaneSegment `yaml:"lanes"`
}

// LaneSegment defines a named segment within a shape, identified by 'from' and 'to' values.
type LaneSegment struct {
	Name string `yaml:"name"`
	From int    `yaml:"from"`
	To   int    `yaml:"to"`
}

const (
	trackYamlPath = "assets/track.yml"
	nextStepTopic = util.RootTopic + "/graph/nextStep"
	arrivedTopic  = util.RootTopic + "/graph/arrived"
)

func ImportYaml() graph.Graph[string, string] {
	b, err := os.ReadFile(trackYamlPath)

	g := graph.New(func(s string) string { return s })

	if err != nil {
		workdir, _ := os.Getwd()
		log.Fatal("Could not read file:", err, "\tWorkdir:", workdir)

	}
	var data TrackConfig
	yaml.Unmarshal(b, &data)

	uniqueVertices := map[string]bool{}
	for _, e := range data.Edges {
		uniqueVertices[e.Source] = true
		uniqueVertices[e.Target] = true
	}

	for k := range uniqueVertices {
		g.AddVertex(k)
	}

	for _, e := range data.Edges {
		g.AddEdge(e.Source, e.Target)
	}

	file, _ := os.Create("assets/track-graph.gv")
	draw.DOT(g, file)
	file.Close()

	return g
}

type nextStepPayload struct {
	NextStep string `json:"next_step"`
}

type arrivedPayload struct {
	Arrived bool `json:"arrived"`
}

type strChannel chan string

func (ch strChannel) targetTopicHandler(c mqtt.Client, m mqtt.Message) {
	var data tilePayload
	err := json.Unmarshal(m.Payload(), &data)
	if err != nil {
		log.Println("Could not unmarshal message:", string(m.Payload()))
		return
	}
	log.Println("[targetTopicHandler] Got a new target:", data.ID)

	// sanitize: we only get a number, and we need to map it to a string with the correct suffix if necessary.
	n := data.ID
	suffix := ""
	if _, exists := trackTypes[n]; exists {
		// if the selected element is an intersection, add a -c suffix (the car shall stop on the straight path).
		suffix = "-c"
	}

	if n == 0 || n == 17 {
		// we cannot stop on the crossing. This is invalid.
		log.Println("It is not allowed to stop on the crossing. Setting it to the default value 15.")
		n = 15
	}

	ch <- fmt.Sprintf("%02d%s", n, suffix)
}

func (ch strChannel) positionTopicHandler(c mqtt.Client, m mqtt.Message) {
	var data positionPayload
	err := json.Unmarshal(m.Payload(), &data)
	if err != nil {
		log.Println("Could not read message:", string(m.Payload()))
		return
	}
	log.Println("[positionTopicHandler] Got new position:", data.ID)

	if data.ID != "" {
		ch <- data.ID
	}
}

func PathCalculation(client mqtt.Client, g graph.Graph[string, string]) {
	targetUpdate := make(chan string)
	client.Subscribe(vehicleTargetTopic, 1, strChannel(targetUpdate).targetTopicHandler)

	positionUpdate := make(chan string)
	client.Subscribe(vehiclePositionTopic, 1, strChannel(positionUpdate).positionTopicHandler)

	var (
		target, position string
		ok               bool
	)
	for {
		select {
		case target, ok = <-targetUpdate:
			if !ok {
				continue
			}

		case position, ok = <-positionUpdate:
			if !ok {
				continue
			}
		}

		if target == "" || position == "" {
			continue
		}

		p, err := graph.ShortestPath(g, position, target)
		if err != nil {
			log.Fatal("Something went horribly wrong.")
		}

		if len(p) <= 1 {
			data, _ := json.Marshal(arrivedPayload{true})
			client.Publish(arrivedTopic, 1, false, data)
		} else {
			nextStep := p[1]
			data, _ := json.Marshal(nextStepPayload{nextStep})
			client.Publish(nextStepTopic, 1, false, data)
		}
	}
}
