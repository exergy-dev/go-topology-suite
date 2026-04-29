// Package prepare offers "prepared" wrappers around geometries that
// pre-compute auxiliary indexes so repeated spatial queries against a single
// fixed geometry run in amortised O(log n + k) time instead of O(n).
//
// The flagship type is PreparedPolygon, which builds an index.RTree of the
// polygon's edges. Construction is O(n log n); the cost is paid once and
// reused across many ContainsPoint / IntersectsEnvelope calls. This pattern
// is the standard JTS/GEOS optimisation for workloads that test many points
// (or candidate envelopes) against the same polygon — overlay candidate
// pruning, point-cloud classification, viewshed clipping, etc.
//
// Prepared geometries are immutable after construction. All query methods
// are safe for concurrent use from multiple goroutines without further
// synchronisation; the underlying R-tree's read path takes only a read lock,
// and no mutating writes occur once Polygon returns.
//
// v0.1 scope: planar kernel only. Spherical / geodesic prepared variants
// are a follow-up; the planar ray-cast assumed by ContainsPoint is not
// valid on the sphere.
package prepare
