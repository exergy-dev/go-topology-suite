package spherical

import (
	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/golang/geo/s2"
)

// CellID returns the S2 cell ID for a point at the finest level (level 30).
// This provides the most precise cell identification for a point.
// Returns 0 if the point is nil or empty.
func CellID(p *geom.Point) s2.CellID {
	return CellIDAtLevel(p, 30)
}

// CellIDAtLevel returns the S2 cell ID at a specific level (0-30).
// Level 0 is the coarsest (6 cells covering the entire sphere).
// Level 30 is the finest (cell side length ~1cm).
//
// Typical levels:
//   - Level 0-2: Continental scale (millions of km²)
//   - Level 10: City scale (~1000 km²)
//   - Level 15: Neighborhood scale (~10 km²)
//   - Level 20: Building scale (~400 m²)
//   - Level 30: Centimeter scale (~1 cm²)
//
// Returns 0 if the point is nil or empty, or if level is out of range.
func CellIDAtLevel(p *geom.Point, level int) s2.CellID {
	if p == nil || p.IsEmpty() {
		return 0
	}

	if level < 0 || level > 30 {
		return 0
	}

	s2Point := ToS2Point(p)
	cellID := s2.CellFromPoint(s2Point).ID()
	return cellID.Parent(level)
}

// CellToken returns the S2 cell token string for a point at a specific level.
// Cell tokens are compact string representations that can be used for indexing.
// Returns an empty string if the point is nil or empty.
//
// Cell tokens are hierarchical - longer tokens represent smaller areas.
// They use base-16 encoding and have the property that cells at the same level
// have tokens of the same length.
func CellToken(p *geom.Point, level int) string {
	cellID := CellIDAtLevel(p, level)
	if cellID == 0 {
		return ""
	}
	return cellID.ToToken()
}

// Covering returns S2 cell IDs that cover a geometry.
// The covering uses cells between minLevel and maxLevel, with at most maxCells cells.
//
// Parameters:
//   - g: The geometry to cover (Point, LineString, Polygon, Multi*, GeometryCollection)
//   - minLevel: Minimum cell level (0-30). Larger cells, fewer total cells.
//   - maxLevel: Maximum cell level (0-30). Smaller cells, more precise coverage.
//   - maxCells: Maximum number of cells in the covering. More cells = better fit.
//
// Returns an empty slice if the geometry is nil or empty, or if parameters are invalid.
// For collection types (MultiPoint, MultiLineString, MultiPolygon, GeometryCollection),
// the covering is computed for each component and merged.
//
// Example usage:
//
//	cells := Covering(polygon, 10, 20, 8)  // 8 cells between city and building scale
func Covering(g geom.Geometry, minLevel, maxLevel, maxCells int) []s2.CellID {
	if g == nil || g.IsEmpty() {
		return nil
	}

	if minLevel < 0 || minLevel > 30 || maxLevel < 0 || maxLevel > 30 || minLevel > maxLevel {
		return nil
	}

	if maxCells <= 0 {
		maxCells = 8 // Default
	}

	rc := &s2.RegionCoverer{
		MinLevel: minLevel,
		MaxLevel: maxLevel,
		MaxCells: maxCells,
	}

	cellUnion := coverGeometry(g, rc, false)
	if len(cellUnion) == 0 {
		return nil
	}

	return []s2.CellID(cellUnion)
}

// CoveringTokens returns S2 cell tokens that cover a geometry.
// This is a convenience wrapper around Covering that returns tokens instead of cell IDs.
//
// Parameters are the same as Covering.
// Returns an empty slice if the geometry is nil or empty, or if parameters are invalid.
func CoveringTokens(g geom.Geometry, minLevel, maxLevel, maxCells int) []string {
	cellIDs := Covering(g, minLevel, maxLevel, maxCells)
	if len(cellIDs) == 0 {
		return nil
	}

	tokens := make([]string, len(cellIDs))
	for i, cellID := range cellIDs {
		tokens[i] = cellID.ToToken()
	}
	return tokens
}

// InteriorCovering returns S2 cell IDs that are completely contained within a geometry.
// Unlike Covering (which may include cells that partially overlap), InteriorCovering
// only returns cells that are entirely inside the geometry.
//
// This is useful for indexing when you want to be sure that all points in the cell
// are definitely inside the geometry.
// For non-area geometries (points/lines), InteriorCovering returns nil.
//
// For collection types (MultiPoint, MultiLineString, MultiPolygon, GeometryCollection),
// the interior covering is computed for each component and merged.
//
// Parameters are the same as Covering.
func InteriorCovering(g geom.Geometry, minLevel, maxLevel, maxCells int) []s2.CellID {
	if g == nil || g.IsEmpty() {
		return nil
	}

	if minLevel < 0 || minLevel > 30 || maxLevel < 0 || maxLevel > 30 || minLevel > maxLevel {
		return nil
	}

	if maxCells <= 0 {
		maxCells = 8
	}

	rc := &s2.RegionCoverer{
		MinLevel: minLevel,
		MaxLevel: maxLevel,
		MaxCells: maxCells,
	}

	cellUnion := coverGeometry(g, rc, true)
	if len(cellUnion) == 0 {
		return nil
	}

	return []s2.CellID(cellUnion)
}

// InteriorCoveringTokens returns S2 cell tokens for cells completely contained within a geometry.
// This is a convenience wrapper around InteriorCovering.
func InteriorCoveringTokens(g geom.Geometry, minLevel, maxLevel, maxCells int) []string {
	cellIDs := InteriorCovering(g, minLevel, maxLevel, maxCells)
	if len(cellIDs) == 0 {
		return nil
	}

	tokens := make([]string, len(cellIDs))
	for i, cellID := range cellIDs {
		tokens[i] = cellID.ToToken()
	}
	return tokens
}

// CellUnion returns an S2 CellUnion covering the geometry.
// A CellUnion is a normalized set of cells that can be used for efficient
// spatial queries and set operations.
func CellUnion(g geom.Geometry, minLevel, maxLevel, maxCells int) s2.CellUnion {
	if g == nil || g.IsEmpty() {
		return nil
	}

	if minLevel < 0 || minLevel > 30 || maxLevel < 0 || maxLevel > 30 || minLevel > maxLevel {
		return nil
	}

	if maxCells <= 0 {
		maxCells = 8
	}

	rc := &s2.RegionCoverer{
		MinLevel: minLevel,
		MaxLevel: maxLevel,
		MaxCells: maxCells,
	}

	cellUnion := coverGeometry(g, rc, false)
	if len(cellUnion) == 0 {
		return nil
	}

	cellUnion.Normalize()
	return cellUnion
}

// geometryToRegion converts a GTS geometry to an S2 region.
// Returns nil if the geometry type is not supported or conversion fails.
func geometryToRegion(g geom.Geometry) s2.Region {
	if g == nil || g.IsEmpty() {
		return nil
	}

	switch gt := g.(type) {
	case *geom.Point:
		// Point as a cell at max level
		s2Point := ToS2Point(gt)
		return s2.CellFromPoint(s2Point)

	case *geom.LineString:
		// LineString as polyline
		polyline := ToS2Polyline(gt)
		if polyline == nil {
			return nil
		}
		return polyline

	case *geom.Polygon:
		// Polygon as S2 polygon
		return ToS2Polygon(gt)

	case *geom.LinearRing:
		// LinearRing as S2 loop
		return ToS2Loop(gt)

	default:
		return nil
	}
}

func coverGeometry(g geom.Geometry, rc *s2.RegionCoverer, interior bool) s2.CellUnion {
	if g == nil || g.IsEmpty() {
		return nil
	}

	if interior && g.Dimension() != geom.DimensionArea {
		return nil
	}

	switch gt := g.(type) {
	case *geom.MultiPoint:
		return coverCollection(gt.NumGeometries(), gt.GeometryN, rc, interior)
	case *geom.MultiLineString:
		return coverCollection(gt.NumGeometries(), gt.GeometryN, rc, interior)
	case *geom.MultiPolygon:
		return coverCollection(gt.NumGeometries(), gt.GeometryN, rc, interior)
	case *geom.GeometryCollection:
		return coverCollection(gt.NumGeometries(), gt.GeometryN, rc, interior)
	default:
		region := geometryToRegion(g)
		if region == nil {
			return nil
		}
		if interior {
			return rc.InteriorCovering(region)
		}
		return rc.Covering(region)
	}
}

func coverCollection(count int, geometryAt func(int) geom.Geometry, rc *s2.RegionCoverer, interior bool) s2.CellUnion {
	var result s2.CellUnion
	for i := 0; i < count; i++ {
		covering := coverGeometry(geometryAt(i), rc, interior)
		if len(covering) > 0 {
			result = append(result, covering...)
		}
	}
	if len(result) == 0 {
		return nil
	}
	result.Normalize()
	return result
}

// CellFromToken converts a cell token string back to a cell ID.
// Returns 0 if the token is invalid.
func CellFromToken(token string) s2.CellID {
	return s2.CellIDFromToken(token)
}

// CellLevel returns the level of a cell ID (0-30).
// Returns -1 if the cell ID is invalid.
func CellLevel(cellID s2.CellID) int {
	if !cellID.IsValid() {
		return -1
	}
	return cellID.Level()
}

// CellsIntersect checks if two cell IDs intersect (share any area).
// This is true if one cell contains the other, or if they overlap.
func CellsIntersect(c1, c2 s2.CellID) bool {
	if !c1.IsValid() || !c2.IsValid() {
		return false
	}
	return c1.Intersects(c2)
}

// CellContains checks if cell c1 contains cell c2.
// A cell contains another if c2 is a descendant of c1 in the cell hierarchy.
func CellContains(c1, c2 s2.CellID) bool {
	if !c1.IsValid() || !c2.IsValid() {
		return false
	}
	return c1.Contains(c2)
}

// GeometryFromCellID converts a cell ID to a polygon representing the cell's area.
// Returns nil if the cell ID is invalid.
func GeometryFromCellID(cellID s2.CellID) *geom.Polygon {
	if !cellID.IsValid() {
		return nil
	}

	cell := s2.CellFromCellID(cellID)
	coords := make(geom.CoordinateSequence, 5) // 4 vertices + closing

	for i := 0; i < 4; i++ {
		vertex := cell.Vertex(i)
		ll := s2.LatLngFromPoint(vertex)
		coords[i] = FromS2LatLng(ll)
	}
	// Close the ring
	coords[4] = coords[0]

	ring := geom.NewLinearRing(coords)
	return geom.NewPolygon(ring, nil)
}

// GeometryFromCellToken converts a cell token to a polygon representing the cell's area.
// Returns nil if the token is invalid.
func GeometryFromCellToken(token string) *geom.Polygon {
	cellID := CellFromToken(token)
	return GeometryFromCellID(cellID)
}
