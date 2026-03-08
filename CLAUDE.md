# Claude Context: Go Topology Suite (GTS)

## Project Overview

Go Topology Suite (GTS) is a computational geometry library for Go, providing a native Go implementation of the functionality found in the Java Topology Suite (JTS). The library enables creation, manipulation, and analysis of 2D vector geometries according to the OGC Simple Features Specification.

**Primary Goal**: Provide a pure Go implementation of robust geometric operations without requiring C bindings (like GEOS), while maintaining compatibility with standard spatial data formats and operations.

**Target Users**: 
- GIS application developers
- Spatial database implementers
- Geospatial data processing pipelines
- Mapping and visualization applications
- Location-based services

## Core Philosophy

### Go Idioms Over Java Patterns
- **Composition over inheritance**: Use embedded structs and interfaces instead of class hierarchies
- **Explicit error handling**: Return errors instead of panics or exceptions
- **Interface segregation**: Small, focused interfaces rather than large monolithic ones
- **Concurrency-aware**: Design for safe concurrent usage where possible

### Precision and Robustness
- **Numerical robustness**: Handle edge cases in geometric computations (collinear points, nearly coincident vertices, etc.)
- **Precision models**: Support different precision requirements (floating, fixed, single precision)
- **Topology validation**: Ensure geometric validity according to OGC specifications

### Performance Considerations
- **Spatial indexing**: Essential for scalability with large datasets
- **Memory efficiency**: Coordinate sequence reuse, object pooling where appropriate
- **Lazy evaluation**: Defer expensive computations until needed
- **Benchmark-driven**: Profile and optimize hot paths

## Architecture

### Package Structure

```
gts/
├── geom/              # Core geometry types and interfaces
│   ├── coordinate.go  # Coordinate and CoordinateSequence
│   ├── geometry.go    # Base Geometry interface
│   ├── point.go       # Point implementation
│   ├── linestring.go  # LineString implementation
│   ├── polygon.go     # Polygon and LinearRing
│   ├── multi.go       # Multi* geometries
│   ├── collection.go  # GeometryCollection
│   └── envelope.go    # Bounding box
│
├── algorithm/         # Geometric algorithms
│   ├── distance.go    # Distance calculations
│   ├── locate.go      # Point location algorithms
│   ├── angle.go       # Angle and orientation
│   ├── area.go        # Area and centroid
│   ├── convexhull.go  # Convex hull computation
│   ├── simplify.go    # Douglas-Peucker and others
│   └── intersection.go # Line intersection
│
├── operation/         # High-level operations
│   ├── overlay/       # Intersection, union, difference
│   ├── buffer/        # Buffer operations
│   ├── relate/        # DE-9IM and spatial predicates
│   ├── valid/         # Geometry validation
│   ├── polygonize/    # Polygonization from lines
│   └── linemerge/     # Line merging
│
├── index/             # Spatial indexes
│   ├── strtree/       # STR-tree implementation
│   ├── quadtree/      # Quadtree implementation
│   └── kdtree/        # KD-tree for point data
│
├── io/                # Input/output formats
│   ├── wkt/           # Well-Known Text
│   ├── wkb/           # Well-Known Binary
│   └── geojson/       # GeoJSON format
│
├── precision/         # Precision models
│   └── precision.go   # Floating, fixed precision
│
├── noding/            # Line segment handling
│   ├── noder.go       # Node and segment intersection
│   └── snap.go        # Snapping operations
│
└── planar/            # Planar graph structures
    └── graph.go       # Graph data structures
```

### Key Design Patterns

#### Geometry Interface Hierarchy

```go
// Base interface - all geometries implement this
type Geometry interface {
    // Fundamental operations that every geometry must support
}

// Specific types embed base behavior
type Point struct {
    coord Coordinate
    srid  int
    // Point-specific fields
}

// Methods on concrete types
func (p *Point) GeometryType() string { return "Point" }
```

#### Factory Pattern for Construction

```go
// Factories ensure valid construction
type GeometryFactory struct {
    precisionModel PrecisionModel
    srid          int
}

func (gf *GeometryFactory) CreatePoint(coord Coordinate) *Point {
    gf.precisionModel.MakePrecise(&coord)
    return &Point{coord: coord, srid: gf.srid}
}
```

#### Visitor Pattern for Operations

```go
// For operations that need to traverse geometry collections
type GeometryVisitor interface {
    VisitPoint(p *Point)
    VisitLineString(ls *LineString)
    VisitPolygon(poly *Polygon)
    // ... other types
}
```

## Key Concepts

### Coordinates and Coordinate Sequences

**Coordinate**: Basic building block, typically (X, Y) with optional Z and M dimensions.

**CoordinateSequence**: Ordered list of coordinates. Must be efficient as it's used extensively.

```go
type Coordinate struct {
    X, Y float64
    Z    *float64 // Optional for 3D
    M    *float64 // Optional measure value
}

type CoordinateSequence []Coordinate
```

**Design Decision**: Use slice by default. Consider implementing an interface if we need packed arrays or other representations.

### Envelope (Bounding Box)

Essential for spatial indexing and quick rejection tests.

```go
type Envelope struct {
    MinX, MinY, MaxX, MaxY float64
}

func (e *Envelope) Intersects(other *Envelope) bool {
    // Quick overlap test
}
```

### Spatial Predicates (DE-9IM Model)

The Dimensionally Extended Nine-Intersection Model defines spatial relationships:
- **Equals**: Geometrically equal
- **Disjoint**: No points in common
- **Intersects**: Share at least one point
- **Touches**: Share boundary but not interior
- **Crosses**: Intersect with different dimensions
- **Within**: All points of A are in B
- **Contains**: All points of B are in A
- **Overlaps**: Share some but not all points

**Implementation Strategy**: Use the `relate` operation to compute the intersection matrix once, then derive all predicates from it.

### Precision Model

Controls coordinate precision and handles floating-point issues.

**Floating Precision**: Full double precision (default)
**Fixed Precision**: Snap to grid (e.g., 1mm resolution)
**Single Precision**: Use float32 for memory savings

### Topology Rules

**Valid Polygon**:
- Exterior ring must be clockwise
- Holes must be counter-clockwise
- Rings must be closed (first = last coordinate)
- Rings must be simple (no self-intersections)
- Holes must be inside exterior

**Valid LineString**:
- At least 2 points
- Simple = no self-intersections

**Valid LinearRing**:
- At least 4 points (including closure)
- First point equals last point
- No self-intersections

## Implementation Priorities

### Phase 1: Core Geometry (MVP)
1. ✓ Coordinate and CoordinateSequence
2. ✓ Envelope
3. ✓ Point, LineString, Polygon
4. ✓ Basic WKT reader/writer
5. ✓ Basic spatial predicates (intersects, contains)

### Phase 2: Essential Operations
1. Buffer operation
2. Intersection and union
3. Distance calculations
4. Simple validation
5. WKB support

### Phase 3: Advanced Features
1. Full overlay operations (difference, symdifference)
2. DE-9IM relate operation
3. Convex hull
4. Spatial indexes (STRtree)
5. GeoJSON support

### Phase 4: Optimization & Polish
1. Performance optimization
2. Precision model refinements
3. Topology validation and repair
4. Advanced algorithms (triangulation, Voronoi)
5. Complete test suite

## Critical Implementation Details

### Orientation and Winding Order

**Critical**: Polygon ring orientation affects area calculation and point-in-polygon tests.

```go
// Compute signed area to determine orientation
func SignedArea(ring LinearRing) float64 {
    // Shoelace formula
    // Positive = counter-clockwise
    // Negative = clockwise
}
```

**OGC Standard**: Exterior ring is counter-clockwise, holes are clockwise.
**Note**: Some systems use opposite convention - support both via options.

### Robustness in Line Intersection

Floating-point arithmetic can cause issues:
- Nearly parallel lines
- T-junctions
- Numerical errors accumulating

**Solution**: Use robust predicates from Shewchuk's adaptive precision arithmetic, or implement snap rounding.

### Buffer Operation Complexity

Buffer is one of the most complex operations:
1. Compute offset curves
2. Handle end cap styles (round, square, flat)
3. Handle join styles (round, mitre, bevel)
4. Resolve self-intersections
5. Construct valid polygon from segments

**Approach**: Use noding framework to handle segment intersections, then polygonize.

### Overlay Operations (Boolean Operations)

Most complex part of the library:
1. Node all line segments (find intersections)
2. Build planar graph
3. Label graph components (inside/outside)
4. Extract relevant portions based on operation
5. Construct output geometry

**Key Challenge**: Handling precision and topology correctly.

**Implementation Strategy**: 
- Use proven noding algorithms
- Implement snap rounding for robustness
- Extensive test suite based on JTS tests

## Testing Strategy

### Unit Tests
- Test each geometry type independently
- Test each algorithm with known inputs/outputs
- Test edge cases (empty geometries, degenerate cases)

### Property-Based Testing
```go
// Example: Buffer should always increase area
func TestBufferIncreasesArea(t *testing.T) {
    quick.Check(func(poly Polygon, dist float64) bool {
        if dist <= 0 { return true }
        buffered := poly.Buffer(dist)
        return Area(buffered) >= Area(poly)
    })
}
```

### Validation Against JTS
- Port JTS test suite
- Compare outputs for identical inputs
- Document intentional differences

### Benchmark Suite
```go
func BenchmarkPolygonIntersection(b *testing.B) {
    poly1 := createComplexPolygon(1000)
    poly2 := createComplexPolygon(1000)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = poly1.Intersection(poly2)
    }
}
```

### Fuzz Testing
- Generate random geometries
- Test operations don't panic
- Validate invariants hold

## Common Pitfalls

### Float Comparison
❌ **Wrong**: `if coord1.X == coord2.X`
✓ **Right**: `if math.Abs(coord1.X - coord2.X) < epsilon`

### Ring Closure
Always verify LinearRing is closed:
```go
func (lr *LinearRing) IsValid() bool {
    if len(lr.coords) < 4 { return false }
    first := lr.coords[0]
    last := lr.coords[len(lr.coords)-1]
    return first.Equals2D(last, epsilon)
}
```

### Nil Geometry Handling
Handle empty/nil geometries gracefully:
```go
func (g *Geometry) IsEmpty() bool {
    return g == nil || len(g.Coordinates()) == 0
}
```

### SRID Consistency
Operations between geometries should check SRID compatibility:
```go
func (g1 *Geometry) Intersects(g2 *Geometry) (bool, error) {
    if g1.SRID() != g2.SRID() && g1.SRID() != 0 && g2.SRID() != 0 {
        return false, ErrSRIDMismatch
    }
    // ... perform operation
}
```

## Performance Optimization Checklist

- [ ] Profile with pprof before optimizing
- [ ] Use spatial index for large datasets
- [ ] Consider coordinate sequence pooling
- [ ] Benchmark memory allocations
- [ ] Cache expensive computations (envelopes, areas)
- [ ] Use SIMD for vector operations where available
- [ ] Parallelize independent operations
- [ ] Minimize interface conversions in hot paths

## External Dependencies Policy

**Principle**: Minimize external dependencies to reduce maintenance burden.

**Allowed**:
- Standard library only for core functionality
- Well-maintained spatial libraries for optional integrations

**Optional Dependencies**:
- `github.com/stretchr/testify` - Testing utilities
- `github.com/golang/geo/s2` - Optional S2 geometry integration
- `github.com/paulmach/orb` - Optional compatibility layer

## API Stability

**v0.x**: Breaking changes allowed
**v1.0+**: Semantic versioning
- Major: Breaking changes
- Minor: New features, backwards compatible
- Patch: Bug fixes

**Deprecation Policy**: Mark deprecated functions for at least one minor version before removal.

## Documentation Standards

### Package Documentation
```go
// Package geom provides types and functions for representing
// and manipulating geometric objects in 2D space.
//
// The geometry model follows the OGC Simple Features Specification.
// All geometries implement the Geometry interface which provides
// standard operations like intersection, union, and spatial predicates.
package geom
```

### Function Documentation
```go
// Buffer returns a geometry representing all points within the
// given distance of this geometry. The distance may be positive
// (expand) or negative (shrink).
//
// The quality of the approximation can be controlled via BufferOp
// for more advanced use cases.
//
// Returns an error if the buffer operation fails.
func (g *Geometry) Buffer(distance float64) (Geometry, error)
```

### Example Tests
```go
func ExamplePolygon_Intersection() {
    poly1 := createSquare(0, 0, 10)
    poly2 := createSquare(5, 5, 10)
    result := poly1.Intersection(poly2)
    fmt.Println(result.Area())
    // Output: 25.0
}
```

## Resources for Contributors

### Essential Reading
1. [JTS Developer Guide](https://locationtech.github.io/jts/javadoc/)
2. [OGC Simple Features Specification](https://www.ogc.org/standards/sfa)
3. [Computational Geometry: Algorithms and Applications](https://www.cs.uu.nl/geobook/) by de Berg et al.
4. [Robust Geometric Computation](http://www.cs.cmu.edu/~quake/robust.html) by Shewchuk

### Related Projects
- **JTS**: Original Java implementation
- **GEOS**: C++ port of JTS (used by PostGIS)
- **Shapely**: Python wrapper around GEOS
- **Turf.js**: JavaScript geospatial library
- **github.com/paulmach/orb**: Alternative Go geometry library

### Useful Tools
- **QGIS**: Visualize and test geometries
- **PostGIS**: Reference implementation for operations
- **JTS TestBuilder**: GUI tool for testing JTS operations

## Contributing Guidelines

1. **Start small**: Begin with simple geometries and operations
2. **Test first**: Write tests before implementation
3. **Follow conventions**: Match existing code style
4. **Document**: Add godoc comments for all exported items
5. **Benchmark**: Add benchmarks for new algorithms
6. **Reference**: Compare against JTS behavior

## Open Questions & Design Decisions

### Mutability
**Question**: Should geometries be mutable or immutable?
**Current**: Mutable for efficiency, but consider immutable variants
**Trade-off**: Immutability is safer but may require more copying

### Error Handling
**Question**: Return errors or panic for invalid inputs?
**Current**: Return errors for user input, panic for programmer errors
**Example**: Invalid WKT → error; nil pointer dereference → panic

### 3D Support
**Question**: First-class 3D or optional Z coordinate?
**Current**: Optional Z coordinate (pointer)
**Future**: May add dedicated 3D types if needed

### Coordinate Storage
**Question**: Slice vs. packed array vs. interface?
**Current**: Slice (simple and fast)
**Future**: May add packed representation for memory efficiency

---

## Quick Start for New Contributors

1. **Setup**: Clone repo, run tests (`go test ./...`)
2. **Pick a task**: Start with a good-first-issue
3. **Read tests**: Understand expected behavior from test cases
4. **Implement**: Write code following existing patterns
5. **Test**: Add tests for new functionality
6. **Document**: Add godoc comments
7. **Submit PR**: Include description and examples

## Contact & Support

- **Issues**: GitHub issue tracker
- **Discussions**: GitHub discussions for questions
- **Slack/Discord**: [If applicable]
- **Email**: [Maintainer email]

---

*Last Updated: 2025-12-07*
*Version: 0.1.0 (Initial Design)*
