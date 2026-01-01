package spherical

import (
	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/golang/geo/s2"
)

// Crosses returns true if g1 crosses g2.
// Geometries cross if they have some but not all interior points in common,
// and the dimension of the intersection is less than the maximum dimension of the inputs.
//
// Rules:
// - Point/Point cannot cross
// - Point/Line: point is in interior of line (not endpoint)
// - Point/Polygon cannot cross
// - Line/Line: cross at a point (not overlap along segment)
// - Line/Polygon: line passes through interior AND exterior
// - Polygon/Polygon cannot cross (same dimension)
func Crosses(g1, g2 geom.Geometry) bool {
	if g1 == nil || g1.IsEmpty() || g2 == nil || g2.IsEmpty() {
		return false
	}

	// Quick envelope check
	if !g1.Envelope().Intersects(g2.Envelope()) {
		return false
	}

	dim1 := g1.Dimension()
	dim2 := g2.Dimension()

	// Point/Point and Area/Area cannot cross
	if dim1 == dim2 && (dim1 == geom.DimensionPoint || dim1 == geom.DimensionArea) {
		return false
	}

	// Line/Line: must have point intersection but not share a line segment
	if dim1 == geom.DimensionLine && dim2 == geom.DimensionLine {
		return linesCrossSpherical(g1, g2)
	}

	// Point/Line or Line/Point: points don't cross lines
	if (dim1 == geom.DimensionPoint && dim2 == geom.DimensionLine) ||
		(dim1 == geom.DimensionLine && dim2 == geom.DimensionPoint) {
		return false
	}

	// Line/Area: line must pass through interior and exterior
	if dim1 == geom.DimensionLine && dim2 == geom.DimensionArea {
		return lineCrossesAreaSpherical(g1, g2)
	}
	if dim1 == geom.DimensionArea && dim2 == geom.DimensionLine {
		return lineCrossesAreaSpherical(g2, g1)
	}

	return false
}

// Covers returns true if no point of g2 is outside g1.
// Similar to Contains but allows boundary-only intersection.
// Covers(a,b) is true if every point of b is in the interior OR boundary of a.
func Covers(g1, g2 geom.Geometry) bool {
	if g1 == nil || g1.IsEmpty() || g2 == nil || g2.IsEmpty() {
		return false
	}

	// Quick envelope check - use Intersects instead of ContainsEnvelope for boundary cases
	if !g1.Envelope().Intersects(g2.Envelope()) {
		return false
	}

	// Check that all coordinates of g2 are inside or on the boundary of g1
	coords2 := g2.Coordinates()
	for _, c := range coords2 {
		if !isPointInInteriorOrBoundarySpherical(c, g1) {
			return false
		}
	}
	return true
}

// CoveredBy returns true if g1 is covered by g2.
// This is the inverse: CoveredBy(a,b) == Covers(b,a)
func CoveredBy(g1, g2 geom.Geometry) bool {
	return Covers(g2, g1)
}

// Equals returns true if g1 and g2 are topologically equal.
// Two geometries are equal if they have the same set of points.
// This is equivalent to: Covers(a,b) && Covers(b,a)
func Equals(g1, g2 geom.Geometry) bool {
	if g1 == nil && g2 == nil {
		return true
	}
	if g1 == nil || g2 == nil {
		return false
	}

	if g1.GeometryType() != g2.GeometryType() {
		return false
	}

	if g1.IsEmpty() && g2.IsEmpty() {
		return true
	}

	if g1.IsEmpty() || g2.IsEmpty() {
		return false
	}

	// Check if each covers the other
	return Covers(g1, g2) && Covers(g2, g1)
}

// linesCrossSpherical checks if two line geometries cross at a point (not overlap).
func linesCrossSpherical(g1, g2 geom.Geometry) bool {
	ls1 := getLineStringsFromGeometry(g1)
	ls2 := getLineStringsFromGeometry(g2)

	for _, l1 := range ls1 {
		for _, l2 := range ls2 {
			if lineStringsCrossSpherical(l1, l2) {
				return true
			}
		}
	}
	return false
}

// lineStringsCrossSpherical checks if two linestrings properly cross.
func lineStringsCrossSpherical(ls1, ls2 *geom.LineString) bool {
	coords1 := ls1.Coordinates()
	coords2 := ls2.Coordinates()

	// Check each segment pair for proper crossing
	for i := 1; i < len(coords1); i++ {
		ll1a := ToS2LatLng(coords1[i-1])
		ll1b := ToS2LatLng(coords1[i])
		p1a := s2.PointFromLatLng(ll1a)
		p1b := s2.PointFromLatLng(ll1b)

		for j := 1; j < len(coords2); j++ {
			ll2a := ToS2LatLng(coords2[j-1])
			ll2b := ToS2LatLng(coords2[j])
			p2a := s2.PointFromLatLng(ll2a)
			p2b := s2.PointFromLatLng(ll2b)

			// Use S2's CrossingSign to detect crossings
			// CrossingSign: -1 = don't cross, 0 = edge case, 1 = cross
			// For "crosses", we want proper crossings (not sharing endpoints)
			sign := s2.CrossingSign(p1a, p1b, p2a, p2b)
			if sign == s2.Cross {
				return true
			}
		}
	}
	return false
}

// lineCrossesAreaSpherical checks if a line crosses a polygon
// (has points both inside and outside).
func lineCrossesAreaSpherical(lineGeom, areaGeom geom.Geometry) bool {
	lines := getLineStringsFromGeometry(lineGeom)
	polys := getPolygonsFromGeometry(areaGeom)

	for _, ls := range lines {
		for _, poly := range polys {
			if lineStringCrossesPolygonSpherical(ls, poly) {
				return true
			}
		}
	}
	return false
}

// lineStringCrossesPolygonSpherical checks if a linestring crosses a polygon.
func lineStringCrossesPolygonSpherical(ls *geom.LineString, poly *geom.Polygon) bool {
	s2Poly := ToS2Polygon(poly)
	if s2Poly == nil {
		return false
	}

	coords := ls.Coordinates()
	hasInside := false
	hasOutside := false

	// Check each vertex
	for _, c := range coords {
		ll := ToS2LatLng(c)
		point := s2.PointFromLatLng(ll)
		if s2Poly.ContainsPoint(point) {
			hasInside = true
		} else {
			hasOutside = true
		}
		if hasInside && hasOutside {
			return true
		}
	}

	// If we didn't find both inside and outside points from vertices,
	// check if line segments cross the polygon boundary
	if !hasInside || !hasOutside {
		shell := poly.ExteriorRing()
		if shell != nil {
			shellCoords := shell.Coordinates()
			for i := 1; i < len(coords); i++ {
				ll1a := ToS2LatLng(coords[i-1])
				ll1b := ToS2LatLng(coords[i])
				p1a := s2.PointFromLatLng(ll1a)
				p1b := s2.PointFromLatLng(ll1b)

				for j := 1; j < len(shellCoords); j++ {
					ll2a := ToS2LatLng(shellCoords[j-1])
					ll2b := ToS2LatLng(shellCoords[j])
					p2a := s2.PointFromLatLng(ll2a)
					p2b := s2.PointFromLatLng(ll2b)

					sign := s2.CrossingSign(p1a, p1b, p2a, p2b)
					if sign == s2.Cross {
						// Line crosses boundary, so it has points inside and outside
						return true
					}
				}
			}
		}
	}

	return false
}

// isPointInInteriorOrBoundarySpherical checks if a point is inside or on the boundary of a geometry.
func isPointInInteriorOrBoundarySpherical(coord geom.Coordinate, g geom.Geometry) bool {
	switch v := g.(type) {
	case *geom.Point:
		if v.IsEmpty() {
			return false
		}
		// Check if coordinates are approximately equal (using small tolerance)
		ll1 := ToS2LatLng(coord)
		ll2 := ToS2LatLng(geom.NewCoordinate(v.X(), v.Y()))
		p1 := s2.PointFromLatLng(ll1)
		p2 := s2.PointFromLatLng(ll2)
		return p1.Distance(p2).Radians()*EarthMeanRadius < 0.01 // 1cm tolerance

	case *geom.LineString:
		if v.IsEmpty() {
			return false
		}
		// Check if point is on the linestring
		point := geom.NewPoint(coord.X, coord.Y)
		return PointOnLineString(point, v, 0.01) // 1cm tolerance

	case *geom.LinearRing:
		if v.IsEmpty() {
			return false
		}
		// Check if point is on the ring
		point := geom.NewPoint(coord.X, coord.Y)
		return PointOnRing(point, v, 0.01) // 1cm tolerance

	case *geom.Polygon:
		if v.IsEmpty() {
			return false
		}
		// Use S2 containment test - check both interior and boundary
		s2Poly := ToS2Polygon(v)
		if s2Poly == nil {
			return false
		}
		ll := ToS2LatLng(coord)
		point := s2.PointFromLatLng(ll)

		// ContainsPoint returns true for interior points only
		// We need to also check if point is on the boundary
		if s2Poly.ContainsPoint(point) {
			return true
		}

		// Check if point is on the boundary using a small tolerance
		// We'll check if the point is on any edge of the polygon
		shell := v.ExteriorRing()
		if shell != nil {
			p := geom.NewPoint(coord.X, coord.Y)
			if PointOnRing(p, shell, 1.0) { // 1 meter tolerance
				return true
			}
		}

		// Check holes
		for i := 0; i < v.NumInteriorRings(); i++ {
			hole := v.InteriorRingN(i)
			if hole != nil {
				p := geom.NewPoint(coord.X, coord.Y)
				if PointOnRing(p, hole, 1.0) {
					return true
				}
			}
		}

		return false

	case *geom.MultiPoint:
		for i := 0; i < v.NumGeometries(); i++ {
			if isPointInInteriorOrBoundarySpherical(coord, v.GeometryN(i)) {
				return true
			}
		}
		return false

	case *geom.MultiLineString:
		for i := 0; i < v.NumGeometries(); i++ {
			if isPointInInteriorOrBoundarySpherical(coord, v.GeometryN(i)) {
				return true
			}
		}
		return false

	case *geom.MultiPolygon:
		for i := 0; i < v.NumGeometries(); i++ {
			if isPointInInteriorOrBoundarySpherical(coord, v.GeometryN(i)) {
				return true
			}
		}
		return false

	case *geom.GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			if isPointInInteriorOrBoundarySpherical(coord, v.GeometryN(i)) {
				return true
			}
		}
		return false
	}

	return false
}

// getLineStringsFromGeometry extracts all LineStrings from a geometry.
func getLineStringsFromGeometry(g geom.Geometry) []*geom.LineString {
	var result []*geom.LineString
	switch v := g.(type) {
	case *geom.LineString:
		result = append(result, v)
	case *geom.MultiLineString:
		for i := 0; i < v.NumGeometries(); i++ {
			result = append(result, v.GeometryN(i).(*geom.LineString))
		}
	case *geom.GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			result = append(result, getLineStringsFromGeometry(v.GeometryN(i))...)
		}
	}
	return result
}

// getPolygonsFromGeometry extracts all Polygons from a geometry.
func getPolygonsFromGeometry(g geom.Geometry) []*geom.Polygon {
	var result []*geom.Polygon
	switch v := g.(type) {
	case *geom.Polygon:
		result = append(result, v)
	case *geom.MultiPolygon:
		for i := 0; i < v.NumGeometries(); i++ {
			result = append(result, v.GeometryN(i).(*geom.Polygon))
		}
	case *geom.GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			result = append(result, getPolygonsFromGeometry(v.GeometryN(i))...)
		}
	}
	return result
}
