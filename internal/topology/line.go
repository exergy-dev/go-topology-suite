// Package topology contains shared planar topology primitives used by
// operations that need consistent noding and labeling.
package topology

import (
	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/noding"
)

// LineSource identifies which input geometry contributed a noded segment.
type LineSource uint8

const (
	LineSourceA LineSource = 1 << iota
	LineSourceB
)

// InA reports whether the source set includes the left input.
func (s LineSource) InA() bool {
	return s&LineSourceA != 0
}

// InB reports whether the source set includes the right input.
func (s LineSource) InB() bool {
	return s&LineSourceB != 0
}

// NodedLineSegment is a line segment split at every intersection point and
// labeled with the input line sets that contain it.
type NodedLineSegment struct {
	Start, End geom.Coordinate
	Sources    LineSource
}

// InA reports whether the segment is covered by the left input.
func (s NodedLineSegment) InA() bool {
	return s.Sources.InA()
}

// InB reports whether the segment is covered by the right input.
func (s NodedLineSegment) InB() bool {
	return s.Sources.InB()
}

// LineString converts the segment into a geometry LineString.
func (s NodedLineSegment) LineString() *geom.LineString {
	return geom.NewLineString(geom.CoordinateSequence{s.Start, s.End})
}

// NodeLineSets nodes two line sets together and returns unique labeled
// segments. The returned segment direction is canonical, so opposite-direction
// duplicates are dissolved into a single labeled segment.
func NodeLineSets(linesA, linesB []*geom.LineString) []NodedLineSegment {
	segStrings := make([]*noding.NodedSegmentString, 0, len(linesA)+len(linesB))
	segStrings = appendLineSegmentStrings(segStrings, linesA, LineSourceA)
	segStrings = appendLineSegmentStrings(segStrings, linesB, LineSourceB)
	if len(segStrings) == 0 {
		return nil
	}

	noder := noding.NewSimpleNoder(noding.NewIntersectionAdder())
	noder.ComputeNodes(segStrings)

	segmentsByKey := make(map[lineSegmentKey]NodedLineSegment)
	for _, ss := range noder.GetNodedSubstrings() {
		source, _ := ss.Context().(LineSource)
		coords := ss.Coordinates()
		for i := 0; i < len(coords)-1; i++ {
			start := coords[i]
			end := coords[i+1]
			if start.Distance(end) <= geom.DefaultEpsilon {
				continue
			}

			key, canonicalStart, canonicalEnd := makeLineSegmentKey(start, end)
			segment := segmentsByKey[key]
			segment.Start = canonicalStart
			segment.End = canonicalEnd
			segment.Sources |= source
			segmentsByKey[key] = segment
		}
	}

	segments := make([]NodedLineSegment, 0, len(segmentsByKey))
	for _, segment := range segmentsByKey {
		segments = append(segments, segment)
	}
	return segments
}

// NodeLineSetsWithPrecision applies a precision model to cloned line
// coordinates before noding. Input geometries are not mutated.
func NodeLineSetsWithPrecision(linesA, linesB []*geom.LineString, pm geom.PrecisionModel) []NodedLineSegment {
	if pm == nil {
		return NodeLineSets(linesA, linesB)
	}
	return NodeLineSets(makePreciseLines(linesA, pm), makePreciseLines(linesB, pm))
}

func appendLineSegmentStrings(dst []*noding.NodedSegmentString, lines []*geom.LineString, source LineSource) []*noding.NodedSegmentString {
	for _, line := range lines {
		if line == nil || line.IsEmpty() {
			continue
		}
		coords := line.Coordinates()
		if len(coords) < 2 {
			continue
		}
		dst = append(dst, noding.NewNodedSegmentString(coords, source))
	}
	return dst
}

func makePreciseLines(lines []*geom.LineString, pm geom.PrecisionModel) []*geom.LineString {
	precise := make([]*geom.LineString, 0, len(lines))
	for _, line := range lines {
		if line == nil {
			continue
		}
		coords := line.Coordinates().Clone()
		geom.MakePreciseSequence(pm, coords)
		precise = append(precise, geom.NewLineString(coords))
	}
	return precise
}

type lineSegmentKey struct {
	x1, y1, x2, y2 float64
}

func makeLineSegmentKey(a, b geom.Coordinate) (lineSegmentKey, geom.Coordinate, geom.Coordinate) {
	if b.X < a.X || (b.X == a.X && b.Y < a.Y) {
		a, b = b, a
	}
	return lineSegmentKey{a.X, a.Y, b.X, b.Y}, a, b
}
