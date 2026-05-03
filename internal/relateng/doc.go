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
// LinearBoundary, DimensionLocation) and the predicate-interface family
// (TopologyPredicate, BasicPredicate, IMPredicate) are ported here.
//
// Not yet ported (planned for the next wave):
//
//   - NodeSection / NodeSections / RelateNode / RelateEdge data classes
//   - EdgeSegmentIntersector / EdgeSetIntersector
//   - TopologyComputer
//   - PolygonNodeConverter
//   - AdjacentEdgeLocator (the multi-polygon edge-adjacency case in
//     RelatePointLocator currently falls back to a conservative answer).
//
// The package is internal and reserved for terra's own predicate layer;
// stable users should continue to use github.com/terra-geo/terra/predicate.
package relateng
