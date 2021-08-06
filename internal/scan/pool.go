package scan

import (
	"context"
	"sync"
)

// ProbeFunc probes a single target and returns its result.
type ProbeFunc func(ctx context.Context, t Target) Result

// Run probes every target using a bounded worker pool. Results are returned in
// the same order as the input targets regardless of completion order.
func Run(ctx context.Context, targets []Target, concurrency int, probe ProbeFunc) []Result {
	results := make([]Result, len(targets))
	if len(targets) == 0 {
		return results
	}
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > len(targets) {
		concurrency = len(targets)
	}

	type job struct {
		idx    int
		target Target
	}
	jobs := make(chan job)

	var wg sync.WaitGroup
	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				results[j.idx] = probe(ctx, j.target)
			}
		}()
	}

	for i, t := range targets {
		jobs <- job{idx: i, target: t}
	}
	close(jobs)
	wg.Wait()

	return results
}
