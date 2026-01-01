package transform_test

import (
	"fmt"
	"math"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/transform"
	"github.com/robert-malhotra/go-topology-suite/transform/projection"
)

// Example demonstrates basic affine transformations
func Example_affineTransform() {
	// Create a square polygon
	shell := geom.NewLinearRingXY(0, 0, 10, 0, 10, 10, 0, 10, 0, 0)
	polygon := geom.NewPolygon(shell, []*geom.LinearRing{})

	// Create a composite transformation: scale by 2, then translate by (5, 5)
	scale := transform.NewAffineScale(2, 2)
	translate := transform.NewAffineTranslation(5, 5)
	composite := transform.NewComposite(scale, translate)

	// Transform the polygon
	result, _ := transform.TransformGeometry(composite, polygon)
	resultPoly := result.(*geom.Polygon)

	// The original square (0,0) to (10,10) becomes (5,5) to (25,25)
	coords := resultPoly.ExteriorRing().Coordinates()
	fmt.Printf("First coordinate: (%.0f, %.0f)\n", coords[0].X, coords[0].Y)
	fmt.Printf("Third coordinate: (%.0f, %.0f)\n", coords[2].X, coords[2].Y)

	// Output:
	// First coordinate: (5, 5)
	// Third coordinate: (25, 25)
}

// Example demonstrates rotation transformation
func Example_rotation() {
	// Create a point at (1, 0)
	point := geom.NewPoint(1, 0)

	// Rotate 90 degrees counter-clockwise
	rotation := transform.NewAffineRotation(math.Pi / 2)

	// Transform the point
	result, _ := transform.TransformGeometry(rotation, point)
	resultPoint := result.(*geom.Point)

	// The point should now be at approximately (0, 1)
	fmt.Printf("Rotated point: (%.0f, %.0f)\n",
		math.Round(resultPoint.X()),
		math.Round(resultPoint.Y()))

	// Output:
	// Rotated point: (0, 1)
}

// Example demonstrates Web Mercator projection
func Example_webMercator() {
	// Create Web Mercator projection
	wm := projection.WebMercator()

	// San Francisco coordinates (WGS84 lon/lat)
	lon, lat := -122.4194, 37.7749

	// Project to Web Mercator (meters)
	x, y, _ := wm.Forward(lon, lat)

	// Print in kilometers for readability
	fmt.Printf("San Francisco in Web Mercator:\n")
	fmt.Printf("  X: %.0f km from central meridian\n", x/1000)
	fmt.Printf("  Y: %.0f km from equator\n", y/1000)

	// Inverse projection back to lon/lat
	lonInv, latInv, _ := wm.Inverse(x, y)
	fmt.Printf("Inverse projection: (%.4f, %.4f)\n", lonInv, latInv)

	// Output:
	// San Francisco in Web Mercator:
	//   X: -13628 km from central meridian
	//   Y: 4548 km from equator
	// Inverse projection: (-122.4194, 37.7749)
}

// Example demonstrates UTM projection
func Example_utm() {
	// Create UTM Zone 10N projection (covers San Francisco)
	utm := projection.UTM(10, true, nil)

	// San Francisco coordinates
	lon, lat := -122.4194, 37.7749

	// Project to UTM
	easting, northing, _ := utm.Forward(lon, lat)

	fmt.Printf("San Francisco in UTM Zone 10N:\n")
	fmt.Printf("  Easting: %.0f m\n", easting)
	fmt.Printf("  Northing: %.0f m\n", northing)

	// Inverse projection
	lonInv, latInv, _ := utm.Inverse(easting, northing)
	fmt.Printf("Inverse: (%.4f, %.4f)\n", lonInv, latInv)

	// Output:
	// San Francisco in UTM Zone 10N:
	//   Easting: 551131 m
	//   Northing: 4180999 m
	// Inverse: (-122.4194, 37.7749)
}

// Example demonstrates transforming a LineString
func Example_transformLineString() {
	// Create a line from (0,0) to (10,10)
	line := geom.NewLineStringXY(0, 0, 5, 5, 10, 10)

	// Scale by 2 in both directions
	scale := transform.NewAffineScale(2, 2)

	// Transform
	result, _ := transform.TransformGeometry(scale, line)
	resultLine := result.(*geom.LineString)

	coords := resultLine.Coordinates()
	for i, c := range coords {
		fmt.Printf("Point %d: (%.0f, %.0f)\n", i+1, c.X, c.Y)
	}

	// Output:
	// Point 1: (0, 0)
	// Point 2: (10, 10)
	// Point 3: (20, 20)
}
