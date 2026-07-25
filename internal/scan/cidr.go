package scan

import (
	"encoding/binary"
	"fmt"
	"net"
)

// MaxCIDRHosts is the largest number of addresses a single CIDR block may
// expand to. It bounds the work and memory a single target specification can
// demand: without it, a short prefix such as /8 would materialise millions of
// addresses before the first dial. 65536 corresponds to a /16.
const MaxCIDRHosts = 65536

// ExpandCIDR expands an IPv4 CIDR block into its host addresses.
//
// For prefixes of /30 and shorter the network and broadcast addresses are
// excluded. A /31 (RFC 3021 point-to-point link) yields both addresses and a
// /32 yields the single host address.
//
// Blocks larger than MaxCIDRHosts addresses are rejected rather than expanded.
func ExpandCIDR(cidr string) ([]string, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR %q: %w", cidr, err)
	}
	ones, bits := ipnet.Mask.Size()
	if bits != 32 {
		return nil, fmt.Errorf("only IPv4 CIDR blocks are supported: %q", cidr)
	}
	base := ipnet.IP.To4()
	if base == nil {
		return nil, fmt.Errorf("invalid IPv4 CIDR %q", cidr)
	}
	start := binary.BigEndian.Uint32(base)
	hostBits := uint(bits - ones)

	// Computed in uint64: a /0 has 2^32 addresses, which overflows uint32 to 0
	// and would otherwise silently yield an empty host list.
	size := uint64(1) << hostBits
	if size > MaxCIDRHosts {
		return nil, fmt.Errorf("CIDR %q covers %d addresses, more than the %d supported by a single target", cidr, size, MaxCIDRHosts)
	}

	skipEdges := hostBits >= 2
	hosts := make([]string, 0, size)
	for i := uint64(0); i < size; i++ {
		if skipEdges && (i == 0 || i == size-1) {
			continue
		}
		hosts = append(hosts, uint32ToIPv4(start+uint32(i)))
	}
	return hosts, nil
}

func uint32ToIPv4(n uint32) string {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], n)
	return net.IPv4(b[0], b[1], b[2], b[3]).String()
}
