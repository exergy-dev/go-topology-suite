package spherical

import (
	"github.com/golang/geo/s2"
	"github.com/robert-malhotra/go-topology-suite/geom"
)

// Disjoint returns true if g1 and g2 have no points in common.
// Works with any geometry type combination using spherical geometry.
func Disjoint(g1, g2 geom.Geometry) bool {
	return !Intersects(g1, g2)
}

// Within returns true if g1 is completely within g2.
// This is the inverse of Contains: Within(a, b) == Contains(b, a).
func Within(g1, g2 geom.Geometry) bool {
	return Contains(g2, g1)
}

// Overlaps returns true if g1 and g2 overlap.
// Geometries overlap if they have the same dimension, intersect,
// and neither contains the other.
func Overlaps(g1, g2 geom.Geometry) bool {
	if isEmptyGeometry(g1) || isEmptyGeometry(g2) {
		return false
	}

	if g1.Dimension() != g2.Dimension() {
		return false
	}

	if !Intersects(g1, g2) {
		return false
	}

	if Contains(g1, g2) || Contains(g2, g1) {
		return false
	}

	return true
}

// Touches returns true if g1 and g2 touch at their boundaries only.
// They share boundary points but not interior points.
func Touches(g1, g2 geom.Geometry) bool {
	if isEmptyGeometry(g1) || isEmptyGeometry(g2) {
		return false
	}

	hasCommonPoint := false

	coords1 := g1.Coordinates()
	coords2 := g2.Coordinates()

	for _, c := range coords1 {
		loc := locatePointSpherical(c, g2)
		if loc == geom.LocationBoundary {
			hasCommonPoint = true
		} else if loc == geom.LocationInterior {
			return false
		}
	}

	for _, c := range coords2 {
		loc := locatePointSpherical(c, g1)
		if loc == geom.LocationBoundary {
			hasCommonPoint = true
		} else if loc == geom.LocationInterior {
			return false
		}
	}

	if g1.Dimension() == geom.DimensionArea && g2.Dimension() == geom.DimensionArea {
		if hasSphericalPolygonInteriorIntersection(g1, g2) {
			return false
		}
	}

	return hasCommonPoint
}

// Crosses returns true if g1 crosses g2.
// Geometries cross if they have some but not all interior points in common,
// and the dimension of the intersection is less than the maximum dimension of the inputs.
func Crosses(g1, g2 geom.Geometry) bool {
	if isEmptyGeometry(g1) || isEmptyGeometry(g2) {
		return false
	}

	if !g1.Envelope().Intersects(g2.Envelope()) {
		return false
	}

	dim1 := g1.Dimension()
	dim2 := g2.Dimension()

	if dim1 == dim2 && (dim1 == geom.DimensionPoint || dim1 == geom.DimensionArea) {
		return false
	}

	if dim1 == geom.DimensionLine && dim2 == geom.DimensionLine {
		return linesCrossSpherical(g1, g2)
	}

	if (dim1 == geom.DimensionPoint && dim2 == geom.DimensionLine) ||
		(dim1 == geom.DimensionLine && dim2 == geom.DimensionPoint) {
		return false
	}

	if dim1 == geom.DimensionLine && dim2 == geom.DimensionArea {
		return lineCrossesAreaSpherical(g1, g2)
	}
	if dim1 == geom.DimensionArea && dim2 == geom.DimensionLine {
		return lineCrossesAreaSpherical(g2, g1)
	}

	return false
}

// Covers returns true if no point of g2 is outside g1.
// Covers(a, b) is true if every point of b is in the interior or boundary of a.
func Covers(g1, g2 geom.Geometry) bool {
	if isEmptyGeometry(g1) || isEmptyGeometry(g2) {
		return false
	}

	if !g1.Envelope().Intersects(g2.Envelope()) {
		return false
	}

	coords2 := g2.Coordinates()
	for _, c := range coords2 {
		if locatePointSpherical(c, g1) == geom.LocationExterior {
			return false
		}
	}
	return true
}

// CoveredBy returns true if g1 is covered by g2.
// CoveredBy(a, b) == Covers(b, a).
func CoveredBy(g1, g2 geom.Geometry) bool {
	return Covers(g2, g1)
}

// Equals returns true if g1 and g2 are topologically equal.
// Two geometries are equal if they have the same set of points.
func Equals(g1, g2 geom.Geometry) bool {
	if isNilGeometry(g1) && isNilGeometry(g2) {
		return true
	}
	if isNilGeometry(g1) || isNilGeometry(g2) {
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

	return Covers(g1, g2) && Covers(g2, g1)
}

// locatePointSpherical returns the location of a point relative to a geometry
// using spherical geometry. Returns Interior, Boundary, or Exterior.
func locatePointSpherical(p geom.Coordinate, g geom.Geometry) geom.Location {
	switch v := g.(type) {
	case *geom.Point:
		if v.IsEmpty() {
			return geom.LocationExterior
		}
		s2Point1 := s2.PointFromLatLng(ToS2LatLng(p))
		s2Point2 := ToS2Point(v)
		if s2Point1.Distance(s2Point2).Radians()*EarthMeanRadius < defaultToleranceMeters {
			return geom.LocationInterior
		}
		return geom.LocationExterior

	case *geom.LineString:
		if v.IsEmpty() {
			return geom.LocationExterior
		}
		coords := v.Coordinates()
		s2Point := s2.PointFromLatLng(ToS2LatLng(p))

		if !v.IsClosed() {
			s2Start := s2.PointFromLatLng(ToS2LatLng(coords[0]))
			s2End := s2.PointFromLatLng(ToS2LatLng(coords[len(coords)-1]))
			if s2Point.Distance(s2Start).Radians()*EarthMeanRadius < defaultToleranceMeters ||
				s2Point.Distance(s2End).Radians()*EarthMeanRadius < defaultToleranceMeters {
				return geom.LocationBoundary
			}
		}

		polyline := ToS2Polyline(v)
		if polyline != nil {
			closest, _ := polyline.Project(s2Point)
			if s2Point.Distance(closest).Radians()*EarthMeanRadius < defaultToleranceMeters {
				return geom.LocationInterior
			}
		}
		return geom.LocationExterior

	case *geom.LinearRing:
		if v.IsEmpty() {
			return geom.LocationExterior
		}
		loop := ToS2Loop(v)
		if loop == nil {
			return geom.LocationExterior
		}
		s2Point := s2.PointFromLatLng(ToS2LatLng(p))

		coords := v.Coordinates()
		for i := 1; i < len(coords); i++ {
			ll1 := ToS2LatLng(coords[i-1])
			ll2 := ToS2LatLng(coords[i])
			p1 := s2.PointFromLatLng(ll1)
			p2 := s2.PointFromLatLng(ll2)

			if distanceToEdge(s2Point, p1, p2) <= defaultToleranceMeters {
				return geom.LocationBoundary
			}
		}

		if loop.ContainsPoint(s2Point) {
			return geom.LocationInterior
		}
		return geom.LocationExterior

	case *geom.Polygon:
		if v.IsEmpty() {
			return geom.LocationExterior
		}
		s2Poly := ToS2Polygon(v)
		if s2Poly == nil {
			return geom.LocationExterior
		}
		s2Point := s2.PointFromLatLng(ToS2LatLng(p))

		shell := v.ExteriorRing()
		if shell != nil {
			shellCoords := shell.Coordinates()
			for i := 1; i < len(shellCoords); i++ {
				ll1 := ToS2LatLng(shellCoords[i-1])
				ll2 := ToS2LatLng(shellCoords[i])
				p1 := s2.PointFromLatLng(ll1)
				p2 := s2.PointFromLatLng(ll2)

				if distanceToEdge(s2Point, p1, p2) <= defaultToleranceMeters {
					return geom.LocationBoundary
				}
			}
		}

		for i := 0; i < v.NumInteriorRings(); i++ {
			hole := v.InteriorRingN(i)
			holeCoords := hole.Coordinates()
			for j := 1; j < len(holeCoords); j++ {
				ll1 := ToS2LatLng(holeCoords[j-1])
				ll2 := ToS2LatLng(holeCoords[j])
				p1 := s2.PointFromLatLng(ll1)
				p2 := s2.PointFromLatLng(ll2)

				if distanceToEdge(s2Point, p1, p2) <= defaultToleranceMeters {
					return geom.LocationBoundary
				}
			}
		}

		if s2Poly.ContainsPoint(s2Point) {
			return geom.LocationInterior
		}
		return geom.LocationExterior

	case *geom.MultiPoint:
		for i := 0; i < v.NumGeometries(); i++ {
			loc := locatePointSpherical(p, v.GeometryN(i))
			if loc != geom.LocationExterior {
				return loc
			}
		}
		return geom.LocationExterior

	case *geom.MultiLineString:
		for i := 0; i < v.NumGeometries(); i++ {
			loc := locatePointSpherical(p, v.GeometryN(i))
			if loc != geom.LocationExterior {
				return loc
			}
		}
		return geom.LocationExterior

	case *geom.MultiPolygon:
		for i := 0; i < v.NumGeometries(); i++ {
			loc := locatePointSpherical(p, v.GeometryN(i))
			if loc != geom.LocationExterior {
				return loc
			}
		}
		return geom.LocationExterior

	case *geom.GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			loc := locatePointSpherical(p, v.GeometryN(i))
			if loc != geom.LocationExterior {
				return loc
			}
		}
		return geom.LocationExterior
	}

	return geom.LocationExterior
}

// hasSphericalPolygonInteriorIntersection checks if two polygon geometries
// have interior points in common (not just boundary touching).
func hasSphericalPolygonInteriorIntersection(g1, g2 geom.Geometry) bool {
	polys1 := getSphericalPolygons(g1)
	polys2 := getSphericalPolygons(g2)

	for _, p1 := range polys1 {
		for _, p2 := range polys2 {
			s2Poly1 := ToS2Polygon(p1)
			s2Poly2 := ToS2Polygon(p2)
			if s2Poly1 == nil || s2Poly2 == nil {
				continue
			}

			if !s2Poly1.Intersects(s2Poly2) {
				continue
			}

			env1 := p1.Envelope()
			env2 := p2.Envelope()
			if !env1.Intersects(env2) {
				continue
			}

			minX := env1.MinX
			if env2.MinX > minX {
				minX = env2.MinX
			}
			minY := env1.MinY
			if env2.MinY > minY {
				minY = env2.MinY
			}
			maxX := env1.MaxX
			if env2.MaxX < maxX {
				maxX = env2.MaxX
			}
			maxY := env1.MaxY
			if env2.MaxY < maxY {
				maxY = env2.MaxY
			}

			intersectEnv := geom.NewEnvelope(minX, minY, maxX, maxY)
			center := intersectEnv.Centre()

			loc1 := locatePointSpherical(center, p1)
			loc2 := locatePointSpherical(center, p2)

			if loc1 == geom.LocationInterior && loc2 == geom.LocationInterior {
				return true
			}

			if s2Poly1.Intersects(s2Poly2) &&
				!s2Poly1.Contains(s2Poly2) &&
				!s2Poly2.Contains(s2Poly1) {
				return true
			}
		}
	}

	return false
}

// getSphericalPolygons extracts all polygon geometries from a geometry.
func getSphericalPolygons(g geom.Geometry) []*geom.Polygon {
	var polys []*geom.Polygon

	switch v := g.(type) {
	case *geom.Polygon:
		if !v.IsEmpty() {
			polys = append(polys, v)
		}
	case *geom.MultiPolygon:
		for i := 0; i < v.NumGeometries(); i++ {
			if poly := v.GeometryN(i).(*geom.Polygon); !poly.IsEmpty() {
				polys = append(polys, poly)
			}
		}
	case *geom.GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			polys = append(polys, getSphericalPolygons(v.GeometryN(i))...)
		}
	}

	return polys
}

// linesCrossSpherical checks if two line geometries cross at a point (not overlap).
func linesCrossSpherical(g1, g2 geom.Geometry) bool {
	ls1 := geom.ExtractLineStrings(g1)
	ls2 := geom.ExtractLineStrings(g2)

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
	lines := geom.ExtractLineStrings(lineGeom)
	polys := geom.ExtractPolygons(areaGeom)

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
						return true
					}
				}
			}
		}
	}

	return false
}
