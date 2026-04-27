// Package relate provides DE-9IM spatial relationship computation.
package relate

import (
	"github.com/robert-malhotra/go-topology-suite/algorithm"
	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/internal/topology"
)

// Relate computes the DE-9IM intersection matrix for two geometries.
func Relate(g1, g2 geom.Geometry) *IntersectionMatrix {
	if g1 == nil || g2 == nil {
		return NewIntersectionMatrix()
	}
	if g1.IsEmpty() || g2.IsEmpty() {
		return computeEmptyMatrix(g1, g2)
	}

	// Quick envelope check
	if !g1.Envelope().Intersects(g2.Envelope()) {
		return computeDisjointMatrix(g1, g2)
	}

	return computeRelate(g1, g2)
}

// RelatePattern tests if two geometries match a DE-9IM pattern.
func RelatePattern(g1, g2 geom.Geometry, pattern string) bool {
	matrix := Relate(g1, g2)
	return matrix.Matches(pattern)
}

// computeEmptyMatrix computes the matrix when one or both geometries are empty.
func computeEmptyMatrix(g1, g2 geom.Geometry) *IntersectionMatrix {
	m := NewIntersectionMatrix()

	// When a geometry is empty, all intersections with it are empty (F)
	// But exterior of empty intersects with everything
	if g1.IsEmpty() && g2.IsEmpty() {
		// Both empty: only E-E is non-empty (2)
		m[Exterior][Exterior] = DimArea
	} else if g1.IsEmpty() {
		// g1 empty: g1's exterior intersects all of g2
		dim2 := geomDimension(g2)
		m[Exterior][Interior] = dim2
		m[Exterior][Boundary] = boundaryDimension(g2)
		m[Exterior][Exterior] = DimArea
	} else {
		// g2 empty: g2's exterior intersects all of g1
		dim1 := geomDimension(g1)
		m[Interior][Exterior] = dim1
		m[Boundary][Exterior] = boundaryDimension(g1)
		m[Exterior][Exterior] = DimArea
	}

	return m
}

// computeDisjointMatrix computes the matrix when envelopes don't intersect.
func computeDisjointMatrix(g1, g2 geom.Geometry) *IntersectionMatrix {
	m := NewIntersectionMatrix()

	dim1 := geomDimension(g1)
	dim2 := geomDimension(g2)
	bnd1 := boundaryDimension(g1)
	bnd2 := boundaryDimension(g2)

	// Disjoint: I-E, B-E, E-I, E-B, E-E
	m[Interior][Exterior] = dim1
	m[Boundary][Exterior] = bnd1
	m[Exterior][Interior] = dim2
	m[Exterior][Boundary] = bnd2
	m[Exterior][Exterior] = DimArea

	return m
}

// computeRelate computes the full DE-9IM matrix.
func computeRelate(g1, g2 geom.Geometry) *IntersectionMatrix {
	if _, ok := g1.(*geom.GeometryCollection); !ok {
		if _, ok := g2.(*geom.GeometryCollection); ok {
			m := computeRelate(g2, g1).Transpose()
			m[Exterior][Exterior] = DimArea
			return m
		}
	}

	m := NewIntersectionMatrix()

	// Compute based on geometry type combinations
	switch a := g1.(type) {
	case *geom.Point:
		computePointRelate(a, g2, m)
	case *geom.LineString:
		computeLineStringRelate(a, g2, m)
	case *geom.LinearRing:
		computeLinearRingRelate(a, g2, m)
	case *geom.Polygon:
		computePolygonRelate(a, g2, m)
	case *geom.MultiPoint:
		computeMultiPointRelate(a, g2, m)
	case *geom.MultiLineString:
		computeMultiLineStringRelate(a, g2, m)
	case *geom.MultiPolygon:
		computeMultiPolygonRelate(a, g2, m)
	case *geom.GeometryCollection:
		computeCollectionRelate(a, g2, m)
	}

	if _, ok := g1.(*geom.GeometryCollection); ok {
		applyCollectionContainmentCorrections(g1, g2, m)
	} else if _, ok := g2.(*geom.GeometryCollection); ok {
		applyCollectionContainmentCorrections(g1, g2, m)
	}

	// Set exterior-exterior to 2D (always true for non-empty geometries)
	m[Exterior][Exterior] = DimArea

	return m
}

func applyCollectionContainmentCorrections(g1, g2 geom.Geometry, m *IntersectionMatrix) {
	g1ContainsG2 := geometryContainsGeometry(g1, g2)
	g2ContainsG1 := geometryContainsGeometry(g2, g1)
	if g1ContainsG2 {
		applyContainsCorrection(g1, g2, m)
	}
	if g2ContainsG1 {
		applyWithinCorrection(g1, g2, m)
	}
	if g1ContainsG2 && g2ContainsG1 {
		m[Interior][Boundary] = DimFalse
		m[Interior][Exterior] = DimFalse
		m[Boundary][Exterior] = DimFalse
		m[Exterior][Interior] = DimFalse
		m[Exterior][Boundary] = DimFalse
		restoreInteriorBoundaryForContainedLinework(g1, g2, m)
	}
	if gc, ok := g1.(*geom.GeometryCollection); ok && g1ContainsG2 {
		if ls, ok := lineStringGeometry(g2); ok && collectionHasLineMember(gc, ls) {
			m[Interior][Boundary] = DimFalse
		}
	}
}

func restoreInteriorBoundaryForContainedLinework(g1, g2 geom.Geometry, m *IntersectionMatrix) {
	for _, point := range geometryLineBoundaryPoints(g1) {
		if topology.PointLocation(point, g2) == geom.LocationInterior {
			m.SetAtLeast(Boundary, Interior, DimPoint)
			break
		}
	}
	for _, point := range geometryLineBoundaryPoints(g2) {
		if topology.PointLocation(point, g1) == geom.LocationInterior {
			m.SetAtLeast(Interior, Boundary, DimPoint)
			break
		}
	}
}

func geometryLineBoundaryPoints(g geom.Geometry) []geom.Coordinate {
	return lineSetBoundaryPoints(geom.ExtractLineStrings(g))
}

// computePointRelate computes the matrix for Point vs any geometry.
func computePointRelate(p *geom.Point, g geom.Geometry, m *IntersectionMatrix) {
	coord := p.Coordinate()
	loc := topology.PointLocation(coord, g)

	// Point has no boundary, so:
	// I-I, I-B, I-E based on location
	// B-* all False
	// E-* filled based on geometry B's components

	switch loc {
	case geom.LocationInterior:
		m[Interior][Interior] = DimPoint
	case geom.LocationBoundary:
		m[Interior][Boundary] = DimPoint
	case geom.LocationExterior:
		m[Interior][Exterior] = DimPoint
	}

	// Exterior of point vs interior/boundary/exterior of g
	dim2 := geomDimension(g)
	bnd2 := boundaryDimension(g)

	// The exterior of a point is everything except the point
	// So E-I is the dimension of g's interior (if g has any interior not at p)
	if dim2 >= DimPoint {
		m[Exterior][Interior] = dim2
	}
	if bnd2 >= DimPoint {
		m[Exterior][Boundary] = bnd2
	}
}

// computeLineStringRelate computes the matrix for LineString vs any geometry.
func computeLineStringRelate(ls *geom.LineString, g geom.Geometry, m *IntersectionMatrix) {
	switch b := g.(type) {
	case *geom.Point:
		computeLinePointRelate(ls, b, m)
	case *geom.LineString:
		computeLineLineRelate(ls, b, m)
	case *geom.LinearRing:
		computeLineLineRelate(ls, b.LineString, m)
	case *geom.Polygon:
		computeLinePolygonRelate(ls, b, m)
	case *geom.MultiPoint:
		computeLineMultiPointRelate(ls, b, m)
	case *geom.MultiLineString:
		computeLineMultiLineRelate(ls, b, m)
	case *geom.MultiPolygon:
		computeLineMultiPolygonRelate(ls, b, m)
	case *geom.GeometryCollection:
		if polys, ok := polygonSetGeometry(b); ok {
			computeLinePolygonSetRelate(ls, polys, m)
			return
		}
		for i := 0; i < b.NumGeometries(); i++ {
			computeLineStringRelate(ls, b.GeometryN(i), m)
		}
	}
}

// computeLinearRingRelate computes the matrix for LinearRing vs any geometry.
func computeLinearRingRelate(lr *geom.LinearRing, g geom.Geometry, m *IntersectionMatrix) {
	// LinearRing is treated as a closed LineString for relate purposes
	computeLineStringRelate(lr.LineString, g, m)
}

// computePolygonRelate computes the matrix for Polygon vs any geometry.
func computePolygonRelate(poly *geom.Polygon, g geom.Geometry, m *IntersectionMatrix) {
	switch b := g.(type) {
	case *geom.Point:
		computePolygonPointRelate(poly, b, m)
	case *geom.LineString:
		computePolygonLineRelate(poly, b, m)
	case *geom.LinearRing:
		computePolygonLineRelate(poly, b.LineString, m)
	case *geom.Polygon:
		computePolygonPolygonRelate(poly, b, m)
	case *geom.MultiPoint:
		computePolygonMultiPointRelate(poly, b, m)
	case *geom.MultiLineString:
		computePolygonMultiLineRelate(poly, b, m)
	case *geom.MultiPolygon:
		computePolygonMultiPolygonRelate(poly, b, m)
	case *geom.GeometryCollection:
		if polys, ok := polygonSetGeometry(b); ok {
			computePolygonSetPolygonSetRelate([]*geom.Polygon{poly}, polys, m)
			return
		}
		for i := 0; i < b.NumGeometries(); i++ {
			computePolygonRelate(poly, b.GeometryN(i), m)
		}
	}
}

// computeMultiPointRelate computes the matrix for MultiPoint vs any geometry.
func computeMultiPointRelate(mp *geom.MultiPoint, g geom.Geometry, m *IntersectionMatrix) {
	for i := 0; i < mp.NumGeometries(); i++ {
		p := mp.GeometryN(i).(*geom.Point)
		computePointRelate(p, g, m)
	}
}

// computeMultiLineStringRelate computes the matrix for MultiLineString vs any geometry.
func computeMultiLineStringRelate(mls *geom.MultiLineString, g geom.Geometry, m *IntersectionMatrix) {
	switch b := g.(type) {
	case *geom.LineString:
		computeLineSetLineSetRelate(multiLineStringLines(mls), []*geom.LineString{b}, m)
		return
	case *geom.LinearRing:
		computeLineSetLineSetRelate(multiLineStringLines(mls), []*geom.LineString{b.LineString}, m)
		return
	case *geom.MultiLineString:
		computeLineSetLineSetRelate(multiLineStringLines(mls), multiLineStringLines(b), m)
		return
	}

	for i := 0; i < mls.NumGeometries(); i++ {
		ls := mls.GeometryN(i).(*geom.LineString)
		computeLineStringRelate(ls, g, m)
	}
}

// computeMultiPolygonRelate computes the matrix for MultiPolygon vs any geometry.
func computeMultiPolygonRelate(mp *geom.MultiPolygon, g geom.Geometry, m *IntersectionMatrix) {
	switch b := g.(type) {
	case *geom.Polygon:
		computePolygonSetPolygonSetRelate(geom.ExtractPolygons(mp), []*geom.Polygon{b}, m)
		if multiPolygonContainsGeometry(mp, g) {
			applyContainsCorrection(mp, g, m)
		}
		return
	case *geom.MultiPolygon:
		computePolygonSetPolygonSetRelate(geom.ExtractPolygons(mp), geom.ExtractPolygons(b), m)
		if multiPolygonContainsGeometry(mp, g) {
			applyContainsCorrection(mp, g, m)
		}
		return
	}

	for i := 0; i < mp.NumGeometries(); i++ {
		poly := mp.GeometryN(i).(*geom.Polygon)
		computePolygonRelate(poly, g, m)
	}
	if multiPolygonContainsGeometry(mp, g) {
		applyContainsCorrection(mp, g, m)
	}
}

func multiPolygonContainsGeometry(mp *geom.MultiPolygon, g geom.Geometry) bool {
	if mp == nil || g == nil || g.IsEmpty() {
		return false
	}
	switch v := g.(type) {
	case *geom.Point:
		return !v.IsEmpty() && topology.PointLocation(v.Coordinate(), mp) == geom.LocationInterior
	case *geom.LineString:
		return lineContainedInMultiPolygon(v, mp)
	case *geom.LinearRing:
		return lineContainedInMultiPolygon(v.LineString, mp)
	case *geom.Polygon:
		return polygonContainedInMultiPolygon(v, mp)
	case *geom.MultiPoint:
		for i := 0; i < v.NumGeometries(); i++ {
			if !multiPolygonContainsGeometry(mp, v.GeometryN(i)) {
				return false
			}
		}
		return true
	case *geom.MultiLineString:
		for i := 0; i < v.NumGeometries(); i++ {
			if !multiPolygonContainsGeometry(mp, v.GeometryN(i)) {
				return false
			}
		}
		return true
	case *geom.MultiPolygon:
		for i := 0; i < v.NumGeometries(); i++ {
			if !multiPolygonContainsGeometry(mp, v.GeometryN(i)) {
				return false
			}
		}
		return true
	case *geom.GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			if !multiPolygonContainsGeometry(mp, v.GeometryN(i)) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func polygonContainedInMultiPolygon(poly *geom.Polygon, mp *geom.MultiPolygon) bool {
	interiorPoint, ok := topology.PolygonInteriorPoint(poly)
	if !ok || topology.PointLocation(interiorPoint, mp) != geom.LocationInterior {
		return false
	}
	for _, boundary := range topology.PolygonBoundaryLines([]*geom.Polygon{poly}) {
		if !lineCoveredByMultiPolygon(boundary, mp) {
			return false
		}
	}
	return true
}

func lineContainedInMultiPolygon(line *geom.LineString, mp *geom.MultiPolygon) bool {
	coords := line.Coordinates()
	for _, coord := range coords {
		if topology.PointLocation(coord, mp) != geom.LocationInterior {
			return false
		}
	}
	return lineCoveredByMultiPolygon(line, mp)
}

func lineCoveredByMultiPolygon(line *geom.LineString, mp *geom.MultiPolygon) bool {
	coords := line.Coordinates()
	for i := 1; i < len(coords); i++ {
		midpoint := geom.Coordinate{
			X: (coords[i-1].X + coords[i].X) / 2,
			Y: (coords[i-1].Y + coords[i].Y) / 2,
		}
		if topology.PointLocation(midpoint, mp) == geom.LocationExterior {
			return false
		}
	}
	return true
}

// computeCollectionRelate computes the matrix for GeometryCollection vs any geometry.
func computeCollectionRelate(gc *geom.GeometryCollection, g geom.Geometry, m *IntersectionMatrix) {
	if polysA, ok := polygonSetGeometry(gc); ok {
		if polysB, ok := polygonSetGeometry(g); ok {
			computePolygonSetPolygonSetRelate(polysA, polysB, m)
			if geometryContainsGeometry(gc, g) && !geometryContainsGeometry(g, gc) {
				applyContainsCorrection(gc, g, m)
			}
			return
		}
		switch b := g.(type) {
		case *geom.LineString:
			computePolygonSetLineRelate(polysA, b, m)
			if geometryContainsGeometry(gc, g) {
				applyContainsCorrection(gc, g, m)
			}
			return
		case *geom.LinearRing:
			computePolygonSetLineRelate(polysA, b.LineString, m)
			return
		case *geom.MultiLineString:
			for i := 0; i < b.NumGeometries(); i++ {
				computePolygonSetLineRelate(polysA, b.GeometryN(i).(*geom.LineString), m)
			}
			return
		}
	}

	for i := 0; i < gc.NumGeometries(); i++ {
		subM := computeRelate(gc.GeometryN(i), g)
		// Merge matrices
		for r := 0; r < 3; r++ {
			for c := 0; c < 3; c++ {
				if subM[r][c] > m[r][c] {
					m[r][c] = subM[r][c]
				}
			}
		}
	}

	if geometryContainsGeometry(gc, g) {
		applyContainsCorrection(gc, g, m)
		if ls, ok := lineStringGeometry(g); ok && collectionInteriorContainsLine(gc, ls) {
			m[Boundary][Interior] = DimFalse
			if collectionHasLineMember(gc, ls) {
				m[Interior][Boundary] = DimFalse
			}
		}
	} else if ls, ok := g.(*geom.LineString); ok && collectionInteriorContainsLine(gc, ls) {
		applyContainsCorrection(gc, ls, m)
	}
	if geometryContainsGeometry(g, gc) {
		applyWithinCorrection(gc, g, m)
		if geometryContainsGeometry(gc, g) {
			m[Interior][Boundary] = DimFalse
		}
	}

	applyCollectionBoundaryPromotion(gc, g, m)
}

func applyCollectionBoundaryPromotion(gc *geom.GeometryCollection, g geom.Geometry, m *IntersectionMatrix) {
	if m[Boundary][Boundary] != DimPoint {
		return
	}

	for _, point := range collectionInteriorPointMembers(gc) {
		if topology.PointLocation(point, g) == geom.LocationBoundary {
			m[Boundary][Boundary] = DimFalse
			return
		}
	}
}

func collectionInteriorPointMembers(gc *geom.GeometryCollection) []geom.Coordinate {
	var points []geom.Coordinate
	geom.ForEachPoint(gc, func(point *geom.Point) bool {
		coord := point.Coordinate()
		if topology.PointLocation(coord, gc) == geom.LocationInterior {
			points = append(points, coord)
		}
		return false
	})
	return points
}

func polygonSetGeometry(g geom.Geometry) ([]*geom.Polygon, bool) {
	if g == nil || g.IsEmpty() {
		return nil, false
	}
	switch v := g.(type) {
	case *geom.Polygon:
		return []*geom.Polygon{v}, true
	case *geom.MultiPolygon:
		return geom.ExtractPolygons(v), true
	case *geom.GeometryCollection:
		var polygons []*geom.Polygon
		for i := 0; i < v.NumGeometries(); i++ {
			componentPolygons, ok := polygonSetGeometry(v.GeometryN(i))
			if !ok {
				return nil, false
			}
			polygons = append(polygons, componentPolygons...)
		}
		return polygons, len(polygons) > 0
	default:
		return nil, false
	}
}

func geometryContainsGeometry(container, contained geom.Geometry) bool {
	if topologyCoversWithInterior(container, contained) {
		return true
	}
	if geom.Contains(container, contained) {
		return true
	}
	if polys, ok := polygonSetGeometry(container); ok {
		mp := geom.NewMultiPolygon(polys)
		return multiPolygonContainsGeometry(mp, contained)
	}
	return false
}

func topologyCoversWithInterior(container, contained geom.Geometry) bool {
	hasInterior := false
	for _, coord := range contained.Coordinates() {
		loc := topology.PointLocation(coord, container)
		if loc == geom.LocationExterior {
			return false
		}
		if loc == geom.LocationInterior {
			hasInterior = true
		}
	}

	lines := geom.ExtractLineStringsWithRings(contained)
	for _, line := range lines {
		coords := line.Coordinates()
		for i := 1; i < len(coords); i++ {
			midpoint := geom.Coordinate{
				X: (coords[i-1].X + coords[i].X) / 2,
				Y: (coords[i-1].Y + coords[i].Y) / 2,
			}
			loc := topology.PointLocation(midpoint, container)
			if loc == geom.LocationExterior {
				return false
			}
			if loc == geom.LocationInterior {
				hasInterior = true
			}
		}
	}

	for _, poly := range geom.ExtractPolygons(contained) {
		interiorPoint, ok := topology.PolygonInteriorPoint(poly)
		if !ok {
			continue
		}
		loc := topology.PointLocation(interiorPoint, container)
		if loc == geom.LocationExterior {
			return false
		}
		if loc == geom.LocationInterior {
			hasInterior = true
		}
	}

	return hasInterior
}

func applyContainsCorrection(container, contained geom.Geometry, m *IntersectionMatrix) {
	if container == nil || contained == nil || contained.IsEmpty() {
		return
	}
	m.SetAtLeast(Interior, Interior, geomDimension(contained))
	m.SetAtLeast(Interior, Boundary, boundaryDimension(contained))
	m.SetAtLeast(Interior, Exterior, geomDimension(container))
	m.SetAtLeast(Boundary, Exterior, boundaryDimension(container))
	m[Exterior][Interior] = DimFalse
	m[Exterior][Boundary] = DimFalse
}

func applyWithinCorrection(contained, container geom.Geometry, m *IntersectionMatrix) {
	if container == nil || contained == nil || contained.IsEmpty() {
		return
	}
	m.SetAtLeast(Interior, Interior, geomDimension(contained))
	m.SetAtLeast(Exterior, Interior, geomDimension(container))
	m.SetAtLeast(Exterior, Boundary, boundaryDimension(container))
	m[Interior][Exterior] = DimFalse
	m[Boundary][Exterior] = DimFalse
}

func lineStringGeometry(g geom.Geometry) (*geom.LineString, bool) {
	switch v := g.(type) {
	case *geom.LineString:
		return v, true
	case *geom.LinearRing:
		return v.LineString, true
	default:
		return nil, false
	}
}

func collectionHasLineMember(gc *geom.GeometryCollection, line *geom.LineString) bool {
	for i := 0; i < gc.NumGeometries(); i++ {
		switch v := gc.GeometryN(i).(type) {
		case *geom.LineString:
			if v.EqualsExact(line, geom.DefaultEpsilon) {
				return true
			}
		case *geom.LinearRing:
			if v.LineString.EqualsExact(line, geom.DefaultEpsilon) {
				return true
			}
		case *geom.GeometryCollection:
			if collectionHasLineMember(v, line) {
				return true
			}
		}
	}
	return false
}

func collectionInteriorContainsLine(gc *geom.GeometryCollection, ls *geom.LineString) bool {
	if polys := geom.ExtractPolygons(gc); len(polys) > 0 {
		return polygonSetInteriorContainsLine(polys, ls)
	}
	for i := 0; i < gc.NumGeometries(); i++ {
		if geometryInteriorContainsLine(gc.GeometryN(i), ls) {
			return true
		}
	}
	return false
}

func polygonSetInteriorContainsLine(polys []*geom.Polygon, ls *geom.LineString) bool {
	coords := ls.Coordinates()
	if len(coords) == 0 {
		return false
	}
	for _, c := range coords {
		if topology.PointLocationInPolygonSet(c, polys) != geom.LocationInterior {
			return false
		}
	}
	for i := 1; i < len(coords); i++ {
		midpoint := geom.Coordinate{
			X: (coords[i-1].X + coords[i].X) / 2,
			Y: (coords[i-1].Y + coords[i].Y) / 2,
		}
		if topology.PointLocationInPolygonSet(midpoint, polys) != geom.LocationInterior {
			return false
		}
	}
	return true
}

func geometryInteriorContainsLine(g geom.Geometry, ls *geom.LineString) bool {
	switch g := g.(type) {
	case *geom.Polygon:
		return polygonInteriorContainsLine(g, ls)
	case *geom.MultiPolygon:
		for i := 0; i < g.NumGeometries(); i++ {
			if polygonInteriorContainsLine(g.GeometryN(i).(*geom.Polygon), ls) {
				return true
			}
		}
	case *geom.GeometryCollection:
		return collectionInteriorContainsLine(g, ls)
	}
	return false
}

func polygonInteriorContainsLine(poly *geom.Polygon, ls *geom.LineString) bool {
	coords := ls.Coordinates()
	if len(coords) == 0 {
		return false
	}

	for _, c := range coords {
		if topology.PointLocationInPolygon(c, poly) != geom.LocationInterior {
			return false
		}
	}

	boundaryLines := topology.PolygonBoundaryLines([]*geom.Polygon{poly})
	for i := 1; i < len(coords); i++ {
		midpoint := geom.Coordinate{
			X: (coords[i-1].X + coords[i].X) / 2,
			Y: (coords[i-1].Y + coords[i].Y) / 2,
		}
		if topology.PointLocationInPolygon(midpoint, poly) != geom.LocationInterior {
			return false
		}

		for _, boundary := range boundaryLines {
			boundaryCoords := boundary.Coordinates()
			for j := 1; j < len(boundaryCoords); j++ {
				if algorithm.LineIntersection(coords[i-1], coords[i], boundaryCoords[j-1], boundaryCoords[j]).HasIntersection {
					return false
				}
			}
		}
	}

	return true
}

// computeLinePointRelate computes LineString vs Point.
func computeLinePointRelate(ls *geom.LineString, p *geom.Point, m *IntersectionMatrix) {
	coord := p.Coordinate()
	loc := pointLocationOnLine(coord, ls)

	switch loc {
	case geom.LocationInterior:
		m[Interior][Interior] = DimPoint
	case geom.LocationBoundary:
		m[Boundary][Interior] = DimPoint
	case geom.LocationExterior:
		m[Exterior][Interior] = DimPoint
	}

	// Line's interior and boundary are not completely covered by point
	m[Interior][Exterior] = DimLine
	if !ls.IsClosed() {
		m[Boundary][Exterior] = DimPoint
	}
}

// computeLineLineRelate computes LineString vs LineString.
func computeLineLineRelate(ls1, ls2 *geom.LineString, m *IntersectionMatrix) {
	computeLineSetLineSetRelate([]*geom.LineString{ls1}, []*geom.LineString{ls2}, m)
}

func computeLineSetLineSetRelate(linesA, linesB []*geom.LineString, m *IntersectionMatrix) {
	boundaryA := lineSetBoundaryPoints(linesA)
	boundaryB := lineSetBoundaryPoints(linesB)

	nodedSegments := topology.NodeLineSets(linesA, linesB)
	hasSharedLine := false
	hasAOnlyLine := false
	hasBOnlyLine := false
	for _, segment := range nodedSegments {
		switch {
		case segment.InA() && segment.InB():
			hasSharedLine = true
		case segment.InA():
			hasAOnlyLine = true
		case segment.InB():
			hasBOnlyLine = true
		}
	}

	if hasSharedLine {
		m[Interior][Interior] = DimLine
	}

	checkLineSetPointIntersections(linesA, linesB, boundaryA, boundaryB, m)
	setLineSetBoundaryIntersections(boundaryA, boundaryB, linesA, linesB, m)

	if hasAOnlyLine {
		m[Interior][Exterior] = DimLine
	}
	if hasBOnlyLine {
		m[Exterior][Interior] = DimLine
	}

	setLineSetBoundaryExterior(boundaryA, boundaryB, linesA, linesB, m)
}

func multiLineStringLines(mls *geom.MultiLineString) []*geom.LineString {
	lines := make([]*geom.LineString, 0, mls.NumGeometries())
	for i := 0; i < mls.NumGeometries(); i++ {
		if ls, ok := mls.GeometryN(i).(*geom.LineString); ok && !ls.IsEmpty() {
			lines = append(lines, ls)
		}
	}
	return lines
}

func checkLineSetPointIntersections(linesA, linesB []*geom.LineString, boundaryA, boundaryB []geom.Coordinate, m *IntersectionMatrix) {
	for _, a := range linesA {
		coordsA := a.Coordinates()
		for _, b := range linesB {
			coordsB := b.Coordinates()
			for i := 1; i < len(coordsA); i++ {
				for j := 1; j < len(coordsB); j++ {
					result := algorithm.LineIntersection(coordsA[i-1], coordsA[i], coordsB[j-1], coordsB[j])
					if !result.HasIntersection {
						continue
					}
					setLineSetPointIntersection(result.Intersection, linesA, linesB, boundaryA, boundaryB, m)
					if result.IsCollinear && !result.Intersection2.IsNaN() {
						setLineSetPointIntersection(result.Intersection2, linesA, linesB, boundaryA, boundaryB, m)
					}
				}
			}
		}
	}
}

func setLineSetPointIntersection(p geom.Coordinate, linesA, linesB []*geom.LineString, boundaryA, boundaryB []geom.Coordinate, m *IntersectionMatrix) {
	locA := pointLocationOnLineSet(p, linesA, boundaryA)
	locB := pointLocationOnLineSet(p, linesB, boundaryB)

	switch locA {
	case geom.LocationInterior:
		switch locB {
		case geom.LocationInterior:
			m.SetAtLeast(Interior, Interior, DimPoint)
		case geom.LocationBoundary:
			m.SetAtLeast(Interior, Boundary, DimPoint)
		}
	case geom.LocationBoundary:
		switch locB {
		case geom.LocationInterior:
			m.SetAtLeast(Boundary, Interior, DimPoint)
		case geom.LocationBoundary:
			m.SetAtLeast(Boundary, Boundary, DimPoint)
		}
	}
}

func setLineSetBoundaryIntersections(boundaryA, boundaryB []geom.Coordinate, linesA, linesB []*geom.LineString, m *IntersectionMatrix) {
	for _, ep := range boundaryA {
		switch pointLocationOnLineSet(ep, linesB, boundaryB) {
		case geom.LocationInterior:
			m.SetAtLeast(Boundary, Interior, DimPoint)
		case geom.LocationBoundary:
			m.SetAtLeast(Boundary, Boundary, DimPoint)
		}
	}
	for _, ep := range boundaryB {
		switch pointLocationOnLineSet(ep, linesA, boundaryA) {
		case geom.LocationInterior:
			m.SetAtLeast(Interior, Boundary, DimPoint)
		case geom.LocationBoundary:
			m.SetAtLeast(Boundary, Boundary, DimPoint)
		}
	}
}

func setLineSetBoundaryExterior(boundaryA, boundaryB []geom.Coordinate, linesA, linesB []*geom.LineString, m *IntersectionMatrix) {
	for _, ep := range boundaryA {
		if pointLocationOnLineSet(ep, linesB, boundaryB) == geom.LocationExterior {
			m.SetAtLeast(Boundary, Exterior, DimPoint)
		}
	}
	for _, ep := range boundaryB {
		if pointLocationOnLineSet(ep, linesA, boundaryA) == geom.LocationExterior {
			m.SetAtLeast(Exterior, Boundary, DimPoint)
		}
	}
}

func pointLocationOnLineSet(p geom.Coordinate, lines []*geom.LineString, boundary []geom.Coordinate) geom.Location {
	onLine := false
	for _, line := range lines {
		if pointLocationOnLine(p, line) != geom.LocationExterior {
			onLine = true
			break
		}
	}
	if !onLine {
		return geom.LocationExterior
	}
	if containsCoordinate(boundary, p) {
		return geom.LocationBoundary
	}
	return geom.LocationInterior
}

func lineSetBoundaryPoints(lines []*geom.LineString) []geom.Coordinate {
	var endpoints []geom.Coordinate
	for _, line := range lines {
		endpoints = append(endpoints, lineBoundaryEndpoints(line)...)
	}

	boundary := endpoints[:0]
	for _, endpoint := range endpoints {
		count := 0
		for _, other := range endpoints {
			if endpoint.Equals2D(other, geom.DefaultEpsilon) {
				count++
			}
		}
		if count%2 == 1 && !containsCoordinate(boundary, endpoint) {
			boundary = append(boundary, endpoint)
		}
	}
	return boundary
}

func containsCoordinate(coords []geom.Coordinate, p geom.Coordinate) bool {
	for _, coord := range coords {
		if coord.Equals2D(p, geom.DefaultEpsilon) {
			return true
		}
	}
	return false
}

func setLineBoundaryExterior(ls1, ls2 *geom.LineString, m *IntersectionMatrix) {
	for _, ep := range lineBoundaryEndpoints(ls1) {
		if pointLocationOnLine(ep, ls2) == geom.LocationExterior {
			m.SetAtLeast(Boundary, Exterior, DimPoint)
		}
	}
	for _, ep := range lineBoundaryEndpoints(ls2) {
		if pointLocationOnLine(ep, ls1) == geom.LocationExterior {
			m.SetAtLeast(Exterior, Boundary, DimPoint)
		}
	}
}

func lineBoundaryEndpoints(ls *geom.LineString) []geom.Coordinate {
	if ls.IsClosed() || ls.IsEmpty() {
		return nil
	}
	coords := ls.Coordinates()
	if len(coords) < 2 {
		return nil
	}
	return []geom.Coordinate{coords[0], coords[len(coords)-1]}
}

// checkLineBoundaryIntersection checks if line boundaries intersect.
func checkLineBoundaryIntersection(ls1, ls2 *geom.LineString, m *IntersectionMatrix) {
	if ls1.IsClosed() && ls2.IsClosed() {
		return // Closed lines have no boundary
	}

	// Check if ls1 endpoints are on ls2
	for _, ep := range lineBoundaryEndpoints(ls1) {
		loc := pointLocationOnLine(ep, ls2)
		switch loc {
		case geom.LocationInterior:
			m.SetAtLeast(Boundary, Interior, DimPoint)
		case geom.LocationBoundary:
			m.SetAtLeast(Boundary, Boundary, DimPoint)
		}
	}

	// Check if ls2 endpoints are on ls1
	for _, ep := range lineBoundaryEndpoints(ls2) {
		loc := pointLocationOnLine(ep, ls1)
		switch loc {
		case geom.LocationInterior:
			m.SetAtLeast(Interior, Boundary, DimPoint)
		case geom.LocationBoundary:
			m.SetAtLeast(Boundary, Boundary, DimPoint)
		}
	}
}

// updateLineEndpointMatrix updates the matrix based on endpoint intersections.
func updateLineEndpointMatrix(ls1, ls2 *geom.LineString, m *IntersectionMatrix) {
	coords1 := ls1.Coordinates()
	coords2 := ls2.Coordinates()

	if len(coords1) < 2 || len(coords2) < 2 {
		return
	}

	// Check all combinations of internal points vs endpoints
	for i := 1; i < len(coords1)-1; i++ {
		c := coords1[i]
		if c.Equals2D(coords2[0], geom.DefaultEpsilon) ||
			c.Equals2D(coords2[len(coords2)-1], geom.DefaultEpsilon) {
			if !ls2.IsClosed() {
				m.SetAtLeast(Interior, Boundary, DimPoint)
			}
		}
	}

	for j := 1; j < len(coords2)-1; j++ {
		c := coords2[j]
		if c.Equals2D(coords1[0], geom.DefaultEpsilon) ||
			c.Equals2D(coords1[len(coords1)-1], geom.DefaultEpsilon) {
			if !ls1.IsClosed() {
				m.SetAtLeast(Boundary, Interior, DimPoint)
			}
		}
	}
}

// computeLinePolygonRelate computes LineString vs Polygon.
func computeLinePolygonRelate(ls *geom.LineString, poly *geom.Polygon, m *IntersectionMatrix) {
	computeLinePolygonSetRelate(ls, []*geom.Polygon{poly}, m)
}

func computeLinePolygonSetRelate(ls *geom.LineString, polys []*geom.Polygon, m *IntersectionMatrix) {
	coords := ls.Coordinates()
	if len(coords) < 2 {
		return
	}

	boundaryLines := topology.PolygonBoundaryLines(polys)
	nodedSegments := topology.NodeLineSets([]*geom.LineString{ls}, boundaryLines)
	for _, segment := range nodedSegments {
		if !segment.InA() {
			continue
		}
		midpoint := geom.NewCoordinate(
			(segment.Start.X+segment.End.X)/2,
			(segment.Start.Y+segment.End.Y)/2,
		)
		switch topology.PointLocationInPolygonSet(midpoint, polys) {
		case geom.LocationInterior:
			m.SetAtLeast(Interior, Interior, DimLine)
		case geom.LocationBoundary:
			m.SetAtLeast(Interior, Boundary, DimLine)
		case geom.LocationExterior:
			m.SetAtLeast(Interior, Exterior, DimLine)
		}

		setLineInteriorPolygonSetBoundaryPoint(segment.Start, ls, polys, m)
		setLineInteriorPolygonSetBoundaryPoint(segment.End, ls, polys, m)
	}

	if !ls.IsClosed() {
		for _, endpoint := range lineBoundaryEndpoints(ls) {
			switch topology.PointLocationInPolygonSet(endpoint, polys) {
			case geom.LocationInterior:
				m.SetAtLeast(Boundary, Interior, DimPoint)
			case geom.LocationBoundary:
				m.SetAtLeast(Boundary, Boundary, DimPoint)
			case geom.LocationExterior:
				m.SetAtLeast(Boundary, Exterior, DimPoint)
			}
		}
	}

	// Exterior-Interior: polygon's interior not covered by line
	m[Exterior][Interior] = DimArea
	// Exterior-Boundary: polygon's boundary not covered by line
	m[Exterior][Boundary] = DimLine
}

func setLineInteriorPolygonBoundaryPoint(point geom.Coordinate, ls *geom.LineString, poly *geom.Polygon, m *IntersectionMatrix) {
	setLineInteriorPolygonSetBoundaryPoint(point, ls, []*geom.Polygon{poly}, m)
}

func setLineInteriorPolygonSetBoundaryPoint(point geom.Coordinate, ls *geom.LineString, polys []*geom.Polygon, m *IntersectionMatrix) {
	if pointLocationOnLine(point, ls) != geom.LocationInterior {
		return
	}
	if topology.PointLocationInPolygonSet(point, polys) == geom.LocationBoundary {
		m.SetAtLeast(Interior, Boundary, DimPoint)
	}
}

// computePolygonPointRelate computes Polygon vs Point.
func computePolygonPointRelate(poly *geom.Polygon, p *geom.Point, m *IntersectionMatrix) {
	coord := p.Coordinate()
	loc := topology.PointLocationInPolygon(coord, poly)

	switch loc {
	case geom.LocationInterior:
		m[Interior][Interior] = DimPoint
	case geom.LocationBoundary:
		m[Boundary][Interior] = DimPoint
	case geom.LocationExterior:
		m[Exterior][Interior] = DimPoint
	}

	// Polygon always has parts not covered by point
	m[Interior][Exterior] = DimArea
	m[Boundary][Exterior] = DimLine
}

// computePolygonLineRelate computes Polygon vs LineString.
func computePolygonLineRelate(poly *geom.Polygon, ls *geom.LineString, m *IntersectionMatrix) {
	computePolygonSetLineRelate([]*geom.Polygon{poly}, ls, m)
}

func computePolygonSetLineRelate(polys []*geom.Polygon, ls *geom.LineString, m *IntersectionMatrix) {
	// This is the transpose of line vs polygon set.
	tempM := NewIntersectionMatrix()
	computeLinePolygonSetRelate(ls, polys, tempM)

	// Transpose results
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			m.SetAtLeast(j, i, tempM[i][j])
		}
	}
}

// computePolygonPolygonRelate computes Polygon vs Polygon.
func computePolygonPolygonRelate(poly1, poly2 *geom.Polygon, m *IntersectionMatrix) {
	computePolygonSetPolygonSetRelate([]*geom.Polygon{poly1}, []*geom.Polygon{poly2}, m)
}

func computePolygonSetPolygonSetRelate(polysA, polysB []*geom.Polygon, m *IntersectionMatrix) {
	boundaryLines1 := topology.PolygonBoundaryLines(polysA)
	boundaryLines2 := topology.PolygonBoundaryLines(polysB)
	graphEdges := topology.BuildPolygonBoundaryGraph(polysA, polysB)

	// Sample points from both polygons
	hasInteriorInterior := false
	hasInteriorBoundary := false
	hasInteriorExterior := false
	hasBoundaryInterior := false
	hasBoundaryBoundary := false
	hasBoundaryExterior := false
	hasExteriorBoundary := false

	for _, line := range boundaryLines1 {
		for _, c := range line.Coordinates() {
			loc := topology.PointLocationInPolygonSet(c, polysB)

			switch loc {
			case geom.LocationInterior:
				hasBoundaryInterior = true
			case geom.LocationBoundary:
				hasBoundaryBoundary = true
			case geom.LocationExterior:
				hasBoundaryExterior = true
			}
		}
	}

	for _, line := range boundaryLines2 {
		for _, c := range line.Coordinates() {
			loc := topology.PointLocationInPolygonSet(c, polysA)

			switch loc {
			case geom.LocationInterior:
				hasInteriorBoundary = true
			case geom.LocationBoundary:
				hasBoundaryBoundary = true
			case geom.LocationExterior:
				hasInteriorExterior = false // Don't set from boundary points
				hasExteriorBoundary = true
			}
		}
	}

	// Check representative interior point of poly1 vs poly2. A centroid can
	// lie in a hole, so use a topology-selected interior point.
	for _, poly := range polysA {
		if interior1, ok := topology.PolygonInteriorPoint(poly); ok {
			loc := topology.PointLocationInPolygonSet(interior1, polysB)
			switch loc {
			case geom.LocationInterior:
				hasInteriorInterior = true
			case geom.LocationExterior:
				hasInteriorExterior = true
			}
		}
	}

	// Check representative interior point of poly2 vs poly1.
	hasExteriorInterior := false
	for _, poly := range polysB {
		if point, ok := topology.PolygonInteriorPoint(poly); ok {
			loc := topology.PointLocationInPolygonSet(point, polysA)
			switch loc {
			case geom.LocationInterior:
				hasInteriorInterior = true
			case geom.LocationExterior:
				hasExteriorInterior = true
			}
		}
	}

	applyPolygonBoundaryGraphRelate(graphEdges, m)

	if dim, ok := topology.PolygonBoundaryIntersectionDimension(polysA, polysB); ok {
		hasBoundaryBoundary = true
		if dim == geom.DimensionLine {
			m.SetAtLeast(Boundary, Boundary, DimLine)
		}
	}

	// Set matrix values
	if hasInteriorInterior {
		m[Interior][Interior] = DimArea
	}
	if hasInteriorBoundary {
		m.SetAtLeast(Interior, Boundary, DimLine)
	}
	if hasInteriorExterior {
		m[Interior][Exterior] = DimArea
	}
	if hasBoundaryInterior {
		m.SetAtLeast(Boundary, Interior, DimLine)
	}
	if hasBoundaryBoundary {
		m.SetAtLeast(Boundary, Boundary, DimPoint)
	}
	if hasBoundaryExterior {
		m.SetAtLeast(Boundary, Exterior, DimLine)
	}

	// Exterior-Interior: poly2's interior not in poly1
	if hasExteriorInterior {
		m[Exterior][Interior] = DimArea
	}

	if hasExteriorBoundary {
		m.SetAtLeast(Exterior, Boundary, DimLine)
	}
}

func applyPolygonBoundaryGraphRelate(edges []topology.PolygonBoundaryEdge, m *IntersectionMatrix) {
	for _, edge := range edges {
		setPolygonFaceSideRelate(edge.Left, m)
		setPolygonFaceSideRelate(edge.Right, m)

		if edge.Sources.InA() && polygonBoundaryEdgeIsSetBoundary(edge, true) {
			setPolygonBoundaryEdgeRelate(Boundary, edge, true, m)
		}
		if edge.Sources.InB() && polygonBoundaryEdgeIsSetBoundary(edge, false) {
			setPolygonBoundaryEdgeRelate(Boundary, edge, false, m)
		}
	}
}

func polygonBoundaryEdgeIsSetBoundary(edge topology.PolygonBoundaryEdge, sourceA bool) bool {
	left := edge.Left.LocA
	right := edge.Right.LocA
	if !sourceA {
		left = edge.Left.LocB
		right = edge.Right.LocB
	}
	return (left == geom.LocationInterior && right == geom.LocationExterior) ||
		(left == geom.LocationExterior && right == geom.LocationInterior)
}

func setPolygonFaceSideRelate(label topology.PolygonEdgeLabel, m *IntersectionMatrix) {
	switch {
	case label.LocA == geom.LocationInterior && label.LocB == geom.LocationInterior:
		m.SetAtLeast(Interior, Interior, DimArea)
	case label.LocA == geom.LocationInterior && label.LocB == geom.LocationExterior:
		m.SetAtLeast(Interior, Exterior, DimArea)
	case label.LocA == geom.LocationExterior && label.LocB == geom.LocationInterior:
		m.SetAtLeast(Exterior, Interior, DimArea)
	}
}

func setPolygonBoundaryEdgeRelate(boundaryLoc int, edge topology.PolygonBoundaryEdge, sourceA bool, m *IntersectionMatrix) {
	if edge.Sources.InA() && edge.Sources.InB() {
		m.SetAtLeast(Boundary, Boundary, DimLine)
		return
	}

	leftLoc := edge.Left.LocB
	rightLoc := edge.Right.LocB
	row := boundaryLoc
	if !sourceA {
		leftLoc = edge.Left.LocA
		rightLoc = edge.Right.LocA
		row = -1
	}

	loc := polygonBoundaryEdgeLocation(leftLoc, rightLoc)
	if loc < 0 {
		return
	}
	if sourceA {
		m.SetAtLeast(row, loc, DimLine)
		return
	}
	m.SetAtLeast(loc, Boundary, DimLine)
}

func polygonBoundaryEdgeLocation(left, right geom.Location) int {
	if left == geom.LocationInterior && right == geom.LocationInterior {
		return Interior
	}
	if left == geom.LocationExterior && right == geom.LocationExterior {
		return Exterior
	}
	return -1
}

// computeLineMultiPointRelate computes LineString vs MultiPoint.
func computeLineMultiPointRelate(ls *geom.LineString, mp *geom.MultiPoint, m *IntersectionMatrix) {
	for i := 0; i < mp.NumGeometries(); i++ {
		p := mp.GeometryN(i).(*geom.Point)
		computeLinePointRelate(ls, p, m)
	}
}

// computeLineMultiLineRelate computes LineString vs MultiLineString.
func computeLineMultiLineRelate(ls *geom.LineString, mls *geom.MultiLineString, m *IntersectionMatrix) {
	computeLineSetLineSetRelate([]*geom.LineString{ls}, multiLineStringLines(mls), m)
}

// computeLineMultiPolygonRelate computes LineString vs MultiPolygon.
func computeLineMultiPolygonRelate(ls *geom.LineString, mp *geom.MultiPolygon, m *IntersectionMatrix) {
	computeLinePolygonSetRelate(ls, geom.ExtractPolygons(mp), m)
}

// computePolygonMultiPointRelate computes Polygon vs MultiPoint.
func computePolygonMultiPointRelate(poly *geom.Polygon, mp *geom.MultiPoint, m *IntersectionMatrix) {
	for i := 0; i < mp.NumGeometries(); i++ {
		p := mp.GeometryN(i).(*geom.Point)
		computePolygonPointRelate(poly, p, m)
	}
}

// computePolygonMultiLineRelate computes Polygon vs MultiLineString.
func computePolygonMultiLineRelate(poly *geom.Polygon, mls *geom.MultiLineString, m *IntersectionMatrix) {
	for i := 0; i < mls.NumGeometries(); i++ {
		ls := mls.GeometryN(i).(*geom.LineString)
		computePolygonLineRelate(poly, ls, m)
	}
}

// computePolygonMultiPolygonRelate computes Polygon vs MultiPolygon.
func computePolygonMultiPolygonRelate(poly *geom.Polygon, mp *geom.MultiPolygon, m *IntersectionMatrix) {
	computePolygonSetPolygonSetRelate([]*geom.Polygon{poly}, geom.ExtractPolygons(mp), m)
}

// pointLocationOnLine determines where a point lies relative to a line string.
func pointLocationOnLine(p geom.Coordinate, ls *geom.LineString) geom.Location {
	if ls.IsEmpty() {
		return geom.LocationExterior
	}

	coords := ls.Coordinates()

	// Check if at endpoint (boundary for non-closed lines)
	if !ls.IsClosed() {
		if p.Equals2D(coords[0], geom.DefaultEpsilon) ||
			p.Equals2D(coords[len(coords)-1], geom.DefaultEpsilon) {
			return geom.LocationBoundary
		}
	}

	// Check if on any segment
	for i := 1; i < len(coords); i++ {
		if isPointOnSegment(p, coords[i-1], coords[i]) {
			return geom.LocationInterior
		}
	}

	return geom.LocationExterior
}

// isPointOnSegment checks if a point is on a line segment.
func isPointOnSegment(p, a, b geom.Coordinate) bool {
	if algorithm.OrientationIndex(a, b, p) != algorithm.Collinear {
		return false
	}

	// Check if within bounding box
	minX, maxX := a.X, b.X
	if minX > maxX {
		minX, maxX = maxX, minX
	}
	minY, maxY := a.Y, b.Y
	if minY > maxY {
		minY, maxY = maxY, minY
	}

	return p.X >= minX-geom.DefaultEpsilon && p.X <= maxX+geom.DefaultEpsilon &&
		p.Y >= minY-geom.DefaultEpsilon && p.Y <= maxY+geom.DefaultEpsilon
}

// geomDimension returns the dimension of a geometry.
func geomDimension(g geom.Geometry) Dimension {
	switch g.Dimension() {
	case geom.DimensionEmpty:
		return DimFalse
	case geom.DimensionPoint:
		return DimPoint
	case geom.DimensionLine:
		return DimLine
	case geom.DimensionArea:
		return DimArea
	default:
		return DimFalse
	}
}

// boundaryDimension returns the dimension of a geometry's boundary.
func boundaryDimension(g geom.Geometry) Dimension {
	switch g := g.(type) {
	case *geom.Point, *geom.MultiPoint:
		return DimFalse // Points have no boundary
	case *geom.LineString:
		if g.IsClosed() {
			return DimFalse // Closed lines have no boundary
		}
		return DimPoint // Boundary is endpoints
	case *geom.LinearRing:
		return DimFalse // Rings are closed, no boundary
	case *geom.Polygon, *geom.MultiPolygon:
		return DimLine // Boundary is rings
	case *geom.MultiLineString:
		for i := 0; i < g.NumGeometries(); i++ {
			ls := g.GeometryN(i).(*geom.LineString)
			if !ls.IsClosed() {
				return DimPoint
			}
		}
		return DimFalse
	case *geom.GeometryCollection:
		maxDim := DimFalse
		for i := 0; i < g.NumGeometries(); i++ {
			dim := boundaryDimension(g.GeometryN(i))
			if dim > maxDim {
				maxDim = dim
			}
		}
		return maxDim
	default:
		return DimFalse
	}
}
