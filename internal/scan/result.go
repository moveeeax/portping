package scan

import "time"

// Result captures the outcome of probing a single target.
type Result struct {
	Target   Target
	Open     bool
	Latency  time.Duration
	Attempts int
	Err      string
}

// Match reports whether the observed state equals the expected one.
func (r Result) Match() bool {
	return r.Open == r.Target.Expect.WantOpen()
}

// Summary aggregates a set of results.
type Summary struct {
	Total  int
	Open   int
	Closed int
	// Mismatched counts results whose state differs from the expected one.
	Mismatched int
	// MismatchedExplicit is the subset of Mismatched whose expectation was
	// given explicitly with "=open" or "=closed".
	MismatchedExplicit int
	Results            []Result
}

// Aggregate tallies open, closed, and mismatched results. The input order is
// preserved.
func Aggregate(results []Result) Summary {
	s := Summary{Results: results, Total: len(results)}
	for _, r := range results {
		if r.Open {
			s.Open++
		} else {
			s.Closed++
		}
		if !r.Match() {
			s.Mismatched++
			if r.Target.Expect.Explicit() {
				s.MismatchedExplicit++
			}
		}
	}
	return s
}

// AllOpen reports whether every target was reachable.
func (s Summary) AllOpen() bool {
	return s.Total > 0 && s.Closed == 0
}
