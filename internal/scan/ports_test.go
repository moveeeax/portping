package scan

import (
	"reflect"
	"testing"
)

func TestParsePortsSingle(t *testing.T) {
	got, err := ParsePorts("80")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, []int{80}) {
		t.Fatalf("got %v, want [80]", got)
	}
}

func TestParsePortsList(t *testing.T) {
	got, err := ParsePorts("22,80,443")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, []int{22, 80, 443}) {
		t.Fatalf("got %v, want [22 80 443]", got)
	}
}

func TestParsePortsRange(t *testing.T) {
	got, err := ParsePorts("8000-8003")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, []int{8000, 8001, 8002, 8003}) {
		t.Fatalf("got %v", got)
	}
}

func TestParsePortsMixedAndDedup(t *testing.T) {
	got, err := ParsePorts("80,8000-8002,80,8001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, []int{80, 8000, 8001, 8002}) {
		t.Fatalf("got %v, want dedup order-preserving", got)
	}
}

func TestParsePortsTrailingCommaLenient(t *testing.T) {
	got, err := ParsePorts("22,80,")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, []int{22, 80}) {
		t.Fatalf("got %v", got)
	}
}

func TestParsePortsErrors(t *testing.T) {
	cases := []string{"", "0", "70000", "abc", "80-", "-80", "100-50"}
	for _, c := range cases {
		if _, err := ParsePorts(c); err == nil {
			t.Errorf("ParsePorts(%q) expected error, got nil", c)
		}
	}
}
