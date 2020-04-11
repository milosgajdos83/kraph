package store

import (
	"github.com/milosgajdos/kraph/api"
	"github.com/milosgajdos/kraph/query"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
)

// DOTAttributes are attributes for Graphiz DOT graph
type DOTAttributes interface {
	Attributes
	DOTAttributes() []encoding.Attribute
}

// Attributes provide a simple key-value store
// for storing arbitrary entity properties
type Attributes interface {
	// Keys returns all attribute keys
	Keys() []string
	// Attributes returns all attributes as a slice of encoding.Attribute
	Attributes() []encoding.Attribute
	// Get returns the attribute value for the given key
	Get(string) string
	// Set sets the value of the attribute for the given key
	Set(string, string)
}

// Metadata provides a simple key-valule store
// for arbitrary entity data of arbitrary type
type Metadata interface {
	// Keys returns all metadata keys
	Keys() []string
	// Get returns the attribute value for the given key
	Get(string) interface{}
	// Set sets the value of the attribute for the given key
	Set(string, interface{})
}

// Entity is an arbitrary store entity
type Entity interface {
	// Attributes returns attributes
	Attributes() Attributes
	// Metadata returns metadata
	Metadata() Metadata
}

// DOTNode is a GraphViz DOT Node
type DOTNode interface {
	Node
	// DOTID returns Graphiz DOT ID
	DOTID() string
	// SetDOTID sets Graphiz DOT ID
	SetDOTID(string)
}

// Node is a graph node
type Node interface {
	Entity
	graph.Node
	// Name returns node name
	Name() string
}

// Edge is an edge between two nodes
type Edge interface {
	Entity
	graph.WeightedEdge
}

// DOTStore returns Graphiz DOT store
type DOTStore interface {
	Store
	// DOTID returns DOT graph ID
	DOTID() string
	// DOTAttributers returns global graph DOT attributes
	DOTAttributers() (graph, node, edge encoding.Attributer)
	// DOT returns Graphiz graph
	DOT() (string, error)
}

// Store allows to store and query the graph of API objects
type Store interface {
	// Add adds an api.Object to the store and returns a Node or error
	Add(api.Object, ...Option) (Node, error)
	// Link links two nodes and returns the new edge between them or error
	Link(Node, Node, ...Option) (Edge, error)
	// Query queries the store and returns the results or error
	Query(...query.Option) ([]Entity, error)
	// Subgraph returns the subgraph of the node up to given depth or returns error
	SubGraph(Node, int) (graph.Graph, error)
}
