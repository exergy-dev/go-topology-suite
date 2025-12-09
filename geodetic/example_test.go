package geodetic_test

import (
	"fmt"
	"github.com/go-topology-suite/gts/geodetic"
)

// ExampleDistance demonstrates calculating the geodesic distance between two points.
func ExampleDistance() {
	// New York City
	lat1, lon1 := 40.7128, -74.0060
	// London
	lat2, lon2 := 51.5074, -0.1278

	// Calculate distance using WGS84 ellipsoid
	distance := geodetic.Distance(lat1, lon1, lat2, lon2, geodetic.WGS84)

	fmt.Printf("Distance from NYC to London: %.0f km\n", distance/1000)
	// Output: Distance from NYC to London: 5585 km
}

// ExampleDistanceWGS84 demonstrates the convenience function for WGS84 calculations.
func ExampleDistanceWGS84() {
	// San Francisco
	lat1, lon1 := 37.7749, -122.4194
	// Tokyo
	lat2, lon2 := 35.6762, 139.6503

	distance := geodetic.DistanceWGS84(lat1, lon1, lat2, lon2)

	fmt.Printf("Distance: %.0f km\n", distance/1000)
	// Output: Distance: 8293 km
}

// ExampleHaversine demonstrates using the faster spherical approximation.
func ExampleHaversine() {
	// Paris
	lat1, lon1 := 48.8566, 2.3522
	// Berlin
	lat2, lon2 := 52.5200, 13.4050

	// Use mean Earth radius for spherical calculation
	distance := geodetic.Haversine(lat1, lon1, lat2, lon2, geodetic.EarthMeanRadius)

	fmt.Printf("Distance (spherical): %.0f km\n", distance/1000)
	// Output: Distance (spherical): 877 km
}

// ExampleInitialBearing demonstrates calculating the initial bearing between two points.
func ExampleInitialBearing() {
	// Start: New York
	lat1, lon1 := 40.7128, -74.0060
	// End: London
	lat2, lon2 := 51.5074, -0.1278

	bearing := geodetic.InitialBearing(lat1, lon1, lat2, lon2)

	fmt.Printf("Initial bearing: %.1f°\n", bearing)
	// Output: Initial bearing: 51.2°
}

// ExampleInverse demonstrates solving the inverse geodesic problem.
func ExampleInverse() {
	lat1, lon1 := -37.951033, 144.424868 // Flinders Peak
	lat2, lon2 := -37.652821, 143.926496 // Buninyong

	distance, azimuth1, azimuth2, err := geodetic.Inverse(lat1, lon1, lat2, lon2, geodetic.WGS84)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Distance: %.0f m\n", distance)
	fmt.Printf("Forward azimuth: %.2f°\n", azimuth1)
	fmt.Printf("Reverse azimuth: %.2f°\n", azimuth2)
	// Output:
	// Distance: 54972 m
	// Forward azimuth: 306.87°
	// Reverse azimuth: 307.17°
}

// ExampleDestinationPoint demonstrates calculating a destination point.
func ExampleDestinationPoint() {
	// Start at Sydney
	lat, lon := -33.8688, 151.2093

	// Travel 1000 km northeast (45 degrees)
	bearing := 45.0
	distance := 1000000.0 // meters

	lat2, lon2 := geodetic.DestinationPoint(lat, lon, bearing, distance, geodetic.WGS84)

	fmt.Printf("Destination: %.4f°, %.4f°\n", lat2, lon2)
	// Output: Destination: -27.2818°, 158.3405°
}

// ExampleDirect demonstrates solving the direct geodesic problem.
func ExampleDirect() {
	// Start at equator
	lat1, lon1 := 0.0, 0.0

	// Travel 1000 km due north
	azimuth1 := 0.0
	distance := 1000000.0

	lat2, lon2, azimuth2, err := geodetic.Direct(lat1, lon1, azimuth1, distance, geodetic.WGS84)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Destination: %.4f°, %.4f°\n", lat2, lon2)
	fmt.Printf("Final azimuth: %.2f°\n", azimuth2)
	// Output:
	// Destination: 9.0429°, 0.0000°
	// Final azimuth: 0.00°
}

// ExamplePolygonArea demonstrates calculating the area of a polygon.
func ExamplePolygonArea() {
	// Define a small rectangle near the equator
	lats := []float64{0, 0, 0.1, 0.1, 0}
	lons := []float64{0, 0.1, 0.1, 0, 0}

	area := geodetic.PolygonArea(lats, lons, geodetic.WGS84)

	fmt.Printf("Area: %.0f km²\n", area/1e6)
	// Output: Area: 123 km²
}

// ExamplePolygonAreaWGS84 demonstrates the convenience function for area calculation.
func ExamplePolygonAreaWGS84() {
	// Define a triangle
	lats := []float64{0, 0, 1, 0}
	lons := []float64{0, 1, 0, 0}

	area := geodetic.PolygonAreaWGS84(lats, lons)

	fmt.Printf("Triangle area: %.0f km²\n", area/1e6)
	// Output: Triangle area: 6154 km²
}

// ExampleSignedPolygonArea demonstrates detecting polygon winding order.
func ExampleSignedPolygonArea() {
	// Counter-clockwise square
	latsCCW := []float64{0, 0, 1, 1, 0}
	lonsCCW := []float64{0, 1, 1, 0, 0}

	// Clockwise square (reversed)
	latsCW := []float64{0, 1, 1, 0, 0}
	lonsCW := []float64{0, 0, 1, 1, 0}

	areaCCW := geodetic.SignedPolygonArea(latsCCW, lonsCCW, geodetic.WGS84)
	areaCW := geodetic.SignedPolygonArea(latsCW, lonsCW, geodetic.WGS84)

	if areaCCW < 0 {
		fmt.Println("First polygon: clockwise")
	}
	if areaCW > 0 {
		fmt.Println("Second polygon: counter-clockwise")
	}
	// Output:
	// First polygon: clockwise
	// Second polygon: counter-clockwise
}

// ExampleEllipsoid demonstrates working with different ellipsoids.
func ExampleEllipsoid() {
	// Compare distances using different ellipsoids
	lat1, lon1 := 40.7128, -74.0060
	lat2, lon2 := 51.5074, -0.1278

	distWGS84 := geodetic.Distance(lat1, lon1, lat2, lon2, geodetic.WGS84)
	distGRS80 := geodetic.Distance(lat1, lon1, lat2, lon2, geodetic.GRS80)
	distSphere := geodetic.Distance(lat1, lon1, lat2, lon2, geodetic.Sphere)

	fmt.Printf("WGS84:  %.1f km\n", distWGS84/1000)
	fmt.Printf("GRS80:  %.1f km\n", distGRS80/1000)
	fmt.Printf("Sphere: %.1f km\n", distSphere/1000)
	// Output:
	// WGS84:  5585.2 km
	// GRS80:  5585.2 km
	// Sphere: 5570.2 km
}

// ExampleNewEllipsoid demonstrates creating a custom ellipsoid.
func ExampleNewEllipsoid() {
	// Create a custom ellipsoid
	custom := geodetic.NewEllipsoidFromAInvF("Custom", 6378137.0, 300.0)

	fmt.Printf("Name: %s\n", custom.Name())
	fmt.Printf("Semi-major axis: %.0f m\n", custom.SemiMajorAxis())
	fmt.Printf("Semi-minor axis: %.0f m\n", custom.SemiMinorAxis())
	fmt.Printf("Flattening: 1/%.1f\n", custom.InverseFlattening())
	// Output:
	// Name: Custom
	// Semi-major axis: 6378137 m
	// Semi-minor axis: 6356877 m
	// Flattening: 1/300.0
}

// Example_roundTrip demonstrates the relationship between Direct and Inverse.
func Example_roundTrip() {
	// Starting point
	lat1, lon1 := 35.0, 45.0

	// Travel 500 km at 60 degrees
	azimuth := 60.0
	distance := 500000.0

	// Direct: find destination
	lat2, lon2, _, _ := geodetic.Direct(lat1, lon1, azimuth, distance, geodetic.WGS84)

	// Inverse: calculate back
	dist, az, _, _ := geodetic.Inverse(lat1, lon1, lat2, lon2, geodetic.WGS84)

	fmt.Printf("Round trip distance: %.0f m\n", dist)
	fmt.Printf("Round trip azimuth: %.1f°\n", az)
	// Output:
	// Round trip distance: 500000 m
	// Round trip azimuth: 60.0°
}

// Example_crossingDateline demonstrates handling the international date line.
func Example_crossingDateline() {
	// Point west of dateline
	lat1, lon1 := 35.6762, 139.6503 // Tokyo
	// Point east of dateline
	lat2, lon2 := 21.3099, -157.8581 // Honolulu

	distance := geodetic.DistanceWGS84(lat1, lon1, lat2, lon2)
	bearing := geodetic.InitialBearing(lat1, lon1, lat2, lon2)

	fmt.Printf("Distance: %.0f km\n", distance/1000)
	fmt.Printf("Bearing: %.1f°\n", bearing)
	// Output:
	// Distance: 6219 km
	// Bearing: 86.9°
}
