package spherical

import (
	"github.com/go-topology-suite/gts/geom"
	"github.com/golang/geo/s2"
)

// ToS2Point converts a GTS Point to S2 Point.
// Coordinates are expected as (lon, lat) in degrees.
func ToS2Point(p *geom.Point) s2.Point {
	ll := s2.LatLngFromDegrees(p.Y(), p.X())
	return s2.PointFromLatLng(ll)
}

// FromS2Point converts S2 Point to GTS Point.
// Returns coordinates as (lon, lat) in degrees.
func FromS2Point(p s2.Point) *geom.Point {
	ll := s2.LatLngFromPoint(p)
	return geom.NewPoint(ll.Lng.Degrees(), ll.Lat.Degrees())
}

// ToS2LatLng converts a GTS coordinate to S2 LatLng.
// The coordinate X is interpreted as longitude, Y as latitude (both in degrees).
func ToS2LatLng(c geom.Coordinate) s2.LatLng {
	return s2.LatLngFromDegrees(c.Y, c.X)
}

// FromS2LatLng converts S2 LatLng to GTS Coordinate.
// Returns a coordinate with X = longitude, Y = latitude (both in degrees).
func FromS2LatLng(ll s2.LatLng) geom.Coordinate {
	return geom.NewCoordinate(ll.Lng.Degrees(), ll.Lat.Degrees())
}

// ToS2Polyline converts a GTS LineString to S2 Polyline.
// Returns nil if the linestring is empty or invalid.
func ToS2Polyline(ls *geom.LineString) *s2.Polyline {
	if ls == nil || ls.IsEmpty() {
		return nil
	}

	coords := ls.Coordinates()
	points := make([]s2.Point, len(coords))
	for i, c := range coords {
		ll := ToS2LatLng(c)
		points[i] = s2.PointFromLatLng(ll)
	}

	polyline := s2.Polyline(points)
	return &polyline
}

// FromS2Polyline converts S2 Polyline to GTS LineString.
// Returns an empty linestring if the polyline is nil or empty.
func FromS2Polyline(pl *s2.Polyline) *geom.LineString {
	if pl == nil || len(*pl) == 0 {
		return geom.NewLineStringEmpty()
	}

	coords := make(geom.CoordinateSequence, len(*pl))
	for i, p := range *pl {
		ll := s2.LatLngFromPoint(p)
		coords[i] = FromS2LatLng(ll)
	}

	return geom.NewLineString(coords)
}

// ToS2Loop converts a GTS LinearRing to S2 Loop.
// Note: S2 loops don't repeat the last point, so we exclude the closing coordinate.
// Returns nil if the ring is empty or invalid.
func ToS2Loop(ring *geom.LinearRing) *s2.Loop {
	if ring == nil || ring.IsEmpty() {
		return nil
	}

	coords := ring.Coordinates()
	if len(coords) < 4 {
		// Need at least 4 points (including closure) for a valid ring
		return nil
	}

	// S2 loops don't repeat the last point, so we exclude it
	points := make([]s2.Point, len(coords)-1)
	for i := 0; i < len(coords)-1; i++ {
		ll := ToS2LatLng(coords[i])
		points[i] = s2.PointFromLatLng(ll)
	}

	return s2.LoopFromPoints(points)
}

// ToS2Polygon converts a GTS Polygon to S2 Polygon.
// Returns nil if the polygon is empty or invalid.
func ToS2Polygon(poly *geom.Polygon) *s2.Polygon {
	if poly == nil || poly.IsEmpty() {
		return nil
	}

	// Convert exterior ring
	shell := ToS2Loop(poly.ExteriorRing())
	if shell == nil {
		return nil
	}

	// Convert holes
	loops := make([]*s2.Loop, 1+poly.NumInteriorRings())
	loops[0] = shell

	for i := 0; i < poly.NumInteriorRings(); i++ {
		hole := ToS2Loop(poly.InteriorRingN(i))
		if hole == nil {
			continue
		}
		loops[i+1] = hole
	}

	return s2.PolygonFromLoops(loops)
}

// FromS2Polygon converts S2 Polygon to GTS Polygon.
// Returns an empty polygon if the S2 polygon is nil or empty.
func FromS2Polygon(poly *s2.Polygon) *geom.Polygon {
	if poly == nil || poly.NumLoops() == 0 {
		return geom.NewPolygonEmpty()
	}

	// Convert the first loop as the exterior ring
	shell := fromS2Loop(poly.Loop(0))
	if shell == nil {
		return geom.NewPolygonEmpty()
	}

	// Convert remaining loops as holes
	holes := make([]*geom.LinearRing, 0, poly.NumLoops()-1)
	for i := 1; i < poly.NumLoops(); i++ {
		hole := fromS2Loop(poly.Loop(i))
		if hole != nil {
			holes = append(holes, hole)
		}
	}

	return geom.NewPolygon(shell, holes)
}

// fromS2Loop converts an S2 Loop to a GTS LinearRing.
// Adds the closing coordinate to make it a valid ring.
func fromS2Loop(loop *s2.Loop) *geom.LinearRing {
	if loop == nil || loop.NumVertices() < 3 {
		return nil
	}

	// S2 loops don't repeat the last point, so we need to add it
	coords := make(geom.CoordinateSequence, loop.NumVertices()+1)
	for i := 0; i < loop.NumVertices(); i++ {
		ll := s2.LatLngFromPoint(loop.Vertex(i))
		coords[i] = FromS2LatLng(ll)
	}
	// Close the ring
	coords[loop.NumVertices()] = coords[0]

	return geom.NewLinearRing(coords)
}
