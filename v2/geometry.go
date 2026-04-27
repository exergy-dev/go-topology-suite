package topology

import (
	"fmt"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

type Geometry = geom.Geometry
type Coordinate = geom.Coordinate
type CoordinateSequence = geom.CoordinateSequence
type GeometryFactory = geom.GeometryFactory

type Point = geom.Point
type LineString = geom.LineString
type LinearRing = geom.LinearRing
type Polygon = geom.Polygon
type MultiPoint = geom.MultiPoint
type MultiLineString = geom.MultiLineString
type MultiPolygon = geom.MultiPolygon
type GeometryCollection = geom.GeometryCollection

type ConstructorOptions struct {
	Factory      *geom.GeometryFactory
	AllowInvalid bool
	Normalize    bool
}

func NewGeometryFactory(srid int) *geom.GeometryFactory {
	return geom.NewGeometryFactoryWithSRID(srid)
}

func NewCoordinate(x, y float64) geom.Coordinate {
	return geom.NewCoordinate(x, y)
}

func NewPoint(x, y float64, opts ...ConstructorOptions) (*geom.Point, error) {
	cfg := constructorOptions(opts...)
	point := cfg.Factory.CreatePoint(x, y)
	if err := validateConstructed(point, cfg); err != nil {
		return nil, err
	}
	if cfg.Normalize {
		point = point.Normalized().(*geom.Point)
	}
	return point, nil
}

func NewPointEmpty(opts ...ConstructorOptions) (*geom.Point, error) {
	cfg := constructorOptions(opts...)
	return cfg.Factory.CreatePointEmpty(), nil
}

func NewLineString(coords geom.CoordinateSequence, opts ...ConstructorOptions) (*geom.LineString, error) {
	cfg := constructorOptions(opts...)
	if coords == nil {
		return nil, fmt.Errorf("v2 geom: linestring coordinates are nil")
	}
	line := cfg.Factory.CreateLineString(coords)
	if err := validateConstructed(line, cfg); err != nil {
		return nil, err
	}
	if cfg.Normalize {
		line = line.Normalized().(*geom.LineString)
	}
	return line, nil
}

func NewLineStringEmpty(opts ...ConstructorOptions) (*geom.LineString, error) {
	cfg := constructorOptions(opts...)
	return cfg.Factory.CreateLineStringEmpty(), nil
}

func NewPolygon(shell geom.CoordinateSequence, holes []geom.CoordinateSequence, opts ...ConstructorOptions) (*geom.Polygon, error) {
	cfg := constructorOptions(opts...)
	if shell == nil {
		return nil, fmt.Errorf("v2 geom: polygon shell coordinates are nil")
	}

	shellRing := cfg.Factory.CreateLinearRing(shell)
	holeRings := make([]*geom.LinearRing, len(holes))
	for i, hole := range holes {
		if hole == nil {
			return nil, fmt.Errorf("v2 geom: polygon hole %d coordinates are nil", i)
		}
		holeRings[i] = cfg.Factory.CreateLinearRing(hole)
	}

	poly := cfg.Factory.CreatePolygon(shellRing, holeRings)
	if err := validateConstructed(poly, cfg); err != nil {
		return nil, err
	}
	if cfg.Normalize {
		poly = poly.Normalized().(*geom.Polygon)
	}
	return poly, nil
}

func NewPolygonEmpty(opts ...ConstructorOptions) (*geom.Polygon, error) {
	cfg := constructorOptions(opts...)
	return cfg.Factory.CreatePolygonEmpty(), nil
}

func NewMultiPoint(points []*geom.Point, opts ...ConstructorOptions) (*geom.MultiPoint, error) {
	cfg := constructorOptions(opts...)
	if points == nil {
		return nil, fmt.Errorf("v2 geom: multipoint points are nil")
	}
	for i, point := range points {
		if point == nil {
			return nil, fmt.Errorf("v2 geom: multipoint point %d is nil", i)
		}
		if !cfg.AllowInvalid {
			if err := Validate(point); err != nil {
				return nil, fmt.Errorf("v2 geom: multipoint point %d: %w", i, err)
			}
		}
	}

	multiPoint := cfg.Factory.CreateMultiPoint(points)
	if err := validateConstructed(multiPoint, cfg); err != nil {
		return nil, err
	}
	if cfg.Normalize {
		multiPoint = multiPoint.Normalized().(*geom.MultiPoint)
	}
	return multiPoint, nil
}

func NewMultiPointFromCoords(coords geom.CoordinateSequence, opts ...ConstructorOptions) (*geom.MultiPoint, error) {
	cfg := constructorOptions(opts...)
	if coords == nil {
		return nil, fmt.Errorf("v2 geom: multipoint coordinates are nil")
	}
	if !cfg.AllowInvalid {
		for i, coord := range coords {
			if coord.IsNaN() {
				return nil, fmt.Errorf("v2 geom: multipoint coordinate %d is invalid", i)
			}
		}
	}

	multiPoint := cfg.Factory.CreateMultiPointFromCoords(coords)
	if err := validateConstructed(multiPoint, cfg); err != nil {
		return nil, err
	}
	if cfg.Normalize {
		multiPoint = multiPoint.Normalized().(*geom.MultiPoint)
	}
	return multiPoint, nil
}

func NewMultiPointEmpty(opts ...ConstructorOptions) (*geom.MultiPoint, error) {
	cfg := constructorOptions(opts...)
	return cfg.Factory.CreateMultiPointEmpty(), nil
}

func NewMultiLineString(lines []*geom.LineString, opts ...ConstructorOptions) (*geom.MultiLineString, error) {
	cfg := constructorOptions(opts...)
	if lines == nil {
		return nil, fmt.Errorf("v2 geom: multilinestring lines are nil")
	}
	for i, line := range lines {
		if line == nil {
			return nil, fmt.Errorf("v2 geom: multilinestring line %d is nil", i)
		}
	}

	multiLine := cfg.Factory.CreateMultiLineString(lines)
	if err := validateConstructed(multiLine, cfg); err != nil {
		return nil, err
	}
	if cfg.Normalize {
		multiLine = multiLine.Normalized().(*geom.MultiLineString)
	}
	return multiLine, nil
}

func NewMultiLineStringEmpty(opts ...ConstructorOptions) (*geom.MultiLineString, error) {
	cfg := constructorOptions(opts...)
	return cfg.Factory.CreateMultiLineStringEmpty(), nil
}

func NewMultiPolygon(polygons []*geom.Polygon, opts ...ConstructorOptions) (*geom.MultiPolygon, error) {
	cfg := constructorOptions(opts...)
	if polygons == nil {
		return nil, fmt.Errorf("v2 geom: multipolygon polygons are nil")
	}
	for i, polygon := range polygons {
		if polygon == nil {
			return nil, fmt.Errorf("v2 geom: multipolygon polygon %d is nil", i)
		}
	}

	multiPolygon := cfg.Factory.CreateMultiPolygon(polygons)
	if err := validateConstructed(multiPolygon, cfg); err != nil {
		return nil, err
	}
	if cfg.Normalize {
		multiPolygon = multiPolygon.Normalized().(*geom.MultiPolygon)
	}
	return multiPolygon, nil
}

func NewMultiPolygonEmpty(opts ...ConstructorOptions) (*geom.MultiPolygon, error) {
	cfg := constructorOptions(opts...)
	return cfg.Factory.CreateMultiPolygonEmpty(), nil
}

func NewGeometryCollection(geometries []geom.Geometry, opts ...ConstructorOptions) (*geom.GeometryCollection, error) {
	cfg := constructorOptions(opts...)
	if geometries == nil {
		return nil, fmt.Errorf("v2 geom: geometrycollection geometries are nil")
	}
	for i, geometry := range geometries {
		if geometry == nil {
			return nil, fmt.Errorf("v2 geom: geometrycollection geometry %d is nil", i)
		}
	}

	collection := cfg.Factory.CreateGeometryCollection(geometries)
	if err := validateConstructed(collection, cfg); err != nil {
		return nil, err
	}
	if cfg.Normalize {
		collection = collection.Normalized().(*geom.GeometryCollection)
	}
	return collection, nil
}

func NewGeometryCollectionEmpty(opts ...ConstructorOptions) (*geom.GeometryCollection, error) {
	cfg := constructorOptions(opts...)
	return cfg.Factory.CreateGeometryCollectionEmpty(), nil
}

func Validate(g geom.Geometry) error {
	if g == nil {
		return fmt.Errorf("v2 geom: geometry is nil")
	}
	if !g.IsValid() {
		return fmt.Errorf("v2 geom: invalid %s", g.GeometryType())
	}
	return nil
}

func constructorOptions(opts ...ConstructorOptions) ConstructorOptions {
	cfg := ConstructorOptions{
		Factory: geom.DefaultFactory,
	}
	if len(opts) > 0 {
		cfg = opts[0]
	}
	if cfg.Factory == nil {
		cfg.Factory = geom.DefaultFactory
	}
	return cfg
}

func validateConstructed(g geom.Geometry, opts ConstructorOptions) error {
	if opts.AllowInvalid {
		return nil
	}
	return Validate(g)
}
