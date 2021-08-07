package scan

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestRunPreservesOrder(t *testing.T) {
	targets := make([]Target, 50)
	for i := range targets {
		targets[i] = Target{Host: "10.0.0.1", Port: i + 1}
	}
	probe := func(ctx context.Context, tg Target) Result {
		return Result{Target: tg, Open: true, Attempts: 1}
	}
	results := Run(context.Background(), targets, 8, probe)
	if len(results) != len(targets) {
		t.Fatalf("got %d results, want %d", len(results), len(targets))
	}
	for i, r := range results {
		if r.Target != targets[i] {
			t.Fatalf("result[%d] = %v, want %v (order not preserved)", i, r.Target, targets[i])
		}
	}
}

func TestRunConcurrencyBounded(t *testing.T) {
	targets := make([]Target, 100)
	for i := range targets {
		targets[i] = Target{Host: "h", Port: i + 1}
	}
	var mu sync.Mutex
	var inFlight, maxInFlight int
	probe := func(ctx context.Context, tg Target) Result {
		mu.Lock()
		inFlight++
		if inFlight > maxInFlight {
			maxInFlight = inFlight
		}
		mu.Unlock()
		time.Sleep(time.Millisecond)
		mu.Lock()
		inFlight--
		mu.Unlock()
		return Result{Target: tg, Open: true}
	}
	Run(context.Background(), targets, 5, probe)
	if maxInFlight > 5 {
		t.Fatalf("max concurrency = %d, want <= 5", maxInFlight)
	}
	if maxInFlight == 0 {
		t.Fatal("probe never ran")
	}
}

func TestRunEmpty(t *testing.T) {
	got := Run(context.Background(), nil, 4, func(context.Context, Target) Result {
		t.Fatal("probe should not be called")
		return Result{}
	})
	if len(got) != 0 {
		t.Fatalf("got %d results, want 0", len(got))
	}
}

func TestRunZeroConcurrencyDefaultsToOne(t *testing.T) {
	targets := []Target{{"a", 1}, {"b", 2}}
	var calls int
	var mu sync.Mutex
	probe := func(ctx context.Context, tg Target) Result {
		mu.Lock()
		calls++
		mu.Unlock()
		return Result{Target: tg}
	}
	got := Run(context.Background(), targets, 0, probe)
	if calls != 2 || len(got) != 2 {
		t.Fatalf("calls=%d results=%d", calls, len(got))
	}
}

func TestRunConcurrencyCappedToTargets(t *testing.T) {
	targets := []Target{{"a", 1}}
	var ports []int
	probe := func(ctx context.Context, tg Target) Result {
		ports = append(ports, tg.Port)
		return Result{Target: tg}
	}
	got := Run(context.Background(), targets, 100, probe)
	if len(got) != 1 {
		t.Fatalf("got %d results, want 1", len(got))
	}
	sort.Ints(ports)
	if len(ports) != 1 || ports[0] != 1 {
		t.Fatalf("ports = %v", ports)
	}
}
