package probe

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/moveeeax/portping/internal/scan"
)

// fakeDialer returns a canned error, counting attempts. Used to exercise the
// retry and timeout paths without touching the network.
type fakeDialer struct {
	err      error
	attempts int32
	delay    time.Duration
}

func (f *fakeDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	atomic.AddInt32(&f.attempts, 1)
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, f.err
}

// flakyDialer fails its first `failures` calls, then dials for real. Used to
// exercise the path where a retry eventually succeeds, distinct from
// fakeDialer which always fails.
type flakyDialer struct {
	failures int
	calls    int32
}

func (f *flakyDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	n := atomic.AddInt32(&f.calls, 1)
	if int(n) <= f.failures {
		return nil, errors.New("simulated transient failure")
	}
	var d net.Dialer
	return d.DialContext(ctx, network, address)
}

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

func TestProbeInjectedTimeout(t *testing.T) {
	fake := &fakeDialer{err: context.DeadlineExceeded, delay: 50 * time.Millisecond}
	p := &Prober{Dialer: fake, Timeout: 10 * time.Millisecond, Count: 1}
	res := p.Probe(context.Background(), scan.Target{Host: "10.255.255.1", Port: 80})
	if res.Open {
		t.Fatal("expected timeout to be reported as closed")
	}
	if res.Attempts != 1 {
		t.Errorf("attempts = %d, want 1", res.Attempts)
	}
}

func TestProbeRetries(t *testing.T) {
	fake := &fakeDialer{err: errors.New("connection refused")}
	p := &Prober{Dialer: fake, Timeout: 10 * time.Millisecond, Count: 3}
	res := p.Probe(context.Background(), scan.Target{Host: "127.0.0.1", Port: 9})
	if res.Open {
		t.Fatal("expected closed")
	}
	if res.Attempts != 3 {
		t.Errorf("attempts = %d, want 3", res.Attempts)
	}
	if got := atomic.LoadInt32(&fake.attempts); got != 3 {
		t.Errorf("dialer called %d times, want 3", got)
	}
}

func TestProbeRetriesThenSucceeds(t *testing.T) {
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

	fake := &flakyDialer{failures: 2}
	p := &Prober{Dialer: fake, Timeout: time.Second, Count: 5}
	res := p.Probe(context.Background(), targetFromAddr(t, ln.Addr().String()))
	if !res.Open {
		t.Fatalf("expected the retry to eventually succeed, got closed: %s", res.Err)
	}
	if res.Attempts != 3 {
		t.Errorf("attempts = %d, want 3 (2 failures then a success)", res.Attempts)
	}
	if got := atomic.LoadInt32(&fake.calls); got != 3 {
		t.Errorf("dialer called %d times, want 3", got)
	}
	if res.Latency <= 0 {
		t.Errorf("latency = %v, want > 0", res.Latency)
	}
}

func TestProbeCountZeroDefaultsToOne(t *testing.T) {
	fake := &fakeDialer{err: errors.New("nope")}
	p := &Prober{Dialer: fake, Timeout: 10 * time.Millisecond, Count: 0}
	res := p.Probe(context.Background(), scan.Target{Host: "127.0.0.1", Port: 9})
	if res.Attempts != 1 {
		t.Errorf("attempts = %d, want 1", res.Attempts)
	}
}

func TestProbeCanceledContextStops(t *testing.T) {
	fake := &fakeDialer{err: errors.New("refused")}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := &Prober{Dialer: fake, Timeout: 10 * time.Millisecond, Count: 5}
	res := p.Probe(ctx, scan.Target{Host: "127.0.0.1", Port: 9})
	if res.Open {
		t.Fatal("expected closed")
	}
	if got := atomic.LoadInt32(&fake.attempts); got != 1 {
		t.Errorf("dialer called %d times, want 1 (should stop on canceled ctx)", got)
	}
	// Attempts must reflect the dials actually made, not the configured Count.
	if res.Attempts != 1 {
		t.Errorf("attempts = %d, want 1 (only one dial was made before the ctx cut the retries short)", res.Attempts)
	}
}
