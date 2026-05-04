package noding

import (
	"math"

	"github.com/exergy-dev/go-topology-suite/geom"
)

// ScaledNoder wraps another noder, applying an integer-grid scale to
// every input coordinate before noding and the inverse scale after.
// Mirrors org.locationtech.jts.noding.ScaledNoder.
//
// The transform is:
//
//	scaled.X   = round((input.X - offset.X) * scale)
//	scaled.Y   = round((input.Y - offset.Y) * scale)
//	output.X   = scaled.X / scale + offset.X
//	output.Y   = scaled.Y / scale + offset.Y
//
// Scaling shifts arbitrary floating-point coordinates onto an integer
// grid so the inner noder (typically MCIndexNoder fed into snap-rounding)
// can take advantage of integer-precise predicates and a deterministic
// hot-pixel grid.
//
// A scale of zero or one is treated as a pass-through: input is fed
// through inner unchanged.
type ScaledNoder struct {
	inner    Noder
	scale    float64
	offset   geom.XY
	isScaled bool
}

// NewScaledNoder wraps inner with the given uniform scale and zero
// offset. Use NewScaledNoderWithOffset for a non-zero translation.
func NewScaledNoder(inner Noder, scale float64) *ScaledNoder {
	return NewScaledNoderWithOffset(inner, scale, geom.XY{})
}

// NewScaledNoderWithOffset wraps inner with the given scale and a
// translation offset applied before scaling.
func NewScaledNoderWithOffset(inner Noder, scale float64, offset geom.XY) *ScaledNoder {
	return &ScaledNoder{
		inner:    inner,
		scale:    scale,
		offset:   offset,
		isScaled: scale != 0 && scale != 1,
	}
}

// Scale returns the configured scale factor.
func (n *ScaledNoder) Scale() float64 { return n.scale }

// IsIntegerPrecision reports whether scaling is in effect (i.e. scale ≠
// 0 and ≠ 1). When false Node is a transparent pass-through to the
// inner noder.
func (n *ScaledNoder) IsIntegerPrecision() bool { return n.isScaled }

// Node satisfies the Noder interface.
func (n *ScaledNoder) Node(input []*SegmentString) []*SegmentString {
	if !n.isScaled {
		return n.inner.Node(input)
	}
	scaled := scaleInputs(input, n.scale, n.offset)
	out := n.inner.Node(scaled)
	return rescaleOutputs(out, n.scale, n.offset)
}

func scaleInputs(input []*SegmentString, scale float64, off geom.XY) []*SegmentString {
	out := make([]*SegmentString, len(input))
	for i, ss := range input {
		coords := make([]geom.XY, 0, len(ss.Coords))
		var prev geom.XY
		havePrev := false
		for _, p := range ss.Coords {
			q := geom.XY{
				X: math.Round((p.X - off.X) * scale),
				Y: math.Round((p.Y - off.Y) * scale),
			}
			if havePrev && q == prev {
				// Drop immediate duplicates introduced by rounding,
				// matching JTS CoordinateList semantics.
				continue
			}
			coords = append(coords, q)
			prev = q
			havePrev = true
		}
		out[i] = &SegmentString{Coords: coords, Tag: ss.Tag}
	}
	return out
}

func rescaleOutputs(strings []*SegmentString, scale float64, off geom.XY) []*SegmentString {
	out := make([]*SegmentString, len(strings))
	for i, ss := range strings {
		coords := make([]geom.XY, len(ss.Coords))
		for j, p := range ss.Coords {
			coords[j] = geom.XY{
				X: p.X/scale + off.X,
				Y: p.Y/scale + off.Y,
			}
		}
		out[i] = &SegmentString{Coords: coords, Tag: ss.Tag}
	}
	return out
}
