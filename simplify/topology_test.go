package simplify

import (
	"testing"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/validate"
	"github.com/terra-geo/terra/wkt"
)

func TestTopologyPreservingStraightLine(t *testing.T) {
	g, _ := wkt.Unmarshal("LINESTRING (0 0, 1 1, 2 2)")
	out := TopologyPreserving(g, 0.5)
	ls := out.(*geom.LineString)
	if ls.NumPoints() != 2 {
		t.Errorf("collinear simplification produced %d points, want 2", ls.NumPoints())
	}
}

func TestTopologyPreservingKeepsBumps(t *testing.T) {
	g, _ := wkt.Unmarshal("LINESTRING (0 0, 1 1, 2 0)")
	// Tolerance 0.5 → threshold 0.25. Triangle area for the three points
	// is 1 (2× = 2), which is > 0.25 → bump kept.
	out := TopologyPreserving(g, 0.5)
	if ls := out.(*geom.LineString); ls.NumPoints() != 3 {
		t.Errorf("expected bump to be kept (3 points), got %d", ls.NumPoints())
	}
	out2 := TopologyPreserving(g, 2)
	if ls := out2.(*geom.LineString); ls.NumPoints() != 2 {
		t.Errorf("aggressive tolerance should drop the bump, got %d", ls.NumPoints())
	}
}

func TestTopologyPreservingPolygonStaysValid(t *testing.T) {
	// A figure with a notch — aggressive simplification could naively
	// flatten the notch and create self-intersection. The
	// topology-preserving variant must NOT introduce one.
	g, _ := wkt.Unmarshal(`POLYGON ((0 0, 10 0, 10 10, 6 10, 6 4, 4 4, 4 10, 0 10, 0 0))`)
	out := TopologyPreserving(g, 5).(*geom.Polygon)
	// Validate: must be simple. (validate package returns error on
	// self-intersecting polygons.)
	if err := validate.Validate(out); err != nil {
		t.Errorf("topology-preserving simplify produced invalid polygon: %v", err)
	}
}

func TestTopologyPreservingZeroToleranceIdentity(t *testing.T) {
	g, _ := wkt.Unmarshal("LINESTRING (0 0, 1 1, 2 2, 3 3)")
	out := TopologyPreserving(g, 0)
	if out != g {
		t.Errorf("zero tolerance should return identity geometry")
	}
}

func TestTopologyPreservingMultiLineString(t *testing.T) {
	g, _ := wkt.Unmarshal(`MULTILINESTRING ((0 0, 1 1, 2 2), (5 5, 6 6, 7 7))`)
	out := TopologyPreserving(g, 0.5).(*geom.MultiLineString)
	for i := 0; i < out.NumGeometries(); i++ {
		if out.LineStringAt(i).NumPoints() != 2 {
			t.Errorf("part %d: expected 2 points after collinear simplification, got %d",
				i, out.LineStringAt(i).NumPoints())
		}
	}
}
