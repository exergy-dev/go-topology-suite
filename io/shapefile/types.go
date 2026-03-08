package shapefile

import (
	"github.com/robert-malhotra/go-topology-suite/geom"
)

// Feature represents a shapefile feature with geometry and DBF attributes.
type Feature struct {
	Index      int
	Geometry   geom.Geometry
	Properties map[string]any // DBF attributes
}

// Field describes a DBF attribute field.
type Field struct {
	Name      string
	Type      FieldType
	Length    int
	Precision int
}

// FieldType represents a DBF field type.
type FieldType byte

const (
	FieldTypeString  FieldType = 'C'
	FieldTypeInteger FieldType = 'N'
	FieldTypeFloat   FieldType = 'F'
	FieldTypeDate    FieldType = 'D'
)

// ShapeType represents the type of geometry stored in a shapefile.
type ShapeType int

// Shapefile geometry type constants as defined by the ESRI Shapefile specification.
const (
	ShapeTypeNull        ShapeType = 0
	ShapeTypePoint       ShapeType = 1
	ShapeTypePolyLine    ShapeType = 3
	ShapeTypePolygon     ShapeType = 5
	ShapeTypeMultiPoint  ShapeType = 8
	ShapeTypePointZ      ShapeType = 11
	ShapeTypePolyLineZ   ShapeType = 13
	ShapeTypePolygonZ    ShapeType = 15
	ShapeTypeMultiPointZ ShapeType = 18
	ShapeTypePointM      ShapeType = 21
	ShapeTypePolyLineM   ShapeType = 23
	ShapeTypePolygonM    ShapeType = 25
	ShapeTypeMultiPointM ShapeType = 28
)

// String returns the string representation of the shape type.
func (st ShapeType) String() string {
	switch st {
	case ShapeTypeNull:
		return "Null"
	case ShapeTypePoint:
		return "Point"
	case ShapeTypePolyLine:
		return "PolyLine"
	case ShapeTypePolygon:
		return "Polygon"
	case ShapeTypeMultiPoint:
		return "MultiPoint"
	case ShapeTypePointZ:
		return "PointZ"
	case ShapeTypePolyLineZ:
		return "PolyLineZ"
	case ShapeTypePolygonZ:
		return "PolygonZ"
	case ShapeTypeMultiPointZ:
		return "MultiPointZ"
	case ShapeTypePointM:
		return "PointM"
	case ShapeTypePolyLineM:
		return "PolyLineM"
	case ShapeTypePolygonM:
		return "PolygonM"
	case ShapeTypeMultiPointM:
		return "MultiPointM"
	default:
		return "Unknown"
	}
}

// Is2D returns true if this shape type is a 2D type.
func (st ShapeType) Is2D() bool {
	return st >= 0 && st <= 8
}

// IsZ returns true if this shape type has Z coordinates.
func (st ShapeType) IsZ() bool {
	return st >= 11 && st <= 18
}

// IsM returns true if this shape type has M coordinates.
func (st ShapeType) IsM() bool {
	return st >= 21 && st <= 28
}

// GeometryToShapeType determines the appropriate shapefile shape type for a geometry.
// Returns ShapeTypeNull for nil or unsupported geometry types.
func GeometryToShapeType(g geom.Geometry) ShapeType {
	if g == nil || g.IsEmpty() {
		return ShapeTypeNull
	}

	hasZ := g.Coordinates().HasZ()

	switch g.(type) {
	case *geom.Point:
		if hasZ {
			return ShapeTypePointZ
		}
		return ShapeTypePoint
	case *geom.LineString, *geom.LinearRing:
		if hasZ {
			return ShapeTypePolyLineZ
		}
		return ShapeTypePolyLine
	case *geom.MultiLineString:
		if hasZ {
			return ShapeTypePolyLineZ
		}
		return ShapeTypePolyLine
	case *geom.Polygon:
		if hasZ {
			return ShapeTypePolygonZ
		}
		return ShapeTypePolygon
	case *geom.MultiPolygon:
		if hasZ {
			return ShapeTypePolygonZ
		}
		return ShapeTypePolygon
	case *geom.MultiPoint:
		if hasZ {
			return ShapeTypeMultiPointZ
		}
		return ShapeTypeMultiPoint
	default:
		return ShapeTypeNull
	}
}

// InferShapeType determines the best shape type for a collection of geometries.
// Returns ShapeTypeNull if the geometries are not homogeneous or the slice is empty.
func InferShapeType(geometries []geom.Geometry) ShapeType {
	if len(geometries) == 0 {
		return ShapeTypeNull
	}

	// Find first non-nil, non-empty geometry
	firstType := ShapeTypeNull
	for _, g := range geometries {
		if g != nil && !g.IsEmpty() {
			firstType = GeometryToShapeType(g)
			break
		}
	}

	if firstType == ShapeTypeNull {
		return ShapeTypeNull
	}

	// Check if all geometries are compatible with the first type
	for _, g := range geometries {
		if g == nil || g.IsEmpty() {
			continue
		}
		gType := GeometryToShapeType(g)
		if !areCompatibleTypes(firstType, gType) {
			return ShapeTypeNull
		}
	}

	return firstType
}

// areCompatibleTypes checks if two shape types are compatible for the same shapefile.
func areCompatibleTypes(t1, t2 ShapeType) bool {
	// Exact match
	if t1 == t2 {
		return true
	}

	// Point types are compatible
	if (t1 == ShapeTypePoint || t1 == ShapeTypePointZ || t1 == ShapeTypePointM) &&
		(t2 == ShapeTypePoint || t2 == ShapeTypePointZ || t2 == ShapeTypePointM) {
		return true
	}

	// PolyLine types are compatible
	if (t1 == ShapeTypePolyLine || t1 == ShapeTypePolyLineZ || t1 == ShapeTypePolyLineM) &&
		(t2 == ShapeTypePolyLine || t2 == ShapeTypePolyLineZ || t2 == ShapeTypePolyLineM) {
		return true
	}

	// Polygon types are compatible
	if (t1 == ShapeTypePolygon || t1 == ShapeTypePolygonZ || t1 == ShapeTypePolygonM) &&
		(t2 == ShapeTypePolygon || t2 == ShapeTypePolygonZ || t2 == ShapeTypePolygonM) {
		return true
	}

	// MultiPoint types are compatible
	if (t1 == ShapeTypeMultiPoint || t1 == ShapeTypeMultiPointZ || t1 == ShapeTypeMultiPointM) &&
		(t2 == ShapeTypeMultiPoint || t2 == ShapeTypeMultiPointZ || t2 == ShapeTypeMultiPointM) {
		return true
	}

	return false
}
