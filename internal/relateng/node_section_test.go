package relateng

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/exergy-dev/go-topology-suite/geom"
)

func TestNodeSection_Basic(t *testing.T) {
	v0 := geom.XY{X: 0, Y: 0}
	v1 := geom.XY{X: 2, Y: 0}
	pt := geom.XY{X: 1, Y: 0}
	s := NewNodeSection(true, DimL, 0, 0, nil, false, &v0, pt, &v1)
	assert.True(t, s.IsA, "IsA expected true")
	assert.True(t, s.IsShell(), "ringId 0 should be shell")
	assert.False(t, s.IsArea(), "dim=L should not be area")
	assert.True(t, s.IsProper(), "isNodeAtVertex=false should be proper")
	assert.Equal(t, &v0, s.Vertex(0), "Vertex(0)")
	assert.Equal(t, &v1, s.Vertex(1), "Vertex(1)")
}

func TestNodeSection_SamePolygon(t *testing.T) {
	pt := geom.XY{X: 0, Y: 0}
	a := NewNodeSection(true, DimA, 1, 0, nil, true, nil, pt, nil)
	b := NewNodeSection(true, DimA, 1, 1, nil, true, nil, pt, nil)
	c := NewNodeSection(true, DimA, 2, 0, nil, true, nil, pt, nil)
	d := NewNodeSection(false, DimA, 1, 0, nil, true, nil, pt, nil)
	assert.True(t, a.IsSamePolygon(b), "same A.id=1 should be same polygon")
	assert.False(t, a.IsSamePolygon(c), "different ids → not same polygon")
	assert.False(t, a.IsSamePolygon(d), "different geometries → not same polygon")
	assert.True(t, a.IsSameGeometry(b), "a/b in A → same geometry")
	assert.True(t, a.IsSameGeometry(c), "a/c in A → same geometry")
	assert.False(t, a.IsSameGeometry(d), "d in B → not same geometry")
	assert.True(t, IsAreaArea(a, b), "two area sections → IsAreaArea")
	assert.False(t, IsProperPair(a, b), "both isNodeAtVertex=true → IsProperPair should be false")
}

func TestNodeSection_LineEndpoint_NilVertex(t *testing.T) {
	pt := geom.XY{X: 5, Y: 5}
	v0 := geom.XY{X: 4, Y: 4}
	// Line ending at pt: V1 = nil
	s := NewNodeSection(true, DimL, 0, 0, nil, true, &v0, pt, nil)
	assert.Nil(t, s.Vertex(1), "V1 should be nil for line endpoint")
}
