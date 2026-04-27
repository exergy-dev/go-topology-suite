package topology

import (
	"fmt"
	"strings"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/operation/relate"
)

type IntersectionMatrix = relate.IntersectionMatrix

type RelateOptions struct {
	AllowInvalidInputs bool
}

func Relate(a, b geom.Geometry, opts ...RelateOptions) (*relate.IntersectionMatrix, error) {
	if err := validatePredicateInputs("relate", a, b, relateOptions(opts...)); err != nil {
		return nil, err
	}
	return relate.Relate(a, b), nil
}

func RelatePattern(a, b geom.Geometry, pattern string, opts ...RelateOptions) (bool, error) {
	if err := validateRelatePattern(pattern); err != nil {
		return false, err
	}
	matrix, err := Relate(a, b, opts...)
	if err != nil {
		return false, err
	}
	return matrix.Matches(pattern), nil
}

func Intersects(a, b geom.Geometry, opts ...RelateOptions) (bool, error) {
	if err := validatePredicateInputs("intersects", a, b, relateOptions(opts...)); err != nil {
		return false, err
	}
	return geom.Intersects(a, b), nil
}

func Contains(a, b geom.Geometry, opts ...RelateOptions) (bool, error) {
	if err := validatePredicateInputs("contains", a, b, relateOptions(opts...)); err != nil {
		return false, err
	}
	return geom.Contains(a, b), nil
}

func Within(a, b geom.Geometry, opts ...RelateOptions) (bool, error) {
	if err := validatePredicateInputs("within", a, b, relateOptions(opts...)); err != nil {
		return false, err
	}
	return geom.Within(a, b), nil
}

func Touches(a, b geom.Geometry, opts ...RelateOptions) (bool, error) {
	if err := validatePredicateInputs("touches", a, b, relateOptions(opts...)); err != nil {
		return false, err
	}
	return geom.Touches(a, b), nil
}

func Crosses(a, b geom.Geometry, opts ...RelateOptions) (bool, error) {
	if err := validatePredicateInputs("crosses", a, b, relateOptions(opts...)); err != nil {
		return false, err
	}
	return geom.Crosses(a, b), nil
}

func Overlaps(a, b geom.Geometry, opts ...RelateOptions) (bool, error) {
	if err := validatePredicateInputs("overlaps", a, b, relateOptions(opts...)); err != nil {
		return false, err
	}
	return geom.Overlaps(a, b), nil
}

func Equals(a, b geom.Geometry, opts ...RelateOptions) (bool, error) {
	if err := validatePredicateInputs("equals", a, b, relateOptions(opts...)); err != nil {
		return false, err
	}
	return geom.Equals(a, b), nil
}

func Disjoint(a, b geom.Geometry, opts ...RelateOptions) (bool, error) {
	if err := validatePredicateInputs("disjoint", a, b, relateOptions(opts...)); err != nil {
		return false, err
	}
	return geom.Disjoint(a, b), nil
}

func relateOptions(opts ...RelateOptions) RelateOptions {
	if len(opts) == 0 {
		return RelateOptions{}
	}
	return opts[0]
}

func validatePredicateInputs(name string, a, b geom.Geometry, opts RelateOptions) error {
	if a == nil {
		return fmt.Errorf("v2 %s: left geometry is nil", name)
	}
	if b == nil {
		return fmt.Errorf("v2 %s: right geometry is nil", name)
	}
	if opts.AllowInvalidInputs {
		return nil
	}
	if err := Validate(a); err != nil {
		return fmt.Errorf("v2 %s: left input: %w", name, err)
	}
	if err := Validate(b); err != nil {
		return fmt.Errorf("v2 %s: right input: %w", name, err)
	}
	return nil
}

func validateRelatePattern(pattern string) error {
	if len(pattern) != 9 {
		return fmt.Errorf("v2 relate pattern: expected 9 characters, got %d", len(pattern))
	}
	for _, c := range strings.ToUpper(pattern) {
		switch c {
		case 'T', 'F', '*', '0', '1', '2':
		default:
			return fmt.Errorf("v2 relate pattern: invalid character %q", c)
		}
	}
	return nil
}
