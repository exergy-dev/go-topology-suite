// Package predicate computes spatial predicates between Terra geometries.
//
// All predicate functions take two geometries and a variadic Option list
// (functional options pattern). The kernel is selected via WithKernel; if
// omitted, planar is used. CRS mismatches return terra.ErrCRSMismatch.
//
// Phase 1 ships the core predicates needed by typical workloads:
// Intersects, Disjoint, Equals, and Contains. Within, Crosses, Overlaps,
// Touches, Covers, CoveredBy, and the full Relate (DE-9IM) implementation
// arrive in Phase 2 once the overlay graph lands.
package predicate
