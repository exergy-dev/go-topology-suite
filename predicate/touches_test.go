package predicate

import (
	"testing"

	"github.com/terra-geo/terra/wkt"
)

func TestPointPointNeverTouches(t *testing.T) {
	a, _ := wkt.Unmarshal("POINT (1 2)")
	b, _ := wkt.Unmarshal("POINT (1 2)")
	if got, _ := Touches(a, b); got {
		t.Errorf("Point/Point should never Touch (no boundary)")
	}
}

func TestPointTouchesLineEndpoint(t *testing.T) {
	endpoint, _ := wkt.Unmarshal("POINT (0 0)")
	mid, _ := wkt.Unmarshal("POINT (5 5)")
	off, _ := wkt.Unmarshal("POINT (1 0)")
	ls, _ := wkt.Unmarshal("LINESTRING (0 0, 5 5, 10 10)")

	if got, _ := Touches(endpoint, ls); !got {
		t.Errorf("point at line endpoint should Touch")
	}
	if got, _ := Touches(mid, ls); got {
		t.Errorf("point in line interior should not Touch (boundary-of-line is endpoints)")
	}
	if got, _ := Touches(off, ls); got {
		t.Errorf("point off line should not Touch")
	}
}

func TestPointTouchesPolygonBoundary(t *testing.T) {
	poly, _ := wkt.Unmarshal("POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))")
	boundary, _ := wkt.Unmarshal("POINT (0 5)")
	interior, _ := wkt.Unmarshal("POINT (5 5)")
	exterior, _ := wkt.Unmarshal("POINT (15 5)")

	if got, _ := Touches(boundary, poly); !got {
		t.Errorf("boundary point should Touch polygon")
	}
	if got, _ := Touches(interior, poly); got {
		t.Errorf("interior point should not Touch")
	}
	if got, _ := Touches(exterior, poly); got {
		t.Errorf("exterior point should not Touch")
	}
}

func TestPolygonsTouchAtEdge(t *testing.T) {
	a, _ := wkt.Unmarshal("POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))")
	// Touches at the right edge x=10.
	right, _ := wkt.Unmarshal("POLYGON ((10 0, 10 10, 20 10, 20 0, 10 0))")
	// Overlaps interior.
	overlap, _ := wkt.Unmarshal("POLYGON ((5 5, 5 15, 15 15, 15 5, 5 5))")
	// Disjoint.
	far, _ := wkt.Unmarshal("POLYGON ((20 0, 20 10, 30 10, 30 0, 20 0))")

	if got, _ := Touches(a, right); !got {
		t.Errorf("edge-sharing polygons should Touch")
	}
	if got, _ := Touches(a, overlap); got {
		t.Errorf("interior-overlapping polygons should not Touch")
	}
	if got, _ := Touches(a, far); got {
		t.Errorf("disjoint polygons should not Touch")
	}
}

func TestLineTouchesPolygonAtBoundary(t *testing.T) {
	poly, _ := wkt.Unmarshal("POLYGON ((0 0, 0 10, 10 10, 10 0, 0 0))")
	// Line that touches the polygon boundary at (0, 5) and goes outward.
	tangent, _ := wkt.Unmarshal("LINESTRING (-5 5, 0 5)")
	// Line that enters the polygon interior.
	entering, _ := wkt.Unmarshal("LINESTRING (-5 5, 5 5)")

	if got, _ := Touches(tangent, poly); !got {
		t.Errorf("tangent line should Touch polygon")
	}
	if got, _ := Touches(entering, poly); got {
		t.Errorf("entering line should not Touch (interior crossing)")
	}
}
