// Package wkb provides Well-Known Binary (WKB) encoding/decoding for geometries.
//
// Use Marshal/Unmarshal for standard WKB operations:
//
//	data, err := wkb.Marshal(point)
//	geom, err := wkb.Unmarshal(data)
//
// Use MarshalEWKB for Extended WKB with SRID:
//
//	data, err := wkb.MarshalEWKB(point)
package wkb

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// WKB geometry type constants
const (
	wkbPoint              = 1
	wkbLineString         = 2
	wkbPolygon            = 3
	wkbMultiPoint         = 4
	wkbMultiLineString    = 5
	wkbMultiPolygon       = 6
	wkbGeometryCollection = 7

	// EWKB SRID flag
	wkbSRIDFlag = 0x20000000
)

// Byte order constants
const (
	wkbXDR = 0 // Big endian
	wkbNDR = 1 // Little endian
)

// Options configures WKB marshaling behavior.
type Options struct {
	// ByteOrder specifies the byte order (default: LittleEndian).
	ByteOrder binary.ByteOrder
	// IncludeSRID specifies whether to include SRID (EWKB format).
	IncludeSRID bool
	// OutputDimension specifies coordinate dimensions (2, 3, or 4).
	OutputDimension int
}

// DefaultOptions returns the default marshaling options.
func DefaultOptions() Options {
	return Options{
		ByteOrder:       binary.LittleEndian,
		IncludeSRID:     false,
		OutputDimension: 2,
	}
}

// Marshal marshals a geometry to WKB bytes.
func Marshal(g geom.Geometry) ([]byte, error) {
	return MarshalWithOptions(g, DefaultOptions())
}

// MarshalWithOptions marshals with custom options.
func MarshalWithOptions(g geom.Geometry, opts Options) ([]byte, error) {
	if g == nil {
		return nil, fmt.Errorf("wkb: cannot marshal nil geometry")
	}
	opts = normalizeOptions(opts)

	buf := &buffer{
		data:  make([]byte, 0, 256),
		order: opts.ByteOrder,
	}

	if err := writeGeometry(buf, g, opts); err != nil {
		return nil, err
	}
	return buf.data, nil
}

func normalizeOptions(opts Options) Options {
	if opts.ByteOrder == nil {
		opts.ByteOrder = binary.LittleEndian
	}
	if opts.OutputDimension == 0 {
		opts.OutputDimension = 2
	}
	return opts
}

// MarshalEWKB marshals a geometry to Extended WKB with SRID.
func MarshalEWKB(g geom.Geometry) ([]byte, error) {
	opts := DefaultOptions()
	opts.IncludeSRID = true
	return MarshalWithOptions(g, opts)
}

// Unmarshal unmarshals WKB bytes to a geometry.
func Unmarshal(data []byte) (geom.Geometry, error) {
	return UnmarshalWithFactory(data, geom.DefaultFactory)
}

// UnmarshalWithFactory unmarshals using a custom geometry factory.
func UnmarshalWithFactory(data []byte, factory *geom.GeometryFactory) (geom.Geometry, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty WKB data")
	}
	if factory == nil {
		factory = geom.DefaultFactory
	}

	p := &parser{
		data:    data,
		pos:     0,
		factory: factory,
	}

	g, err := p.readGeometry()
	if err != nil {
		return nil, err
	}

	if p.pos != len(p.data) {
		return nil, fmt.Errorf("unexpected %d trailing bytes after WKB geometry", len(p.data)-p.pos)
	}

	return g, nil
}

// --- Marshaling implementation ---

func writeGeometry(buf *buffer, g geom.Geometry, opts Options) error {
	if opts.ByteOrder != binary.LittleEndian && opts.ByteOrder != binary.BigEndian {
		return fmt.Errorf("wkb: unsupported byte order %T", opts.ByteOrder)
	}
	if opts.OutputDimension < 2 || opts.OutputDimension > 4 {
		return fmt.Errorf("wkb: unsupported output dimension %d", opts.OutputDimension)
	}

	// Write byte order
	if opts.ByteOrder == binary.LittleEndian {
		buf.writeByte(wkbNDR)
	} else {
		buf.writeByte(wkbXDR)
	}

	// Determine geometry type
	baseType, ok := geometryType(g)
	if !ok {
		return fmt.Errorf("wkb: unsupported geometry type %T", g)
	}

	// Add dimension flags
	coords := g.Coordinates()
	hasZ := opts.OutputDimension >= 3 && coords.HasZ()
	hasM := opts.OutputDimension >= 4 && coords.HasM()

	geomType := baseType
	if hasZ && hasM {
		geomType += 3000
	} else if hasZ {
		geomType += 1000
	} else if hasM {
		geomType += 2000
	}

	// Add SRID flag if needed
	if opts.IncludeSRID && g.SRID() != 0 {
		geomType |= wkbSRIDFlag
	}

	buf.writeUint32(geomType)

	// Write SRID if flagged
	if opts.IncludeSRID && g.SRID() != 0 {
		buf.writeUint32(uint32(g.SRID()))
	}

	// Write geometry data
	switch v := g.(type) {
	case *geom.Point:
		writePoint(buf, v, hasZ, hasM)
	case *geom.LineString:
		writeLineString(buf, v, hasZ, hasM)
	case *geom.LinearRing:
		writeLineString(buf, v.LineString, hasZ, hasM)
	case *geom.Polygon:
		writePolygon(buf, v, hasZ, hasM)
	case *geom.MultiPoint:
		return writeMultiPoint(buf, v, opts)
	case *geom.MultiLineString:
		return writeMultiLineString(buf, v, opts)
	case *geom.MultiPolygon:
		return writeMultiPolygon(buf, v, opts)
	case *geom.GeometryCollection:
		return writeGeometryCollection(buf, v, opts)
	}

	return nil
}

func geometryType(g geom.Geometry) (uint32, bool) {
	switch g.(type) {
	case *geom.Point:
		return wkbPoint, true
	case *geom.LineString:
		return wkbLineString, true
	case *geom.LinearRing:
		return wkbLineString, true // LinearRing is encoded as LineString.
	case *geom.Polygon:
		return wkbPolygon, true
	case *geom.MultiPoint:
		return wkbMultiPoint, true
	case *geom.MultiLineString:
		return wkbMultiLineString, true
	case *geom.MultiPolygon:
		return wkbMultiPolygon, true
	case *geom.GeometryCollection:
		return wkbGeometryCollection, true
	default:
		return 0, false
	}
}

func writePoint(buf *buffer, p *geom.Point, hasZ, hasM bool) {
	if p.IsEmpty() {
		buf.writeFloat64(math.NaN())
		buf.writeFloat64(math.NaN())
		if hasZ {
			buf.writeFloat64(math.NaN())
		}
		if hasM {
			buf.writeFloat64(math.NaN())
		}
		return
	}

	coord := p.Coordinate()
	buf.writeFloat64(coord.X)
	buf.writeFloat64(coord.Y)
	if hasZ {
		buf.writeFloat64(coord.GetZ())
	}
	if hasM {
		buf.writeFloat64(coord.GetM())
	}
}

func writeLineString(buf *buffer, ls *geom.LineString, hasZ, hasM bool) {
	coords := ls.Coordinates()
	buf.writeUint32(uint32(len(coords)))
	for _, c := range coords {
		buf.writeFloat64(c.X)
		buf.writeFloat64(c.Y)
		if hasZ {
			buf.writeFloat64(c.GetZ())
		}
		if hasM {
			buf.writeFloat64(c.GetM())
		}
	}
}

func writePolygon(buf *buffer, p *geom.Polygon, hasZ, hasM bool) {
	if p.IsEmpty() {
		buf.writeUint32(0)
		return
	}

	numRings := uint32(1 + p.NumInteriorRings())
	buf.writeUint32(numRings)

	// Write shell
	shellCoords := p.ExteriorRing().Coordinates()
	buf.writeUint32(uint32(len(shellCoords)))
	for _, c := range shellCoords {
		buf.writeFloat64(c.X)
		buf.writeFloat64(c.Y)
		if hasZ {
			buf.writeFloat64(c.GetZ())
		}
		if hasM {
			buf.writeFloat64(c.GetM())
		}
	}

	// Write holes
	for i := 0; i < p.NumInteriorRings(); i++ {
		holeCoords := p.InteriorRingN(i).Coordinates()
		buf.writeUint32(uint32(len(holeCoords)))
		for _, c := range holeCoords {
			buf.writeFloat64(c.X)
			buf.writeFloat64(c.Y)
			if hasZ {
				buf.writeFloat64(c.GetZ())
			}
			if hasM {
				buf.writeFloat64(c.GetM())
			}
		}
	}
}

func writeMultiPoint(buf *buffer, mp *geom.MultiPoint, opts Options) error {
	buf.writeUint32(uint32(mp.NumGeometries()))
	for i := 0; i < mp.NumGeometries(); i++ {
		if err := writeGeometry(buf, mp.GeometryN(i), opts); err != nil {
			return err
		}
	}
	return nil
}

func writeMultiLineString(buf *buffer, mls *geom.MultiLineString, opts Options) error {
	buf.writeUint32(uint32(mls.NumGeometries()))
	for i := 0; i < mls.NumGeometries(); i++ {
		if err := writeGeometry(buf, mls.GeometryN(i), opts); err != nil {
			return err
		}
	}
	return nil
}

func writeMultiPolygon(buf *buffer, mp *geom.MultiPolygon, opts Options) error {
	buf.writeUint32(uint32(mp.NumGeometries()))
	for i := 0; i < mp.NumGeometries(); i++ {
		if err := writeGeometry(buf, mp.GeometryN(i), opts); err != nil {
			return err
		}
	}
	return nil
}

func writeGeometryCollection(buf *buffer, gc *geom.GeometryCollection, opts Options) error {
	buf.writeUint32(uint32(gc.NumGeometries()))
	for i := 0; i < gc.NumGeometries(); i++ {
		if err := writeGeometry(buf, gc.GeometryN(i), opts); err != nil {
			return err
		}
	}
	return nil
}

// buffer is a simple byte buffer for WKB writing.
type buffer struct {
	data  []byte
	order binary.ByteOrder
}

func (b *buffer) writeByte(v byte) {
	b.data = append(b.data, v)
}

func (b *buffer) writeUint32(v uint32) {
	buf := make([]byte, 4)
	b.order.PutUint32(buf, v)
	b.data = append(b.data, buf...)
}

func (b *buffer) writeFloat64(v float64) {
	buf := make([]byte, 8)
	b.order.PutUint64(buf, math.Float64bits(v))
	b.data = append(b.data, buf...)
}

// --- Parsing implementation ---

type parserState struct {
	hasZ      bool
	hasM      bool
	hasSRID   bool
	srid      int
	coordSize int
	order     binary.ByteOrder
}

type parser struct {
	data      []byte
	pos       int
	order     binary.ByteOrder
	factory   *geom.GeometryFactory
	hasZ      bool
	hasM      bool
	hasSRID   bool
	srid      int
	coordSize int
}

func (p *parser) saveState() parserState {
	return parserState{
		hasZ:      p.hasZ,
		hasM:      p.hasM,
		hasSRID:   p.hasSRID,
		srid:      p.srid,
		coordSize: p.coordSize,
		order:     p.order,
	}
}

func (p *parser) restoreState(s parserState) {
	p.hasZ = s.hasZ
	p.hasM = s.hasM
	p.hasSRID = s.hasSRID
	p.srid = s.srid
	p.coordSize = s.coordSize
	p.order = s.order
}

func (p *parser) readGeometry() (geom.Geometry, error) {
	// Read byte order
	if p.pos >= len(p.data) {
		return nil, fmt.Errorf("unexpected end of data")
	}

	byteOrder := p.data[p.pos]
	p.pos++

	switch byteOrder {
	case wkbXDR:
		p.order = binary.BigEndian
	case wkbNDR:
		p.order = binary.LittleEndian
	default:
		return nil, fmt.Errorf("invalid byte order: %d", byteOrder)
	}

	// Read geometry type
	geomType, err := p.readUint32()
	if err != nil {
		return nil, err
	}

	// Check for SRID flag (EWKB)
	p.hasSRID = (geomType & wkbSRIDFlag) != 0
	if p.hasSRID {
		geomType &= ^uint32(wkbSRIDFlag)
		srid, err := p.readUint32()
		if err != nil {
			return nil, err
		}
		p.srid = int(srid)
	}

	// Determine coordinate dimensions
	baseType := geomType % 1000
	dimFlag := geomType / 1000
	if dimFlag > 3 {
		return nil, fmt.Errorf("unsupported coordinate dimension flag: %d", dimFlag)
	}

	p.hasZ = dimFlag == 1 || dimFlag == 3
	p.hasM = dimFlag == 2 || dimFlag == 3

	p.coordSize = 2
	if p.hasZ {
		p.coordSize++
	}
	if p.hasM {
		p.coordSize++
	}

	var g geom.Geometry

	switch baseType {
	case wkbPoint:
		g, err = p.readPoint()
	case wkbLineString:
		g, err = p.readLineString()
	case wkbPolygon:
		g, err = p.readPolygon()
	case wkbMultiPoint:
		g, err = p.readMultiPoint()
	case wkbMultiLineString:
		g, err = p.readMultiLineString()
	case wkbMultiPolygon:
		g, err = p.readMultiPolygon()
	case wkbGeometryCollection:
		g, err = p.readGeometryCollection()
	default:
		return nil, fmt.Errorf("unsupported geometry type: %d", baseType)
	}

	if err != nil {
		return nil, err
	}

	if p.hasSRID {
		// SetSRID is available on all concrete geometry types but not the Geometry interface.
		type sridSetter interface{ SetSRID(int) }
		if s, ok := g.(sridSetter); ok {
			s.SetSRID(p.srid)
		}
	}

	return g, nil
}

func (p *parser) readPoint() (*geom.Point, error) {
	coord, err := p.readCoordinate()
	if err != nil {
		return nil, err
	}

	// Check for empty point (NaN coordinates)
	if math.IsNaN(coord.X) && math.IsNaN(coord.Y) {
		return p.factory.CreatePointEmpty(), nil
	}

	return p.factory.CreatePointFromCoordinate(coord), nil
}

func (p *parser) readLineString() (*geom.LineString, error) {
	numPoints, err := p.readUint32()
	if err != nil {
		return nil, err
	}

	if numPoints == 0 {
		return p.factory.CreateLineStringEmpty(), nil
	}
	if numPoints == 1 {
		return nil, fmt.Errorf("wkb: LineString must have 0 or at least 2 points, got %d", numPoints)
	}
	if err := p.ensureCoordinatePayload(numPoints, "LineString"); err != nil {
		return nil, err
	}

	coords, err := p.readCoordinates(int(numPoints))
	if err != nil {
		return nil, err
	}

	return p.factory.CreateLineString(coords), nil
}

func (p *parser) readPolygon() (*geom.Polygon, error) {
	numRings, err := p.readUint32()
	if err != nil {
		return nil, err
	}

	if numRings == 0 {
		return p.factory.CreatePolygonEmpty(), nil
	}
	if err := p.ensureCountFitsRemaining(numRings, 4, "polygon ring"); err != nil {
		return nil, err
	}

	// Read shell
	shellPoints, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	if err := p.ensureCoordinatePayload(shellPoints, "polygon shell"); err != nil {
		return nil, err
	}
	shellCoords, err := p.readCoordinates(int(shellPoints))
	if err != nil {
		return nil, err
	}
	if err := validateWKBRing("polygon shell", shellCoords); err != nil {
		return nil, err
	}
	shell := p.factory.CreateLinearRing(shellCoords)

	// Read holes
	holes := make([]*geom.LinearRing, numRings-1)
	for i := uint32(1); i < numRings; i++ {
		holePoints, err := p.readUint32()
		if err != nil {
			return nil, err
		}
		if err := p.ensureCoordinatePayload(holePoints, fmt.Sprintf("polygon hole %d", i-1)); err != nil {
			return nil, err
		}
		holeCoords, err := p.readCoordinates(int(holePoints))
		if err != nil {
			return nil, err
		}
		if err := validateWKBRing(fmt.Sprintf("polygon hole %d", i-1), holeCoords); err != nil {
			return nil, err
		}
		holes[i-1] = p.factory.CreateLinearRing(holeCoords)
	}

	return p.factory.CreatePolygon(shell, holes), nil
}

func validateWKBRing(label string, coords geom.CoordinateSequence) error {
	if len(coords) < 4 {
		return fmt.Errorf("wkb: %s must have at least 4 points, got %d", label, len(coords))
	}
	if !coords.IsClosed(geom.DefaultEpsilon) {
		return fmt.Errorf("wkb: %s is not closed", label)
	}
	return nil
}

func (p *parser) readMultiPoint() (*geom.MultiPoint, error) {
	numGeoms, err := p.readUint32()
	if err != nil {
		return nil, err
	}

	if numGeoms == 0 {
		return p.factory.CreateMultiPointEmpty(), nil
	}
	if err := p.ensureCountFitsRemaining(numGeoms, 5, "MultiPoint child geometry"); err != nil {
		return nil, err
	}

	points := make([]*geom.Point, numGeoms)
	childSRIDs := make([]int, numGeoms)
	for i := uint32(0); i < numGeoms; i++ {
		state := p.saveState()
		g, err := p.readGeometry()
		p.restoreState(state)
		if err != nil {
			return nil, err
		}
		pt, ok := g.(*geom.Point)
		if !ok {
			return nil, fmt.Errorf("expected Point in MultiPoint, got %T", g)
		}
		points[i] = pt
		childSRIDs[i] = pt.SRID()
	}

	mp := p.factory.CreateMultiPoint(points)
	for i, srid := range childSRIDs {
		if child, ok := mp.GeometryN(i).(interface{ SetSRID(int) }); ok {
			child.SetSRID(srid)
		}
	}
	return mp, nil
}

func (p *parser) readMultiLineString() (*geom.MultiLineString, error) {
	numGeoms, err := p.readUint32()
	if err != nil {
		return nil, err
	}

	if numGeoms == 0 {
		return p.factory.CreateMultiLineStringEmpty(), nil
	}
	if err := p.ensureCountFitsRemaining(numGeoms, 5, "MultiLineString child geometry"); err != nil {
		return nil, err
	}

	lines := make([]*geom.LineString, numGeoms)
	childSRIDs := make([]int, numGeoms)
	for i := uint32(0); i < numGeoms; i++ {
		state := p.saveState()
		g, err := p.readGeometry()
		p.restoreState(state)
		if err != nil {
			return nil, err
		}
		ls, ok := g.(*geom.LineString)
		if !ok {
			return nil, fmt.Errorf("expected LineString in MultiLineString, got %T", g)
		}
		lines[i] = ls
		childSRIDs[i] = ls.SRID()
	}

	mls := p.factory.CreateMultiLineString(lines)
	for i, srid := range childSRIDs {
		if child, ok := mls.GeometryN(i).(interface{ SetSRID(int) }); ok {
			child.SetSRID(srid)
		}
	}
	return mls, nil
}

func (p *parser) readMultiPolygon() (*geom.MultiPolygon, error) {
	numGeoms, err := p.readUint32()
	if err != nil {
		return nil, err
	}

	if numGeoms == 0 {
		return p.factory.CreateMultiPolygonEmpty(), nil
	}
	if err := p.ensureCountFitsRemaining(numGeoms, 5, "MultiPolygon child geometry"); err != nil {
		return nil, err
	}

	polys := make([]*geom.Polygon, numGeoms)
	childSRIDs := make([]int, numGeoms)
	for i := uint32(0); i < numGeoms; i++ {
		state := p.saveState()
		g, err := p.readGeometry()
		p.restoreState(state)
		if err != nil {
			return nil, err
		}
		poly, ok := g.(*geom.Polygon)
		if !ok {
			return nil, fmt.Errorf("expected Polygon in MultiPolygon, got %T", g)
		}
		polys[i] = poly
		childSRIDs[i] = poly.SRID()
	}

	mp := p.factory.CreateMultiPolygon(polys)
	for i, srid := range childSRIDs {
		if child, ok := mp.GeometryN(i).(interface{ SetSRID(int) }); ok {
			child.SetSRID(srid)
		}
	}
	return mp, nil
}

func (p *parser) readGeometryCollection() (*geom.GeometryCollection, error) {
	numGeoms, err := p.readUint32()
	if err != nil {
		return nil, err
	}

	if numGeoms == 0 {
		return p.factory.CreateGeometryCollectionEmpty(), nil
	}
	if err := p.ensureCountFitsRemaining(numGeoms, 5, "GeometryCollection child geometry"); err != nil {
		return nil, err
	}

	geoms := make([]geom.Geometry, numGeoms)
	for i := uint32(0); i < numGeoms; i++ {
		state := p.saveState()
		g, err := p.readGeometry()
		p.restoreState(state)
		if err != nil {
			return nil, err
		}
		geoms[i] = g
	}

	return p.factory.CreateGeometryCollection(geoms), nil
}

func (p *parser) readCoordinate() (geom.Coordinate, error) {
	x, err := p.readFloat64()
	if err != nil {
		return geom.Coordinate{}, err
	}
	y, err := p.readFloat64()
	if err != nil {
		return geom.Coordinate{}, err
	}

	coord := geom.NewCoordinate(x, y)

	if p.hasZ {
		z, err := p.readFloat64()
		if err != nil {
			return geom.Coordinate{}, err
		}
		coord.Z = z
	}

	if p.hasM {
		m, err := p.readFloat64()
		if err != nil {
			return geom.Coordinate{}, err
		}
		coord.M = m
	}

	return coord, nil
}

func (p *parser) readCoordinates(n int) (geom.CoordinateSequence, error) {
	coords := make(geom.CoordinateSequence, n)
	for i := 0; i < n; i++ {
		coord, err := p.readCoordinate()
		if err != nil {
			return nil, err
		}
		coords[i] = coord
	}
	return coords, nil
}

func (p *parser) ensureCoordinatePayload(count uint32, label string) error {
	bytesPerCoordinate := uint64(p.coordSize) * 8
	return p.ensureCountFitsRemaining(count, bytesPerCoordinate, label+" coordinate")
}

func (p *parser) ensureCountFitsRemaining(count uint32, bytesPerItem uint64, label string) error {
	if count == 0 {
		return nil
	}
	if bytesPerItem == 0 {
		return fmt.Errorf("wkb: invalid byte width for %s payload", label)
	}
	remaining := uint64(len(p.data) - p.pos)
	needed := uint64(count) * bytesPerItem
	if needed/bytesPerItem != uint64(count) || needed > remaining {
		return fmt.Errorf("wkb: %s count %d exceeds remaining payload", label, count)
	}
	if uint64(count) > uint64(maxInt()) {
		return fmt.Errorf("wkb: %s count %d exceeds platform allocation limit", label, count)
	}
	return nil
}

func maxInt() int {
	return int(^uint(0) >> 1)
}

func (p *parser) readUint32() (uint32, error) {
	if p.pos+4 > len(p.data) {
		return 0, io.ErrUnexpectedEOF
	}
	val := p.order.Uint32(p.data[p.pos:])
	p.pos += 4
	return val, nil
}

func (p *parser) readFloat64() (float64, error) {
	if p.pos+8 > len(p.data) {
		return 0, io.ErrUnexpectedEOF
	}
	bits := p.order.Uint64(p.data[p.pos:])
	p.pos += 8
	return math.Float64frombits(bits), nil
}
