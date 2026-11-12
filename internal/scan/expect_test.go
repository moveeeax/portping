package scan

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseTargetExpectSuffix(t *testing.T) {
	cases := []struct {
		spec string
		want Expect
	}{
		{"10.0.0.5:22", ExpectDefault},
		{"10.0.0.5:22=open", ExpectOpen},
		{"10.0.0.5:3306=closed", ExpectClosed},
		{"[::1]:80=closed", ExpectClosed},
	}
	for _, c := range cases {
		got, err := ParseTarget(c.spec)
		if err != nil {
			t.Fatalf("ParseTarget(%q): unexpected error: %v", c.spec, err)
		}
		if len(got) != 1 || got[0].Expect != c.want {
			t.Errorf("ParseTarget(%q) = %+v, want Expect %v", c.spec, got, c.want)
		}
	}
}

func TestParseTargetExpectAppliesToExpansion(t *testing.T) {
	got, err := ParseTarget("10.0.0.0/30:22,80=closed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []Target{
		{Host: "10.0.0.1", Port: 22, Expect: ExpectClosed},
		{Host: "10.0.0.1", Port: 80, Expect: ExpectClosed},
		{Host: "10.0.0.2", Port: 22, Expect: ExpectClosed},
		{Host: "10.0.0.2", Port: 80, Expect: ExpectClosed},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParseTargetExpectErrors(t *testing.T) {
	cases := []string{"10.0.0.5:22=", "10.0.0.5:22=up", "10.0.0.5:22=Open", "=open", "10.0.0.5=closed", "10.0.0.5:22=open=closed"}
	for _, c := range cases {
		if _, err := ParseTarget(c); err == nil {
			t.Errorf("ParseTarget(%q) expected error, got nil", c)
		}
	}
}

func TestParseTargetsExplicitOverridesDefault(t *testing.T) {
	got, err := ParseTargets([]string{"10.0.0.0/30:22", "10.0.0.2:22=closed", "10.0.0.1:22=open"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []Target{
		{Host: "10.0.0.1", Port: 22, Expect: ExpectOpen},
		{Host: "10.0.0.2", Port: 22, Expect: ExpectClosed},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParseTargetsDefaultDoesNotOverrideExplicit(t *testing.T) {
	got, err := ParseTargets([]string{"10.0.0.2:22=closed", "10.0.0.0/30:22"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []Target{
		{Host: "10.0.0.2", Port: 22, Expect: ExpectClosed},
		{Host: "10.0.0.1", Port: 22, Expect: ExpectDefault},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParseTargetsConflictingExpectations(t *testing.T) {
	_, err := ParseTargets([]string{"10.0.0.5:22=open", "10.0.0.5:22=closed"})
	if err == nil {
		t.Fatal("expected an error for contradictory expectations, got nil")
	}
	if !strings.Contains(err.Error(), "both open and closed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResultMatch(t *testing.T) {
	cases := []struct {
		expect Expect
		open   bool
		want   bool
	}{
		{ExpectDefault, true, true},
		{ExpectDefault, false, false},
		{ExpectOpen, true, true},
		{ExpectOpen, false, false},
		{ExpectClosed, false, true},
		{ExpectClosed, true, false},
	}
	for _, c := range cases {
		r := Result{Target: Target{Host: "h", Port: 1, Expect: c.expect}, Open: c.open}
		if got := r.Match(); got != c.want {
			t.Errorf("expect=%v open=%v: Match() = %v, want %v", c.expect, c.open, got, c.want)
		}
	}
}

func TestAggregateMismatches(t *testing.T) {
	s := Aggregate([]Result{
		{Target: Target{Host: "a", Port: 1}, Open: false},
		{Target: Target{Host: "b", Port: 2, Expect: ExpectClosed}, Open: true},
		{Target: Target{Host: "c", Port: 3, Expect: ExpectClosed}, Open: false},
		{Target: Target{Host: "d", Port: 4, Expect: ExpectOpen}, Open: true},
	})
	if s.Mismatched != 2 || s.MismatchedExplicit != 1 {
		t.Fatalf("summary = %+v, want 2 mismatched, 1 explicit", s)
	}
}
