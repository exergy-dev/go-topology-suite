package measure

import (
	"math"
	"testing"
)

func TestCombineSimilarities_GeometricMean(t *testing.T) {
	got := CombineSimilarities(0.5, 0.5)
	if math.Abs(got-0.5) > 1e-12 {
		t.Fatalf("geo-mean of 0.5,0.5 = 0.5, got %v", got)
	}
	got = CombineSimilarities(1.0, 1.0, 1.0)
	if math.Abs(got-1.0) > 1e-12 {
		t.Fatalf("geo-mean of all 1s = 1, got %v", got)
	}
	// √(0.25 * 0.81) = 0.45
	got = CombineSimilarities(0.25, 0.81)
	if math.Abs(got-0.45) > 1e-12 {
		t.Fatalf("geo-mean(0.25,0.81)=0.45, got %v", got)
	}
}

func TestCombineSimilarities_ZeroPropagates(t *testing.T) {
	if got := CombineSimilarities(0.0, 0.5, 0.9); got != 0 {
		t.Fatalf("zero input should force 0, got %v", got)
	}
}

func TestCombineSimilarities_NaNPropagates(t *testing.T) {
	if got := CombineSimilarities(math.NaN(), 0.5); !math.IsNaN(got) {
		t.Fatalf("NaN input should propagate, got %v", got)
	}
}

func TestCombineSimilarities_Empty(t *testing.T) {
	if got := CombineSimilarities(); !math.IsNaN(got) {
		t.Fatalf("empty input: want NaN, got %v", got)
	}
}

func TestCombineMin(t *testing.T) {
	if got := CombineMin(0.7, 0.3, 0.9); math.Abs(got-0.3) > 1e-12 {
		t.Fatalf("min: want 0.3, got %v", got)
	}
	if got := CombineMin(); !math.IsNaN(got) {
		t.Fatalf("empty: want NaN, got %v", got)
	}
}
