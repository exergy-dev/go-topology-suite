// Boolean overlay operations: Union, Intersection, Difference,
// SymmetricDifference. Inputs and outputs are round-tripped via WKT for
// readability.
package main

import (
	"fmt"
	"log"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/overlay"
	"github.com/exergy-dev/go-topology-suite/wkt"
)

func main() {
	a := mustWKT("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	b := mustWKT("POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")

	for _, op := range []struct {
		name string
		fn   func(x, y geom.Geometry) (geom.Geometry, error)
	}{
		{"Intersection", overlay.Intersection},
		{"Union", overlay.Union},
		{"Difference (a - b)", overlay.Difference},
		{"SymmetricDifference", overlay.SymmetricDifference},
	} {
		out, err := op.fn(a, b)
		if err != nil {
			log.Fatalf("%s: %v", op.name, err)
		}
		text, err := wkt.Marshal(out)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%-22s  %s\n", op.name, text)
	}
}

func mustWKT(s string) geom.Geometry {
	g, err := wkt.Unmarshal(s)
	if err != nil {
		log.Fatalf("wkt.Unmarshal(%q): %v", s, err)
	}
	return g
}
