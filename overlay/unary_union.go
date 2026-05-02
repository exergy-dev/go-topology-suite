package overlay

import (
	"github.com/terra-geo/terra/crs"
	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/predicate"
)

// UnaryUnion returns the union of g with itself: deduplicates and merges
// any overlapping or touching members of a Multi* or GeometryCollection
// into a single canonical representative.
//
// Behaviour by type:
//   - Point / LineString / Polygon: returned as-is.
//   - MultiPoint: deduplicated; result is Point if a single point survives.
//   - MultiLineString: returned as-is (proper noding deferred).
//   - MultiPolygon: pairwise union of members.
//   - GeometryCollection: members are partitioned by dimension; areal
//     members are pairwise unioned, lineal and pointal members are
//     carried through unchanged. The result is a GeometryCollection iff
//     more than one dimensional class survives.
func UnaryUnion(g geom.Geometry) (geom.Geometry, error) {
	if g == nil || g.IsEmpty() {
		return g, nil
	}
	g = unwrapLinearRing(g)
	switch v := g.(type) {
	case *geom.Point, *geom.LineString, *geom.Polygon, *geom.MultiLineString:
		return v, nil
	case *geom.MultiPoint:
		return dedupeMultiPoint(v), nil
	case *geom.MultiPolygon:
		polys := make([]geom.Geometry, 0, v.NumGeometries())
		for i := 0; i < v.NumGeometries(); i++ {
			p := v.PolygonAt(i)
			if !p.IsEmpty() {
				polys = append(polys, p)
			}
		}
		return unionAllAreal(v.CRS(), polys)
	case *geom.GeometryCollection:
		return unionGeometryCollection(v)
	}
	return g, nil
}

func dedupeMultiPoint(mp *geom.MultiPoint) geom.Geometry {
	seen := map[geom.XY]struct{}{}
	pts := make([]geom.XY, 0, mp.NumGeometries())
	for i := 0; i < mp.NumGeometries(); i++ {
		p := mp.PointAt(i)
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		pts = append(pts, p)
	}
	switch len(pts) {
	case 0:
		return geom.NewEmptyPoint(mp.CRS(), geom.LayoutXY)
	case 1:
		return geom.NewPoint(mp.CRS(), pts[0])
	default:
		return geom.NewMultiPoint(mp.CRS(), pts)
	}
}

func unionAllAreal(c *crs.CRS, polys []geom.Geometry) (geom.Geometry, error) {
	if len(polys) == 0 {
		return geom.NewEmptyPolygon(c, geom.LayoutXY), nil
	}
	acc := polys[0]
	for i := 1; i < len(polys); i++ {
		next, err := Union(acc, polys[i])
		if err != nil {
			return nil, err
		}
		acc = next
	}
	return acc, nil
}

func unionGeometryCollection(gc *geom.GeometryCollection) (geom.Geometry, error) {
	var polys, lines, points []geom.Geometry
	for i := 0; i < gc.NumGeometries(); i++ {
		m := gc.GeometryAt(i)
		if m.IsEmpty() {
			continue
		}
		switch v := m.(type) {
		case *geom.Polygon:
			polys = append(polys, v)
		case *geom.MultiPolygon:
			for j := 0; j < v.NumGeometries(); j++ {
				if !v.PolygonAt(j).IsEmpty() {
					polys = append(polys, v.PolygonAt(j))
				}
			}
		case *geom.LineString:
			lines = append(lines, v)
		case *geom.MultiLineString:
			for j := 0; j < v.NumGeometries(); j++ {
				if !v.LineStringAt(j).IsEmpty() {
					lines = append(lines, v.LineStringAt(j))
				}
			}
		case *geom.Point:
			points = append(points, v)
		case *geom.MultiPoint:
			for j := 0; j < v.NumGeometries(); j++ {
				points = append(points, geom.NewPoint(v.CRS(), v.PointAt(j)))
			}
		case *geom.GeometryCollection:
			sub, err := unionGeometryCollection(v)
			if err != nil {
				return nil, err
			}
			switch sv := sub.(type) {
			case *geom.Polygon:
				polys = append(polys, sv)
			case *geom.MultiPolygon:
				for j := 0; j < sv.NumGeometries(); j++ {
					polys = append(polys, sv.PolygonAt(j))
				}
			case *geom.LineString:
				lines = append(lines, sv)
			case *geom.MultiLineString:
				for j := 0; j < sv.NumGeometries(); j++ {
					lines = append(lines, sv.LineStringAt(j))
				}
			case *geom.Point:
				points = append(points, sv)
			case *geom.MultiPoint:
				for j := 0; j < sv.NumGeometries(); j++ {
					points = append(points, geom.NewPoint(sv.CRS(), sv.PointAt(j)))
				}
			}
		}
	}

	var areal geom.Geometry
	if len(polys) > 0 {
		a, err := unionAllAreal(gc.CRS(), polys)
		if err != nil {
			return nil, err
		}
		areal = a
	}

	// Absorption: drop lineal members that are fully covered by the
	// areal union, since their topology is subsumed.
	if areal != nil && len(lines) > 0 {
		filtered := lines[:0]
		for _, l := range lines {
			if covered, err := predicate.Covers(areal, l); err != nil || !covered {
				filtered = append(filtered, l)
			}
		}
		lines = filtered
	}
	// Drop pointal members covered by the areal or lineal union.
	if (areal != nil || len(lines) > 0) && len(points) > 0 {
		filtered := points[:0]
		for _, p := range points {
			absorbed := false
			if areal != nil {
				if covered, err := predicate.Covers(areal, p); err == nil && covered {
					absorbed = true
				}
			}
			if !absorbed {
				for _, l := range lines {
					if covered, err := predicate.Covers(l, p); err == nil && covered {
						absorbed = true
						break
					}
				}
			}
			if !absorbed {
				filtered = append(filtered, p)
			}
		}
		points = filtered
	}

	var lineal geom.Geometry
	if len(lines) > 0 {
		ls := make([]*geom.LineString, len(lines))
		for i, l := range lines {
			ls[i] = l.(*geom.LineString)
		}
		if len(ls) == 1 {
			lineal = ls[0]
		} else {
			lineal = geom.NewMultiLineString(gc.CRS(), ls...)
		}
	}
	var pointal geom.Geometry
	if len(points) > 0 {
		seen := map[geom.XY]struct{}{}
		var pts []geom.XY
		for _, p := range points {
			xy := p.(*geom.Point).XY()
			if _, ok := seen[xy]; ok {
				continue
			}
			seen[xy] = struct{}{}
			pts = append(pts, xy)
		}
		switch len(pts) {
		case 1:
			pointal = geom.NewPoint(gc.CRS(), pts[0])
		default:
			pointal = geom.NewMultiPoint(gc.CRS(), pts)
		}
	}

	var members []geom.Geometry
	if areal != nil {
		members = append(members, areal)
	}
	if lineal != nil {
		members = append(members, lineal)
	}
	if pointal != nil {
		members = append(members, pointal)
	}
	switch len(members) {
	case 0:
		return geom.NewGeometryCollection(gc.CRS()), nil
	case 1:
		return members[0], nil
	default:
		return geom.NewGeometryCollection(gc.CRS(), members...), nil
	}
}

func init() {
	predicate.SetUnaryUnion(UnaryUnion)
}
