package geojson

import (
	"encoding/json"
	"testing"

	"github.com/go-topology-suite/gts/geom"
)

func TestMarshalUnmarshalPoint(t *testing.T) {
	factory := geom.DefaultFactory
	p := factory.CreatePoint(1.5, 2.5)

	data, err := MarshalGeometry(p)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	g, err := UnmarshalGeometry(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	point, ok := g.(*geom.Point)
	if !ok {
		t.Fatalf("Expected Point, got %T", g)
	}

	coord := point.Coordinate()
	if coord.X != 1.5 || coord.Y != 2.5 {
		t.Errorf("Expected (1.5, 2.5), got (%v, %v)", coord.X, coord.Y)
	}
}

func TestMarshalUnmarshalLineString(t *testing.T) {
	factory := geom.DefaultFactory
	coords := geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(1, 1),
		geom.NewCoordinate(2, 0),
	}
	ls := factory.CreateLineString(coords)

	data, err := MarshalGeometry(ls)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	g, err := UnmarshalGeometry(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	lineString, ok := g.(*geom.LineString)
	if !ok {
		t.Fatalf("Expected LineString, got %T", g)
	}

	if len(lineString.Coordinates()) != 3 {
		t.Errorf("Expected 3 coordinates, got %d", len(lineString.Coordinates()))
	}
}

func TestMarshalUnmarshalPolygon(t *testing.T) {
	factory := geom.DefaultFactory
	shell := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	poly := factory.CreatePolygon(shell, nil)

	data, err := MarshalGeometry(poly)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	g, err := UnmarshalGeometry(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	polygon, ok := g.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon, got %T", g)
	}

	if polygon.IsEmpty() {
		t.Error("Expected non-empty polygon")
	}
}

func TestMarshalUnmarshalMultiGeometries(t *testing.T) {
	factory := geom.DefaultFactory

	// MultiPoint
	mp := factory.CreateMultiPoint([]*geom.Point{
		factory.CreatePoint(1, 2),
		factory.CreatePoint(3, 4),
	})

	data, _ := MarshalGeometry(mp)
	g, _ := UnmarshalGeometry(data)
	if g.GeometryType() != "MultiPoint" {
		t.Errorf("Expected MultiPoint, got %s", g.GeometryType())
	}

	// MultiLineString
	mls := factory.CreateMultiLineString([]*geom.LineString{
		factory.CreateLineString(geom.CoordinateSequence{
			geom.NewCoordinate(0, 0),
			geom.NewCoordinate(1, 1),
		}),
	})

	data, _ = MarshalGeometry(mls)
	g, _ = UnmarshalGeometry(data)
	if g.GeometryType() != "MultiLineString" {
		t.Errorf("Expected MultiLineString, got %s", g.GeometryType())
	}

	// GeometryCollection
	gc := factory.CreateGeometryCollection([]geom.Geometry{
		factory.CreatePoint(1, 2),
		factory.CreateLineString(geom.CoordinateSequence{
			geom.NewCoordinate(0, 0),
			geom.NewCoordinate(1, 1),
		}),
	})

	data, _ = MarshalGeometry(gc)
	g, _ = UnmarshalGeometry(data)
	if g.GeometryType() != "GeometryCollection" {
		t.Errorf("Expected GeometryCollection, got %s", g.GeometryType())
	}
}

func TestUnmarshalFeatureExtractsGeometry(t *testing.T) {
	geojsonStr := `{
		"type": "Feature",
		"geometry": {"type": "Point", "coordinates": [102.0, 0.5]},
		"properties": {"name": "test"}
	}`

	g, err := UnmarshalGeometry([]byte(geojsonStr))
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	point, ok := g.(*geom.Point)
	if !ok {
		t.Fatalf("Expected Point, got %T", g)
	}

	coord := point.Coordinate()
	if coord.X != 102.0 || coord.Y != 0.5 {
		t.Errorf("Expected (102.0, 0.5), got (%v, %v)", coord.X, coord.Y)
	}
}

func TestUnmarshalFeatureCollectionExtractsGeometries(t *testing.T) {
	geojsonStr := `{
		"type": "FeatureCollection",
		"features": [
			{"type": "Feature", "geometry": {"type": "Point", "coordinates": [1, 2]}, "properties": {}},
			{"type": "Feature", "geometry": {"type": "Point", "coordinates": [3, 4]}, "properties": {}}
		]
	}`

	g, err := UnmarshalGeometry([]byte(geojsonStr))
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	gc, ok := g.(*geom.GeometryCollection)
	if !ok {
		t.Fatalf("Expected GeometryCollection, got %T", g)
	}

	if gc.NumGeometries() != 2 {
		t.Errorf("Expected 2 geometries, got %d", gc.NumGeometries())
	}
}

func TestTypedFeatureMarshalUnmarshal(t *testing.T) {
	factory := geom.DefaultFactory

	type Props struct {
		Name string `json:"name"`
		Pop  int    `json:"population"`
	}

	f := NewFeature(factory.CreatePoint(1, 2), Props{Name: "NYC", Pop: 8000000})
	f.ID = NewStringID("nyc")

	// Marshal using standard json
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal using standard json
	var f2 Feature[Props]
	if err := json.Unmarshal(data, &f2); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if f2.Properties.Name != "NYC" {
		t.Errorf("Expected Name=NYC, got %s", f2.Properties.Name)
	}
	if f2.Properties.Pop != 8000000 {
		t.Errorf("Expected Pop=8000000, got %d", f2.Properties.Pop)
	}
	if f2.ID.String != "nyc" {
		t.Errorf("Expected ID=nyc, got %s", f2.ID.String)
	}
}

func TestTypedFeatureCollectionMarshalUnmarshal(t *testing.T) {
	factory := geom.DefaultFactory

	type Props struct {
		Name string `json:"name"`
	}

	fc := NewFeatureCollection[Props]()
	fc.Add(NewFeature(factory.CreatePoint(1, 2), Props{Name: "A"}))
	fc.Add(NewFeature(factory.CreatePoint(3, 4), Props{Name: "B"}))

	data, err := json.Marshal(fc)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var fc2 FeatureCollection[Props]
	if err := json.Unmarshal(data, &fc2); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(fc2.Features) != 2 {
		t.Errorf("Expected 2 features, got %d", len(fc2.Features))
	}
	if fc2.Features[0].Properties.Name != "A" {
		t.Errorf("Expected first feature Name=A")
	}
}

func TestUntypedFeature(t *testing.T) {
	factory := geom.DefaultFactory

	f := NewUntypedFeature(factory.CreatePoint(1, 2), map[string]any{"key": "value"})

	data, _ := json.Marshal(f)

	var f2 UntypedFeature
	json.Unmarshal(data, &f2)

	if f2.Properties["key"] != "value" {
		t.Errorf("Expected key=value")
	}
}

func TestForeignMembers(t *testing.T) {
	geojsonStr := `{
		"type": "Feature",
		"geometry": {"type": "Point", "coordinates": [1, 2]},
		"properties": {},
		"custom": "value",
		"count": 42
	}`

	var f UntypedFeature
	if err := json.Unmarshal([]byte(geojsonStr), &f); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if f.ForeignMembers["custom"] != "value" {
		t.Errorf("Expected custom=value")
	}
	if f.ForeignMembers["count"] != float64(42) {
		t.Errorf("Expected count=42")
	}

	// Round-trip preserves foreign members
	data, _ := json.Marshal(f)
	var parsed map[string]any
	json.Unmarshal(data, &parsed)

	if parsed["custom"] != "value" {
		t.Error("Foreign member not preserved")
	}
}

func TestFeatureCollectionForeignMembers(t *testing.T) {
	geojsonStr := `{
		"type": "FeatureCollection",
		"features": [],
		"name": "test collection"
	}`

	var fc UntypedFeatureCollection
	json.Unmarshal([]byte(geojsonStr), &fc)

	if fc.ForeignMembers["name"] != "test collection" {
		t.Errorf("Expected foreign member name")
	}
}

func TestBBox(t *testing.T) {
	geojsonStr := `{
		"type": "Feature",
		"bbox": [-10, -10, 10, 10],
		"geometry": {"type": "Point", "coordinates": [0, 0]},
		"properties": {}
	}`

	var f UntypedFeature
	json.Unmarshal([]byte(geojsonStr), &f)

	if len(f.BBox) != 4 {
		t.Fatalf("Expected bbox with 4 elements")
	}
	if f.BBox[0] != -10 || f.BBox[2] != 10 {
		t.Errorf("Unexpected bbox values")
	}

	// ToEnvelope
	env := f.BBox.ToEnvelope()
	if env.MinX != -10 || env.MaxX != 10 {
		t.Errorf("ToEnvelope failed")
	}
}

func TestSetBBox(t *testing.T) {
	factory := geom.DefaultFactory
	shell := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	poly := factory.CreatePolygon(shell, nil)

	f := NewUntypedFeature(poly, nil)
	f.SetBBox()

	if len(f.BBox) != 4 {
		t.Fatalf("Expected bbox to be set")
	}
	if f.BBox[0] != 0 || f.BBox[2] != 10 {
		t.Errorf("Unexpected bbox: %v", f.BBox)
	}
}

func TestFeatureID(t *testing.T) {
	// String ID
	id1 := NewStringID("abc")
	data, _ := json.Marshal(id1)
	if string(data) != `"abc"` {
		t.Errorf("Expected string ID")
	}

	// Number ID
	id2 := NewNumberID(123)
	data, _ = json.Marshal(id2)
	if string(data) != `123` {
		t.Errorf("Expected number ID")
	}

	// Unmarshal string
	var id3 FeatureID
	json.Unmarshal([]byte(`"test"`), &id3)
	if !id3.IsValid || id3.IsNum || id3.String != "test" {
		t.Error("Failed to unmarshal string ID")
	}

	// Unmarshal number
	var id4 FeatureID
	json.Unmarshal([]byte(`456`), &id4)
	if !id4.IsValid || !id4.IsNum || id4.Number != 456 {
		t.Error("Failed to unmarshal number ID")
	}
}

func TestGeometryWrapper(t *testing.T) {
	factory := geom.DefaultFactory
	p := factory.CreatePoint(1, 2)

	// Wrap and marshal
	wrapped := Geometry{Geometry: p}
	data, _ := json.Marshal(wrapped)

	// Unmarshal back
	var wrapped2 Geometry
	json.Unmarshal(data, &wrapped2)

	point := wrapped2.Geometry.(*geom.Point)
	coord := point.Coordinate()
	if coord.X != 1 || coord.Y != 2 {
		t.Errorf("Round-trip failed")
	}
}

func TestMarshalIndent(t *testing.T) {
	factory := geom.DefaultFactory
	p := factory.CreatePoint(1, 2)

	data, err := MarshalGeometryIndent(p, "  ")
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	// Should contain newlines
	if len(data) < 30 {
		t.Error("Expected indented output")
	}
}

func TestNullGeometry(t *testing.T) {
	geojsonStr := `{
		"type": "Feature",
		"geometry": null,
		"properties": {}
	}`

	g, err := UnmarshalGeometry([]byte(geojsonStr))
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	// Should return empty geometry collection for null geometry
	if g.GeometryType() != "GeometryCollection" {
		t.Errorf("Expected GeometryCollection for null geometry")
	}
}

func TestInvalidGeoJSON(t *testing.T) {
	testCases := []string{
		"",
		"{invalid}",
		`{"coordinates": [1, 2]}`,
		`{"type": "Unknown", "coordinates": [1, 2]}`,
	}

	for _, tc := range testCases {
		_, err := UnmarshalGeometry([]byte(tc))
		if err == nil {
			t.Errorf("Expected error for: %s", tc)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	factory := geom.DefaultFactory

	geoms := []geom.Geometry{
		factory.CreatePoint(1.5, 2.5),
		factory.CreateLineString(geom.CoordinateSequence{
			geom.NewCoordinate(0, 0),
			geom.NewCoordinate(1, 1),
		}),
		factory.CreatePolygon(
			factory.CreateLinearRing(geom.CoordinateSequence{
				geom.NewCoordinate(0, 0),
				geom.NewCoordinate(10, 0),
				geom.NewCoordinate(10, 10),
				geom.NewCoordinate(0, 10),
				geom.NewCoordinate(0, 0),
			}), nil),
	}

	for _, g := range geoms {
		data, err := MarshalGeometry(g)
		if err != nil {
			t.Fatalf("Marshal failed for %s: %v", g.GeometryType(), err)
		}

		g2, err := UnmarshalGeometry(data)
		if err != nil {
			t.Fatalf("Unmarshal failed for %s: %v", g.GeometryType(), err)
		}

		if g.GeometryType() != g2.GeometryType() {
			t.Errorf("Type mismatch: %s vs %s", g.GeometryType(), g2.GeometryType())
		}
	}
}
