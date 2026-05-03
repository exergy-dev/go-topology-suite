package measure

import (
	"math"

	"github.com/terra-geo/terra/densify"
	"github.com/terra-geo/terra/geom"
)

// IntersectionFunc and UnionFunc are pluggable hooks used by
// AreaSimilarity to compute geometric intersection and union without
// pulling overlay into the measure package's dependency graph: overlay
// already depends on measure (for Area / Centroid helpers), so a direct
// import would form a cycle.
//
// Importing the github.com/terra-geo/terra/measure/match package wires
// these hooks via that package's init(); blank-importing it from a
// build is the simplest way to enable AreaSimilarity:
//
//	import _ "github.com/terra-geo/terra/measure/match"
//
// Callers that do not register the hooks see NaN from AreaSimilarity.
var (
	IntersectionFunc func(a, b geom.Geometry) (geom.Geometry, error)
	UnionFunc        func(a, b geom.Geometry) (geom.Geometry, error)
)

// hausdorffSimilarityDensifyFraction is the relative segment-length
// fraction used to densify both inputs before computing the discrete
// Hausdorff distance (mirrors the JTS DENSIFY_FRACTION = 0.25 constant).
const hausdorffSimilarityDensifyFraction = 0.25

// HausdorffSimilarity returns a similarity score in [0, 1] between two
// geometries based on the Hausdorff distance metric. A score of 1
// indicates identical geometries; lower scores indicate divergence.
//
// The score is `1 - HD(a, b) / diag(env(a) ∪ env(b))`, where HD is the
// (densified) discrete Hausdorff distance and diag is the diagonal
// length of the combined envelope of the two inputs. Both geometries
// are first densified to ensure the discrete Hausdorff distance closely
// approximates the continuous one — the densification segment length
// is `densifyFraction × diag`, matching JTS exactly.
//
// Empty inputs: if both empty, returns 1 (vacuously identical). If only
// one is empty, returns 0.
//
// Port of org.locationtech.jts.algorithm.match.HausdorffSimilarityMeasure.
func HausdorffSimilarity(a, b geom.Geometry) float64 {
	if a == nil || b == nil {
		return math.NaN()
	}
	aEmpty := a.IsEmpty()
	bEmpty := b.IsEmpty()
	if aEmpty && bEmpty {
		return 1
	}
	if aEmpty || bEmpty {
		return 0
	}

	envA := a.Envelope()
	envB := b.Envelope()
	env := envA.ExpandToInclude(envB)
	envSize := envelopeDiagonal(env)
	if envSize == 0 {
		// Both inputs collapsed to a point — distance also collapses
		// to 0, so similarity is 1 if they coincide.
		if DiscreteHausdorff(a, b) == 0 {
			return 1
		}
		return 0
	}

	// Densify both inputs to improve the accuracy of the discrete
	// Hausdorff approximation. Segment length = fraction × diagonal,
	// matching DiscreteHausdorffDistance.distance(g1, g2, fraction).
	maxLen := hausdorffSimilarityDensifyFraction * envSize
	dA := densify.Densify(a, maxLen)
	dB := densify.Densify(b, maxLen)

	dist := DiscreteHausdorff(dA, dB)
	if dist == 0 {
		return 1
	}
	if math.IsInf(dist, +1) {
		return 0
	}
	return 1 - dist/envSize
}

// AreaSimilarity returns a similarity score in [0, 1] between two
// geometries based on the area of their intersection over the area of
// their union: `area(a ∩ b) / area(a ∪ b)`. A score of 1 indicates
// areas perfectly coincide; 0 means the geometries are areally
// disjoint.
//
// Empty inputs: both empty returns 1; only one empty returns 0.
//
// Returns NaN if either input is nil, or if the overlay hook
// (IntersectionFunc / UnionFunc) has not been registered. Importing
// the github.com/terra-geo/terra/overlay package wires the hook
// automatically; otherwise the caller must set IntersectionFunc and
// UnionFunc explicitly.
//
// Port of org.locationtech.jts.algorithm.match.AreaSimilarityMeasure.
func AreaSimilarity(a, b geom.Geometry) float64 {
	if a == nil || b == nil {
		return math.NaN()
	}
	aEmpty := a.IsEmpty()
	bEmpty := b.IsEmpty()
	if aEmpty && bEmpty {
		return 1
	}
	if aEmpty || bEmpty {
		return 0
	}
	if IntersectionFunc == nil || UnionFunc == nil {
		return math.NaN()
	}
	inter, err := IntersectionFunc(a, b)
	if err != nil {
		return math.NaN()
	}
	un, err := UnionFunc(a, b)
	if err != nil {
		return math.NaN()
	}
	areaInter := Area(inter)
	areaUnion := Area(un)
	if areaUnion == 0 {
		// Both inputs are non-areal (lines/points) or coincide on a
		// measure-zero set. Treat as fully similar iff the union's
		// area also vanishes (i.e., both reduce to the same set).
		return 1
	}
	return areaInter / areaUnion
}

// FrechetSimilarity returns a similarity score in [0, 1] between two
// LineStrings based on the discrete Fréchet distance metric. A score
// of 1 indicates identical curves; 0 indicates fully divergent curves.
//
// The score is `1 - F(a, b) / diag(env(a) ∪ env(b))`, where F is the
// discrete Fréchet distance and diag is the diagonal length of the
// combined envelope. Unlike Hausdorff, Fréchet is order-sensitive, so
// the input vertex sequences must be ordered consistently.
//
// Empty inputs: both empty returns 1; only one empty returns 0.
//
// Port of org.locationtech.jts.algorithm.match.FrechetSimilarityMeasure.
func FrechetSimilarity(a, b *geom.LineString) float64 {
	if a == nil || b == nil {
		return math.NaN()
	}
	aEmpty := a.IsEmpty()
	bEmpty := b.IsEmpty()
	if aEmpty && bEmpty {
		return 1
	}
	if aEmpty || bEmpty {
		return 0
	}

	dist := DiscreteFrechet(a, b)
	if dist == 0 {
		return 1
	}
	if math.IsInf(dist, +1) {
		return 0
	}

	envA := a.Envelope()
	envB := b.Envelope()
	env := envA.ExpandToInclude(envB)
	envSize := envelopeDiagonal(env)
	if envSize == 0 {
		// Both inputs degenerate to a single point — distance is 0
		// already handled above; otherwise treat as fully divergent.
		return 0
	}
	return 1 - dist/envSize
}

// envelopeDiagonal returns the length of the envelope's diagonal, or 0
// if the envelope is empty. Mirrors JTS HausdorffSimilarityMeasure.diagonalSize.
func envelopeDiagonal(env geom.Envelope) float64 {
	if env.IsEmpty() {
		return 0
	}
	w := env.Width()
	h := env.Height()
	return math.Sqrt(w*w + h*h)
}
