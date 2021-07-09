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

func TestParsePortsInvalid(t *testing.T) {
	cases := []string{"", "0", "70000", "abc"}
	for _, c := range cases {
		if _, err := ParsePorts(c); err == nil {
			t.Errorf("ParsePorts(%q) expected error, got nil", c)
		}
	}
}
