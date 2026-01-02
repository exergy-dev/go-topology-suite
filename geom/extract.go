package geom

// ForEachPoint applies fn to each non-empty point in the geometry.
// Returns true if fn indicates to stop iteration.
func ForEachPoint(g Geometry, fn func(*Point) bool) bool {
	switch v := g.(type) {
	case *Point:
		if v.IsEmpty() {
			return false
		}
		return fn(v)
	case *MultiPoint:
		for i := 0; i < v.NumGeometries(); i++ {
			if ForEachPoint(v.GeometryN(i), fn) {
				return true
			}
		}
	case *GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			if ForEachPoint(v.GeometryN(i), fn) {
				return true
			}
		}
	}
	return false
}

// ForEachLineString applies fn to each non-empty line string in the geometry.
// Linear rings are excluded.
func ForEachLineString(g Geometry, fn func(*LineString) bool) bool {
	return forEachLineString(g, false, fn)
}

// ForEachLineStringWithRings applies fn to each non-empty line string in the geometry.
// Linear rings are included as line strings.
func ForEachLineStringWithRings(g Geometry, fn func(*LineString) bool) bool {
	return forEachLineString(g, true, fn)
}

func forEachLineString(g Geometry, includeLinearRing bool, fn func(*LineString) bool) bool {
	switch v := g.(type) {
	case *LineString:
		if v.IsEmpty() {
			return false
		}
		return fn(v)
	case *LinearRing:
		if !includeLinearRing || v.IsEmpty() {
			return false
		}
		return fn(v.LineString)
	case *MultiLineString:
		for i := 0; i < v.NumGeometries(); i++ {
			if forEachLineString(v.GeometryN(i), includeLinearRing, fn) {
				return true
			}
		}
	case *GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			if forEachLineString(v.GeometryN(i), includeLinearRing, fn) {
				return true
			}
		}
	}
	return false
}

// ForEachPolygon applies fn to each non-empty polygon in the geometry.
func ForEachPolygon(g Geometry, fn func(*Polygon) bool) bool {
	switch v := g.(type) {
	case *Polygon:
		if v.IsEmpty() {
			return false
		}
		return fn(v)
	case *MultiPolygon:
		for i := 0; i < v.NumGeometries(); i++ {
			if ForEachPolygon(v.GeometryN(i), fn) {
				return true
			}
		}
	case *GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			if ForEachPolygon(v.GeometryN(i), fn) {
				return true
			}
		}
	}
	return false
}

// ExtractPoints returns all non-empty point components in the geometry.
func ExtractPoints(g Geometry) []*Point {
	var points []*Point
	ForEachPoint(g, func(p *Point) bool {
		points = append(points, p)
		return false
	})
	return points
}

// ExtractLineStrings returns all non-empty line string components in the geometry.
// Linear rings are excluded.
func ExtractLineStrings(g Geometry) []*LineString {
	var lines []*LineString
	ForEachLineString(g, func(ls *LineString) bool {
		lines = append(lines, ls)
		return false
	})
	return lines
}

// ExtractLineStringsWithRings returns all non-empty line string components in the geometry.
// Linear rings are included as line strings.
func ExtractLineStringsWithRings(g Geometry) []*LineString {
	var lines []*LineString
	ForEachLineStringWithRings(g, func(ls *LineString) bool {
		lines = append(lines, ls)
		return false
	})
	return lines
}

// ExtractPolygons returns all non-empty polygon components in the geometry.
func ExtractPolygons(g Geometry) []*Polygon {
	var polygons []*Polygon
	ForEachPolygon(g, func(p *Polygon) bool {
		polygons = append(polygons, p)
		return false
	})
	return polygons
}
