package scan

import (
	"reflect"
	"testing"
)

func TestExpandCIDR28(t *testing.T) {
	got, err := ExpandCIDR("10.0.0.0/28")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 14 {
		t.Fatalf("got %d hosts, want 14 (network+broadcast excluded)", len(got))
	}
	if got[0] != "10.0.0.1" {
		t.Errorf("first host = %s, want 10.0.0.1", got[0])
	}
	if got[len(got)-1] != "10.0.0.14" {
		t.Errorf("last host = %s, want 10.0.0.14", got[len(got)-1])
	}
}

func TestExpandCIDR31(t *testing.T) {
	got, err := ExpandCIDR("192.168.1.0/31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"192.168.1.0", "192.168.1.1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v (RFC 3021 both usable)", got, want)
	}
}

func TestExpandCIDR32(t *testing.T) {
	got, err := ExpandCIDR("192.168.1.5/32")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"192.168.1.5"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestExpandCIDR30(t *testing.T) {
	got, err := ExpandCIDR("172.16.0.0/30")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"172.16.0.1", "172.16.0.2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestExpandCIDRNonZeroBase(t *testing.T) {
	got, err := ExpandCIDR("10.1.2.192/29")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"10.1.2.193", "10.1.2.194", "10.1.2.195", "10.1.2.196", "10.1.2.197", "10.1.2.198"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestExpandCIDRErrors(t *testing.T) {
	cases := []string{"not-a-cidr", "10.0.0.0/33", "2001:db8::/64", "10.0.0.0"}
	for _, c := range cases {
		if _, err := ExpandCIDR(c); err == nil {
			t.Errorf("ExpandCIDR(%q) expected error, got nil", c)
		}
	}
}

// A /0 has 2^32 addresses, which overflows a uint32 counter to zero. That used
// to make ExpandCIDR return an empty host list with a nil error, so the CLI
// reported the misleading "no targets given" instead of rejecting the block.
func TestExpandCIDRSlashZeroIsRejected(t *testing.T) {
	hosts, err := ExpandCIDR("0.0.0.0/0")
	if err == nil {
		t.Fatalf("expected an error for /0, got %d hosts and nil error", len(hosts))
	}
	if len(hosts) != 0 {
		t.Errorf("expected no hosts alongside the error, got %d", len(hosts))
	}
}

func TestExpandCIDRTooLargeIsRejected(t *testing.T) {
	// A /8 covers 16.7M addresses; expanding it would allocate them all up
	// front, before a single dial.
	for _, c := range []string{"10.0.0.0/8", "10.0.0.0/1", "10.0.0.0/15"} {
		if _, err := ExpandCIDR(c); err == nil {
			t.Errorf("ExpandCIDR(%q) expected a too-large error, got nil", c)
		}
	}
}

func TestExpandCIDRAtLimitSucceeds(t *testing.T) {
	// A /16 is exactly MaxCIDRHosts addresses and must still be allowed.
	got, err := ExpandCIDR("10.0.0.0/16")
	if err != nil {
		t.Fatalf("unexpected error at the limit: %v", err)
	}
	if len(got) != MaxCIDRHosts-2 {
		t.Fatalf("got %d hosts, want %d (network+broadcast excluded)", len(got), MaxCIDRHosts-2)
	}
	if got[0] != "10.0.0.1" {
		t.Errorf("first host = %s, want 10.0.0.1", got[0])
	}
	if got[len(got)-1] != "10.0.255.254" {
		t.Errorf("last host = %s, want 10.0.255.254", got[len(got)-1])
	}
}
