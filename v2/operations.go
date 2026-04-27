package topology

import (
	"fmt"
	"math"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/operation/buffer"
	"github.com/robert-malhotra/go-topology-suite/operation/overlay"
)

type OverlayOp = overlay.Op

const (
	IntersectionOp  = overlay.OpIntersection
	UnionOp         = overlay.OpUnion
	DifferenceOp    = overlay.OpDifference
	SymDifferenceOp = overlay.OpSymDifference
)

type OverlayOptions struct {
	AllowInvalidInputs bool
	NormalizeResult    bool
	PrecisionModel     geom.PrecisionModel
}

type BufferOptions struct {
	Params            *buffer.Params
	AllowInvalidInput bool
	NormalizeResult   bool
	PrecisionModel    geom.PrecisionModel
}

func Overlay(a, b geom.Geometry, op overlay.Op, opts ...OverlayOptions) (geom.Geometry, error) {
	cfg := overlayOptions(opts...)
	if a == nil {
		return nil, fmt.Errorf("v2 overlay: left geometry is nil")
	}
	if b == nil {
		return nil, fmt.Errorf("v2 overlay: right geometry is nil")
	}
	if err := validateOverlayInputs(a, b, cfg); err != nil {
		return nil, err
	}
	if op < overlay.OpIntersection || op > overlay.OpSymDifference {
		return nil, fmt.Errorf("v2 overlay: unsupported operation %d", op)
	}

	result := overlay.OverlayWithPrecision(a, b, op, cfg.PrecisionModel)
	if cfg.NormalizeResult && result != nil {
		result = result.Normalized()
	}
	return result, nil
}

func Intersection(a, b geom.Geometry, opts ...OverlayOptions) (geom.Geometry, error) {
	return Overlay(a, b, overlay.OpIntersection, opts...)
}

func Union(a, b geom.Geometry, opts ...OverlayOptions) (geom.Geometry, error) {
	return Overlay(a, b, overlay.OpUnion, opts...)
}

func Difference(a, b geom.Geometry, opts ...OverlayOptions) (geom.Geometry, error) {
	return Overlay(a, b, overlay.OpDifference, opts...)
}

func SymDifference(a, b geom.Geometry, opts ...OverlayOptions) (geom.Geometry, error) {
	return Overlay(a, b, overlay.OpSymDifference, opts...)
}

func Buffer(g geom.Geometry, distance float64, opts ...BufferOptions) (geom.Geometry, error) {
	cfg := bufferOptions(opts...)
	if g == nil {
		return nil, fmt.Errorf("v2 buffer: geometry is nil")
	}
	if cfg.PrecisionModel != nil {
		g = makePreciseGeometry(g, cfg.PrecisionModel)
	}
	if !cfg.AllowInvalidInput {
		if err := Validate(g); err != nil {
			return nil, err
		}
	}

	params := cfg.Params
	if params == nil {
		params = buffer.DefaultParams()
	}
	if err := validateBufferParams(params); err != nil {
		return nil, err
	}
	result := buffer.BufferWithParams(g, distance, params)
	if cfg.NormalizeResult && result != nil {
		result = result.Normalized()
	}
	return result, nil
}

func overlayOptions(opts ...OverlayOptions) OverlayOptions {
	if len(opts) == 0 {
		return OverlayOptions{}
	}
	return opts[0]
}

func bufferOptions(opts ...BufferOptions) BufferOptions {
	if len(opts) == 0 {
		return BufferOptions{}
	}
	return opts[0]
}

func validateOverlayInputs(a, b geom.Geometry, opts OverlayOptions) error {
	if opts.AllowInvalidInputs {
		return nil
	}
	if err := Validate(a); err != nil {
		return fmt.Errorf("v2 overlay: left input: %w", err)
	}
	if err := Validate(b); err != nil {
		return fmt.Errorf("v2 overlay: right input: %w", err)
	}
	return nil
}

func validateBufferParams(params *buffer.Params) error {
	if params.QuadrantSegments <= 0 {
		return fmt.Errorf("v2 buffer: quadrant segments must be > 0")
	}
	switch params.EndCapStyle {
	case buffer.CapRound, buffer.CapFlat, buffer.CapSquare:
	default:
		return fmt.Errorf("v2 buffer: unsupported end cap style %d", params.EndCapStyle)
	}
	switch params.JoinStyle {
	case buffer.JoinRound, buffer.JoinMitre, buffer.JoinBevel:
	default:
		return fmt.Errorf("v2 buffer: unsupported join style %d", params.JoinStyle)
	}
	if params.JoinStyle == buffer.JoinMitre && (math.IsNaN(params.MitreLimit) || math.IsInf(params.MitreLimit, 0) || params.MitreLimit <= 0) {
		return fmt.Errorf("v2 buffer: mitre limit must be finite and > 0 for mitre joins")
	}
	return nil
}

func makePreciseGeometry(g geom.Geometry, pm geom.PrecisionModel) geom.Geometry {
	if g == nil || pm == nil {
		return g
	}
	clone := g.Clone()
	if filterable, ok := clone.(geom.CoordinateFilterer); ok {
		filterable.ApplyCoordinateFilter(precisionFilter{pm: pm})
	}
	return clone
}

type precisionFilter struct {
	pm geom.PrecisionModel
}

func (f precisionFilter) Filter(coord *geom.Coordinate) {
	f.pm.MakePrecise(coord)
}
