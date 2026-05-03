package predicate

import "github.com/terra-geo/terra/geom"

// SetUnaryUnion was previously wired by the overlay package to give
// the legacy DE-9IM pipeline a way to simplify GeometryCollection
// operands before classification. The RelateNG driver
// (internal/relateng) handles GC operands natively, so this hook is
// now a no-op kept as a stable surface for the overlay package's
// init wiring.
func SetUnaryUnion(_ func(g geom.Geometry) (geom.Geometry, error)) {}
