package relateng

import "github.com/terra-geo/terra/geom"

// Operand-side flags. JTS uses a boolean isA; we expose named
// constants for readability at call sites.
const (
	GeomA = true
	GeomB = false
)

// Geometry wraps a terra geom.Geometry with cached metadata used
// throughout RelateNG: envelope, dimension, hasPoints/hasLines/hasAreas
// flags, and a lazily constructed PointLocator.
//
// Port of org.locationtech.jts.operation.relateng.RelateGeometry.
//
// Once constructed the wrapper is read-mostly (the locator is built
// on first locate*) so an instance can be reused across many
// predicate calls — that's the JTS "prepared" mode.
type Geometry struct {
	geom         geom.Geometry
	isPrepared   bool
	rule         BoundaryNodeRule
	envelope     geom.Envelope
	dim          int
	hasPoints    bool
	hasLines     bool
	hasAreas     bool
	isLineZeroLen bool
	isEmpty      bool
	locator      *PointLocator
}

// NewGeometry wraps g with the OGC SFS rule (default).
func NewGeometry(g geom.Geometry) *Geometry {
	return NewGeometryRule(g, false, OGCSFSBoundaryRule)
}

// NewGeometryRule wraps g with explicit prepared/rule settings.
func NewGeometryRule(g geom.Geometry, isPrepared bool, rule BoundaryNodeRule) *Geometry {
	if rule == nil {
		rule = OGCSFSBoundaryRule
	}
	rg := &Geometry{
		geom:       g,
		isPrepared: isPrepared,
		rule:       rule,
		dim:        DimFalse,
	}
	if g != nil {
		rg.envelope = g.Envelope()
		rg.isEmpty = g.IsEmpty()
		rg.analyzeDimensions()
		rg.isLineZeroLen = rg.dim == DimL && isAllZeroLength(g)
	} else {
		rg.isEmpty = true
	}
	return rg
}

// Geometry returns the wrapped geometry.
func (g *Geometry) Geometry() geom.Geometry { return g.geom }

// IsPrepared reports whether the wrapper is in prepared mode.
func (g *Geometry) IsPrepared() bool { return g.isPrepared }

// Envelope returns the cached envelope.
func (g *Geometry) Envelope() geom.Envelope { return g.envelope }

// Dimension returns the maximal element dimension (P/L/A) present
// in the geometry. Empty geometries report DimFalse.
func (g *Geometry) Dimension() int { return g.dim }

// HasDimension reports whether the geometry contains any element
// of the given dimension.
func (g *Geometry) HasDimension(dim int) bool {
	switch dim {
	case DimP:
		return g.hasPoints
	case DimL:
		return g.hasLines
	case DimA:
		return g.hasAreas
	}
	return false
}

// HasAreaAndLine is true for mixed-dim collections containing both
// areas and lines.
func (g *Geometry) HasAreaAndLine() bool { return g.hasAreas && g.hasLines }

// DimensionReal returns the *non-empty* dimension: a zero-length
// LineString is reported as DimP (it is topologically a point),
// and a collection's dimension is the max of its non-empty members.
func (g *Geometry) DimensionReal() int {
	if g.isEmpty {
		return DimFalse
	}
	if g.dim == DimL && g.isLineZeroLen {
		return DimP
	}
	if g.hasAreas {
		return DimA
	}
	if g.hasLines {
		return DimL
	}
	return DimP
}

// HasEdges reports whether the geometry has any 1D or 2D parts.
func (g *Geometry) HasEdges() bool { return g.hasLines || g.hasAreas }

// IsEmpty mirrors the wrapped geometry's IsEmpty.
func (g *Geometry) IsEmpty() bool { return g.isEmpty }

// IsPolygonal is true for Polygon and MultiPolygon (exclusively;
// GeometryCollections containing only polygons are not "polygonal"
// for RelateNG purposes — overlapping members may need adjacent-edge
// locator handling).
func (g *Geometry) IsPolygonal() bool {
	switch g.geom.(type) {
	case *geom.Polygon, *geom.MultiPolygon:
		return true
	}
	return false
}

// IsSelfNodingRequired reports whether the geometry requires
// self-noding (its component edges may cross). Lines and mixed
// GeometryCollections do; pure point/polygon inputs do not.
func (g *Geometry) IsSelfNodingRequired() bool {
	switch g.geom.(type) {
	case *geom.Point, *geom.MultiPoint, *geom.Polygon, *geom.MultiPolygon:
		return false
	}
	if gc, ok := g.geom.(*geom.GeometryCollection); ok {
		// A GC with a single polygonal member needs no noding.
		if g.hasAreas && gc.NumGeometries() == 1 {
			return false
		}
	}
	if !g.hasAreas && !g.hasLines {
		return false
	}
	return true
}

// HasBoundary reports whether the linear part of the geometry has
// any boundary points under the configured rule. Lazily constructs
// the underlying PointLocator.
func (g *Geometry) HasBoundary() bool {
	return g.getLocator().HasBoundary()
}

// LocateWithDim is a convenience wrapping the underlying
// PointLocator.LocateWithDim.
func (g *Geometry) LocateWithDim(p geom.XY) int {
	return g.getLocator().LocateWithDim(p)
}

// LocateLineEndWithDim wraps PointLocator.LocateLineEndWithDim.
func (g *Geometry) LocateLineEndWithDim(p geom.XY) int {
	return g.getLocator().LocateLineEndWithDim(p)
}

// LocateNode wraps PointLocator.LocateNode.
func (g *Geometry) LocateNode(p geom.XY, parentPoly geom.Geometry) int {
	return g.getLocator().LocateNode(p, parentPoly)
}

// LocateNodeWithDim wraps PointLocator.LocateNodeWithDim.
func (g *Geometry) LocateNodeWithDim(p geom.XY, parentPoly geom.Geometry) int {
	return g.getLocator().LocateNodeWithDim(p, parentPoly)
}

// LocateAreaVertex matches JTS RelateGeometry.locateAreaVertex: the
// parent polygon is passed as nil because the point is itself a
// vertex and will resolve to BOUNDARY without needing a parent
// reference.
func (g *Geometry) LocateAreaVertex(p geom.XY) int {
	return g.LocateNode(p, nil)
}

// IsNodeInArea checks whether the supplied node point lies in the
// interior of an area component (i.e. classifies as AREA_INTERIOR).
// Used by the topology computer to determine whether an edge node
// "punches through" a polygon.
func (g *Geometry) IsNodeInArea(p geom.XY, parentPoly geom.Geometry) bool {
	return g.LocateNodeWithDim(p, parentPoly) == DLAreaInterior
}

func (g *Geometry) getLocator() *PointLocator {
	if g.locator == nil {
		g.locator = NewPointLocatorRule(g.geom, g.isPrepared, g.rule)
	}
	return g.locator
}

// analyzeDimensions populates hasPoints/hasLines/hasAreas and the
// max dim across single-typed inputs and (recursively) collections.
func (g *Geometry) analyzeDimensions() {
	if g.isEmpty {
		return
	}
	switch v := g.geom.(type) {
	case *geom.Point, *geom.MultiPoint:
		_ = v
		g.hasPoints = true
		g.dim = DimP
	case *geom.LineString, *geom.LinearRing, *geom.MultiLineString:
		g.hasLines = true
		g.dim = DimL
	case *geom.Polygon, *geom.MultiPolygon:
		g.hasAreas = true
		g.dim = DimA
	case *geom.GeometryCollection:
		analyzeCollection(v, &g.hasPoints, &g.hasLines, &g.hasAreas, &g.dim)
	}
}

func analyzeCollection(gc *geom.GeometryCollection, hp, hl, ha *bool, dim *int) {
	for i := 0; i < gc.NumGeometries(); i++ {
		c := gc.GeometryAt(i)
		if c.IsEmpty() {
			continue
		}
		switch v := c.(type) {
		case *geom.Point, *geom.MultiPoint:
			_ = v
			*hp = true
			if *dim < DimP {
				*dim = DimP
			}
		case *geom.LineString, *geom.LinearRing, *geom.MultiLineString:
			*hl = true
			if *dim < DimL {
				*dim = DimL
			}
		case *geom.Polygon, *geom.MultiPolygon:
			*ha = true
			if *dim < DimA {
				*dim = DimA
			}
		case *geom.GeometryCollection:
			analyzeCollection(v, hp, hl, ha, dim)
		}
	}
}

// isAllZeroLength reports whether every linear element in g is
// zero-length (every vertex equal to vertex 0). For non-linear
// geometries the result is meaningless; callers must check the
// geometry's dimension first (see DimensionReal).
func isAllZeroLength(g geom.Geometry) bool {
	switch v := g.(type) {
	case *geom.LineString:
		return lineStringZeroLen(v)
	case *geom.LinearRing:
		return lineStringZeroLen(v.AsLineString())
	case *geom.MultiLineString:
		for i := 0; i < v.NumGeometries(); i++ {
			if !lineStringZeroLen(v.LineStringAt(i)) {
				return false
			}
		}
		return true
	case *geom.GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			c := v.GeometryAt(i)
			switch c.(type) {
			case *geom.LineString, *geom.LinearRing, *geom.MultiLineString:
				if !isAllZeroLength(c) {
					return false
				}
			}
		}
		return true
	}
	return false
}

func lineStringZeroLen(ls *geom.LineString) bool {
	if ls == nil || ls.NumPoints() < 2 {
		return true
	}
	p0 := ls.PointAt(0)
	for i := 1; i < ls.NumPoints(); i++ {
		if ls.PointAt(i) != p0 {
			return false
		}
	}
	return true
}

// Name returns "A" or "B" for the standard boolean operand flag.
// Mirrors RelateGeometry.name.
func Name(isA bool) string {
	if isA {
		return "A"
	}
	return "B"
}
