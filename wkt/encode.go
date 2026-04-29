package wkt

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/terra-geo/terra/geom"
)

// Marshal returns the WKT representation of g.
// CRS is intentionally not encoded — WKT proper has no SRID slot. Use
// MarshalEWKT for the PostGIS "SRID=...;..." extension.
func Marshal(g geom.Geometry) (string, error) {
	var b strings.Builder
	if err := appendGeometry(&b, g); err != nil {
		return "", err
	}
	return b.String(), nil
}

// MarshalEWKT returns "SRID=<code>;<wkt>" if the geometry has an EPSG-coded
// CRS attached; otherwise it is identical to Marshal.
func MarshalEWKT(g geom.Geometry) (string, error) {
	core, err := Marshal(g)
	if err != nil {
		return "", err
	}
	if c := g.CRS(); c != nil && c.Authority == "EPSG" && c.Code != 0 {
		return "SRID=" + strconv.Itoa(c.Code) + ";" + core, nil
	}
	return core, nil
}

func appendGeometry(b *strings.Builder, g geom.Geometry) error {
	switch v := g.(type) {
	case *geom.Point:
		return appendPoint(b, v)
	case *geom.LineString:
		return appendLineString(b, v)
	case *geom.Polygon:
		return appendPolygon(b, v)
	case *geom.MultiPoint:
		return appendMultiPoint(b, v)
	case *geom.MultiLineString:
		return appendMultiLineString(b, v)
	case *geom.MultiPolygon:
		return appendMultiPolygon(b, v)
	case *geom.GeometryCollection:
		return appendGeometryCollection(b, v)
	default:
		return fmt.Errorf("wkt: unsupported geometry type %T", g)
	}
}

// layoutSuffix returns the keyword that follows the type name for non-XY
// layouts ("", " Z", " M", " ZM").
func layoutSuffix(l geom.Layout) string {
	switch l {
	case geom.LayoutXYZ:
		return " Z"
	case geom.LayoutXYM:
		return " M"
	case geom.LayoutXYZM:
		return " ZM"
	default:
		return ""
	}
}

func writeNumber(b *strings.Builder, f float64) {
	b.WriteString(strconv.FormatFloat(f, 'g', -1, 64))
}

// writeFlatPoint writes one stride-sized vertex starting at coords[off].
func writeFlatPoint(b *strings.Builder, coords []float64, off, stride int) {
	for i := 0; i < stride; i++ {
		if i > 0 {
			b.WriteByte(' ')
		}
		writeNumber(b, coords[off+i])
	}
}

func appendPoint(b *strings.Builder, p *geom.Point) error {
	b.WriteString("POINT")
	b.WriteString(layoutSuffix(p.Layout()))
	if p.IsEmpty() {
		b.WriteString(" EMPTY")
		return nil
	}
	b.WriteString(" (")
	flat := p.FlatCoords()
	stride := p.Layout().Stride()
	writeFlatPoint(b, flat, 0, stride)
	b.WriteByte(')')
	return nil
}

func appendCoordSequence(b *strings.Builder, flat []float64, stride int) {
	b.WriteByte('(')
	n := len(flat) / stride
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		writeFlatPoint(b, flat, i*stride, stride)
	}
	b.WriteByte(')')
}

func appendLineString(b *strings.Builder, ls *geom.LineString) error {
	b.WriteString("LINESTRING")
	b.WriteString(layoutSuffix(ls.Layout()))
	if ls.IsEmpty() {
		b.WriteString(" EMPTY")
		return nil
	}
	b.WriteByte(' ')
	appendCoordSequence(b, ls.FlatCoords(), ls.Layout().Stride())
	return nil
}

func appendPolygon(b *strings.Builder, p *geom.Polygon) error {
	b.WriteString("POLYGON")
	b.WriteString(layoutSuffix(p.Layout()))
	if p.IsEmpty() {
		b.WriteString(" EMPTY")
		return nil
	}
	b.WriteString(" (")
	for i := 0; i < p.NumRings(); i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		ring := p.Ring(i)
		writeRingXY(b, ring)
	}
	b.WriteByte(')')
	return nil
}

func writeRingXY(b *strings.Builder, ring []geom.XY) {
	b.WriteByte('(')
	for i, p := range ring {
		if i > 0 {
			b.WriteString(", ")
		}
		writeNumber(b, p.X)
		b.WriteByte(' ')
		writeNumber(b, p.Y)
	}
	b.WriteByte(')')
}

func appendMultiPoint(b *strings.Builder, mp *geom.MultiPoint) error {
	b.WriteString("MULTIPOINT")
	b.WriteString(layoutSuffix(mp.Layout()))
	if mp.IsEmpty() {
		b.WriteString(" EMPTY")
		return nil
	}
	stride := mp.Layout().Stride()
	flat := mp.FlatCoords()
	n := len(flat) / stride
	b.WriteString(" (")
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteByte('(')
		writeFlatPoint(b, flat, i*stride, stride)
		b.WriteByte(')')
	}
	b.WriteByte(')')
	return nil
}

func appendMultiLineString(b *strings.Builder, m *geom.MultiLineString) error {
	b.WriteString("MULTILINESTRING")
	b.WriteString(layoutSuffix(m.Layout()))
	if m.IsEmpty() {
		b.WriteString(" EMPTY")
		return nil
	}
	b.WriteString(" (")
	for i := 0; i < m.NumGeometries(); i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		ls := m.LineStringAt(i)
		appendCoordSequence(b, ls.FlatCoords(), ls.Layout().Stride())
	}
	b.WriteByte(')')
	return nil
}

func appendMultiPolygon(b *strings.Builder, m *geom.MultiPolygon) error {
	b.WriteString("MULTIPOLYGON")
	b.WriteString(layoutSuffix(m.Layout()))
	if m.IsEmpty() {
		b.WriteString(" EMPTY")
		return nil
	}
	b.WriteString(" (")
	for i := 0; i < m.NumGeometries(); i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		p := m.PolygonAt(i)
		b.WriteByte('(')
		for j := 0; j < p.NumRings(); j++ {
			if j > 0 {
				b.WriteString(", ")
			}
			writeRingXY(b, p.Ring(j))
		}
		b.WriteByte(')')
	}
	b.WriteByte(')')
	return nil
}

func appendGeometryCollection(b *strings.Builder, gc *geom.GeometryCollection) error {
	b.WriteString("GEOMETRYCOLLECTION")
	b.WriteString(layoutSuffix(gc.Layout()))
	if gc.IsEmpty() {
		b.WriteString(" EMPTY")
		return nil
	}
	b.WriteString(" (")
	for i := 0; i < gc.NumGeometries(); i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		if err := appendGeometry(b, gc.GeometryAt(i)); err != nil {
			return err
		}
	}
	b.WriteByte(')')
	return nil
}
