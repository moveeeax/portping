package scan

import (
	"fmt"
	"strconv"
	"strings"
)

// parsePort parses a single port number and validates its range.
func parsePort(s string) (int, error) {
	s = strings.TrimSpace(s)
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid port %q", s)
	}
	if n < 1 || n > 65535 {
		return 0, fmt.Errorf("port %d out of range (1-65535)", n)
	}
	return n, nil
}

// ParsePorts expands a comma-separated port list such as "22,80,443" into a
// de-duplicated, order-preserving slice of ports.
func ParsePorts(spec string) ([]int, error) {
	parts := strings.Split(spec, ",")
	seen := make(map[int]bool)
	var ports []int
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		p, err := parsePort(part)
		if err != nil {
			return nil, err
		}
		if !seen[p] {
			seen[p] = true
			ports = append(ports, p)
		}
	}
	if len(ports) == 0 {
		return nil, fmt.Errorf("no ports in %q", spec)
	}
	return ports, nil
}
