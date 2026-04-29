package geojson

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/terra-geo/terra/geom"
)

// Marshal returns the GeoJSON encoding of g. CRS is implicit WGS84 per
// RFC 7946; any non-nil CRS attached to g is silently ignored on output.
func Marshal(g geom.Geometry) ([]byte, error) {
	var b strings.Builder
	if err := writeGeometry(&b, g); err != nil {
		return nil, err
	}
	return []byte(b.String()), nil
}

func writeGeometry(b *strings.Builder, g geom.Geometry) error {
	switch v := g.(type) {
	case *geom.Point:
		return writePoint(b, v)
	case *geom.LineString:
		return writeLineString(b, v)
	case *geom.Polygon:
		return writePolygon(b, v)
	case *geom.MultiPoint:
		return writeMultiPoint(b, v)
	case *geom.MultiLineString:
		return writeMultiLineString(b, v)
	case *geom.MultiPolygon:
		return writeMultiPolygon(b, v)
	case *geom.GeometryCollection:
		return writeGeometryCollection(b, v)
	default:
		return fmt.Errorf("geojson: unsupported %T", g)
	}
}

func writeNumber(b *strings.Builder, f float64) {
	b.WriteString(strconv.FormatFloat(f, 'g', -1, 64))
}

// writeVertex writes [x, y] or [x, y, z] depending on layout. M is
// dropped per RFC 7946 §3.1.1 (M not part of GeoJSON).
func writeVertex(b *strings.Builder, flat []float64, off int, layout geom.Layout) {
	stride := layout.Stride()
	b.WriteByte('[')
	writeNumber(b, flat[off])
	b.WriteByte(',')
	writeNumber(b, flat[off+1])
	if layout.HasZ() {
		// Z is at index 2 for XYZ and XYZM; XYM has no Z.
		zIdx := 2
		b.WriteByte(',')
		writeNumber(b, flat[off+zIdx])
	}
	_ = stride
	b.WriteByte(']')
}

func writePoint(b *strings.Builder, p *geom.Point) error {
	b.WriteString(`{"type":"Point","coordinates":`)
	if p.IsEmpty() {
		// RFC 7946 §3.1: GeoJSON has no first-class empty geometry; emit
		// an empty array. (A foreign "EMPTY" sentinel is sometimes used;
		// we deliberately do not emit it.)
		b.WriteString("[]}")
		return nil
	}
	writeVertex(b, p.FlatCoords(), 0, p.Layout())
	b.WriteByte('}')
	return nil
}

func writeCoordSequence(b *strings.Builder, flat []float64, layout geom.Layout) {
	stride := layout.Stride()
	b.WriteByte('[')
	n := len(flat) / stride
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		writeVertex(b, flat, i*stride, layout)
	}
	b.WriteByte(']')
}

func writeLineString(b *strings.Builder, ls *geom.LineString) error {
	b.WriteString(`{"type":"LineString","coordinates":`)
	writeCoordSequence(b, ls.FlatCoords(), ls.Layout())
	b.WriteByte('}')
	return nil
}

func writePolygon(b *strings.Builder, p *geom.Polygon) error {
	b.WriteString(`{"type":"Polygon","coordinates":[`)
	for r := 0; r < p.NumRings(); r++ {
		if r > 0 {
			b.WriteByte(',')
		}
		writeRing(b, p.Ring(r))
	}
	b.WriteString("]}")
	return nil
}

func writeRing(b *strings.Builder, ring []geom.XY) {
	b.WriteByte('[')
	for i, v := range ring {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('[')
		writeNumber(b, v.X)
		b.WriteByte(',')
		writeNumber(b, v.Y)
		b.WriteByte(']')
	}
	b.WriteByte(']')
}

func writeMultiPoint(b *strings.Builder, mp *geom.MultiPoint) error {
	b.WriteString(`{"type":"MultiPoint","coordinates":`)
	writeCoordSequence(b, mp.FlatCoords(), mp.Layout())
	b.WriteByte('}')
	return nil
}

func writeMultiLineString(b *strings.Builder, m *geom.MultiLineString) error {
	b.WriteString(`{"type":"MultiLineString","coordinates":[`)
	for i := 0; i < m.NumGeometries(); i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		ls := m.LineStringAt(i)
		writeCoordSequence(b, ls.FlatCoords(), ls.Layout())
	}
	b.WriteString("]}")
	return nil
}

func writeMultiPolygon(b *strings.Builder, m *geom.MultiPolygon) error {
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
			writeRing(b, p.Ring(r))
		}
		b.WriteByte(']')
	}
	b.WriteString("]}")
	return nil
}

func writeGeometryCollection(b *strings.Builder, gc *geom.GeometryCollection) error {
	b.WriteString(`{"type":"GeometryCollection","geometries":[`)
	for i := 0; i < gc.NumGeometries(); i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		if err := writeGeometry(b, gc.GeometryAt(i)); err != nil {
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
