package scan

import (
	"net"
	"strconv"
)

// Target is a single host:port pair to probe.
type Target struct {
	Host string
	Port int
}

// Addr returns the dial address for the target.
func (t Target) Addr() string {
	return net.JoinHostPort(t.Host, strconv.Itoa(t.Port))
}
