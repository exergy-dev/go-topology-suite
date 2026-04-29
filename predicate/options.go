package predicate

import (
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel"
	"github.com/terra-geo/terra/kernel/planar"
	"github.com/terra-geo/terra/kernel/spherical"
)

// Option configures a predicate call.
type Option func(*config)

type config struct {
	kernel    kernel.Kernel
	kernelSet bool
	prepared  preparedHandle // optional cached prepared geometry for `a`
}

// preparedHandle is a thin interface so this package doesn't directly
// depend on terra/prepare (which would create a cycle once prepare grows
// to use predicates). Concrete instance: *prepare.PreparedPolygon.
type preparedHandle interface {
	ContainsPoint(p geom.XY) kernel.Containment
	IntersectsEnvelope(e geom.Envelope) bool
}

// WithKernel selects an explicit geometric kernel. When omitted the
// kernel is chosen based on the operands' CRS: geographic → spherical,
// projected (or no CRS) → planar.
func WithKernel(k kernel.Kernel) Option {
	return func(c *config) { c.kernel = k; c.kernelSet = true }
}

// WithPrepared attaches a pre-computed acceleration structure for `a`,
// the first operand of subsequent predicate calls. When the same polygon
// is tested against many points (or many other geometries), preparing
// once and passing it via this option amortises the cost across all
// queries.
//
// Example:
//
//	prepared := prepare.Polygon(myPolygon)
//	for _, pt := range points {
//	    in, _ := predicate.Contains(myPolygon, pt, predicate.WithPrepared(prepared))
//	    ...
//	}
//
// The handle interface is intentionally narrow so the predicate package
// doesn't import the prepare package (which itself uses index/predicate).
func WithPrepared(h preparedHandle) Option {
	return func(c *config) { c.prepared = h }
}

// resolve chooses a kernel given the operands. If the user explicitly
// passed WithKernel, that wins; otherwise we route by CRS kind.
func resolve(g geom.Geometry, opts []Option) config {
	c := config{}
	for _, opt := range opts {
		opt(&c)
	}
	if !c.kernelSet {
		c.kernel = defaultKernelFor(g)
	}
	return c
}

// defaultKernelFor returns spherical for geographic CRSes, planar otherwise.
// For predicates the cheap spherical model is preferred to the geodesic;
// the topological answer is the same for non-degenerate inputs.
func defaultKernelFor(g geom.Geometry) kernel.Kernel {
	if g != nil && g.CRS().IsGeographic() {
		return spherical.Default
	}
	return planar.Default
}
