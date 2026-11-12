package scan

import (
	"fmt"
	"strings"
)

// Expect is the state a target is expected to be in.
type Expect int

const (
	// ExpectDefault means no suffix was given. The target is expected to be
	// open, but the expectation was not asserted explicitly, so
	// --fail-on-unreachable=false may waive it.
	ExpectDefault Expect = iota
	// ExpectOpen is an explicit "=open" assertion.
	ExpectOpen
	// ExpectClosed is an explicit "=closed" assertion.
	ExpectClosed
)

// WantOpen reports whether the target is expected to be reachable.
func (e Expect) WantOpen() bool {
	return e != ExpectClosed
}

// Explicit reports whether the expectation came from an "=open" or "=closed"
// suffix rather than the default.
func (e Expect) Explicit() bool {
	return e != ExpectDefault
}

// String returns "open" or "closed".
func (e Expect) String() string {
	if e.WantOpen() {
		return "open"
	}
	return "closed"
}

// splitExpect strips an optional "=open" or "=closed" suffix from a target
// specification. The suffix is matched case-sensitively; any other value after
// "=" is an error.
func splitExpect(spec string) (rest string, expect Expect, err error) {
	idx := strings.LastIndex(spec, "=")
	if idx < 0 {
		return spec, ExpectDefault, nil
	}
	switch v := spec[idx+1:]; v {
	case "open":
		return spec[:idx], ExpectOpen, nil
	case "closed":
		return spec[:idx], ExpectClosed, nil
	default:
		return "", ExpectDefault, fmt.Errorf("invalid target %q: expected state must be open or closed, got %q", spec, v)
	}
}
