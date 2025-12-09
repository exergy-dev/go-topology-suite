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

	"github.com/go-topology-suite/gts/geom"
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

	// With Z
	wkbPointZ              = 1001
	wkbLineStringZ         = 1002
	wkbPolygonZ            = 1003
	wkbMultiPointZ         = 1004
	wkbMultiLineStringZ    = 1005
	wkbMultiPolygonZ       = 1006
	wkbGeometryCollectionZ = 1007

	// With M
	wkbPointM              = 2001
	wkbLineStringM         = 2002
	wkbPolygonM            = 2003
	wkbMultiPointM         = 2004
	wkbMultiLineStringM    = 2005
	wkbMultiPolygonM       = 2006
	wkbGeometryCollectionM = 2007

	// With ZM
	wkbPointZM              = 3001
	wkbLineStringZM         = 3002
	wkbPolygonZM            = 3003
	wkbMultiPointZM         = 3004
	wkbMultiLineStringZM    = 3005
	wkbMultiPolygonZM       = 3006
	wkbGeometryCollectionZM = 3007

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
		return nil, nil
	}

	buf := &buffer{
		data:  make([]byte, 0, 256),
		order: opts.ByteOrder,
	}

	writeGeometry(buf, g, opts)
	return buf.data, nil
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

	p := &parser{
		data:    data,
		pos:     0,
		factory: factory,
	}

	return p.readGeometry()
}

// --- Marshaling implementation ---

func writeGeometry(buf *buffer, g geom.Geometry, opts Options) {
	// Write byte order
	if opts.ByteOrder == binary.LittleEndian {
		buf.writeByte(wkbNDR)
	} else {
		buf.writeByte(wkbXDR)
	}

	// Determine geometry type
	var baseType uint32
	switch g.(type) {
	case *geom.Point:
		baseType = wkbPoint
	case *geom.LineString:
		baseType = wkbLineString
	case *geom.LinearRing:
		baseType = wkbLineString // LinearRing is encoded as LineString
	case *geom.Polygon:
		baseType = wkbPolygon
	case *geom.MultiPoint:
		baseType = wkbMultiPoint
	case *geom.MultiLineString:
		baseType = wkbMultiLineString
	case *geom.MultiPolygon:
		baseType = wkbMultiPolygon
	case *geom.GeometryCollection:
		baseType = wkbGeometryCollection
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
		writeMultiPoint(buf, v, opts)
	case *geom.MultiLineString:
		writeMultiLineString(buf, v, opts)
	case *geom.MultiPolygon:
		writeMultiPolygon(buf, v, opts)
	case *geom.GeometryCollection:
		writeGeometryCollection(buf, v, opts)
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

func writeMultiPoint(buf *buffer, mp *geom.MultiPoint, opts Options) {
	buf.writeUint32(uint32(mp.NumGeometries()))
	for i := 0; i < mp.NumGeometries(); i++ {
		writeGeometry(buf, mp.GeometryN(i), opts)
	}
}

func writeMultiLineString(buf *buffer, mls *geom.MultiLineString, opts Options) {
	buf.writeUint32(uint32(mls.NumGeometries()))
	for i := 0; i < mls.NumGeometries(); i++ {
		writeGeometry(buf, mls.GeometryN(i), opts)
	}
}

func writeMultiPolygon(buf *buffer, mp *geom.MultiPolygon, opts Options) {
	buf.writeUint32(uint32(mp.NumGeometries()))
	for i := 0; i < mp.NumGeometries(); i++ {
		writeGeometry(buf, mp.GeometryN(i), opts)
	}
}

func writeGeometryCollection(buf *buffer, gc *geom.GeometryCollection, opts Options) {
	buf.writeUint32(uint32(gc.NumGeometries()))
	for i := 0; i < gc.NumGeometries(); i++ {
		writeGeometry(buf, gc.GeometryN(i), opts)
	}
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

func (p *parser) readGeometry() (geom.Geometry, error) {
	// Read byte order
	if p.pos >= len(p.data) {
		return nil, fmt.Errorf("unexpected end of data")
	}

	byteOrder := p.data[p.pos]
	p.pos++

	if byteOrder == wkbXDR {
		p.order = binary.BigEndian
	} else if byteOrder == wkbNDR {
		p.order = binary.LittleEndian
	} else {
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
		g.SetSRID(p.srid)
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

	// Read shell
	shellPoints, err := p.readUint32()
	if err != nil {
		return nil, err
	}
	shellCoords, err := p.readCoordinates(int(shellPoints))
	if err != nil {
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
		holeCoords, err := p.readCoordinates(int(holePoints))
		if err != nil {
			return nil, err
		}
		holes[i-1] = p.factory.CreateLinearRing(holeCoords)
	}

	return p.factory.CreatePolygon(shell, holes), nil
}

func (p *parser) readMultiPoint() (*geom.MultiPoint, error) {
	numGeoms, err := p.readUint32()
	if err != nil {
		return nil, err
	}

	if numGeoms == 0 {
		return p.factory.CreateMultiPointEmpty(), nil
	}

	points := make([]*geom.Point, numGeoms)
	for i := uint32(0); i < numGeoms; i++ {
		g, err := p.readGeometry()
		if err != nil {
			return nil, err
		}
		pt, ok := g.(*geom.Point)
		if !ok {
			return nil, fmt.Errorf("expected Point in MultiPoint, got %T", g)
		}
		points[i] = pt
	}

	return p.factory.CreateMultiPoint(points), nil
}

func (p *parser) readMultiLineString() (*geom.MultiLineString, error) {
	numGeoms, err := p.readUint32()
	if err != nil {
		return nil, err
	}

	if numGeoms == 0 {
		return p.factory.CreateMultiLineStringEmpty(), nil
	}

	lines := make([]*geom.LineString, numGeoms)
	for i := uint32(0); i < numGeoms; i++ {
		g, err := p.readGeometry()
		if err != nil {
			return nil, err
		}
		ls, ok := g.(*geom.LineString)
		if !ok {
			return nil, fmt.Errorf("expected LineString in MultiLineString, got %T", g)
		}
		lines[i] = ls
	}

	return p.factory.CreateMultiLineString(lines), nil
}

func (p *parser) readMultiPolygon() (*geom.MultiPolygon, error) {
	numGeoms, err := p.readUint32()
	if err != nil {
		return nil, err
	}

	if numGeoms == 0 {
		return p.factory.CreateMultiPolygonEmpty(), nil
	}

	polys := make([]*geom.Polygon, numGeoms)
	for i := uint32(0); i < numGeoms; i++ {
		g, err := p.readGeometry()
		if err != nil {
			return nil, err
		}
		poly, ok := g.(*geom.Polygon)
		if !ok {
			return nil, fmt.Errorf("expected Polygon in MultiPolygon, got %T", g)
		}
		polys[i] = poly
	}

	return p.factory.CreateMultiPolygon(polys), nil
}

func (p *parser) readGeometryCollection() (*geom.GeometryCollection, error) {
	numGeoms, err := p.readUint32()
	if err != nil {
		return nil, err
	}

	if numGeoms == 0 {
		return p.factory.CreateGeometryCollectionEmpty(), nil
	}

	geoms := make([]geom.Geometry, numGeoms)
	for i := uint32(0); i < numGeoms; i++ {
		g, err := p.readGeometry()
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
		coord.Z = &z
	}

	if p.hasM {
		m, err := p.readFloat64()
		if err != nil {
			return geom.Coordinate{}, err
		}
		coord.M = &m
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
