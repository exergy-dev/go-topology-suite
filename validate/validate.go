package validate

import (
	"fmt"
	"strings"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel/planar"
)

// DefectKind classifies a structural defect.
type DefectKind string

const (
	DefectRingNotClosed     DefectKind = "ring-not-closed"
	DefectRingTooFewPoints  DefectKind = "ring-too-few-points"
	DefectLineTooFewPoints  DefectKind = "line-too-few-points"
	DefectSelfIntersection  DefectKind = "self-intersection"
	DefectHoleOutsideShell  DefectKind = "hole-outside-shell"
	DefectInvalidLayout     DefectKind = "invalid-layout"
)

// Defect describes one specific failure.
type Defect struct {
	Kind     DefectKind
	Message  string
	Location geom.XY // approximate location, zero if not applicable
}

// ValidationError aggregates all defects found in a single Validate call.
type ValidationError struct {
	Defects []Defect
}

func (e *ValidationError) Error() string {
	var b strings.Builder
	b.WriteString("terra: invalid geometry: ")
	for i, d := range e.Defects {
		if i > 0 {
			b.WriteString("; ")
		}
		fmt.Fprintf(&b, "%s: %s", d.Kind, d.Message)
	}
	return b.String()
}

// Validate returns nil if g is a valid OGC geometry, or *ValidationError
// listing every defect detected.
func Validate(g geom.Geometry) error {
	v := &validator{}
	v.check(g)
	if len(v.defects) == 0 {
		return nil
	}
	return &ValidationError{Defects: v.defects}
}

type validator struct {
	defects []Defect
}

func (v *validator) add(kind DefectKind, msg string, loc geom.XY) {
	v.defects = append(v.defects, Defect{Kind: kind, Message: msg, Location: loc})
}

func (v *validator) check(g geom.Geometry) {
	if g.Layout() == geom.NoLayout && !g.IsEmpty() {
		v.add(DefectInvalidLayout, "geometry has NoLayout but is not empty", geom.XY{})
	}
	switch x := g.(type) {
	case *geom.Point:
		// Empty or single coordinate; nothing further to validate.
	case *geom.LineString:
		v.checkLineString(x)
	case *geom.Polygon:
		v.checkPolygon(x)
	case *geom.MultiPoint:
		// Each member is a single coordinate; no structural rule beyond layout.
	case *geom.MultiLineString:
		for i := 0; i < x.NumGeometries(); i++ {
			v.checkLineString(x.LineStringAt(i))
		}
	case *geom.MultiPolygon:
		for i := 0; i < x.NumGeometries(); i++ {
			v.checkPolygon(x.PolygonAt(i))
		}
	case *geom.GeometryCollection:
		for i := 0; i < x.NumGeometries(); i++ {
			v.check(x.GeometryAt(i))
		}
	}
}

func (v *validator) checkLineString(ls *geom.LineString) {
	if ls.IsEmpty() {
		return
	}
	if ls.NumPoints() < 2 {
		v.add(DefectLineTooFewPoints,
			fmt.Sprintf("line has %d points, need ≥2", ls.NumPoints()),
			ls.PointAt(0))
	}
}

func (v *validator) checkPolygon(p *geom.Polygon) {
	if p.IsEmpty() {
		return
	}
	for r := 0; r < p.NumRings(); r++ {
		ring := p.Ring(r)
		if len(ring) < 4 {
			loc := geom.XY{}
			if len(ring) > 0 {
				loc = ring[0]
			}
			v.add(DefectRingTooFewPoints,
				fmt.Sprintf("ring %d has %d vertices, need ≥4", r, len(ring)),
				loc)
			continue
		}
		if ring[0] != ring[len(ring)-1] {
			v.add(DefectRingNotClosed,
				fmt.Sprintf("ring %d not closed: first=%v last=%v", r, ring[0], ring[len(ring)-1]),
				ring[0])
		}
		// Self-intersection within the ring.
		if loc, ok := ringSelfIntersection(ring); ok {
			v.add(DefectSelfIntersection,
				fmt.Sprintf("ring %d self-intersects", r), loc)
		}
	}
	// Hole containment: every interior ring must lie inside the outer ring.
	if p.NumRings() > 1 {
		outer := p.Ring(0)
		k := planar.Default
		for r := 1; r < p.NumRings(); r++ {
			ring := p.Ring(r)
			for _, vert := range ring {
				if k.PointInRing(vert, outer) == 0 { // Outside
					v.add(DefectHoleOutsideShell,
						fmt.Sprintf("hole %d vertex outside shell", r-1), vert)
					break
				}
			}
		}
	}
}

// ringSelfIntersection returns the first pair of non-adjacent ring edges
// that properly cross. Endpoint-touch is allowed.
func ringSelfIntersection(ring []geom.XY) (geom.XY, bool) {
	k := planar.Default
	n := len(ring)
	for i := 0; i+1 < n; i++ {
		a1, a2 := ring[i], ring[i+1]
		for j := i + 2; j+1 < n; j++ {
			// Skip the closing edge that shares endpoint with edge 0.
			if i == 0 && j+1 == n-1 {
				continue
			}
			b1, b2 := ring[j], ring[j+1]
			ip, ok := k.SegmentIntersection(a1, a2, b1, b2)
			if !ok {
				continue
			}
			// Touching at a shared vertex is not a self-intersection.
			if ip == a1 || ip == a2 || ip == b1 || ip == b2 {
				continue
			}
			return ip, true
		}
	}
	return geom.XY{}, false
}
