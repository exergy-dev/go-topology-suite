// Package noding provides segment "noding" — splitting a set of input
// line segments at every interior intersection so the output set has
// the property that any two segments either don't intersect at all or
// touch only at a vertex endpoint (a "node").
//
// Noding is the precursor to overlay topology construction: once an
// edge graph is noded, two edges can never cross in their interiors,
// so a planar subdivision can be built by walking shared endpoints
// without further geometric work.
//
// This package is internal — it is consumed by the overlay-NG port
// (Pillar A1 of the Terra roadmap) and is not part of the public API.
//
// Exposed types:
//
//   - SegmentString: a sequence of vertices defining a connected
//     polyline (or polygon ring boundary, when first==last). Carries
//     a caller-defined Tag used by the overlay package to mark which
//     input geometry an edge originated from.
//
//   - Noder: the strategy interface. Node takes input segment strings
//     and returns a topologically equivalent noded set.
//
//   - SimpleNoder: a brute-force O(n^2) pairwise noder. Correct on all
//     inputs, suitable for small-to-medium polygons. The follow-up
//     optimisation is a monotone-chain index (cf. JTS MCIndexNoder).
//
// Limitations:
//
//   - Collinear-overlap is not handled: when two segments lie on the
//     same line and partially overlap, the underlying segment-
//     intersection primitive returns no intersection, so the overlap
//     passes through as-is. This matches the behaviour of
//     kernel/planar.SegmentIntersection and is documented as a known
//     gap until the overlay package adds explicit overlap handling.
package noding
