// Package geom provides core geometry types implementing the OGC Simple Features
// specification. It includes Point, LineString, Polygon, and their Multi* variants,
// along with supporting types like Coordinate, Envelope, and GeometryFactory.
//
// All geometry types implement the Geometry interface which provides standard
// operations like intersection testing, boundary computation, and spatial predicates.
//
// # Concurrency
//
// Geometry objects are safe for concurrent read access after construction.
// Multiple goroutines may call read-only methods (Envelope, Coordinates,
// IsEmpty, EqualsExact, String, etc.) concurrently without synchronization.
//
// Methods that mutate a geometry (SetSRID, ApplyCoordinateFilter) are NOT
// safe for concurrent use on the same geometry. Use Clone or Normalized to
// create independent copies when concurrent mutation is needed.
//
// Operation types in the operation/ packages are designed to be created
// per-operation and should not be shared between goroutines.
//
// Example usage:
//
//	// Create a polygon
//	seq, err := geom.NewCoordinateSequenceXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	shell := geom.NewLinearRing(seq)
//	polygon := geom.NewPolygon(shell, nil)
//
//	// Check if a point is inside
//	if polygon.ContainsPoint(geom.NewCoordinate(5, 5)) {
//	    fmt.Println("Point is inside the polygon")
//	}
package geom
