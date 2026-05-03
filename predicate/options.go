package predicate

import (
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/kernel"
	"github.com/terra-geo/terra/kernel/planar"
	"github.com/terra-geo/terra/kernel/spherical"
)

// Option configures a predicate call.
//
// An Option is a value type carrying a kernel choice and/or a prepared
// handle. Callers construct Options via WithKernel / WithPrepared.
//
// The historic shape (a closure) caused config to escape to the heap on
// every predicate call that used options; the value-type representation
// keeps the per-call config struct on the stack.
type Option struct {
	kernel       kernel.Kernel
	kernelSet    bool
	prepared     preparedHandle
	bnr          BoundaryNodeRule
	bnrSet       bool
	useRelateNG  bool
	useRelateSet bool
}

type config struct {
	kernel       kernel.Kernel
	kernelSet    bool
	prepared     preparedHandle // optional cached prepared geometry for `a`
	bnr          BoundaryNodeRule
	bnrSet       bool
	useRelateNG  bool
	useRelateSet bool
}

// preparedHandle is a thin interface so this package doesn't directly
// depend on terra/prepare (which would create a cycle once prepare grows
// to use predicates). Concrete instance: *prepare.PreparedPolygon.
//
// Methods beyond ContainsPoint/IntersectsEnvelope are reached via type
// assertion in the per-predicate fast paths; only the core two are
// required so a prepared form for a non-polygonal geometry can implement
// the minimal interface.
type preparedHandle interface {
	ContainsPoint(p geom.XY) kernel.Containment
	IntersectsEnvelope(e geom.Envelope) bool
}

// preparedIntersector is implemented by prepared geometries that can
// answer the generic Intersects(geometry) question without a fall-through
// to the slow path.
type preparedIntersector interface {
	Intersects(g geom.Geometry) bool
}

// preparedCoverer is implemented by prepared geometries that can answer
// the generic Covers(geometry) question.
type preparedCoverer interface {
	Covers(g geom.Geometry) bool
}

// WithKernel selects an explicit geometric kernel. When omitted the
// kernel is chosen based on the operands' CRS: geographic → spherical,
// projected (or no CRS) → planar.
func WithKernel(k kernel.Kernel) Option {
	return Option{kernel: k, kernelSet: true}
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
	return Option{prepared: h}
}

// UseRelateNG opts the call into the experimental RelateNG topology
// driver (internal/relateng). When set to true, Relate first attempts
// to evaluate the predicate via the new path; if RelateNG cannot
// produce a definitive answer (e.g. inputs whose answer depends on
// non-vertex edge intersections, which are not yet ported), the call
// falls back to the legacy DE-9IM computation in this package.
//
// The default is false (legacy path). The switch is intended for
// equivalence testing during the RelateNG rollout — once the missing
// edge-segment intersection pipeline lands, the default may flip.
func UseRelateNG(use bool) Option {
	return Option{useRelateNG: use, useRelateSet: true}
}

// WithBoundaryNodeRule selects a non-default rule for classifying
// MultiLineString endpoint nodes as boundary or interior. The OGC
// SFS default is Mod2BoundaryNodeRule; pass
// EndpointBoundaryNodeRule (or one of the others) when modelling
// linear-network topology where every endpoint is meaningful.
//
// Currently honoured by relate-driven predicates (Relate / Touches /
// Crosses / Intersects against MultiLineStrings). Polygonal rules
// are unaffected.
func WithBoundaryNodeRule(rule BoundaryNodeRule) Option {
	return Option{bnr: rule, bnrSet: true}
}

// resolve chooses a kernel given the operands. If the user explicitly
// passed WithKernel, that wins; otherwise we route by CRS kind.
//
// Fast path: when no options are passed (the by-far common case in hot
// loops) the config stays on the stack.
func resolve(g geom.Geometry, opts []Option) config {
	if len(opts) == 0 {
		return config{kernel: defaultKernelFor(g)}
	}
	c := config{}
	for _, o := range opts {
		if o.kernelSet {
			c.kernel = o.kernel
			c.kernelSet = true
		}
		if o.prepared != nil {
			c.prepared = o.prepared
		}
		if o.bnrSet {
			c.bnr = o.bnr
			c.bnrSet = true
		}
		if o.useRelateSet {
			c.useRelateNG = o.useRelateNG
			c.useRelateSet = true
		}
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
