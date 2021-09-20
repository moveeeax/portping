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
