// Build an R-tree from a small set of envelopes and run an
// envelope-intersects query. The R-tree is generic over the value type;
// here we associate a string label with each envelope.
package main

import (
	"fmt"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/index"
)

func main() {
	tree := index.New[string]()

	cities := []struct {
		name string
		env  geom.Envelope
	}{
		{"Manhattan", box(-74.02, 40.70, -73.93, 40.88)},
		{"Brooklyn", box(-74.05, 40.57, -73.83, 40.74)},
		{"Queens", box(-73.96, 40.54, -73.70, 40.80)},
		{"Bronx", box(-73.93, 40.78, -73.76, 40.92)},
		{"Staten Island", box(-74.26, 40.49, -74.05, 40.65)},
	}
	for _, c := range cities {
		tree.Insert(c.env, c.name)
	}

	// Query: everything intersecting a small box around Times Square.
	query := box(-74.00, 40.74, -73.97, 40.78)

	fmt.Printf("Boroughs whose bounding box intersects %v:\n", query)
	tree.Search(query, func(it index.Item[string]) bool {
		fmt.Printf("  - %s (env=%v)\n", it.Value, it.Env)
		return true // keep iterating; return false to stop early.
	})
	fmt.Printf("(tree holds %d items)\n", tree.Len())
}

func box(minX, minY, maxX, maxY float64) geom.Envelope {
	return geom.Envelope{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}
}
