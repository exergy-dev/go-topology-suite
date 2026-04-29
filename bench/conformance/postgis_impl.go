//go:build postgis

package conformance

import (
	"errors"

	"github.com/terra-geo/terra/geom"
)

// TODO(b2-postgis): wire this stub up to a running Postgres + PostGIS
// instance. The intended shape:
//
//   import (
//       "database/sql"
//       _ "github.com/jackc/pgx/v5/stdlib"
//   )
//
//   type postgisImpl struct{ db *sql.DB }
//
//   func NewPostGIS(connStr string) (Impl, error) {
//       db, err := sql.Open("pgx", connStr)
//       if err != nil { return nil, err }
//       return &postgisImpl{db: db}, nil
//   }
//
// Each operation issues a `SELECT ST_<Op>(ST_GeomFromText($1), ...)`
// query, scans the result back as WKT (`ST_AsText`), and parses with
// wkt.Unmarshal. Use a dedicated test schema so concurrent runs don't
// collide.
//
// Connection-string example:
//
//   postgres://terra:terra@127.0.0.1:5432/terra_conformance?sslmode=disable
//
// CI integration: run a Postgres+PostGIS service container in the
// `postgis` workflow job and gate the build on the postgis tag, e.g.:
//
//   go test -tags postgis ./bench/conformance/...

// postgisImpl is the placeholder. Calls fail loudly so a missed wiring
// step is impossible to overlook.
type postgisImpl struct{}

// NewPostGIS returns a stubbed PostGIS Impl. Replace the body with a
// real connection-aware constructor when wiring this up.
func NewPostGIS() Impl { return postgisImpl{} }

func (postgisImpl) Name() string { return "postgis" }

func (postgisImpl) Intersection(a, b geom.Geometry) (geom.Geometry, error) {
	return nil, errPostGISNotImplemented
}

func (postgisImpl) Union(a, b geom.Geometry) (geom.Geometry, error) {
	return nil, errPostGISNotImplemented
}

func (postgisImpl) Difference(a, b geom.Geometry) (geom.Geometry, error) {
	return nil, errPostGISNotImplemented
}

func (postgisImpl) Area(g geom.Geometry) (float64, error) {
	return 0, errPostGISNotImplemented
}

func (postgisImpl) Length(g geom.Geometry) (float64, error) {
	return 0, errPostGISNotImplemented
}

func (postgisImpl) Relate(a, b geom.Geometry) (string, error) {
	return "", errPostGISNotImplemented
}

var errPostGISNotImplemented = errors.New(
	"conformance: postgis impl is a stub; see TODO(b2-postgis) in postgis_impl.go")
