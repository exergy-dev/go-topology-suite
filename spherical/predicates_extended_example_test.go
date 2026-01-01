package spherical_test

import (
	"fmt"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/spherical"
)

// ExampleGenericWithin demonstrates checking if one geometry is completely within another
// using spherical geometry calculations.
func ExampleGenericWithin() {
	// Create a small polygon
	smallPolygon := geom.NewPolygon(
		geom.NewLinearRing(geom.CoordinateSequence{
			{X: -0.5, Y: -0.5},
			{X: 0.5, Y: -0.5},
			{X: 0.5, Y: 0.5},
			{X: -0.5, Y: 0.5},
			{X: -0.5, Y: -0.5},
		}),
		nil,
	)

	// Create a larger polygon that contains the small one
	largePolygon := geom.NewPolygon(
		geom.NewLinearRing(geom.CoordinateSequence{
			{X: -1, Y: -1},
			{X: 1, Y: -1},
			{X: 1, Y: 1},
			{X: -1, Y: 1},
			{X: -1, Y: -1},
		}),
		nil,
	)

	// Check if small polygon is within large polygon
	isWithin := spherical.GenericWithin(smallPolygon, largePolygon)
	fmt.Printf("Small polygon within large polygon: %v\n", isWithin)

	// Output:
	// Small polygon within large polygon: true
}

// ExampleGenericDisjoint demonstrates checking if two geometries have no points in common
// using spherical geometry calculations.
func ExampleGenericDisjoint() {
	// Create two separate polygons
	polygon1 := geom.NewPolygon(
		geom.NewLinearRing(geom.CoordinateSequence{
			{X: 0, Y: 0},
			{X: 1, Y: 0},
			{X: 1, Y: 1},
			{X: 0, Y: 1},
			{X: 0, Y: 0},
		}),
		nil,
	)

	polygon2 := geom.NewPolygon(
		geom.NewLinearRing(geom.CoordinateSequence{
			{X: 2, Y: 2},
			{X: 3, Y: 2},
			{X: 3, Y: 3},
			{X: 2, Y: 3},
			{X: 2, Y: 2},
		}),
		nil,
	)

	// Check if polygons are disjoint
	areDisjoint := spherical.GenericDisjoint(polygon1, polygon2)
	fmt.Printf("Polygons are disjoint: %v\n", areDisjoint)

	// Output:
	// Polygons are disjoint: true
}

// ExampleGenericOverlaps demonstrates checking if two geometries of the same dimension
// intersect but neither contains the other, using spherical geometry.
func ExampleGenericOverlaps() {
	// Create two overlapping polygons
	polygon1 := geom.NewPolygon(
		geom.NewLinearRing(geom.CoordinateSequence{
			{X: 0, Y: 0},
			{X: 2, Y: 0},
			{X: 2, Y: 2},
			{X: 0, Y: 2},
			{X: 0, Y: 0},
		}),
		nil,
	)

	polygon2 := geom.NewPolygon(
		geom.NewLinearRing(geom.CoordinateSequence{
			{X: 1, Y: 1},
			{X: 3, Y: 1},
			{X: 3, Y: 3},
			{X: 1, Y: 3},
			{X: 1, Y: 1},
		}),
		nil,
	)

	// Check if polygons overlap
	doOverlap := spherical.GenericOverlaps(polygon1, polygon2)
	fmt.Printf("Polygons overlap: %v\n", doOverlap)

	// Output:
	// Polygons overlap: true
}

// ExampleGenericTouches demonstrates checking if two geometries touch at their boundaries
// but have no interior points in common, using spherical geometry.
func ExampleGenericTouches() {
	// Create two adjacent polygons that share an edge
	polygon1 := geom.NewPolygon(
		geom.NewLinearRing(geom.CoordinateSequence{
			{X: 0, Y: 0},
			{X: 1, Y: 0},
			{X: 1, Y: 1},
			{X: 0, Y: 1},
			{X: 0, Y: 0},
		}),
		nil,
	)

	polygon2 := geom.NewPolygon(
		geom.NewLinearRing(geom.CoordinateSequence{
			{X: 1, Y: 0},
			{X: 2, Y: 0},
			{X: 2, Y: 1},
			{X: 1, Y: 1},
			{X: 1, Y: 0},
		}),
		nil,
	)

	// Check if polygons touch
	doTouch := spherical.GenericTouches(polygon1, polygon2)
	fmt.Printf("Polygons touch: %v\n", doTouch)

	// Output:
	// Polygons touch: true
}

// ExampleGenericWithin_multiGeometry demonstrates using Within with MultiGeometry types.
func ExampleGenericWithin_multiGeometry() {
	// Create a MultiPoint with several points
	multiPoint := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(0, 0),
		geom.NewPoint(0.5, 0.5),
		geom.NewPoint(-0.5, -0.5),
	})

	// Create a polygon that contains all points
	polygon := geom.NewPolygon(
		geom.NewLinearRing(geom.CoordinateSequence{
			{X: -1, Y: -1},
			{X: 1, Y: -1},
			{X: 1, Y: 1},
			{X: -1, Y: 1},
			{X: -1, Y: -1},
		}),
		nil,
	)

	// Check if all points are within the polygon
	allWithin := spherical.GenericWithin(multiPoint, polygon)
	fmt.Printf("All points within polygon: %v\n", allWithin)

	// Output:
	// All points within polygon: true
}

// ExampleGenericDisjoint_realWorld demonstrates using Disjoint with real-world coordinates.
func ExampleGenericDisjoint_realWorld() {
	// New York City approximate bounding box (lon, lat)
	nyc := geom.NewPolygon(
		geom.NewLinearRing(geom.CoordinateSequence{
			{X: -74.05, Y: 40.68},  // Southwest
			{X: -73.75, Y: 40.68},  // Southeast
			{X: -73.75, Y: 40.88},  // Northeast
			{X: -74.05, Y: 40.88},  // Northwest
			{X: -74.05, Y: 40.68},  // Close
		}),
		nil,
	)

	// London approximate bounding box (lon, lat)
	london := geom.NewPolygon(
		geom.NewLinearRing(geom.CoordinateSequence{
			{X: -0.3, Y: 51.4},  // Southwest
			{X: 0.1, Y: 51.4},   // Southeast
			{X: 0.1, Y: 51.6},   // Northeast
			{X: -0.3, Y: 51.6},  // Northwest
			{X: -0.3, Y: 51.4},  // Close
		}),
		nil,
	)

	// Check if NYC and London are disjoint (they should be!)
	areDisjoint := spherical.GenericDisjoint(nyc, london)
	fmt.Printf("NYC and London are disjoint: %v\n", areDisjoint)

	// Output:
	// NYC and London are disjoint: true
}
