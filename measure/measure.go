package measure

import (
	"math"

	terra "github.com/terra-geo/terra"
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel"
	"github.com/terra-geo/terra/kernel/geodesic"
	"github.com/terra-geo/terra/kernel/planar"
)

// Option configures a measurement call.
type Option func(*config)

type config struct {
	kernel    kernel.Kernel
	kernelSet bool
}

// WithKernel selects the geometric kernel. When omitted the kernel is
// chosen by CRS: geographic → geodesic (accurate metres on WGS84),
// projected (or no CRS) → planar (units of the projection).
func WithKernel(k kernel.Kernel) Option {
	return func(c *config) { c.kernel = k; c.kernelSet = true }
}

// resolve chooses the kernel given the geometry. Geographic CRSes route
// to the geodesic kernel (accurate area/length in metres); projected and
// CRS-less geometries use planar.
func resolve(g geom.Geometry, opts []Option) config {
	c := config{}
	for _, opt := range opts {
		opt(&c)
	}
	if !c.kernelSet {
		c.kernel = defaultKernel(g)
	}
	return c
}

func defaultKernel(g geom.Geometry) kernel.Kernel {
	if g != nil && g.CRS().IsGeographic() {
		return geodesic.Default
	}
	return planar.Default
}

// Distance returns the kernel-appropriate distance between a and b.
// Returns 0 if either is empty (with no error). CRS mismatch returns
// ErrCRSMismatch and a NaN distance.
func Distance(a, b geom.Geometry, opts ...Option) (float64, error) {
	if !crs.Equal(a.CRS(), b.CRS()) {
		return math.NaN(), terra.ErrCRSMismatch
	}
	if a.IsEmpty() || b.IsEmpty() {
		return 0, nil
	}
	c := resolve(a, opts)
	d := geometryDistance(a, b, c.kernel)
	return d, nil
}

func geometryDistance(a, b geom.Geometry, k kernel.Kernel) float64 {
	// Quick path for point-point.
	if pa, ok := a.(*geom.Point); ok {
		if pb, ok := b.(*geom.Point); ok {
			return k.Distance(pa.XY(), pb.XY())
		}
	}
	// Otherwise: minimum distance between any vertex of a and any segment of b
	// (and vice versa). This is correct for non-overlapping geometries.
	min := math.Inf(+1)
	visitVertices(a, func(p geom.XY) {
		visitSegments(b, func(s1, s2 geom.XY) {
			if d := k.SegmentDistance(p, s1, s2); d < min {
				min = d
			}
		})
	})
	visitVertices(b, func(p geom.XY) {
		visitSegments(a, func(s1, s2 geom.XY) {
			if d := k.SegmentDistance(p, s1, s2); d < min {
				min = d
			}
		})
	})
	if math.IsInf(min, +1) {
		return 0
	}
	return min
}

// Length returns the total length of all linear components in g.
// Polygons return the perimeter (outer + holes).
func Length(g geom.Geometry, opts ...Option) float64 {
	c := resolve(g, opts)
	var total float64
	visitSegments(g, func(s1, s2 geom.XY) {
		total += c.kernel.Distance(s1, s2)
	})
	return total
}

// Area returns the area of polygonal components. Lines and points return 0.
// Holes subtract from the outer ring.
func Area(g geom.Geometry, opts ...Option) float64 {
	c := resolve(g, opts)
	switch v := g.(type) {
	case *geom.Polygon:
		return polygonArea(v, c.kernel)
	case *geom.MultiPolygon:
		var total float64
		for i := 0; i < v.NumGeometries(); i++ {
			total += polygonArea(v.PolygonAt(i), c.kernel)
		}
		return total
	case *geom.GeometryCollection:
		var total float64
		for i := 0; i < v.NumGeometries(); i++ {
			total += Area(v.GeometryAt(i), opts...)
		}
		return total
	}
	return 0
}

func polygonArea(p *geom.Polygon, k kernel.Kernel) float64 {
	if p.NumRings() == 0 {
		return 0
	}
	outer := math.Abs(k.RingArea(p.Ring(0)))
	for r := 1; r < p.NumRings(); r++ {
		outer -= math.Abs(k.RingArea(p.Ring(r)))
	}
	return outer
}

// Centroid returns the geometric centroid as a Point in the same CRS.
// For points: the point itself. For lines: length-weighted midpoint of
// segments. For polygons: signed-area centroid (Bashein-Detmer).
//
// The kernel selects how segment lengths and ring areas are computed,
// which affects the relative weights of sub-geometries on a
// MultiLineString / MultiPolygon and the length-weighting of segments
// on a LineString. The per-polygon centroid is the planar shoelace
// regardless of kernel; see the package doc.
func Centroid(g geom.Geometry, opts ...Option) *geom.Point {
	if g.IsEmpty() {
		return geom.NewEmptyPoint(g.CRS(), geom.LayoutXY)
	}
	c := resolve(g, opts)
	switch v := g.(type) {
	case *geom.Point:
		return geom.NewPoint(v.CRS(), v.XY())
	case *geom.LineString:
		return lineStringCentroid(v, c.kernel)
	case *geom.Polygon:
		return polygonCentroid(v)
	case *geom.MultiPoint:
		var sx, sy float64
		n := v.NumGeometries()
		for i := 0; i < n; i++ {
			p := v.PointAt(i)
			sx += p.X
			sy += p.Y
		}
		return geom.NewPoint(v.CRS(), geom.XY{X: sx / float64(n), Y: sy / float64(n)})
	case *geom.MultiLineString:
		return multiLineCentroid(v, c.kernel)
	case *geom.MultiPolygon:
		return multiPolygonCentroid(v, c.kernel)
	}
	return geom.NewEmptyPoint(g.CRS(), geom.LayoutXY)
}

func lineStringCentroid(ls *geom.LineString, k kernel.Kernel) *geom.Point {
	n := ls.NumPoints()
	if n == 0 {
		return geom.NewEmptyPoint(ls.CRS(), geom.LayoutXY)
	}
	if n == 1 {
		return geom.NewPoint(ls.CRS(), ls.PointAt(0))
	}
	var totalLen, cx, cy float64
	for i := 0; i+1 < n; i++ {
		a, b := ls.PointAt(i), ls.PointAt(i+1)
		segLen := k.Distance(a, b)
		mid := k.Midpoint(a, b)
		totalLen += segLen
		cx += mid.X * segLen
		cy += mid.Y * segLen
	}
	if totalLen == 0 {
		return geom.NewPoint(ls.CRS(), ls.PointAt(0))
	}
	return geom.NewPoint(ls.CRS(), geom.XY{X: cx / totalLen, Y: cy / totalLen})
}

func polygonCentroid(p *geom.Polygon) *geom.Point {
	if p.NumRings() == 0 {
		return geom.NewEmptyPoint(p.CRS(), geom.LayoutXY)
	}
	var sx, sy, sa float64
	for r := 0; r < p.NumRings(); r++ {
		ring := p.Ring(r)
		rx, ry, ra := ringCentroidContribution(ring)
		sign := 1.0
		if r > 0 {
			sign = -1.0
		}
		sx += sign * rx
		sy += sign * ry
		sa += sign * ra
	}
	if sa == 0 {
		// Degenerate; fall back to vertex average.
		ring := p.Ring(0)
		var x, y float64
		for _, v := range ring {
			x += v.X
			y += v.Y
		}
		n := float64(len(ring))
		return geom.NewPoint(p.CRS(), geom.XY{X: x / n, Y: y / n})
	}
	return geom.NewPoint(p.CRS(), geom.XY{X: sx / sa, Y: sy / sa})
}

// ringCentroidContribution returns 6*Cx*A, 6*Cy*A, 6*A for the ring (the
// uncomputed denominators of the polygon centroid sum).
func ringCentroidContribution(ring []geom.XY) (cxNum, cyNum, areaSum float64) {
	for i := 0; i+1 < len(ring); i++ {
		x0, y0 := ring[i].X, ring[i].Y
		x1, y1 := ring[i+1].X, ring[i+1].Y
		cross := x0*y1 - x1*y0
		cxNum += (x0 + x1) * cross
		cyNum += (y0 + y1) * cross
		areaSum += cross
	}
	// Multiply x/y numerators by (1/3) and divide by area*2 = areaSum.
	// Combined: cxNum / (3*areaSum) = centroid X.
	return cxNum / 3, cyNum / 3, areaSum
}

func multiLineCentroid(m *geom.MultiLineString, k kernel.Kernel) *geom.Point {
	var sx, sy, totalLen float64
	for i := 0; i < m.NumGeometries(); i++ {
		ls := m.LineStringAt(i)
		c := lineStringCentroid(ls, k)
		var segLen float64
		for j := 0; j+1 < ls.NumPoints(); j++ {
			a, b := ls.PointAt(j), ls.PointAt(j+1)
			segLen += k.Distance(a, b)
		}
		sx += c.XY().X * segLen
		sy += c.XY().Y * segLen
		totalLen += segLen
	}
	if totalLen == 0 {
		return geom.NewEmptyPoint(m.CRS(), geom.LayoutXY)
	}
	return geom.NewPoint(m.CRS(), geom.XY{X: sx / totalLen, Y: sy / totalLen})
}

func multiPolygonCentroid(m *geom.MultiPolygon, k kernel.Kernel) *geom.Point {
	var sx, sy, totalArea float64
	for i := 0; i < m.NumGeometries(); i++ {
		p := m.PolygonAt(i)
		c := polygonCentroid(p)
		a := polygonArea(p, k)
		sx += c.XY().X * a
		sy += c.XY().Y * a
		totalArea += a
	}
	if totalArea == 0 {
		return geom.NewEmptyPoint(m.CRS(), geom.LayoutXY)
	}
	return geom.NewPoint(m.CRS(), geom.XY{X: sx / totalArea, Y: sy / totalArea})
}

// visitVertices yields every vertex in g.
func visitVertices(g geom.Geometry, fn func(geom.XY)) {
	switch v := g.(type) {
	case *geom.Point:
		if !v.IsEmpty() {
			fn(v.XY())
		}
	case *geom.LineString:
		for i := 0; i < v.NumPoints(); i++ {
			fn(v.PointAt(i))
		}
	case *geom.Polygon:
		for r := 0; r < v.NumRings(); r++ {
			for _, p := range v.Ring(r) {
				fn(p)
			}
		}
	case *geom.MultiPoint:
		for i := 0; i < v.NumGeometries(); i++ {
			fn(v.PointAt(i))
		}
	case *geom.MultiLineString:
		for i := 0; i < v.NumGeometries(); i++ {
			visitVertices(v.LineStringAt(i), fn)
		}
	case *geom.MultiPolygon:
		for i := 0; i < v.NumGeometries(); i++ {
			visitVertices(v.PolygonAt(i), fn)
		}
	case *geom.GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			visitVertices(v.GeometryAt(i), fn)
		}
	}
}

// visitSegments yields every linear segment in g (Polygon edges + LineString
// edges). Points contribute nothing.
func visitSegments(g geom.Geometry, fn func(a, b geom.XY)) {
	switch v := g.(type) {
	case *geom.LineString:
		for i := 0; i+1 < v.NumPoints(); i++ {
			fn(v.PointAt(i), v.PointAt(i+1))
		}
	case *geom.Polygon:
		for r := 0; r < v.NumRings(); r++ {
			ring := v.Ring(r)
			for i := 0; i+1 < len(ring); i++ {
				fn(ring[i], ring[i+1])
			}
		}
	case *geom.MultiLineString:
		for i := 0; i < v.NumGeometries(); i++ {
			visitSegments(v.LineStringAt(i), fn)
		}
	case *geom.MultiPolygon:
		for i := 0; i < v.NumGeometries(); i++ {
			visitSegments(v.PolygonAt(i), fn)
		}
	case *geom.GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			visitSegments(v.GeometryAt(i), fn)
		}
	}
}
