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

func TestParseTargetIPv6Bracket(t *testing.T) {
	got, err := ParseTarget("[::1]:80")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []Target{{Host: "::1", Port: 80}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParseTargetErrors(t *testing.T) {
	cases := []string{"", "nohost", "host:", ":80", "host:99999", "10.0.0.0/28", "[::1]80"}
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

// hosts*ports is what actually gets allocated, so the product must be bounded
// even when each dimension is individually acceptable: a /16 (within
// MaxCIDRHosts) crossed with every port is ~4.3e9 pairs, which also overflows
// a 32-bit int.
func TestParseTargetRejectsOversizedProduct(t *testing.T) {
	if _, err := ParseTarget("10.0.0.0/16:1-65535"); err == nil {
		t.Fatal("expected an error for a /16 crossed with all ports, got nil")
	}
}

func TestParseTargetLargeButAllowedProduct(t *testing.T) {
	// 65534 hosts * 4 ports = 262136 pairs, under MaxTargets.
	got, err := ParseTarget("10.0.0.0/16:22,80,443,8080")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := (MaxCIDRHosts - 2) * 4; len(got) != want {
		t.Fatalf("got %d targets, want %d", len(got), want)
	}
}

func TestParseTargetsRejectsOversizedTotal(t *testing.T) {
	// Individually fine, but together they exceed MaxTargets.
	specs := []string{
		"10.0.0.0/16:1-8",
		"11.0.0.0/16:1-8",
		"12.0.0.0/16:1-8",
	}
	if _, err := ParseTargets(specs); err == nil {
		t.Fatal("expected an error once the cumulative total exceeds MaxTargets, got nil")
	}
}
