package index

import (
	"math"
	"math/rand"
	"testing"

	"github.com/terra-geo/terra/geom"
)

func TestKdTree_InsertAndQueryPoint(t *testing.T) {
	tree := NewKdTree[int](0)
	pts := []geom.XY{
		{X: 1, Y: 1},
		{X: 2, Y: 5},
		{X: -3, Y: 7},
		{X: 0, Y: 0},
		{X: 4, Y: -1},
	}
	for i, p := range pts {
		n, isNew := tree.Insert(p, i)
		if !isNew {
			t.Errorf("Insert(%v): expected isNew=true on first insert", p)
		}
		if !n.Coordinate.EqualBitwise(p) {
			t.Errorf("inserted node coord = %v, want %v", n.Coordinate, p)
		}
	}
	if got := tree.Len(); got != len(pts) {
		t.Fatalf("Len = %d, want %d", got, len(pts))
	}
	for _, p := range pts {
		n := tree.QueryPoint(p)
		if n == nil {
			t.Errorf("QueryPoint(%v) = nil, want hit", p)
			continue
		}
		if !n.Coordinate.EqualBitwise(p) {
			t.Errorf("QueryPoint(%v).Coordinate = %v", p, n.Coordinate)
		}
	}
	if got := tree.QueryPoint(geom.XY{X: 99, Y: 99}); got != nil {
		t.Errorf("QueryPoint miss returned %v, want nil", got)
	}
}

func TestKdTree_ToleranceDedup(t *testing.T) {
	tree := NewKdTree[string](0.5)
	a, _ := tree.Insert(geom.XY{X: 1, Y: 1}, "a")
	b, isNewB := tree.Insert(geom.XY{X: 1.2, Y: 1.1}, "b")
	if isNewB {
		t.Fatalf("near-duplicate insert allocated new node")
	}
	if a != b {
		t.Fatalf("dedup did not return same node")
	}
	if a.Count != 2 {
		t.Errorf("a.Count = %d, want 2", a.Count)
	}
	c, isNewC := tree.Insert(geom.XY{X: 5, Y: 5}, "c")
	if !isNewC || c == a {
		t.Fatalf("far insert should create a new node")
	}
	if got := tree.Len(); got != 2 {
		t.Errorf("Len = %d, want 2", got)
	}
}

func TestKdTree_QueryEnvelope(t *testing.T) {
	tree := NewKdTree[int](0)
	pts := []geom.XY{
		{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 2},
		{X: 3, Y: 3}, {X: 4, Y: 4}, {X: -1, Y: -1},
		{X: 5, Y: 0}, {X: 0, Y: 5},
	}
	for i, p := range pts {
		tree.Insert(p, i)
	}
	env := geom.Envelope{MinX: 1, MinY: 1, MaxX: 3, MaxY: 3}
	got := tree.QueryAll(env)
	if len(got) != 3 {
		t.Errorf("QueryAll = %d nodes, want 3", len(got))
	}
	for _, n := range got {
		if !env.ContainsXY(n.Coordinate) {
			t.Errorf("QueryAll returned %v outside env", n.Coordinate)
		}
	}
}

func TestKdTree_NearestNeighbor(t *testing.T) {
	tree := NewKdTree[int](0)
	pts := []geom.XY{
		{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 0, Y: 10},
		{X: 10, Y: 10}, {X: 5, Y: 5},
	}
	for i, p := range pts {
		tree.Insert(p, i)
	}
	q := geom.XY{X: 6, Y: 6}
	n, ok := tree.NearestNeighbor(q)
	if !ok {
		t.Fatalf("NearestNeighbor(empty?) returned false")
	}
	want := geom.XY{X: 5, Y: 5}
	if !n.Coordinate.EqualBitwise(want) {
		t.Errorf("NearestNeighbor = %v, want %v", n.Coordinate, want)
	}
}

func TestKdTree_NearestNeighbor_Random(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	tree := NewKdTree[int](0)
	const N = 500
	pts := make([]geom.XY, N)
	for i := range pts {
		pts[i] = geom.XY{X: rng.Float64() * 100, Y: rng.Float64() * 100}
		tree.Insert(pts[i], i)
	}
	for q := 0; q < 50; q++ {
		query := geom.XY{X: rng.Float64() * 100, Y: rng.Float64() * 100}
		got, ok := tree.NearestNeighbor(query)
		if !ok {
			t.Fatal("nearest empty")
		}
		// Brute-force verify
		bestSq := math.Inf(+1)
		var want geom.XY
		for _, p := range pts {
			dx := query.X - p.X
			dy := query.Y - p.Y
			d := dx*dx + dy*dy
			if d < bestSq {
				bestSq = d
				want = p
			}
		}
		if !got.Coordinate.EqualBitwise(want) {
			t.Errorf("query=%v: got %v, want %v", query, got.Coordinate, want)
		}
	}
}

func TestKdTree_NearestNeighborEmpty(t *testing.T) {
	tree := NewKdTree[int](0)
	if _, ok := tree.NearestNeighbor(geom.XY{X: 0, Y: 0}); ok {
		t.Errorf("NearestNeighbor on empty tree returned ok=true")
	}
}
