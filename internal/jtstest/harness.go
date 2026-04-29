//go:build jts

package jtstest

import (
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/terra-geo/terra/geom"
	"github.com/terra-geo/terra/measure"
	"github.com/terra-geo/terra/overlay"
	"github.com/terra-geo/terra/predicate"
	"github.com/terra-geo/terra/wkt"
)

// run is the top-level <run> element in a JTS XML test file.
type run struct {
	XMLName xml.Name  `xml:"run"`
	Desc    string    `xml:"desc"`
	Cases   []xmlCase `xml:"case"`
}

// xmlCase is one <case>: shared operands and one or more tests.
type xmlCase struct {
	Desc  string    `xml:"desc"`
	A     string    `xml:"a"`
	B     string    `xml:"b"`
	Tests []xmlTest `xml:"test"`
}

type xmlTest struct {
	Desc string `xml:"desc"`
	Op   xmlOp  `xml:"op"`
}

// xmlOp is the <op name="..." arg1="A" arg2="B">expected</op> element.
// The expected result is the element text. arg3 carries an optional extra
// argument (e.g. a DE-9IM pattern for relate, a buffer distance, etc.).
type xmlOp struct {
	Name     string `xml:"name,attr"`
	Arg1     string `xml:"arg1,attr"`
	Arg2     string `xml:"arg2,attr"`
	Arg3     string `xml:"arg3,attr"`
	Expected string `xml:",chardata"`
}

// loadFile parses a single JTS XML test file.
func loadFile(path string) (*run, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return decodeRun(f)
}

func decodeRun(r io.Reader) (*run, error) {
	var rn run
	dec := xml.NewDecoder(r)
	if err := dec.Decode(&rn); err != nil {
		return nil, err
	}
	return &rn, nil
}

// findCorpus walks dir and returns every *.xml file found.
func findCorpus(dir string) ([]string, error) {
	var out []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".xml") {
			out = append(out, path)
		}
		return nil
	})
	return out, err
}

// resolveOperand returns the WKT string referenced by an arg attribute.
// JTS uses the literal tokens "A" or "B" to reference the case-level
// operands; otherwise the attribute value is itself the WKT.
func resolveOperand(c *xmlCase, arg string) string {
	switch strings.TrimSpace(strings.ToUpper(arg)) {
	case "A":
		return c.A
	case "B":
		return c.B
	default:
		return arg
	}
}

// parseWKT decodes WKT, treating empty/whitespace input as an error so the
// caller can decide how to surface the gap.
func parseWKT(s string) (geom.Geometry, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty WKT")
	}
	return wkt.Unmarshal(s)
}

// dispatchResult is the outcome of running a single op.
type dispatchResult struct {
	Skipped bool   // op not implemented by the harness
	Reason  string // reason for skip
	Pass    bool
	Detail  string // failure detail when Pass == false
}

// runOp dispatches an op to the appropriate terra function and compares
// the result against the expected XML payload.
func runOp(c *xmlCase, op xmlOp) dispatchResult {
	name := strings.ToLower(strings.TrimSpace(op.Name))

	switch name {
	case "intersection", "union", "difference", "symdifference":
		return runOverlayOp(c, op, name)
	case "relate":
		return runRelate(c, op)
	case "intersects", "disjoint", "contains", "within",
		"covers", "coveredby", "touches", "crosses",
		"overlaps", "equals", "equalstopo":
		return runPredicate(c, op, name)
	case "getarea":
		return runScalar(c, op, measure.Area)
	case "getlength":
		return runScalar(c, op, measure.Length)
	default:
		return dispatchResult{Skipped: true, Reason: "unsupported op: " + op.Name}
	}
}

func runOverlayOp(c *xmlCase, op xmlOp, name string) dispatchResult {
	a, err := parseWKT(resolveOperand(c, op.Arg1))
	if err != nil {
		return dispatchResult{Detail: "parse arg1: " + err.Error()}
	}
	b, err := parseWKT(resolveOperand(c, op.Arg2))
	if err != nil {
		return dispatchResult{Detail: "parse arg2: " + err.Error()}
	}

	var got geom.Geometry
	switch name {
	case "intersection":
		got, err = overlay.Intersection(a, b)
	case "union":
		got, err = overlay.Union(a, b)
	case "difference":
		got, err = overlay.Difference(a, b)
	case "symdifference":
		got, err = overlay.SymmetricDifference(a, b)
	}
	if err != nil {
		return dispatchResult{Detail: name + ": " + err.Error()}
	}

	expected, err := parseWKT(op.Expected)
	if err != nil {
		// Fall back to comparing trimmed string equality for cases
		// where the expected payload is non-WKT (rare in the JTS suite
		// for these ops, but we want a useful message either way).
		return dispatchResult{Detail: "parse expected: " + err.Error()}
	}

	eq, eerr := predicate.Equals(got, expected)
	if eerr != nil {
		return dispatchResult{Detail: "equals: " + eerr.Error()}
	}
	if !eq {
		return dispatchResult{
			Detail: fmt.Sprintf("expected %s, got %s",
				op.Expected, geomString(got)),
		}
	}
	return dispatchResult{Pass: true}
}

func runRelate(c *xmlCase, op xmlOp) dispatchResult {
	a, err := parseWKT(resolveOperand(c, op.Arg1))
	if err != nil {
		return dispatchResult{Detail: "parse arg1: " + err.Error()}
	}
	b, err := parseWKT(resolveOperand(c, op.Arg2))
	if err != nil {
		return dispatchResult{Detail: "parse arg2: " + err.Error()}
	}
	im, err := predicate.Relate(a, b)
	if err != nil {
		return dispatchResult{Detail: "relate: " + err.Error()}
	}

	// JTS uses arg3 for a DE-9IM pattern when the expected payload is a
	// boolean, and the element text for the matrix string itself. We
	// support both.
	pattern := strings.TrimSpace(op.Arg3)
	expected := strings.TrimSpace(op.Expected)

	if pattern != "" {
		want, err := parseBool(expected)
		if err != nil {
			return dispatchResult{Detail: "parse expected bool: " + err.Error()}
		}
		got := im.Matches(pattern)
		if got != want {
			return dispatchResult{
				Detail: fmt.Sprintf("relate %s against %q: want %v got %v (im=%s)",
					pattern, pattern, want, got, im),
			}
		}
		return dispatchResult{Pass: true}
	}

	if len(expected) != 9 {
		return dispatchResult{Detail: "expected 9-char DE-9IM, got " + expected}
	}
	if !im.Matches(expected) {
		return dispatchResult{
			Detail: fmt.Sprintf("relate: want %s got %s", expected, im),
		}
	}
	return dispatchResult{Pass: true}
}

func runPredicate(c *xmlCase, op xmlOp, name string) dispatchResult {
	a, err := parseWKT(resolveOperand(c, op.Arg1))
	if err != nil {
		return dispatchResult{Detail: "parse arg1: " + err.Error()}
	}
	b, err := parseWKT(resolveOperand(c, op.Arg2))
	if err != nil {
		return dispatchResult{Detail: "parse arg2: " + err.Error()}
	}

	var got bool
	switch name {
	case "intersects":
		got, err = predicate.Intersects(a, b)
	case "disjoint":
		got, err = predicate.Disjoint(a, b)
	case "contains":
		got, err = predicate.Contains(a, b)
	case "within":
		got, err = predicate.Within(a, b)
	case "covers":
		got, err = predicate.Covers(a, b)
	case "coveredby":
		got, err = predicate.CoveredBy(a, b)
	case "touches":
		got, err = predicate.Touches(a, b)
	case "crosses":
		got, err = predicate.Crosses(a, b)
	case "overlaps":
		got, err = predicate.Overlaps(a, b)
	case "equals", "equalstopo":
		got, err = predicate.Equals(a, b)
	}
	if err != nil {
		return dispatchResult{Detail: name + ": " + err.Error()}
	}
	want, perr := parseBool(op.Expected)
	if perr != nil {
		return dispatchResult{Detail: "parse expected bool: " + perr.Error()}
	}
	if got != want {
		return dispatchResult{
			Detail: fmt.Sprintf("%s: want %v got %v", name, want, got),
		}
	}
	return dispatchResult{Pass: true}
}

func runScalar(c *xmlCase, op xmlOp, fn func(geom.Geometry, ...measure.Option) float64) dispatchResult {
	a, err := parseWKT(resolveOperand(c, op.Arg1))
	if err != nil {
		return dispatchResult{Detail: "parse arg1: " + err.Error()}
	}
	got := fn(a)
	want, perr := strconv.ParseFloat(strings.TrimSpace(op.Expected), 64)
	if perr != nil {
		return dispatchResult{Detail: "parse expected float: " + perr.Error()}
	}
	if !nearlyEqual(got, want, 1e-9) {
		return dispatchResult{
			Detail: fmt.Sprintf("scalar: want %g got %g", want, got),
		}
	}
	return dispatchResult{Pass: true}
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "t", "1":
		return true, nil
	case "false", "f", "0":
		return false, nil
	default:
		return false, fmt.Errorf("not a bool: %q", s)
	}
}

func nearlyEqual(a, b, tol float64) bool {
	if math.IsNaN(a) || math.IsNaN(b) {
		return false
	}
	d := math.Abs(a - b)
	if d <= tol {
		return true
	}
	scale := math.Max(math.Abs(a), math.Abs(b))
	return d <= tol*scale
}

// geomString returns a best-effort textual form of g for failure messages.
// It is intentionally lossy and is *not* a stable serialization.
func geomString(g geom.Geometry) string {
	if g == nil {
		return "<nil>"
	}
	if g.IsEmpty() {
		return strings.ToUpper(g.Type().String()) + " EMPTY"
	}
	return fmt.Sprintf("%s(env=%v)", strings.ToUpper(g.Type().String()), g.Envelope())
}
