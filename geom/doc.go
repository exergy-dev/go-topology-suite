// Package geom provides core geometry types implementing the OGC Simple Features
// specification. It includes Point, LineString, Polygon, and their Multi* variants,
// along with supporting types like Coordinate, Envelope, and GeometryFactory.
//
// All geometry types implement the Geometry interface which provides standard
// operations like intersection testing, boundary computation, and spatial predicates.
//
// Example usage:
//
//	// Create a polygon
//	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
//	polygon := geom.NewPolygon(shell, nil)
//
//	// Check if a point is inside
//	if polygon.ContainsPoint(geom.NewCoordinate(5, 5)) {
//	    fmt.Println("Point is inside the polygon")
//	}
package geom
