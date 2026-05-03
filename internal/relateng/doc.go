// Package relateng is an in-progress port of JTS's
// org.locationtech.jts.operation.relateng package.
//
// JTS RelateNG is a re-architecture of the classic RelateOp that:
//
//  1. Uses a tri-state TopologyPredicate (UNKNOWN / TRUE / FALSE) so each
//     predicate can short-circuit as soon as enough of the DE-9IM matrix
//     has been determined.
//  2. Builds the topology graph incrementally via a TopologyComputer fed
//     by per-segment intersections, instead of constructing the whole
//     graph up front.
//  3. Locates points via a dedicated RelatePointLocator that supports
//     mixed-dimension GeometryCollections with proper union semantics.
//
// Wave-state: the foundation classes (RelateGeometry, RelatePointLocator,
// LinearBoundary, DimensionLocation), the predicate-interface family
// (TopologyPredicate, BasicPredicate, IMPredicate, IMPatternMatcher,
// IntersectsPredicate, DisjointPredicate, NewMatchesPredicate), and
// the NodeSection data class are ported here.
//
// Not yet ported (planned for the next wave — these are interlocking
// and best landed together):
//
//   - NodeSections / RelateNode / RelateEdge: the per-vertex topology
//     graph nodes, with CCW edge ordering by angle. RelateNode.finish
//     (which propagates side locations around the node) is the
//     non-trivial piece — it depends on JTS PolygonNodeTopology, which
//     is itself a separate port.
//   - PolygonNodeConverter: re-orients polygon node sections so the
//     shell is CW.
//   - EdgeSegmentIntersector + EdgeSegmentOverlapAction: drives
//     segment-pair intersection across self-noding monotone chains,
//     emitting NodeSections.
//   - EdgeSetIntersector: top-level driver that pairs edge segments
//     of A with B (and A with itself if self-noding required).
//   - TopologyComputer: the final orchestrator. Walks vertices /
//     edges of A and B, pushes nodes into RelateNodes, consults the
//     PointLocator for "lone" vertices, and feeds the resulting
//     DE-9IM cells into the bound TopologyPredicate.
//   - AdjacentEdgeLocator: the multi-polygon edge-adjacency case in
//     RelatePointLocator currently falls back to a conservative
//     answer (BOUNDARY); AdjacentEdgeLocator depends on RelateNode.
//
// The Wave 3 plan: once RelateNode + EdgeSegmentIntersector +
// TopologyComputer are in place, predicate/relate.go's `Relate(a,b)`
// public entry can dispatch to a TopologyComputer-backed path. The
// existing legacy DE-9IM builder remains as a fallback for the
// not-yet-supported geometry-pair cases.
//
// The package is internal and reserved for terra's own predicate layer;
// stable users should continue to use github.com/terra-geo/terra/predicate.
package relateng
