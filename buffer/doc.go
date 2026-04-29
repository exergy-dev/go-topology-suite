// Package buffer constructs offset polygons around geometries.
//
// # Scope (v0.1)
//
// This release supports buffering of:
//
//   - Point             → regular polygon approximating a circle
//   - LineString        → "thickened" polygon via parallel offsets and caps
//   - MultiPoint        → MultiPolygon of per-member buffers (no union yet)
//   - MultiLineString   → MultiPolygon of per-member buffers (no union yet)
//   - Polygon           → outward-grown / inward-shrunk polygon via the
//     overlay-NG (Greiner-Hormann) Union and parallel-offset rings
//   - MultiPolygon      → per-member polygon buffer, then pairwise Union
//
// GeometryCollection input is still rejected with an explicit error.
//
// # Polygon buffer limitations (v0.1)
//
// Polygon buffering is implemented on top of the v0.1 overlay-NG path
// (Greiner-Hormann). Known limitations carry over:
//
//   - Concave polygons with sharp reflex corners may produce a self-
//     intersecting offset ring; the result is then geometrically
//     approximate rather than exact.
//   - Coincident input edges in the original polygon (slivers, exact
//     boundary touches) may produce minor degeneracy in the offset.
//   - Negative-distance buffers that fully erode the polygon return an
//     empty polygon. Inputs near the collapse threshold may return a
//     vanishingly thin polygon rather than an empty one.
//
// Callers needing fully robust polygon buffers should preprocess holes
// separately or wait for the planned exact-arithmetic overlay (Phase 4).
//
// # Coordinate system
//
// Buffer is purely planar. The input distance is interpreted in the units
// of the geometry's coordinate reference system. Geographic CRSes
// (lat/lon degrees) will produce nonsense output — project to a metric CRS
// (UTM, Web Mercator with appropriate scale factor, etc.) before calling.
//
// # Validity
//
// For non-self-intersecting LineString input the result is a simple polygon.
// Self-intersecting input may yield a self-intersecting buffer; cleaning
// such output requires the union operation (overlay-NG). Callers requiring
// a guaranteed-simple buffer should pre-noding their LineString.
//
// # Options
//
// Cap and join styles, mitre limit, and arc resolution are controlled via
// functional options. Defaults: round caps, round joins, 8 segments per
// quadrant, mitre limit 5.0.
package buffer
