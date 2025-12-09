// Package kdtree provides a K-dimensional tree spatial index optimized for point data.
//
// Overview
//
// A KD-tree (k-dimensional tree) is a space-partitioning data structure for organizing
// points in k-dimensional space. This implementation is specialized for 2D point data
// and provides excellent performance for nearest neighbor queries.
//
// Key Features:
//   - Efficient nearest neighbor search: O(log n) average case
//   - K-nearest neighbors search
//   - Range queries (rectangular regions)
//   - Radius queries (circular regions)
//   - Incremental construction (no build phase required)
//
// When to Use KD-Tree:
//
// Use a KD-tree when:
//   - Working primarily with point data
//   - Nearest neighbor queries are important
//   - Data is relatively static (infrequent insertions/deletions)
//   - Points are well-distributed in space
//
// Use STR-tree instead when:
//   - Working with various geometry types (polygons, lines, etc.)
//   - All data is available upfront (bulk loading)
//   - Need envelope-based queries only
//
// Use Quadtree instead when:
//   - Need frequent insertions and deletions
//   - Data has unknown or dynamic bounds
//   - Working with heterogeneous geometry types
//
// Performance Characteristics:
//
//   - Insert: O(log n) average, O(n) worst case (degenerate tree)
//   - Query (range): O(√n + k) where k is number of results
//   - Nearest neighbor: O(log n) average case
//   - K-nearest neighbors: O(log n + k) average case
//   - Space: O(n)
//
// The tree is built incrementally during insertion, alternating split axes at each
// level (X axis at even depths, Y axis at odd depths). This creates a balanced tree
// when insertion order is random, but may degenerate if points are inserted in sorted order.
//
// Example Usage:
//
//	tree := kdtree.New()
//
//	// Insert points
//	tree.InsertXY(0, 0, "Origin")
//	tree.InsertXY(10, 10, "Point A")
//	tree.InsertXY(20, 5, "Point B")
//
//	// Find nearest neighbor
//	nearest := tree.NearestNeighbor(11, 11)
//	fmt.Println(nearest) // "Point A"
//
//	// Find k nearest neighbors
//	nearestK := tree.NearestK(0, 0, 3)
//
//	// Query points in a region
//	env := geom.NewEnvelope(0, 0, 15, 15)
//	results := tree.Query(env)
//
//	// Query points within a radius
//	nearby := tree.QueryRadius(10, 10, 5.0)
//
// Thread Safety:
//
// KD-tree is NOT thread-safe. External synchronization is required for concurrent access.
// For read-heavy workloads, consider using a read-write mutex (sync.RWMutex) to allow
// concurrent queries while serializing insertions.
//
// Implementation Notes:
//
// The KD-tree alternates splitting dimensions at each level:
//   - Level 0 (root): splits on X axis
//   - Level 1: splits on Y axis
//   - Level 2: splits on X axis
//   - And so on...
//
// Each node stores:
//   - A point (coordinate)
//   - Associated data
//   - Bounding envelope (for query optimization)
//   - Left and right child pointers
//
// The bounding envelope at each node allows pruning of entire subtrees during
// queries, significantly improving performance for range and nearest neighbor searches.
//
// Comparison with Other Spatial Indexes:
//
// KD-Tree vs STR-Tree:
//   - KD-tree: Better for point data and nearest neighbor queries
//   - STR-tree: Better for bulk loading and mixed geometry types
//
// KD-Tree vs Quadtree:
//   - KD-tree: More efficient space partitioning, better for static data
//   - Quadtree: Better for dynamic data with frequent updates
//
// KD-Tree vs R-Tree:
//   - KD-tree: Simpler, faster for point data
//   - R-tree: Better for overlapping rectangles and complex geometries
package kdtree
