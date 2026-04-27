package kml

import (
	"strings"
	"testing"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalUnmarshalPoint(t *testing.T) {
	factory := geom.DefaultFactory
	p := factory.CreatePoint(-122.084, 37.422)

	data, err := Marshal(p)
	require.NoError(t, err, "Failed to marshal")

	g, err := Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal")

	point, ok := g.(*geom.Point)
	require.True(t, ok, "Expected Point, got %T", g)

	coord := point.Coordinate()
	assert.InDelta(t, -122.084, coord.X, 0.0001)
	assert.InDelta(t, 37.422, coord.Y, 0.0001)
	assert.Equal(t, SRID4326, point.SRID(), "Expected SRID 4326")
}

func TestMarshalUnmarshalLineString(t *testing.T) {
	factory := geom.DefaultFactory
	coords := geom.CoordinateSequence{
		geom.NewCoordinate(-122.084, 37.422),
		geom.NewCoordinate(-122.085, 37.423),
		geom.NewCoordinate(-122.086, 37.424),
	}
	ls := factory.CreateLineString(coords)

	data, err := Marshal(ls)
	require.NoError(t, err, "Failed to marshal")

	g, err := Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal")

	lineString, ok := g.(*geom.LineString)
	require.True(t, ok, "Expected LineString, got %T", g)

	assert.Len(t, lineString.Coordinates(), 3, "Expected 3 coordinates")
	assert.Equal(t, SRID4326, lineString.SRID(), "Expected SRID 4326")
}

func TestMarshalUnmarshalPolygon(t *testing.T) {
	factory := geom.DefaultFactory
	shell := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(-122.084, 37.422),
		geom.NewCoordinate(-122.085, 37.422),
		geom.NewCoordinate(-122.085, 37.423),
		geom.NewCoordinate(-122.084, 37.423),
		geom.NewCoordinate(-122.084, 37.422),
	})
	poly := factory.CreatePolygon(shell, nil)

	data, err := Marshal(poly)
	require.NoError(t, err, "Failed to marshal")

	g, err := Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal")

	polygon, ok := g.(*geom.Polygon)
	require.True(t, ok, "Expected Polygon, got %T", g)

	assert.False(t, polygon.IsEmpty(), "Expected non-empty polygon")
	assert.Equal(t, 5, len(polygon.ExteriorRing().Coordinates()), "Expected 5 coordinates in exterior ring")
	assert.Equal(t, 0, polygon.NumInteriorRings(), "Expected no holes")
	assert.Equal(t, SRID4326, polygon.SRID(), "Expected SRID 4326")
}

func TestMarshalUnmarshalPolygonWithHole(t *testing.T) {
	factory := geom.DefaultFactory
	shell := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	hole := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(2, 2),
		geom.NewCoordinate(8, 2),
		geom.NewCoordinate(8, 8),
		geom.NewCoordinate(2, 8),
		geom.NewCoordinate(2, 2),
	})
	poly := factory.CreatePolygon(shell, []*geom.LinearRing{hole})

	data, err := Marshal(poly)
	require.NoError(t, err, "Failed to marshal")

	g, err := Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal")

	polygon, ok := g.(*geom.Polygon)
	require.True(t, ok, "Expected Polygon, got %T", g)

	assert.False(t, polygon.IsEmpty(), "Expected non-empty polygon")
	assert.Equal(t, 1, polygon.NumInteriorRings(), "Expected 1 hole")

	// Verify hole coordinates
	holeRing := polygon.InteriorRingN(0)
	assert.Equal(t, 5, len(holeRing.Coordinates()), "Expected 5 coordinates in hole")
}

func TestMarshalUnmarshalMultiGeometry(t *testing.T) {
	factory := geom.DefaultFactory

	gc := factory.CreateGeometryCollection([]geom.Geometry{
		factory.CreatePoint(1, 2),
		factory.CreateLineString(geom.CoordinateSequence{
			geom.NewCoordinate(0, 0),
			geom.NewCoordinate(1, 1),
		}),
	})

	data, err := Marshal(gc)
	require.NoError(t, err, "Failed to marshal")

	g, err := Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal")

	collection, ok := g.(*geom.GeometryCollection)
	require.True(t, ok, "Expected GeometryCollection, got %T", g)

	assert.Equal(t, 2, collection.NumGeometries(), "Expected 2 geometries")
	assert.Equal(t, SRID4326, collection.SRID(), "Expected SRID 4326")
}

func TestEmptyGeometries(t *testing.T) {
	factory := geom.DefaultFactory

	// Empty point
	emptyPoint := factory.CreatePointEmpty()
	data, err := Marshal(emptyPoint)
	require.NoError(t, err, "Failed to marshal empty point")
	g, err := Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal empty point")
	assert.True(t, g.IsEmpty(), "Expected empty geometry")

	// Empty linestring
	emptyLs := factory.CreateLineStringEmpty()
	data, err = Marshal(emptyLs)
	require.NoError(t, err, "Failed to marshal empty linestring")
	g, err = Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal empty linestring")
	assert.True(t, g.IsEmpty(), "Expected empty geometry")

	// Empty polygon
	emptyPoly := factory.CreatePolygonEmpty()
	data, err = Marshal(emptyPoly)
	require.NoError(t, err, "Failed to marshal empty polygon")
	g, err = Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal empty polygon")
	assert.True(t, g.IsEmpty(), "Expected empty geometry")
}

func TestCoordinateOrder(t *testing.T) {
	// KML uses lon,lat order (X=lon, Y=lat)
	kmlData := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
  <Placemark>
    <Point>
      <coordinates>-122.084,37.422</coordinates>
    </Point>
  </Placemark>
</kml>`

	g, err := Unmarshal([]byte(kmlData))
	require.NoError(t, err, "Failed to unmarshal")

	point, ok := g.(*geom.Point)
	require.True(t, ok, "Expected Point, got %T", g)

	coord := point.Coordinate()
	// X should be longitude (-122.084), Y should be latitude (37.422)
	assert.InDelta(t, -122.084, coord.X, 0.0001, "X should be longitude")
	assert.InDelta(t, 37.422, coord.Y, 0.0001, "Y should be latitude")
}

func TestAltitude(t *testing.T) {
	// Test parsing altitude
	kmlData := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
  <Placemark>
    <Point>
      <coordinates>-122.084,37.422,100.5</coordinates>
    </Point>
  </Placemark>
</kml>`

	g, err := Unmarshal([]byte(kmlData))
	require.NoError(t, err, "Failed to unmarshal")

	point, ok := g.(*geom.Point)
	require.True(t, ok, "Expected Point, got %T", g)

	coord := point.Coordinate()
	require.True(t, coord.HasZ(), "Expected Z value")
	assert.InDelta(t, 100.5, coord.Z, 0.0001, "Z should be altitude")

	// Test marshaling with altitude
	opts := Options{
		IncludeAltitude: true,
		Precision:       1, // Force 1 decimal place to preserve .5
	}
	data, err := MarshalWithOptions(point, opts)
	require.NoError(t, err, "Failed to marshal with altitude")
	assert.Contains(t, string(data), "100.5", "Output should contain altitude")
}

func TestRoundTrip(t *testing.T) {
	factory := geom.DefaultFactory

	testCases := []struct {
		name string
		geom geom.Geometry
	}{
		{
			name: "Point",
			geom: factory.CreatePoint(-122.084, 37.422),
		},
		{
			name: "LineString",
			geom: factory.CreateLineString(geom.CoordinateSequence{
				geom.NewCoordinate(0, 0),
				geom.NewCoordinate(1, 1),
				geom.NewCoordinate(2, 0),
			}),
		},
		{
			name: "Polygon",
			geom: factory.CreatePolygon(
				factory.CreateLinearRing(geom.CoordinateSequence{
					geom.NewCoordinate(0, 0),
					geom.NewCoordinate(10, 0),
					geom.NewCoordinate(10, 10),
					geom.NewCoordinate(0, 10),
					geom.NewCoordinate(0, 0),
				}), nil),
		},
		{
			name: "MultiPoint",
			geom: factory.CreateMultiPoint([]*geom.Point{
				factory.CreatePoint(1, 2),
				factory.CreatePoint(3, 4),
			}),
		},
		{
			name: "GeometryCollection",
			geom: factory.CreateGeometryCollection([]geom.Geometry{
				factory.CreatePoint(1, 2),
				factory.CreateLineString(geom.CoordinateSequence{
					geom.NewCoordinate(0, 0),
					geom.NewCoordinate(1, 1),
				}),
			}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := Marshal(tc.geom)
			require.NoError(t, err, "Marshal failed for %s", tc.name)

			g2, err := Unmarshal(data)
			require.NoError(t, err, "Unmarshal failed for %s", tc.name)

			// Compare coordinates
			coords1 := tc.geom.Coordinates()
			coords2 := g2.Coordinates()
			require.Equal(t, len(coords1), len(coords2), "Coordinate count mismatch")

			for i := range coords1 {
				assert.InDelta(t, coords1[i].X, coords2[i].X, 0.0001, "X mismatch at %d", i)
				assert.InDelta(t, coords1[i].Y, coords2[i].Y, 0.0001, "Y mismatch at %d", i)
			}
		})
	}
}

func TestParseCoordinates(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected int
		hasZ     bool
	}{
		{
			name:     "single 2D coordinate",
			input:    "-122.084,37.422",
			expected: 1,
			hasZ:     false,
		},
		{
			name:     "single 3D coordinate",
			input:    "-122.084,37.422,100",
			expected: 1,
			hasZ:     true,
		},
		{
			name:     "multiple 2D coordinates",
			input:    "-122.084,37.422 -122.085,37.423 -122.086,37.424",
			expected: 3,
			hasZ:     false,
		},
		{
			name:     "multiple 3D coordinates",
			input:    "-122.084,37.422,100 -122.085,37.423,200",
			expected: 2,
			hasZ:     true,
		},
		{
			name:     "coordinates with newlines",
			input:    "-122.084,37.422\n-122.085,37.423\n-122.086,37.424",
			expected: 3,
			hasZ:     false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
			hasZ:     false,
		},
		{
			name:     "whitespace only",
			input:    "   \n\t  ",
			expected: 0,
			hasZ:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			coords, err := parseCoordinates(tc.input)
			require.NoError(t, err, "parseCoordinates failed")
			assert.Len(t, coords, tc.expected, "Unexpected coordinate count")

			if tc.expected > 0 && tc.hasZ {
				assert.True(t, coords[0].HasZ(), "Expected Z value")
			}
		})
	}
}

func TestParseCoordinatesErrors(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "single value",
			input: "123",
		},
		{
			name:  "invalid number",
			input: "abc,def",
		},
		{
			name:  "empty longitude",
			input: ",2",
		},
		{
			name:  "empty latitude",
			input: "1,",
		},
		{
			name:  "empty altitude",
			input: "1,2,",
		},
		{
			name:  "too many ordinates",
			input: "1,2,3,4",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseCoordinates(tc.input)
			assert.Error(t, err, "Expected error for input: %s", tc.input)
		})
	}
}

func TestUnmarshalMalformedKMLGeometries(t *testing.T) {
	testCases := []struct {
		name string
		kml  string
	}{
		{
			name: "point with multiple coordinate tuples",
			kml:  `<kml><Placemark><Point><coordinates>1,2 3,4</coordinates></Point></Placemark></kml>`,
		},
		{
			name: "linestring with one coordinate tuple",
			kml:  `<kml><Placemark><LineString><coordinates>1,2</coordinates></LineString></Placemark></kml>`,
		},
		{
			name: "linearring with too few coordinate tuples",
			kml:  `<kml><Placemark><LinearRing><coordinates>0,0 1,0 0,0</coordinates></LinearRing></Placemark></kml>`,
		},
		{
			name: "linearring not closed",
			kml:  `<kml><Placemark><LinearRing><coordinates>0,0 1,0 1,1 0,1</coordinates></LinearRing></Placemark></kml>`,
		},
		{
			name: "polygon outer ring not closed",
			kml:  `<kml><Placemark><Polygon><outerBoundaryIs><LinearRing><coordinates>0,0 1,0 1,1 0,1</coordinates></LinearRing></outerBoundaryIs></Polygon></Placemark></kml>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Unmarshal([]byte(tc.kml))
			assert.Error(t, err)
		})
	}
}

func TestMarshalOptions(t *testing.T) {
	factory := geom.DefaultFactory
	p := factory.CreatePoint(-122.0845678, 37.4229012)

	// Test precision
	opts := Options{Precision: 3}
	data, err := MarshalWithOptions(p, opts)
	require.NoError(t, err)
	assert.Contains(t, string(data), "-122.085", "Should have 3 decimal places")

	// Test formatted output
	opts = Options{Formatted: true}
	data, err = MarshalWithOptions(p, opts)
	require.NoError(t, err)
	assert.Contains(t, string(data), "\n", "Should contain newlines")
}

func TestKMLNamespace(t *testing.T) {
	factory := geom.DefaultFactory
	p := factory.CreatePoint(1, 2)

	data, err := Marshal(p)
	require.NoError(t, err)

	assert.Contains(t, string(data), KMLNamespace, "Should contain KML namespace")
}

func TestMultipleCoordinateFormats(t *testing.T) {
	// Test various spacing formats
	testCases := []struct {
		name        string
		coordinates string
		expected    int
	}{
		{
			name:        "space separated",
			coordinates: "1,2 3,4 5,6",
			expected:    3,
		},
		{
			name:        "newline separated",
			coordinates: "1,2\n3,4\n5,6",
			expected:    3,
		},
		{
			name:        "tab separated",
			coordinates: "1,2\t3,4\t5,6",
			expected:    3,
		},
		{
			name:        "mixed whitespace",
			coordinates: "1,2  \n\t 3,4   5,6",
			expected:    3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			kmlData := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
  <Placemark>
    <LineString>
      <coordinates>` + tc.coordinates + `</coordinates>
    </LineString>
  </Placemark>
</kml>`

			g, err := Unmarshal([]byte(kmlData))
			require.NoError(t, err, "Failed to unmarshal")

			ls, ok := g.(*geom.LineString)
			require.True(t, ok, "Expected LineString, got %T", g)
			assert.Len(t, ls.Coordinates(), tc.expected, "Expected %d coordinates", tc.expected)
		})
	}
}

func TestUnmarshalDocumentWithMultiplePlacemarks(t *testing.T) {
	kmlData := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
  <Document>
    <name>Test Document</name>
    <Placemark>
      <name>Point 1</name>
      <Point><coordinates>1,2</coordinates></Point>
    </Placemark>
    <Placemark>
      <name>Point 2</name>
      <Point><coordinates>3,4</coordinates></Point>
    </Placemark>
  </Document>
</kml>`

	g, err := Unmarshal([]byte(kmlData))
	require.NoError(t, err, "Failed to unmarshal")

	gc, ok := g.(*geom.GeometryCollection)
	require.True(t, ok, "Expected GeometryCollection, got %T", g)
	assert.Equal(t, 2, gc.NumGeometries(), "Expected 2 geometries")
}

func TestUnmarshalFolder(t *testing.T) {
	kmlData := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
  <Folder>
    <name>Test Folder</name>
    <Placemark>
      <Point><coordinates>1,2</coordinates></Point>
    </Placemark>
  </Folder>
</kml>`

	g, err := Unmarshal([]byte(kmlData))
	require.NoError(t, err, "Failed to unmarshal")

	point, ok := g.(*geom.Point)
	require.True(t, ok, "Expected Point, got %T", g)

	coord := point.Coordinate()
	assert.InDelta(t, 1.0, coord.X, 0.0001)
	assert.InDelta(t, 2.0, coord.Y, 0.0001)
}

func TestLinearRing(t *testing.T) {
	factory := geom.DefaultFactory
	lr := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})

	data, err := Marshal(lr)
	require.NoError(t, err, "Failed to marshal")

	// Check that it contains LinearRing element
	assert.True(t, strings.Contains(string(data), "LinearRing"), "Should contain LinearRing element")

	g, err := Unmarshal(data)
	require.NoError(t, err, "Failed to unmarshal")

	ring, ok := g.(*geom.LinearRing)
	require.True(t, ok, "Expected LinearRing, got %T", g)
	assert.Equal(t, 5, len(ring.Coordinates()), "Expected 5 coordinates")
}

func TestUnmarshalFeatures(t *testing.T) {
	kmlData := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
  <Document>
    <Placemark id="pm1">
      <name>Test Point</name>
      <description>A test point feature</description>
      <Point><coordinates>1,2</coordinates></Point>
    </Placemark>
  </Document>
</kml>`

	var features []*Feature
	for f, err := range UnmarshalFeatures([]byte(kmlData)) {
		require.NoError(t, err, "Unexpected error during iteration")
		features = append(features, f)
	}

	require.Len(t, features, 1, "Expected 1 feature")
	assert.Equal(t, "pm1", features[0].ID, "Expected ID 'pm1'")
	assert.Equal(t, "Test Point", features[0].Name, "Expected name 'Test Point'")
	assert.Equal(t, "A test point feature", features[0].Description, "Expected description")
	assert.NotNil(t, features[0].Geometry, "Expected non-nil geometry")

	point, ok := features[0].Geometry.(*geom.Point)
	require.True(t, ok, "Expected Point geometry, got %T", features[0].Geometry)
	coord := point.Coordinate()
	assert.InDelta(t, 1.0, coord.X, 0.0001)
	assert.InDelta(t, 2.0, coord.Y, 0.0001)
}

func TestFeatureIterator(t *testing.T) {
	kmlData := `<?xml version="1.0" encoding="UTF-8"?>
<kml xmlns="http://www.opengis.net/kml/2.2">
  <Document>
    <Placemark id="first">
      <name>First</name>
      <Point><coordinates>1,1</coordinates></Point>
    </Placemark>
    <Placemark id="second">
      <name>Second</name>
      <Point><coordinates>2,2</coordinates></Point>
    </Placemark>
    <Placemark id="third">
      <name>Third</name>
      <Point><coordinates>3,3</coordinates></Point>
    </Placemark>
  </Document>
</kml>`

	// Test that features are yielded in order
	var names []string
	for f, err := range UnmarshalFeatures([]byte(kmlData)) {
		require.NoError(t, err, "Unexpected error during iteration")
		names = append(names, f.Name)
	}

	require.Len(t, names, 3, "Expected 3 features")
	assert.Equal(t, []string{"First", "Second", "Third"}, names, "Features should be in order")
}

func TestMarshalFeatures(t *testing.T) {
	factory := geom.DefaultFactory

	features := []*Feature{
		{
			ID:          "feature1",
			Name:        "My Point",
			Description: "A point description",
			Geometry:    factory.CreatePoint(1, 2),
		},
		{
			ID:          "feature2",
			Name:        "My Line",
			Description: "A line description",
			Geometry: factory.CreateLineString(geom.CoordinateSequence{
				geom.NewCoordinate(0, 0),
				geom.NewCoordinate(1, 1),
			}),
		},
	}

	data, err := MarshalFeatures(features)
	require.NoError(t, err, "Failed to marshal features")

	// Verify output contains expected elements
	kmlStr := string(data)
	assert.Contains(t, kmlStr, "feature1", "Should contain feature1 ID")
	assert.Contains(t, kmlStr, "feature2", "Should contain feature2 ID")
	assert.Contains(t, kmlStr, "My Point", "Should contain feature name")
	assert.Contains(t, kmlStr, "A point description", "Should contain feature description")
	assert.Contains(t, kmlStr, "<Document>", "Should contain Document element for multiple features")

	// Round-trip test: unmarshal and verify
	var roundTripped []*Feature
	for f, err := range UnmarshalFeatures(data) {
		require.NoError(t, err, "Failed to unmarshal round-tripped data")
		roundTripped = append(roundTripped, f)
	}

	require.Len(t, roundTripped, 2, "Expected 2 features after round-trip")
	assert.Equal(t, "feature1", roundTripped[0].ID)
	assert.Equal(t, "My Point", roundTripped[0].Name)
	assert.Equal(t, "A point description", roundTripped[0].Description)
}
