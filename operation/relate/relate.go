// Package relate provides DE-9IM spatial relationship computation.
package relate

import (
	"github.com/robert-malhotra/go-topology-suite/algorithm"
	"github.com/robert-malhotra/go-topology-suite/geom"
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

	// Set exterior-exterior to 2D (always true for non-empty geometries)
	m[Exterior][Exterior] = DimArea

	return m
}

// computePointRelate computes the matrix for Point vs any geometry.
func computePointRelate(p *geom.Point, g geom.Geometry, m *IntersectionMatrix) {
	coord := p.Coordinate()
	loc := algorithm.PointLocation(coord, g)

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
	for i := 0; i < mls.NumGeometries(); i++ {
		ls := mls.GeometryN(i).(*geom.LineString)
		computeLineStringRelate(ls, g, m)
	}
}

// computeMultiPolygonRelate computes the matrix for MultiPolygon vs any geometry.
func computeMultiPolygonRelate(mp *geom.MultiPolygon, g geom.Geometry, m *IntersectionMatrix) {
	for i := 0; i < mp.NumGeometries(); i++ {
		poly := mp.GeometryN(i).(*geom.Polygon)
		computePolygonRelate(poly, g, m)
	}
}

// computeCollectionRelate computes the matrix for GeometryCollection vs any geometry.
func computeCollectionRelate(gc *geom.GeometryCollection, g geom.Geometry, m *IntersectionMatrix) {
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
	coords1 := ls1.Coordinates()
	coords2 := ls2.Coordinates()

	// Check for intersections between segments
	hasProperIntersection := false
	hasEndpointIntersection := false
	hasCollinearOverlap := false

	for i := 1; i < len(coords1); i++ {
		for j := 1; j < len(coords2); j++ {
			result := algorithm.LineIntersection(coords1[i-1], coords1[i], coords2[j-1], coords2[j])
			if result.HasIntersection {
				if result.IsCollinear {
					hasCollinearOverlap = true
				} else if result.IsProper {
					hasProperIntersection = true
				} else {
					hasEndpointIntersection = true
				}
			}
		}
	}

	// Determine dimensions
	if hasCollinearOverlap {
		m[Interior][Interior] = DimLine
	} else if hasProperIntersection {
		m[Interior][Interior] = DimPoint
	}

	// Check boundary intersections (endpoints)
	checkLineBoundaryIntersection(ls1, ls2, m)

	// Interior-Exterior: lines always have parts in each other's exterior
	m[Interior][Exterior] = DimLine
	m[Exterior][Interior] = DimLine

	// Boundary-Exterior
	if !ls1.IsClosed() {
		m[Boundary][Exterior] = DimPoint
	}
	if !ls2.IsClosed() {
		m[Exterior][Boundary] = DimPoint
	}

	// Update based on endpoint intersections
	if hasEndpointIntersection {
		updateLineEndpointMatrix(ls1, ls2, m)
	}
}

// checkLineBoundaryIntersection checks if line boundaries intersect.
func checkLineBoundaryIntersection(ls1, ls2 *geom.LineString, m *IntersectionMatrix) {
	if ls1.IsClosed() && ls2.IsClosed() {
		return // Closed lines have no boundary
	}

	coords1 := ls1.Coordinates()
	coords2 := ls2.Coordinates()

	// Get endpoints of ls1 (its boundary)
	var endpoints1 []geom.Coordinate
	if !ls1.IsClosed() && len(coords1) >= 2 {
		endpoints1 = []geom.Coordinate{coords1[0], coords1[len(coords1)-1]}
	}

	// Get endpoints of ls2
	var endpoints2 []geom.Coordinate
	if !ls2.IsClosed() && len(coords2) >= 2 {
		endpoints2 = []geom.Coordinate{coords2[0], coords2[len(coords2)-1]}
	}

	// Check if ls1 endpoints are on ls2
	for _, ep := range endpoints1 {
		loc := pointLocationOnLine(ep, ls2)
		switch loc {
		case geom.LocationInterior:
			m.SetAtLeast(Boundary, Interior, DimPoint)
		case geom.LocationBoundary:
			m.SetAtLeast(Boundary, Boundary, DimPoint)
		}
	}

	// Check if ls2 endpoints are on ls1
	for _, ep := range endpoints2 {
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
	coords := ls.Coordinates()

	hasInteriorInterior := false
	hasInteriorBoundary := false
	hasInteriorExterior := false
	hasBoundaryInterior := false
	hasBoundaryBoundary := false
	hasBoundaryExterior := false

	// Check each point on the line
	for i, c := range coords {
		loc := algorithm.PointLocationInPolygon(c, poly)
		isEndpoint := !ls.IsClosed() && (i == 0 || i == len(coords)-1)

		switch loc {
		case geom.LocationInterior:
			if isEndpoint {
				hasBoundaryInterior = true
			} else {
				hasInteriorInterior = true
			}
		case geom.LocationBoundary:
			if isEndpoint {
				hasBoundaryBoundary = true
			} else {
				hasInteriorBoundary = true
			}
		case geom.LocationExterior:
			if isEndpoint {
				hasBoundaryExterior = true
			} else {
				hasInteriorExterior = true
			}
		}
	}

	// Check for segment intersections with polygon boundary
	shellCoords := poly.ExteriorRing().Coordinates()
	for i := 1; i < len(coords); i++ {
		for j := 1; j < len(shellCoords); j++ {
			result := algorithm.LineIntersection(coords[i-1], coords[i], shellCoords[j-1], shellCoords[j])
			if result.HasIntersection {
				if result.IsProper {
					hasInteriorBoundary = true
				}
			}
		}
	}

	// Set matrix values
	if hasInteriorInterior {
		m[Interior][Interior] = DimLine
	}
	if hasInteriorBoundary {
		m.SetAtLeast(Interior, Boundary, DimPoint)
	}
	if hasInteriorExterior {
		m[Interior][Exterior] = DimLine
	}
	if hasBoundaryInterior {
		m.SetAtLeast(Boundary, Interior, DimPoint)
	}
	if hasBoundaryBoundary {
		m.SetAtLeast(Boundary, Boundary, DimPoint)
	}
	if hasBoundaryExterior {
		m.SetAtLeast(Boundary, Exterior, DimPoint)
	}

	// Exterior-Interior: polygon's interior not covered by line
	m[Exterior][Interior] = DimArea
	// Exterior-Boundary: polygon's boundary not covered by line
	m[Exterior][Boundary] = DimLine
}

// computePolygonPointRelate computes Polygon vs Point.
func computePolygonPointRelate(poly *geom.Polygon, p *geom.Point, m *IntersectionMatrix) {
	coord := p.Coordinate()
	loc := algorithm.PointLocationInPolygon(coord, poly)

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
	// This is the transpose of line vs polygon
	tempM := NewIntersectionMatrix()
	computeLinePolygonRelate(ls, poly, tempM)

	// Transpose results
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			m.SetAtLeast(j, i, tempM[i][j])
		}
	}
}

// computePolygonPolygonRelate computes Polygon vs Polygon.
func computePolygonPolygonRelate(poly1, poly2 *geom.Polygon, m *IntersectionMatrix) {
	// Check shell intersections
	shell1 := poly1.ExteriorRing().Coordinates()
	shell2 := poly2.ExteriorRing().Coordinates()

	// Sample points from both polygons
	hasInteriorInterior := false
	hasInteriorBoundary := false
	hasInteriorExterior := false
	hasBoundaryInterior := false
	hasBoundaryBoundary := false
	hasBoundaryExterior := false

	// Check shell1 points against poly2
	for _, c := range shell1 {
		loc := algorithm.PointLocationInPolygon(c, poly2)

		switch loc {
		case geom.LocationInterior:
			hasBoundaryInterior = true
		case geom.LocationBoundary:
			hasBoundaryBoundary = true
		case geom.LocationExterior:
			hasBoundaryExterior = true
		}
	}

	// Check shell2 points against poly1
	for _, c := range shell2 {
		loc := algorithm.PointLocationInPolygon(c, poly1)

		switch loc {
		case geom.LocationInterior:
			hasInteriorBoundary = true
		case geom.LocationBoundary:
			hasBoundaryBoundary = true
		case geom.LocationExterior:
			hasInteriorExterior = false // Don't set from boundary points
		}
	}

	// Check centroid of poly1 vs poly2
	centroid1 := poly1.Centroid()
	if !centroid1.IsEmpty() {
		loc := algorithm.PointLocationInPolygon(centroid1.Coordinate(), poly2)
		switch loc {
		case geom.LocationInterior:
			hasInteriorInterior = true
		case geom.LocationExterior:
			hasInteriorExterior = true
		}
	}

	// Check centroid of poly2 vs poly1
	centroid2 := poly2.Centroid()
	if !centroid2.IsEmpty() {
		loc := algorithm.PointLocationInPolygon(centroid2.Coordinate(), poly1)
		if loc == geom.LocationExterior {
			hasInteriorExterior = true
		}
	}

	// Check for boundary-boundary intersections
	for i := 1; i < len(shell1); i++ {
		for j := 1; j < len(shell2); j++ {
			result := algorithm.LineIntersection(shell1[i-1], shell1[i], shell2[j-1], shell2[j])
			if result.HasIntersection {
				if result.IsCollinear {
					hasBoundaryBoundary = true
					m.SetAtLeast(Boundary, Boundary, DimLine)
				} else {
					hasBoundaryBoundary = true
				}
			}
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
	loc := algorithm.PointLocationInPolygon(centroid2.Coordinate(), poly1)
	if loc == geom.LocationExterior {
		m[Exterior][Interior] = DimArea
	}

	// Exterior-Boundary: poly2's boundary not in poly1's interior
	m.SetAtLeast(Exterior, Boundary, DimLine)
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
	for i := 0; i < mls.NumGeometries(); i++ {
		ls2 := mls.GeometryN(i).(*geom.LineString)
		computeLineLineRelate(ls, ls2, m)
	}
}

// computeLineMultiPolygonRelate computes LineString vs MultiPolygon.
func computeLineMultiPolygonRelate(ls *geom.LineString, mp *geom.MultiPolygon, m *IntersectionMatrix) {
	for i := 0; i < mp.NumGeometries(); i++ {
		poly := mp.GeometryN(i).(*geom.Polygon)
		computeLinePolygonRelate(ls, poly, m)
	}
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
	for i := 0; i < mp.NumGeometries(); i++ {
		poly2 := mp.GeometryN(i).(*geom.Polygon)
		computePolygonPolygonRelate(poly, poly2, m)
	}
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
