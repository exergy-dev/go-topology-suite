package buffer

import (
	"math"

	"github.com/exergy-dev/go-topology-suite/geom"
	"github.com/exergy-dev/go-topology-suite/measure"
)

// maxBufferDistanceDiffFrac is the maximum allowable fractional deviation
// between the requested buffer distance and the actual minimum/maximum
// distance from the input to the buffer boundary. JTS uses 1.2% (1%
// caused occasional false positives).
const maxBufferDistanceDiffFrac = 0.012

// distanceValidationResult captures the maximum absolute distance error
// observed between the input geometry and the buffer boundary, along
// with the witness location and a human-readable error message. ok is
// true iff the buffer distance is within tolerance.
type distanceValidationResult struct {
	ok            bool
	errorMessage  string
	errorLocation geom.XY
	maxError      float64 // signed: positive = too far, negative = too close
}

// ValidateBufferDistance verifies that the boundary of bufferGeom lies
// approximately distance away from input. Useful only for round buffers
// (round caps and joins). Works for positive and negative distances;
// negative distances are checked only when input is areal.
//
// On success returns (0, zero-XY, true). On failure returns the largest
// signed distance error observed and a witness coordinate. Empty inputs
// pass through as ok.
//
// Port of org.locationtech.jts.operation.buffer.validate.BufferDistanceValidator.
func ValidateBufferDistance(input, bufferGeom geom.Geometry, distance float64) (errorMagnitude float64, errorLocation geom.XY, ok bool) {
	r := validateBufferDistance(input, bufferGeom, distance)
	if r.ok {
		return 0, geom.XY{}, true
	}
	return r.maxError, r.errorLocation, false
}

// validateBufferDistance is the shared engine used by both the public
// helper and BufferResultValidator.
func validateBufferDistance(input, bufferGeom geom.Geometry, distance float64) distanceValidationResult {
	if input == nil || bufferGeom == nil {
		return distanceValidationResult{ok: true}
	}
	if input.IsEmpty() || bufferGeom.IsEmpty() {
		return distanceValidationResult{ok: true}
	}
	posDistance := math.Abs(distance)
	delta := maxBufferDistanceDiffFrac * posDistance
	minValid := posDistance - delta
	maxValid := posDistance + delta

	if distance > 0 {
		// Positive buffer: check input vs. buffer boundary.
		bufCurve := geometryBoundaryLines(bufferGeom)
		if bufCurve == nil || bufCurve.IsEmpty() {
			return distanceValidationResult{ok: true}
		}
		if r := checkMinimumBufferDistance(input, bufCurve, minValid); !r.ok {
			return r
		}
		return checkMaximumBufferDistance(input, bufCurve, maxValid)
	}
	// Negative buffer: only check polygonal inputs (lines/points have no
	// area to inset).
	if !isArealOrCollection(input) {
		return distanceValidationResult{ok: true}
	}
	inputCurve := geometryBoundaryLines(input)
	if inputCurve == nil || inputCurve.IsEmpty() {
		return distanceValidationResult{ok: true}
	}
	if r := checkMinimumBufferDistance(inputCurve, bufferGeom, minValid); !r.ok {
		return r
	}
	return checkMaximumBufferDistance(inputCurve, bufferGeom, maxValid)
}

// checkMinimumBufferDistance asserts that g1 and g2 are at least minDist
// apart. The closer pair becomes the witness on failure.
func checkMinimumBufferDistance(g1, g2 geom.Geometry, minDist float64) distanceValidationResult {
	d := measure.DistanceOp(g1, g2)
	if d >= minDist {
		return distanceValidationResult{ok: true}
	}
	_, p2 := measure.NearestPoints(g1, g2)
	return distanceValidationResult{
		ok:            false,
		errorMessage:  "Distance between buffer curve and input is too small",
		errorLocation: p2,
		maxError:      d - minDist, // negative
	}
}

// checkMaximumBufferDistance asserts that no point of bufCurve lies
// further than maxDist from the input. The witness is the discrete
// Hausdorff witness on bufCurve.
func checkMaximumBufferDistance(input, bufCurve geom.Geometry, maxDist float64) distanceValidationResult {
	// orientedHausdorff(bufCurve, input): the largest distance from any
	// vertex of bufCurve to its nearest point on input. JTS densifies at
	// fraction 0.25 to limit discretisation error; we approximate by
	// densifying bufCurve below.
	dense := densifyForHausdorff(bufCurve, 0.25)
	d, witness := orientedHausdorffWithWitness(dense, input)
	if d <= maxDist {
		return distanceValidationResult{ok: true}
	}
	return distanceValidationResult{
		ok:            false,
		errorMessage:  "Distance between buffer curve and input is too large",
		errorLocation: witness,
		maxError:      d - maxDist, // positive
	}
}

// orientedHausdorffWithWitness returns the directed discrete Hausdorff
// distance from a's vertices to b, plus the vertex of a that realises
// the maximum.
func orientedHausdorffWithWitness(a, b geom.Geometry) (float64, geom.XY) {
	if a == nil || b == nil || a.IsEmpty() || b.IsEmpty() {
		return 0, geom.XY{}
	}
	max := 0.0
	var witness geom.XY
	visitGeometryVertices(a, func(p geom.XY) {
		d := measure.DistanceOp(geom.NewPoint(a.CRS(), p), b)
		if d > max {
			max = d
			witness = p
		}
	})
	return max, witness
}

// densifyForHausdorff returns a copy of g with every linear segment
// subdivided so that the longest sub-segment is at most fraction times
// the original. Mirrors DiscreteHausdorffDistance.setDensifyFraction in
// JTS, which densifies before sampling vertices.
//
// Polygons and rings are normalised to a MultiLineString of their rings;
// points pass through.
func densifyForHausdorff(g geom.Geometry, fraction float64) geom.Geometry {
	if g == nil || g.IsEmpty() || fraction <= 0 || fraction >= 1 {
		return g
	}
	// Walk segments, emitting per-line densified copies.
	parts := []*geom.LineString{}
	visitConnectedComponents(g, func(c geom.Geometry) {
		switch v := c.(type) {
		case *geom.LineString:
			parts = append(parts, densifyLineString(v, fraction))
		case *geom.LinearRing:
			parts = append(parts, densifyLineString(v.AsLineString(), fraction))
		case *geom.Polygon:
			for r := 0; r < v.NumRings(); r++ {
				ring := v.Ring(r)
				ls := geom.NewLineString(v.CRS(), ring)
				parts = append(parts, densifyLineString(ls, fraction))
			}
		}
	})
	if len(parts) == 0 {
		return g
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return geom.NewMultiLineString(g.CRS(), parts...)
}

func densifyLineString(ls *geom.LineString, fraction float64) *geom.LineString {
	n := ls.NumPoints()
	if n < 2 {
		return ls
	}
	out := make([]geom.XY, 0, n*2)
	for i := 0; i+1 < n; i++ {
		a, b := ls.PointAt(i), ls.PointAt(i+1)
		out = append(out, a)
		segLen := math.Hypot(b.X-a.X, b.Y-a.Y)
		// Number of additional points so that each piece is <= fraction * segLen.
		// Solving (1 / (k+1)) <= fraction → k >= 1/fraction - 1.
		numSubs := int(math.Ceil(1.0/fraction)) - 1
		_ = segLen
		for k := 1; k <= numSubs; k++ {
			t := float64(k) / float64(numSubs+1)
			out = append(out, geom.XY{X: a.X + t*(b.X-a.X), Y: a.Y + t*(b.Y-a.Y)})
		}
	}
	out = append(out, ls.PointAt(n-1))
	return geom.NewLineString(ls.CRS(), out)
}

// visitGeometryVertices visits every vertex of g (polygon ring vertices
// included).
func visitGeometryVertices(g geom.Geometry, fn func(geom.XY)) {
	visitConnectedComponents(g, func(c geom.Geometry) {
		switch v := c.(type) {
		case *geom.Point:
			fn(v.XY())
		case *geom.LineString:
			n := v.NumPoints()
			for i := 0; i < n; i++ {
				fn(v.PointAt(i))
			}
		case *geom.LinearRing:
			ls := v.AsLineString()
			n := ls.NumPoints()
			for i := 0; i < n; i++ {
				fn(ls.PointAt(i))
			}
		case *geom.Polygon:
			for r := 0; r < v.NumRings(); r++ {
				ring := v.Ring(r)
				for _, p := range ring {
					fn(p)
				}
			}
		}
	})
}

// visitConnectedComponents walks the connected (non-collection) parts of
// g, calling fn for each.
func visitConnectedComponents(g geom.Geometry, fn func(geom.Geometry)) {
	switch v := g.(type) {
	case nil:
		return
	case *geom.Point:
		fn(v)
	case *geom.LineString:
		fn(v)
	case *geom.LinearRing:
		fn(v)
	case *geom.Polygon:
		fn(v)
	case *geom.MultiPoint:
		for i := 0; i < v.NumGeometries(); i++ {
			fn(geom.NewPoint(v.CRS(), v.PointAt(i)))
		}
	case *geom.MultiLineString:
		for i := 0; i < v.NumGeometries(); i++ {
			fn(v.LineStringAt(i))
		}
	case *geom.MultiPolygon:
		for i := 0; i < v.NumGeometries(); i++ {
			fn(v.PolygonAt(i))
		}
	case *geom.GeometryCollection:
		for i := 0; i < v.NumGeometries(); i++ {
			visitConnectedComponents(v.GeometryAt(i), fn)
		}
	}
}

// geometryBoundaryLines returns the linear boundary of g as a (possibly
// multi-) LineString. For polygonal inputs, this is the union of all
// rings. For linear inputs, returns a copy. For points, returns nil.
//
// Mirrors the role of LinearComponentExtracter / Geometry.getBoundary
// for the validator's purposes — we only need to compute distances
// to/from "the curve."
func geometryBoundaryLines(g geom.Geometry) geom.Geometry {
	if g == nil || g.IsEmpty() {
		return nil
	}
	parts := []*geom.LineString{}
	visitConnectedComponents(g, func(c geom.Geometry) {
		switch v := c.(type) {
		case *geom.LineString:
			parts = append(parts, v)
		case *geom.LinearRing:
			parts = append(parts, v.AsLineString())
		case *geom.Polygon:
			for r := 0; r < v.NumRings(); r++ {
				ring := v.Ring(r)
				if len(ring) >= 2 {
					parts = append(parts, geom.NewLineString(v.CRS(), ring))
				}
			}
		}
	})
	if len(parts) == 0 {
		return nil
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return geom.NewMultiLineString(g.CRS(), parts...)
}

// isArealOrCollection reports whether g is a Polygon, MultiPolygon, or a
// GeometryCollection (which may contain polygonal members).
func isArealOrCollection(g geom.Geometry) bool {
	switch g.(type) {
	case *geom.Polygon, *geom.MultiPolygon, *geom.GeometryCollection:
		return true
	}
	return false
}
