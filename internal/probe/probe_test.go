package probe

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/moveeeax/portping/internal/scan"
)

func targetFromAddr(t *testing.T, addr string) scan.Target {
	t.Helper()
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("split %q: %v", addr, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("atoi %q: %v", portStr, err)
	}
	return scan.Target{Host: host, Port: port}
}

func TestProbeOpenPort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	p := New(time.Second, 1)
	res := p.Probe(context.Background(), targetFromAddr(t, ln.Addr().String()))
	if !res.Open {
		t.Fatalf("expected open, got closed: %s", res.Err)
	}
	if res.Attempts != 1 {
		t.Errorf("attempts = %d, want 1", res.Attempts)
	}
	if res.Latency <= 0 {
		t.Errorf("latency = %v, want > 0", res.Latency)
	}
}

func TestProbeClosedPort(t *testing.T) {
	// Bind then close to obtain a port that is very likely unused.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()

	p := New(500*time.Millisecond, 1)
	res := p.Probe(context.Background(), targetFromAddr(t, addr))
	if res.Open {
		t.Fatal("expected closed port to be unreachable")
	}
	if res.Err == "" {
		t.Error("expected error message on closed port")
	}
}
