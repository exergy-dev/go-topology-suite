package spherical

import (
	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/golang/geo/s2"
)

// GenericWithin returns true if g1 is completely within g2.
// This is the inverse of Contains: GenericWithin(a,b) == Contains(b,a)
// Works with any geometry type combination using spherical geometry.
func GenericWithin(g1, g2 geom.Geometry) bool {
	return Contains(g2, g1)
}

// GenericDisjoint returns true if g1 and g2 have no points in common.
// This is the inverse of Intersects.
// Works with any geometry type combination using spherical geometry.
func GenericDisjoint(g1, g2 geom.Geometry) bool {
	return !Intersects(g1, g2)
}

// GenericOverlaps returns true if g1 and g2 overlap.
// Geometries overlap if they:
// - Have the same dimension
// - Intersect
// - Neither contains the other
// Works with any geometry type combination using spherical geometry.
func GenericOverlaps(g1, g2 geom.Geometry) bool {
	if g1 == nil || g1.IsEmpty() || g2 == nil || g2.IsEmpty() {
		return false
	}

	// Must have same dimension
	if g1.Dimension() != g2.Dimension() {
		return false
	}

	// Must intersect
	if !Intersects(g1, g2) {
		return false
	}

	// Neither must contain the other
	if Contains(g1, g2) || Contains(g2, g1) {
		return false
	}

	return true
}

// GenericTouches returns true if g1 and g2 touch at their boundaries only.
// They share boundary points but not interior points.
// Works with any geometry type combination using spherical geometry.
func GenericTouches(g1, g2 geom.Geometry) bool {
	if g1 == nil || g1.IsEmpty() || g2 == nil || g2.IsEmpty() {
		return false
	}

	// Must have boundary intersection but no interior intersection
	hasCommonPoint := false
	hasInteriorIntersection := false

	coords1 := g1.Coordinates()
	coords2 := g2.Coordinates()

	// Check if any point of g1 is in the interior of g2
	for _, c := range coords1 {
		loc := locatePointSpherical(c, g2)
		if loc == geom.LocationBoundary {
			hasCommonPoint = true
		} else if loc == geom.LocationInterior {
			hasInteriorIntersection = true
			break
		}
	}

	if hasInteriorIntersection {
		return false
	}

	// Check if any point of g2 is in the interior of g1
	for _, c := range coords2 {
		loc := locatePointSpherical(c, g1)
		if loc == geom.LocationBoundary {
			hasCommonPoint = true
		} else if loc == geom.LocationInterior {
			hasInteriorIntersection = true
			break
		}
	}

	if hasInteriorIntersection {
		return false
	}

	// For polygons, also check if boundaries intersect causing interior overlap
	if g1.Dimension() == geom.DimensionArea && g2.Dimension() == geom.DimensionArea {
		if hasSphericalPolygonInteriorIntersection(g1, g2) {
			return false
		}
	}

	return hasCommonPoint
}

// locatePointSpherical returns the location of a point relative to a geometry
// using spherical geometry. Returns Interior, Boundary, or Exterior.
func locatePointSpherical(p geom.Coordinate, g geom.Geometry) geom.Location {
	switch v := g.(type) {
	case *geom.Point:
		if v.IsEmpty() {
			return geom.LocationExterior
		}
		// Check if points are coincident (within small tolerance)
		s2Point1 := s2.PointFromLatLng(ToS2LatLng(p))
		s2Point2 := ToS2Point(v)
		dist := s2Point1.Distance(s2Point2)
		if dist.Radians()*EarthMeanRadius < 1.0 { // 1 meter tolerance
			return geom.LocationInterior
		}
		return geom.LocationExterior

	case *geom.LineString:
		if v.IsEmpty() {
			return geom.LocationExterior
		}
		coords := v.Coordinates()
		s2Point := s2.PointFromLatLng(ToS2LatLng(p))

		// Check endpoints (boundary)
		if !v.IsClosed() {
			s2Start := s2.PointFromLatLng(ToS2LatLng(coords[0]))
			s2End := s2.PointFromLatLng(ToS2LatLng(coords[len(coords)-1]))
			if s2Point.Distance(s2Start).Radians()*EarthMeanRadius < 1.0 ||
				s2Point.Distance(s2End).Radians()*EarthMeanRadius < 1.0 {
				return geom.LocationBoundary
			}
		}

		// Check if point is on the linestring
		polyline := ToS2Polyline(v)
		if polyline != nil {
			closest, _ := polyline.Project(s2Point)
			dist := s2Point.Distance(closest)
			if dist.Radians()*EarthMeanRadius < 1.0 { // 1 meter tolerance
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

		// Check if on boundary
		coords := v.Coordinates()
		for i := 1; i < len(coords); i++ {
			ll1 := ToS2LatLng(coords[i-1])
			ll2 := ToS2LatLng(coords[i])
			p1 := s2.PointFromLatLng(ll1)
			p2 := s2.PointFromLatLng(ll2)

			dist := distanceToEdge(s2Point, p1, p2)
			if dist <= 1.0 { // 1 meter tolerance
				return geom.LocationBoundary
			}
		}

		// Check if inside
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

		// Check if on boundary (exterior ring or any hole)
		shell := v.ExteriorRing()
		if shell != nil {
			shellCoords := shell.Coordinates()
			for i := 1; i < len(shellCoords); i++ {
				ll1 := ToS2LatLng(shellCoords[i-1])
				ll2 := ToS2LatLng(shellCoords[i])
				p1 := s2.PointFromLatLng(ll1)
				p2 := s2.PointFromLatLng(ll2)

				dist := distanceToEdge(s2Point, p1, p2)
				if dist <= 1.0 { // 1 meter tolerance
					return geom.LocationBoundary
				}
			}
		}

		// Check holes
		for i := 0; i < v.NumInteriorRings(); i++ {
			hole := v.InteriorRingN(i)
			holeCoords := hole.Coordinates()
			for j := 1; j < len(holeCoords); j++ {
				ll1 := ToS2LatLng(holeCoords[j-1])
				ll2 := ToS2LatLng(holeCoords[j])
				p1 := s2.PointFromLatLng(ll1)
				p2 := s2.PointFromLatLng(ll2)

				dist := distanceToEdge(s2Point, p1, p2)
				if dist <= 1.0 { // 1 meter tolerance
					return geom.LocationBoundary
				}
			}
		}

		// Check if in interior
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

			// Check if they intersect
			if !s2Poly1.Intersects(s2Poly2) {
				continue
			}

			// Sample point from the interior of the envelope intersection
			env1 := p1.Envelope()
			env2 := p2.Envelope()
			if !env1.Intersects(env2) {
				continue
			}

			// Create intersection envelope
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

			// Also check if edges properly cross (not just touch)
			// If the polygons intersect and we can find a sample interior point in both,
			// they have interior intersection
			// For spherical geometry, if they intersect but don't contain each other,
			// and are both polygons, they likely have interior intersection
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
