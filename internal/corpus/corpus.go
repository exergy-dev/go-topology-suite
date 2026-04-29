package corpus

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/terra-geo/terra/geojson"
	"github.com/terra-geo/terra/geom"
)

//go:embed testdata/natural_earth_admin0_sample.geojson
//go:embed testdata/tiger_counties_sample.geojson
//go:embed testdata/osm_buildings_sample.geojson
var fixtureFS embed.FS

// fixtureSource describes one embedded corpus.
type fixtureSource struct {
	short string
	path  string
}

var sources = []fixtureSource{
	{"ne", "testdata/natural_earth_admin0_sample.geojson"},
	{"tiger", "testdata/tiger_counties_sample.geojson"},
	{"osm", "testdata/osm_buildings_sample.geojson"},
}

// Fixture is a named bundle of geometries decoded from a real-world style
// GeoJSON FeatureCollection.
type Fixture struct {
	// Name is the short identifier (e.g. "ne", "tiger", "osm").
	Name string
	// Features holds one geometry per Feature, preserving file order.
	Features []geom.Geometry
}

// Load returns the fixture registered under the given short name.
// Valid names: "ne", "tiger", "osm". Unknown names return an error.
func Load(name string) (*Fixture, error) {
	for _, s := range sources {
		if s.short == name {
			return decodeFixture(s)
		}
	}
	return nil, fmt.Errorf("corpus: unknown fixture %q", name)
}

// All returns every embedded fixture, in registration order.
// The slice is freshly decoded on each call; callers may mutate it freely.
func All() []*Fixture {
	out := make([]*Fixture, 0, len(sources))
	for _, s := range sources {
		f, err := decodeFixture(s)
		if err != nil {
			// Embedded data is compiled in: a decode failure is a programmer
			// error, not a runtime condition. Surface it loudly.
			panic(fmt.Errorf("corpus: decoding %s: %w", s.short, err))
		}
		out = append(out, f)
	}
	return out
}

func decodeFixture(s fixtureSource) (*Fixture, error) {
	data, err := fixtureFS.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("corpus: read %s: %w", s.path, err)
	}
	var fc geojson.FeatureCollection
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("corpus: parse %s: %w", s.short, err)
	}
	feats := make([]geom.Geometry, 0, len(fc.Features))
	for i, f := range fc.Features {
		if f == nil || f.Geometry == nil {
			return nil, fmt.Errorf("corpus: %s feature %d has nil geometry", s.short, i)
		}
		feats = append(feats, f.Geometry)
	}
	return &Fixture{Name: s.short, Features: feats}, nil
}
