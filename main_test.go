package main

import (
	"bytes"
	"encoding/json"
	"net"
	"strings"
	"testing"
)

func TestCollectSpecsFromArgs(t *testing.T) {
	got, err := collectSpecs([]string{"127.0.0.1:80", "10.0.0.0/30:22"}, "", strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
}

func TestCollectSpecsFromStdin(t *testing.T) {
	in := strings.NewReader("# comment\n127.0.0.1:80\n\n10.0.0.1:22 10.0.0.2:22\n")
	got, err := collectSpecs(nil, "-", in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"127.0.0.1:80", "10.0.0.1:22", "10.0.0.2:22"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestRunOpenTargetExitZero(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()

	var out, errBuf bytes.Buffer
	code := run([]string{ln.Addr().String()}, strings.NewReader(""), &out, &errBuf)
	if code != 0 {
		t.Fatalf("exit = %d, want 0; stderr=%s", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "open") {
		t.Fatalf("output missing open state: %s", out.String())
	}
}

func TestRunClosedTargetExitOne(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()

	var out, errBuf bytes.Buffer
	code := run([]string{"--timeout=300ms", addr}, strings.NewReader(""), &out, &errBuf)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
}

func TestRunClosedTargetNoFail(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()

	var out, errBuf bytes.Buffer
	code := run([]string{"--timeout=300ms", "--fail-on-unreachable=false", addr}, strings.NewReader(""), &out, &errBuf)
	if code != 0 {
		t.Fatalf("exit = %d, want 0 with fail disabled", code)
	}
}

func TestRunJSONOutput(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()

	var out, errBuf bytes.Buffer
	code := run([]string{"--json", ln.Addr().String()}, strings.NewReader(""), &out, &errBuf)
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	var parsed []jsonResult
	if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out.String())
	}
	if len(parsed) != 1 || !parsed[0].Open {
		t.Fatalf("unexpected json: %+v", parsed)
	}
}

func TestRunBadTargetExitTwo(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run([]string{"not-a-target"}, strings.NewReader(""), &out, &errBuf)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
}

func TestRunOpenOnlyFiltersClosed(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()
	openAddr := ln.Addr().String()

	closedLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	closedAddr := closedLn.Addr().String()
	closedLn.Close()

	var out, errBuf bytes.Buffer
	run([]string{"--timeout=300ms", "--open-only", "--fail-on-unreachable=false", openAddr, closedAddr}, strings.NewReader(""), &out, &errBuf)
	if strings.Contains(out.String(), "closed") {
		t.Fatalf("open-only output contained closed row: %s", out.String())
	}
}

func TestRunHelpExitsZero(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run([]string{"-h"}, strings.NewReader(""), &out, &errBuf)
	if code != 0 {
		t.Fatalf("exit = %d, want 0 for explicitly requested help", code)
	}
	if !strings.Contains(errBuf.String(), "usage: portping") {
		t.Fatalf("help output missing usage line: %s", errBuf.String())
	}
}

func TestRunOversizedCIDRExitsTwo(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run([]string{"10.0.0.0/8:80"}, strings.NewReader(""), &out, &errBuf)
	if code != 2 {
		t.Fatalf("exit = %d, want 2 for an oversized CIDR", code)
	}
	if !strings.Contains(errBuf.String(), "more than") {
		t.Fatalf("expected a size-limit error, got: %s", errBuf.String())
	}
}

func TestRunNonPositiveTimeoutExitsTwo(t *testing.T) {
	cases := []string{"0", "0s", "-1s", "-500ms"}
	for _, tc := range cases {
		var out, errBuf bytes.Buffer
		code := run([]string{"--timeout=" + tc, "127.0.0.1:80"}, strings.NewReader(""), &out, &errBuf)
		if code != 2 {
			t.Errorf("timeout=%s: exit = %d, want 2; stderr=%s", tc, code, errBuf.String())
		}
		if !strings.Contains(errBuf.String(), "--timeout must be greater than zero") {
			t.Errorf("timeout=%s: expected a timeout validation error, got: %s", tc, errBuf.String())
		}
		if out.Len() != 0 {
			t.Errorf("timeout=%s: expected no stdout on a validation error, got: %s", tc, out.String())
		}
	}
}

func TestRunSlashZeroExitsTwo(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run([]string{"0.0.0.0/0:80"}, strings.NewReader(""), &out, &errBuf)
	if code != 2 {
		t.Fatalf("exit = %d, want 2 for a /0", code)
	}
	if strings.Contains(errBuf.String(), "no targets given") {
		t.Fatalf("a /0 should be rejected on its own terms, not as an empty target list: %s", errBuf.String())
	}
}

// listenOpen starts a listener that accepts and closes connections and returns
// its address.
func listenOpen(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()
	return ln.Addr().String()
}

// closedAddr returns an address that nothing listens on.
func closedAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()
	return addr
}

func TestRunExpectClosedMatches(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run([]string{"--timeout=300ms", listenOpen(t) + "=open", closedAddr(t) + "=closed"}, strings.NewReader(""), &out, &errBuf)
	if code != 0 {
		t.Fatalf("exit = %d, want 0 when every target matches; stdout=%s stderr=%s", code, out.String(), errBuf.String())
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("want header and 2 rows, got:\n%s", out.String())
	}
	if h := strings.Fields(lines[0]); strings.Join(h, " ") != "TARGET STATE EXPECT MATCH LATENCY ATTEMPTS DETAIL" {
		t.Fatalf("unexpected header: %q", lines[0])
	}
	if f := strings.Fields(lines[2]); f[1] != "closed" || f[2] != "closed" || f[3] != "yes" {
		t.Fatalf("unexpected closed row: %q", lines[2])
	}
}

func TestRunExpectClosedButOpenExitOne(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run([]string{listenOpen(t) + "=closed"}, strings.NewReader(""), &out, &errBuf)
	if code != 1 {
		t.Fatalf("exit = %d, want 1 for an open port expected closed", code)
	}
	if f := strings.Fields(strings.Split(out.String(), "\n")[1]); f[1] != "open" || f[2] != "closed" || f[3] != "no" {
		t.Fatalf("unexpected row: %s", out.String())
	}
}

// --fail-on-unreachable=false waives only targets without a suffix; an
// explicit assertion still fails the run.
func TestRunExplicitMismatchIgnoresFailFlag(t *testing.T) {
	cases := [][]string{
		{listenOpen(t) + "=closed"},
		{closedAddr(t) + "=open"},
	}
	for _, targets := range cases {
		var out, errBuf bytes.Buffer
		args := append([]string{"--timeout=300ms", "--fail-on-unreachable=false"}, targets...)
		if code := run(args, strings.NewReader(""), &out, &errBuf); code != 1 {
			t.Errorf("%v: exit = %d, want 1", targets, code)
		}
	}
}

func TestRunExpectFromStdin(t *testing.T) {
	in := strings.NewReader("# firewall regression\n" + listenOpen(t) + "=open\n" + closedAddr(t) + "=closed\n")
	var out, errBuf bytes.Buffer
	code := run([]string{"--timeout=300ms", "--file", "-"}, in, &out, &errBuf)
	if code != 0 {
		t.Fatalf("exit = %d, want 0; stdout=%s stderr=%s", code, out.String(), errBuf.String())
	}
}

func TestRunJSONExpectFields(t *testing.T) {
	open := listenOpen(t)
	var out, errBuf bytes.Buffer
	run([]string{"--json", "--timeout=300ms", open, open + "=closed"}, strings.NewReader(""), &out, &errBuf)
	// The second spec overrides the default expectation for the same pair.
	var parsed []map[string]any
	if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out.String())
	}
	if len(parsed) != 1 {
		t.Fatalf("want 1 result, got %+v", parsed)
	}
	if parsed[0]["expected"] != "closed" || parsed[0]["ok"] != false {
		t.Fatalf("unexpected json: %+v", parsed[0])
	}

	out.Reset()
	run([]string{"--json", open}, strings.NewReader(""), &out, &errBuf)
	parsed = nil
	if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out.String())
	}
	if parsed[0]["expected"] != "open" || parsed[0]["ok"] != true {
		t.Fatalf("unexpected json for default expectation: %+v", parsed[0])
	}
}

func TestRunInvalidExpectExitTwo(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := run([]string{"127.0.0.1:80=filtered"}, strings.NewReader(""), &out, &errBuf); code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if code := run([]string{"127.0.0.1:80=open", "127.0.0.1:80=closed"}, strings.NewReader(""), &out, &errBuf); code != 2 {
		t.Fatalf("exit = %d, want 2 for contradictory expectations", code)
	}
}
