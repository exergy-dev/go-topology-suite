// Package buffer provides geometry buffer operations.
//
// Buffer operations create a geometry representing all points within a
// specified distance of the input geometry. Positive distances expand
// the geometry, negative distances shrink it (for polygonal geometries).
package buffer

import (
	"math"

	"github.com/go-topology-suite/gts/geom"
)

// CapStyle defines how line endpoints are handled in buffer operations.
type CapStyle int

const (
	// CapRound creates semicircular caps at line endpoints.
	CapRound CapStyle = iota
	// CapFlat creates flat caps that end exactly at the line endpoint.
	CapFlat
	// CapSquare creates square caps that extend beyond the line endpoint.
	CapSquare
)

// JoinStyle defines how corners are handled in buffer operations.
type JoinStyle int

const (
	// JoinRound creates rounded corners.
	JoinRound JoinStyle = iota
	// JoinMitre creates sharp corners (limited by mitre limit).
	JoinMitre
	// JoinBevel creates flat corners.
	JoinBevel
)

// Params contains parameters for buffer operations.
type Params struct {
	// QuadrantSegments is the number of segments used to approximate
	// a quarter circle. Default is 8.
	QuadrantSegments int

	// EndCapStyle defines how line endpoints are buffered.
	EndCapStyle CapStyle

	// JoinStyle defines how corners are buffered.
	JoinStyle JoinStyle

	// MitreLimit limits the length of mitre joins. Only used when
	// JoinStyle is JoinMitre. Default is 5.0.
	MitreLimit float64

	// SingleSided if true creates a single-sided buffer for lines.
	SingleSided bool
}

// DefaultParams returns default buffer parameters.
func DefaultParams() *Params {
	return &Params{
		QuadrantSegments: 8,
		EndCapStyle:      CapRound,
		JoinStyle:        JoinRound,
		MitreLimit:       5.0,
		SingleSided:      false,
	}
}

// Buffer computes the buffer of a geometry using default parameters.
func Buffer(g geom.Geometry, distance float64) geom.Geometry {
	return BufferWithParams(g, distance, DefaultParams())
}

// BufferWithParams computes the buffer of a geometry with custom parameters.
func BufferWithParams(g geom.Geometry, distance float64, params *Params) geom.Geometry {
	if g == nil || g.IsEmpty() {
		return geom.NewPolygonEmpty()
	}

	if params == nil {
		params = DefaultParams()
	}

	// Handle zero distance
	if math.Abs(distance) < geom.DefaultEpsilon {
		return g.Clone()
	}

	// Compute buffer based on geometry type
	switch v := g.(type) {
	case *geom.Point:
		return bufferPoint(v, distance, params)
	case *geom.LineString:
		return bufferLineString(v, distance, params)
	case *geom.LinearRing:
		return bufferLineString(v.LineString, distance, params)
	case *geom.Polygon:
		return bufferPolygon(v, distance, params)
	case *geom.MultiPoint:
		return bufferMultiPoint(v, distance, params)
	case *geom.MultiLineString:
		return bufferMultiLineString(v, distance, params)
	case *geom.MultiPolygon:
		return bufferMultiPolygon(v, distance, params)
	case *geom.GeometryCollection:
		return bufferGeometryCollection(v, distance, params)
	default:
		return geom.NewPolygonEmpty()
	}
}

// bufferPoint creates a circular buffer around a point.
func bufferPoint(p *geom.Point, distance float64, params *Params) geom.Geometry {
	if distance <= 0 {
		return geom.NewPolygonEmpty()
	}

	coord := p.Coordinate()
	numSegments := params.QuadrantSegments * 4

	coords := make(geom.CoordinateSequence, numSegments+1)
	for i := 0; i < numSegments; i++ {
		angle := 2 * math.Pi * float64(i) / float64(numSegments)
		x := coord.X + distance*math.Cos(angle)
		y := coord.Y + distance*math.Sin(angle)
		coords[i] = geom.NewCoordinate(x, y)
	}
	coords[numSegments] = coords[0].Clone() // Close the ring

	ring := geom.NewLinearRing(coords)
	return geom.NewPolygon(ring, nil)
}

// bufferLineString creates a buffer around a line string.
func bufferLineString(ls *geom.LineString, distance float64, params *Params) geom.Geometry {
	if ls.IsEmpty() || distance <= 0 {
		return geom.NewPolygonEmpty()
	}

	coords := ls.Coordinates()
	if len(coords) < 2 {
		return geom.NewPolygonEmpty()
	}

	// For a simple 2-point line, create a rectangle with end caps
	if len(coords) == 2 {
		return bufferSegment(coords[0], coords[1], distance, params)
	}

	// Build buffer polygon by creating offset curves and connecting them
	return buildLineBuffer(coords, distance, params)
}

// bufferSegment creates a buffer around a single line segment.
func bufferSegment(p0, p1 geom.Coordinate, distance float64, params *Params) geom.Geometry {
	dx := p1.X - p0.X
	dy := p1.Y - p0.Y
	length := math.Sqrt(dx*dx + dy*dy)

	if length < geom.DefaultEpsilon {
		// Degenerate segment - buffer as point
		return bufferPoint(geom.NewPointFromCoordinate(p0), distance, params)
	}

	// Normalize direction
	dx /= length
	dy /= length

	// Perpendicular direction (left side)
	px := -dy
	py := dx

	// Create the buffer rectangle with end caps
	var shellCoords geom.CoordinateSequence

	// Left side of the line (from p0 to p1)
	leftStart := geom.NewCoordinate(p0.X+px*distance, p0.Y+py*distance)
	leftEnd := geom.NewCoordinate(p1.X+px*distance, p1.Y+py*distance)

	// Right side of the line (from p1 to p0)
	rightStart := geom.NewCoordinate(p1.X-px*distance, p1.Y-py*distance)
	rightEnd := geom.NewCoordinate(p0.X-px*distance, p0.Y-py*distance)

	// Add left side
	shellCoords = append(shellCoords, leftStart, leftEnd)

	// Add end cap at p1
	endCap := computeCapPoints(p1, dx, dy, distance, params)
	shellCoords = append(shellCoords, endCap...)

	// Add right side
	shellCoords = append(shellCoords, rightStart, rightEnd)

	// Add start cap at p0
	startCap := computeCapPoints(p0, -dx, -dy, distance, params)
	shellCoords = append(shellCoords, startCap...)

	// Close the ring
	shellCoords = append(shellCoords, shellCoords[0].Clone())

	ring := geom.NewLinearRing(shellCoords)
	return geom.NewPolygon(ring, nil)
}

// buildLineBuffer builds a buffer polygon for a multi-segment line.
func buildLineBuffer(coords geom.CoordinateSequence, distance float64, params *Params) geom.Geometry {
	n := len(coords)
	if n < 2 {
		return geom.NewPolygonEmpty()
	}

	// Compute offset points for each segment on both sides
	var leftSide, rightSide geom.CoordinateSequence

	for i := 0; i < n-1; i++ {
		p0 := coords[i]
		p1 := coords[i+1]

		dx := p1.X - p0.X
		dy := p1.Y - p0.Y
		length := math.Sqrt(dx*dx + dy*dy)

		if length < geom.DefaultEpsilon {
			continue
		}

		// Normalize
		dx /= length
		dy /= length

		// Perpendicular (left)
		px := -dy
		py := dx

		// Left offset points
		lo0 := geom.NewCoordinate(p0.X+px*distance, p0.Y+py*distance)
		lo1 := geom.NewCoordinate(p1.X+px*distance, p1.Y+py*distance)

		// Right offset points
		ro0 := geom.NewCoordinate(p0.X-px*distance, p0.Y-py*distance)
		ro1 := geom.NewCoordinate(p1.X-px*distance, p1.Y-py*distance)

		if i == 0 {
			leftSide = append(leftSide, lo0)
			rightSide = append(rightSide, ro0)
		} else {
			// Add corner join on left side
			prevLeft := leftSide[len(leftSide)-1]
			leftCorner := computeJoinPoints(prevLeft, lo0, coords[i], distance, params)
			leftSide = append(leftSide, leftCorner...)

			// Add corner join on right side
			prevRight := rightSide[len(rightSide)-1]
			rightCorner := computeJoinPoints(prevRight, ro0, coords[i], -distance, params)
			rightSide = append(rightSide, rightCorner...)
		}

		leftSide = append(leftSide, lo1)
		rightSide = append(rightSide, ro1)
	}

	// Build the complete shell
	var shellCoords geom.CoordinateSequence

	// Add left side (forward)
	shellCoords = append(shellCoords, leftSide...)

	// Add end cap
	lastPt := coords[n-1]
	prevPt := coords[n-2]
	dx := lastPt.X - prevPt.X
	dy := lastPt.Y - prevPt.Y
	length := math.Sqrt(dx*dx + dy*dy)
	if length > geom.DefaultEpsilon {
		dx /= length
		dy /= length
		endCap := computeCapPoints(lastPt, dx, dy, distance, params)
		shellCoords = append(shellCoords, endCap...)
	}

	// Add right side (reversed)
	for i := len(rightSide) - 1; i >= 0; i-- {
		shellCoords = append(shellCoords, rightSide[i])
	}

	// Add start cap
	firstPt := coords[0]
	nextPt := coords[1]
	dx = firstPt.X - nextPt.X
	dy = firstPt.Y - nextPt.Y
	length = math.Sqrt(dx*dx + dy*dy)
	if length > geom.DefaultEpsilon {
		dx /= length
		dy /= length
		startCap := computeCapPoints(firstPt, dx, dy, distance, params)
		shellCoords = append(shellCoords, startCap...)
	}

	// Close the ring
	if len(shellCoords) > 0 {
		shellCoords = append(shellCoords, shellCoords[0].Clone())
	}

	if len(shellCoords) < 4 {
		return geom.NewPolygonEmpty()
	}

	ring := geom.NewLinearRing(shellCoords)
	return geom.NewPolygon(ring, nil)
}

// computeJoinPoints computes the points for a corner join.
func computeJoinPoints(prevOffset, nextOffset, vertex geom.Coordinate, distance float64, params *Params) geom.CoordinateSequence {
	// Check if the offset points are close enough (no corner needed)
	if prevOffset.Distance(nextOffset) < geom.DefaultEpsilon*2 {
		return geom.CoordinateSequence{}
	}

	// Determine if this is a convex or concave corner
	// For a convex corner (outside of turn), we add points
	// For a concave corner (inside of turn), we just connect

	// Vector from vertex to prevOffset
	v1x := prevOffset.X - vertex.X
	v1y := prevOffset.Y - vertex.Y

	// Vector from vertex to nextOffset
	v2x := nextOffset.X - vertex.X
	v2y := nextOffset.Y - vertex.Y

	// Cross product to determine turn direction
	cross := v1x*v2y - v1y*v2x

	// If cross and distance have same sign, it's convex (need fillet)
	// If opposite signs, it's concave (just connect)
	isConvex := (cross > 0 && distance > 0) || (cross < 0 && distance < 0)

	if !isConvex {
		// Concave corner - just return next offset point
		return geom.CoordinateSequence{nextOffset}
	}

	switch params.JoinStyle {
	case JoinRound:
		return computeRoundJoin(prevOffset, nextOffset, vertex, math.Abs(distance), params)
	case JoinMitre:
		return computeMitreJoin(prevOffset, nextOffset, vertex, math.Abs(distance), params)
	case JoinBevel:
		return geom.CoordinateSequence{nextOffset}
	default:
		return geom.CoordinateSequence{nextOffset}
	}
}

// computeRoundJoin computes a rounded corner join.
func computeRoundJoin(p0, p1, vertex geom.Coordinate, distance float64, params *Params) geom.CoordinateSequence {
	// Calculate angles from vertex to each offset point
	angle0 := math.Atan2(p0.Y-vertex.Y, p0.X-vertex.X)
	angle1 := math.Atan2(p1.Y-vertex.Y, p1.X-vertex.X)

	// Normalize angle difference to go the short way
	angleDiff := angle1 - angle0
	for angleDiff > math.Pi {
		angleDiff -= 2 * math.Pi
	}
	for angleDiff < -math.Pi {
		angleDiff += 2 * math.Pi
	}

	// Number of segments based on angle
	absAngle := math.Abs(angleDiff)
	numSegments := int(math.Ceil(absAngle / (math.Pi / 2) * float64(params.QuadrantSegments)))
	if numSegments < 1 {
		numSegments = 1
	}

	result := make(geom.CoordinateSequence, 0, numSegments)

	angleStep := angleDiff / float64(numSegments)
	for i := 1; i < numSegments; i++ {
		angle := angle0 + angleStep*float64(i)
		x := vertex.X + distance*math.Cos(angle)
		y := vertex.Y + distance*math.Sin(angle)
		result = append(result, geom.NewCoordinate(x, y))
	}
	result = append(result, p1)

	return result
}

// computeMitreJoin computes a mitred corner join.
func computeMitreJoin(p0, p1, vertex geom.Coordinate, distance float64, params *Params) geom.CoordinateSequence {
	// Vectors from vertex to offset points
	d0x := p0.X - vertex.X
	d0y := p0.Y - vertex.Y
	d1x := p1.X - vertex.X
	d1y := p1.Y - vertex.Y

	len0 := math.Sqrt(d0x*d0x + d0y*d0y)
	len1 := math.Sqrt(d1x*d1x + d1y*d1y)

	if len0 < geom.DefaultEpsilon || len1 < geom.DefaultEpsilon {
		return geom.CoordinateSequence{p1}
	}

	// Normalize
	d0x /= len0
	d0y /= len0
	d1x /= len1
	d1y /= len1

	// Compute bisector direction
	bx := d0x + d1x
	by := d0y + d1y
	bLen := math.Sqrt(bx*bx + by*by)

	if bLen < geom.DefaultEpsilon {
		// Opposite directions, use bevel
		return geom.CoordinateSequence{p1}
	}

	bx /= bLen
	by /= bLen

	// Calculate angle between the two directions
	dot := d0x*d1x + d0y*d1y
	if dot < -1 {
		dot = -1
	}
	if dot > 1 {
		dot = 1
	}
	halfAngle := math.Acos(dot) / 2

	if math.Abs(math.Sin(halfAngle)) < geom.DefaultEpsilon {
		return geom.CoordinateSequence{p1}
	}

	// Mitre length
	mitreLength := distance / math.Sin(halfAngle)

	// Check mitre limit
	if mitreLength > params.MitreLimit*distance {
		// Exceed limit, use bevel
		return geom.CoordinateSequence{p1}
	}

	mitrePoint := geom.NewCoordinate(vertex.X+bx*mitreLength, vertex.Y+by*mitreLength)
	return geom.CoordinateSequence{mitrePoint, p1}
}

// computeCapPoints computes points for an end cap based on the cap style.
func computeCapPoints(endpoint geom.Coordinate, dx, dy, distance float64, params *Params) geom.CoordinateSequence {
	switch params.EndCapStyle {
	case CapFlat:
		// Flat cap - no additional points needed
		return geom.CoordinateSequence{}
	case CapSquare:
		// Square cap - extend by distance in the direction of the line
		px := -dy // perpendicular
		py := dx
		// Create two points extending past the endpoint
		p1 := geom.NewCoordinate(endpoint.X+dx*distance+px*distance, endpoint.Y+dy*distance+py*distance)
		p2 := geom.NewCoordinate(endpoint.X+dx*distance-px*distance, endpoint.Y+dy*distance-py*distance)
		return geom.CoordinateSequence{p1, p2}
	default: // CapRound
		return computeRoundCapPoints(endpoint, dx, dy, distance, params)
	}
}

// computeRoundCapPoints computes points for a round end cap.
func computeRoundCapPoints(endpoint geom.Coordinate, dx, dy, distance float64, params *Params) geom.CoordinateSequence {
	baseAngle := math.Atan2(dy, dx)
	numSegments := params.QuadrantSegments * 2

	// Generate semicircle from perpendicular left to perpendicular right
	result := make(geom.CoordinateSequence, 0, numSegments-1)

	for i := 1; i < numSegments; i++ {
		// Go from -90 degrees to +90 degrees relative to direction
		t := float64(i) / float64(numSegments)
		angle := baseAngle + math.Pi/2 - math.Pi*t
		x := endpoint.X + distance*math.Cos(angle)
		y := endpoint.Y + distance*math.Sin(angle)
		result = append(result, geom.NewCoordinate(x, y))
	}

	return result
}

// bufferPolygon creates a buffer around a polygon.
func bufferPolygon(poly *geom.Polygon, distance float64, params *Params) geom.Geometry {
	if poly.IsEmpty() {
		return geom.NewPolygonEmpty()
	}

	// For negative distance, we're eroding
	if distance < 0 {
		return erodePolygon(poly, -distance, params)
	}

	// For positive distance, expand the exterior ring outward
	shellCoords := poly.ExteriorRing().Coordinates()

	// Determine if shell is CCW (standard) or CW
	isCCW := geom.SignedArea(shellCoords) > 0

	// For CCW ring, positive buffer means offset to the left (positive distance)
	// For CW ring, we'd need to reverse
	offsetDist := distance
	if !isCCW {
		offsetDist = -distance
	}

	bufferedShell := offsetClosedRing(shellCoords, offsetDist, params)

	if len(bufferedShell) < 4 {
		return geom.NewPolygonEmpty()
	}

	// Ensure proper orientation (CCW for exterior)
	if geom.SignedArea(bufferedShell) < 0 {
		reverseCoords(bufferedShell)
	}

	// Buffer holes (shrink them - offset inward)
	var bufferedHoles []*geom.LinearRing
	for i := 0; i < poly.NumInteriorRings(); i++ {
		hole := poly.InteriorRingN(i)
		holeCoords := hole.Coordinates()

		// Holes are typically CW, shrinking means offsetting to their left (which is inward)
		isHoleCCW := geom.SignedArea(holeCoords) > 0
		holeOffset := -distance
		if isHoleCCW {
			holeOffset = distance
		}

		shrunkHole := offsetClosedRing(holeCoords, holeOffset, params)
		if len(shrunkHole) >= 4 {
			// Ensure CW orientation for holes
			if geom.SignedArea(shrunkHole) > 0 {
				reverseCoords(shrunkHole)
			}
			bufferedHoles = append(bufferedHoles, geom.NewLinearRing(shrunkHole))
		}
	}

	shell := geom.NewLinearRing(bufferedShell)
	return geom.NewPolygon(shell, bufferedHoles)
}

// erodePolygon shrinks a polygon by the given distance.
func erodePolygon(poly *geom.Polygon, distance float64, params *Params) geom.Geometry {
	if poly.IsEmpty() {
		return geom.NewPolygonEmpty()
	}

	// Shrink the exterior ring (offset inward)
	shellCoords := poly.ExteriorRing().Coordinates()

	// For CCW ring, shrinking means offset to the right (negative distance)
	isCCW := geom.SignedArea(shellCoords) > 0
	offsetDist := -distance
	if !isCCW {
		offsetDist = distance
	}

	shrunkShell := offsetClosedRing(shellCoords, offsetDist, params)

	if len(shrunkShell) < 4 {
		return geom.NewPolygonEmpty()
	}

	// Ensure CCW orientation
	if geom.SignedArea(shrunkShell) < 0 {
		reverseCoords(shrunkShell)
	}

	// Expand holes (offset outward)
	var expandedHoles []*geom.LinearRing
	for i := 0; i < poly.NumInteriorRings(); i++ {
		hole := poly.InteriorRingN(i)
		holeCoords := hole.Coordinates()

		isHoleCCW := geom.SignedArea(holeCoords) > 0
		holeOffset := distance
		if isHoleCCW {
			holeOffset = -distance
		}

		expandedHole := offsetClosedRing(holeCoords, holeOffset, params)
		if len(expandedHole) >= 4 {
			// Ensure CW orientation for holes
			if geom.SignedArea(expandedHole) > 0 {
				reverseCoords(expandedHole)
			}
			expandedHoles = append(expandedHoles, geom.NewLinearRing(expandedHole))
		}
	}

	shell := geom.NewLinearRing(shrunkShell)
	return geom.NewPolygon(shell, expandedHoles)
}

// offsetClosedRing computes an offset curve for a closed ring.
func offsetClosedRing(coords geom.CoordinateSequence, distance float64, params *Params) geom.CoordinateSequence {
	n := len(coords)
	if n < 4 { // Minimum valid ring: 3 points + closing point
		return geom.CoordinateSequence{}
	}

	// Work with the open ring (exclude closing point)
	openRing := coords[:n-1]
	numPts := len(openRing)

	var result geom.CoordinateSequence

	for i := 0; i < numPts; i++ {
		prevIdx := (i - 1 + numPts) % numPts
		nextIdx := (i + 1) % numPts

		prev := openRing[prevIdx]
		curr := openRing[i]
		next := openRing[nextIdx]

		// Compute offset point at this vertex
		offsetPts := computeVertexOffset(prev, curr, next, distance, params)
		result = append(result, offsetPts...)
	}

	if len(result) < 3 {
		return geom.CoordinateSequence{}
	}

	// Close the ring
	result = append(result, result[0].Clone())

	return result
}

// computeVertexOffset computes offset points at a vertex of a closed ring.
// For CCW rings (exterior), positive distance expands outward.
// For CW rings (holes), positive distance shrinks them.
func computeVertexOffset(prev, curr, next geom.Coordinate, distance float64, params *Params) geom.CoordinateSequence {
	// Direction vectors
	d1x := curr.X - prev.X
	d1y := curr.Y - prev.Y
	len1 := math.Sqrt(d1x*d1x + d1y*d1y)

	d2x := next.X - curr.X
	d2y := next.Y - curr.Y
	len2 := math.Sqrt(d2x*d2x + d2y*d2y)

	if len1 < geom.DefaultEpsilon || len2 < geom.DefaultEpsilon {
		// Degenerate segment
		return geom.CoordinateSequence{curr}
	}

	// Normalize
	d1x /= len1
	d1y /= len1
	d2x /= len2
	d2y /= len2

	// Perpendiculars (pointing right - exterior for CCW, interior for CW)
	// Right perpendicular of (dx, dy) is (dy, -dx)
	n1x := d1y
	n1y := -d1x
	n2x := d2y
	n2y := -d2x

	// Offset points on each segment
	o1 := geom.NewCoordinate(curr.X+n1x*distance, curr.Y+n1y*distance)
	o2 := geom.NewCoordinate(curr.X+n2x*distance, curr.Y+n2y*distance)

	// Check if we need corner treatment
	if o1.Distance(o2) < geom.DefaultEpsilon {
		return geom.CoordinateSequence{o1}
	}

	// Cross product to determine turn direction
	cross := d1x*d2y - d1y*d2x

	// With right perpendicular:
	// For CCW ring with positive buffer (expanding):
	//   - Turning left (cross > 0) at convex corner needs fillet
	//   - Turning right (cross < 0) at concave corner needs intersection
	// The sign relationship is inverted from left perpendicular
	isConvex := (cross < 0 && distance > 0) || (cross > 0 && distance < 0)

	if !isConvex {
		// Concave: compute intersection of the two offset lines
		// Line 1: o1 + t * d1
		// Line 2: o2 - s * d2
		denom := d1x*(-d2y) - d1y*(-d2x)
		if math.Abs(denom) < geom.DefaultEpsilon {
			return geom.CoordinateSequence{o1}
		}

		dx := o2.X - o1.X
		dy := o2.Y - o1.Y
		t := (dx*(-d2y) - dy*(-d2x)) / denom

		intersection := geom.NewCoordinate(o1.X+t*d1x, o1.Y+t*d1y)
		return geom.CoordinateSequence{intersection}
	}

	// Convex corner: add fillet
	switch params.JoinStyle {
	case JoinRound:
		return computeRoundJoin(o1, o2, curr, math.Abs(distance), params)
	case JoinMitre:
		result := computeMitreJoin(o1, o2, curr, math.Abs(distance), params)
		// Prepend o1
		return append(geom.CoordinateSequence{o1}, result...)
	case JoinBevel:
		return geom.CoordinateSequence{o1, o2}
	default:
		return geom.CoordinateSequence{o1, o2}
	}
}

// bufferMultiPoint buffers each point and unions them.
func bufferMultiPoint(mp *geom.MultiPoint, distance float64, params *Params) geom.Geometry {
	var polygons []*geom.Polygon
	for i := 0; i < mp.NumGeometries(); i++ {
		p := mp.GeometryN(i).(*geom.Point)
		buffered := bufferPoint(p, distance, params)
		if poly, ok := buffered.(*geom.Polygon); ok && !poly.IsEmpty() {
			polygons = append(polygons, poly)
		}
	}
	if len(polygons) == 0 {
		return geom.NewPolygonEmpty()
	}
	if len(polygons) == 1 {
		return polygons[0]
	}
	return geom.NewMultiPolygon(polygons)
}

// bufferMultiLineString buffers each line string and unions them.
func bufferMultiLineString(mls *geom.MultiLineString, distance float64, params *Params) geom.Geometry {
	var polygons []*geom.Polygon
	for i := 0; i < mls.NumGeometries(); i++ {
		ls := mls.GeometryN(i).(*geom.LineString)
		buffered := bufferLineString(ls, distance, params)
		if poly, ok := buffered.(*geom.Polygon); ok && !poly.IsEmpty() {
			polygons = append(polygons, poly)
		}
	}
	if len(polygons) == 0 {
		return geom.NewPolygonEmpty()
	}
	if len(polygons) == 1 {
		return polygons[0]
	}
	return geom.NewMultiPolygon(polygons)
}

// bufferMultiPolygon buffers each polygon and unions them.
func bufferMultiPolygon(mp *geom.MultiPolygon, distance float64, params *Params) geom.Geometry {
	var polygons []*geom.Polygon
	for i := 0; i < mp.NumGeometries(); i++ {
		poly := mp.GeometryN(i).(*geom.Polygon)
		buffered := bufferPolygon(poly, distance, params)
		switch v := buffered.(type) {
		case *geom.Polygon:
			if !v.IsEmpty() {
				polygons = append(polygons, v)
			}
		case *geom.MultiPolygon:
			for j := 0; j < v.NumGeometries(); j++ {
				p := v.GeometryN(j).(*geom.Polygon)
				if !p.IsEmpty() {
					polygons = append(polygons, p)
				}
			}
		}
	}
	if len(polygons) == 0 {
		return geom.NewPolygonEmpty()
	}
	if len(polygons) == 1 {
		return polygons[0]
	}
	return geom.NewMultiPolygon(polygons)
}

// bufferGeometryCollection buffers each geometry in the collection.
func bufferGeometryCollection(gc *geom.GeometryCollection, distance float64, params *Params) geom.Geometry {
	var results []geom.Geometry
	for i := 0; i < gc.NumGeometries(); i++ {
		buffered := BufferWithParams(gc.GeometryN(i), distance, params)
		if !buffered.IsEmpty() {
			results = append(results, buffered)
		}
	}
	if len(results) == 0 {
		return geom.NewGeometryCollectionEmpty()
	}
	if len(results) == 1 {
		return results[0]
	}
	return geom.NewGeometryCollection(results)
}

// reverseCoords reverses a coordinate sequence in place.
func reverseCoords(coords geom.CoordinateSequence) {
	for i, j := 0, len(coords)-1; i < j; i, j = i+1, j-1 {
		coords[i], coords[j] = coords[j], coords[i]
	}
}
