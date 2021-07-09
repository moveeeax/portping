package scan

import "testing"

func TestTargetAddr(t *testing.T) {
	if got := (Target{Host: "10.0.0.1", Port: 22}).Addr(); got != "10.0.0.1:22" {
		t.Fatalf("Addr() = %s", got)
	}
	if got := (Target{Host: "::1", Port: 80}).Addr(); got != "[::1]:80" {
		t.Fatalf("Addr() = %s, want bracketed IPv6", got)
	}
}
