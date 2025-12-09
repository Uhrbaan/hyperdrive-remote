package main

import (
	"bytes"
	"encoding/binary"
	"hash/fnv"
	"log"
	"sort"

	"github.com/dominikbraun/graph"
)

type Node struct {
	TrackID               int
	AllowedTrackLocations []int
}

var nodeHash = func(n Node) uint64 {
	// Create the constant list of numbers
	l := make([]int, len(n.AllowedTrackLocations)+1)
	l[0] = n.TrackID
	s := l[1:]
	copy(s, n.AllowedTrackLocations)
	sort.Ints(s) // sort only what is *after* the track ID

	// Convert to bytes
	buf := new(bytes.Buffer)
	for _, v := range s {
		err := binary.Write(buf, binary.BigEndian, int32(v))
		if err != nil {
			log.Fatal("This should not happen.")
		}
	}

	// Produce the hash
	h := fnv.New64a()
	h.Write(buf.Bytes())
	return h.Sum64()
}

// func buildTrackGraph() graph.Graph[int, int] {
// 	g := graph.New(graph.IntHash)
// 	g.AddVertex(13)
// 	g.AddVertex(20)
// 	g.AddVertex(4)
// 	g.AddVertex(21)
// 	g.AddVertex(16)

// 	g.AddVertex(1)
// 	g.AddVertex(7)
// 	g.AddVertex(5)
// 	g.AddVertex(8)
// 	g.AddVertex(2)

// 	g.AddVertex(9)
// 	g.AddVertex(11)
// 	g.AddVertex(6)
// 	g.AddVertex(10)
// 	g.AddVertex(12)

// 	g.AddVertex(18)
// 	g.AddVertex(22)
// 	g.AddVertex(17)
// 	g.AddVertex(23)
// 	g.AddVertex(19)

// 	g.AddVertex(14)
// 	g.AddVertex(24)
// 	g.AddVertex(3)
// 	g.AddVertex(25)
// 	g.AddVertex(15)

// 	// Outermost cycle
// 	g.AddEdge(13, 20, graph.EdgeWeight(1))
// 	g.AddEdge(20, 4, graph.EdgeWeight(1))
// 	g.AddEdge(4, 21, graph.EdgeWeight(1))
// 	g.AddEdge(21, 16, graph.EdgeWeight(1))
// 	g.AddEdge(16, 2, graph.EdgeWeight(1))
// 	g.AddEdge(2, 12, graph.EdgeWeight(1))
// 	g.AddEdge(12, 19, graph.EdgeWeight(1))
// 	g.AddEdge(19, 15, graph.EdgeWeight(1))
// 	g.AddEdge(15, 25, graph.EdgeWeight(1))
// 	g.AddEdge(25, 3, graph.EdgeWeight(1))
// 	g.AddEdge(3, 24, graph.EdgeWeight(1))
// 	g.AddEdge(24, 14, graph.EdgeWeight(1))
// 	g.AddEdge(14, 18, graph.EdgeWeight(1))
// 	g.AddEdge(18, 9, graph.EdgeWeight(1))
// 	g.AddEdge(9, 1, graph.EdgeWeight(1))
// 	g.AddEdge(1, 13, graph.EdgeWeight(1))

// 	// Leftmost intersections
// 	g.AddEdge(1, 7, graph.EdgeWeight(1))
// 	g.AddEdge(9, 11, graph.EdgeWeight(1))
// 	g.AddEdge(18, 22, graph.EdgeWeight(1))

// 	// Second column
// 	g.AddEdge(7, 5, graph.EdgeWeight(1))
// 	g.AddEdge(11, 6, graph.EdgeWeight(1))
// 	g.AddEdge(22, 17, graph.EdgeWeight(1))

// 	// Middle column
// 	g.AddEdge(5, 4, graph.EdgeWeight(1))
// 	g.AddEdge(5, 8, graph.EdgeWeight(1))
// 	g.AddEdge(6, 10, graph.EdgeWeight(1))
// 	g.AddEdge(6, 17, graph.EdgeWeight(1))
// 	g.AddEdge(17, 23, graph.EdgeWeight(1))
// 	g.AddEdge(17, 3, graph.EdgeWeight(1))

// 	// Fourth column
// 	g.AddEdge(8, 2, graph.EdgeWeight(1))
// 	g.AddEdge(10, 12, graph.EdgeWeight(1))
// 	g.AddEdge(23, 19, graph.EdgeWeight(1))
// 	g.AddEdge(23, 19, graph.EdgeWeight(1))

// 	// Visualize graph
// 	file, _ := os.Create("assets/track-graph.gv")
// 	draw.DOT(g, file)
// 	// To render the graph, you can run
// 	// dot -Tsvg -O "assets/track-graph.gv"

// 	return g
// }

func trackGraph() {
	g := graph.New(nodeHash)
	g.AddVertex()
}
