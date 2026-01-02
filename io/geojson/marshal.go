package geojson

import (
	"encoding/json"
	"fmt"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// MarshalJSON implements json.Marshaler for Geometry.
func (g Geometry) MarshalJSON() ([]byte, error) {
	if g.Geometry == nil || g.Geometry.IsEmpty() {
		return json.Marshal(nil)
	}
	return json.Marshal(geometryToMap(g.Geometry))
}

// UnmarshalJSON implements json.Unmarshaler for Geometry.
func (g *Geometry) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		g.Geometry = nil
		return nil
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	geomObj, err := parseGeometry(raw, DefaultFactory)
	if err != nil {
		return err
	}

	g.Geometry = geomObj
	return nil
}

// MarshalJSON implements json.Marshaler for Feature.
func (f Feature[P]) MarshalJSON() ([]byte, error) {
	obj := map[string]any{
		"type":       "Feature",
		"properties": f.Properties,
	}

	if f.ID.IsValid {
		obj["id"] = f.ID
	}

	if f.Geometry != nil && f.Geometry.Geometry != nil {
		obj["geometry"] = geometryToMap(f.Geometry.Geometry)
	} else {
		obj["geometry"] = nil
	}

	if len(f.BBox) > 0 {
		obj["bbox"] = f.BBox
	}

	// Add foreign members
	for k, v := range f.ForeignMembers {
		if _, exists := obj[k]; !exists {
			obj[k] = v
		}
	}

	return json.Marshal(obj)
}

// UnmarshalJSON implements json.Unmarshaler for Feature.
func (f *Feature[P]) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Verify type
	if typeData, ok := raw["type"]; ok {
		var typeName string
		if err := json.Unmarshal(typeData, &typeName); err != nil {
			return fmt.Errorf("invalid type field: %w", err)
		}
		if typeName != "Feature" {
			return fmt.Errorf("expected type 'Feature', got '%s'", typeName)
		}
		delete(raw, "type")
	}

	// Parse ID
	if idData, ok := raw["id"]; ok {
		if err := json.Unmarshal(idData, &f.ID); err != nil {
			return fmt.Errorf("invalid id: %w", err)
		}
		delete(raw, "id")
	}

	// Parse geometry
	if geomData, ok := raw["geometry"]; ok {
		if string(geomData) != "null" {
			f.Geometry = &Geometry{}
			if err := f.Geometry.UnmarshalJSON(geomData); err != nil {
				return fmt.Errorf("invalid geometry: %w", err)
			}
		}
		delete(raw, "geometry")
	}

	// Parse properties
	if propsData, ok := raw["properties"]; ok {
		if err := json.Unmarshal(propsData, &f.Properties); err != nil {
			return fmt.Errorf("invalid properties: %w", err)
		}
		delete(raw, "properties")
	}

	// Parse bbox
	if bboxData, ok := raw["bbox"]; ok {
		if err := json.Unmarshal(bboxData, &f.BBox); err != nil {
			return fmt.Errorf("invalid bbox: %w", err)
		}
		delete(raw, "bbox")
	}

	// Collect foreign members
	if len(raw) > 0 {
		f.ForeignMembers = make(ForeignMembers)
		for k, v := range raw {
			var val any
			if err := json.Unmarshal(v, &val); err != nil {
				return fmt.Errorf("invalid foreign member '%s': %w", k, err)
			}
			f.ForeignMembers[k] = val
		}
	}

	return nil
}

// MarshalJSON implements json.Marshaler for FeatureCollection.
func (fc FeatureCollection[P]) MarshalJSON() ([]byte, error) {
	obj := map[string]any{
		"type":     "FeatureCollection",
		"features": fc.Features,
	}

	if len(fc.BBox) > 0 {
		obj["bbox"] = fc.BBox
	}

	// Add foreign members
	for k, v := range fc.ForeignMembers {
		if _, exists := obj[k]; !exists {
			obj[k] = v
		}
	}

	return json.Marshal(obj)
}

// UnmarshalJSON implements json.Unmarshaler for FeatureCollection.
func (fc *FeatureCollection[P]) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Verify type
	if typeData, ok := raw["type"]; ok {
		var typeName string
		if err := json.Unmarshal(typeData, &typeName); err != nil {
			return fmt.Errorf("invalid type field: %w", err)
		}
		if typeName != "FeatureCollection" {
			return fmt.Errorf("expected type 'FeatureCollection', got '%s'", typeName)
		}
		delete(raw, "type")
	}

	// Parse features
	if featuresData, ok := raw["features"]; ok {
		if err := json.Unmarshal(featuresData, &fc.Features); err != nil {
			return fmt.Errorf("invalid features: %w", err)
		}
		delete(raw, "features")
	} else {
		fc.Features = make([]*Feature[P], 0)
	}

	// Parse bbox
	if bboxData, ok := raw["bbox"]; ok {
		if err := json.Unmarshal(bboxData, &fc.BBox); err != nil {
			return fmt.Errorf("invalid bbox: %w", err)
		}
		delete(raw, "bbox")
	}

	// Collect foreign members
	if len(raw) > 0 {
		fc.ForeignMembers = make(ForeignMembers)
		for k, v := range raw {
			var val any
			if err := json.Unmarshal(v, &val); err != nil {
				return fmt.Errorf("invalid foreign member '%s': %w", k, err)
			}
			fc.ForeignMembers[k] = val
		}
	}

	return nil
}

// geometryToMap converts a geometry to a GeoJSON map structure.
func geometryToMap(g geom.Geometry) map[string]any {
	if g == nil || g.IsEmpty() {
		return nil
	}

	var result map[string]any
	geom.VisitGeometry(g, geom.GeometryVisitor{
		Point: func(p *geom.Point) {
			result = map[string]any{
				"type":        "Point",
				"coordinates": coordToArray(p.Coordinate()),
			}
		},
		LineString: func(ls *geom.LineString) {
			result = map[string]any{
				"type":        "LineString",
				"coordinates": coordsToArray(ls.Coordinates()),
			}
		},
		LinearRing: func(lr *geom.LinearRing) {
			result = map[string]any{
				"type":        "LineString",
				"coordinates": coordsToArray(lr.Coordinates()),
			}
		},
		Polygon: func(p *geom.Polygon) {
			result = polygonToMap(p)
		},
		MultiPoint: func(mp *geom.MultiPoint) {
			result = multiPointToMap(mp)
		},
		MultiLineString: func(mls *geom.MultiLineString) {
			result = multiLineStringToMap(mls)
		},
		MultiPolygon: func(mp *geom.MultiPolygon) {
			result = multiPolygonToMap(mp)
		},
		GeometryCollection: func(gc *geom.GeometryCollection) {
			result = geometryCollectionToMap(gc)
		},
		Default: func(geom.Geometry) {
			result = map[string]any{"type": "GeometryCollection", "geometries": []any{}}
		},
	})
	return result
}

func polygonToMap(p *geom.Polygon) map[string]any {
	if p.IsEmpty() {
		return map[string]any{"type": "Polygon", "coordinates": []any{}}
	}

	rings := make([]any, 1+p.NumInteriorRings())
	rings[0] = coordsToArray(p.ExteriorRing().Coordinates())
	for i := 0; i < p.NumInteriorRings(); i++ {
		rings[i+1] = coordsToArray(p.InteriorRingN(i).Coordinates())
	}
	return map[string]any{
		"type":        "Polygon",
		"coordinates": rings,
	}
}

func multiPointToMap(mp *geom.MultiPoint) map[string]any {
	coords := make([]any, mp.NumGeometries())
	for i := 0; i < mp.NumGeometries(); i++ {
		coords[i] = coordToArray(mp.GeometryN(i).(*geom.Point).Coordinate())
	}
	return map[string]any{
		"type":        "MultiPoint",
		"coordinates": coords,
	}
}

func multiLineStringToMap(mls *geom.MultiLineString) map[string]any {
	lines := make([]any, mls.NumGeometries())
	for i := 0; i < mls.NumGeometries(); i++ {
		lines[i] = coordsToArray(mls.GeometryN(i).(*geom.LineString).Coordinates())
	}
	return map[string]any{
		"type":        "MultiLineString",
		"coordinates": lines,
	}
}

func multiPolygonToMap(mp *geom.MultiPolygon) map[string]any {
	polys := make([]any, mp.NumGeometries())
	for i := 0; i < mp.NumGeometries(); i++ {
		p := mp.GeometryN(i).(*geom.Polygon)
		rings := make([]any, 1+p.NumInteriorRings())
		rings[0] = coordsToArray(p.ExteriorRing().Coordinates())
		for j := 0; j < p.NumInteriorRings(); j++ {
			rings[j+1] = coordsToArray(p.InteriorRingN(j).Coordinates())
		}
		polys[i] = rings
	}
	return map[string]any{
		"type":        "MultiPolygon",
		"coordinates": polys,
	}
}

func geometryCollectionToMap(gc *geom.GeometryCollection) map[string]any {
	geoms := make([]any, gc.NumGeometries())
	for i := 0; i < gc.NumGeometries(); i++ {
		geoms[i] = geometryToMap(gc.GeometryN(i))
	}
	return map[string]any{
		"type":       "GeometryCollection",
		"geometries": geoms,
	}
}

func coordToArray(c geom.Coordinate) []float64 {
	if c.Z != nil {
		return []float64{c.X, c.Y, *c.Z}
	}
	return []float64{c.X, c.Y}
}

func coordsToArray(coords geom.CoordinateSequence) [][]float64 {
	result := make([][]float64, len(coords))
	for i, c := range coords {
		result[i] = coordToArray(c)
	}
	return result
}

// parseGeometry parses a geometry from raw JSON.
func parseGeometry(raw map[string]json.RawMessage, factory *geom.GeometryFactory) (geom.Geometry, error) {
	typeData, ok := raw["type"]
	if !ok {
		return nil, fmt.Errorf("missing 'type' field")
	}

	var geomType string
	if err := json.Unmarshal(typeData, &geomType); err != nil {
		return nil, fmt.Errorf("invalid 'type' field: %w", err)
	}

	coordsData := raw["coordinates"]

	switch geomType {
	case "Point":
		return parsePoint(coordsData, factory)
	case "LineString":
		return parseLineString(coordsData, factory)
	case "Polygon":
		return parsePolygon(coordsData, factory)
	case "MultiPoint":
		return parseMultiPoint(coordsData, factory)
	case "MultiLineString":
		return parseMultiLineString(coordsData, factory)
	case "MultiPolygon":
		return parseMultiPolygon(coordsData, factory)
	case "GeometryCollection":
		return parseGeometryCollection(raw, factory)
	default:
		return nil, fmt.Errorf("unsupported geometry type: %s", geomType)
	}
}

func parsePoint(data json.RawMessage, factory *geom.GeometryFactory) (*geom.Point, error) {
	var coords []float64
	if err := json.Unmarshal(data, &coords); err != nil {
		return nil, fmt.Errorf("invalid Point coordinates: %w", err)
	}

	if len(coords) == 0 {
		return factory.CreatePointEmpty(), nil
	}

	if len(coords) < 2 {
		return nil, fmt.Errorf("Point requires at least 2 coordinates")
	}

	coord := geom.NewCoordinate(coords[0], coords[1])
	if len(coords) >= 3 {
		z := coords[2]
		coord.Z = &z
	}

	return factory.CreatePointFromCoordinate(coord), nil
}

func parseLineString(data json.RawMessage, factory *geom.GeometryFactory) (*geom.LineString, error) {
	coords, err := parseCoordinateArray(data)
	if err != nil {
		return nil, fmt.Errorf("invalid LineString coordinates: %w", err)
	}

	if len(coords) == 0 {
		return factory.CreateLineStringEmpty(), nil
	}

	return factory.CreateLineString(coords), nil
}

func parsePolygon(data json.RawMessage, factory *geom.GeometryFactory) (*geom.Polygon, error) {
	var rings []json.RawMessage
	if err := json.Unmarshal(data, &rings); err != nil {
		return nil, fmt.Errorf("invalid Polygon coordinates: %w", err)
	}

	if len(rings) == 0 {
		return factory.CreatePolygonEmpty(), nil
	}

	shellCoords, err := parseCoordinateArray(rings[0])
	if err != nil {
		return nil, fmt.Errorf("invalid Polygon shell: %w", err)
	}
	shell := factory.CreateLinearRing(shellCoords)

	holes := make([]*geom.LinearRing, len(rings)-1)
	for i := 1; i < len(rings); i++ {
		holeCoords, err := parseCoordinateArray(rings[i])
		if err != nil {
			return nil, fmt.Errorf("invalid Polygon hole %d: %w", i, err)
		}
		holes[i-1] = factory.CreateLinearRing(holeCoords)
	}

	return factory.CreatePolygon(shell, holes), nil
}

func parseMultiPoint(data json.RawMessage, factory *geom.GeometryFactory) (*geom.MultiPoint, error) {
	var coordArrays [][]float64
	if err := json.Unmarshal(data, &coordArrays); err != nil {
		return nil, fmt.Errorf("invalid MultiPoint coordinates: %w", err)
	}

	if len(coordArrays) == 0 {
		return factory.CreateMultiPointEmpty(), nil
	}

	points := make([]*geom.Point, len(coordArrays))
	for i, coords := range coordArrays {
		if len(coords) < 2 {
			return nil, fmt.Errorf("MultiPoint coordinate %d requires at least 2 values", i)
		}
		coord := geom.NewCoordinate(coords[0], coords[1])
		if len(coords) >= 3 {
			z := coords[2]
			coord.Z = &z
		}
		points[i] = factory.CreatePointFromCoordinate(coord)
	}

	return factory.CreateMultiPoint(points), nil
}

func parseMultiLineString(data json.RawMessage, factory *geom.GeometryFactory) (*geom.MultiLineString, error) {
	var lineArrays []json.RawMessage
	if err := json.Unmarshal(data, &lineArrays); err != nil {
		return nil, fmt.Errorf("invalid MultiLineString coordinates: %w", err)
	}

	if len(lineArrays) == 0 {
		return factory.CreateMultiLineStringEmpty(), nil
	}

	lines := make([]*geom.LineString, len(lineArrays))
	for i, lineData := range lineArrays {
		coords, err := parseCoordinateArray(lineData)
		if err != nil {
			return nil, fmt.Errorf("invalid MultiLineString line %d: %w", i, err)
		}
		lines[i] = factory.CreateLineString(coords)
	}

	return factory.CreateMultiLineString(lines), nil
}

func parseMultiPolygon(data json.RawMessage, factory *geom.GeometryFactory) (*geom.MultiPolygon, error) {
	var polyArrays []json.RawMessage
	if err := json.Unmarshal(data, &polyArrays); err != nil {
		return nil, fmt.Errorf("invalid MultiPolygon coordinates: %w", err)
	}

	if len(polyArrays) == 0 {
		return factory.CreateMultiPolygonEmpty(), nil
	}

	polys := make([]*geom.Polygon, len(polyArrays))
	for i, polyData := range polyArrays {
		poly, err := parsePolygon(polyData, factory)
		if err != nil {
			return nil, fmt.Errorf("invalid MultiPolygon polygon %d: %w", i, err)
		}
		polys[i] = poly
	}

	return factory.CreateMultiPolygon(polys), nil
}

func parseGeometryCollection(raw map[string]json.RawMessage, factory *geom.GeometryFactory) (*geom.GeometryCollection, error) {
	geomsData, ok := raw["geometries"]
	if !ok {
		return nil, fmt.Errorf("missing 'geometries' field in GeometryCollection")
	}

	var geomsRaw []json.RawMessage
	if err := json.Unmarshal(geomsData, &geomsRaw); err != nil {
		return nil, fmt.Errorf("invalid 'geometries' array: %w", err)
	}

	geoms := make([]geom.Geometry, len(geomsRaw))
	for i, gData := range geomsRaw {
		var geomRaw map[string]json.RawMessage
		if err := json.Unmarshal(gData, &geomRaw); err != nil {
			return nil, fmt.Errorf("invalid geometry %d: %w", i, err)
		}
		g, err := parseGeometry(geomRaw, factory)
		if err != nil {
			return nil, fmt.Errorf("error reading geometry %d: %w", i, err)
		}
		geoms[i] = g
	}

	return factory.CreateGeometryCollection(geoms), nil
}

func parseCoordinateArray(data json.RawMessage) (geom.CoordinateSequence, error) {
	var coordArrays [][]float64
	if err := json.Unmarshal(data, &coordArrays); err != nil {
		return nil, err
	}

	coords := make(geom.CoordinateSequence, len(coordArrays))
	for i, arr := range coordArrays {
		if len(arr) < 2 {
			return nil, fmt.Errorf("coordinate %d requires at least 2 values", i)
		}
		coords[i] = geom.NewCoordinate(arr[0], arr[1])
		if len(arr) >= 3 {
			z := arr[2]
			coords[i].Z = &z
		}
	}

	return coords, nil
}
