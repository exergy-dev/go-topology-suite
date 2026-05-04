package geojson

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/exergy-dev/go-topology-suite/geom"
)

// Option configures the GeoJSON writer.
type Option func(*config)

type config struct {
	// precision: -1 means default ('g' shortest round-trip); >=0 selects
	// strconv 'f' with that many decimal digits.
	precision int
	// forceCCW rewinds polygon rings to RFC 7946 orientation on output:
	// outer ring CCW, holes CW.
	forceCCW bool
}

func defaults() config { return config{precision: -1} }

// WithPrecision selects fixed-point output with the given number of decimal
// digits. Default behaviour is unchanged (Go's 'g' format, shortest
// round-trip representation).
//
// JTS reference: GeoJsonWriter(int decimals)
// (org.locationtech.jts.io.geojson.GeoJsonWriter).
func WithPrecision(decimals int) Option {
	return func(c *config) {
		if decimals < 0 {
			decimals = -1
		}
		c.precision = decimals
	}
}

// WithForceCCW rewinds polygon rings on output so that outer rings are
// counter-clockwise and holes are clockwise, per RFC 7946 §3.1.6. The
// stored geometry is not modified.
//
// JTS reference: GeoJsonWriter.setForceCCW
// (org.locationtech.jts.io.geojson.GeoJsonWriter).
func WithForceCCW() Option { return func(c *config) { c.forceCCW = true } }

// ringSignedArea returns 2*signed area of a closed ring (last==first).
// Positive => CCW, negative => CW. Local copy avoids depending on the
// algorithm package from the io layer.
func ringSignedArea(ring []geom.XY) float64 {
	if len(ring) < 4 {
		return 0
	}
	var sum float64
	for i := 0; i+1 < len(ring); i++ {
		sum += ring[i].X*ring[i+1].Y - ring[i+1].X*ring[i].Y
	}
	return sum
}

func reverseRing(ring []geom.XY) []geom.XY {
	out := make([]geom.XY, len(ring))
	for i, v := range ring {
		out[len(ring)-1-i] = v
	}
	return out
}

// orientedRing returns ring rewound to the orientation required for ring
// index r under RFC 7946 (outer CCW, holes CW). r==0 is the outer ring.
// The original slice is returned unchanged when already correct.
func orientedRing(ring []geom.XY, r int) []geom.XY {
	area := ringSignedArea(ring)
	wantCCW := r == 0
	isCCW := area > 0
	if isCCW == wantCCW {
		return ring
	}
	return reverseRing(ring)
}

// Marshal returns the GeoJSON encoding of g. CRS is implicit WGS84 per
// RFC 7946; any non-nil CRS attached to g is silently ignored on output.
func Marshal(g geom.Geometry, opts ...Option) ([]byte, error) {
	c := defaults()
	for _, o := range opts {
		o(&c)
	}
	var b strings.Builder
	if err := writeGeometry(&b, g, &c); err != nil {
		return nil, err
	}
	return []byte(b.String()), nil
}

func writeGeometry(b *strings.Builder, g geom.Geometry, c *config) error {
	switch v := g.(type) {
	case *geom.Point:
		return writePoint(b, v, c)
	case *geom.LineString:
		return writeLineString(b, v, c)
	case *geom.LinearRing:
		return writeLineString(b, v.AsLineString(), c)
	case *geom.Polygon:
		return writePolygon(b, v, c)
	case *geom.MultiPoint:
		return writeMultiPoint(b, v, c)
	case *geom.MultiLineString:
		return writeMultiLineString(b, v, c)
	case *geom.MultiPolygon:
		return writeMultiPolygon(b, v, c)
	case *geom.GeometryCollection:
		return writeGeometryCollection(b, v, c)
	default:
		return fmt.Errorf("geojson: unsupported %T", g)
	}
}

func writeNumber(b *strings.Builder, f float64, c *config) {
	if c != nil && c.precision >= 0 {
		b.WriteString(strconv.FormatFloat(f, 'f', c.precision, 64))
		return
	}
	b.WriteString(strconv.FormatFloat(f, 'g', -1, 64))
}

// writeVertex writes [x, y] or [x, y, z] depending on layout. M is
// dropped per RFC 7946 §3.1.1 (M not part of GeoJSON).
func writeVertex(b *strings.Builder, flat []float64, off int, layout geom.Layout, c *config) {
	stride := layout.Stride()
	b.WriteByte('[')
	writeNumber(b, flat[off], c)
	b.WriteByte(',')
	writeNumber(b, flat[off+1], c)
	if layout.HasZ() {
		// Z is at index 2 for XYZ and XYZM; XYM has no Z.
		zIdx := 2
		b.WriteByte(',')
		writeNumber(b, flat[off+zIdx], c)
	}
	_ = stride
	b.WriteByte(']')
}

func writePoint(b *strings.Builder, p *geom.Point, c *config) error {
	b.WriteString(`{"type":"Point","coordinates":`)
	if p.IsEmpty() {
		// RFC 7946 §3.1: GeoJSON has no first-class empty geometry; emit
		// an empty array. (A foreign "EMPTY" sentinel is sometimes used;
		// we deliberately do not emit it.)
		b.WriteString("[]}")
		return nil
	}
	writeVertex(b, p.FlatCoords(), 0, p.Layout(), c)
	b.WriteByte('}')
	return nil
}

func writeCoordSequence(b *strings.Builder, flat []float64, layout geom.Layout, c *config) {
	stride := layout.Stride()
	b.WriteByte('[')
	n := len(flat) / stride
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		writeVertex(b, flat, i*stride, layout, c)
	}
	b.WriteByte(']')
}

func writeLineString(b *strings.Builder, ls *geom.LineString, c *config) error {
	b.WriteString(`{"type":"LineString","coordinates":`)
	writeCoordSequence(b, ls.FlatCoords(), ls.Layout(), c)
	b.WriteByte('}')
	return nil
}

func writePolygon(b *strings.Builder, p *geom.Polygon, c *config) error {
	b.WriteString(`{"type":"Polygon","coordinates":[`)
	for r := 0; r < p.NumRings(); r++ {
		if r > 0 {
			b.WriteByte(',')
		}
		ring := p.Ring(r)
		if c != nil && c.forceCCW {
			ring = orientedRing(ring, r)
		}
		writeRing(b, ring, c)
	}
	b.WriteString("]}")
	return nil
}

func writeRing(b *strings.Builder, ring []geom.XY, c *config) {
	b.WriteByte('[')
	for i, v := range ring {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('[')
		writeNumber(b, v.X, c)
		b.WriteByte(',')
		writeNumber(b, v.Y, c)
		b.WriteByte(']')
	}
	b.WriteByte(']')
}

func writeMultiPoint(b *strings.Builder, mp *geom.MultiPoint, c *config) error {
	b.WriteString(`{"type":"MultiPoint","coordinates":`)
	writeCoordSequence(b, mp.FlatCoords(), mp.Layout(), c)
	b.WriteByte('}')
	return nil
}

func writeMultiLineString(b *strings.Builder, m *geom.MultiLineString, c *config) error {
	b.WriteString(`{"type":"MultiLineString","coordinates":[`)
	for i := 0; i < m.NumGeometries(); i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		ls := m.LineStringAt(i)
		writeCoordSequence(b, ls.FlatCoords(), ls.Layout(), c)
	}
	b.WriteString("]}")
	return nil
}

func writeMultiPolygon(b *strings.Builder, m *geom.MultiPolygon, c *config) error {
	b.WriteString(`{"type":"MultiPolygon","coordinates":[`)
	for i := 0; i < m.NumGeometries(); i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		p := m.PolygonAt(i)
		b.WriteByte('[')
		for r := 0; r < p.NumRings(); r++ {
			if r > 0 {
				b.WriteByte(',')
			}
			ring := p.Ring(r)
			if c != nil && c.forceCCW {
				ring = orientedRing(ring, r)
			}
			writeRing(b, ring, c)
		}
		b.WriteByte(']')
	}
	b.WriteString("]}")
	return nil
}

func writeGeometryCollection(b *strings.Builder, gc *geom.GeometryCollection, c *config) error {
	b.WriteString(`{"type":"GeometryCollection","geometries":[`)
	for i := 0; i < gc.NumGeometries(); i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		if err := writeGeometry(b, gc.GeometryAt(i), c); err != nil {
			return err
		}
	}
	b.WriteString("]}")
	return nil
}

// rawJSONOrNull emits raw JSON or "null" for nil.
func rawJSONOrNull(v json.RawMessage) string {
	if len(v) == 0 {
		return "null"
	}
	return string(v)
}
