package scan

import "testing"

func TestAggregate(t *testing.T) {
	results := []Result{
		{Target: Target{"a", 1}, Open: true},
		{Target: Target{"b", 2}, Open: false},
		{Target: Target{"c", 3}, Open: true},
	}
	s := Aggregate(results)
	if s.Total != 3 || s.Open != 2 || s.Closed != 1 {
		t.Fatalf("summary = %+v", s)
	}
	if s.AllOpen() {
		t.Fatal("AllOpen() should be false with a closed target")
	}
}

func TestAggregateAllOpen(t *testing.T) {
	s := Aggregate([]Result{{Open: true}, {Open: true}})
	if !s.AllOpen() {
		t.Fatal("AllOpen() should be true")
	}
}

func TestAggregateEmptyNotAllOpen(t *testing.T) {
	if Aggregate(nil).AllOpen() {
		t.Fatal("empty summary must not report AllOpen")
	}
}
