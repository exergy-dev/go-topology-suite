// Package jtstest provides a JTS XML conformance test harness for terra.
//
// The harness reads JTS-format XML test cases (as used by the
// locationtech/jts project for its testxml suite) and dispatches each
// op to the corresponding terra function. It exists purely to validate
// behavior parity with JTS and is not part of the terra public API.
//
// Build tag
//
// The package is gated behind the "jts" build tag so that the default
// `go test ./...` invocation never compiles or runs it. To execute the
// harness use:
//
//	go test -tags=jts ./internal/jtstest/
//
// Sample corpus
//
// A small hand-crafted corpus lives in testdata/. It is not a full
// mirror of the upstream JTS suite — that would be brought in as a
// vendored or fetched-on-demand dataset in a follow-up. The shipped
// cases exist to prove the runner works end-to-end.
package jtstest
