package spherical_test

import (
	"fmt"

	"github.com/go-topology-suite/gts/geom"
	"github.com/go-topology-suite/gts/spherical"
)

// ExampleCrosses demonstrates the Crosses predicate for spherical geometries.
func ExampleCrosses() {
	// Create two lines that cross
	line1 := geom.NewLineStringXY(
		-1.0, 0.0, // West to East
		1.0, 0.0,
	)
	line2 := geom.NewLineStringXY(
		0.0, -1.0, // South to North
		0.0, 1.0,
	)

	crosses := spherical.Crosses(line1, line2)
	fmt.Println("Lines cross:", crosses)

	// Line crossing through a polygon
	poly := geom.NewPolygon(geom.NewLinearRingXY(
		-0.5, -0.5,
		0.5, -0.5,
		0.5, 0.5,
		-0.5, 0.5,
		-0.5, -0.5,
	), nil)

	lineThroughPoly := geom.NewLineStringXY(
		-1.0, 0.0,
		1.0, 0.0,
	)

	crossesPoly := spherical.Crosses(lineThroughPoly, poly)
	fmt.Println("Line crosses polygon:", crossesPoly)

	// Output:
	// Lines cross: true
	// Line crosses polygon: true
}

// ExampleCovers demonstrates the Covers predicate for spherical geometries.
func ExampleCovers() {
	// Create a polygon
	poly := geom.NewPolygon(geom.NewLinearRingXY(
		0.0, 0.0,
		2.0, 0.0,
		2.0, 2.0,
		0.0, 2.0,
		0.0, 0.0,
	), nil)

	// Point inside
	pointInside := geom.NewPoint(1.0, 1.0)
	covers := spherical.Covers(poly, pointInside)
	fmt.Println("Polygon covers interior point:", covers)

	// Point on boundary (corner)
	pointOnBoundary := geom.NewPoint(0.0, 0.0)
	coversCorner := spherical.Covers(poly, pointOnBoundary)
	fmt.Println("Polygon covers corner point:", coversCorner)

	// Point outside
	pointOutside := geom.NewPoint(3.0, 3.0)
	coversOutside := spherical.Covers(poly, pointOutside)
	fmt.Println("Polygon covers outside point:", coversOutside)

	// Output:
	// Polygon covers interior point: true
	// Polygon covers corner point: true
	// Polygon covers outside point: false
}

// ExampleCoveredBy demonstrates the CoveredBy predicate for spherical geometries.
func ExampleCoveredBy() {
	// Create a polygon
	bigPoly := geom.NewPolygon(geom.NewLinearRingXY(
		0.0, 0.0,
		3.0, 0.0,
		3.0, 3.0,
		0.0, 3.0,
		0.0, 0.0,
	), nil)

	// Smaller polygon inside
	smallPoly := geom.NewPolygon(geom.NewLinearRingXY(
		1.0, 1.0,
		2.0, 1.0,
		2.0, 2.0,
		1.0, 2.0,
		1.0, 1.0,
	), nil)

	coveredBy := spherical.CoveredBy(smallPoly, bigPoly)
	fmt.Println("Small polygon covered by big polygon:", coveredBy)

	// This is equivalent to Covers(bigPoly, smallPoly)
	covers := spherical.Covers(bigPoly, smallPoly)
	fmt.Println("Big polygon covers small polygon:", covers)
	fmt.Println("CoveredBy is inverse of Covers:", coveredBy == covers)

	// Output:
	// Small polygon covered by big polygon: true
	// Big polygon covers small polygon: true
	// CoveredBy is inverse of Covers: true
}

// ExampleEquals demonstrates the Equals predicate for spherical geometries.
func ExampleEquals() {
	// Create two identical polygons
	poly1 := geom.NewPolygon(geom.NewLinearRingXY(
		0.0, 0.0,
		1.0, 0.0,
		1.0, 1.0,
		0.0, 1.0,
		0.0, 0.0,
	), nil)

	poly2 := geom.NewPolygon(geom.NewLinearRingXY(
		0.0, 0.0,
		1.0, 0.0,
		1.0, 1.0,
		0.0, 1.0,
		0.0, 0.0,
	), nil)

	equals := spherical.Equals(poly1, poly2)
	fmt.Println("Identical polygons are equal:", equals)

	// Different polygons
	poly3 := geom.NewPolygon(geom.NewLinearRingXY(
		0.0, 0.0,
		2.0, 0.0,
		2.0, 2.0,
		0.0, 2.0,
		0.0, 0.0,
	), nil)

	notEquals := spherical.Equals(poly1, poly3)
	fmt.Println("Different polygons are equal:", notEquals)

	// Output:
	// Identical polygons are equal: true
	// Different polygons are equal: false
}
