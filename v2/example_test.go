package topology_test

import (
	"fmt"

	"github.com/robert-malhotra/go-topology-suite/geom"
	"github.com/robert-malhotra/go-topology-suite/operation/buffer"
	topology "github.com/robert-malhotra/go-topology-suite/v2"
)

func ExampleNewPolygon() {
	polygon, err := topology.NewPolygon(topology.CoordinateSequence{
		topology.NewCoordinate(0, 0),
		topology.NewCoordinate(10, 0),
		topology.NewCoordinate(10, 10),
		topology.NewCoordinate(0, 10),
		topology.NewCoordinate(0, 0),
	}, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(polygon.GeometryType(), polygon.Area())
	// Output: Polygon 100
}

func ExampleOverlay() {
	left, err := topology.NewPolygon(topology.CoordinateSequence{
		topology.NewCoordinate(0, 0),
		topology.NewCoordinate(10, 0),
		topology.NewCoordinate(10, 10),
		topology.NewCoordinate(0, 10),
		topology.NewCoordinate(0, 0),
	}, nil)
	if err != nil {
		panic(err)
	}
	right, err := topology.NewPolygon(topology.CoordinateSequence{
		topology.NewCoordinate(5, 5),
		topology.NewCoordinate(15, 5),
		topology.NewCoordinate(15, 15),
		topology.NewCoordinate(5, 15),
		topology.NewCoordinate(5, 5),
	}, nil)
	if err != nil {
		panic(err)
	}

	intersection, err := topology.Intersection(left, right)
	if err != nil {
		panic(err)
	}

	fmt.Println(intersection.GeometryType(), intersection.(interface{ Area() float64 }).Area())
	// Output: Polygon 25
}

func ExampleOverlayOptions_precisionModel() {
	left, err := topology.NewPolygon(topology.CoordinateSequence{
		topology.NewCoordinate(0.2, 0.2),
		topology.NewCoordinate(4.2, 0.2),
		topology.NewCoordinate(4.2, 4.2),
		topology.NewCoordinate(0.2, 4.2),
		topology.NewCoordinate(0.2, 0.2),
	}, nil)
	if err != nil {
		panic(err)
	}
	right, err := topology.NewPolygon(topology.CoordinateSequence{
		topology.NewCoordinate(2.6, 2.6),
		topology.NewCoordinate(6.6, 2.6),
		topology.NewCoordinate(6.6, 6.6),
		topology.NewCoordinate(2.6, 6.6),
		topology.NewCoordinate(2.6, 2.6),
	}, nil)
	if err != nil {
		panic(err)
	}

	opts := topology.OverlayOptions{PrecisionModel: geom.NewFixedPrecision(1)}
	intersection, err := topology.Intersection(left, right, opts)
	if err != nil {
		panic(err)
	}
	union, err := topology.Union(left, right, opts)
	if err != nil {
		panic(err)
	}

	intersectionArea := intersection.(interface{ Area() float64 }).Area()
	unionArea := union.(interface{ Area() float64 }).Area()
	fmt.Println(intersection.GeometryType(), intersectionArea, union.GeometryType(), unionArea)
	// Output: Polygon 1 Polygon 31
}

func ExampleBuffer() {
	point, err := topology.NewPoint(0, 0)
	if err != nil {
		panic(err)
	}
	buffered, err := topology.Buffer(point, 5)
	if err != nil {
		panic(err)
	}

	fmt.Println(buffered.GeometryType(), buffered.IsValid())
	// Output: Polygon true
}

func ExampleBuffer_customParams() {
	line, err := topology.NewLineString(topology.CoordinateSequence{
		topology.NewCoordinate(0, 0),
		topology.NewCoordinate(10, 0),
	})
	if err != nil {
		panic(err)
	}

	params := buffer.DefaultParams()
	params.QuadrantSegments = 4
	params.EndCapStyle = buffer.CapFlat
	params.JoinStyle = buffer.JoinBevel

	buffered, err := topology.Buffer(line, 2, topology.BufferOptions{
		Params:          params,
		NormalizeResult: true,
	})
	if err != nil {
		panic(err)
	}

	params.QuadrantSegments = 0
	_, invalidErr := topology.Buffer(line, 2, topology.BufferOptions{Params: params})

	fmt.Println(buffered.GeometryType(), buffered.IsValid(), invalidErr != nil)
	// Output: Polygon true true
}

func ExampleRelate() {
	point, err := topology.NewPoint(1, 1)
	if err != nil {
		panic(err)
	}
	polygon, err := topology.NewPolygon(topology.CoordinateSequence{
		topology.NewCoordinate(0, 0),
		topology.NewCoordinate(10, 0),
		topology.NewCoordinate(10, 10),
		topology.NewCoordinate(0, 10),
		topology.NewCoordinate(0, 0),
	}, nil)
	if err != nil {
		panic(err)
	}

	matrix, err := topology.Relate(point, polygon)
	if err != nil {
		panic(err)
	}

	fmt.Println(matrix.Matches("T*F**F***"))
	// Output: true
}

func ExampleRelatePattern() {
	point, err := topology.NewPoint(1, 1)
	if err != nil {
		panic(err)
	}
	polygon, err := topology.NewPolygon(topology.CoordinateSequence{
		topology.NewCoordinate(0, 0),
		topology.NewCoordinate(10, 0),
		topology.NewCoordinate(10, 10),
		topology.NewCoordinate(0, 10),
		topology.NewCoordinate(0, 0),
	}, nil)
	if err != nil {
		panic(err)
	}

	matches, err := topology.RelatePattern(point, polygon, "T*F**F***")
	if err != nil {
		panic(err)
	}
	_, invalidErr := topology.RelatePattern(point, polygon, "T*F**F**X")

	fmt.Println(matches, invalidErr != nil)
	// Output: true true
}

func ExampleIntersects() {
	left, err := topology.NewPolygon(topology.CoordinateSequence{
		topology.NewCoordinate(0, 0),
		topology.NewCoordinate(10, 0),
		topology.NewCoordinate(10, 10),
		topology.NewCoordinate(0, 10),
		topology.NewCoordinate(0, 0),
	}, nil)
	if err != nil {
		panic(err)
	}
	right, err := topology.NewPolygon(topology.CoordinateSequence{
		topology.NewCoordinate(5, 5),
		topology.NewCoordinate(15, 5),
		topology.NewCoordinate(15, 15),
		topology.NewCoordinate(5, 15),
		topology.NewCoordinate(5, 5),
	}, nil)
	if err != nil {
		panic(err)
	}

	ok, err := topology.Intersects(left, right)
	if err != nil {
		panic(err)
	}

	fmt.Println(ok)
	// Output: true
}

func ExampleReadWKT() {
	geometry, err := topology.ReadWKT("POINT (1 2)")
	if err != nil {
		panic(err)
	}
	_, trailingErr := topology.ReadWKT("POINT (1 2) trailing")

	fmt.Println(geometry.GeometryType(), trailingErr != nil)
	// Output: Point true
}

func ExampleReadGeoJSON() {
	geometry, err := topology.ReadGeoJSON([]byte(`{"type":"Point","coordinates":[1,2]}`))
	if err != nil {
		panic(err)
	}

	data, err := topology.WriteGeoJSON(geometry)
	if err != nil {
		panic(err)
	}

	fmt.Println(geometry.GeometryType(), len(data) > 0)
	// Output: Point true
}

func ExampleReadKML() {
	geometry, err := topology.ReadKML([]byte(`<kml><Placemark><Point><coordinates>1,2</coordinates></Point></Placemark></kml>`))
	if err != nil {
		panic(err)
	}

	data, err := topology.WriteKML(geometry)
	if err != nil {
		panic(err)
	}

	fmt.Println(geometry.GeometryType(), len(data) > 0)
	// Output: Point true
}
