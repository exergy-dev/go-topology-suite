package algorithm

import (
	"math"

	"github.com/go-topology-suite/gts/geom"
)

// Area computes the area of a geometry.
func Area(g geom.Geometry) float64 {
	switch v := g.(type) {
	case *geom.Point:
		return 0
	case *geom.LineString:
		return 0
	case *geom.LinearRing:
		return RingArea(v.Coordinates())
	case *geom.Polygon:
		return PolygonArea(v)
	case *geom.MultiPoint:
		return 0
	case *geom.MultiLineString:
		return 0
	case *geom.MultiPolygon:
		return MultiPolygonArea(v)
	case *geom.GeometryCollection:
		total := 0.0
		for i := 0; i < v.NumGeometries(); i++ {
			total += Area(v.GeometryN(i))
		}
		return total
	default:
		return 0
	}
}

// RingArea computes the area of a coordinate ring.
func RingArea(ring geom.CoordinateSequence) float64 {
	return math.Abs(SignedArea(ring))
}

// PolygonArea computes the area of a polygon (exterior - holes).
func PolygonArea(p *geom.Polygon) float64 {
	if p.IsEmpty() {
		return 0
	}

	area := RingArea(p.ExteriorRing().Coordinates())
	for i := 0; i < p.NumInteriorRings(); i++ {
		area -= RingArea(p.InteriorRingN(i).Coordinates())
	}
	return area
}

// MultiPolygonArea computes the total area of a multi-polygon.
func MultiPolygonArea(mp *geom.MultiPolygon) float64 {
	total := 0.0
	for i := 0; i < mp.NumGeometries(); i++ {
		total += PolygonArea(mp.GeometryN(i).(*geom.Polygon))
	}
	return total
}

// Length computes the length of a geometry.
func Length(g geom.Geometry) float64 {
	switch v := g.(type) {
	case *geom.Point:
		return 0
	case *geom.LineString:
		return LineLength(v.Coordinates())
	case *geom.LinearRing:
		return LineLength(v.Coordinates())
	case *geom.Polygon:
		return PolygonPerimeter(v)
	case *geom.MultiPoint:
		return 0
	case *geom.MultiLineString:
		total := 0.0
		for i := 0; i < v.NumGeometries(); i++ {
			total += LineLength(v.GeometryN(i).(*geom.LineString).Coordinates())
		}
		return total
	case *geom.MultiPolygon:
		total := 0.0
		for i := 0; i < v.NumGeometries(); i++ {
			total += PolygonPerimeter(v.GeometryN(i).(*geom.Polygon))
		}
		return total
	case *geom.GeometryCollection:
		total := 0.0
		for i := 0; i < v.NumGeometries(); i++ {
			total += Length(v.GeometryN(i))
		}
		return total
	default:
		return 0
	}
}

// LineLength computes the length of a coordinate sequence.
func LineLength(coords geom.CoordinateSequence) float64 {
	if len(coords) < 2 {
		return 0
	}
	length := 0.0
	for i := 1; i < len(coords); i++ {
		length += coords[i-1].Distance(coords[i])
	}
	return length
}

// PolygonPerimeter computes the perimeter of a polygon.
func PolygonPerimeter(p *geom.Polygon) float64 {
	if p.IsEmpty() {
		return 0
	}
	perimeter := LineLength(p.ExteriorRing().Coordinates())
	for i := 0; i < p.NumInteriorRings(); i++ {
		perimeter += LineLength(p.InteriorRingN(i).Coordinates())
	}
	return perimeter
}

// Centroid computes the centroid of a geometry.
func Centroid(g geom.Geometry) geom.Coordinate {
	switch v := g.(type) {
	case *geom.Point:
		return v.Coordinate()
	case *geom.LineString:
		return LineCentroid(v.Coordinates())
	case *geom.LinearRing:
		return RingCentroid(v.Coordinates())
	case *geom.Polygon:
		return PolygonCentroid(v)
	case *geom.MultiPoint:
		return MultiPointCentroid(v)
	case *geom.MultiLineString:
		return MultiLineStringCentroid(v)
	case *geom.MultiPolygon:
		return MultiPolygonCentroid(v)
	case *geom.GeometryCollection:
		return GeometryCollectionCentroid(v)
	default:
		return geom.Coordinate{X: math.NaN(), Y: math.NaN()}
	}
}

// LineCentroid computes the centroid of a line string.
func LineCentroid(coords geom.CoordinateSequence) geom.Coordinate {
	if len(coords) == 0 {
		return geom.Coordinate{X: math.NaN(), Y: math.NaN()}
	}
	if len(coords) == 1 {
		return coords[0].Clone()
	}

	totalLength := 0.0
	sumX, sumY := 0.0, 0.0

	for i := 1; i < len(coords); i++ {
		p1, p2 := coords[i-1], coords[i]
		segLen := p1.Distance(p2)
		midX := (p1.X + p2.X) / 2
		midY := (p1.Y + p2.Y) / 2
		sumX += midX * segLen
		sumY += midY * segLen
		totalLength += segLen
	}

	if totalLength == 0 {
		return coords[0].Clone()
	}

	return geom.NewCoordinate(sumX/totalLength, sumY/totalLength)
}

// RingCentroid computes the centroid of a ring using the polygon centroid formula.
func RingCentroid(coords geom.CoordinateSequence) geom.Coordinate {
	if len(coords) < 3 {
		return LineCentroid(coords)
	}

	n := len(coords)
	if coords.IsClosed(geom.DefaultEpsilon) {
		n-- // Exclude closing point
	}

	sumX, sumY := 0.0, 0.0
	signedArea := 0.0

	for i := 0; i < n; i++ {
		x0, y0 := coords[i].X, coords[i].Y
		x1, y1 := coords[(i+1)%n].X, coords[(i+1)%n].Y
		cross := x0*y1 - x1*y0
		signedArea += cross
		sumX += (x0 + x1) * cross
		sumY += (y0 + y1) * cross
	}

	signedArea /= 2
	if math.Abs(signedArea) < geom.DefaultEpsilon {
		return LineCentroid(coords)
	}

	return geom.NewCoordinate(sumX/(6*signedArea), sumY/(6*signedArea))
}

// PolygonCentroid computes the centroid of a polygon.
func PolygonCentroid(p *geom.Polygon) geom.Coordinate {
	if p.IsEmpty() {
		return geom.Coordinate{X: math.NaN(), Y: math.NaN()}
	}

	shellCentroid := RingCentroid(p.ExteriorRing().Coordinates())
	shellArea := RingArea(p.ExteriorRing().Coordinates())

	if p.NumInteriorRings() == 0 {
		return shellCentroid
	}

	// Weighted average considering holes
	totalArea := shellArea
	sumX := shellCentroid.X * shellArea
	sumY := shellCentroid.Y * shellArea

	for i := 0; i < p.NumInteriorRings(); i++ {
		holeCoords := p.InteriorRingN(i).Coordinates()
		holeCentroid := RingCentroid(holeCoords)
		holeArea := RingArea(holeCoords)
		totalArea -= holeArea
		sumX -= holeCentroid.X * holeArea
		sumY -= holeCentroid.Y * holeArea
	}

	if math.Abs(totalArea) < geom.DefaultEpsilon {
		return shellCentroid
	}

	return geom.NewCoordinate(sumX/totalArea, sumY/totalArea)
}

// MultiPointCentroid computes the centroid of a multi-point.
func MultiPointCentroid(mp *geom.MultiPoint) geom.Coordinate {
	if mp.IsEmpty() {
		return geom.Coordinate{X: math.NaN(), Y: math.NaN()}
	}

	sumX, sumY := 0.0, 0.0
	n := mp.NumGeometries()

	for i := 0; i < n; i++ {
		p := mp.GeometryN(i).(*geom.Point)
		sumX += p.X()
		sumY += p.Y()
	}

	return geom.NewCoordinate(sumX/float64(n), sumY/float64(n))
}

// MultiLineStringCentroid computes the centroid of a multi-linestring.
func MultiLineStringCentroid(mls *geom.MultiLineString) geom.Coordinate {
	if mls.IsEmpty() {
		return geom.Coordinate{X: math.NaN(), Y: math.NaN()}
	}

	totalLength := 0.0
	sumX, sumY := 0.0, 0.0

	for i := 0; i < mls.NumGeometries(); i++ {
		ls := mls.GeometryN(i).(*geom.LineString)
		coords := ls.Coordinates()
		length := LineLength(coords)
		centroid := LineCentroid(coords)
		sumX += centroid.X * length
		sumY += centroid.Y * length
		totalLength += length
	}

	if totalLength == 0 {
		// Fall back to first point
		return mls.GeometryN(0).Coordinates()[0]
	}

	return geom.NewCoordinate(sumX/totalLength, sumY/totalLength)
}

// MultiPolygonCentroid computes the centroid of a multi-polygon.
func MultiPolygonCentroid(mp *geom.MultiPolygon) geom.Coordinate {
	if mp.IsEmpty() {
		return geom.Coordinate{X: math.NaN(), Y: math.NaN()}
	}

	totalArea := 0.0
	sumX, sumY := 0.0, 0.0

	for i := 0; i < mp.NumGeometries(); i++ {
		p := mp.GeometryN(i).(*geom.Polygon)
		area := PolygonArea(p)
		centroid := PolygonCentroid(p)
		sumX += centroid.X * area
		sumY += centroid.Y * area
		totalArea += area
	}

	if totalArea == 0 {
		// Fall back to first polygon's centroid
		return PolygonCentroid(mp.GeometryN(0).(*geom.Polygon))
	}

	return geom.NewCoordinate(sumX/totalArea, sumY/totalArea)
}

// GeometryCollectionCentroid computes the centroid of a geometry collection.
func GeometryCollectionCentroid(gc *geom.GeometryCollection) geom.Coordinate {
	if gc.IsEmpty() {
		return geom.Coordinate{X: math.NaN(), Y: math.NaN()}
	}

	// Weight by dimension: areas > lines > points
	var areaWeighted, lineWeighted []geom.Coordinate
	var areaWeights, lineWeights []float64
	var points []geom.Coordinate

	for i := 0; i < gc.NumGeometries(); i++ {
		g := gc.GeometryN(i)
		switch v := g.(type) {
		case *geom.Polygon:
			c := PolygonCentroid(v)
			areaWeighted = append(areaWeighted, c)
			areaWeights = append(areaWeights, PolygonArea(v))
		case *geom.MultiPolygon:
			c := MultiPolygonCentroid(v)
			areaWeighted = append(areaWeighted, c)
			areaWeights = append(areaWeights, MultiPolygonArea(v))
		case *geom.LineString:
			c := LineCentroid(v.Coordinates())
			lineWeighted = append(lineWeighted, c)
			lineWeights = append(lineWeights, v.Length())
		case *geom.MultiLineString:
			c := MultiLineStringCentroid(v)
			lineWeighted = append(lineWeighted, c)
			lineWeights = append(lineWeights, v.Length())
		case *geom.Point:
			points = append(points, v.Coordinate())
		case *geom.MultiPoint:
			for j := 0; j < v.NumGeometries(); j++ {
				points = append(points, v.GeometryN(j).(*geom.Point).Coordinate())
			}
		}
	}

	// Use highest dimension available
	if len(areaWeighted) > 0 {
		return weightedCentroid(areaWeighted, areaWeights)
	}
	if len(lineWeighted) > 0 {
		return weightedCentroid(lineWeighted, lineWeights)
	}
	if len(points) > 0 {
		sumX, sumY := 0.0, 0.0
		for _, p := range points {
			sumX += p.X
			sumY += p.Y
		}
		return geom.NewCoordinate(sumX/float64(len(points)), sumY/float64(len(points)))
	}

	return geom.Coordinate{X: math.NaN(), Y: math.NaN()}
}

func weightedCentroid(centroids []geom.Coordinate, weights []float64) geom.Coordinate {
	totalWeight := 0.0
	sumX, sumY := 0.0, 0.0

	for i, c := range centroids {
		w := weights[i]
		sumX += c.X * w
		sumY += c.Y * w
		totalWeight += w
	}

	if totalWeight == 0 {
		return centroids[0]
	}

	return geom.NewCoordinate(sumX/totalWeight, sumY/totalWeight)
}
