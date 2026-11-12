package scan

import (
	"net"
	"strconv"
)

// Target is a single host:port pair to probe, together with the state it is
// expected to be in.
type Target struct {
	Host   string
	Port   int
	Expect Expect
}

// Addr returns the dial address for the target.
func (t Target) Addr() string {
	return net.JoinHostPort(t.Host, strconv.Itoa(t.Port))
}

// String returns a human readable representation of the target.
func (t Target) String() string {
	return t.Addr()
}
