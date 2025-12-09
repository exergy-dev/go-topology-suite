package spherical_test

import (
	"fmt"

	"github.com/go-topology-suite/gts/geom"
	"github.com/go-topology-suite/gts/spherical"
)

// ExampleDistance demonstrates calculating geodesic distance between two cities.
func ExampleDistance() {
	// New York City
	nyc := geom.NewPoint(-74.0060, 40.7128)

	// London
	london := geom.NewPoint(-0.1278, 51.5074)

	// Calculate distance in meters
	distanceMeters := spherical.Distance(nyc, london)

	// Convert to kilometers
	distanceKm := distanceMeters / 1000.0

	fmt.Printf("Distance from NYC to London: %.0f km\n", distanceKm)
	// Output: Distance from NYC to London: 5570 km
}

// ExampleArea demonstrates calculating the area of a geographic polygon.
func ExampleArea() {
	// Create a polygon representing approximately 1 degree square near the equator
	ring := geom.NewLinearRingXY(
		0.0, 0.0,   // Southwest corner
		1.0, 0.0,   // Southeast corner
		1.0, 1.0,   // Northeast corner
		0.0, 1.0,   // Northwest corner
		0.0, 0.0,   // Close the ring
	)
	polygon := geom.NewPolygon(ring, nil)

	// Calculate area in square meters
	areaM2 := spherical.Area(polygon)

	// Convert to square kilometers
	areaKm2 := areaM2 / 1000000.0

	fmt.Printf("Area of 1° square near equator: %.0f km²\n", areaKm2)
	// Output: Area of 1° square near equator: 12364 km²
}

// ExampleContains demonstrates point-in-polygon testing on the sphere.
func ExampleContains() {
	// Create a polygon around Paris
	ring := geom.NewLinearRingXY(
		2.0, 48.5,  // Southwest
		2.7, 48.5,  // Southeast
		2.7, 49.0,  // Northeast
		2.0, 49.0,  // Northwest
		2.0, 48.5,  // Close
	)
	polygon := geom.NewPolygon(ring, nil)

	// Paris coordinates (Eiffel Tower)
	paris := geom.NewPoint(2.2945, 48.8584)

	// London coordinates
	london := geom.NewPoint(-0.1278, 51.5074)

	fmt.Printf("Paris in polygon: %v\n", spherical.Contains(polygon, paris))
	fmt.Printf("London in polygon: %v\n", spherical.Contains(polygon, london))
	// Output:
	// Paris in polygon: true
	// London in polygon: false
}

// ExampleCellToken demonstrates S2 cell indexing for spatial queries.
func ExampleCellToken() {
	// San Francisco coordinates
	sf := geom.NewPoint(-122.4194, 37.7749)

	// Get S2 cell token at different levels
	// Level 10: City scale (~1000 km²)
	cityLevel := spherical.CellToken(sf, 10)

	// Level 15: Neighborhood scale (~10 km²)
	neighborhoodLevel := spherical.CellToken(sf, 15)

	// Level 20: Building scale (~400 m²)
	buildingLevel := spherical.CellToken(sf, 20)

	fmt.Printf("City level token: %s\n", cityLevel)
	fmt.Printf("Neighborhood level token: %s\n", neighborhoodLevel)
	fmt.Printf("Building level token: %s\n", buildingLevel)

	// Note: Longer tokens = smaller areas
	// All levels of a point share a common prefix
	// Output:
	// City level token: 80858004
	// Neighborhood level token: 8085800c94
	// Building level token: 8085800c9479c
}

// ExampleCovering demonstrates covering a polygon with S2 cells for indexing.
func ExampleCovering() {
	// Create a polygon
	ring := geom.NewLinearRingXY(
		-122.5, 37.7,
		-122.3, 37.7,
		-122.3, 37.8,
		-122.5, 37.8,
		-122.5, 37.7,
	)
	polygon := geom.NewPolygon(ring, nil)

	// Get S2 cell covering
	// minLevel=10 (city scale), maxLevel=15 (neighborhood scale), maxCells=8
	cells := spherical.Covering(polygon, 10, 15, 8)

	fmt.Printf("Polygon covered by %d S2 cells\n", len(cells))
	fmt.Printf("First cell level: %d\n", spherical.CellLevel(cells[0]))

	// These cells can be used for spatial indexing
	// Store them in a database to enable efficient spatial queries
	// Output:
	// Polygon covered by 8 S2 cells
	// First cell level: 12
}

// ExampleLength demonstrates calculating the length of a path on the sphere.
func ExampleLength() {
	// Create a path: NYC -> London -> Paris
	path := geom.NewLineStringXY(
		-74.0060, 40.7128, // NYC
		-0.1278, 51.5074,  // London
		2.3522, 48.8566,   // Paris
	)

	// Calculate total path length
	lengthMeters := spherical.Length(path)
	lengthKm := lengthMeters / 1000.0

	fmt.Printf("Total path length: %.0f km\n", lengthKm)
	// Output: Total path length: 5914 km
}

// ExamplePerimeter demonstrates calculating polygon perimeter.
func ExamplePerimeter() {
	// Create a rectangle
	ring := geom.NewLinearRingXY(
		0.0, 0.0,
		1.0, 0.0,
		1.0, 1.0,
		0.0, 1.0,
		0.0, 0.0,
	)
	polygon := geom.NewPolygon(ring, nil)

	perimeterMeters := spherical.Perimeter(polygon)
	perimeterKm := perimeterMeters / 1000.0

	fmt.Printf("Perimeter: %.0f km\n", perimeterKm)
	// Output: Perimeter: 445 km
}

// ExampleCentroid demonstrates finding the spherical centroid of a polygon.
func ExampleCentroid() {
	// Create a triangle
	ring := geom.NewLinearRingXY(
		0.0, 0.0,
		2.0, 0.0,
		1.0, 2.0,
		0.0, 0.0,
	)
	polygon := geom.NewPolygon(ring, nil)

	centroid := spherical.Centroid(polygon)

	fmt.Printf("Centroid: (%.2f, %.2f)\n", centroid.X(), centroid.Y())
	// Output: Centroid: (1.00, 0.67)
}
