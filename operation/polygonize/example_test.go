package polygonize_test

import (
	"fmt"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/operation/polygonize"
)

// ExamplePolygonize demonstrates basic polygonization from a set of line segments.
func ExamplePolygonize() {
	// Create edges forming a rectangle
	edges := []*geom.LineString{
		mustLineStringXY(0, 0, 10, 0),   // Bottom edge
		mustLineStringXY(10, 0, 10, 10), // Right edge
		mustLineStringXY(10, 10, 0, 10), // Top edge
		mustLineStringXY(0, 10, 0, 0),   // Left edge
	}

	// Polygonize the edges
	polygons := polygonize.Polygonize(edges)

	// Print results
	fmt.Printf("Number of polygons: %d\n", len(polygons))
	if len(polygons) > 0 {
		fmt.Printf("Area of first polygon: %.1f\n", polygons[0].Area())
	}

	// Output:
	// Number of polygons: 1
	// Area of first polygon: 100.0
}

// ExamplePolygonizer demonstrates using the Polygonizer API.
func ExamplePolygonizer() {
	// Create a polygonizer
	p := polygonize.NewPolygonizer()

	// Add edges one by one
	p.Add(mustLineStringXY(0, 0, 5, 0))
	p.Add(mustLineStringXY(5, 0, 5, 5))
	p.Add(mustLineStringXY(5, 5, 0, 5))
	p.Add(mustLineStringXY(0, 5, 0, 0))

	// Get the resulting polygons
	polygons := p.GetPolygons()

	fmt.Printf("Number of polygons: %d\n", len(polygons))
	if len(polygons) > 0 {
		fmt.Printf("Area: %.1f\n", polygons[0].Area())
	}

	// Output:
	// Number of polygons: 1
	// Area: 25.0
}

// ExamplePolygonize_triangle demonstrates polygonizing a triangle.
func ExamplePolygonize_triangle() {
	// Create edges forming a triangle
	edges := []*geom.LineString{
		mustLineStringXY(0, 0, 10, 0),
		mustLineStringXY(10, 0, 5, 8),
		mustLineStringXY(5, 8, 0, 0),
	}

	polygons := polygonize.Polygonize(edges)

	fmt.Printf("Number of polygons: %d\n", len(polygons))
	fmt.Printf("Is valid: %v\n", polygons[0].IsValid())

	// Output:
	// Number of polygons: 1
	// Is valid: true
}
