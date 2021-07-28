package scan

import (
	"encoding/binary"
	"fmt"
	"net"
)

// ExpandCIDR expands an IPv4 CIDR block into its host addresses.
//
// For prefixes of /30 and shorter the network and broadcast addresses are
// excluded. A /31 (RFC 3021 point-to-point link) yields both addresses and a
// /32 yields the single host address.
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
	size := uint32(1) << hostBits

	skipEdges := hostBits >= 2
	var hosts []string
	for i := uint32(0); i < size; i++ {
		if skipEdges && (i == 0 || i == size-1) {
			continue
		}
		hosts = append(hosts, uint32ToIPv4(start+i))
	}
	return hosts, nil
}

func uint32ToIPv4(n uint32) string {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], n)
	return net.IPv4(b[0], b[1], b[2], b[3]).String()
}
