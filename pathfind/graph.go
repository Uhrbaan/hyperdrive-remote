package main

import (
	"os"

	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"
)

// type Node struct {
// 	TrackID               int
// 	AllowedTrackLocations []int
// }

type Node struct {
	ID       string // unique name for the node
	Track    int    // Track number (number of the piece)
	FromLane int    // Lower end of the lane
	ToLane   int    // Higher bound of the lane
}

type Edge struct {
	Source string
	Target string
	Weight int
}

var vertices = []Node{
	Node{"13", 13, 1, 12},
	Node{"20", 20, 1, 16},
	Node{"04-a", 4, 1, 4},
	Node{"04-b", 4, 5, 8},
	Node{"04-c", 4, 9, 16},
	Node{"21", 21, 1, 16},
	Node{"16", 16, 1, 12},
}

var edges = []Edge{
	Edge{"13", "20", 1},
	// Edge{"13-a", "01-a", 1},
	// Edge{"13-b", "01-b", 1},

	Edge{"20", "04-c", 1},
	Edge{"20", "04-a", 1},

	Edge{"04-c", "21", 1},
	Edge{"04-b", "21", 1},

	Edge{"21", "16", 1},
}

var nodeHash = func(n Node) string {
	return n.ID
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

func trackGraph() graph.Graph[string, Node] {
	g := graph.New(nodeHash)
	for _, node := range vertices {
		g.AddVertex(node)
	}
	for _, edge := range edges {
		g.AddEdge(edge.Source, edge.Target, graph.EdgeWeight(edge.Weight))
	}

	file, _ := os.Create("assets/track-graph.gv")
	draw.DOT(g, file)
	// To render the graph, you can run
	// dot -Tsvg -O "assets/track-graph.gv"

	return g
}
