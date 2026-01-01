package kdtree_test

import (
	"fmt"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/index/kdtree"
)

// Example demonstrates basic KD-tree usage
func Example() {
	tree := kdtree.New()

	// Insert points
	tree.InsertXY(0, 0, "Origin")
	tree.InsertXY(10, 10, "Point A")
	tree.InsertXY(20, 5, "Point B")

	// Find nearest neighbor
	nearest := tree.NearestNeighbor(11, 11)
	fmt.Println("Nearest to (11,11):", nearest)

	// Query a region
	env := geom.NewEnvelope(0, 0, 15, 15)
	results := tree.Query(env)
	fmt.Println("Points in region:", len(results))

	// Output:
	// Nearest to (11,11): Point A
	// Points in region: 2
}

// ExampleKDTree_NearestK demonstrates finding k nearest neighbors
func ExampleKDTree_NearestK() {
	tree := kdtree.New()

	// Insert points in a grid
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			tree.InsertXY(float64(i*10), float64(j*10), fmt.Sprintf("P(%d,%d)", i, j))
		}
	}

	// Find 3 nearest neighbors to (15, 15)
	nearest := tree.NearestK(15, 15, 3)
	fmt.Println("Found", len(nearest), "nearest neighbors")

	// Output:
	// Found 3 nearest neighbors
}

// ExampleKDTree_QueryRadius demonstrates radius-based queries
func ExampleKDTree_QueryRadius() {
	tree := kdtree.New()

	// Insert city locations (simplified coordinates)
	tree.InsertXY(0, 0, "City Center")
	tree.InsertXY(5, 0, "East District")
	tree.InsertXY(0, 5, "North District")
	tree.InsertXY(20, 20, "Distant Town")

	// Find all locations within 10 units of city center
	nearby := tree.QueryRadius(0, 0, 10)
	fmt.Println("Locations within 10 units:", len(nearby))

	// Output:
	// Locations within 10 units: 3
}

// ExampleKDTree_InsertGeometry shows how to index geometries
func ExampleKDTree_InsertGeometry() {
	tree := kdtree.New()
	factory := geom.DefaultFactory

	// Create and insert point geometries
	p1 := factory.CreatePoint(10, 10)
	p2 := factory.CreatePoint(20, 20)

	tree.InsertGeometry(p1)
	tree.InsertGeometry(p2)

	fmt.Println("Tree size:", tree.Size())

	// Output:
	// Tree size: 2
}
