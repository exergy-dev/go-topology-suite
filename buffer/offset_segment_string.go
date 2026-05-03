package buffer

import "github.com/terra-geo/terra/geom"

// offsetSegmentString is a dynamic list of vertices in a constructed
// offset curve. Adjacent vertices closer than minVertexDistance are
// silently dropped — this is the JTS equivalent of OffsetSegmentString
// (org.locationtech.jts.operation.buffer.OffsetSegmentString) and is the
// engine that keeps the offset curve from accumulating ULP-scale
// duplicate vertices at every corner.
//
// A typical configuration: minVertexDistance =
// |bufferDistance| * curveVertexSnapDistanceFactor (= 1e-4).
type offsetSegmentString struct {
	pts               []geom.XY
	minVertexDistance float64
}

// newOffsetSegmentString returns an empty accumulator with the given
// minimum vertex distance. Pass 0 to disable adjacent-vertex dedup
// (only exact equality will be filtered).
func newOffsetSegmentString(minVertexDistance float64) *offsetSegmentString {
	return &offsetSegmentString{minVertexDistance: minVertexDistance}
}

// addPt appends p to the accumulator unless it would coincide with the
// previous accumulated vertex within minVertexDistance.
func (s *offsetSegmentString) addPt(p geom.XY) {
	if s.isRedundant(p) {
		return
	}
	s.pts = append(s.pts, p)
}

// isRedundant reports whether p lies within minVertexDistance of the
// most recently accumulated vertex. Always false on an empty accumulator.
func (s *offsetSegmentString) isRedundant(p geom.XY) bool {
	if len(s.pts) == 0 {
		return false
	}
	last := s.pts[len(s.pts)-1]
	dx := p.X - last.X
	dy := p.Y - last.Y
	if dx == 0 && dy == 0 {
		return true
	}
	if s.minVertexDistance <= 0 {
		return false
	}
	d2 := dx*dx + dy*dy
	return d2 < s.minVertexDistance*s.minVertexDistance
}

// closeRing appends the first vertex if the accumulator's last vertex
// does not already match it. Empty accumulators are left alone.
func (s *offsetSegmentString) closeRing() {
	if len(s.pts) < 1 {
		return
	}
	start := s.pts[0]
	last := s.pts[len(s.pts)-1]
	if start == last {
		return
	}
	s.pts = append(s.pts, start)
}

// coordinates returns the accumulated vertex sequence. Callers must not
// mutate the returned slice.
func (s *offsetSegmentString) coordinates() []geom.XY {
	return s.pts
}
