package path

import (
	"log"
	"os"

	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"
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
