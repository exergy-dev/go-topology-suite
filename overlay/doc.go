// Package overlay computes the four boolean polygon operations
// (Intersection, Union, Difference, SymmetricDifference).
//
// **v0.1 status: PARTIAL.** Only convex-clipper cases are supported,
// implemented via Sutherland-Hodgman: Intersection works correctly when
// the second operand (the "clipper") is a convex polygon and the first
// (the "subject") is any polygon. Union, Difference, SymmetricDifference,
// and the general non-convex Intersection all return ErrUnsupportedKernel
// with a message pointing at the JTS overlay-NG port that gates this
// work (Phase 3, E5 in the parallel plan).
//
// Callers requiring full overlay today should not use this package — they
// should either project to PostGIS and call ST_Intersection, or wait for
// the overlay-NG port. The convex-clip path is included because it is
// genuinely useful (clip-to-bbox, clip-to-tile-quad) and the algorithm
// is standalone, robust, and easy to audit.
package overlay
