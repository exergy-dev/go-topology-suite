// Package match wires the overlay-backed intersection and union
// functions into the measure package's AreaSimilarity hook. It exists
// solely to break the otherwise-cyclic dependency between measure
// (which is imported by overlay) and overlay (which AreaSimilarity
// needs to compute intersection/union areas).
//
// Importers that need AreaSimilarity to work should blank-import this
// package, typically from main:
//
//	import _ "github.com/exergy-dev/go-topology-suite/measure/match"
//
// The init() below is then run before any caller's code, registering
// overlay.Intersection / overlay.Union as the implementation backing
// measure.IntersectionFunc / measure.UnionFunc.
package match

import (
	"github.com/exergy-dev/go-topology-suite/measure"
	"github.com/exergy-dev/go-topology-suite/overlay"
)

func init() {
	measure.IntersectionFunc = overlay.Intersection
	measure.UnionFunc = overlay.Union
}
