package geojson

import (
	"encoding/json"
	"testing"

	"github.com/go-topology-suite/gts/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalUnmarshalPoint(t *testing.T) {
	factory := geom.DefaultFactory
	p := factory.CreatePoint(1.5, 2.5)

	data, err := MarshalGeometry(p)
	require.NoError(t, err, "Failed to marshal")

	g, err := UnmarshalGeometry(data)
	require.NoError(t, err, "Failed to unmarshal")

	point, ok := g.(*geom.Point)
	require.True(t, ok, "Expected Point, got %T", g)

	coord := point.Coordinate()
	assert.Equal(t, 1.5, coord.X)
	assert.Equal(t, 2.5, coord.Y)
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
	require.NoError(t, err, "Failed to marshal")

	g, err := UnmarshalGeometry(data)
	require.NoError(t, err, "Failed to unmarshal")

	lineString, ok := g.(*geom.LineString)
	require.True(t, ok, "Expected LineString, got %T", g)

	assert.Len(t, lineString.Coordinates(), 3, "Expected 3 coordinates")
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
	require.NoError(t, err, "Failed to marshal")

	g, err := UnmarshalGeometry(data)
	require.NoError(t, err, "Failed to unmarshal")

	polygon, ok := g.(*geom.Polygon)
	require.True(t, ok, "Expected Polygon, got %T", g)

	assert.False(t, polygon.IsEmpty(), "Expected non-empty polygon")
}

func TestMarshalUnmarshalMultiGeometries(t *testing.T) {
	factory := geom.DefaultFactory

	// MultiPoint
	mp := factory.CreateMultiPoint([]*geom.Point{
		factory.CreatePoint(1, 2),
		factory.CreatePoint(3, 4),
	})

	data, err := MarshalGeometry(mp)
	require.NoError(t, err)
	g, err := UnmarshalGeometry(data)
	require.NoError(t, err)
	assert.Equal(t, "MultiPoint", g.GeometryType())

	// MultiLineString
	mls := factory.CreateMultiLineString([]*geom.LineString{
		factory.CreateLineString(geom.CoordinateSequence{
			geom.NewCoordinate(0, 0),
			geom.NewCoordinate(1, 1),
		}),
	})

	data, err = MarshalGeometry(mls)
	require.NoError(t, err)
	g, err = UnmarshalGeometry(data)
	require.NoError(t, err)
	assert.Equal(t, "MultiLineString", g.GeometryType())

	// GeometryCollection
	gc := factory.CreateGeometryCollection([]geom.Geometry{
		factory.CreatePoint(1, 2),
		factory.CreateLineString(geom.CoordinateSequence{
			geom.NewCoordinate(0, 0),
			geom.NewCoordinate(1, 1),
		}),
	})

	data, err = MarshalGeometry(gc)
	require.NoError(t, err)
	g, err = UnmarshalGeometry(data)
	require.NoError(t, err)
	assert.Equal(t, "GeometryCollection", g.GeometryType())
}

func TestUnmarshalFeatureExtractsGeometry(t *testing.T) {
	geojsonStr := `{
		"type": "Feature",
		"geometry": {"type": "Point", "coordinates": [102.0, 0.5]},
		"properties": {"name": "test"}
	}`

	g, err := UnmarshalGeometry([]byte(geojsonStr))
	require.NoError(t, err, "Failed to unmarshal")

	point, ok := g.(*geom.Point)
	require.True(t, ok, "Expected Point, got %T", g)

	coord := point.Coordinate()
	assert.Equal(t, 102.0, coord.X)
	assert.Equal(t, 0.5, coord.Y)
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
	require.NoError(t, err, "Failed to unmarshal")

	gc, ok := g.(*geom.GeometryCollection)
	require.True(t, ok, "Expected GeometryCollection, got %T", g)

	assert.Equal(t, 2, gc.NumGeometries(), "Expected 2 geometries")
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
	require.NoError(t, err, "Failed to marshal")

	// Unmarshal using standard json
	var f2 Feature[Props]
	err = json.Unmarshal(data, &f2)
	require.NoError(t, err, "Failed to unmarshal")

	assert.Equal(t, "NYC", f2.Properties.Name)
	assert.Equal(t, 8000000, f2.Properties.Pop)
	assert.Equal(t, "nyc", f2.ID.String)
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
	require.NoError(t, err, "Failed to marshal")

	var fc2 FeatureCollection[Props]
	err = json.Unmarshal(data, &fc2)
	require.NoError(t, err, "Failed to unmarshal")

	assert.Len(t, fc2.Features, 2, "Expected 2 features")
	assert.Equal(t, "A", fc2.Features[0].Properties.Name, "Expected first feature Name=A")
}

func TestUntypedFeature(t *testing.T) {
	factory := geom.DefaultFactory

	f := NewUntypedFeature(factory.CreatePoint(1, 2), map[string]any{"key": "value"})

	data, err := json.Marshal(f)
	require.NoError(t, err)

	var f2 UntypedFeature
	err = json.Unmarshal(data, &f2)
	require.NoError(t, err)

	assert.Equal(t, "value", f2.Properties["key"], "Expected key=value")
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
	err := json.Unmarshal([]byte(geojsonStr), &f)
	require.NoError(t, err, "Failed to unmarshal")

	assert.Equal(t, "value", f.ForeignMembers["custom"], "Expected custom=value")
	assert.Equal(t, float64(42), f.ForeignMembers["count"], "Expected count=42")

	// Round-trip preserves foreign members
	data, err := json.Marshal(f)
	require.NoError(t, err)
	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "value", parsed["custom"], "Foreign member not preserved")
}

func TestFeatureCollectionForeignMembers(t *testing.T) {
	geojsonStr := `{
		"type": "FeatureCollection",
		"features": [],
		"name": "test collection"
	}`

	var fc UntypedFeatureCollection
	err := json.Unmarshal([]byte(geojsonStr), &fc)
	require.NoError(t, err)

	assert.Equal(t, "test collection", fc.ForeignMembers["name"], "Expected foreign member name")
}

func TestBBox(t *testing.T) {
	geojsonStr := `{
		"type": "Feature",
		"bbox": [-10, -10, 10, 10],
		"geometry": {"type": "Point", "coordinates": [0, 0]},
		"properties": {}
	}`

	var f UntypedFeature
	err := json.Unmarshal([]byte(geojsonStr), &f)
	require.NoError(t, err)

	require.Len(t, f.BBox, 4, "Expected bbox with 4 elements")
	assert.Equal(t, float64(-10), f.BBox[0], "Unexpected bbox min value")
	assert.Equal(t, float64(10), f.BBox[2], "Unexpected bbox max value")

	// ToEnvelope
	env := f.BBox.ToEnvelope()
	assert.Equal(t, float64(-10), env.MinX, "ToEnvelope MinX failed")
	assert.Equal(t, float64(10), env.MaxX, "ToEnvelope MaxX failed")
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

	require.Len(t, f.BBox, 4, "Expected bbox to be set")
	assert.Equal(t, float64(0), f.BBox[0], "Unexpected bbox min")
	assert.Equal(t, float64(10), f.BBox[2], "Unexpected bbox max")
}

func TestFeatureID(t *testing.T) {
	// String ID
	id1 := NewStringID("abc")
	data, err := json.Marshal(id1)
	require.NoError(t, err)
	assert.Equal(t, `"abc"`, string(data), "Expected string ID")

	// Number ID
	id2 := NewNumberID(123)
	data, err = json.Marshal(id2)
	require.NoError(t, err)
	assert.Equal(t, `123`, string(data), "Expected number ID")

	// Unmarshal string
	var id3 FeatureID
	err = json.Unmarshal([]byte(`"test"`), &id3)
	require.NoError(t, err)
	assert.True(t, id3.IsValid, "Failed to unmarshal string ID - not valid")
	assert.False(t, id3.IsNum, "Failed to unmarshal string ID - is num")
	assert.Equal(t, "test", id3.String, "Failed to unmarshal string ID - wrong value")

	// Unmarshal number
	var id4 FeatureID
	err = json.Unmarshal([]byte(`456`), &id4)
	require.NoError(t, err)
	assert.True(t, id4.IsValid, "Failed to unmarshal number ID - not valid")
	assert.True(t, id4.IsNum, "Failed to unmarshal number ID - not num")
	assert.Equal(t, float64(456), id4.Number, "Failed to unmarshal number ID - wrong value")
}

func TestGeometryWrapper(t *testing.T) {
	factory := geom.DefaultFactory
	p := factory.CreatePoint(1, 2)

	// Wrap and marshal
	wrapped := Geometry{Geometry: p}
	data, err := json.Marshal(wrapped)
	require.NoError(t, err)

	// Unmarshal back
	var wrapped2 Geometry
	err = json.Unmarshal(data, &wrapped2)
	require.NoError(t, err)

	point := wrapped2.Geometry.(*geom.Point)
	coord := point.Coordinate()
	assert.Equal(t, float64(1), coord.X, "Round-trip X failed")
	assert.Equal(t, float64(2), coord.Y, "Round-trip Y failed")
}

func TestMarshalIndent(t *testing.T) {
	factory := geom.DefaultFactory
	p := factory.CreatePoint(1, 2)

	data, err := MarshalGeometryIndent(p, "  ")
	require.NoError(t, err)

	// Should contain newlines
	assert.GreaterOrEqual(t, len(data), 30, "Expected indented output")
}

func TestNullGeometry(t *testing.T) {
	geojsonStr := `{
		"type": "Feature",
		"geometry": null,
		"properties": {}
	}`

	g, err := UnmarshalGeometry([]byte(geojsonStr))
	require.NoError(t, err)

	// Should return empty geometry collection for null geometry
	assert.Equal(t, "GeometryCollection", g.GeometryType(), "Expected GeometryCollection for null geometry")
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
		assert.Error(t, err, "Expected error for: %s", tc)
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
		require.NoError(t, err, "Marshal failed for %s", g.GeometryType())

		g2, err := UnmarshalGeometry(data)
		require.NoError(t, err, "Unmarshal failed for %s", g.GeometryType())

		assert.Equal(t, g.GeometryType(), g2.GeometryType(), "Type mismatch")
	}
}

// TestGeometry_UnmarshalJSON_SRID verifies that geometries decoded via json.Unmarshal
// have the correct SRID (4326) as mandated by RFC 7946 for GeoJSON.
func TestGeometry_UnmarshalJSON_SRID(t *testing.T) {
	tests := []struct {
		name    string
		geojson string
		gtype   string
	}{
		{
			name:    "Point",
			geojson: `{"type": "Point", "coordinates": [1.5, 2.5]}`,
			gtype:   "Point",
		},
		{
			name:    "LineString",
			geojson: `{"type": "LineString", "coordinates": [[0, 0], [1, 1], [2, 0]]}`,
			gtype:   "LineString",
		},
		{
			name:    "Polygon",
			geojson: `{"type": "Polygon", "coordinates": [[[0, 0], [10, 0], [10, 10], [0, 10], [0, 0]]]}`,
			gtype:   "Polygon",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var g Geometry
			err := json.Unmarshal([]byte(tc.geojson), &g)
			require.NoError(t, err, "Failed to unmarshal %s", tc.name)
			require.NotNil(t, g.Geometry, "Geometry should not be nil")

			assert.Equal(t, tc.gtype, g.Geometry.GeometryType(), "Expected geometry type %s", tc.gtype)
			assert.Equal(t, SRID4326, g.Geometry.SRID(), "Expected SRID 4326 for %s", tc.name)
		})
	}
}

// TestFeature_UnmarshalJSON_SRID verifies that Feature geometry decoded via json.Unmarshal
// has the correct SRID (4326).
func TestFeature_UnmarshalJSON_SRID(t *testing.T) {
	geojsonStr := `{
		"type": "Feature",
		"geometry": {"type": "Point", "coordinates": [102.0, 0.5]},
		"properties": {"name": "test"}
	}`

	var f UntypedFeature
	err := json.Unmarshal([]byte(geojsonStr), &f)
	require.NoError(t, err, "Failed to unmarshal Feature")
	require.NotNil(t, f.Geometry, "Feature geometry should not be nil")
	require.NotNil(t, f.Geometry.Geometry, "Feature geometry.Geometry should not be nil")

	assert.Equal(t, SRID4326, f.Geometry.Geometry.SRID(), "Expected SRID 4326 for Feature geometry")
}

// TestFeatureCollection_UnmarshalJSON_SRID verifies that all geometries in a FeatureCollection
// decoded via json.Unmarshal have the correct SRID (4326).
func TestFeatureCollection_UnmarshalJSON_SRID(t *testing.T) {
	geojsonStr := `{
		"type": "FeatureCollection",
		"features": [
			{"type": "Feature", "geometry": {"type": "Point", "coordinates": [1, 2]}, "properties": {}},
			{"type": "Feature", "geometry": {"type": "LineString", "coordinates": [[0, 0], [1, 1]]}, "properties": {}},
			{"type": "Feature", "geometry": {"type": "Polygon", "coordinates": [[[0, 0], [1, 0], [1, 1], [0, 1], [0, 0]]]}, "properties": {}}
		]
	}`

	var fc UntypedFeatureCollection
	err := json.Unmarshal([]byte(geojsonStr), &fc)
	require.NoError(t, err, "Failed to unmarshal FeatureCollection")
	require.Len(t, fc.Features, 3, "Expected 3 features")

	for i, f := range fc.Features {
		require.NotNil(t, f.Geometry, "Feature %d geometry should not be nil", i)
		require.NotNil(t, f.Geometry.Geometry, "Feature %d geometry.Geometry should not be nil", i)
		assert.Equal(t, SRID4326, f.Geometry.Geometry.SRID(), "Expected SRID 4326 for Feature %d", i)
	}
}

// TestGeometry_RoundTrip_SRID verifies that SRID is preserved through marshal/unmarshal cycle.
func TestGeometry_RoundTrip_SRID(t *testing.T) {
	// Create geometry with SRID 4326 using the GeoJSON factory
	factory := DefaultFactory
	p := factory.CreatePoint(1.5, 2.5)
	require.Equal(t, SRID4326, p.SRID(), "Source geometry should have SRID 4326")

	// Marshal to GeoJSON
	wrapped := Geometry{Geometry: p}
	data, err := json.Marshal(wrapped)
	require.NoError(t, err, "Failed to marshal")

	// Unmarshal back
	var wrapped2 Geometry
	err = json.Unmarshal(data, &wrapped2)
	require.NoError(t, err, "Failed to unmarshal")

	// Verify SRID is preserved
	assert.Equal(t, SRID4326, wrapped2.Geometry.SRID(), "SRID should be preserved through round-trip")
}
