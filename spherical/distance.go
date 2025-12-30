package spherical

import (
	"github.com/go-topology-suite/gts/geom"
	"github.com/golang/geo/s2"
)

// EarthMeanRadius is the mean radius of Earth in meters (WGS84).
// This is used to convert angular distances to linear distances.
const EarthMeanRadius = 6371008.8

// Distance returns the geodesic distance in meters between two WGS84 points.
// The points are expected to have coordinates in (longitude, latitude) format in degrees.
// Returns 0 if either point is nil or empty.
func Distance(p1, p2 *geom.Point) float64 {
	if p1 == nil || p2 == nil || p1.IsEmpty() || p2.IsEmpty() {
		return 0
	}

	ll1 := ToS2LatLng(p1.Coordinate())
	ll2 := ToS2LatLng(p2.Coordinate())

	// Get angular distance in radians
	angle := ll1.Distance(ll2)

	// Convert to meters
	return angle.Radians() * EarthMeanRadius
}

// DistanceCoords returns the geodesic distance in meters between two coordinates.
// Parameters are (lon1, lat1, lon2, lat2) in degrees.
func DistanceCoords(lon1, lat1, lon2, lat2 float64) float64 {
	ll1 := s2.LatLngFromDegrees(lat1, lon1)
	ll2 := s2.LatLngFromDegrees(lat2, lon2)

	angle := ll1.Distance(ll2)
	return angle.Radians() * EarthMeanRadius
}

// Length returns the total geodesic length of a LineString in meters.
// Returns 0 if the linestring is nil, empty, or has fewer than 2 points.
func Length(ls *geom.LineString) float64 {
	if ls == nil || ls.IsEmpty() || ls.NumPoints() < 2 {
		return 0
	}

	polyline := ToS2Polyline(ls)
	if polyline == nil {
		return 0
	}

	// S2 polyline length returns angular distance in radians
	angleRadians := polyline.Length().Radians()
	return angleRadians * EarthMeanRadius
}

// Perimeter returns the total geodesic perimeter of a Polygon in meters.
// This includes the exterior ring and all interior rings (holes).
// Returns 0 if the polygon is nil or empty.
func Perimeter(poly *geom.Polygon) float64 {
	if poly == nil || poly.IsEmpty() {
		return 0
	}

	totalLength := 0.0

	// Add exterior ring length
	if shell := poly.ExteriorRing(); shell != nil && !shell.IsEmpty() {
		totalLength += ringLength(shell)
	}

	// Add hole lengths
	for i := 0; i < poly.NumInteriorRings(); i++ {
		if hole := poly.InteriorRingN(i); hole != nil && !hole.IsEmpty() {
			totalLength += ringLength(hole)
		}
	}

	return totalLength
}

// ringLength calculates the geodesic length of a ring.
func ringLength(ring *geom.LinearRing) float64 {
	if ring == nil || ring.IsEmpty() || ring.NumPoints() < 2 {
		return 0
	}

	coords := ring.Coordinates()
	if len(coords) < 2 {
		return 0
	}

	totalLength := 0.0
	for i := 1; i < len(coords); i++ {
		ll1 := ToS2LatLng(coords[i-1])
		ll2 := ToS2LatLng(coords[i])
		angle := ll1.Distance(ll2)
		totalLength += angle.Radians() * EarthMeanRadius
	}

	return totalLength
}

// DistanceToLineString returns the minimum distance from a point to a linestring in meters.
// Returns 0 if the point is on the linestring, or if either geometry is nil/empty.
func DistanceToLineString(p *geom.Point, ls *geom.LineString) float64 {
	if p == nil || p.IsEmpty() || ls == nil || ls.IsEmpty() {
		return 0
	}

	s2Point := ToS2Point(p)
	polyline := ToS2Polyline(ls)
	if polyline == nil {
		return 0
	}

	// Find closest point on polyline
	projectedPoint, _ := polyline.Project(s2Point)

	// Calculate angular distance between original point and projected point
	angle := s2Point.Distance(projectedPoint)
	return angle.Radians() * EarthMeanRadius
}

// DistanceToPolygon returns the minimum distance from a point to a polygon in meters.
// Returns 0 if the point is inside or on the polygon boundary.
func DistanceToPolygon(p *geom.Point, poly *geom.Polygon) float64 {
	if p == nil || p.IsEmpty() || poly == nil || poly.IsEmpty() {
		return 0
	}

	s2Point := ToS2Point(p)
	s2Poly := ToS2Polygon(poly)
	if s2Poly == nil {
		return 0
	}

	// If point is inside, distance is 0
	if s2Poly.ContainsPoint(s2Point) {
		return 0
	}

	// Find minimum distance to boundary
	minDist := float64(1e10) // Large initial value

	// Check distance to exterior ring
	if shell := poly.ExteriorRing(); shell != nil {
		dist := distanceToRing(s2Point, shell)
		if dist < minDist {
			minDist = dist
		}
	}

	// Note: For holes, if point is outside the polygon, we only need
	// to check the exterior ring distance
	return minDist
}

// distanceToRing calculates the minimum distance from a point to a ring.
func distanceToRing(p s2.Point, ring *geom.LinearRing) float64 {
	if ring == nil || ring.IsEmpty() {
		return 0
	}

	coords := ring.Coordinates()
	minDist := float64(1e10)

	for i := 1; i < len(coords); i++ {
		ll1 := ToS2LatLng(coords[i-1])
		ll2 := ToS2LatLng(coords[i])
		p1 := s2.PointFromLatLng(ll1)
		p2 := s2.PointFromLatLng(ll2)

		// Calculate distance to this edge
		dist := distanceToEdge(p, p1, p2)
		if dist < minDist {
			minDist = dist
		}
	}

	return minDist
}

// distanceToEdge calculates the minimum distance from a point to a great circle edge.
func distanceToEdge(p, a, b s2.Point) float64 {
	// Calculate distance to the edge using S2's edge distance
	distance := s2.DistanceFromSegment(p, a, b)
	return distance.Radians() * EarthMeanRadius
}
