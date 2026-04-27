package topology

import (
	"fmt"
	"math"
	"strings"

	"github.com/robert-malhotra/go-topology-suite/geom"
)

// PolygonsFromRings classifies closed rings as shells or holes, assigns holes
// to containing shells, and returns valid non-empty polygon results.
func PolygonsFromRings(rings []geom.CoordinateSequence, deduplicate bool) []*geom.Polygon {
	if len(rings) == 0 {
		return nil
	}
	if deduplicate {
		rings = DeduplicateRings(rings)
	}

	shells, holes := classifyRings(rings)
	if len(shells) == 0 && len(holes) == 0 {
		return nil
	}
	if len(shells) == 0 {
		shells, holes = promoteLargestHoleToShell(holes)
	}

	polygons := make([]*geom.Polygon, 0, len(shells))
	for _, shell := range shells {
		shellRing := geom.NewLinearRing(shell)
		shellPoly := geom.NewPolygon(shellRing, nil)

		var assignedHoles []*geom.LinearRing
		for _, hole := range holes {
			holePoint := RingInteriorPoint(hole)
			if PointLocationInPolygon(holePoint, shellPoly) == geom.LocationInterior {
				assignedHoles = append(assignedHoles, geom.NewLinearRing(hole))
			}
		}

		polygon := geom.NewPolygon(shellRing, assignedHoles)
		if !polygon.IsEmpty() && polygon.Area() > geom.DefaultEpsilon {
			polygons = append(polygons, polygon)
		}
	}
	return polygons
}

// DeduplicateRings removes duplicate rings independent of direction and start
// vertex.
func DeduplicateRings(rings []geom.CoordinateSequence) []geom.CoordinateSequence {
	if len(rings) <= 1 {
		return rings
	}

	unique := make([]geom.CoordinateSequence, 0, len(rings))
	seen := make(map[string]struct{})
	for _, ring := range rings {
		key := RingKey(ring)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, ring)
	}
	return unique
}

// RingKey creates a canonical key for a ring independent of direction and
// start vertex.
func RingKey(ring geom.CoordinateSequence) string {
	if len(ring) == 0 {
		return ""
	}

	n := ringCoordinateCount(ring)
	if n == 0 {
		return ""
	}
	minIdx := 0
	for i := 1; i < n; i++ {
		if ring[i].X < ring[minIdx].X ||
			(ring[i].X == ring[minIdx].X && ring[i].Y < ring[minIdx].Y) {
			minIdx = i
		}
	}

	var coords []string
	area := geom.SignedArea(ring)
	if area >= 0 {
		for i := 0; i < n; i++ {
			idx := (minIdx + i) % n
			coords = append(coords, fmt.Sprintf("%.10f,%.10f", ring[idx].X, ring[idx].Y))
		}
	} else {
		for i := 0; i < n; i++ {
			idx := (minIdx - i + n) % n
			coords = append(coords, fmt.Sprintf("%.10f,%.10f", ring[idx].X, ring[idx].Y))
		}
	}
	return strings.Join(coords, ";")
}

func classifyRings(rings []geom.CoordinateSequence) ([]geom.CoordinateSequence, []geom.CoordinateSequence) {
	var shells []geom.CoordinateSequence
	var holes []geom.CoordinateSequence
	for _, ring := range rings {
		if len(ring) == 0 {
			continue
		}
		if !ring.IsClosed(geom.DefaultEpsilon) {
			ring = append(ring, ring[0].Clone())
		}
		if len(ring) < 4 {
			continue
		}

		area := geom.SignedArea(ring)
		switch {
		case area > geom.DefaultEpsilon:
			shells = append(shells, ring)
		case area < -geom.DefaultEpsilon:
			holes = append(holes, ring)
		case math.Abs(area) <= geom.DefaultEpsilon:
			continue
		}
	}
	return shells, holes
}

func promoteLargestHoleToShell(holes []geom.CoordinateSequence) ([]geom.CoordinateSequence, []geom.CoordinateSequence) {
	if len(holes) == 0 {
		return nil, holes
	}
	largestIdx := 0
	largestArea := -geom.SignedArea(holes[0])
	for i := 1; i < len(holes); i++ {
		area := -geom.SignedArea(holes[i])
		if area > largestArea {
			largestArea = area
			largestIdx = i
		}
	}

	shells := []geom.CoordinateSequence{holes[largestIdx].Reverse()}
	holes = append(holes[:largestIdx], holes[largestIdx+1:]...)
	return shells, holes
}
