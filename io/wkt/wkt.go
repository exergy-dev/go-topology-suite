// Package wkt provides Well-Known Text (WKT) encoding/decoding for geometries.
//
// Use Marshal/Unmarshal for []byte operations:
//
//	data, err := wkt.Marshal(point)
//	geom, err := wkt.Unmarshal(data)
//
// Use MarshalString/UnmarshalString for string operations:
//
//	str := wkt.MarshalString(point)
//	geom, err := wkt.UnmarshalString(str)
package wkt

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/io/ioutil"
)

// Options configures WKT marshaling behavior.
type Options struct {
	// Precision is the number of decimal places to output (-1 for default).
	Precision int
	// Formatted controls whether output includes newlines and indentation.
	Formatted bool
	// OutputDimension specifies dimensions to include (2, 3 for Z, 4 for ZM).
	OutputDimension int
}

// DefaultOptions returns the default marshaling options.
func DefaultOptions() Options {
	return Options{
		Precision:       -1,
		Formatted:       false,
		OutputDimension: 2,
	}
}

// Marshal marshals a geometry to WKT bytes.
func Marshal(g geom.Geometry) ([]byte, error) {
	return []byte(MarshalString(g)), nil
}

// MarshalString marshals a geometry to a WKT string.
func MarshalString(g geom.Geometry) string {
	return MarshalStringWithOptions(g, DefaultOptions())
}

// MarshalStringWithOptions marshals with custom options.
func MarshalStringWithOptions(g geom.Geometry, opts Options) string {
	if g == nil {
		return ""
	}
	var sb strings.Builder
	writeGeometry(&sb, g, 0, opts)
	return sb.String()
}

// MarshalIndent marshals a geometry with indentation.
func MarshalIndent(g geom.Geometry) ([]byte, error) {
	opts := DefaultOptions()
	opts.Formatted = true
	return []byte(MarshalStringWithOptions(g, opts)), nil
}

// Unmarshal unmarshals WKT bytes to a geometry.
func Unmarshal(data []byte) (geom.Geometry, error) {
	return UnmarshalString(string(data))
}

// UnmarshalString unmarshals a WKT string to a geometry.
func UnmarshalString(wkt string) (geom.Geometry, error) {
	return UnmarshalStringWithFactory(wkt, geom.DefaultFactory)
}

// UnmarshalStringWithFactory unmarshals using a custom geometry factory.
func UnmarshalStringWithFactory(wkt string, factory *geom.GeometryFactory) (geom.Geometry, error) {
	if factory == nil {
		factory = geom.DefaultFactory
	}
	p := &parser{
		input:   strings.TrimSpace(wkt),
		pos:     0,
		factory: factory,
	}
	g, err := p.parse()
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()
	if !p.atEnd() {
		return nil, fmt.Errorf("unexpected content after geometry at position %d", p.pos)
	}
	return g, nil
}

// UnmarshalWithFactory unmarshals WKT bytes using a custom geometry factory.
func UnmarshalWithFactory(data []byte, factory *geom.GeometryFactory) (geom.Geometry, error) {
	return UnmarshalStringWithFactory(string(data), factory)
}

// --- Marshaling implementation ---

func writeGeometry(sb *strings.Builder, g geom.Geometry, indent int, opts Options) {
	geom.VisitGeometry(g, geom.GeometryVisitor{
		Point: func(p *geom.Point) {
			writePoint(sb, p, opts)
		},
		LineString: func(ls *geom.LineString) {
			writeLineString(sb, ls, opts)
		},
		LinearRing: func(lr *geom.LinearRing) {
			writeLinearRing(sb, lr, opts)
		},
		Polygon: func(p *geom.Polygon) {
			writePolygon(sb, p, indent, opts)
		},
		MultiPoint: func(mp *geom.MultiPoint) {
			writeMultiPoint(sb, mp, indent, opts)
		},
		MultiLineString: func(mls *geom.MultiLineString) {
			writeMultiLineString(sb, mls, indent, opts)
		},
		MultiPolygon: func(mp *geom.MultiPolygon) {
			writeMultiPolygon(sb, mp, indent, opts)
		},
		GeometryCollection: func(gc *geom.GeometryCollection) {
			writeGeometryCollection(sb, gc, indent, opts)
		},
		Default: func(g geom.Geometry) {
			sb.WriteString(g.String())
		},
	})
}

func writePoint(sb *strings.Builder, p *geom.Point, opts Options) {
	sb.WriteString("POINT ")
	if p.IsEmpty() {
		sb.WriteString("EMPTY")
		return
	}
	writeDimensionMarker(sb, p.Coordinates(), opts)
	sb.WriteString("(")
	writeCoordinate(sb, p.Coordinate(), opts)
	sb.WriteString(")")
}

func writeLineString(sb *strings.Builder, ls *geom.LineString, opts Options) {
	sb.WriteString("LINESTRING ")
	if ls.IsEmpty() {
		sb.WriteString("EMPTY")
		return
	}
	writeDimensionMarker(sb, ls.Coordinates(), opts)
	writeCoordinateSequence(sb, ls.Coordinates(), opts)
}

func writeLinearRing(sb *strings.Builder, lr *geom.LinearRing, opts Options) {
	sb.WriteString("LINEARRING ")
	if lr.IsEmpty() {
		sb.WriteString("EMPTY")
		return
	}
	writeDimensionMarker(sb, lr.Coordinates(), opts)
	writeCoordinateSequence(sb, lr.Coordinates(), opts)
}

func writePolygon(sb *strings.Builder, p *geom.Polygon, indent int, opts Options) {
	sb.WriteString("POLYGON ")
	if p.IsEmpty() {
		sb.WriteString("EMPTY")
		return
	}

	coords := p.ExteriorRing().Coordinates()
	writeDimensionMarker(sb, coords, opts)

	sb.WriteString("(")
	if opts.Formatted {
		sb.WriteString("\n")
		writeIndent(sb, indent+1)
	}

	writeCoordinateSequence(sb, coords, opts)

	for i := 0; i < p.NumInteriorRings(); i++ {
		sb.WriteString(",")
		if opts.Formatted {
			sb.WriteString("\n")
			writeIndent(sb, indent+1)
		} else {
			sb.WriteString(" ")
		}
		writeCoordinateSequence(sb, p.InteriorRingN(i).Coordinates(), opts)
	}

	if opts.Formatted {
		sb.WriteString("\n")
		writeIndent(sb, indent)
	}
	sb.WriteString(")")
}

func writeMultiPoint(sb *strings.Builder, mp *geom.MultiPoint, indent int, opts Options) {
	sb.WriteString("MULTIPOINT ")
	if mp.IsEmpty() {
		sb.WriteString("EMPTY")
		return
	}

	writeDimensionMarker(sb, mp.Coordinates(), opts)
	sb.WriteString("(")

	for i := 0; i < mp.NumGeometries(); i++ {
		if i > 0 {
			sb.WriteString(",")
			if opts.Formatted {
				sb.WriteString("\n")
				writeIndent(sb, indent+1)
			} else {
				sb.WriteString(" ")
			}
		} else if opts.Formatted {
			sb.WriteString("\n")
			writeIndent(sb, indent+1)
		}

		p := mp.GeometryN(i).(*geom.Point)
		sb.WriteString("(")
		writeCoordinate(sb, p.Coordinate(), opts)
		sb.WriteString(")")
	}

	if opts.Formatted {
		sb.WriteString("\n")
		writeIndent(sb, indent)
	}
	sb.WriteString(")")
}

func writeMultiLineString(sb *strings.Builder, mls *geom.MultiLineString, indent int, opts Options) {
	sb.WriteString("MULTILINESTRING ")
	if mls.IsEmpty() {
		sb.WriteString("EMPTY")
		return
	}

	writeDimensionMarker(sb, mls.Coordinates(), opts)
	sb.WriteString("(")

	for i := 0; i < mls.NumGeometries(); i++ {
		if i > 0 {
			sb.WriteString(",")
			if opts.Formatted {
				sb.WriteString("\n")
				writeIndent(sb, indent+1)
			} else {
				sb.WriteString(" ")
			}
		} else if opts.Formatted {
			sb.WriteString("\n")
			writeIndent(sb, indent+1)
		}

		ls := mls.GeometryN(i).(*geom.LineString)
		writeCoordinateSequence(sb, ls.Coordinates(), opts)
	}

	if opts.Formatted {
		sb.WriteString("\n")
		writeIndent(sb, indent)
	}
	sb.WriteString(")")
}

func writeMultiPolygon(sb *strings.Builder, mp *geom.MultiPolygon, indent int, opts Options) {
	sb.WriteString("MULTIPOLYGON ")
	if mp.IsEmpty() {
		sb.WriteString("EMPTY")
		return
	}

	writeDimensionMarker(sb, mp.Coordinates(), opts)
	sb.WriteString("(")

	for i := 0; i < mp.NumGeometries(); i++ {
		if i > 0 {
			sb.WriteString(",")
			if opts.Formatted {
				sb.WriteString("\n")
				writeIndent(sb, indent+1)
			} else {
				sb.WriteString(" ")
			}
		} else if opts.Formatted {
			sb.WriteString("\n")
			writeIndent(sb, indent+1)
		}

		p := mp.GeometryN(i).(*geom.Polygon)
		writePolygonInner(sb, p, indent+1, opts)
	}

	if opts.Formatted {
		sb.WriteString("\n")
		writeIndent(sb, indent)
	}
	sb.WriteString(")")
}

func writePolygonInner(sb *strings.Builder, p *geom.Polygon, indent int, opts Options) {
	sb.WriteString("(")
	writeCoordinateSequence(sb, p.ExteriorRing().Coordinates(), opts)

	for i := 0; i < p.NumInteriorRings(); i++ {
		sb.WriteString(",")
		if opts.Formatted {
			sb.WriteString("\n")
			writeIndent(sb, indent+1)
		} else {
			sb.WriteString(" ")
		}
		writeCoordinateSequence(sb, p.InteriorRingN(i).Coordinates(), opts)
	}
	sb.WriteString(")")
}

func writeGeometryCollection(sb *strings.Builder, gc *geom.GeometryCollection, indent int, opts Options) {
	sb.WriteString("GEOMETRYCOLLECTION ")
	if gc.IsEmpty() {
		sb.WriteString("EMPTY")
		return
	}

	sb.WriteString("(")

	for i := 0; i < gc.NumGeometries(); i++ {
		if i > 0 {
			sb.WriteString(",")
			if opts.Formatted {
				sb.WriteString("\n")
				writeIndent(sb, indent+1)
			} else {
				sb.WriteString(" ")
			}
		} else if opts.Formatted {
			sb.WriteString("\n")
			writeIndent(sb, indent+1)
		}

		writeGeometry(sb, gc.GeometryN(i), indent+1, opts)
	}

	if opts.Formatted {
		sb.WriteString("\n")
		writeIndent(sb, indent)
	}
	sb.WriteString(")")
}

func writeCoordinateSequence(sb *strings.Builder, coords geom.CoordinateSequence, opts Options) {
	sb.WriteString("(")
	for i, c := range coords {
		if i > 0 {
			sb.WriteString(", ")
		}
		writeCoordinate(sb, c, opts)
	}
	sb.WriteString(")")
}

func writeCoordinate(sb *strings.Builder, c geom.Coordinate, opts Options) {
	ioutil.WriteNumber(sb, c.X, opts.Precision)
	sb.WriteString(" ")
	ioutil.WriteNumber(sb, c.Y, opts.Precision)

	if opts.OutputDimension >= 3 && c.HasZ() {
		sb.WriteString(" ")
		ioutil.WriteNumber(sb, c.Z, opts.Precision)
	}
	if opts.OutputDimension >= 4 && c.HasM() {
		sb.WriteString(" ")
		ioutil.WriteNumber(sb, c.M, opts.Precision)
	}
}

func writeDimensionMarker(sb *strings.Builder, coords geom.CoordinateSequence, opts Options) {
	if opts.OutputDimension <= 2 {
		return
	}

	hasZ := coords.HasZ()
	hasM := coords.HasM()

	if hasZ && hasM && opts.OutputDimension >= 4 {
		sb.WriteString("ZM ")
	} else if hasZ && opts.OutputDimension >= 3 {
		sb.WriteString("Z ")
	} else if hasM && opts.OutputDimension >= 4 {
		sb.WriteString("M ")
	}
}

func writeIndent(sb *strings.Builder, level int) {
	for i := 0; i < level; i++ {
		sb.WriteString("  ")
	}
}

// --- Parsing implementation ---

type parser struct {
	input       string
	pos         int
	factory     *geom.GeometryFactory
	defaultHasZ bool
	defaultHasM bool
	hasZ        bool
	hasM        bool
}

func (p *parser) parse() (geom.Geometry, error) {
	p.skipWhitespace()

	typeName := p.readWord()
	typeUpper := strings.ToUpper(typeName)

	p.hasZ = p.defaultHasZ
	p.hasM = p.defaultHasM

	p.skipWhitespace()
	if p.peek() != '(' && !p.atEnd() {
		modifier := strings.ToUpper(p.readWord())
		switch modifier {
		case "Z":
			p.hasZ = true
			p.hasM = false
		case "M":
			p.hasZ = false
			p.hasM = true
		case "ZM":
			p.hasZ = true
			p.hasM = true
		case "EMPTY":
			return p.createEmpty(typeUpper)
		default:
			if modifier != "" {
				return nil, fmt.Errorf("unknown dimension modifier %q for geometry type %s", modifier, typeName)
			}
		}
	}

	p.skipWhitespace()

	if p.matchWord("EMPTY") {
		return p.createEmpty(typeUpper)
	}

	switch typeUpper {
	case "POINT":
		return p.parsePoint()
	case "LINESTRING":
		return p.parseLineString()
	case "LINEARRING":
		return p.parseLinearRing()
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
		return nil, fmt.Errorf("unknown geometry type: %s", typeName)
	}
}

func (p *parser) createEmpty(typeName string) (geom.Geometry, error) {
	switch typeName {
	case "POINT":
		return p.factory.CreatePointEmpty(), nil
	case "LINESTRING":
		return p.factory.CreateLineStringEmpty(), nil
	case "LINEARRING":
		return p.factory.CreateLinearRingEmpty(), nil
	case "POLYGON":
		return p.factory.CreatePolygonEmpty(), nil
	case "MULTIPOINT":
		return p.factory.CreateMultiPointEmpty(), nil
	case "MULTILINESTRING":
		return p.factory.CreateMultiLineStringEmpty(), nil
	case "MULTIPOLYGON":
		return p.factory.CreateMultiPolygonEmpty(), nil
	case "GEOMETRYCOLLECTION":
		return p.factory.CreateGeometryCollectionEmpty(), nil
	default:
		return nil, fmt.Errorf("unknown geometry type: %s", typeName)
	}
}

func (p *parser) parsePoint() (*geom.Point, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	coord, err := p.parseCoordinate()
	if err != nil {
		return nil, err
	}
	if err := p.expect(')'); err != nil {
		return nil, err
	}
	return p.factory.CreatePointFromCoordinate(coord), nil
}

func (p *parser) parseLineString() (*geom.LineString, error) {
	coords, err := p.parseCoordinateSequence()
	if err != nil {
		return nil, err
	}
	if err := validateWKTLineString("LINESTRING", coords); err != nil {
		return nil, err
	}
	return p.factory.CreateLineString(coords), nil
}

func (p *parser) parseLinearRing() (*geom.LinearRing, error) {
	coords, err := p.parseCoordinateSequence()
	if err != nil {
		return nil, err
	}
	if err := validateWKTRing("LINEARRING", coords); err != nil {
		return nil, err
	}
	return p.factory.CreateLinearRing(coords), nil
}

func (p *parser) parsePolygon() (*geom.Polygon, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}

	shellCoords, err := p.parseCoordinateSequence()
	if err != nil {
		return nil, err
	}
	if err := validateWKTRing("polygon shell", shellCoords); err != nil {
		return nil, err
	}
	shell := p.factory.CreateLinearRing(shellCoords)

	var holes []*geom.LinearRing
	for {
		p.skipWhitespace()
		if p.peek() == ')' {
			break
		}
		if err := p.expect(','); err != nil {
			return nil, err
		}
		holeCoords, err := p.parseCoordinateSequence()
		if err != nil {
			return nil, err
		}
		if err := validateWKTRing(fmt.Sprintf("polygon hole %d", len(holes)), holeCoords); err != nil {
			return nil, err
		}
		holes = append(holes, p.factory.CreateLinearRing(holeCoords))
	}

	if err := p.expect(')'); err != nil {
		return nil, err
	}

	return p.factory.CreatePolygon(shell, holes), nil
}

func (p *parser) parseMultiPoint() (*geom.MultiPoint, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	if err := p.rejectEmptyCollection("MULTIPOINT"); err != nil {
		return nil, err
	}

	var points []*geom.Point
	for {
		p.skipWhitespace()

		if p.peek() == '(' {
			p.advance()
			coord, err := p.parseCoordinate()
			if err != nil {
				return nil, err
			}
			if err := p.expect(')'); err != nil {
				return nil, err
			}
			points = append(points, p.factory.CreatePointFromCoordinate(coord))
		} else if p.peek() == ')' {
			break
		} else {
			coord, err := p.parseCoordinate()
			if err != nil {
				return nil, err
			}
			points = append(points, p.factory.CreatePointFromCoordinate(coord))
		}

		p.skipWhitespace()
		if p.peek() == ')' {
			break
		}
		if err := p.expect(','); err != nil {
			return nil, err
		}
		if err := p.rejectTrailingComma("MULTIPOINT"); err != nil {
			return nil, err
		}
	}

	if err := p.expect(')'); err != nil {
		return nil, err
	}

	return p.factory.CreateMultiPoint(points), nil
}

func (p *parser) parseMultiLineString() (*geom.MultiLineString, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	if err := p.rejectEmptyCollection("MULTILINESTRING"); err != nil {
		return nil, err
	}

	var lines []*geom.LineString
	for {
		p.skipWhitespace()
		if p.peek() == ')' {
			break
		}

		coords, err := p.parseCoordinateSequence()
		if err != nil {
			return nil, err
		}
		if err := validateWKTLineString("MULTILINESTRING element", coords); err != nil {
			return nil, err
		}
		lines = append(lines, p.factory.CreateLineString(coords))

		p.skipWhitespace()
		if p.peek() == ')' {
			break
		}
		if err := p.expect(','); err != nil {
			return nil, err
		}
		if err := p.rejectTrailingComma("MULTILINESTRING"); err != nil {
			return nil, err
		}
	}

	if err := p.expect(')'); err != nil {
		return nil, err
	}

	return p.factory.CreateMultiLineString(lines), nil
}

func (p *parser) parseMultiPolygon() (*geom.MultiPolygon, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	if err := p.rejectEmptyCollection("MULTIPOLYGON"); err != nil {
		return nil, err
	}

	var polygons []*geom.Polygon
	for {
		p.skipWhitespace()
		if p.peek() == ')' {
			break
		}

		poly, err := p.parsePolygonInner()
		if err != nil {
			return nil, err
		}
		polygons = append(polygons, poly)

		p.skipWhitespace()
		if p.peek() == ')' {
			break
		}
		if err := p.expect(','); err != nil {
			return nil, err
		}
		if err := p.rejectTrailingComma("MULTIPOLYGON"); err != nil {
			return nil, err
		}
	}

	if err := p.expect(')'); err != nil {
		return nil, err
	}

	return p.factory.CreateMultiPolygon(polygons), nil
}

func (p *parser) parsePolygonInner() (*geom.Polygon, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}

	shellCoords, err := p.parseCoordinateSequence()
	if err != nil {
		return nil, err
	}
	if err := validateWKTRing("polygon shell", shellCoords); err != nil {
		return nil, err
	}
	shell := p.factory.CreateLinearRing(shellCoords)

	var holes []*geom.LinearRing
	for {
		p.skipWhitespace()
		if p.peek() == ')' {
			break
		}
		if err := p.expect(','); err != nil {
			return nil, err
		}
		holeCoords, err := p.parseCoordinateSequence()
		if err != nil {
			return nil, err
		}
		if err := validateWKTRing(fmt.Sprintf("polygon hole %d", len(holes)), holeCoords); err != nil {
			return nil, err
		}
		holes = append(holes, p.factory.CreateLinearRing(holeCoords))
	}

	if err := p.expect(')'); err != nil {
		return nil, err
	}

	return p.factory.CreatePolygon(shell, holes), nil
}

func (p *parser) parseGeometryCollection() (*geom.GeometryCollection, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	if err := p.rejectEmptyCollection("GEOMETRYCOLLECTION"); err != nil {
		return nil, err
	}

	var geometries []geom.Geometry
	for {
		p.skipWhitespace()
		if p.peek() == ')' {
			break
		}

		nestedParser := &parser{
			input:       p.input[p.pos:],
			pos:         0,
			factory:     p.factory,
			defaultHasZ: p.hasZ,
			defaultHasM: p.hasM,
		}
		g, err := nestedParser.parse()
		if err != nil {
			return nil, err
		}
		geometries = append(geometries, g)
		p.pos += nestedParser.pos

		p.skipWhitespace()
		if p.peek() == ')' {
			break
		}
		if err := p.expect(','); err != nil {
			return nil, err
		}
		if err := p.rejectTrailingComma("GEOMETRYCOLLECTION"); err != nil {
			return nil, err
		}
	}

	if err := p.expect(')'); err != nil {
		return nil, err
	}

	return p.factory.CreateGeometryCollection(geometries), nil
}

func (p *parser) parseCoordinateSequence() (geom.CoordinateSequence, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	if err := p.rejectEmptyCollection("coordinate sequence"); err != nil {
		return nil, err
	}

	var coords geom.CoordinateSequence
	for {
		p.skipWhitespace()
		if p.peek() == ')' {
			break
		}

		coord, err := p.parseCoordinate()
		if err != nil {
			return nil, err
		}
		coords = append(coords, coord)

		p.skipWhitespace()
		if p.peek() == ')' {
			break
		}
		if err := p.expect(','); err != nil {
			return nil, err
		}
		if err := p.rejectTrailingComma("coordinate sequence"); err != nil {
			return nil, err
		}
	}

	if err := p.expect(')'); err != nil {
		return nil, err
	}

	return coords, nil
}

func validateWKTLineString(label string, coords geom.CoordinateSequence) error {
	if len(coords) < 2 {
		return fmt.Errorf("%s must have at least 2 points", label)
	}
	return nil
}

func validateWKTRing(label string, coords geom.CoordinateSequence) error {
	if len(coords) < 4 {
		return fmt.Errorf("%s must have at least 4 points", label)
	}
	if !coords.IsClosed(geom.DefaultEpsilon) {
		return fmt.Errorf("%s is not closed", label)
	}
	return nil
}

func (p *parser) rejectEmptyCollection(name string) error {
	p.skipWhitespace()
	if p.peek() == ')' {
		return fmt.Errorf("%s cannot be empty; use EMPTY", name)
	}
	return nil
}

func (p *parser) rejectTrailingComma(name string) error {
	p.skipWhitespace()
	if p.peek() == ')' {
		return fmt.Errorf("%s has trailing comma at position %d", name, p.pos)
	}
	return nil
}

func (p *parser) parseCoordinate() (geom.Coordinate, error) {
	p.skipWhitespace()

	x, err := p.parseNumber()
	if err != nil {
		return geom.Coordinate{}, fmt.Errorf("expected X coordinate: %w", err)
	}

	p.skipWhitespace()
	y, err := p.parseNumber()
	if err != nil {
		return geom.Coordinate{}, fmt.Errorf("expected Y coordinate: %w", err)
	}

	coord := geom.NewCoordinate(x, y)

	// When modifier is set, enforce strict arity
	if p.hasZ && p.hasM {
		// ZM: require exactly 2 more numbers
		p.skipWhitespace()
		if !p.isDigitOrSign() {
			return geom.Coordinate{}, fmt.Errorf("expected Z coordinate for ZM geometry at position %d", p.pos)
		}
		z, err := p.parseNumber()
		if err != nil {
			return geom.Coordinate{}, fmt.Errorf("expected Z coordinate: %w", err)
		}
		coord.Z = z

		p.skipWhitespace()
		if !p.isDigitOrSign() {
			return geom.Coordinate{}, fmt.Errorf("expected M coordinate for ZM geometry at position %d", p.pos)
		}
		m, err := p.parseNumber()
		if err != nil {
			return geom.Coordinate{}, fmt.Errorf("expected M coordinate: %w", err)
		}
		coord.M = m
	} else if p.hasZ {
		// Z: require exactly 1 more number
		p.skipWhitespace()
		if !p.isDigitOrSign() {
			return geom.Coordinate{}, fmt.Errorf("expected Z coordinate for Z geometry at position %d", p.pos)
		}
		z, err := p.parseNumber()
		if err != nil {
			return geom.Coordinate{}, fmt.Errorf("expected Z coordinate: %w", err)
		}
		coord.Z = z
	} else if p.hasM {
		// M: require exactly 1 more number (stored as M)
		p.skipWhitespace()
		if !p.isDigitOrSign() {
			return geom.Coordinate{}, fmt.Errorf("expected M coordinate for M geometry at position %d", p.pos)
		}
		m, err := p.parseNumber()
		if err != nil {
			return geom.Coordinate{}, fmt.Errorf("expected M coordinate: %w", err)
		}
		coord.M = m
	} else {
		// No modifier: auto-detect extra coords (backward compat)
		p.skipWhitespace()
		if p.isDigitOrSign() {
			z, err := p.parseNumber()
			if err == nil {
				coord.Z = z

				p.skipWhitespace()
				if p.isDigitOrSign() {
					m, err := p.parseNumber()
					if err == nil {
						coord.M = m
					}
				}
			}
		}
	}

	return coord, nil
}

func (p *parser) parseNumber() (float64, error) {
	p.skipWhitespace()
	start := p.pos

	if p.peek() == '-' || p.peek() == '+' {
		p.advance()
	}

	for p.isDigit() {
		p.advance()
	}

	if p.peek() == '.' {
		p.advance()
		for p.isDigit() {
			p.advance()
		}
	}

	if p.peek() == 'e' || p.peek() == 'E' {
		p.advance()
		if p.peek() == '-' || p.peek() == '+' {
			p.advance()
		}
		for p.isDigit() {
			p.advance()
		}
	}

	if start == p.pos {
		return 0, fmt.Errorf("expected number at position %d", p.pos)
	}

	return strconv.ParseFloat(p.input[start:p.pos], 64)
}

func (p *parser) skipWhitespace() {
	for p.pos < len(p.input) && unicode.IsSpace(rune(p.input[p.pos])) {
		p.pos++
	}
}

func (p *parser) peek() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *parser) advance() {
	if p.pos < len(p.input) {
		p.pos++
	}
}

func (p *parser) expect(c byte) error {
	p.skipWhitespace()
	if p.peek() != c {
		return fmt.Errorf("expected '%c' at position %d, got '%c'", c, p.pos, p.peek())
	}
	p.advance()
	return nil
}

func (p *parser) readWord() string {
	p.skipWhitespace()
	start := p.pos
	for p.pos < len(p.input) && (unicode.IsLetter(rune(p.input[p.pos])) || p.input[p.pos] == '_') {
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *parser) matchWord(word string) bool {
	p.skipWhitespace()
	if p.pos+len(word) > len(p.input) {
		return false
	}
	if strings.EqualFold(p.input[p.pos:p.pos+len(word)], word) {
		if p.pos+len(word) < len(p.input) && unicode.IsLetter(rune(p.input[p.pos+len(word)])) {
			return false
		}
		p.pos += len(word)
		return true
	}
	return false
}

func (p *parser) isDigit() bool {
	c := p.peek()
	return c >= '0' && c <= '9'
}

func (p *parser) isDigitOrSign() bool {
	c := p.peek()
	return (c >= '0' && c <= '9') || c == '-' || c == '+'
}

func (p *parser) atEnd() bool {
	return p.pos >= len(p.input)
}
