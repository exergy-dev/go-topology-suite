package terra_test

import (
	"fmt"
	"math"

	"github.com/terra-geo/terra"
	"github.com/terra-geo/terra/buffer"
	"github.com/terra-geo/terra/crs/epsg"
	"github.com/terra-geo/terra/geojson"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/predicate"
	"github.com/terra-geo/terra/wkt"
)

// Decode a polygon from WKT, test a point against it, and re-encode the
// result of buffering it as GeoJSON. This is the canonical "first program"
// shape: parse → operate → encode.
func Example() {
	square, _ := wkt.Unmarshal("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	pt, _ := wkt.Unmarshal("POINT (5 5)")

	hit, _ := predicate.Intersects(square, pt)
	fmt.Println("intersects:", hit)
	// Output:
	// intersects: true
}

// ExampleIntersects shows the predicate API: two geometries plus optional
// kernel/precision options. CRS mismatches return terra.ErrCRSMismatch
// rather than silently coercing.
func ExampleIntersects() {
	a, _ := wkt.Unmarshal("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	b, _ := wkt.Unmarshal("POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")

	hit, err := predicate.Intersects(a, b)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("overlapping squares intersect:", hit)
	// Output:
	// overlapping squares intersect: true
}

// ExampleBuffer expands a geometry by a positive distance. Negative
// distances erode (and may produce an empty geometry on small inputs).
func ExampleBuffer() {
	pt, _ := wkt.Unmarshal("POINT (0 0)")
	disk, _ := buffer.Buffer(pt, 1.0)

	// The disk approximates a unit circle; vertex count is implementation-
	// defined but always non-empty for a positive distance buffer of a
	// non-empty geometry.
	fmt.Println("disk is empty:", disk.IsEmpty())
	// Output:
	// disk is empty: false
}

// ExampleUnmarshal parses WKT. POINT EMPTY, GEOMETRYCOLLECTION EMPTY,
// and the dimension suffixes (Z, M, ZM, glued or spaced) are all
// supported.
func ExampleUnmarshal() {
	g, err := wkt.Unmarshal("POINT Z (1 2 3)")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("layout:", g.Layout())
	// Output:
	// layout: XYZ
}

// ExampleTransform reprojects a geometry from one CRS to another.
// Terra never reprojects implicitly: predicates and overlays on
// geometries with mismatched CRS pointers return ErrCRSMismatch.
func ExampleTransform() {
	pt := geom.NewPoint(epsg.WGS84, geom.XY{X: 0, Y: 0})

	projected, err := terra.Transform(pt, epsg.WebMercator)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	xy := projected.(*geom.Point).XY()
	// (0°, 0°) maps to (0, 0) on Web Mercator.
	fmt.Printf("(%g, %g)\n", math.Abs(xy.X), math.Abs(xy.Y))
	// Output:
	// (0, 0)
}

// ExampleMarshal_geojson round-trips a geometry through GeoJSON. The
// WithForceCCW writer option enforces RFC 7946 ring orientation.
func ExampleMarshal_geojson() {
	square, _ := wkt.Unmarshal("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
	out, err := geojson.Marshal(square)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(string(out))
	// Output:
	// {"type":"Polygon","coordinates":[[[0,0],[10,0],[10,10],[0,10],[0,0]]]}
}
