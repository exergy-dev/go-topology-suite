// Decode geometries from WKT and run spatial predicates against them.
package main

import (
	"fmt"
	"log"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/predicate"
	"github.com/exergy-dev/go-topology-suite/wkt"
)

func main() {
	square := mustWKT("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")

	cases := []struct {
		label string
		wkt   string
	}{
		{"interior", "POINT (5 5)"},
		{"boundary", "POINT (10 5)"},
		{"exterior", "POINT (20 20)"},
	}

	for _, c := range cases {
		pt := mustWKT(c.wkt)
		intersects, err := predicate.Intersects(square, pt)
		if err != nil {
			log.Fatal(err)
		}
		contains, err := predicate.Contains(square, pt)
		if err != nil {
			log.Fatal(err)
		}
		touches, err := predicate.Touches(square, pt)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%-8s  %-14s  intersects=%-5v contains=%-5v touches=%-5v\n",
			c.label, c.wkt, intersects, contains, touches)
	}
}

func mustWKT(s string) geom.Geometry {
	g, err := wkt.Unmarshal(s)
	if err != nil {
		log.Fatalf("wkt.Unmarshal(%q): %v", s, err)
	}
	return g
}
