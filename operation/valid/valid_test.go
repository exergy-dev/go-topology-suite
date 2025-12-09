package valid

import (
	"math"
	"testing"

	"github.com/go-topology-suite/gts/geom"
)

func TestValidPoint(t *testing.T) {
	p := geom.NewPoint(1, 2)
	result := Validate(p)
	if !result.IsValid {
		t.Errorf("Valid point should be valid: %s", result.Error())
	}
}

func TestEmptyPoint(t *testing.T) {
	p := geom.NewPointEmpty()
	result := Validate(p)
	if !result.IsValid {
		t.Errorf("Empty point should be valid: %s", result.Error())
	}
}

func TestPointWithNaN(t *testing.T) {
	p := geom.NewPoint(math.NaN(), 2)
	result := Validate(p)
	if result.IsValid {
		t.Error("Point with NaN should be invalid")
	}
	if len(result.Errors) == 0 {
		t.Error("Expected at least one error")
	}
	if result.Errors[0].Type != ErrInvalidCoordinate {
		t.Errorf("Expected ErrInvalidCoordinate, got %v", result.Errors[0].Type)
	}
}

func TestPointWithInf(t *testing.T) {
	p := geom.NewPoint(1, math.Inf(1))
	result := Validate(p)
	if result.IsValid {
		t.Error("Point with Inf should be invalid")
	}
}

func TestValidLineString(t *testing.T) {
	ls := geom.NewLineStringXY(0, 0, 1, 1, 2, 0)
	result := Validate(ls)
	if !result.IsValid {
		t.Errorf("Valid linestring should be valid: %s", result.Error())
	}
}

func TestEmptyLineString(t *testing.T) {
	ls := geom.NewLineStringEmpty()
	result := Validate(ls)
	if !result.IsValid {
		t.Errorf("Empty linestring should be valid: %s", result.Error())
	}
}

func TestLineStringWithOnePoint(t *testing.T) {
	ls := geom.NewLineString(geom.CoordinateSequence{geom.NewCoordinate(0, 0)})
	result := Validate(ls)
	if result.IsValid {
		t.Error("LineString with one point should be invalid")
	}
	if len(result.Errors) == 0 || result.Errors[0].Type != ErrTooFewPoints {
		t.Error("Expected ErrTooFewPoints")
	}
}

func TestValidLinearRing(t *testing.T) {
	lr := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	result := Validate(lr)
	if !result.IsValid {
		t.Errorf("Valid ring should be valid: %s", result.Error())
	}
}

func TestEmptyLinearRing(t *testing.T) {
	lr := geom.NewLinearRingEmpty()
	result := Validate(lr)
	if !result.IsValid {
		t.Errorf("Empty ring should be valid: %s", result.Error())
	}
}

func TestLinearRingTooFewPoints(t *testing.T) {
	// Create a ring with only 3 points (including closure)
	lr := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(1, 1),
		geom.NewCoordinate(0, 0),
	})
	result := Validate(lr)
	if result.IsValid {
		t.Error("Ring with too few points should be invalid")
	}
}

func TestSelfIntersectingRing(t *testing.T) {
	// Figure-8 shape
	lr := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	result := Validate(lr)
	if result.IsValid {
		t.Error("Self-intersecting ring should be invalid")
	}
	if len(result.Errors) == 0 || result.Errors[0].Type != ErrSelfIntersection {
		t.Error("Expected ErrSelfIntersection")
	}
}

func TestValidPolygon(t *testing.T) {
	factory := geom.DefaultFactory
	shell := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	poly := factory.CreatePolygon(shell, nil)
	result := Validate(poly)
	if !result.IsValid {
		t.Errorf("Valid polygon should be valid: %s", result.Error())
	}
}

func TestEmptyPolygon(t *testing.T) {
	poly := geom.NewPolygonEmpty()
	result := Validate(poly)
	if !result.IsValid {
		t.Errorf("Empty polygon should be valid: %s", result.Error())
	}
}

func TestPolygonWithWrongOrientation(t *testing.T) {
	// Clockwise shell (should be counter-clockwise)
	shell := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(0, 0),
	})
	poly := geom.NewPolygon(shell, nil)
	result := Validate(poly)
	if result.IsValid {
		t.Error("Polygon with wrong orientation should be invalid")
	}
	if len(result.Errors) == 0 || result.Errors[0].Type != ErrInvalidOrientation {
		t.Error("Expected ErrInvalidOrientation")
	}
}

func TestPolygonWithHole(t *testing.T) {
	factory := geom.DefaultFactory
	// Counter-clockwise shell
	shell := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(20, 0),
		geom.NewCoordinate(20, 20),
		geom.NewCoordinate(0, 20),
		geom.NewCoordinate(0, 0),
	})
	// Clockwise hole
	hole := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(5, 5),
		geom.NewCoordinate(5, 15),
		geom.NewCoordinate(15, 15),
		geom.NewCoordinate(15, 5),
		geom.NewCoordinate(5, 5),
	})
	poly := factory.CreatePolygon(shell, []*geom.LinearRing{hole})
	result := Validate(poly)
	if !result.IsValid {
		t.Errorf("Valid polygon with hole should be valid: %s", result.Error())
	}
}

func TestPolygonWithHoleWrongOrientation(t *testing.T) {
	factory := geom.DefaultFactory
	// Counter-clockwise shell
	shell := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(20, 0),
		geom.NewCoordinate(20, 20),
		geom.NewCoordinate(0, 20),
		geom.NewCoordinate(0, 0),
	})
	// Counter-clockwise hole (should be clockwise)
	hole := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(5, 5),
		geom.NewCoordinate(15, 5),
		geom.NewCoordinate(15, 15),
		geom.NewCoordinate(5, 15),
		geom.NewCoordinate(5, 5),
	})
	poly := factory.CreatePolygon(shell, []*geom.LinearRing{hole})
	result := Validate(poly)
	if result.IsValid {
		t.Error("Polygon with CCW hole should be invalid")
	}
}

func TestPolygonWithHoleOutside(t *testing.T) {
	factory := geom.DefaultFactory
	shell := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	// Hole outside shell
	hole := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(20, 20),
		geom.NewCoordinate(20, 30),
		geom.NewCoordinate(30, 30),
		geom.NewCoordinate(30, 20),
		geom.NewCoordinate(20, 20),
	})
	poly := factory.CreatePolygon(shell, []*geom.LinearRing{hole})
	result := Validate(poly)
	if result.IsValid {
		t.Error("Polygon with hole outside should be invalid")
	}
	hasHoleOutsideError := false
	for _, err := range result.Errors {
		if err.Type == ErrHoleOutsideShell {
			hasHoleOutsideError = true
			break
		}
	}
	if !hasHoleOutsideError {
		t.Error("Expected ErrHoleOutsideShell")
	}
}

func TestValidMultiPoint(t *testing.T) {
	mp := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(1, 2),
		geom.NewPoint(3, 4),
	})
	result := Validate(mp)
	if !result.IsValid {
		t.Errorf("Valid MultiPoint should be valid: %s", result.Error())
	}
}

func TestMultiPointWithInvalidPoint(t *testing.T) {
	mp := geom.NewMultiPoint([]*geom.Point{
		geom.NewPoint(1, 2),
		geom.NewPoint(math.NaN(), 4),
	})
	result := Validate(mp)
	if result.IsValid {
		t.Error("MultiPoint with invalid point should be invalid")
	}
}

func TestValidMultiLineString(t *testing.T) {
	mls := geom.NewMultiLineString([]*geom.LineString{
		geom.NewLineStringXY(0, 0, 1, 1),
		geom.NewLineStringXY(2, 2, 3, 3),
	})
	result := Validate(mls)
	if !result.IsValid {
		t.Errorf("Valid MultiLineString should be valid: %s", result.Error())
	}
}

func TestValidMultiPolygon(t *testing.T) {
	factory := geom.DefaultFactory
	shell1 := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	shell2 := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(20, 0),
		geom.NewCoordinate(30, 0),
		geom.NewCoordinate(30, 10),
		geom.NewCoordinate(20, 10),
		geom.NewCoordinate(20, 0),
	})
	mp := geom.NewMultiPolygon([]*geom.Polygon{
		factory.CreatePolygon(shell1, nil),
		factory.CreatePolygon(shell2, nil),
	})
	result := Validate(mp)
	if !result.IsValid {
		t.Errorf("Valid MultiPolygon should be valid: %s", result.Error())
	}
}

func TestMultiPolygonWithNestedShells(t *testing.T) {
	factory := geom.DefaultFactory
	// Outer polygon
	shell1 := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(20, 0),
		geom.NewCoordinate(20, 20),
		geom.NewCoordinate(0, 20),
		geom.NewCoordinate(0, 0),
	})
	// Inner polygon (inside the outer)
	shell2 := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(5, 5),
		geom.NewCoordinate(15, 5),
		geom.NewCoordinate(15, 15),
		geom.NewCoordinate(5, 15),
		geom.NewCoordinate(5, 5),
	})
	mp := geom.NewMultiPolygon([]*geom.Polygon{
		factory.CreatePolygon(shell1, nil),
		factory.CreatePolygon(shell2, nil),
	})
	result := Validate(mp)
	if result.IsValid {
		t.Error("MultiPolygon with nested shells should be invalid")
	}
	hasNestedError := false
	for _, err := range result.Errors {
		if err.Type == ErrNestedShells {
			hasNestedError = true
			break
		}
	}
	if !hasNestedError {
		t.Error("Expected ErrNestedShells")
	}
}

func TestValidGeometryCollection(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(1, 2),
		geom.NewLineStringXY(0, 0, 1, 1),
	})
	result := Validate(gc)
	if !result.IsValid {
		t.Errorf("Valid GeometryCollection should be valid: %s", result.Error())
	}
}

func TestGeometryCollectionWithInvalidGeometry(t *testing.T) {
	gc := geom.NewGeometryCollection([]geom.Geometry{
		geom.NewPoint(1, 2),
		geom.NewPoint(math.NaN(), 0),
	})
	result := Validate(gc)
	if result.IsValid {
		t.Error("GeometryCollection with invalid geometry should be invalid")
	}
}

func TestIsValid(t *testing.T) {
	p := geom.NewPoint(1, 2)
	if !IsValid(p) {
		t.Error("IsValid should return true for valid geometry")
	}

	invalid := geom.NewPoint(math.NaN(), 0)
	if IsValid(invalid) {
		t.Error("IsValid should return false for invalid geometry")
	}
}

func TestValidateNil(t *testing.T) {
	result := Validate(nil)
	if !result.IsValid {
		t.Error("nil geometry should be considered valid")
	}
}

func TestMakeValidPolygonOrientation(t *testing.T) {
	// Clockwise shell (should be counter-clockwise)
	shell := geom.NewLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(0, 0),
	})
	poly := geom.NewPolygon(shell, nil)

	if IsValid(poly) {
		t.Error("Original polygon should be invalid")
	}

	repaired, wasRepaired := MakeValid(poly)
	if !wasRepaired {
		t.Error("Polygon should have been repaired")
	}

	if !IsValid(repaired) {
		t.Error("Repaired polygon should be valid")
	}
}

func TestMakeValidAlreadyValid(t *testing.T) {
	factory := geom.DefaultFactory
	shell := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	poly := factory.CreatePolygon(shell, nil)

	repaired, wasRepaired := MakeValid(poly)
	if wasRepaired {
		t.Error("Already valid polygon should not be repaired")
	}
	if repaired != poly {
		t.Error("Should return same polygon if already valid")
	}
}

func TestValidationErrorString(t *testing.T) {
	err := &ValidationError{
		Type:    ErrSelfIntersection,
		Message: "test error",
	}
	s := err.Error()
	if s == "" {
		t.Error("Error string should not be empty")
	}

	errWithLocation := &ValidationError{
		Type:     ErrSelfIntersection,
		Location: &geom.Coordinate{X: 1, Y: 2},
		Message:  "at location",
	}
	s = errWithLocation.Error()
	if s == "" {
		t.Error("Error string with location should not be empty")
	}
}

func TestValidationErrorTypeString(t *testing.T) {
	tests := []struct {
		errType ValidationErrorType
		want    string
	}{
		{ErrNone, "no error"},
		{ErrInvalidCoordinate, "invalid coordinate (NaN or Inf)"},
		{ErrTooFewPoints, "too few points"},
		{ErrRingNotClosed, "ring not closed"},
		{ErrSelfIntersection, "self-intersection"},
		{ErrHoleOutsideShell, "hole outside shell"},
		{ErrNestedHoles, "nested holes"},
		{ErrNestedShells, "nested shells in MultiPolygon"},
		{ErrInvalidOrientation, "invalid ring orientation"},
	}

	for _, tt := range tests {
		got := tt.errType.String()
		if got != tt.want {
			t.Errorf("ValidationErrorType(%d).String() = %q, want %q", tt.errType, got, tt.want)
		}
	}
}

func BenchmarkValidateSimplePolygon(b *testing.B) {
	factory := geom.DefaultFactory
	shell := factory.CreateLinearRing(geom.CoordinateSequence{
		geom.NewCoordinate(0, 0),
		geom.NewCoordinate(10, 0),
		geom.NewCoordinate(10, 10),
		geom.NewCoordinate(0, 10),
		geom.NewCoordinate(0, 0),
	})
	poly := factory.CreatePolygon(shell, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(poly)
	}
}

func BenchmarkValidateComplexPolygon(b *testing.B) {
	factory := geom.DefaultFactory

	// Create a polygon with many points
	n := 100
	coords := make(geom.CoordinateSequence, n+1)
	for i := 0; i < n; i++ {
		angle := 2 * math.Pi * float64(i) / float64(n)
		coords[i] = geom.NewCoordinate(math.Cos(angle)*10, math.Sin(angle)*10)
	}
	coords[n] = coords[0].Clone()

	shell := factory.CreateLinearRing(coords)
	poly := factory.CreatePolygon(shell, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Validate(poly)
	}
}
