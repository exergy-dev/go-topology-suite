package geom

import (
	"iter"

	"github.com/terra-geo/terra/crs"
)

// LineString is an ordered sequence of two or more vertices joined by
// straight (or great-circle, depending on the kernel) edges.
type LineString struct {
	baseGeom
}

// NewLineString constructs a LineString from a slice of XY coordinates.
// The input is cloned; the caller retains ownership.
func NewLineString(c *crs.CRS, pts []XY) *LineString {
	flat := make([]float64, 0, 2*len(pts))
	for _, p := range pts {
		flat = append(flat, p.X, p.Y)
	}
	return &LineString{baseGeom{layout: LayoutXY, coords: flat, crs: c}}
}

// NewLineStringFlat constructs a LineString directly from a flat coordinate
// buffer. The buffer is cloned. Callers wanting zero-copy behaviour should
// donate ownership via NewLineStringFlatNoClone (intended for format
// decoders only).
func NewLineStringFlat(layout Layout, c *crs.CRS, flat []float64) *LineString {
	return &LineString{baseGeom{layout: layout, coords: cloneFloats(flat), crs: c}}
}

// NewLineStringFlatNoClone takes ownership of flat without copying. Intended
// for format decoders that have just allocated the buffer themselves.
func NewLineStringFlatNoClone(layout Layout, c *crs.CRS, flat []float64) *LineString {
	return &LineString{baseGeom{layout: layout, coords: flat, crs: c}}
}

func (ls *LineString) isGeometry()       {}
func (ls *LineString) Type() Type        { return LineStringType }
func (ls *LineString) Envelope() Envelope  { return ls.envelope() }
func (ls *LineString) IsEmpty() bool       { return len(ls.coords) == 0 }
func (ls *LineString) NumGeometries() int  { return 1 }

// NumPoints returns the number of vertices in the line string.
func (ls *LineString) NumPoints() int { return ls.numCoords() }

// PointAt returns the i-th vertex projected to XY. Panics on out-of-range
// i — programmer error, not a runtime failure mode.
func (ls *LineString) PointAt(i int) XY {
	stride := ls.stride()
	off := i * stride
	return XY{ls.coords[off], ls.coords[off+1]}
}

// IsClosed reports whether the line string is closed: i.e. the first and
// last vertices coincide (under XY.Equal). An empty line string is not
// closed. Mirrors JTS LineString.isClosed().
func (ls *LineString) IsClosed() bool {
	n := ls.numCoords()
	if n < 2 {
		return false
	}
	return ls.PointAt(0).Equal(ls.PointAt(n - 1))
}

// XYs returns the line string's vertices as a fresh []XY slice. The result
// is independent of the LineString's internal storage; mutating it does not
// affect the geometry.
func (ls *LineString) XYs() []XY {
	n := ls.numCoords()
	stride := ls.stride()
	out := make([]XY, n)
	for i, off := 0, 0; i < n; i, off = i+1, off+stride {
		out[i] = XY{ls.coords[off], ls.coords[off+1]}
	}
	return out
}

// CoordsXY returns a range-over-func iterator yielding each vertex as XY.
// Use:
//
//	for p := range ls.CoordsXY() {
//	    fmt.Println(p.X, p.Y)
//	}
func (ls *LineString) CoordsXY() iter.Seq[XY] {
	stride := ls.stride()
	coords := ls.coords
	return func(yield func(XY) bool) {
		for i := 0; i+1 < len(coords); i += stride {
			if !yield(XY{coords[i], coords[i+1]}) {
				return
			}
		}
	}
}
