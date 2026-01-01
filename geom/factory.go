package geom

// GeometryFactory creates geometry objects with a specific precision model and SRID.
type GeometryFactory struct {
	precisionModel PrecisionModel
	srid           int
}

// NewGeometryFactory creates a new factory with the given precision model and SRID.
func NewGeometryFactory(pm PrecisionModel, srid int) *GeometryFactory {
	if pm == nil {
		pm = Floating
	}
	return &GeometryFactory{
		precisionModel: pm,
		srid:           srid,
	}
}

// NewGeometryFactoryDefault creates a factory with floating precision and no SRID.
func NewGeometryFactoryDefault() *GeometryFactory {
	return NewGeometryFactory(Floating, 0)
}

// NewGeometryFactoryWithPrecision creates a factory with specified precision.
func NewGeometryFactoryWithPrecision(pm PrecisionModel) *GeometryFactory {
	return NewGeometryFactory(pm, 0)
}

// NewGeometryFactoryWithSRID creates a factory with specified SRID.
func NewGeometryFactoryWithSRID(srid int) *GeometryFactory {
	return NewGeometryFactory(Floating, srid)
}

// PrecisionModel returns the factory's precision model.
func (gf *GeometryFactory) PrecisionModel() PrecisionModel {
	return gf.precisionModel
}

// SRID returns the factory's SRID.
func (gf *GeometryFactory) SRID() int {
	return gf.srid
}

// CreatePoint creates a Point from x,y coordinates.
func (gf *GeometryFactory) CreatePoint(x, y float64) *Point {
	coord := NewCoordinate(x, y)
	gf.precisionModel.MakePrecise(&coord)
	p := NewPointFromCoordinate(coord)
	p.srid = gf.srid
	return p
}

// CreatePointFromCoordinate creates a Point from a Coordinate.
func (gf *GeometryFactory) CreatePointFromCoordinate(coord Coordinate) *Point {
	c := coord.Clone()
	gf.precisionModel.MakePrecise(&c)
	p := NewPointFromCoordinate(c)
	p.srid = gf.srid
	return p
}

// CreatePointEmpty creates an empty Point.
func (gf *GeometryFactory) CreatePointEmpty() *Point {
	p := NewPointEmpty()
	p.srid = gf.srid
	return p
}

// CreateLineString creates a LineString from coordinates.
func (gf *GeometryFactory) CreateLineString(coords CoordinateSequence) *LineString {
	c := coords.Clone()
	MakePreciseSequence(gf.precisionModel, c)
	ls := NewLineString(c)
	ls.srid = gf.srid
	return ls
}

// CreateLineStringXY creates a LineString from x,y pairs.
func (gf *GeometryFactory) CreateLineStringXY(values ...float64) *LineString {
	return gf.CreateLineString(NewCoordinateSequenceXY(values...))
}

// CreateLineStringEmpty creates an empty LineString.
func (gf *GeometryFactory) CreateLineStringEmpty() *LineString {
	ls := NewLineStringEmpty()
	ls.srid = gf.srid
	return ls
}

// CreateLinearRing creates a LinearRing from coordinates.
func (gf *GeometryFactory) CreateLinearRing(coords CoordinateSequence) *LinearRing {
	c := coords.Clone()
	MakePreciseSequence(gf.precisionModel, c)
	lr := NewLinearRing(c)
	lr.srid = gf.srid
	return lr
}

// CreateLinearRingXY creates a LinearRing from x,y pairs.
func (gf *GeometryFactory) CreateLinearRingXY(values ...float64) *LinearRing {
	return gf.CreateLinearRing(NewCoordinateSequenceXY(values...))
}

// CreateLinearRingEmpty creates an empty LinearRing.
func (gf *GeometryFactory) CreateLinearRingEmpty() *LinearRing {
	lr := NewLinearRingEmpty()
	lr.srid = gf.srid
	return lr
}

// CreatePolygon creates a Polygon from a shell and holes.
func (gf *GeometryFactory) CreatePolygon(shell *LinearRing, holes []*LinearRing) *Polygon {
	// Apply precision to shell
	shellCoords := shell.coords.Clone()
	MakePreciseSequence(gf.precisionModel, shellCoords)
	preciseShell := NewLinearRing(shellCoords)
	preciseShell.srid = gf.srid

	// Apply precision to holes
	preciseHoles := make([]*LinearRing, len(holes))
	for i, hole := range holes {
		holeCoords := hole.coords.Clone()
		MakePreciseSequence(gf.precisionModel, holeCoords)
		preciseHoles[i] = NewLinearRing(holeCoords)
		preciseHoles[i].srid = gf.srid
	}

	p := NewPolygon(preciseShell, preciseHoles)
	p.srid = gf.srid
	return p
}

// CreatePolygonFromCoords creates a Polygon from coordinate sequences.
func (gf *GeometryFactory) CreatePolygonFromCoords(shell CoordinateSequence, holes ...CoordinateSequence) *Polygon {
	shellRing := gf.CreateLinearRing(shell)
	holeRings := make([]*LinearRing, len(holes))
	for i, h := range holes {
		holeRings[i] = gf.CreateLinearRing(h)
	}
	p := NewPolygon(shellRing, holeRings)
	p.srid = gf.srid
	return p
}

// CreatePolygonEmpty creates an empty Polygon.
func (gf *GeometryFactory) CreatePolygonEmpty() *Polygon {
	p := NewPolygonEmpty()
	p.srid = gf.srid
	return p
}

// CreateMultiPoint creates a MultiPoint from points.
func (gf *GeometryFactory) CreateMultiPoint(points []*Point) *MultiPoint {
	precisePoints := make([]*Point, len(points))
	for i, p := range points {
		precisePoints[i] = gf.CreatePointFromCoordinate(p.coord)
	}
	mp := &MultiPoint{points: precisePoints}
	mp.srid = gf.srid
	return mp
}

// CreateMultiPointFromCoords creates a MultiPoint from coordinates.
func (gf *GeometryFactory) CreateMultiPointFromCoords(coords CoordinateSequence) *MultiPoint {
	points := make([]*Point, len(coords))
	for i, c := range coords {
		points[i] = gf.CreatePointFromCoordinate(c)
	}
	mp := &MultiPoint{points: points}
	mp.srid = gf.srid
	return mp
}

// CreateMultiPointEmpty creates an empty MultiPoint.
func (gf *GeometryFactory) CreateMultiPointEmpty() *MultiPoint {
	mp := NewMultiPointEmpty()
	mp.srid = gf.srid
	return mp
}

// CreateMultiLineString creates a MultiLineString from linestrings.
func (gf *GeometryFactory) CreateMultiLineString(lines []*LineString) *MultiLineString {
	preciseLines := make([]*LineString, len(lines))
	for i, l := range lines {
		preciseLines[i] = gf.CreateLineString(l.coords)
	}
	mls := &MultiLineString{lines: preciseLines}
	mls.srid = gf.srid
	return mls
}

// CreateMultiLineStringEmpty creates an empty MultiLineString.
func (gf *GeometryFactory) CreateMultiLineStringEmpty() *MultiLineString {
	mls := NewMultiLineStringEmpty()
	mls.srid = gf.srid
	return mls
}

// CreateMultiPolygon creates a MultiPolygon from polygons.
func (gf *GeometryFactory) CreateMultiPolygon(polygons []*Polygon) *MultiPolygon {
	precisePolygons := make([]*Polygon, len(polygons))
	for i, p := range polygons {
		precisePolygons[i] = gf.CreatePolygon(p.shell, p.holes)
	}
	mp := &MultiPolygon{polygons: precisePolygons}
	mp.srid = gf.srid
	return mp
}

// CreateMultiPolygonEmpty creates an empty MultiPolygon.
func (gf *GeometryFactory) CreateMultiPolygonEmpty() *MultiPolygon {
	mp := NewMultiPolygonEmpty()
	mp.srid = gf.srid
	return mp
}

// CreateGeometryCollection creates a GeometryCollection.
func (gf *GeometryFactory) CreateGeometryCollection(geometries []Geometry) *GeometryCollection {
	gc := NewGeometryCollection(geometries)
	gc.srid = gf.srid
	return gc
}

// CreateGeometryCollectionEmpty creates an empty GeometryCollection.
func (gf *GeometryFactory) CreateGeometryCollectionEmpty() *GeometryCollection {
	gc := NewGeometryCollectionEmpty()
	gc.srid = gf.srid
	return gc
}

// ToGeometry converts a slice of coordinates to an appropriate geometry.
// 0 coords -> empty point
// 1 coord -> point
// 2 coords -> linestring
// 3+ coords (unclosed) -> linestring
// 4+ coords (closed) -> polygon
func (gf *GeometryFactory) ToGeometry(coords CoordinateSequence) Geometry {
	switch len(coords) {
	case 0:
		return gf.CreatePointEmpty()
	case 1:
		return gf.CreatePointFromCoordinate(coords[0])
	default:
		if len(coords) >= 4 && coords.IsClosed(DefaultEpsilon) {
			return gf.CreatePolygonFromCoords(coords)
		}
		return gf.CreateLineString(coords)
	}
}

// BuildGeometry constructs a geometry from a collection of geometries.
// If all are same type, returns multi version. Otherwise GeometryCollection.
func (gf *GeometryFactory) BuildGeometry(geometries []Geometry) Geometry {
	if len(geometries) == 0 {
		return gf.CreateGeometryCollectionEmpty()
	}

	if len(geometries) == 1 {
		return geometries[0].Clone()
	}

	// Check if all same type
	firstType := geometries[0].GeometryType()
	allSame := true
	for _, g := range geometries[1:] {
		if g.GeometryType() != firstType {
			allSame = false
			break
		}
	}

	if allSame {
		switch firstType {
		case "Point":
			points := make([]*Point, len(geometries))
			for i, g := range geometries {
				points[i] = g.(*Point)
			}
			return gf.CreateMultiPoint(points)
		case "LineString":
			lines := make([]*LineString, len(geometries))
			for i, g := range geometries {
				lines[i] = g.(*LineString)
			}
			return gf.CreateMultiLineString(lines)
		case "Polygon":
			polys := make([]*Polygon, len(geometries))
			for i, g := range geometries {
				polys[i] = g.(*Polygon)
			}
			return gf.CreateMultiPolygon(polys)
		}
	}

	return gf.CreateGeometryCollection(geometries)
}

// Default factory with floating precision and no SRID.
var DefaultFactory = NewGeometryFactoryDefault()
