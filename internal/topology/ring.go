package topology

import (
	"math"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// DirectedSegment is an oriented edge used for ring tracing.
type DirectedSegment struct {
	Start, End geom.Coordinate
}

// TraceRingsFromDirectedSegments builds closed rings from directed edges using
// a rightmost-turn traversal at graph nodes.
func TraceRingsFromDirectedSegments(segments []DirectedSegment) []geom.CoordinateSequence {
	if len(segments) == 0 {
		return nil
	}

	edgeMap := make(map[geom.CoordinateXY][]int)
	for i, segment := range segments {
		edgeMap[segment.Start.XY()] = append(edgeMap[segment.Start.XY()], i)
	}

	used := make([]bool, len(segments))
	rings := make([]geom.CoordinateSequence, 0)
	for i := range segments {
		if used[i] {
			continue
		}
		ring := traceRing(i, segments, edgeMap, used)
		if len(ring) >= 4 {
			rings = append(rings, ring)
		}
	}
	return rings
}

func traceRing(startIdx int, segments []DirectedSegment, edgeMap map[geom.CoordinateXY][]int, used []bool) geom.CoordinateSequence {
	start := segments[startIdx]
	ring := geom.CoordinateSequence{start.Start}
	currentIdx := startIdx
	used[currentIdx] = true

	maxSteps := len(segments) + 1
	for steps := 0; steps < maxSteps; steps++ {
		current := segments[currentIdx]
		ring = append(ring, current.End)
		if current.End.Equals2D(start.Start, geom.DefaultEpsilon) {
			return ring
		}

		nextIdx := findRightmostNextSegment(currentIdx, segments, edgeMap, used)
		if nextIdx < 0 {
			return nil
		}
		used[nextIdx] = true
		currentIdx = nextIdx
	}
	return nil
}

func findRightmostNextSegment(currentIdx int, segments []DirectedSegment, edgeMap map[geom.CoordinateXY][]int, used []bool) int {
	current := segments[currentIdx]
	candidateIndexes := edgeMap[current.End.XY()]
	bestIdx := -1
	bestAngleDiff := math.MaxFloat64
	incomingAngle := math.Atan2(current.End.Y-current.Start.Y, current.End.X-current.Start.X)

	for _, candidateIdx := range candidateIndexes {
		if used[candidateIdx] {
			continue
		}
		candidate := segments[candidateIdx]
		outgoingAngle := math.Atan2(candidate.End.Y-candidate.Start.Y, candidate.End.X-candidate.Start.X)
		angleDiff := normalizeAngle(outgoingAngle - incomingAngle)
		if angleDiff < bestAngleDiff {
			bestAngleDiff = angleDiff
			bestIdx = candidateIdx
		}
	}
	return bestIdx
}

func normalizeAngle(angle float64) float64 {
	for angle <= -math.Pi {
		angle += 2 * math.Pi
	}
	for angle > math.Pi {
		angle -= 2 * math.Pi
	}
	return angle
}
