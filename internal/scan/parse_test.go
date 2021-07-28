package scan

import (
	"reflect"
	"testing"
)

func TestParseTargetSingle(t *testing.T) {
	got, err := ParseTarget("example.com:443")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []Target{{Host: "example.com", Port: 443}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParseTargetCIDRAndPortRange(t *testing.T) {
	got, err := ParseTarget("10.0.0.0/30:22,80")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []Target{
		{Host: "10.0.0.1", Port: 22},
		{Host: "10.0.0.1", Port: 80},
		{Host: "10.0.0.2", Port: 22},
		{Host: "10.0.0.2", Port: 80},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParseTargetErrors(t *testing.T) {
	cases := []string{"", "nohost", "host:", ":80", "host:99999", "10.0.0.0/28"}
	for _, c := range cases {
		if _, err := ParseTarget(c); err == nil {
			t.Errorf("ParseTarget(%q) expected error, got nil", c)
		}
	}
}

func TestParseTargetsDedup(t *testing.T) {
	got, err := ParseTargets([]string{"127.0.0.1:80", "127.0.0.1:80,443", "  "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []Target{
		{Host: "127.0.0.1", Port: 80},
		{Host: "127.0.0.1", Port: 443},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParseTargetsEmpty(t *testing.T) {
	if _, err := ParseTargets([]string{"  ", ""}); err == nil {
		t.Fatal("expected error for no targets")
	}
}
