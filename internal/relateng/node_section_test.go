package relateng

import (
	"testing"

	"github.com/terra-geo/terra/geom"
)

func TestNodeSection_Basic(t *testing.T) {
	v0 := geom.XY{X: 0, Y: 0}
	v1 := geom.XY{X: 2, Y: 0}
	pt := geom.XY{X: 1, Y: 0}
	s := NewNodeSection(true, DimL, 0, 0, nil, false, &v0, pt, &v1)
	if !s.IsA {
		t.Error("IsA expected true")
	}
	if !s.IsShell() {
		t.Error("ringId 0 should be shell")
	}
	if s.IsArea() {
		t.Error("dim=L should not be area")
	}
	if !s.IsProper() {
		t.Error("isNodeAtVertex=false should be proper")
	}
	if got := s.Vertex(0); got != &v0 {
		t.Errorf("Vertex(0) = %v, want &v0", got)
	}
	if got := s.Vertex(1); got != &v1 {
		t.Errorf("Vertex(1) = %v, want &v1", got)
	}
}

func TestNodeSection_SamePolygon(t *testing.T) {
	pt := geom.XY{X: 0, Y: 0}
	a := NewNodeSection(true, DimA, 1, 0, nil, true, nil, pt, nil)
	b := NewNodeSection(true, DimA, 1, 1, nil, true, nil, pt, nil)
	c := NewNodeSection(true, DimA, 2, 0, nil, true, nil, pt, nil)
	d := NewNodeSection(false, DimA, 1, 0, nil, true, nil, pt, nil)
	if !a.IsSamePolygon(b) {
		t.Error("same A.id=1 should be same polygon")
	}
	if a.IsSamePolygon(c) {
		t.Error("different ids → not same polygon")
	}
	if a.IsSamePolygon(d) {
		t.Error("different geometries → not same polygon")
	}
	if !a.IsSameGeometry(b) || !a.IsSameGeometry(c) {
		t.Error("a/b/c all in A → same geometry")
	}
	if a.IsSameGeometry(d) {
		t.Error("d in B → not same geometry")
	}
	if !IsAreaArea(a, b) {
		t.Error("two area sections → IsAreaArea")
	}
	if IsProperPair(a, b) {
		t.Error("both isNodeAtVertex=true → IsProperPair should be false")
	}
}

func TestNodeSection_LineEndpoint_NilVertex(t *testing.T) {
	pt := geom.XY{X: 5, Y: 5}
	v0 := geom.XY{X: 4, Y: 4}
	// Line ending at pt: V1 = nil
	s := NewNodeSection(true, DimL, 0, 0, nil, true, &v0, pt, nil)
	if s.Vertex(1) != nil {
		t.Error("V1 should be nil for line endpoint")
	}
}
