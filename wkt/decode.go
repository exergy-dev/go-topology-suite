package wkt

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
)

// Unmarshal parses a WKT (or EWKT with SRID prefix) string and returns the
// constructed geometry. The CRS is set to the SRID-prefixed CRS if present;
// otherwise nil.
func Unmarshal(s string) (geom.Geometry, error) {
	p := &parser{src: s}
	if err := p.parseSRIDPrefix(); err != nil {
		return nil, err
	}
	g, err := p.parseGeometry()
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()
	if p.pos < len(p.src) {
		return nil, fmt.Errorf("wkt: trailing input at offset %d: %q", p.pos, p.src[p.pos:])
	}
	return g, nil
}

type parser struct {
	src string
	pos int
	crs *crs.CRS
}

func (p *parser) skipWhitespace() {
	for p.pos < len(p.src) && unicode.IsSpace(rune(p.src[p.pos])) {
		p.pos++
	}
}

// peek returns the next byte without consuming, or 0 at EOF.
func (p *parser) peek() byte {
	p.skipWhitespace()
	if p.pos >= len(p.src) {
		return 0
	}
	return p.src[p.pos]
}

func (p *parser) consume(b byte) error {
	p.skipWhitespace()
	if p.pos >= len(p.src) || p.src[p.pos] != b {
		return fmt.Errorf("wkt: expected %q at offset %d", b, p.pos)
	}
	p.pos++
	return nil
}

// readWord consumes an alphabetic word (case-folded to upper) and returns it.
func (p *parser) readWord() string {
	p.skipWhitespace()
	start := p.pos
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			p.pos++
		} else {
			break
		}
	}
	return strings.ToUpper(p.src[start:p.pos])
}

func (p *parser) readNumber() (float64, error) {
	p.skipWhitespace()
	start := p.pos
	if p.pos < len(p.src) && (p.src[p.pos] == '-' || p.src[p.pos] == '+') {
		p.pos++
	}
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if (c >= '0' && c <= '9') || c == '.' || c == 'e' || c == 'E' || c == '-' || c == '+' {
			p.pos++
		} else {
			break
		}
	}
	if start == p.pos {
		return 0, fmt.Errorf("wkt: expected number at offset %d", p.pos)
	}
	return strconv.ParseFloat(p.src[start:p.pos], 64)
}

// parseSRIDPrefix consumes "SRID=<int>;" if present and stores the
// resulting CRS on the parser.
func (p *parser) parseSRIDPrefix() error {
	p.skipWhitespace()
	if !strings.HasPrefix(strings.ToUpper(p.src[p.pos:]), "SRID=") {
		return nil
	}
	p.pos += 5
	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] != ';' {
		p.pos++
	}
	code, err := strconv.Atoi(strings.TrimSpace(p.src[start:p.pos]))
	if err != nil {
		return fmt.Errorf("wkt: invalid SRID: %w", err)
	}
	if p.pos >= len(p.src) {
		return errors.New("wkt: SRID prefix missing terminator ';'")
	}
	p.pos++ // consume ';'
	p.crs = &crs.CRS{Authority: "EPSG", Code: code}
	return nil
}

// parseLayout consumes an optional layout suffix ("Z"/"M"/"ZM") after a
// type keyword. EMPTY tokens are NOT layouts; the caller handles them.
func (p *parser) parseLayout() geom.Layout {
	save := p.pos
	w := p.readWord()
	switch w {
	case "Z":
		return geom.LayoutXYZ
	case "M":
		return geom.LayoutXYM
	case "ZM":
		return geom.LayoutXYZM
	default:
		p.pos = save
		return geom.LayoutXY
	}
}

// parseGeometry dispatches on the leading type word.
func (p *parser) parseGeometry() (geom.Geometry, error) {
	w := p.readWord()
	switch w {
	case "POINT":
		return p.parsePoint()
	case "LINESTRING":
		return p.parseLineString()
	case "POLYGON":
		return p.parsePolygon()
	case "MULTIPOINT":
		return p.parseMultiPoint()
	case "MULTILINESTRING":
		return p.parseMultiLineString()
	case "MULTIPOLYGON":
		return p.parseMultiPolygon()
	case "GEOMETRYCOLLECTION":
		return p.parseGeometryCollection()
	default:
		return nil, fmt.Errorf("wkt: unknown geometry type %q at offset %d", w, p.pos)
	}
}

// parseEmptyOrLayout looks ahead. If the next word is EMPTY (after an
// optional layout suffix) it returns layout, true (empty). Otherwise it
// returns layout, false and leaves the parser positioned just before '('.
func (p *parser) parseEmptyOrLayout() (geom.Layout, bool) {
	layout := p.parseLayout()
	save := p.pos
	w := p.readWord()
	if w == "EMPTY" {
		return layout, true
	}
	p.pos = save
	return layout, false
}

func (p *parser) parsePoint() (geom.Geometry, error) {
	layout, empty := p.parseEmptyOrLayout()
	if empty {
		return geom.NewEmptyPoint(p.crs, layout), nil
	}
	if err := p.consume('('); err != nil {
		return nil, err
	}
	coord, err := p.readCoord(layout.Stride())
	if err != nil {
		return nil, err
	}
	if err := p.consume(')'); err != nil {
		return nil, err
	}
	switch layout {
	case geom.LayoutXY:
		return geom.NewPoint(p.crs, geom.XY{X: coord[0], Y: coord[1]}), nil
	case geom.LayoutXYZ:
		return geom.NewPointXYZ(p.crs, geom.XYZ{X: coord[0], Y: coord[1], Z: coord[2]}), nil
	case geom.LayoutXYM:
		return geom.NewPointXYM(p.crs, geom.XYM{X: coord[0], Y: coord[1], M: coord[2]}), nil
	case geom.LayoutXYZM:
		return geom.NewPointXYZM(p.crs, geom.XYZM{X: coord[0], Y: coord[1], Z: coord[2], M: coord[3]}), nil
	default:
		return geom.NewPoint(p.crs, geom.XY{X: coord[0], Y: coord[1]}), nil
	}
}

func (p *parser) readCoord(stride int) ([]float64, error) {
	out := make([]float64, stride)
	for i := 0; i < stride; i++ {
		v, err := p.readNumber()
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}

func (p *parser) readCoordSequence(stride int) ([]float64, error) {
	if err := p.consume('('); err != nil {
		return nil, err
	}
	var out []float64
	for {
		c, err := p.readCoord(stride)
		if err != nil {
			return nil, err
		}
		out = append(out, c...)
		p.skipWhitespace()
		if p.pos >= len(p.src) {
			return nil, errors.New("wkt: unexpected EOF in coord sequence")
		}
		switch p.src[p.pos] {
		case ',':
			p.pos++
		case ')':
			p.pos++
			return out, nil
		default:
			return nil, fmt.Errorf("wkt: expected ',' or ')' at offset %d", p.pos)
		}
	}
}

func (p *parser) parseLineString() (geom.Geometry, error) {
	layout, empty := p.parseEmptyOrLayout()
	if empty {
		return geom.NewLineStringFlat(layout, p.crs, nil), nil
	}
	flat, err := p.readCoordSequence(layout.Stride())
	if err != nil {
		return nil, err
	}
	return geom.NewLineStringFlatNoClone(layout, p.crs, flat), nil
}

func (p *parser) parsePolygon() (geom.Geometry, error) {
	layout, empty := p.parseEmptyOrLayout()
	if empty {
		return geom.NewEmptyPolygon(p.crs, layout), nil
	}
	if err := p.consume('('); err != nil {
		return nil, err
	}
	stride := layout.Stride()
	var rings [][]geom.XY
	for {
		flat, err := p.readCoordSequence(stride)
		if err != nil {
			return nil, err
		}
		rings = append(rings, flatToXY(flat, stride))
		p.skipWhitespace()
		if p.pos >= len(p.src) {
			return nil, errors.New("wkt: unexpected EOF in polygon")
		}
		switch p.src[p.pos] {
		case ',':
			p.pos++
		case ')':
			p.pos++
			return geom.NewPolygon(p.crs, rings...), nil
		default:
			return nil, fmt.Errorf("wkt: expected ',' or ')' at offset %d", p.pos)
		}
	}
}

func flatToXY(flat []float64, stride int) []geom.XY {
	n := len(flat) / stride
	out := make([]geom.XY, n)
	for i := 0; i < n; i++ {
		out[i] = geom.XY{X: flat[i*stride], Y: flat[i*stride+1]}
	}
	return out
}

func (p *parser) parseMultiPoint() (geom.Geometry, error) {
	layout, empty := p.parseEmptyOrLayout()
	if empty {
		return geom.NewMultiPoint(p.crs, nil), nil
	}
	if err := p.consume('('); err != nil {
		return nil, err
	}
	stride := layout.Stride()
	var pts []geom.XY
	for {
		// Each member may be either parenthesised "(x y)" or bare "x y".
		p.skipWhitespace()
		var flat []float64
		if p.peek() == '(' {
			p.pos++
			c, err := p.readCoord(stride)
			if err != nil {
				return nil, err
			}
			if err := p.consume(')'); err != nil {
				return nil, err
			}
			flat = c
		} else {
			c, err := p.readCoord(stride)
			if err != nil {
				return nil, err
			}
			flat = c
		}
		pts = append(pts, geom.XY{X: flat[0], Y: flat[1]})
		p.skipWhitespace()
		if p.pos >= len(p.src) {
			return nil, errors.New("wkt: unexpected EOF in multipoint")
		}
		switch p.src[p.pos] {
		case ',':
			p.pos++
		case ')':
			p.pos++
			return geom.NewMultiPoint(p.crs, pts), nil
		default:
			return nil, fmt.Errorf("wkt: expected ',' or ')' at offset %d", p.pos)
		}
	}
}

func (p *parser) parseMultiLineString() (geom.Geometry, error) {
	layout, empty := p.parseEmptyOrLayout()
	if empty {
		return geom.NewMultiLineString(p.crs), nil
	}
	if err := p.consume('('); err != nil {
		return nil, err
	}
	stride := layout.Stride()
	var lines []*geom.LineString
	for {
		flat, err := p.readCoordSequence(stride)
		if err != nil {
			return nil, err
		}
		lines = append(lines, geom.NewLineStringFlatNoClone(layout, p.crs, flat))
		p.skipWhitespace()
		if p.pos >= len(p.src) {
			return nil, errors.New("wkt: unexpected EOF in multilinestring")
		}
		switch p.src[p.pos] {
		case ',':
			p.pos++
		case ')':
			p.pos++
			return geom.NewMultiLineString(p.crs, lines...), nil
		default:
			return nil, fmt.Errorf("wkt: expected ',' or ')' at offset %d", p.pos)
		}
	}
}

func (p *parser) parseMultiPolygon() (geom.Geometry, error) {
	layout, empty := p.parseEmptyOrLayout()
	if empty {
		return geom.NewMultiPolygon(p.crs), nil
	}
	if err := p.consume('('); err != nil {
		return nil, err
	}
	stride := layout.Stride()
	var polys []*geom.Polygon
	for {
		if err := p.consume('('); err != nil {
			return nil, err
		}
		var rings [][]geom.XY
		for {
			flat, err := p.readCoordSequence(stride)
			if err != nil {
				return nil, err
			}
			rings = append(rings, flatToXY(flat, stride))
			p.skipWhitespace()
			if p.pos >= len(p.src) {
				return nil, errors.New("wkt: unexpected EOF in multipolygon")
			}
			if p.src[p.pos] == ',' {
				p.pos++
				continue
			}
			if p.src[p.pos] == ')' {
				p.pos++
				break
			}
			return nil, fmt.Errorf("wkt: expected ',' or ')' at offset %d", p.pos)
		}
		polys = append(polys, geom.NewPolygon(p.crs, rings...))
		p.skipWhitespace()
		if p.pos >= len(p.src) {
			return nil, errors.New("wkt: unexpected EOF in multipolygon")
		}
		if p.src[p.pos] == ',' {
			p.pos++
			continue
		}
		if p.src[p.pos] == ')' {
			p.pos++
			return geom.NewMultiPolygon(p.crs, polys...), nil
		}
		return nil, fmt.Errorf("wkt: expected ',' or ')' at offset %d", p.pos)
	}
}

func (p *parser) parseGeometryCollection() (geom.Geometry, error) {
	_, empty := p.parseEmptyOrLayout()
	if empty {
		return geom.NewGeometryCollection(p.crs), nil
	}
	if err := p.consume('('); err != nil {
		return nil, err
	}
	var members []geom.Geometry
	for {
		g, err := p.parseGeometry()
		if err != nil {
			return nil, err
		}
		members = append(members, g)
		p.skipWhitespace()
		if p.pos >= len(p.src) {
			return nil, errors.New("wkt: unexpected EOF in collection")
		}
		switch p.src[p.pos] {
		case ',':
			p.pos++
		case ')':
			p.pos++
			return geom.NewGeometryCollection(p.crs, members...), nil
		default:
			return nil, fmt.Errorf("wkt: expected ',' or ')' at offset %d", p.pos)
		}
	}
}
