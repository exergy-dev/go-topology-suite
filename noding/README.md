# Noding Package

The `noding` package provides robust segment noding functionality for the Go Topology Suite. Noding is the process of finding all intersection points between line segments and splitting those segments at the intersection points, ensuring topological consistency.

## Overview

Segment noding is a critical component for overlay operations (union, intersection, difference) in computational geometry. It ensures that:

1. All intersection points between segments are found
2. Segments are properly split at those intersection points
3. The result maintains topological consistency

## Core Types

### SegmentString

A `SegmentString` represents a sequence of line segments defined by a coordinate sequence. It can carry arbitrary context data.

```go
coords := geom.NewCoordinateSequenceXY(0, 0, 10, 10, 20, 0)
ss := noding.NewSegmentString(coords, "my-segment")
```

### NodedSegmentString

A `NodedSegmentString` extends `SegmentString` and tracks nodes (intersection points) that have been added. After noding, you can retrieve the modified coordinate sequence with all nodes inserted.

```go
nss := noding.NewNodedSegmentString(coords, nil)
// ... noding happens ...
nodedCoords := nss.NodedCoordinates()  // Coords with intersections inserted
```

### Noder Interface

The `Noder` interface defines the contract for noding algorithms:

```go
type Noder interface {
    ComputeNodes(segmentStrings []*NodedSegmentString)
    GetNodedSubstrings() []*NodedSegmentString
}
```

## Implementations

### SimpleNoder

A basic O(n²) noder that compares every segment with every other segment. Suitable for small datasets.

```go
adder := noding.NewIntersectionAdder()
noder := noding.NewSimpleNoder(adder)
noder.ComputeNodes(segmentStrings)
result := noder.GetNodedSubstrings()
```

### ScaledNoder

Wraps another noder and applies coordinate scaling for improved numerical robustness:

```go
baseNoder := noding.NewSimpleNoder(noding.NewIntersectionAdder())
scaledNoder := noding.NewScaledNoder(baseNoder, 1000.0)  // Scale by 1000
scaledNoder.ComputeNodes(segmentStrings)
```

### ValidatingNoder

Validates that noding is complete (no intersections remain):

```go
baseNoder := noding.NewSimpleNoder(noding.NewIntersectionAdder())
validatingNoder := noding.NewValidatingNoder(baseNoder)
validatingNoder.ComputeNodes(segmentStrings)
```

### IteratedNoder

Runs noding multiple times until no more intersections are found:

```go
baseNoder := noding.NewSimpleNoder(noding.NewIntersectionAdder())
iteratedNoder := noding.NewIteratedNoder(baseNoder, 5)  // Max 5 iterations
iteratedNoder.ComputeNodes(segmentStrings)
```

## SegmentIntersector Implementations

### IntersectionAdder

Finds all intersections and adds them as nodes to the segment strings. This is the most common use case for overlay operations.

```go
adder := noding.NewIntersectionAdder()
noder := noding.NewSimpleNoder(adder)
noder.ComputeNodes(segmentStrings)

fmt.Printf("Found %d proper intersections\n", adder.ProperIntersectionCount())
```

### IntersectionCounter

Simply counts intersections without modifying the segment strings:

```go
counter := noding.NewIntersectionCounter()
noder := noding.NewSimpleNoder(counter)
noder.ComputeNodes(segmentStrings)

fmt.Printf("Total intersections: %d\n", counter.Count())
```

### IntersectionFinderAdder

Finds and records interior intersections while also adding them as nodes:

```go
finder := noding.NewIntersectionFinderAdder()
noder := noding.NewSimpleNoder(finder)
noder.ComputeNodes(segmentStrings)

intersections := finder.Intersections()
```

## Usage Examples

### Basic Noding

```go
package main

import (
    "fmt"
    "github.com/robert-malhotra/go-topology-suite/geom"
    "github.com/robert-malhotra/go-topology-suite/noding"
)

func main() {
    // Create two crossing lines
    line1 := noding.NewNodedSegmentString(
        geom.NewCoordinateSequenceXY(0, 0, 10, 10),
        "line1",
    )
    line2 := noding.NewNodedSegmentString(
        geom.NewCoordinateSequenceXY(0, 10, 10, 0),
        "line2",
    )

    // Compute nodes
    noder := noding.NewSimpleNoder(noding.NewIntersectionAdder())
    noder.ComputeNodes([]*noding.NodedSegmentString{line1, line2})

    // Get noded result
    nodedStrings := noder.GetNodedSubstrings()

    for _, nss := range nodedStrings {
        coords := nss.Coordinates()
        // Each line now has 3 coordinates: start, intersection, end
        fmt.Printf("Noded segment: %d coordinates\n", len(coords))
    }
}
```

### Counting Intersections

```go
// Create a grid of lines
var segments []*noding.NodedSegmentString

// Add horizontal lines
for y := 0; y < 5; y++ {
    seg := noding.NewNodedSegmentString(
        geom.NewCoordinateSequenceXY(0, float64(y), 4, float64(y)),
        nil,
    )
    segments = append(segments, seg)
}

// Add vertical lines
for x := 0; x < 5; x++ {
    seg := noding.NewNodedSegmentString(
        geom.NewCoordinateSequenceXY(float64(x), 0, float64(x), 4),
        nil,
    )
    segments = append(segments, seg)
}

// Count intersections
counter := noding.NewIntersectionCounter()
noder := noding.NewSimpleNoder(counter)
noder.ComputeNodes(segments)

fmt.Printf("Grid has %d intersections\n", counter.Count())
```

### Working with Closed Rings

```go
// Create a closed ring (triangle)
ring := noding.NewNodedSegmentString(
    geom.NewCoordinateSequenceXY(0, 0, 10, 0, 5, 8, 0, 0),
    "triangle",
)

// Create a line crossing the ring
line := noding.NewNodedSegmentString(
    geom.NewCoordinateSequenceXY(0, 4, 10, 4),
    "line",
)

// Node them
adder := noding.NewIntersectionAdder()
noder := noding.NewSimpleNoder(adder)
noder.ComputeNodes([]*noding.NodedSegmentString{ring, line})

// The noder properly handles closed rings and only finds
// proper intersections, not endpoint coincidences
fmt.Printf("Intersections: %d\n", adder.ProperIntersectionCount())
```

## Design Decisions

### Adjacent Segment Handling

The implementation automatically skips testing adjacent segments within the same segment string, as they always share an endpoint. For closed rings, it also skips testing the first and last segments.

### Proper vs. Non-Proper Intersections

- **Proper intersection**: Segments cross in their interiors
- **Non-proper intersection**: Segments touch at endpoints or are collinear

The `IntersectionAdder` adds nodes for all intersections but tracks proper intersections separately.

### Node Ordering

When multiple nodes are added to a segment, they are automatically sorted by their parameter value (position along the segment) when generating noded coordinates.

### Context Preservation

Each `SegmentString` can carry arbitrary context data that is preserved through the noding process. This is useful for tracking which original geometry a segment came from.

## Performance Considerations

### SimpleNoder Performance

- **Time Complexity**: O(n²) where n is the number of segments
- **Space Complexity**: O(n + k) where k is the number of intersections
- **Best for**: Small to medium datasets (< 1000 segments)

### Optimization Strategies

For large datasets, consider:

1. **Spatial indexing**: Use an R-tree or STR-tree to avoid comparing distant segments
2. **Envelope filtering**: Check bounding boxes before detailed intersection tests
3. **Scaling**: Use `ScaledNoder` to improve numerical robustness
4. **Iteration**: Use `IteratedNoder` to handle numerical precision issues

## Testing

The package includes comprehensive tests:

```bash
# Run all tests
go test ./noding/...

# Run with verbose output
go test ./noding/... -v

# Run benchmarks
go test ./noding/... -bench=. -benchmem

# Run examples
go test ./noding/... -run Example -v
```

## Future Enhancements

Potential improvements for future versions:

1. **MCIndexNoder**: Spatial index-based noder for better performance
2. **SnapRoundingNoder**: Snap coordinates to a grid for robustness
3. **Parallel noding**: Concurrent processing for large datasets
4. **Incremental noding**: Add segments one at a time
5. **Chain-based noding**: Optimize for long chains of segments

## References

- [JTS Topology Suite Noding Package](https://github.com/locationtech/jts/tree/master/modules/core/src/main/java/org/locationtech/jts/noding)
- [GEOS Noding](https://github.com/libgeos/geos/tree/main/src/noding)
- "Computational Geometry: Algorithms and Applications" by de Berg et al.

## License

Part of the Go Topology Suite project.
