package kml

import (
	"encoding/xml"
	"fmt"
	"iter"
	"strconv"
	"strings"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/io/ioutil"
)

// SRID4326 is the SRID for WGS84, which is the coordinate reference system
// used by KML as mandated by the OGC KML specification.
const SRID4326 = 4326

// KMLNamespace is the XML namespace for KML 2.2.
const KMLNamespace = "http://www.opengis.net/kml/2.2"

// DefaultFactory is a geometry factory configured for KML (WGS84/EPSG:4326).
var DefaultFactory = geom.NewGeometryFactoryWithSRID(SRID4326)

// Options configures KML marshaling behavior.
type Options struct {
	// Precision is the number of decimal places to output (-1 for default).
	Precision int
	// Formatted controls whether output includes indentation.
	Formatted bool
	// IncludeAltitude controls whether Z values are included in coordinates.
	IncludeAltitude bool
}

// DefaultOptions returns the default marshaling options.
func DefaultOptions() Options {
	return Options{
		Precision:       -1,
		Formatted:       false,
		IncludeAltitude: false,
	}
}

// Marshal marshals a geometry to KML bytes.
func Marshal(g geom.Geometry) ([]byte, error) {
	return MarshalWithOptions(g, DefaultOptions())
}

// MarshalWithOptions marshals a geometry to KML bytes with custom options.
func MarshalWithOptions(g geom.Geometry, opts Options) ([]byte, error) {
	kml := createKML(g, opts)
	if opts.Formatted {
		return xml.MarshalIndent(kml, "", "  ")
	}
	return xml.Marshal(kml)
}

// Unmarshal unmarshals KML bytes to a geometry.
func Unmarshal(data []byte) (geom.Geometry, error) {
	return UnmarshalWithFactory(data, DefaultFactory)
}

// UnmarshalWithFactory unmarshals KML bytes using a custom geometry factory.
func UnmarshalWithFactory(data []byte, factory *geom.GeometryFactory) (geom.Geometry, error) {
	var kml KML
	if err := xml.Unmarshal(data, &kml); err != nil {
		return nil, fmt.Errorf("failed to parse KML: %w", err)
	}

	return parseKML(&kml, factory)
}

// createKML creates a KML structure from a geometry.
func createKML(g geom.Geometry, opts Options) *KML {
	kml := &KML{
		Namespace: KMLNamespace,
		Placemark: &Placemark{},
	}

	setPlacemarkGeometry(kml.Placemark, g, opts)
	return kml
}

// setPlacemarkGeometry sets the geometry on a Placemark.
func setPlacemarkGeometry(pm *Placemark, g geom.Geometry, opts Options) {
	if g == nil || g.IsEmpty() {
		return
	}

	geom.VisitGeometry(g, geom.GeometryVisitor{
		Point: func(p *geom.Point) {
			pm.Point = createKMLPoint(p, opts)
		},
		LineString: func(ls *geom.LineString) {
			pm.LineString = createKMLLineString(ls, opts)
		},
		LinearRing: func(lr *geom.LinearRing) {
			pm.LinearRing = createKMLLinearRing(lr, opts)
		},
		Polygon: func(p *geom.Polygon) {
			pm.Polygon = createKMLPolygon(p, opts)
		},
		MultiPoint: func(mp *geom.MultiPoint) {
			pm.MultiGeometry = createKMLMultiGeometry(mp, opts)
		},
		MultiLineString: func(mls *geom.MultiLineString) {
			pm.MultiGeometry = createKMLMultiGeometry(mls, opts)
		},
		MultiPolygon: func(mp *geom.MultiPolygon) {
			pm.MultiGeometry = createKMLMultiGeometry(mp, opts)
		},
		GeometryCollection: func(gc *geom.GeometryCollection) {
			pm.MultiGeometry = createKMLMultiGeometry(gc, opts)
		},
	})
}

// createKMLPoint creates a KML Point from a geom.Point.
func createKMLPoint(p *geom.Point, opts Options) *Point {
	if p.IsEmpty() {
		return nil
	}
	return &Point{
		Coordinates: formatCoordinate(p.Coordinate(), opts),
	}
}

// createKMLLineString creates a KML LineString from a geom.LineString.
func createKMLLineString(ls *geom.LineString, opts Options) *LineString {
	if ls.IsEmpty() {
		return nil
	}
	return &LineString{
		Coordinates: formatCoordinates(ls.Coordinates(), opts),
	}
}

// createKMLLinearRing creates a KML LinearRing from a geom.LinearRing.
func createKMLLinearRing(lr *geom.LinearRing, opts Options) *LinearRing {
	if lr.IsEmpty() {
		return nil
	}
	return &LinearRing{
		Coordinates: formatCoordinates(lr.Coordinates(), opts),
	}
}

// createKMLPolygon creates a KML Polygon from a geom.Polygon.
func createKMLPolygon(poly *geom.Polygon, opts Options) *Polygon {
	if poly.IsEmpty() {
		return nil
	}

	kmlPoly := &Polygon{
		OuterBoundary: &OuterBoundaryIs{
			LinearRing: createKMLLinearRing(poly.ExteriorRing(), opts),
		},
	}

	for i := 0; i < poly.NumInteriorRings(); i++ {
		hole := poly.InteriorRingN(i)
		kmlPoly.InnerBoundries = append(kmlPoly.InnerBoundries, &InnerBoundaryIs{
			LinearRing: createKMLLinearRing(hole, opts),
		})
	}

	return kmlPoly
}

// createKMLMultiGeometry creates a KML MultiGeometry from a multi geometry.
func createKMLMultiGeometry(g geom.Geometry, opts Options) *MultiGeometry {
	if g == nil || g.IsEmpty() {
		return nil
	}

	mg := &MultiGeometry{}

	for i := 0; i < g.NumGeometries(); i++ {
		child := g.GeometryN(i)
		switch v := child.(type) {
		case *geom.Point:
			if kmlPt := createKMLPoint(v, opts); kmlPt != nil {
				mg.Points = append(mg.Points, kmlPt)
			}
		case *geom.LineString:
			if kmlLs := createKMLLineString(v, opts); kmlLs != nil {
				mg.LineStrings = append(mg.LineStrings, kmlLs)
			}
		case *geom.LinearRing:
			if kmlLr := createKMLLinearRing(v, opts); kmlLr != nil {
				mg.LinearRings = append(mg.LinearRings, kmlLr)
			}
		case *geom.Polygon:
			if kmlPoly := createKMLPolygon(v, opts); kmlPoly != nil {
				mg.Polygons = append(mg.Polygons, kmlPoly)
			}
		case *geom.MultiPoint, *geom.MultiLineString, *geom.MultiPolygon, *geom.GeometryCollection:
			if nested := createKMLMultiGeometry(v, opts); nested != nil {
				mg.MultiGeometries = append(mg.MultiGeometries, nested)
			}
		}
	}

	return mg
}

// formatCoordinate formats a single coordinate as "lon,lat[,alt]".
func formatCoordinate(c geom.Coordinate, opts Options) string {
	var sb strings.Builder
	ioutil.WriteNumber(&sb, c.X, opts.Precision)
	sb.WriteByte(',')
	ioutil.WriteNumber(&sb, c.Y, opts.Precision)
	if opts.IncludeAltitude && c.HasZ() {
		sb.WriteByte(',')
		ioutil.WriteNumber(&sb, c.Z, opts.Precision)
	}
	return sb.String()
}

// formatCoordinates formats a coordinate sequence as whitespace-separated tuples.
func formatCoordinates(coords geom.CoordinateSequence, opts Options) string {
	if len(coords) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, c := range coords {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(formatCoordinate(c, opts))
	}
	return sb.String()
}

// parseKML parses a KML structure into a geometry.
func parseKML(kml *KML, factory *geom.GeometryFactory) (geom.Geometry, error) {
	// Collect all placemarks from the document
	var placemarks []*Placemark

	if kml.Placemark != nil {
		placemarks = append(placemarks, kml.Placemark)
	}
	if kml.Document != nil {
		placemarks = append(placemarks, collectPlacemarks(kml.Document.Placemarks, kml.Document.Folders)...)
	}
	if kml.Folder != nil {
		placemarks = append(placemarks, collectPlacemarks(kml.Folder.Placemarks, kml.Folder.Folders)...)
	}

	if len(placemarks) == 0 {
		return factory.CreateGeometryCollectionEmpty(), nil
	}

	if len(placemarks) == 1 {
		return parsePlacemark(placemarks[0], factory)
	}

	// Multiple placemarks -> GeometryCollection
	var geoms []geom.Geometry
	for _, pm := range placemarks {
		g, err := parsePlacemark(pm, factory)
		if err != nil {
			return nil, err
		}
		if !g.IsEmpty() {
			geoms = append(geoms, g)
		}
	}

	if len(geoms) == 0 {
		return factory.CreateGeometryCollectionEmpty(), nil
	}
	if len(geoms) == 1 {
		return geoms[0], nil
	}

	return factory.CreateGeometryCollection(geoms), nil
}

// collectPlacemarks recursively collects placemarks from folders.
func collectPlacemarks(placemarks []*Placemark, folders []*Folder) []*Placemark {
	result := make([]*Placemark, 0, len(placemarks))
	result = append(result, placemarks...)
	for _, folder := range folders {
		result = append(result, collectPlacemarks(folder.Placemarks, folder.Folders)...)
	}
	return result
}

// parsePlacemark parses a Placemark into a geometry.
func parsePlacemark(pm *Placemark, factory *geom.GeometryFactory) (geom.Geometry, error) {
	if pm.Point != nil {
		return parsePoint(pm.Point, factory)
	}
	if pm.LineString != nil {
		return parseLineString(pm.LineString, factory)
	}
	if pm.LinearRing != nil {
		return parseLinearRing(pm.LinearRing, factory)
	}
	if pm.Polygon != nil {
		return parsePolygon(pm.Polygon, factory)
	}
	if pm.MultiGeometry != nil {
		return parseMultiGeometry(pm.MultiGeometry, factory)
	}
	return factory.CreateGeometryCollectionEmpty(), nil
}

// parsePoint parses a KML Point into a geom.Point.
func parsePoint(pt *Point, factory *geom.GeometryFactory) (*geom.Point, error) {
	coords, err := parseCoordinates(pt.Coordinates)
	if err != nil {
		return nil, fmt.Errorf("invalid point coordinates: %w", err)
	}
	if len(coords) == 0 {
		return factory.CreatePointEmpty(), nil
	}
	if len(coords) != 1 {
		return nil, fmt.Errorf("point requires exactly one coordinate tuple")
	}
	return factory.CreatePointFromCoordinate(coords[0]), nil
}

// parseLineString parses a KML LineString into a geom.LineString.
func parseLineString(ls *LineString, factory *geom.GeometryFactory) (*geom.LineString, error) {
	coords, err := parseCoordinates(ls.Coordinates)
	if err != nil {
		return nil, fmt.Errorf("invalid linestring coordinates: %w", err)
	}
	if len(coords) == 0 {
		return factory.CreateLineStringEmpty(), nil
	}
	if len(coords) < 2 {
		return nil, fmt.Errorf("linestring requires 0 or at least 2 coordinate tuples")
	}
	return factory.CreateLineString(coords), nil
}

// parseLinearRing parses a KML LinearRing into a geom.LinearRing.
func parseLinearRing(lr *LinearRing, factory *geom.GeometryFactory) (*geom.LinearRing, error) {
	coords, err := parseCoordinates(lr.Coordinates)
	if err != nil {
		return nil, fmt.Errorf("invalid linearring coordinates: %w", err)
	}
	if len(coords) == 0 {
		return factory.CreateLinearRingEmpty(), nil
	}
	if len(coords) < 4 {
		return nil, fmt.Errorf("linearring requires 0 or at least 4 coordinate tuples")
	}
	if !coords.IsClosed(geom.DefaultEpsilon) {
		return nil, fmt.Errorf("linearring must be closed")
	}
	return factory.CreateLinearRing(coords), nil
}

// parsePolygon parses a KML Polygon into a geom.Polygon.
func parsePolygon(poly *Polygon, factory *geom.GeometryFactory) (*geom.Polygon, error) {
	if poly.OuterBoundary == nil || poly.OuterBoundary.LinearRing == nil {
		return factory.CreatePolygonEmpty(), nil
	}

	shell, err := parseLinearRing(poly.OuterBoundary.LinearRing, factory)
	if err != nil {
		return nil, fmt.Errorf("invalid polygon outer boundary: %w", err)
	}

	var holes []*geom.LinearRing
	for i, inner := range poly.InnerBoundries {
		if inner.LinearRing == nil {
			continue
		}
		hole, err := parseLinearRing(inner.LinearRing, factory)
		if err != nil {
			return nil, fmt.Errorf("invalid polygon inner boundary %d: %w", i, err)
		}
		holes = append(holes, hole)
	}

	return factory.CreatePolygon(shell, holes), nil
}

// parseMultiGeometry parses a KML MultiGeometry into a geom.GeometryCollection.
func parseMultiGeometry(mg *MultiGeometry, factory *geom.GeometryFactory) (*geom.GeometryCollection, error) {
	var geoms []geom.Geometry

	for _, pt := range mg.Points {
		g, err := parsePoint(pt, factory)
		if err != nil {
			return nil, err
		}
		if g != nil && !g.IsEmpty() {
			geoms = append(geoms, g)
		}
	}

	for _, ls := range mg.LineStrings {
		g, err := parseLineString(ls, factory)
		if err != nil {
			return nil, err
		}
		if g != nil && !g.IsEmpty() {
			geoms = append(geoms, g)
		}
	}

	for _, lr := range mg.LinearRings {
		g, err := parseLinearRing(lr, factory)
		if err != nil {
			return nil, err
		}
		if g != nil && !g.IsEmpty() {
			geoms = append(geoms, g)
		}
	}

	for _, poly := range mg.Polygons {
		g, err := parsePolygon(poly, factory)
		if err != nil {
			return nil, err
		}
		if g != nil && !g.IsEmpty() {
			geoms = append(geoms, g)
		}
	}

	for _, nested := range mg.MultiGeometries {
		g, err := parseMultiGeometry(nested, factory)
		if err != nil {
			return nil, err
		}
		if g != nil && !g.IsEmpty() {
			geoms = append(geoms, g)
		}
	}

	if len(geoms) == 0 {
		return factory.CreateGeometryCollectionEmpty(), nil
	}

	return factory.CreateGeometryCollection(geoms), nil
}

// parseCoordinates parses a KML coordinate string into a CoordinateSequence.
// Format: "lon,lat[,alt] lon,lat[,alt] ..."
func parseCoordinates(s string) (geom.CoordinateSequence, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	// Split by whitespace
	tuples := strings.Fields(s)
	coords := make(geom.CoordinateSequence, 0, len(tuples))

	for _, tuple := range tuples {
		parts := strings.Split(tuple, ",")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid coordinate tuple: %q", tuple)
		}
		if len(parts) > 3 {
			return nil, fmt.Errorf("coordinate tuple has unsupported arity %d: %q", len(parts), tuple)
		}
		if strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return nil, fmt.Errorf("coordinate tuple has empty longitude or latitude: %q", tuple)
		}

		lon, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid longitude: %w", err)
		}

		lat, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid latitude: %w", err)
		}

		// KML uses lon,lat order, so X=lon, Y=lat
		coord := geom.NewCoordinate(lon, lat)

		// Parse optional altitude
		if len(parts) == 3 {
			if strings.TrimSpace(parts[2]) == "" {
				return nil, fmt.Errorf("coordinate tuple has empty altitude: %q", tuple)
			}
			alt, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
			if err != nil {
				return nil, fmt.Errorf("invalid altitude: %w", err)
			}
			coord.Z = alt
		}

		coords = append(coords, coord)
	}

	return coords, nil
}

// UnmarshalFeatures returns an iterator over features in KML data.
// Usage: for f, err := range kml.UnmarshalFeatures(data) { ... }
func UnmarshalFeatures(data []byte) iter.Seq2[*Feature, error] {
	return UnmarshalFeaturesWithFactory(data, DefaultFactory)
}

// UnmarshalFeaturesWithFactory uses a custom geometry factory.
func UnmarshalFeaturesWithFactory(data []byte, factory *geom.GeometryFactory) iter.Seq2[*Feature, error] {
	return func(yield func(*Feature, error) bool) {
		var kml KML
		if err := xml.Unmarshal(data, &kml); err != nil {
			yield(nil, fmt.Errorf("failed to parse KML: %w", err))
			return
		}

		// Collect all placemarks from the document
		var placemarks []*Placemark

		if kml.Placemark != nil {
			placemarks = append(placemarks, kml.Placemark)
		}
		if kml.Document != nil {
			placemarks = append(placemarks, collectPlacemarks(kml.Document.Placemarks, kml.Document.Folders)...)
		}
		if kml.Folder != nil {
			placemarks = append(placemarks, collectPlacemarks(kml.Folder.Placemarks, kml.Folder.Folders)...)
		}

		for _, pm := range placemarks {
			g, err := parsePlacemark(pm, factory)
			if err != nil {
				if !yield(nil, err) {
					return
				}
				continue
			}

			feature := &Feature{
				ID:          pm.ID,
				Name:        pm.Name,
				Description: pm.Description,
				Geometry:    g,
			}

			if !yield(feature, nil) {
				return
			}
		}
	}
}

// MarshalFeatures marshals features to KML.
func MarshalFeatures(features []*Feature) ([]byte, error) {
	return MarshalFeaturesWithOptions(features, DefaultOptions())
}

// MarshalFeaturesWithOptions marshals with custom options.
func MarshalFeaturesWithOptions(features []*Feature, opts Options) ([]byte, error) {
	kml := &KML{
		Namespace: KMLNamespace,
	}

	if len(features) == 0 {
		// Empty document
		kml.Document = &Document{}
	} else if len(features) == 1 {
		// Single feature as root placemark
		pm := featureToPlacemark(features[0], opts)
		kml.Placemark = pm
	} else {
		// Multiple features in a document
		doc := &Document{}
		for _, f := range features {
			pm := featureToPlacemark(f, opts)
			doc.Placemarks = append(doc.Placemarks, pm)
		}
		kml.Document = doc
	}

	if opts.Formatted {
		return xml.MarshalIndent(kml, "", "  ")
	}
	return xml.Marshal(kml)
}

// featureToPlacemark converts a Feature to a Placemark.
func featureToPlacemark(f *Feature, opts Options) *Placemark {
	pm := &Placemark{
		ID:          f.ID,
		Name:        f.Name,
		Description: f.Description,
	}
	setPlacemarkGeometry(pm, f.Geometry, opts)
	return pm
}
