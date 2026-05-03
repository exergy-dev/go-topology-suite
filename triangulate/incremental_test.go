package triangulate

import (
	"math"
	"math/rand"
	"testing"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/triangulate/quadedge"
)

func TestDelaunayOf_Empty(t *testing.T) {
	tris, err := DelaunayOf(nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(tris) != 0 {
		t.Fatalf("want 0 triangles, got %d", len(tris))
	}
}

func TestDelaunayOf_Triangle(t *testing.T) {
	pts := []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}}
	tris, err := DelaunayOf(pts)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(tris) != 1 {
		t.Fatalf("want 1 triangle, got %d", len(tris))
	}
}

func TestDelaunayOf_Collinear(t *testing.T) {
	pts := []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}, {X: 3, Y: 0}}
	tris, err := DelaunayOf(pts)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	// Collinear input has no interior triangles.
	if len(tris) != 0 {
		t.Fatalf("want 0 triangles for collinear input, got %d", len(tris))
	}
}

func TestDelaunayOf_Square(t *testing.T) {
	pts := []geom.XY{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}}
	tris, err := DelaunayOf(pts)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(tris) != 2 {
		t.Fatalf("want 2 triangles, got %d", len(tris))
	}
	assertEmptyCircumcircle(t, pts, tris)
}

func TestDelaunayOf_Random10(t *testing.T) {
	pts := randomPoints(10, 1)
	tris, err := DelaunayOf(pts)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(tris) == 0 {
		t.Fatal("expected some triangles")
	}
	assertEmptyCircumcircle(t, pts, tris)
}

func TestDelaunayOf_Random100(t *testing.T) {
	pts := randomPoints(100, 2)
	tris, err := DelaunayOf(pts)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(tris) == 0 {
		t.Fatal("expected some triangles")
	}
	assertEmptyCircumcircle(t, pts, tris)
}

func TestDelaunayOf_Random1000(t *testing.T) {
	pts := randomPoints(1000, 3)
	tris, err := DelaunayOf(pts)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(tris) == 0 {
		t.Fatal("expected some triangles")
	}
	assertEmptyCircumcircle(t, pts, tris)
}

func TestSubdivisionIsDelaunay(t *testing.T) {
	pts := randomPoints(50, 4)
	env := envelopeOf(pts)
	subdiv := quadedge.NewSubdivision(env, 0.0)
	tri := NewIncrementalDelaunayTriangulator(subdiv)
	verts := make([]*quadedge.Vertex, len(pts))
	for i, p := range pts {
		verts[i] = quadedge.NewVertex(p)
	}
	if err := tri.InsertSites(verts); err != nil {
		t.Fatal(err)
	}
	if !subdiv.IsDelaunay() {
		t.Fatal("subdivision is not Delaunay")
	}
}

func randomPoints(n int, seed int64) []geom.XY {
	r := rand.New(rand.NewSource(seed))
	pts := make([]geom.XY, n)
	for i := range pts {
		pts[i] = geom.XY{X: r.Float64() * 100, Y: r.Float64() * 100}
	}
	return pts
}

// assertEmptyCircumcircle verifies that no input point lies strictly
// inside the circumcircle of any triangle.
func assertEmptyCircumcircle(t *testing.T, pts []geom.XY, tris []Triangle) {
	t.Helper()
	const eps = 1e-9
	for _, tri := range tris {
		// Ensure CCW orientation for the predicate.
		a, b, c := tri.P0, tri.P1, tri.P2
		if (b.X-a.X)*(c.Y-a.Y)-(b.Y-a.Y)*(c.X-a.X) < 0 {
			b, c = c, b
		}
		for _, p := range pts {
			if approxEq(p, a) || approxEq(p, b) || approxEq(p, c) {
				continue
			}
			if inCircleStrict(a, b, c, p, eps) {
				t.Fatalf("empty-circumcircle property violated:\n  tri = %v %v %v\n  pt  = %v", a, b, c, p)
			}
		}
	}
}

func approxEq(a, b geom.XY) bool {
	return math.Abs(a.X-b.X) < 1e-12 && math.Abs(a.Y-b.Y) < 1e-12
}

func inCircleStrict(a, b, c, p geom.XY, eps float64) bool {
	adx := a.X - p.X
	ady := a.Y - p.Y
	bdx := b.X - p.X
	bdy := b.Y - p.Y
	cdx := c.X - p.X
	cdy := c.Y - p.Y
	abdet := adx*bdy - bdx*ady
	bcdet := bdx*cdy - cdx*bdy
	cadet := cdx*ady - adx*cdy
	alift := adx*adx + ady*ady
	blift := bdx*bdx + bdy*bdy
	clift := cdx*cdx + cdy*cdy
	disc := alift*bcdet + blift*cadet + clift*abdet
	return disc > eps
}
