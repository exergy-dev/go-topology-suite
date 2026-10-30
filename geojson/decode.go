package geojson

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/exergy-dev/go-topology-suite/crs"
	"github.com/exergy-dev/go-topology-suite/geom"
)

// Unmarshal parses a GeoJSON geometry object. The returned geometry has
// CRS = nil unless the caller explicitly attaches one (RFC 7946 implies
// WGS84 but the package does not silently set it — that would conceal
// CRS-mismatch bugs in mixed pipelines).
func Unmarshal(data []byte) (geom.Geometry, error) {
	return UnmarshalWithCRS(data, nil)
}

// UnmarshalWithCRS is like Unmarshal but attaches the supplied CRS to the
// resulting geometry. Pass crs.WGS84 to honour the RFC 7946 default
// explicitly.
func UnmarshalWithCRS(data []byte, c *crs.CRS) (geom.Geometry, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("geojson: %w", err)
	}
	return decodeGeometry(raw, c)
}

func decodeGeometry(raw map[string]json.RawMessage, c *crs.CRS) (geom.Geometry, error) {
	tRaw, ok := raw["type"]
	if !ok {
		return nil, fmt.Errorf("geojson: missing 'type' field")
	}
	var typ string
	if err := json.Unmarshal(tRaw, &typ); err != nil {
		return nil, fmt.Errorf("geojson: bad 'type': %w", err)
	}
	if typ == "GeometryCollection" {
		return decodeGeometryCollection(raw, c)
	}
	coords, ok := raw["coordinates"]
	if !ok {
		return nil, fmt.Errorf("geojson: missing 'coordinates' for %q", typ)
	}
	switch typ {
	case "Point":
		return decodePoint(coords, c)
	case "LineString":
		return decodeLineString(coords, c)
	case "Polygon":
		return decodePolygon(coords, c)
	case "MultiPoint":
		return decodeMultiPoint(coords, c)
	case "MultiLineString":
		return decodeMultiLineString(coords, c)
	case "MultiPolygon":
		return decodeMultiPolygon(coords, c)
	default:
		return nil, fmt.Errorf("geojson: unknown type %q", typ)
	}
}

// decodeVertex reads a [x, y] or [x, y, z] coordinate. Layout is inferred
// from the array length on the first vertex.
// scanner is a minimal cursor over an already syntax-validated JSON
// fragment (the coordinates member). It decodes nested number arrays without
// going through encoding/json's reflection-based path.
type scanner struct {
	b []byte
	i int
}

func (s *scanner) ws() {
	for s.i < len(s.b) {
		switch s.b[s.i] {
		case ' ', '\t', '\n', '\r':
			s.i++
		default:
			return
		}
	}
}

func (s *scanner) isNullLit() bool {
	if len(s.b)-s.i >= 4 && string(s.b[s.i:s.i+4]) == "null" {
		s.i += 4
		return true
	}
	return false
}

// open consumes '[' (or the literal null, reported as isNull).
func (s *scanner) open() (isNull bool, err error) {
	s.ws()
	if s.i >= len(s.b) {
		return false, fmt.Errorf("geojson: unexpected end of coordinates")
	}
	switch s.b[s.i] {
	case '[':
		s.i++
		return false, nil
	case 'n':
		if s.isNullLit() {
			return true, nil
		}
	}
	return false, fmt.Errorf("geojson: cannot decode coordinates: expected array at offset %d", s.i)
}

// next advances past a separator and reports whether another element follows
// (false once the closing ']' has been consumed).
func (s *scanner) next(first bool) (bool, error) {
	s.ws()
	if s.i >= len(s.b) {
		return false, fmt.Errorf("geojson: unexpected end of coordinates")
	}
	c := s.b[s.i]
	if c == ']' {
		s.i++
		return false, nil
	}
	if !first {
		if c != ',' {
			return false, fmt.Errorf("geojson: malformed coordinates at offset %d", s.i)
		}
		s.i++
		s.ws()
	}
	return true, nil
}

func (s *scanner) number() (float64, error) {
	s.ws()
	st := s.i
	if s.isNullLit() {
		return 0, nil // matches encoding/json: null leaves the zero value
	}
	for s.i < len(s.b) {
		c := s.b[s.i]
		if (c >= '0' && c <= '9') || c == '-' || c == '+' || c == '.' || c == 'e' || c == 'E' {
			s.i++
			continue
		}
		break
	}
	if s.i == st {
		return 0, fmt.Errorf("geojson: cannot decode coordinates: expected number at offset %d", st)
	}
	f, err := strconv.ParseFloat(string(s.b[st:s.i]), 64)
	if err != nil {
		return 0, fmt.Errorf("geojson: bad number %q: %w", s.b[st:s.i], err)
	}
	return f, nil
}

// vertex reads one position. It stores the first (up to) three components in
// buf and returns the total component count.
func (s *scanner) vertex(buf *[3]float64) (int, error) {
	isNull, err := s.open()
	if err != nil {
		return 0, err
	}
	if isNull {
		return 0, fmt.Errorf("geojson: vertex needs at least 2 components, got 0")
	}
	n := 0
	for first := true; ; first = false {
		more, err := s.next(first)
		if err != nil {
			return 0, err
		}
		if !more {
			break
		}
		f, err := s.number()
		if err != nil {
			return 0, err
		}
		if n < 3 {
			buf[n] = f
		}
		n++
	}
	if n < 2 {
		return 0, fmt.Errorf("geojson: vertex needs at least 2 components, got %d", n)
	}
	return n, nil
}

func (s *scanner) end() error {
	s.ws()
	if s.i != len(s.b) {
		return fmt.Errorf("geojson: trailing data in coordinates at offset %d", s.i)
	}
	return nil
}

// readLine reads an array of positions, appending to flat. polyStride is the
// stride the result must be normalised to (0 = use this line's own layout).
// It returns the grown buffer, the line's own layout and its vertex count.
func (s *scanner) readLine(flat []float64, polyStride int) ([]float64, geom.Layout, int, error) {
	isNull, err := s.open()
	if err != nil {
		return flat, geom.NoLayout, 0, err
	}
	layout := geom.LayoutXY
	if isNull {
		return flat, layout, 0, nil
	}
	own, stride := 0, polyStride
	nv := 0
	var buf [3]float64
	for first := true; ; first = false {
		more, err := s.next(first)
		if err != nil {
			return flat, geom.NoLayout, 0, err
		}
		if !more {
			break
		}
		n, err := s.vertex(&buf)
		if err != nil {
			if nv > 0 {
				err = fmt.Errorf("vertex %d: %w", nv, err)
			}
			return flat, geom.NoLayout, 0, err
		}
		if nv == 0 {
			own = 2
			if n >= 3 {
				own = 3
				layout = geom.LayoutXYZ
			}
			if stride == 0 {
				stride = own
			}
		}
		for j := 0; j < stride; j++ {
			if j < own && j < n {
				flat = append(flat, buf[j])
			} else {
				flat = append(flat, 0)
			}
		}
		nv++
	}
	return flat, layout, nv, nil
}

func decodePoint(coords json.RawMessage, c *crs.CRS) (geom.Geometry, error) {
	// Treat empty array as POINT EMPTY for cross-format compatibility.
	if string(coords) == "[]" {
		return geom.NewEmptyPoint(c, geom.LayoutXY), nil
	}
	s := scanner{b: coords}
	var v [3]float64
	n, err := s.vertex(&v)
	if err != nil {
		return nil, err
	}
	if err := s.end(); err != nil {
		return nil, err
	}
	if n >= 3 {
		return geom.NewPointXYZ(c, geom.XYZ{X: v[0], Y: v[1], Z: v[2]}), nil
	}
	return geom.NewPoint(c, geom.XY{X: v[0], Y: v[1]}), nil
}

// capHint estimates the flat buffer capacity for the array whose contents
// start at b[0]: one bracket per vertex (plus ring brackets), assuming XY.
func capHint(b []byte) int {
	// Limit the count to the array being decoded (depth 0 reached).
	depth, n := 0, 0
	for _, c := range b {
		if c == '[' {
			depth++
			n++
		} else if c == ']' {
			if depth == 0 {
				break
			}
			depth--
		}
	}
	return n * 2
}

func decodeFlatLine(s *scanner) ([]float64, geom.Layout, error) {
	flat := make([]float64, 0, capHint(s.b[min(s.i+1, len(s.b)):]))
	flat, layout, nv, err := s.readLine(flat, 0)
	if err != nil {
		return nil, geom.NoLayout, err
	}
	if nv == 0 {
		return nil, geom.LayoutXY, nil
	}
	return flat, layout, nil
}

func decodeLineString(coords json.RawMessage, c *crs.CRS) (geom.Geometry, error) {
	s := scanner{b: coords}
	flat, layout, err := decodeFlatLine(&s)
	if err != nil {
		return nil, err
	}
	if err := s.end(); err != nil {
		return nil, err
	}
	return geom.NewLineStringOwned(layout, c, flat), nil
}

// readPolygon reads a ring array. Rings are normalised to the first ring's
// layout (extra dims dropped, missing dims zero-padded).
func (s *scanner) readPolygon(c *crs.CRS) (*geom.Polygon, error) {
	isNull, err := s.open()
	if err != nil {
		return nil, err
	}
	if isNull {
		return geom.NewEmptyPolygon(c, geom.LayoutXY), nil
	}
	var flat []float64
	var ringStarts []int
	layout := geom.NoLayout
	vertexOff := 0
	for first := true; ; first = false {
		more, err := s.next(first)
		if err != nil {
			return nil, err
		}
		if !more {
			break
		}
		if first {
			flat = make([]float64, 0, capHint(s.b[s.i:]))
		}
		stride := 0
		if layout != geom.NoLayout {
			stride = layout.Stride()
		}
		var ringLayout geom.Layout
		var nv int
		flat, ringLayout, nv, err = s.readLine(flat, stride)
		if err != nil {
			return nil, err
		}
		if layout == geom.NoLayout {
			layout = ringLayout
		}
		ringStarts = append(ringStarts, vertexOff)
		vertexOff += nv
	}
	if len(ringStarts) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY), nil
	}
	if len(flat) == 0 {
		flat = nil // match the historical nil (not empty) buffer
	}
	return geom.NewPolygonOwned(layout, c, flat, ringStarts), nil
}

func decodePolygon(coords json.RawMessage, c *crs.CRS) (geom.Geometry, error) {
	s := scanner{b: coords}
	p, err := s.readPolygon(c)
	if err != nil {
		return nil, err
	}
	if err := s.end(); err != nil {
		return nil, err
	}
	return p, nil
}

func decodeMultiPoint(coords json.RawMessage, c *crs.CRS) (geom.Geometry, error) {
	s := scanner{b: coords}
	flat, layout, err := decodeFlatLine(&s)
	if err != nil {
		return nil, err
	}
	if err := s.end(); err != nil {
		return nil, err
	}
	return geom.NewMultiPointOwned(layout, c, flat), nil
}

func decodeMultiLineString(coords json.RawMessage, c *crs.CRS) (geom.Geometry, error) {
	s := scanner{b: coords}
	isNull, err := s.open()
	if err != nil {
		return nil, err
	}
	parts := make([]*geom.LineString, 0)
	if !isNull {
		for first := true; ; first = false {
			more, err := s.next(first)
			if err != nil {
				return nil, err
			}
			if !more {
				break
			}
			flat, layout, err := decodeFlatLine(&s)
			if err != nil {
				return nil, err
			}
			parts = append(parts, geom.NewLineStringOwned(layout, c, flat))
		}
	}
	if err := s.end(); err != nil {
		return nil, err
	}
	return geom.NewMultiLineString(c, parts...), nil
}

func decodeMultiPolygon(coords json.RawMessage, c *crs.CRS) (geom.Geometry, error) {
	s := scanner{b: coords}
	isNull, err := s.open()
	if err != nil {
		return nil, err
	}
	polys := make([]*geom.Polygon, 0)
	if !isNull {
		for first := true; ; first = false {
			more, err := s.next(first)
			if err != nil {
				return nil, err
			}
			if !more {
				break
			}
			p, err := s.readPolygon(c)
			if err != nil {
				return nil, err
			}
			polys = append(polys, p)
		}
	}
	if err := s.end(); err != nil {
		return nil, err
	}
	return geom.NewMultiPolygon(c, polys...), nil
}

func decodeGeometryCollection(raw map[string]json.RawMessage, c *crs.CRS) (geom.Geometry, error) {
	geomsRaw, ok := raw["geometries"]
	if !ok {
		return nil, fmt.Errorf("geojson: GeometryCollection missing 'geometries'")
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(geomsRaw, &arr); err != nil {
		return nil, err
	}
	parts := make([]geom.Geometry, 0, len(arr))
	for _, m := range arr {
		var inner map[string]json.RawMessage
		if err := json.Unmarshal(m, &inner); err != nil {
			return nil, err
		}
		g, err := decodeGeometry(inner, c)
		if err != nil {
			return nil, err
		}
		parts = append(parts, g)
	}
	return geom.NewGeometryCollection(c, parts...), nil
}
