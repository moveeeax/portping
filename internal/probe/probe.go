package probe

import (
	"context"
	"net"
	"time"

	"github.com/moveeeax/portping/internal/scan"
)

// Dialer is the subset of net.Dialer used by the prober. It is satisfied by
// *net.Dialer and can be replaced with a fake in tests.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// Prober probes targets over TCP using an injectable dialer.
type Prober struct {
	Dialer  Dialer
	Timeout time.Duration
	// Count is the number of dial attempts before a target is declared
	// unreachable. A target is reported open as soon as one attempt succeeds.
	Count int
}

// New returns a Prober backed by the standard net.Dialer.
func New(timeout time.Duration, count int) *Prober {
	return &Prober{
		Dialer:  &net.Dialer{},
		Timeout: timeout,
		Count:   count,
	}
}

// Probe dials the target, retrying up to Count times, and returns the result
// including the connect latency of the successful attempt.
func (p *Prober) Probe(ctx context.Context, t scan.Target) scan.Result {
	attempts := p.Count
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	made := 0
	for i := 0; i < attempts; i++ {
		made = i + 1
		start := time.Now()
		dialCtx := ctx
		var cancel context.CancelFunc
		if p.Timeout > 0 {
			dialCtx, cancel = context.WithTimeout(ctx, p.Timeout)
		}
		conn, err := p.Dialer.DialContext(dialCtx, "tcp", t.Addr())
		if cancel != nil {
			cancel()
		}
		if err == nil {
			latency := time.Since(start)
			_ = conn.Close()
			return scan.Result{
				Target:   t,
				Open:     true,
				Latency:  latency,
				Attempts: i + 1,
			}
		}
		lastErr = err
		if ctx.Err() != nil {
			break
		}
	}

	// Report the attempts actually made, which is fewer than the configured
	// count when the context was canceled part-way through the retries.
	res := scan.Result{
		Target:   t,
		Open:     false,
		Attempts: made,
	}
	if lastErr != nil {
		res.Err = lastErr.Error()
	}
	return res
}
