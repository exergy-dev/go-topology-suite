package geom_test

import (
	"math"
	"testing"

	"github.com/exergy-dev/go-topology-suite/crs"
	"github.com/exergy-dev/go-topology-suite/geom"
)

func TestEqualsExact(t *testing.T) {
	nan := math.NaN()
	sq := []geom.XY{{0, 0}, {4, 0}, {4, 4}, {0, 4}, {0, 0}}
	hole := []geom.XY{{1, 1}, {2, 1}, {2, 2}, {1, 1}}
	// The same 9 vertices as shell+hole, but split 4/5 instead of 5/4.
	sqHole := append(append([]geom.XY{}, sq...), hole...)
	ls := func(pts ...geom.XY) *geom.LineString { return geom.NewLineString(nil, pts) }

	tests := []struct {
		name string
		a, b geom.Geometry
		want bool
	}{
		// nil
		{"nil vs nil", nil, nil, true},
		{"nil vs point", nil, geom.NewPoint(nil, geom.XY{X: 1, Y: 2}), false},
		{"point vs nil", geom.NewPoint(nil, geom.XY{X: 1, Y: 2}), nil, false},
		{"typed nil vs nil", (*geom.Point)(nil), nil, true},
		{"typed nil vs typed nil", (*geom.Polygon)(nil), (*geom.LineString)(nil), true},
		{"typed nil vs point", (*geom.Point)(nil), geom.NewPoint(nil, geom.XY{}), false},

		// Point, all layouts
		{"point XY equal", geom.NewPoint(nil, geom.XY{X: 1, Y: 2}), geom.NewPoint(nil, geom.XY{X: 1, Y: 2}), true},
		{"point XY differ", geom.NewPoint(nil, geom.XY{X: 1, Y: 2}), geom.NewPoint(nil, geom.XY{X: 1, Y: 3}), false},
		{"point XYZ equal", geom.NewPointXYZ(nil, geom.XYZ{X: 1, Y: 2, Z: 3}), geom.NewPointXYZ(nil, geom.XYZ{X: 1, Y: 2, Z: 3}), true},
		{"point XYZ differs only in Z", geom.NewPointXYZ(nil, geom.XYZ{X: 1, Y: 2, Z: 3}), geom.NewPointXYZ(nil, geom.XYZ{X: 1, Y: 2, Z: 4}), false},
		{"point XYM equal", geom.NewPointXYM(nil, geom.XYM{X: 1, Y: 2, M: 3}), geom.NewPointXYM(nil, geom.XYM{X: 1, Y: 2, M: 3}), true},
		{"point XYM differs only in M", geom.NewPointXYM(nil, geom.XYM{X: 1, Y: 2, M: 3}), geom.NewPointXYM(nil, geom.XYM{X: 1, Y: 2, M: 5}), false},
		{"point XYZM equal", geom.NewPointXYZM(nil, geom.XYZM{X: 1, Y: 2, Z: 3, M: 4}), geom.NewPointXYZM(nil, geom.XYZM{X: 1, Y: 2, Z: 3, M: 4}), true},
		{"point XYZM differs only in M", geom.NewPointXYZM(nil, geom.XYZM{X: 1, Y: 2, Z: 3, M: 4}), geom.NewPointXYZM(nil, geom.XYZM{X: 1, Y: 2, Z: 3, M: 9}), false},
		{"point layout mismatch XYZ vs XYM", geom.NewPointXYZ(nil, geom.XYZ{X: 1, Y: 2, Z: 3}), geom.NewPointXYM(nil, geom.XYM{X: 1, Y: 2, M: 3}), false},
		{"point layout mismatch XY vs XYZ", geom.NewPoint(nil, geom.XY{X: 1, Y: 2}), geom.NewPointXYZ(nil, geom.XYZ{X: 1, Y: 2}), false},
		{"point NaN vs NaN", geom.NewPointXYZ(nil, geom.XYZ{X: 1, Y: 2, Z: nan}), geom.NewPointXYZ(nil, geom.XYZ{X: 1, Y: 2, Z: nan}), true},
		{"point NaN vs number", geom.NewPointXYZ(nil, geom.XYZ{X: 1, Y: 2, Z: nan}), geom.NewPointXYZ(nil, geom.XYZ{X: 1, Y: 2, Z: 0}), false},

		// LineString
		{"linestring XY equal", ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 1}), ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 1}), true},
		{"linestring reordered", ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 1}), ls(geom.XY{X: 1, Y: 1}, geom.XY{X: 0, Y: 0}), false},
		{"linestring different length", ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 1}), ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 1}, geom.XY{X: 2, Y: 2}), false},
		{"linestring XYZ equal",
			geom.NewLineString(nil, []geom.XYZ{{0, 0, 1}, {1, 1, 2}}),
			geom.NewLineString(nil, []geom.XYZ{{0, 0, 1}, {1, 1, 2}}), true},
		{"linestring XYZ differs only in Z",
			geom.NewLineString(nil, []geom.XYZ{{0, 0, 1}, {1, 1, 2}}),
			geom.NewLineString(nil, []geom.XYZ{{0, 0, 1}, {1, 1, 7}}), false},
		{"linestring XYM equal",
			geom.NewLineString(nil, []geom.XYM{{0, 0, 1}, {1, 1, 2}}),
			geom.NewLineString(nil, []geom.XYM{{0, 0, 1}, {1, 1, 2}}), true},
		{"linestring XYZM equal",
			geom.NewLineString(nil, []geom.XYZM{{0, 0, 1, 2}, {1, 1, 3, 4}}),
			geom.NewLineString(nil, []geom.XYZM{{0, 0, 1, 2}, {1, 1, 3, 4}}), true},
		{"linestring XYZ vs XYM same numbers",
			geom.NewLineString(nil, []geom.XYZ{{0, 0, 1}, {1, 1, 2}}),
			geom.NewLineString(nil, []geom.XYM{{0, 0, 1}, {1, 1, 2}}), false},
		{"linestring NaN vs NaN",
			geom.NewLineString(nil, []geom.XYM{{0, 0, nan}, {1, 1, nan}}),
			geom.NewLineString(nil, []geom.XYM{{0, 0, nan}, {1, 1, nan}}), true},

		// LinearRing and type mismatch
		{"linearring equal", geom.NewLinearRing(nil, sq), geom.NewLinearRing(nil, sq), true},
		{"linestring vs linearring same coords", geom.NewLineString(nil, sq), geom.NewLinearRing(nil, sq), false},
		{"point vs multipoint same coords", geom.NewPoint(nil, geom.XY{X: 1, Y: 2}), geom.NewMultiPoint(nil, []geom.XY{{1, 2}}), false},

		// MultiPoint
		{"multipoint XYZ equal",
			geom.NewMultiPoint(nil, []geom.XYZ{{1, 2, 3}, {4, 5, 6}}),
			geom.NewMultiPoint(nil, []geom.XYZ{{1, 2, 3}, {4, 5, 6}}), true},
		{"multipoint reordered",
			geom.NewMultiPoint(nil, []geom.XYZ{{1, 2, 3}, {4, 5, 6}}),
			geom.NewMultiPoint(nil, []geom.XYZ{{4, 5, 6}, {1, 2, 3}}), false},

		// Polygon
		{"polygon equal", geom.NewPolygon(nil, sq, hole), geom.NewPolygon(nil, sq, hole), true},
		{"polygon missing hole", geom.NewPolygon(nil, sq, hole), geom.NewPolygon(nil, sq), false},
		{"polygon same coords different ring split",
			geom.NewPolygon(nil, sq, hole),
			geom.NewPolygon(nil, sqHole[:4], sqHole[4:]), false},
		{"polygon same coords one ring vs two",
			geom.NewPolygon(nil, sq, hole),
			geom.NewPolygon(nil, sqHole), false},
		{"polygon XYZM differs only in Z",
			geom.NewPolygon(nil, []geom.XYZM{{0, 0, 1, 1}, {1, 0, 1, 1}, {1, 1, 1, 1}, {0, 0, 1, 1}}),
			geom.NewPolygon(nil, []geom.XYZM{{0, 0, 1, 1}, {1, 0, 2, 1}, {1, 1, 1, 1}, {0, 0, 1, 1}}), false},

		// MultiLineString
		{"multilinestring equal",
			geom.NewMultiLineString(nil, ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 1}), ls(geom.XY{X: 2, Y: 2}, geom.XY{X: 3, Y: 3})),
			geom.NewMultiLineString(nil, ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 1}), ls(geom.XY{X: 2, Y: 2}, geom.XY{X: 3, Y: 3})), true},
		{"multilinestring members reordered",
			geom.NewMultiLineString(nil, ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 1}), ls(geom.XY{X: 2, Y: 2}, geom.XY{X: 3, Y: 3})),
			geom.NewMultiLineString(nil, ls(geom.XY{X: 2, Y: 2}, geom.XY{X: 3, Y: 3}), ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 1})), false},
		{"multilinestring different member count",
			geom.NewMultiLineString(nil, ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 1})),
			geom.NewMultiLineString(nil, ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 1}), ls(geom.XY{X: 2, Y: 2}, geom.XY{X: 3, Y: 3})), false},
		{"multilinestring same coords different split",
			geom.NewMultiLineString(nil, ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 1}), ls(geom.XY{X: 2, Y: 2}, geom.XY{X: 3, Y: 3})),
			geom.NewMultiLineString(nil, ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 1}, geom.XY{X: 2, Y: 2}), ls(geom.XY{X: 3, Y: 3})), false},

		// MultiPolygon
		{"multipolygon equal",
			geom.NewMultiPolygon(nil, geom.NewPolygon(nil, sq, hole), geom.NewPolygon(nil, hole)),
			geom.NewMultiPolygon(nil, geom.NewPolygon(nil, sq, hole), geom.NewPolygon(nil, hole)), true},
		{"multipolygon ring moved between members",
			geom.NewMultiPolygon(nil, geom.NewPolygon(nil, sq, hole), geom.NewPolygon(nil, hole)),
			geom.NewMultiPolygon(nil, geom.NewPolygon(nil, sq), geom.NewPolygon(nil, hole, hole)), false},

		// GeometryCollection
		{"collection equal",
			geom.NewGeometryCollection(nil, geom.NewPoint(nil, geom.XY{X: 1, Y: 2}), geom.NewPolygon(nil, sq)),
			geom.NewGeometryCollection(nil, geom.NewPoint(nil, geom.XY{X: 1, Y: 2}), geom.NewPolygon(nil, sq)), true},
		{"collection member type differs",
			geom.NewGeometryCollection(nil, geom.NewLineString(nil, sq)),
			geom.NewGeometryCollection(nil, geom.NewLinearRing(nil, sq)), false},
		{"collection nested",
			geom.NewGeometryCollection(nil, geom.NewGeometryCollection(nil, geom.NewPoint(nil, geom.XY{X: 1, Y: 2}))),
			geom.NewGeometryCollection(nil, geom.NewGeometryCollection(nil, geom.NewPoint(nil, geom.XY{X: 1, Y: 2}))), true},
		{"collection nested differs",
			geom.NewGeometryCollection(nil, geom.NewGeometryCollection(nil, geom.NewPoint(nil, geom.XY{X: 1, Y: 2}))),
			geom.NewGeometryCollection(nil, geom.NewGeometryCollection(nil, geom.NewPoint(nil, geom.XY{X: 1, Y: 3}))), false},
		{"collection vs multipoint", geom.NewGeometryCollection(nil, geom.NewPoint(nil, geom.XY{X: 1, Y: 2})), geom.NewMultiPoint(nil, []geom.XY{{1, 2}}), false},

		// Empty
		{"empty point vs empty point", geom.NewEmptyPoint(nil, geom.LayoutXY), geom.NewEmptyPoint(nil, geom.LayoutXY), true},
		{"empty point XYZ vs empty point XYZ", geom.NewEmptyPoint(nil, geom.LayoutXYZ), geom.NewEmptyPoint(nil, geom.LayoutXYZ), true},
		{"empty point layout mismatch", geom.NewEmptyPoint(nil, geom.LayoutXY), geom.NewEmptyPoint(nil, geom.LayoutXYZ), false},
		{"empty point vs NaN point", geom.NewEmptyPoint(nil, geom.LayoutXY), geom.NewPoint(nil, geom.XY{X: nan, Y: nan}), false},
		{"empty point vs point", geom.NewEmptyPoint(nil, geom.LayoutXY), geom.NewPoint(nil, geom.XY{X: 1, Y: 2}), false},
		{"empty linestring vs empty linestring", geom.NewEmptyLineString(nil, geom.LayoutXYM), geom.NewEmptyLineString(nil, geom.LayoutXYM), true},
		{"empty linestring vs typed-nil-slice linestring", geom.NewEmptyLineString(nil, geom.LayoutXY), geom.NewLineString(nil, []geom.XY(nil)), true},
		{"empty linestring vs linestring", geom.NewEmptyLineString(nil, geom.LayoutXY), ls(geom.XY{X: 0, Y: 0}, geom.XY{X: 1, Y: 1}), false},
		{"empty linestring vs empty point", geom.NewEmptyLineString(nil, geom.LayoutXY), geom.NewEmptyPoint(nil, geom.LayoutXY), false},
		{"empty polygon vs empty polygon", geom.NewEmptyPolygon(nil, geom.LayoutXY), geom.NewEmptyPolygon(nil, geom.LayoutXY), true},
		{"empty polygon vs zero-ring polygon", geom.NewEmptyPolygon(nil, geom.LayoutXY), geom.NewPolygon[geom.XY](nil), true},
		{"empty polygon vs polygon", geom.NewEmptyPolygon(nil, geom.LayoutXY), geom.NewPolygon(nil, sq), false},
		{"empty multipoint vs empty multipoint", geom.NewEmptyMultiPoint(nil, geom.LayoutXY), geom.NewEmptyMultiPoint(nil, geom.LayoutXY), true},
		{"empty multilinestring vs empty", geom.NewEmptyMultiLineString(nil, geom.LayoutXY), geom.NewEmptyMultiLineString(nil, geom.LayoutXY), true},
		{"empty multipolygon vs multipolygon", geom.NewEmptyMultiPolygon(nil, geom.LayoutXY), geom.NewMultiPolygon(nil, geom.NewPolygon(nil, sq)), false},
		{"empty collection vs empty collection", geom.NewEmptyGeometryCollection(nil, geom.LayoutXY), geom.NewEmptyGeometryCollection(nil, geom.LayoutXY), true},
		{"empty collection vs collection", geom.NewEmptyGeometryCollection(nil, geom.LayoutXY), geom.NewGeometryCollection(nil, geom.NewPoint(nil, geom.XY{X: 1, Y: 2})), false},

		// CRS is ignored
		{"point different CRS", geom.NewPoint(crs.WGS84, geom.XY{X: 1, Y: 2}), geom.NewPoint(crs.WebMercator, geom.XY{X: 1, Y: 2}), true},
		{"point CRS vs nil CRS", geom.NewPoint(crs.WGS84, geom.XY{X: 1, Y: 2}), geom.NewPoint(nil, geom.XY{X: 1, Y: 2}), true},
		{"polygon different CRS", geom.NewPolygon(crs.WGS84, sq, hole), geom.NewPolygon(crs.WebMercator, sq, hole), true},
		{"collection different CRS",
			geom.NewGeometryCollection(crs.WGS84, geom.NewPoint(crs.WGS84, geom.XY{X: 1, Y: 2})),
			geom.NewGeometryCollection(crs.WebMercator, geom.NewPoint(crs.WebMercator, geom.XY{X: 1, Y: 2})), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := geom.EqualsExact(tt.a, tt.b); got != tt.want {
				t.Errorf("EqualsExact(a, b) = %v, want %v", got, tt.want)
			}
			if got := geom.EqualsExact(tt.b, tt.a); got != tt.want {
				t.Errorf("EqualsExact(b, a) = %v, want %v (not symmetric)", got, tt.want)
			}
		})
	}
}

// TestEqualsExactReflexive checks every geometry equals itself (including
// NaN ordinates, which a plain == comparison would reject).
func TestEqualsExactReflexive(t *testing.T) {
	nan := math.NaN()
	for _, g := range []geom.Geometry{
		geom.NewPointXYZM(nil, geom.XYZM{X: 1, Y: nan, Z: 3, M: nan}),
		geom.NewLineString(nil, []geom.XYZ{{0, 0, nan}, {1, 1, 1}}),
		geom.NewPolygon(nil, []geom.XY{{0, 0}, {1, 0}, {1, 1}, {0, 0}}),
		geom.NewEmptyGeometryCollection(nil, geom.LayoutXYZM),
	} {
		if !geom.EqualsExact(g, g) {
			t.Errorf("EqualsExact(g, g) = false for %v", g.Type())
		}
	}
}
