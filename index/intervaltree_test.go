package index

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func collectIntervals[T any](t *IntervalRTree[T], min, max float64) []IntervalItem[T] {
	var out []IntervalItem[T]
	t.Query(min, max, func(it IntervalItem[T]) bool {
		out = append(out, it)
		return true
	})
	return out
}

func TestIntervalRTreeEmpty(t *testing.T) {
	tr := NewIntervalRTree[int]()
	assert.Empty(t, collectIntervals(tr, 0, 100))
}

func TestIntervalRTreeBasic(t *testing.T) {
	tr := NewIntervalRTree[int]()
	tr.Insert(1, 3, 1)
	tr.Insert(5, 7, 2)
	tr.Insert(10, 12, 3)
	tr.Insert(2, 6, 4)

	got := collectIntervals(tr, 4, 5)
	values := map[int]bool{}
	for _, it := range got {
		values[it.Value] = true
	}
	// items intersecting [4,5]: 2 ([5,7]), 4 ([2,6])
	assert.True(t, values[2], "missing item 2")
	assert.True(t, values[4], "missing item 4")
	assert.False(t, values[1], "item 1 should not be present")
	assert.False(t, values[3], "item 3 should not be present")
}

func TestIntervalRTreePoint(t *testing.T) {
	tr := NewIntervalRTree[int]()
	tr.Insert(1, 3, 1)
	tr.Insert(2, 4, 2)
	tr.Insert(5, 7, 3)

	// Point query at 2.5 should hit items 1 and 2.
	got := collectIntervals(tr, 2.5, 2.5)
	values := map[int]bool{}
	for _, it := range got {
		values[it.Value] = true
	}
	assert.True(t, values[1])
	assert.True(t, values[2])
	assert.False(t, values[3])
}

func TestIntervalRTreeInsertAfterQueryPanics(t *testing.T) {
	tr := NewIntervalRTree[int]()
	tr.Insert(0, 1, 1)
	collectIntervals(tr, 0, 1)
	assert.Panics(t, func() {
		tr.Insert(2, 3, 2)
	})
}

func TestIntervalRTreeRandom(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	type entry struct {
		min, max float64
		id       int
	}
	const n = 500
	entries := make([]entry, n)
	tr := NewIntervalRTree[int]()
	for i := 0; i < n; i++ {
		min := rng.Float64() * 1000
		w := rng.Float64() * 20
		entries[i] = entry{min: min, max: min + w, id: i}
		tr.Insert(min, min+w, i)
	}
	qmin, qmax := 200.0, 250.0
	expected := map[int]bool{}
	for _, e := range entries {
		if !(e.min > qmax || e.max < qmin) {
			expected[e.id] = true
		}
	}
	got := collectIntervals(tr, qmin, qmax)
	gotSet := map[int]bool{}
	for _, it := range got {
		gotSet[it.Value] = true
	}
	for id := range expected {
		assert.True(t, gotSet[id], "missing expected id %d", id)
	}
	for id := range gotSet {
		assert.True(t, expected[id], "false positive id %d", id)
	}
}

func TestIntervalRTreeEarlyExit(t *testing.T) {
	tr := NewIntervalRTree[int]()
	for i := 0; i < 50; i++ {
		x := float64(i)
		tr.Insert(x, x+1, i)
	}
	count := 0
	tr.Query(0, 100, func(IntervalItem[int]) bool {
		count++
		return count < 5
	})
	assert.Equal(t, 5, count)
}

func TestIntervalRTreeSingleItem(t *testing.T) {
	tr := NewIntervalRTree[int]()
	tr.Insert(5, 10, 42)
	got := collectIntervals(tr, 7, 8)
	assert.Len(t, got, 1)
	assert.Equal(t, 42, got[0].Value)

	got = collectIntervals(tr, 100, 200)
	assert.Empty(t, got)
}
