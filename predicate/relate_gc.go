package predicate

import "github.com/exergy-dev/go-topology-suite/geom"

// SetUnaryUnion is a no-op retained for overlay's init wiring;
// RelateNG handles GeometryCollection operands natively.
func SetUnaryUnion(_ func(g geom.Geometry) (geom.Geometry, error)) {}
