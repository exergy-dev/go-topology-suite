//go:build cgo

package conformance

import (
	"errors"

	"github.com/terra-geo/terra/geom"
)

// TODO(b2-geos): wire this stub up to github.com/twpayne/go-geos. The
// intended shape:
//
//   import goeos "github.com/twpayne/go-geos"
//
//   type geosImpl struct{ ctx *goeos.Context }
//
//   func NewGEOS() Impl {
//       return &geosImpl{ctx: goeos.NewContext()}
//   }
//
// Each operation marshals the Terra input to WKT (or WKB), constructs
// a geos.Geom via ctx.NewGeomFromWKT, runs the corresponding method
// (Intersection / Union / Difference / Area / Length / Relate), and
// converts back to Terra via the resulting ToWKT.
//
// Build incantation (note cgo and the libgeos-dev system package):
//
//   apt-get install libgeos-dev
//   CGO_ENABLED=1 go test -tags cgo ./bench/conformance/...
//
// Because go-geos uses cgo it MUST stay behind the `cgo` build tag —
// the default `go test ./...` invocation is required to remain
// cgo-free for portability and CI matrix simplicity.

// geosImpl is the placeholder. Calls fail loudly so a missed wiring
// step is impossible to overlook.
type geosImpl struct{}

// NewGEOS returns a stubbed go-geos Impl. Replace the body with the
// real go-geos-backed implementation when wiring this up.
func NewGEOS() Impl { return geosImpl{} }

func (geosImpl) Name() string { return "geos" }

func (geosImpl) Intersection(a, b geom.Geometry) (geom.Geometry, error) {
	return nil, errGEOSNotImplemented
}

func (geosImpl) Union(a, b geom.Geometry) (geom.Geometry, error) {
	return nil, errGEOSNotImplemented
}

func (geosImpl) Difference(a, b geom.Geometry) (geom.Geometry, error) {
	return nil, errGEOSNotImplemented
}

func (geosImpl) Area(g geom.Geometry) (float64, error) {
	return 0, errGEOSNotImplemented
}

func (geosImpl) Length(g geom.Geometry) (float64, error) {
	return 0, errGEOSNotImplemented
}

func (geosImpl) Relate(a, b geom.Geometry) (string, error) {
	return "", errGEOSNotImplemented
}

var errGEOSNotImplemented = errors.New(
	"conformance: geos impl is a stub; see TODO(b2-geos) in geos_impl.go")
