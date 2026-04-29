package noding

import (
	"sort"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel/planar"
)

// Noder takes input segment strings and returns a noded equivalent set.
// The output strings collectively cover the same point set as the input
// but no two output strings have an interior intersection — they only
// touch at vertex endpoints.
//
// Implementations must not mutate input strings; the returned slice and
// strings are freshly allocated.
type Noder interface {
	Node(input []*SegmentString) []*SegmentString
}

// SimpleNoder is a brute-force pairwise noder: O(n^2) edge comparisons.
// Correct on all inputs that the underlying segment-intersection
// primitive handles. See package doc for the collinear-overlap
// limitation.
//
// The zero value is ready to use.
type SimpleNoder struct{}

// Node returns a noded copy of input. Output strings share Tag with the
// input string they were derived from. Adjacent edges within a single
// SegmentString are not compared against each other (they share a
// vertex by construction); non-adjacent self-intersections within the
// same string ARE noded (e.g. an L-/figure-eight-shaped string is
// split into multiple non-self-crossing strings).
func (SimpleNoder) Node(input []*SegmentString) []*SegmentString {
	if len(input) == 0 {
		return nil
	}

	// For each edge (string i, segment j) collect parameter values t
	// in [0,1] of computed intersection points along that edge,
	// together with the corresponding intersection point. We carry
	// the point itself rather than recomputing it from t, so the
	// output uses the exact value the kernel returned.
	type split struct {
		t float64
		p geom.XY
	}
	splits := make([][][]split, len(input))
	for i, ss := range input {
		splits[i] = make([][]split, ss.NumSegments())
	}

	add := func(i, j int, t float64, p geom.XY) {
		// Skip endpoint hits — they're already vertices on this edge
		// and would only introduce duplicates. Use a tight tolerance
		// to handle the floating-point round-off that
		// SegmentIntersection can leave at exact endpoints.
		const eps = 1e-12
		if t <= eps || t >= 1-eps {
			return
		}
		splits[i][j] = append(splits[i][j], split{t: t, p: p})
	}

	// Pairwise edge intersection. We iterate every unordered pair of
	// edges across all strings, including pairs within the same
	// string (skipping adjacent edges, whose only intersection is
	// the shared vertex).
	for i1, ss1 := range input {
		n1 := ss1.NumSegments()
		for j1 := 0; j1 < n1; j1++ {
			a1, a2 := ss1.Segment(j1)

			for i2 := i1; i2 < len(input); i2++ {
				ss2 := input[i2]
				n2 := ss2.NumSegments()
				j2Start := 0
				if i2 == i1 {
					j2Start = j1 + 1
				}
				for j2 := j2Start; j2 < n2; j2++ {
					if i1 == i2 && (j2 == j1+1 || j1 == j2+1) {
						// Adjacent edges in the same string share
						// a vertex by construction.
						continue
					}
					b1, b2 := ss2.Segment(j2)
					p, ok := planar.Default.SegmentIntersection(a1, a2, b1, b2)
					if !ok {
						continue
					}
					t1 := segmentParam(a1, a2, p)
					t2 := segmentParam(b1, b2, p)
					add(i1, j1, t1, p)
					add(i2, j2, t2, p)
				}
			}
		}
	}

	// Build output strings: walk each input string and, between every
	// consecutive vertex pair, insert any intersection points along
	// that edge sorted by parameter.
	out := make([]*SegmentString, 0, len(input))
	for i, ss := range input {
		n := ss.NumSegments()
		if n == 0 {
			// Degenerate input — pass through unchanged.
			out = append(out, &SegmentString{
				Coords: append([]geom.XY(nil), ss.Coords...),
				Tag:    ss.Tag,
			})
			continue
		}

		// Build the full sequence of nodes along the original string,
		// then split into substrings at each interior intersection
		// point. A point is an interior intersection iff it was
		// inserted by add() above (i.e. is in splits[i][j] for some
		// j); original vertices are NOT split points unless they
		// happen to equal an intersection — which we don't track,
		// matching JTS behaviour where a "node" is any point a
		// downstream consumer might wish to break on.
		//
		// In practice: a self-crossing string should be broken at
		// the crossing point, while a string crossed by another
		// string should also be broken at that crossing. Both fall
		// out of the same rule: split at every recorded intersection
		// point.
		nodes := make([]geom.XY, 0, len(ss.Coords))
		// breaks[k] == true means split the output string at nodes[k]
		// (i.e. nodes[k] terminates the current piece and starts the
		// next). The first and last nodes always terminate pieces.
		breaks := make([]bool, 0, len(ss.Coords))

		for j := 0; j < n; j++ {
			a, b := ss.Segment(j)
			nodes = append(nodes, a)
			breaks = append(breaks, false)
			ints := splits[i][j]
			if len(ints) > 0 {
				sort.Slice(ints, func(p, q int) bool { return ints[p].t < ints[q].t })
				for k, s := range ints {
					// Skip duplicates produced when the same edge
					// is hit by multiple other edges at near-equal
					// parameter values.
					if k > 0 && s.t-ints[k-1].t < 1e-12 {
						continue
					}
					nodes = append(nodes, s.p)
					breaks = append(breaks, true)
				}
			}
			_ = b
		}
		// Append final vertex.
		nodes = append(nodes, ss.Coords[len(ss.Coords)-1])
		breaks = append(breaks, false)

		// Slice nodes at every break: each break point is the end of
		// the current piece AND the start of the next.
		start := 0
		for k := 1; k < len(nodes); k++ {
			if breaks[k] || k == len(nodes)-1 {
				piece := make([]geom.XY, k-start+1)
				copy(piece, nodes[start:k+1])
				out = append(out, &SegmentString{Coords: piece, Tag: ss.Tag})
				start = k
			}
		}
	}

	return out
}

// segmentParam returns the parameter t in [0,1] such that p ≈ a + t*(b-a).
// The caller has already established that p lies on segment [a,b]; this
// just picks the more numerically stable axis.
func segmentParam(a, b, p geom.XY) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	if abs(dx) >= abs(dy) {
		if dx == 0 {
			return 0
		}
		return (p.X - a.X) / dx
	}
	if dy == 0 {
		return 0
	}
	return (p.Y - a.Y) / dy
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
