# KD-Tree Spatial Index

A K-dimensional tree implementation optimized for 2D point data, providing efficient nearest neighbor and range queries.

## Features

- **Nearest Neighbor Search**: Find the closest point to a query location in O(log n) average time
- **K-Nearest Neighbors**: Find the k closest points efficiently
- **Range Queries**: Find all points within a rectangular region
- **Radius Queries**: Find all points within a circular radius
- **Incremental Construction**: No build phase required - tree is constructed during insertion
- **High Performance**: Optimized for point data with envelope-based pruning

## Installation

```go
import "github.com/robert-malhotra/go-topology-suite/index/kdtree"
```

## Quick Start

```go
// Create a new KD-tree
tree := kdtree.New()

// Insert points
tree.InsertXY(0, 0, "Origin")
tree.InsertXY(10, 10, "Point A")
tree.InsertXY(20, 5, "Point B")

// Find nearest neighbor
nearest := tree.NearestNeighbor(11, 11)
// Returns: "Point A"

// Find 3 nearest neighbors
nearestK := tree.NearestK(0, 0, 3)

// Query a rectangular region
env := geom.NewEnvelope(0, 0, 15, 15)
results := tree.Query(env)

// Query points within a radius
nearby := tree.QueryRadius(10, 10, 5.0)
```

## API Reference

### Construction

- `New() *KDTree` - Create a new empty KD-tree

### Insertion

- `Insert(coord Coordinate, data interface{})` - Insert a point with associated data
- `InsertXY(x, y float64, data interface{})` - Insert a point by X,Y coordinates
- `InsertGeometry(g Geometry)` - Insert a geometry using its centroid

### Queries

- `Query(envelope *Envelope) []interface{}` - Find all points in a rectangular region
- `QueryPoint(x, y float64) []interface{}` - Find points at exact location
- `QueryRadius(x, y, radius float64) []interface{}` - Find points within circular radius
- `QueryGeometry(g Geometry) []interface{}` - Find points within geometry's envelope

### Nearest Neighbor

- `NearestNeighbor(x, y float64) interface{}` - Find the single nearest point
- `NearestNeighborCoord(coord Coordinate) interface{}` - Find nearest using Coordinate
- `NearestK(x, y float64, k int) []interface{}` - Find k nearest neighbors (ordered by distance)

### Utilities

- `Size() int` - Get number of points in tree
- `IsEmpty() bool` - Check if tree is empty
- `Depth() int` - Get maximum depth of tree
- `Envelope() *Envelope` - Get bounding box of all points
- `Items() []interface{}` - Get all data items
- `Visit(visitor func(Coordinate, interface{}) bool)` - Traverse all points
- `Clear()` - Remove all points
- `Remove(coord Coordinate, data interface{}) bool` - Remove a point (rebuilds tree)

## Performance

Benchmarks on Intel Core i9-14900KF:

```
BenchmarkInsert-32              119989 ns/op      119 B/op       2 allocs/op
BenchmarkQuerySmallRegion-32     14902 ns/op    18800 B/op      10 allocs/op
BenchmarkNearestNeighbor-32       1743 ns/op        0 B/op       0 allocs/op
BenchmarkNearestK-32              5502 ns/op     4176 B/op     121 allocs/op
BenchmarkQueryRadius-32          16.85 ns/op       32 B/op       1 allocs/op
```

## When to Use

### Use KD-Tree When:

- Working primarily with **point data**
- **Nearest neighbor queries** are important
- Data is relatively **static** (infrequent insertions/deletions)
- Points are **well-distributed** in space

### Use STR-Tree Instead When:

- Working with various **geometry types** (polygons, lines, etc.)
- All data is available **upfront** (bulk loading)
- Need **envelope-based queries** only

### Use Quadtree Instead When:

- Need **frequent insertions and deletions**
- Data has **unknown or dynamic bounds**
- Working with **heterogeneous geometry types**

## Algorithm Details

### Tree Structure

The KD-tree recursively partitions space by alternating split axes:

- **Level 0** (root): splits on X axis
- **Level 1**: splits on Y axis
- **Level 2**: splits on X axis
- And so on...

Each node stores:
- A point (coordinate)
- Associated data
- Bounding envelope (for optimization)
- Left and right subtree pointers

### Query Optimization

Each node maintains a bounding envelope, allowing entire subtrees to be pruned during queries if they cannot possibly contain matching points.

### Complexity

- **Insert**: O(log n) average, O(n) worst case
- **Range Query**: O(√n + k) where k = results
- **Nearest Neighbor**: O(log n) average case
- **K-Nearest**: O(log n + k) average case
- **Space**: O(n)

### Balancing

The tree is built incrementally without explicit balancing. For best performance:

- Insert points in **random order** (produces balanced tree)
- Avoid inserting in **sorted order** (produces degenerate tree)
- Consider rebuilding if tree becomes unbalanced

## Thread Safety

**KD-tree is NOT thread-safe.** Use external synchronization for concurrent access:

```go
var mu sync.RWMutex
tree := kdtree.New()

// For queries (read-only):
mu.RLock()
results := tree.Query(env)
mu.RUnlock()

// For insertions:
mu.Lock()
tree.Insert(coord, data)
mu.Unlock()
```

## Examples

### Nearest Neighbor Search

```go
tree := kdtree.New()

// Insert cities
tree.InsertXY(40.7128, -74.0060, "New York")
tree.InsertXY(34.0522, -118.2437, "Los Angeles")
tree.InsertXY(41.8781, -87.6298, "Chicago")

// Find nearest city to coordinates
nearest := tree.NearestNeighbor(40.0, -75.0)
fmt.Println(nearest) // "New York"
```

### K-Nearest Neighbors

```go
tree := kdtree.New()

// Insert weather stations
for _, station := range stations {
    tree.InsertXY(station.Lat, station.Lon, station)
}

// Find 5 nearest stations to a location
nearest := tree.NearestK(userLat, userLon, 5)
```

### Radius Search

```go
tree := kdtree.New()

// Insert points of interest
tree.InsertXY(0, 0, "Restaurant A")
tree.InsertXY(0.5, 0.5, "Restaurant B")
tree.InsertXY(10, 10, "Restaurant C")

// Find restaurants within 1km radius
nearby := tree.QueryRadius(0, 0, 1.0)
// Returns: ["Restaurant A", "Restaurant B"]
```

### Geometry Integration

```go
tree := kdtree.New()
factory := geom.DefaultFactory

// Index point geometries
points := []geom.Geometry{
    factory.CreatePoint(10, 10),
    factory.CreatePoint(20, 20),
    factory.CreatePoint(30, 30),
}

for _, pt := range points {
    tree.InsertGeometry(pt)
}

// Query using a polygon's envelope
poly := factory.CreatePolygon(...)
results := tree.QueryGeometry(poly)
```

## Testing

Run tests with coverage:

```bash
go test ./index/kdtree/... -cover
```

Run benchmarks:

```bash
go test ./index/kdtree/... -bench=. -benchmem
```

## References

- [KD-tree on Wikipedia](https://en.wikipedia.org/wiki/K-d_tree)
- [JTS Topology Suite](https://github.com/locationtech/jts)
- [Computational Geometry: Algorithms and Applications](https://www.cs.uu.nl/geobook/)

## License

See the main GTS repository for license information.
