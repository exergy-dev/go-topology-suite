package geom

import (
	"github.com/terra-geo/terra/crs"
)

// Polygon is a planar region bounded by an outer ring and zero or more
// inner rings (holes). Rings are closed line strings: the first and last
// vertex must coincide. Construction does not validate this — callers
// requiring validation should use validate.Validate.
//
// Storage: all rings live in a single flat coordinate buffer. ringStarts
// records the vertex index where each ring begins; ringStarts[0] is always 0
// and the implicit "ringStarts[len]" boundary equals the total vertex count.
type Polygon struct {
	baseGeom
	ringStarts []int // vertex offsets; len = numRings; first element is 0
}

// NewPolygon constructs a polygon from XY rings. The first ring is the
// outer shell; remaining rings are holes. Rings are cloned.
func NewPolygon(c *crs.CRS, rings ...[]XY) *Polygon {
	totalVerts := 0
	for _, r := range rings {
		totalVerts += len(r)
	}
	flat := make([]float64, 0, 2*totalVerts)
	starts := make([]int, 0, len(rings))
	off := 0
	for _, r := range rings {
		starts = append(starts, off)
		for _, p := range r {
			flat = append(flat, p.X, p.Y)
		}
		off += len(r)
	}
	return &Polygon{
		baseGeom:   baseGeom{layout: LayoutXY, coords: flat, crs: c},
		ringStarts: starts,
	}
}

// NewEmptyPolygon constructs a POLYGON EMPTY in the given layout.
func NewEmptyPolygon(c *crs.CRS, layout Layout) *Polygon {
	return &Polygon{baseGeom: baseGeom{layout: layout, crs: c}}
}

func (p *Polygon) isGeometry()       {}
func (p *Polygon) Type() Type        { return PolygonType }
func (p *Polygon) Envelope() Envelope  { return p.envelope() }
func (p *Polygon) IsEmpty() bool       { return len(p.coords) == 0 }
func (p *Polygon) NumGeometries() int  { return 1 }

// NumRings returns the number of rings (1 outer + n holes).
func (p *Polygon) NumRings() int { return len(p.ringStarts) }

// Ring returns the i-th ring (0 = exterior shell, 1..n = holes) as an
// XY slice. The slice is freshly allocated; callers may mutate it without
// affecting the polygon.
func (p *Polygon) Ring(i int) []XY {
	if i < 0 || i >= len(p.ringStarts) {
		return nil
	}
	stride := p.stride()
	startVertex := p.ringStarts[i]
	endVertex := p.numCoords()
	if i+1 < len(p.ringStarts) {
		endVertex = p.ringStarts[i+1]
	}
	out := make([]XY, 0, endVertex-startVertex)
	for v := startVertex; v < endVertex; v++ {
		off := v * stride
		out = append(out, XY{p.coords[off], p.coords[off+1]})
	}
	return out
}

// ExteriorRing returns the outer shell as XY.
func (p *Polygon) ExteriorRing() []XY {
	if p.NumRings() == 0 {
		return nil
	}
	return p.Ring(0)
}

// InteriorRings returns all holes (zero or more).
func (p *Polygon) InteriorRings() [][]XY {
	if p.NumRings() < 2 {
		return nil
	}
	out := make([][]XY, p.NumRings()-1)
	for i := 1; i < p.NumRings(); i++ {
		out[i-1] = p.Ring(i)
	}
	return out
}
