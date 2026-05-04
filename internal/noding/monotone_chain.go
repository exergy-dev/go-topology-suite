package noding

import "github.com/terra-geo/terra/geom"

// MonotoneChain is a contiguous run of segments within a SegmentString
// whose direction lies in a single quadrant — i.e. dx and dy each have
// constant sign over the chain. Because of monotonicity:
//
//  1. No two segments inside the chain intersect each other (so segment
//     pairs within one chain need not be tested).
//  2. The bounding envelope of any sub-range [s,e] of the chain equals
//     the envelope of just its endpoints {pts[s], pts[e]}, which lets
//     overlap queries between two chains be resolved by mutual binary
//     subdivision rather than scanning every segment.
//
// This is a Go port of org.locationtech.jts.index.chain.MonotoneChain.
// Like JTS we hold a reference to the original coordinate slice and
// just record the [Start, End] index range, so chain construction
// allocates O(numChains) — not O(numCoords).
//
// The Tag carries the originating SegmentString.Tag through callbacks so
// MCIndexNoder can recover which input string a chain came from.
type MonotoneChain struct {
	Pts    []geom.XY // backing slice (NOT copied)
	Start  int       // index of first vertex of the chain
	End    int       // index of last vertex of the chain (Start+1 minimum)
	ID     int       // unique id assigned by the noder for pair-ordering
	Tag    int       // Tag of the originating SegmentString
	env    geom.Envelope
	hasEnv bool
}

// Envelope returns the chain's bounding envelope. Because the chain is
// monotone, this is just the envelope of its two endpoints.
func (mc *MonotoneChain) Envelope() geom.Envelope {
	if !mc.hasEnv {
		mc.env = geom.SegmentEnvelope(mc.Pts[mc.Start], mc.Pts[mc.End])
		mc.hasEnv = true
	}
	return mc.env
}

// EnvelopeExpanded returns the chain envelope expanded by overlapTolerance
// on every side. Useful when the noder is doing a tolerance-buffered
// overlap test (e.g. for SnappingNoder).
func (mc *MonotoneChain) EnvelopeExpanded(overlapTolerance float64) geom.Envelope {
	env := mc.Envelope()
	if overlapTolerance > 0 {
		return env.ExpandBy(overlapTolerance)
	}
	return env
}

// NumSegments is the number of segments in the chain.
func (mc *MonotoneChain) NumSegments() int { return mc.End - mc.Start }

// ComputeOverlaps reports every pair of segments (one from mc, one from
// other) whose envelopes might overlap, using a binary-subdivision
// recursion. visit receives the start index of each candidate segment
// in its respective chain. If overlapTolerance > 0 the envelope test is
// inflated by that distance.
//
// As in JTS this may report pairs whose segments do not actually
// intersect — the visit function must do the precise test itself.
func (mc *MonotoneChain) ComputeOverlaps(other *MonotoneChain, overlapTolerance float64, visit func(mc1 *MonotoneChain, start1 int, mc2 *MonotoneChain, start2 int)) {
	mc.computeOverlapsRange(mc.Start, mc.End, other, other.Start, other.End, overlapTolerance, visit)
}

func (mc *MonotoneChain) computeOverlapsRange(
	start0, end0 int,
	other *MonotoneChain,
	start1, end1 int,
	overlapTolerance float64,
	visit func(*MonotoneChain, int, *MonotoneChain, int),
) {
	// Terminal: both ranges are single segments — emit the pair.
	if end0-start0 == 1 && end1-start1 == 1 {
		visit(mc, start0, other, start1)
		return
	}
	// Bounding-envelope rejection on the sub-ranges.
	if !rangesOverlap(mc.Pts[start0], mc.Pts[end0], other.Pts[start1], other.Pts[end1], overlapTolerance) {
		return
	}
	mid0 := (start0 + end0) / 2
	mid1 := (start1 + end1) / 2
	if start0 < mid0 {
		if start1 < mid1 {
			mc.computeOverlapsRange(start0, mid0, other, start1, mid1, overlapTolerance, visit)
		}
		if mid1 < end1 {
			mc.computeOverlapsRange(start0, mid0, other, mid1, end1, overlapTolerance, visit)
		}
	}
	if mid0 < end0 {
		if start1 < mid1 {
			mc.computeOverlapsRange(mid0, end0, other, start1, mid1, overlapTolerance, visit)
		}
		if mid1 < end1 {
			mc.computeOverlapsRange(mid0, end0, other, mid1, end1, overlapTolerance, visit)
		}
	}
}

func rangesOverlap(p1, p2, q1, q2 geom.XY, tol float64) bool {
	// X axis
	minp, maxp := p1.X, p2.X
	if minp > maxp {
		minp, maxp = maxp, minp
	}
	minq, maxq := q1.X, q2.X
	if minq > maxq {
		minq, maxq = maxq, minq
	}
	if minp > maxq+tol || maxp < minq-tol {
		return false
	}
	// Y axis
	minp, maxp = p1.Y, p2.Y
	if minp > maxp {
		minp, maxp = maxp, minp
	}
	minq, maxq = q1.Y, q2.Y
	if minq > maxq {
		minq, maxq = maxq, minq
	}
	if minp > maxq+tol || maxp < minq-tol {
		return false
	}
	return true
}

// BuildMonotoneChains splits ss.Coords into a sequence of monotone
// chains. Zero-length segments are absorbed into the surrounding chain
// (matching JTS MonotoneChainBuilder.findChainEnd).
//
// The returned chains all reference ss.Coords directly — do not mutate
// the slice while chains are in use.
func BuildMonotoneChains(ss *SegmentString) []*MonotoneChain {
	pts := ss.Coords
	if len(pts) < 2 {
		return nil
	}
	var chains []*MonotoneChain
	chainStart := 0
	for {
		chainEnd := findChainEnd(pts, chainStart)
		chains = append(chains, &MonotoneChain{
			Pts:   pts,
			Start: chainStart,
			End:   chainEnd,
			Tag:   ss.Tag,
		})
		chainStart = chainEnd
		if chainStart >= len(pts)-1 {
			break
		}
	}
	return chains
}

// findChainEnd returns the index of the last vertex of the monotone
// chain starting at start. Mirrors JTS MonotoneChainBuilder.findChainEnd:
// zero-length segments are silently absorbed (they cannot define a
// quadrant), and the chain extends as long as every non-zero segment
// shares the chain's starting quadrant.
func findChainEnd(pts []geom.XY, start int) int {
	safeStart := start
	for safeStart < len(pts)-1 && pts[safeStart] == pts[safeStart+1] {
		safeStart++
	}
	if safeStart >= len(pts)-1 {
		return len(pts) - 1
	}
	chainQuad := quadrant(pts[safeStart], pts[safeStart+1])
	last := start + 1
	for last < len(pts) {
		if pts[last-1] != pts[last] {
			q := quadrant(pts[last-1], pts[last])
			if q != chainQuad {
				break
			}
		}
		last++
	}
	return last - 1
}

// quadrant returns the JTS quadrant index of the directed segment
// p0 -> p1. Quadrants are numbered:
//
//	1 | 0
//	-----
//	2 | 3
//
// matching org.locationtech.jts.geom.Quadrant.quadrant(Coordinate, Coordinate).
// p0 == p1 is forbidden by the caller (chain builder strips zero-length
// runs first), but we fall back to quadrant 0 to keep the function total.
func quadrant(p0, p1 geom.XY) int {
	dx := p1.X - p0.X
	dy := p1.Y - p0.Y
	switch {
	case dx >= 0 && dy >= 0:
		return 0
	case dx < 0 && dy >= 0:
		return 1
	case dx < 0 && dy < 0:
		return 2
	default:
		return 3
	}
}
