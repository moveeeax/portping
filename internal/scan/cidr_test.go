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
