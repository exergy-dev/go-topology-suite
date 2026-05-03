package measure

import (
	"math"

	"github.com/terra-geo/terra/densify"
	"github.com/terra-geo/terra/geom"
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
