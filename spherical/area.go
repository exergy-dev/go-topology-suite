package spherical

import (
	"math"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// Area returns the area of a Polygon in square meters on the WGS84 ellipsoid.
// The area is computed using spherical geometry on the mean Earth radius.
// Returns 0 if the polygon is nil or empty.
//
// The result is always positive, regardless of the winding order of the rings.
func Area(poly *geom.Polygon) float64 {
	return math.Abs(SignedArea(poly))
}

// SignedArea returns the signed area of a Polygon in square meters.
// The sign depends on the orientation:
//   - Positive for counter-clockwise exterior ring
//   - Negative for clockwise exterior ring
//
// Returns 0 if the polygon is nil or empty.
func SignedArea(poly *geom.Polygon) float64 {
	if poly == nil || poly.IsEmpty() {
		return 0
	}

	s2Poly := ToS2Polygon(poly)
	if s2Poly == nil {
		return 0
	}

	// S2 area is in steradians (solid angle)
	// Convert to square meters: steradians * radius^2
	areaSteradians := s2Poly.Area()
	return areaSteradians * EarthMeanRadius * EarthMeanRadius
}

// RingArea returns the area of a LinearRing in square meters.
// The result is always positive.
// Returns 0 if the ring is nil or empty.
func RingArea(ring *geom.LinearRing) float64 {
	return math.Abs(SignedRingArea(ring))
}

// SignedRingArea returns the signed area of a LinearRing in square meters.
// The sign depends on the orientation:
//   - Positive for counter-clockwise ring
//   - Negative for clockwise ring
//
// Returns 0 if the ring is nil or empty.
func SignedRingArea(ring *geom.LinearRing) float64 {
	if ring == nil || ring.IsEmpty() {
		return 0
	}

	loop := ToS2Loop(ring)
	if loop == nil {
		return 0
	}

	// S2 loop area is in steradians
	// Convert to square meters
	areaSteradians := loop.Area()
	return areaSteradians * EarthMeanRadius * EarthMeanRadius
}

// PolygonAreaWithHoles calculates the area of a polygon accounting for holes.
// This returns: exterior_area - sum(hole_areas).
// Returns 0 if the polygon is nil or empty.
func PolygonAreaWithHoles(poly *geom.Polygon) float64 {
	if poly == nil || poly.IsEmpty() {
		return 0
	}

	// Get exterior ring area
	totalArea := RingArea(poly.ExteriorRing())

	// Subtract hole areas
	for i := 0; i < poly.NumInteriorRings(); i++ {
		if hole := poly.InteriorRingN(i); hole != nil && !hole.IsEmpty() {
			totalArea -= RingArea(hole)
		}
	}

	return totalArea
}

// MultiPolygonArea returns the total area of all polygons in a collection.
// This is a convenience function for calculating area of multiple polygons.
func MultiPolygonArea(polygons []*geom.Polygon) float64 {
	totalArea := 0.0
	for _, poly := range polygons {
		if poly != nil && !poly.IsEmpty() {
			totalArea += Area(poly)
		}
	}
	return totalArea
}

// Centroid returns the spherical centroid of a polygon.
// The centroid is computed on the surface of the sphere.
// Returns an empty point if the polygon is nil or empty.
func Centroid(poly *geom.Polygon) *geom.Point {
	if poly == nil || poly.IsEmpty() {
		return geom.NewPointEmpty()
	}

	s2Poly := ToS2Polygon(poly)
	if s2Poly == nil {
		return geom.NewPointEmpty()
	}

	centroid := s2Poly.Centroid()
	return FromS2Point(centroid)
}

// RingCentroid returns the spherical centroid of a ring.
// Returns an empty point if the ring is nil or empty.
func RingCentroid(ring *geom.LinearRing) *geom.Point {
	if ring == nil || ring.IsEmpty() {
		return geom.NewPointEmpty()
	}

	loop := ToS2Loop(ring)
	if loop == nil {
		return geom.NewPointEmpty()
	}

	centroid := loop.Centroid()
	return FromS2Point(centroid)
}
