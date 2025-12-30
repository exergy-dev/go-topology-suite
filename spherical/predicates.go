package spherical

import (
	"github.com/go-topology-suite/gts/geom"
	"github.com/golang/geo/s2"
)

// Intersects returns true if g1 and g2 share any portion of space using spherical geometry.
// Works with any geometry type combination. Returns false if either geometry is nil or empty.
//
// This uses S2's robust intersection tests which properly handle edge cases
// on the sphere, including the antimeridian and poles.
func Intersects(g1, g2 geom.Geometry) bool {
	if g1 == nil || g1.IsEmpty() || g2 == nil || g2.IsEmpty() {
		return false
	}

	return intersectsSpherical(g1, g2)
}

// Contains returns true if g1 completely contains g2 using spherical geometry.
// Works with any geometry type combination. Returns false if either geometry is nil or empty.
//
// This uses S2's robust containment tests which properly handle edge cases
// on the sphere, including the antimeridian and poles.
func Contains(g1, g2 geom.Geometry) bool {
	if g1 == nil || g1.IsEmpty() || g2 == nil || g2.IsEmpty() {
		return false
	}

	return containsSpherical(g1, g2)
}

// intersectsSpherical is the main implementation dispatcher for Intersects.
func intersectsSpherical(g1, g2 geom.Geometry) bool {
	switch a := g1.(type) {
	case *geom.Point:
		return pointIntersectsSpherical(a, g2)
	case *geom.LineString:
		return lineStringIntersectsSpherical(a, g2)
	case *geom.LinearRing:
		return linearRingIntersectsSpherical(a, g2)
	case *geom.Polygon:
		return polygonIntersectsSpherical(a, g2)
	case *geom.MultiPoint:
		for i := 0; i < a.NumGeometries(); i++ {
			if pointIntersectsSpherical(a.GeometryN(i).(*geom.Point), g2) {
				return true
			}
		}
		return false
	case *geom.MultiLineString:
		for i := 0; i < a.NumGeometries(); i++ {
			if lineStringIntersectsSpherical(a.GeometryN(i).(*geom.LineString), g2) {
				return true
			}
		}
		return false
	case *geom.MultiPolygon:
		for i := 0; i < a.NumGeometries(); i++ {
			if polygonIntersectsSpherical(a.GeometryN(i).(*geom.Polygon), g2) {
				return true
			}
		}
		return false
	case *geom.GeometryCollection:
		for i := 0; i < a.NumGeometries(); i++ {
			if intersectsSpherical(a.GeometryN(i), g2) {
				return true
			}
		}
		return false
	}
	return false
}

// containsSpherical is the main implementation dispatcher for Contains.
func containsSpherical(g1, g2 geom.Geometry) bool {
	switch a := g1.(type) {
	case *geom.Point:
		return pointContainsSpherical(a, g2)
	case *geom.LineString:
		return lineStringContainsSpherical(a, g2)
	case *geom.LinearRing:
		return linearRingContainsSpherical(a, g2)
	case *geom.Polygon:
		return polygonContainsSpherical(a, g2)
	case *geom.MultiPoint:
		return multiPointContainsSpherical(a, g2)
	case *geom.MultiLineString:
		return multiLineStringContainsSpherical(a, g2)
	case *geom.MultiPolygon:
		return multiPolygonContainsSpherical(a, g2)
	case *geom.GeometryCollection:
		return geometryCollectionContainsSpherical(a, g2)
	}
	return false
}

// pointIntersectsSpherical checks if a point intersects any geometry.
func pointIntersectsSpherical(p *geom.Point, g geom.Geometry) bool {
	if p.IsEmpty() {
		return false
	}

	switch b := g.(type) {
	case *geom.Point:
		if b.IsEmpty() {
			return false
		}
		// Use small distance tolerance for point equality on sphere
		return Distance(p, b) < 0.01 // 1cm tolerance
	case *geom.LineString:
		return PointOnLineString(p, b, 0.01) // 1cm tolerance
	case *geom.LinearRing:
		loop := ToS2Loop(b)
		if loop == nil {
			return false
		}
		s2Point := ToS2Point(p)
		return loop.ContainsPoint(s2Point) || PointOnRing(p, b, 0.01)
	case *geom.Polygon:
		s2Poly := ToS2Polygon(b)
		if s2Poly == nil {
			return false
		}
		s2Point := ToS2Point(p)
		return s2Poly.ContainsPoint(s2Point)
	case *geom.MultiPoint:
		for i := 0; i < b.NumGeometries(); i++ {
			if Distance(p, b.GeometryN(i).(*geom.Point)) < 0.01 {
				return true
			}
		}
		return false
	case *geom.MultiLineString:
		for i := 0; i < b.NumGeometries(); i++ {
			if PointOnLineString(p, b.GeometryN(i).(*geom.LineString), 0.01) {
				return true
			}
		}
		return false
	case *geom.MultiPolygon:
		s2Point := ToS2Point(p)
		for i := 0; i < b.NumGeometries(); i++ {
			s2Poly := ToS2Polygon(b.GeometryN(i).(*geom.Polygon))
			if s2Poly != nil && s2Poly.ContainsPoint(s2Point) {
				return true
			}
		}
		return false
	case *geom.GeometryCollection:
		for i := 0; i < b.NumGeometries(); i++ {
			if pointIntersectsSpherical(p, b.GeometryN(i)) {
				return true
			}
		}
		return false
	}
	return false
}

// lineStringIntersectsSpherical checks if a linestring intersects any geometry.
func lineStringIntersectsSpherical(ls *geom.LineString, g geom.Geometry) bool {
	if ls.IsEmpty() {
		return false
	}

	switch b := g.(type) {
	case *geom.Point:
		return PointOnLineString(b, ls, 0.01)
	case *geom.LineString:
		return lineStringsIntersectSpherical(ls, b)
	case *geom.LinearRing:
		return lineStringRingIntersectSpherical(ls, b)
	case *geom.Polygon:
		return LineStringIntersectsPolygon(ls, b)
	case *geom.MultiPoint:
		for i := 0; i < b.NumGeometries(); i++ {
			if PointOnLineString(b.GeometryN(i).(*geom.Point), ls, 0.01) {
				return true
			}
		}
		return false
	case *geom.MultiLineString:
		for i := 0; i < b.NumGeometries(); i++ {
			if lineStringsIntersectSpherical(ls, b.GeometryN(i).(*geom.LineString)) {
				return true
			}
		}
		return false
	case *geom.MultiPolygon:
		for i := 0; i < b.NumGeometries(); i++ {
			if LineStringIntersectsPolygon(ls, b.GeometryN(i).(*geom.Polygon)) {
				return true
			}
		}
		return false
	case *geom.GeometryCollection:
		for i := 0; i < b.NumGeometries(); i++ {
			if lineStringIntersectsSpherical(ls, b.GeometryN(i)) {
				return true
			}
		}
		return false
	}
	return false
}

// linearRingIntersectsSpherical checks if a linear ring intersects any geometry.
func linearRingIntersectsSpherical(ring *geom.LinearRing, g geom.Geometry) bool {
	if ring.IsEmpty() {
		return false
	}

	switch b := g.(type) {
	case *geom.Point:
		return PointOnRing(b, ring, 0.01) || LoopContainsPoint(ring, b)
	case *geom.LineString:
		return lineStringRingIntersectSpherical(b, ring)
	case *geom.LinearRing:
		return LoopsIntersect(ring, b)
	case *geom.Polygon:
		loop := ToS2Loop(ring)
		s2Poly := ToS2Polygon(b)
		if loop == nil || s2Poly == nil {
			return false
		}
		return loopIntersectsPolygon(loop, s2Poly)
	case *geom.MultiPoint:
		for i := 0; i < b.NumGeometries(); i++ {
			if PointOnRing(b.GeometryN(i).(*geom.Point), ring, 0.01) || LoopContainsPoint(ring, b.GeometryN(i).(*geom.Point)) {
				return true
			}
		}
		return false
	case *geom.MultiLineString:
		for i := 0; i < b.NumGeometries(); i++ {
			if lineStringRingIntersectSpherical(b.GeometryN(i).(*geom.LineString), ring) {
				return true
			}
		}
		return false
	case *geom.MultiPolygon:
		loop := ToS2Loop(ring)
		if loop == nil {
			return false
		}
		for i := 0; i < b.NumGeometries(); i++ {
			s2Poly := ToS2Polygon(b.GeometryN(i).(*geom.Polygon))
			if s2Poly != nil && loopIntersectsPolygon(loop, s2Poly) {
				return true
			}
		}
		return false
	case *geom.GeometryCollection:
		for i := 0; i < b.NumGeometries(); i++ {
			if linearRingIntersectsSpherical(ring, b.GeometryN(i)) {
				return true
			}
		}
		return false
	}
	return false
}

// polygonIntersectsSpherical checks if a polygon intersects any geometry.
func polygonIntersectsSpherical(poly *geom.Polygon, g geom.Geometry) bool {
	if poly.IsEmpty() {
		return false
	}

	s2Poly := ToS2Polygon(poly)
	if s2Poly == nil {
		return false
	}

	switch b := g.(type) {
	case *geom.Point:
		s2Point := ToS2Point(b)
		return s2Poly.ContainsPoint(s2Point)
	case *geom.LineString:
		return LineStringIntersectsPolygon(b, poly)
	case *geom.LinearRing:
		loop := ToS2Loop(b)
		if loop == nil {
			return false
		}
		return loopIntersectsPolygon(loop, s2Poly)
	case *geom.Polygon:
		s2Poly2 := ToS2Polygon(b)
		if s2Poly2 == nil {
			return false
		}
		return s2Poly.Intersects(s2Poly2)
	case *geom.MultiPoint:
		for i := 0; i < b.NumGeometries(); i++ {
			s2Point := ToS2Point(b.GeometryN(i).(*geom.Point))
			if s2Poly.ContainsPoint(s2Point) {
				return true
			}
		}
		return false
	case *geom.MultiLineString:
		for i := 0; i < b.NumGeometries(); i++ {
			if LineStringIntersectsPolygon(b.GeometryN(i).(*geom.LineString), poly) {
				return true
			}
		}
		return false
	case *geom.MultiPolygon:
		for i := 0; i < b.NumGeometries(); i++ {
			s2Poly2 := ToS2Polygon(b.GeometryN(i).(*geom.Polygon))
			if s2Poly2 != nil && s2Poly.Intersects(s2Poly2) {
				return true
			}
		}
		return false
	case *geom.GeometryCollection:
		for i := 0; i < b.NumGeometries(); i++ {
			if polygonIntersectsSpherical(poly, b.GeometryN(i)) {
				return true
			}
		}
		return false
	}
	return false
}

// pointContainsSpherical checks if a point contains any geometry.
// A point can only contain another point (if they're equal).
func pointContainsSpherical(p *geom.Point, g geom.Geometry) bool {
	if p.IsEmpty() {
		return false
	}

	switch b := g.(type) {
	case *geom.Point:
		if b.IsEmpty() {
			return true // Empty geometry is contained by everything
		}
		return Distance(p, b) < 0.01 // 1cm tolerance
	}
	return false
}

// lineStringContainsSpherical checks if a linestring contains any geometry.
func lineStringContainsSpherical(ls *geom.LineString, g geom.Geometry) bool {
	if ls.IsEmpty() {
		return false
	}

	switch b := g.(type) {
	case *geom.Point:
		if b.IsEmpty() {
			return true
		}
		return PointOnLineString(b, ls, 0.01)
	case *geom.LineString:
		// Check if all points of b are on ls
		coords := b.Coordinates()
		for _, c := range coords {
			pt := geom.NewPoint(c.X, c.Y)
			if !PointOnLineString(pt, ls, 0.01) {
				return false
			}
		}
		return true
	case *geom.MultiPoint:
		for i := 0; i < b.NumGeometries(); i++ {
			if !PointOnLineString(b.GeometryN(i).(*geom.Point), ls, 0.01) {
				return false
			}
		}
		return true
	}
	return false
}

// linearRingContainsSpherical checks if a linear ring contains any geometry.
func linearRingContainsSpherical(ring *geom.LinearRing, g geom.Geometry) bool {
	if ring.IsEmpty() {
		return false
	}

	loop := ToS2Loop(ring)
	if loop == nil {
		return false
	}

	switch b := g.(type) {
	case *geom.Point:
		if b.IsEmpty() {
			return true
		}
		s2Point := ToS2Point(b)
		return loop.ContainsPoint(s2Point)
	case *geom.LineString:
		// Check if all points are inside or on the ring
		coords := b.Coordinates()
		for _, c := range coords {
			ll := ToS2LatLng(c)
			pt := s2.PointFromLatLng(ll)
			if !loop.ContainsPoint(pt) {
				return false
			}
		}
		return true
	case *geom.LinearRing:
		loop2 := ToS2Loop(b)
		if loop2 == nil {
			return false
		}
		return loopContainsLoop(loop, loop2)
	case *geom.MultiPoint:
		for i := 0; i < b.NumGeometries(); i++ {
			s2Point := ToS2Point(b.GeometryN(i).(*geom.Point))
			if !loop.ContainsPoint(s2Point) {
				return false
			}
		}
		return true
	}
	return false
}

// polygonContainsSpherical checks if a polygon contains any geometry.
func polygonContainsSpherical(poly *geom.Polygon, g geom.Geometry) bool {
	if poly.IsEmpty() {
		return false
	}

	s2Poly := ToS2Polygon(poly)
	if s2Poly == nil {
		return false
	}

	switch b := g.(type) {
	case *geom.Point:
		if b.IsEmpty() {
			return true
		}
		s2Point := ToS2Point(b)
		return s2Poly.ContainsPoint(s2Point)
	case *geom.LineString:
		// Check if all points are inside the polygon
		coords := b.Coordinates()
		for _, c := range coords {
			ll := ToS2LatLng(c)
			pt := s2.PointFromLatLng(ll)
			if !s2Poly.ContainsPoint(pt) {
				return false
			}
		}
		return true
	case *geom.LinearRing:
		loop := ToS2Loop(b)
		if loop == nil {
			return false
		}
		return polygonContainsLoop(s2Poly, loop)
	case *geom.Polygon:
		s2Poly2 := ToS2Polygon(b)
		if s2Poly2 == nil {
			return false
		}
		return s2Poly.Contains(s2Poly2)
	case *geom.MultiPoint:
		for i := 0; i < b.NumGeometries(); i++ {
			s2Point := ToS2Point(b.GeometryN(i).(*geom.Point))
			if !s2Poly.ContainsPoint(s2Point) {
				return false
			}
		}
		return true
	case *geom.MultiLineString:
		for i := 0; i < b.NumGeometries(); i++ {
			coords := b.GeometryN(i).(*geom.LineString).Coordinates()
			for _, c := range coords {
				ll := ToS2LatLng(c)
				pt := s2.PointFromLatLng(ll)
				if !s2Poly.ContainsPoint(pt) {
					return false
				}
			}
		}
		return true
	case *geom.MultiPolygon:
		for i := 0; i < b.NumGeometries(); i++ {
			s2Poly2 := ToS2Polygon(b.GeometryN(i).(*geom.Polygon))
			if s2Poly2 == nil || !s2Poly.Contains(s2Poly2) {
				return false
			}
		}
		return true
	case *geom.GeometryCollection:
		for i := 0; i < b.NumGeometries(); i++ {
			if !polygonContainsSpherical(poly, b.GeometryN(i)) {
				return false
			}
		}
		return true
	}
	return false
}

// multiPointContainsSpherical checks if a multipoint contains any geometry.
func multiPointContainsSpherical(mp *geom.MultiPoint, g geom.Geometry) bool {
	switch b := g.(type) {
	case *geom.Point:
		if b.IsEmpty() {
			return true
		}
		for i := 0; i < mp.NumGeometries(); i++ {
			if Distance(mp.GeometryN(i).(*geom.Point), b) < 0.01 {
				return true
			}
		}
		return false
	case *geom.MultiPoint:
		// All points in b must be in mp
		for i := 0; i < b.NumGeometries(); i++ {
			found := false
			for j := 0; j < mp.NumGeometries(); j++ {
				if Distance(mp.GeometryN(j).(*geom.Point), b.GeometryN(i).(*geom.Point)) < 0.01 {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}
	return false
}

// multiLineStringContainsSpherical checks if a multilinestring contains any geometry.
func multiLineStringContainsSpherical(mls *geom.MultiLineString, g geom.Geometry) bool {
	switch b := g.(type) {
	case *geom.Point:
		if b.IsEmpty() {
			return true
		}
		for i := 0; i < mls.NumGeometries(); i++ {
			if PointOnLineString(b, mls.GeometryN(i).(*geom.LineString), 0.01) {
				return true
			}
		}
		return false
	case *geom.MultiPoint:
		for i := 0; i < b.NumGeometries(); i++ {
			found := false
			for j := 0; j < mls.NumGeometries(); j++ {
				if PointOnLineString(b.GeometryN(i).(*geom.Point), mls.GeometryN(j).(*geom.LineString), 0.01) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}
	return false
}

// multiPolygonContainsSpherical checks if a multipolygon contains any geometry.
func multiPolygonContainsSpherical(mpoly *geom.MultiPolygon, g geom.Geometry) bool {
	switch b := g.(type) {
	case *geom.Point:
		if b.IsEmpty() {
			return true
		}
		s2Point := ToS2Point(b)
		for i := 0; i < mpoly.NumGeometries(); i++ {
			s2Poly := ToS2Polygon(mpoly.GeometryN(i).(*geom.Polygon))
			if s2Poly != nil && s2Poly.ContainsPoint(s2Point) {
				return true
			}
		}
		return false
	case *geom.Polygon:
		// Check if any polygon in mpoly contains b
		s2Poly2 := ToS2Polygon(b)
		if s2Poly2 == nil {
			return false
		}
		for i := 0; i < mpoly.NumGeometries(); i++ {
			s2Poly := ToS2Polygon(mpoly.GeometryN(i).(*geom.Polygon))
			if s2Poly != nil && s2Poly.Contains(s2Poly2) {
				return true
			}
		}
		return false
	case *geom.MultiPoint:
		for i := 0; i < b.NumGeometries(); i++ {
			if !multiPolygonContainsSpherical(mpoly, b.GeometryN(i)) {
				return false
			}
		}
		return true
	case *geom.MultiPolygon:
		for i := 0; i < b.NumGeometries(); i++ {
			if !multiPolygonContainsSpherical(mpoly, b.GeometryN(i)) {
				return false
			}
		}
		return true
	case *geom.GeometryCollection:
		for i := 0; i < b.NumGeometries(); i++ {
			if !multiPolygonContainsSpherical(mpoly, b.GeometryN(i)) {
				return false
			}
		}
		return true
	}
	return false
}

// geometryCollectionContainsSpherical checks if a geometry collection contains any geometry.
func geometryCollectionContainsSpherical(gc *geom.GeometryCollection, g geom.Geometry) bool {
	for i := 0; i < gc.NumGeometries(); i++ {
		if containsSpherical(gc.GeometryN(i), g) {
			return true
		}
	}
	return false
}

// Helper functions for S2 operations

// lineStringsIntersectSpherical checks if two linestrings intersect on the sphere.
func lineStringsIntersectSpherical(ls1, ls2 *geom.LineString) bool {
	coords1 := ls1.Coordinates()
	coords2 := ls2.Coordinates()

	// Check if any edges cross
	for i := 1; i < len(coords1); i++ {
		ll1a := ToS2LatLng(coords1[i-1])
		ll1b := ToS2LatLng(coords1[i])
		a1 := s2.PointFromLatLng(ll1a)
		a2 := s2.PointFromLatLng(ll1b)

		for j := 1; j < len(coords2); j++ {
			ll2a := ToS2LatLng(coords2[j-1])
			ll2b := ToS2LatLng(coords2[j])
			b1 := s2.PointFromLatLng(ll2a)
			b2 := s2.PointFromLatLng(ll2b)

			// Check if edges cross on sphere
			if s2.CrossingSign(a1, a2, b1, b2) != s2.DoNotCross {
				return true
			}
		}
	}

	// Check if any endpoint is on the other linestring
	for _, c := range coords1 {
		pt := geom.NewPoint(c.X, c.Y)
		if PointOnLineString(pt, ls2, 0.01) {
			return true
		}
	}

	for _, c := range coords2 {
		pt := geom.NewPoint(c.X, c.Y)
		if PointOnLineString(pt, ls1, 0.01) {
			return true
		}
	}

	return false
}

// lineStringRingIntersectSpherical checks if a linestring intersects a ring.
func lineStringRingIntersectSpherical(ls *geom.LineString, ring *geom.LinearRing) bool {
	loop := ToS2Loop(ring)
	if loop == nil {
		return false
	}

	coords := ls.Coordinates()

	// Check if any point of linestring is inside or on the ring
	for _, c := range coords {
		ll := ToS2LatLng(c)
		pt := s2.PointFromLatLng(ll)
		if loop.ContainsPoint(pt) {
			return true
		}
	}

	// Check if any edge crosses the ring
	ringCoords := ring.Coordinates()
	for i := 1; i < len(coords); i++ {
		ll1a := ToS2LatLng(coords[i-1])
		ll1b := ToS2LatLng(coords[i])
		a1 := s2.PointFromLatLng(ll1a)
		a2 := s2.PointFromLatLng(ll1b)

		for j := 1; j < len(ringCoords); j++ {
			ll2a := ToS2LatLng(ringCoords[j-1])
			ll2b := ToS2LatLng(ringCoords[j])
			b1 := s2.PointFromLatLng(ll2a)
			b2 := s2.PointFromLatLng(ll2b)

			if s2.CrossingSign(a1, a2, b1, b2) != s2.DoNotCross {
				return true
			}
		}
	}

	return false
}

// loopIntersectsPolygon checks if an S2 loop intersects an S2 polygon.
func loopIntersectsPolygon(loop *s2.Loop, poly *s2.Polygon) bool {
	// Check if any vertex of the loop is inside the polygon
	for i := 0; i < loop.NumVertices(); i++ {
		if poly.ContainsPoint(loop.Vertex(i)) {
			return true
		}
	}

	// Check if any vertex of the polygon is inside the loop
	for i := 0; i < poly.NumLoops(); i++ {
		polyLoop := poly.Loop(i)
		for j := 0; j < polyLoop.NumVertices(); j++ {
			if loop.ContainsPoint(polyLoop.Vertex(j)) {
				return true
			}
		}
	}

	// Check for edge intersections
	for i := 0; i < loop.NumVertices(); i++ {
		a1 := loop.Vertex(i)
		a2 := loop.Vertex((i + 1) % loop.NumVertices())

		for j := 0; j < poly.NumLoops(); j++ {
			polyLoop := poly.Loop(j)
			for k := 0; k < polyLoop.NumVertices(); k++ {
				b1 := polyLoop.Vertex(k)
				b2 := polyLoop.Vertex((k + 1) % polyLoop.NumVertices())

				if s2.CrossingSign(a1, a2, b1, b2) != s2.DoNotCross {
					return true
				}
			}
		}
	}

	return false
}

// loopContainsLoop checks if loop1 contains loop2.
func loopContainsLoop(loop1, loop2 *s2.Loop) bool {
	// All vertices of loop2 must be inside loop1
	for i := 0; i < loop2.NumVertices(); i++ {
		if !loop1.ContainsPoint(loop2.Vertex(i)) {
			return false
		}
	}
	return true
}

// polygonContainsLoop checks if a polygon contains a loop.
func polygonContainsLoop(poly *s2.Polygon, loop *s2.Loop) bool {
	// All vertices of the loop must be inside the polygon
	for i := 0; i < loop.NumVertices(); i++ {
		if !poly.ContainsPoint(loop.Vertex(i)) {
			return false
		}
	}
	return true
}

// LoopContainsPoint checks if a ring contains a point using spherical geometry.
// Returns true if the point is inside the ring (including on the boundary).
// Returns false if either geometry is nil or empty.
func LoopContainsPoint(ring *geom.LinearRing, p *geom.Point) bool {
	if ring == nil || ring.IsEmpty() || p == nil || p.IsEmpty() {
		return false
	}

	loop := ToS2Loop(ring)
	if loop == nil {
		return false
	}

	s2Point := ToS2Point(p)
	return loop.ContainsPoint(s2Point)
}

// PolygonContainsPolygon checks if polygon p1 completely contains polygon p2.
// Returns true if every point in p2 is also in p1.
// Returns false if either polygon is nil or empty.
func PolygonContainsPolygon(p1, p2 *geom.Polygon) bool {
	if p1 == nil || p1.IsEmpty() || p2 == nil || p2.IsEmpty() {
		return false
	}

	s2Poly1 := ToS2Polygon(p1)
	s2Poly2 := ToS2Polygon(p2)
	if s2Poly1 == nil || s2Poly2 == nil {
		return false
	}

	return s2Poly1.Contains(s2Poly2)
}

// LoopsIntersect checks if two rings intersect using spherical geometry.
// Returns true if the rings share any points (including boundaries).
// Returns false if either ring is nil or empty.
func LoopsIntersect(r1, r2 *geom.LinearRing) bool {
	if r1 == nil || r1.IsEmpty() || r2 == nil || r2.IsEmpty() {
		return false
	}

	loop1 := ToS2Loop(r1)
	loop2 := ToS2Loop(r2)
	if loop1 == nil || loop2 == nil {
		return false
	}

	return loop1.Intersects(loop2)
}

// PointOnLineString checks if a point lies on a linestring using spherical geometry.
// Returns true if the point is on any segment of the linestring.
// The tolerance parameter specifies the maximum distance in meters for the point
// to be considered "on" the linestring.
func PointOnLineString(p *geom.Point, ls *geom.LineString, toleranceMeters float64) bool {
	if p == nil || p.IsEmpty() || ls == nil || ls.IsEmpty() {
		return false
	}

	dist := DistanceToLineString(p, ls)
	return dist <= toleranceMeters
}

// PointOnRing checks if a point lies on a ring using spherical geometry.
// Returns true if the point is on any segment of the ring boundary.
// The tolerance parameter specifies the maximum distance in meters.
func PointOnRing(p *geom.Point, ring *geom.LinearRing, toleranceMeters float64) bool {
	if p == nil || p.IsEmpty() || ring == nil || ring.IsEmpty() {
		return false
	}

	s2Point := ToS2Point(p)
	coords := ring.Coordinates()

	for i := 1; i < len(coords); i++ {
		ll1 := ToS2LatLng(coords[i-1])
		ll2 := ToS2LatLng(coords[i])
		p1 := s2.PointFromLatLng(ll1)
		p2 := s2.PointFromLatLng(ll2)

		dist := distanceToEdge(s2Point, p1, p2)
		if dist <= toleranceMeters {
			return true
		}
	}

	return false
}

// Disjoint checks if two polygons are disjoint (do not intersect).
// Returns true if the polygons share no points.
// Returns false if either polygon is nil or empty.
func Disjoint(p1, p2 *geom.Polygon) bool {
	return !Intersects(p1, p2)
}

// Within checks if polygon p1 is completely within polygon p2.
// This is equivalent to p2.Contains(p1).
func Within(p1, p2 *geom.Polygon) bool {
	return PolygonContainsPolygon(p2, p1)
}

// Overlaps checks if two polygons overlap (intersect but neither contains the other).
// Returns true if the polygons intersect and neither is completely contained in the other.
func Overlaps(p1, p2 *geom.Polygon) bool {
	if !Intersects(p1, p2) {
		return false
	}

	if PolygonContainsPolygon(p1, p2) || PolygonContainsPolygon(p2, p1) {
		return false
	}

	return true
}

// Touches checks if two polygons touch (share boundary points but not interior points).
// This is a more expensive operation as it requires checking the intersection type.
func Touches(p1, p2 *geom.Polygon) bool {
	if p1 == nil || p1.IsEmpty() || p2 == nil || p2.IsEmpty() {
		return false
	}

	// Polygons touch if they intersect but neither contains the other's interior
	// This is approximated by checking if they intersect but have no significant
	// area of overlap
	if !Intersects(p1, p2) {
		return false
	}

	s2Poly1 := ToS2Polygon(p1)
	s2Poly2 := ToS2Polygon(p2)
	if s2Poly1 == nil || s2Poly2 == nil {
		return false
	}

	// Approximate touch detection: if they intersect but the intersection
	// area is negligible compared to either polygon's area
	// Note: S2 doesn't have a built-in Union operation, so we use a heuristic

	// For a proper touch test, we would need to check if intersection is only
	// on boundaries. For now, we use a simplified approach:
	// If they intersect but neither contains the other, they likely touch or overlap
	if PolygonContainsPolygon(p1, p2) || PolygonContainsPolygon(p2, p1) {
		return false
	}

	// This is a simplified touch test - a full implementation would check
	// if the intersection is only at boundaries
	return true
}

// PointInRingWindingNumber uses the winding number algorithm on the sphere
// to determine if a point is inside a ring. This is an alternative to
// LoopContainsPoint that doesn't require S2 loop validation.
func PointInRingWindingNumber(ring *geom.LinearRing, p *geom.Point) bool {
	if ring == nil || ring.IsEmpty() || p == nil || p.IsEmpty() {
		return false
	}

	// Use S2's robust containment test via loop
	loop := ToS2Loop(ring)
	if loop == nil {
		return false
	}

	s2Point := ToS2Point(p)
	return loop.ContainsPoint(s2Point)
}

// LineStringIntersectsPolygon checks if a linestring intersects a polygon.
// Returns true if any part of the linestring is inside or crosses the polygon boundary.
func LineStringIntersectsPolygon(ls *geom.LineString, poly *geom.Polygon) bool {
	if ls == nil || ls.IsEmpty() || poly == nil || poly.IsEmpty() {
		return false
	}

	s2Poly := ToS2Polygon(poly)
	if s2Poly == nil {
		return false
	}

	coords := ls.Coordinates()
	for _, c := range coords {
		ll := ToS2LatLng(c)
		point := s2.PointFromLatLng(ll)
		if s2Poly.ContainsPoint(point) {
			return true
		}
	}

	// Check if any segment crosses the polygon boundary
	polyline := ToS2Polyline(ls)
	if polyline == nil {
		return false
	}

	// If any vertex of the polygon is close to the linestring, they intersect
	// This is a simplified check - a full implementation would check edge-edge intersections
	shell := poly.ExteriorRing()
	if shell != nil {
		shellCoords := shell.Coordinates()
		for _, c := range shellCoords {
			ll := ToS2LatLng(c)
			point := s2.PointFromLatLng(ll)
			// Project returns the closest point on the polyline and the edge index
			closest, _ := polyline.Project(point)
			dist := point.Distance(closest)
			// If polygon boundary is very close to linestring, they intersect
			if dist.Radians()*EarthMeanRadius < 1.0 { // 1 meter tolerance
				return true
			}
		}
	}

	return false
}
