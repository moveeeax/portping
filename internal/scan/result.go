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

// Summary aggregates a set of results.
type Summary struct {
	Total   int
	Open    int
	Closed  int
	Results []Result
}

// Aggregate tallies open and closed results. The input order is preserved.
func Aggregate(results []Result) Summary {
	s := Summary{Results: results, Total: len(results)}
	for _, r := range results {
		if r.Open {
			s.Open++
		} else {
			s.Closed++
		}
	}
	return s
}

// AllOpen reports whether every target was reachable.
func (s Summary) AllOpen() bool {
	return s.Total > 0 && s.Closed == 0
}
