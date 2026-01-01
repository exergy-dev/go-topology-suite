package shapefile

import (
	"fmt"
	"iter"
	"strings"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/jonas-p/go-shp"
)

// Reader reads geometries from a shapefile.
type Reader struct {
	reader    *shp.Reader
	factory   *geom.GeometryFactory
	shapeType ShapeType
	current   int
	err       error
}

// NewReader creates a new shapefile reader.
func NewReader(filename string) (*Reader, error) {
	return NewReaderWithFactory(filename, geom.DefaultFactory)
}

// NewReaderWithFactory creates a new shapefile reader with a custom geometry factory.
func NewReaderWithFactory(filename string, factory *geom.GeometryFactory) (*Reader, error) {
	reader, err := shp.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open shapefile: %w", err)
	}

	return &Reader{
		reader:    reader,
		factory:   factory,
		shapeType: ShapeType(reader.GeometryType),
		current:   -1,
	}, nil
}

// Next advances to the next record. Returns false when there are no more records.
func (r *Reader) Next() bool {
	if r.reader.Next() {
		r.current++
		return true
	}
	return false
}

// Geometry returns the current geometry.
func (r *Reader) Geometry() (geom.Geometry, error) {
	_, shape := r.reader.Shape()
	if shape == nil {
		return nil, fmt.Errorf("no shape at current position")
	}

	return shapeToGeometry(shape, r.factory)
}

// ShapeType returns the shapefile's geometry type.
func (r *Reader) ShapeType() ShapeType {
	return r.shapeType
}

// BoundingBox returns the bounding box of all geometries in the shapefile.
func (r *Reader) BoundingBox() *geom.Envelope {
	bbox := r.reader.BBox()
	return geom.NewEnvelope(bbox.MinX, bbox.MinY, bbox.MaxX, bbox.MaxY)
}

// Close closes the shapefile reader.
func (r *Reader) Close() error {
	return r.reader.Close()
}

// Err returns any error that occurred during iteration.
func (r *Reader) Err() error {
	return r.err
}

// Fields returns the DBF field names.
func (r *Reader) Fields() []string {
	shpFields := r.reader.Fields()
	names := make([]string, len(shpFields))
	for i, f := range shpFields {
		names[i] = strings.TrimRight(string(f.Name[:]), "\x00")
	}
	return names
}

// Feature returns the current feature with geometry and DBF attributes.
func (r *Reader) Feature() (*Feature, error) {
	g, err := r.Geometry()
	if err != nil {
		return nil, err
	}

	// Build properties map from DBF attributes
	shpFields := r.reader.Fields()
	props := make(map[string]any, len(shpFields))
	for i, f := range shpFields {
		name := strings.TrimRight(string(f.Name[:]), "\x00")
		value := r.reader.ReadAttribute(r.current, i)
		// DBF values are padded with spaces or null bytes
		value = strings.TrimRight(value, " \x00")
		props[name] = value
	}

	return &Feature{
		Index:      r.current,
		Geometry:   g,
		Properties: props,
	}, nil
}

// ReadAll reads all geometries from a shapefile.
func ReadAll(filename string) ([]geom.Geometry, error) {
	return ReadAllWithFactory(filename, geom.DefaultFactory)
}

// ReadAllWithFactory reads all geometries from a shapefile using a custom factory.
func ReadAllWithFactory(filename string, factory *geom.GeometryFactory) ([]geom.Geometry, error) {
	reader, err := NewReaderWithFactory(filename, factory)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var geometries []geom.Geometry
	for reader.Next() {
		g, err := reader.Geometry()
		if err != nil {
			return nil, fmt.Errorf("error reading geometry at index %d: %w", reader.current, err)
		}
		geometries = append(geometries, g)
	}

	return geometries, nil
}

// Features returns an iterator over all features in a shapefile.
// Usage: for f, err := range shapefile.Features("file.shp") { ... }
func Features(filename string) iter.Seq2[*Feature, error] {
	return FeaturesWithFactory(filename, geom.DefaultFactory)
}

// FeaturesWithFactory returns an iterator using a custom geometry factory.
func FeaturesWithFactory(filename string, factory *geom.GeometryFactory) iter.Seq2[*Feature, error] {
	return func(yield func(*Feature, error) bool) {
		reader, err := NewReaderWithFactory(filename, factory)
		if err != nil {
			yield(nil, err)
			return
		}
		defer reader.Close()

		for reader.Next() {
			f, err := reader.Feature()
			if !yield(f, err) {
				return
			}
			if err != nil {
				return
			}
		}
	}
}

// Writer writes geometries to a shapefile.
type Writer struct {
	writer    *shp.Writer
	shapeType ShapeType
	count     int
	fields    []Field
}

// NewWriter creates a new shapefile writer.
func NewWriter(filename string, shapeType ShapeType) (*Writer, error) {
	writer, err := shp.Create(filename, shp.ShapeType(shapeType))
	if err != nil {
		return nil, fmt.Errorf("failed to create shapefile: %w", err)
	}

	return &Writer{
		writer:    writer,
		shapeType: shapeType,
		count:     0,
	}, nil
}

// Write writes a geometry to the shapefile.
func (w *Writer) Write(g geom.Geometry) error {
	if g == nil || g.IsEmpty() {
		// Write null shape for empty geometries
		w.writer.Write(&shp.Null{})
		w.count++
		return nil
	}

	shape, err := geometryToShape(g, w.shapeType)
	if err != nil {
		return fmt.Errorf("failed to convert geometry: %w", err)
	}

	w.writer.Write(shape)
	w.count++
	return nil
}

// Close closes the shapefile writer.
func (w *Writer) Close() error {
	w.writer.Close()
	return nil
}

// Count returns the number of geometries written.
func (w *Writer) Count() int {
	return w.count
}

// SetFields sets the DBF field definitions for writing attributes.
func (w *Writer) SetFields(fields []Field) error {
	shpFields := make([]shp.Field, len(fields))
	for i, f := range fields {
		switch f.Type {
		case FieldTypeString:
			shpFields[i] = shp.StringField(f.Name, uint8(f.Length))
		case FieldTypeInteger:
			shpFields[i] = shp.NumberField(f.Name, uint8(f.Length))
		case FieldTypeFloat:
			shpFields[i] = shp.FloatField(f.Name, uint8(f.Length), uint8(f.Precision))
		case FieldTypeDate:
			shpFields[i] = shp.DateField(f.Name)
		default:
			shpFields[i] = shp.StringField(f.Name, uint8(f.Length))
		}
	}
	w.fields = fields
	return w.writer.SetFields(shpFields)
}

// WriteFeature writes a feature with geometry and attributes.
func (w *Writer) WriteFeature(f *Feature) error {
	// Write geometry
	if err := w.Write(f.Geometry); err != nil {
		return err
	}

	// Write attributes (count was already incremented by Write, so use count-1 as row index)
	row := w.count - 1
	for i, field := range w.fields {
		if val, ok := f.Properties[field.Name]; ok {
			if err := w.writer.WriteAttribute(row, i, val); err != nil {
				return fmt.Errorf("failed to write attribute %s: %w", field.Name, err)
			}
		}
	}

	return nil
}

// WriteAll writes all geometries to a shapefile.
// The shape type is inferred from the geometries.
func WriteAll(filename string, geometries []geom.Geometry) error {
	if len(geometries) == 0 {
		return fmt.Errorf("cannot write empty geometry slice")
	}

	shapeType := InferShapeType(geometries)
	if shapeType == ShapeTypeNull {
		return fmt.Errorf("cannot infer consistent shape type from geometries")
	}

	writer, err := NewWriter(filename, shapeType)
	if err != nil {
		return err
	}
	defer writer.Close()

	for i, g := range geometries {
		if err := writer.Write(g); err != nil {
			return fmt.Errorf("error writing geometry at index %d: %w", i, err)
		}
	}

	return nil
}

// shapeToGeometry converts a go-shp shape to a GTS geometry.
func shapeToGeometry(shape shp.Shape, factory *geom.GeometryFactory) (geom.Geometry, error) {
	switch s := shape.(type) {
	case *shp.Point:
		return factory.CreatePoint(s.X, s.Y), nil

	case *shp.PointZ:
		coord := geom.NewCoordinateZ(s.X, s.Y, s.Z)
		return factory.CreatePointFromCoordinate(coord), nil

	case *shp.PointM:
		coord := geom.NewCoordinateM(s.X, s.Y, s.M)
		return factory.CreatePointFromCoordinate(coord), nil

	case *shp.PolyLine:
		return polyLineToGeometry(s.Points, s.Parts, false, nil, factory)

	case *shp.PolyLineZ:
		return polyLineToGeometry(s.Points, s.Parts, true, s.ZArray, factory)

	case *shp.PolyLineM:
		return polyLineToGeometry(s.Points, s.Parts, false, nil, factory)

	case *shp.Polygon:
		return polygonToGeometry(s.Points, s.Parts, false, nil, factory)

	case *shp.PolygonZ:
		return polygonToGeometry(s.Points, s.Parts, true, s.ZArray, factory)

	case *shp.PolygonM:
		return polygonToGeometry(s.Points, s.Parts, false, nil, factory)

	case *shp.MultiPoint:
		return multiPointToGeometry(s.Points, false, nil, factory)

	case *shp.MultiPointZ:
		return multiPointToGeometry(s.Points, true, s.ZArray, factory)

	case *shp.MultiPointM:
		return multiPointToGeometry(s.Points, false, nil, factory)

	case *shp.Null:
		return factory.CreatePointEmpty(), nil

	default:
		return nil, fmt.Errorf("unsupported shape type: %T", shape)
	}
}

// polyLineToGeometry converts polyline points and parts to a LineString or MultiLineString.
func polyLineToGeometry(points []shp.Point, parts []int32, hasZ bool, zValues []float64, factory *geom.GeometryFactory) (geom.Geometry, error) {
	if len(points) == 0 {
		return factory.CreateLineStringEmpty(), nil
	}

	if len(parts) == 0 {
		parts = []int32{0}
	}

	lines := make([]*geom.LineString, len(parts))

	for i, partStart := range parts {
		var partEnd int
		if i+1 < len(parts) {
			partEnd = int(parts[i+1])
		} else {
			partEnd = len(points)
		}

		coords := make(geom.CoordinateSequence, partEnd-int(partStart))
		for j := int(partStart); j < partEnd; j++ {
			idx := j - int(partStart)
			if hasZ && j < len(zValues) {
				coords[idx] = geom.NewCoordinateZ(points[j].X, points[j].Y, zValues[j])
			} else {
				coords[idx] = geom.NewCoordinate(points[j].X, points[j].Y)
			}
		}

		lines[i] = factory.CreateLineString(coords)
	}

	if len(lines) == 1 {
		return lines[0], nil
	}

	return factory.CreateMultiLineString(lines), nil
}

// polygonToGeometry converts polygon points and parts to a Polygon or MultiPolygon.
func polygonToGeometry(points []shp.Point, parts []int32, hasZ bool, zValues []float64, factory *geom.GeometryFactory) (geom.Geometry, error) {
	if len(points) == 0 {
		return factory.CreatePolygonEmpty(), nil
	}

	if len(parts) == 0 {
		parts = []int32{0}
	}

	// Build rings from parts
	rings := make([]*geom.LinearRing, len(parts))

	for i, partStart := range parts {
		var partEnd int
		if i+1 < len(parts) {
			partEnd = int(parts[i+1])
		} else {
			partEnd = len(points)
		}

		coords := make(geom.CoordinateSequence, partEnd-int(partStart))
		for j := int(partStart); j < partEnd; j++ {
			idx := j - int(partStart)
			if hasZ && j < len(zValues) {
				coords[idx] = geom.NewCoordinateZ(points[j].X, points[j].Y, zValues[j])
			} else {
				coords[idx] = geom.NewCoordinate(points[j].X, points[j].Y)
			}
		}

		rings[i] = factory.CreateLinearRing(coords)
	}

	// Determine which rings are shells and which are holes
	// In shapefiles, clockwise rings are exterior (shells), counter-clockwise are holes
	// Note: This is opposite of OGC convention
	return buildPolygonsFromRings(rings, factory)
}

// buildPolygonsFromRings organizes rings into polygons based on orientation and containment.
func buildPolygonsFromRings(rings []*geom.LinearRing, factory *geom.GeometryFactory) (geom.Geometry, error) {
	if len(rings) == 0 {
		return factory.CreatePolygonEmpty(), nil
	}

	if len(rings) == 1 {
		return factory.CreatePolygon(rings[0], nil), nil
	}

	// Identify shells and holes based on orientation
	// In shapefiles: clockwise = exterior, counter-clockwise = interior
	var shells []*geom.LinearRing
	var holes []*geom.LinearRing

	for _, ring := range rings {
		if ring.IsCW() {
			shells = append(shells, ring)
		} else {
			holes = append(holes, ring)
		}
	}

	// If no shells found, treat all as shells (fallback)
	if len(shells) == 0 {
		shells = rings
		holes = nil
	}

	// Single shell case - all holes belong to it
	if len(shells) == 1 {
		return factory.CreatePolygon(shells[0], holes), nil
	}

	// Multiple shells - build MultiPolygon
	// Associate holes with shells based on containment
	polygons := make([]*geom.Polygon, len(shells))
	shellHoles := make([][]*geom.LinearRing, len(shells))

	for _, hole := range holes {
		holeCentroid := hole.Centroid()
		if holeCentroid.IsEmpty() {
			continue
		}
		holeCoord := holeCentroid.Coordinate()

		// Find the smallest shell that contains this hole
		minArea := -1.0
		shellIdx := -1

		for i, shell := range shells {
			shellPoly := geom.NewPolygon(shell, nil)
			if shellPoly.ContainsPoint(holeCoord) {
				area := shellPoly.Area()
				if shellIdx == -1 || area < minArea {
					minArea = area
					shellIdx = i
				}
			}
		}

		if shellIdx >= 0 {
			shellHoles[shellIdx] = append(shellHoles[shellIdx], hole)
		}
	}

	for i, shell := range shells {
		polygons[i] = factory.CreatePolygon(shell, shellHoles[i])
	}

	return factory.CreateMultiPolygon(polygons), nil
}

// multiPointToGeometry converts multipoint data to a MultiPoint geometry.
func multiPointToGeometry(points []shp.Point, hasZ bool, zValues []float64, factory *geom.GeometryFactory) (geom.Geometry, error) {
	if len(points) == 0 {
		return factory.CreateMultiPointEmpty(), nil
	}

	pts := make([]*geom.Point, len(points))
	for i, p := range points {
		if hasZ && i < len(zValues) {
			coord := geom.NewCoordinateZ(p.X, p.Y, zValues[i])
			pts[i] = factory.CreatePointFromCoordinate(coord)
		} else {
			pts[i] = factory.CreatePoint(p.X, p.Y)
		}
	}

	return factory.CreateMultiPoint(pts), nil
}

// geometryToShape converts a GTS geometry to a go-shp shape.
func geometryToShape(g geom.Geometry, shapeType ShapeType) (shp.Shape, error) {
	if g == nil || g.IsEmpty() {
		return &shp.Null{}, nil
	}

	switch shapeType {
	case ShapeTypePoint, ShapeTypePointZ, ShapeTypePointM:
		return geometryToPoint(g, shapeType)
	case ShapeTypePolyLine, ShapeTypePolyLineZ, ShapeTypePolyLineM:
		return geometryToPolyLine(g, shapeType)
	case ShapeTypePolygon, ShapeTypePolygonZ, ShapeTypePolygonM:
		return geometryToPolygon(g, shapeType)
	case ShapeTypeMultiPoint, ShapeTypeMultiPointZ, ShapeTypeMultiPointM:
		return geometryToMultiPoint(g, shapeType)
	default:
		return nil, fmt.Errorf("unsupported shape type: %v", shapeType)
	}
}

// geometryToPoint converts a Point geometry to a shp.Point.
func geometryToPoint(g geom.Geometry, shapeType ShapeType) (shp.Shape, error) {
	var coord geom.Coordinate

	switch pt := g.(type) {
	case *geom.Point:
		if pt.IsEmpty() {
			return &shp.Null{}, nil
		}
		coord = pt.Coordinate()
	default:
		return nil, fmt.Errorf("expected Point, got %T", g)
	}

	switch shapeType {
	case ShapeTypePointZ:
		return &shp.PointZ{
			X: coord.X,
			Y: coord.Y,
			Z: coord.GetZ(),
		}, nil
	case ShapeTypePointM:
		return &shp.PointM{
			X: coord.X,
			Y: coord.Y,
			M: coord.GetM(),
		}, nil
	default:
		return &shp.Point{
			X: coord.X,
			Y: coord.Y,
		}, nil
	}
}

// geometryToPolyLine converts a LineString or MultiLineString to a shp.PolyLine.
func geometryToPolyLine(g geom.Geometry, shapeType ShapeType) (shp.Shape, error) {
	var lines []*geom.LineString

	switch ls := g.(type) {
	case *geom.LineString:
		lines = []*geom.LineString{ls}
	case *geom.LinearRing:
		lines = []*geom.LineString{ls.LineString}
	case *geom.MultiLineString:
		lines = make([]*geom.LineString, ls.NumGeometries())
		for i := 0; i < ls.NumGeometries(); i++ {
			lines[i] = ls.LineStringN(i)
		}
	default:
		return nil, fmt.Errorf("expected LineString or MultiLineString, got %T", g)
	}

	// Build parts as [][]shp.Point for NewPolyLine constructor
	parts := make([][]shp.Point, len(lines))
	var zValues []float64
	zIdx := 0

	for i, line := range lines {
		coords := line.Coordinates()
		parts[i] = make([]shp.Point, len(coords))
		for j, c := range coords {
			parts[i][j] = shp.Point{X: c.X, Y: c.Y}
			if shapeType.IsZ() {
				zValues = append(zValues, c.GetZ())
				zIdx++
			}
		}
	}

	switch shapeType {
	case ShapeTypePolyLineZ:
		pl := shp.NewPolyLine(parts)
		return &shp.PolyLineZ{
			Box:       pl.Box,
			NumParts:  pl.NumParts,
			NumPoints: pl.NumPoints,
			Parts:     pl.Parts,
			Points:    pl.Points,
			ZArray:    zValues,
		}, nil
	case ShapeTypePolyLineM:
		pl := shp.NewPolyLine(parts)
		return &shp.PolyLineM{
			Box:       pl.Box,
			NumParts:  pl.NumParts,
			NumPoints: pl.NumPoints,
			Parts:     pl.Parts,
			Points:    pl.Points,
		}, nil
	default:
		return shp.NewPolyLine(parts), nil
	}
}

// geometryToPolygon converts a Polygon or MultiPolygon to a shp.Polygon.
func geometryToPolygon(g geom.Geometry, shapeType ShapeType) (shp.Shape, error) {
	var polygons []*geom.Polygon

	switch p := g.(type) {
	case *geom.Polygon:
		polygons = []*geom.Polygon{p}
	case *geom.MultiPolygon:
		polygons = make([]*geom.Polygon, p.NumGeometries())
		for i := 0; i < p.NumGeometries(); i++ {
			polygons[i] = p.PolygonN(i)
		}
	default:
		return nil, fmt.Errorf("expected Polygon or MultiPolygon, got %T", g)
	}

	// Collect all rings (shells and holes) as [][]shp.Point
	var allParts [][]shp.Point
	var zValues []float64

	for _, poly := range polygons {
		// Add exterior ring (ensure clockwise for shapefile)
		shell := poly.ExteriorRing()
		if shell.IsCCW() {
			shell = shell.Reverse()
		}
		shellCoords := shell.Coordinates()
		shellPart := make([]shp.Point, len(shellCoords))
		for j, c := range shellCoords {
			shellPart[j] = shp.Point{X: c.X, Y: c.Y}
			if shapeType.IsZ() {
				zValues = append(zValues, c.GetZ())
			}
		}
		allParts = append(allParts, shellPart)

		// Add hole rings (ensure counter-clockwise for shapefile)
		for i := 0; i < poly.NumInteriorRings(); i++ {
			hole := poly.InteriorRingN(i)
			if hole.IsCW() {
				hole = hole.Reverse()
			}
			holeCoords := hole.Coordinates()
			holePart := make([]shp.Point, len(holeCoords))
			for j, c := range holeCoords {
				holePart[j] = shp.Point{X: c.X, Y: c.Y}
				if shapeType.IsZ() {
					zValues = append(zValues, c.GetZ())
				}
			}
			allParts = append(allParts, holePart)
		}
	}

	// Use NewPolyLine to build the base structure (Polygon uses same structure)
	pl := shp.NewPolyLine(allParts)

	switch shapeType {
	case ShapeTypePolygonZ:
		return &shp.PolygonZ{
			Box:       pl.Box,
			NumParts:  pl.NumParts,
			NumPoints: pl.NumPoints,
			Parts:     pl.Parts,
			Points:    pl.Points,
			ZArray:    zValues,
		}, nil
	case ShapeTypePolygonM:
		return &shp.PolygonM{
			Box:       pl.Box,
			NumParts:  pl.NumParts,
			NumPoints: pl.NumPoints,
			Parts:     pl.Parts,
			Points:    pl.Points,
		}, nil
	default:
		// Convert PolyLine to Polygon (same structure, type alias)
		return (*shp.Polygon)(pl), nil
	}
}

// geometryToMultiPoint converts a MultiPoint geometry to a shp.MultiPoint.
func geometryToMultiPoint(g geom.Geometry, shapeType ShapeType) (shp.Shape, error) {
	var mp *geom.MultiPoint

	switch m := g.(type) {
	case *geom.MultiPoint:
		mp = m
	case *geom.Point:
		// Single point - wrap in MultiPoint
		mp = geom.NewMultiPoint([]*geom.Point{m})
	default:
		return nil, fmt.Errorf("expected MultiPoint, got %T", g)
	}

	numPoints := int32(mp.NumGeometries())
	points := make([]shp.Point, numPoints)
	var zValues []float64
	if shapeType.IsZ() {
		zValues = make([]float64, numPoints)
	}

	for i := 0; i < int(numPoints); i++ {
		pt := mp.PointN(i)
		coord := pt.Coordinate()
		points[i] = shp.Point{X: coord.X, Y: coord.Y}
		if zValues != nil {
			zValues[i] = coord.GetZ()
		}
	}

	// Calculate bounding box
	box := shp.BBoxFromPoints(points)

	switch shapeType {
	case ShapeTypeMultiPointZ:
		return &shp.MultiPointZ{
			Box:       box,
			NumPoints: numPoints,
			Points:    points,
			ZArray:    zValues,
		}, nil
	case ShapeTypeMultiPointM:
		return &shp.MultiPointM{
			Box:       box,
			NumPoints: numPoints,
			Points:    points,
		}, nil
	default:
		return &shp.MultiPoint{
			Box:       box,
			NumPoints: numPoints,
			Points:    points,
		}, nil
	}
}
