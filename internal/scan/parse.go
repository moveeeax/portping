package scan

import (
	"fmt"
	"strings"
)

// MaxTargets is the largest number of distinct host:port pairs a single
// invocation may expand to. Hosts multiply by ports, so even a CIDR within
// MaxCIDRHosts can produce billions of pairs when combined with a wide port
// range; this bounds the total rather than each dimension.
const MaxTargets = 1 << 20

// splitHostPorts splits a "host:ports" specification into its host and port
// components. Bracketed IPv6 literals ("[::1]:80") are supported.
func splitHostPorts(spec string) (host, ports string, err error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "", "", fmt.Errorf("empty target")
	}
	if strings.HasPrefix(spec, "[") {
		end := strings.Index(spec, "]")
		if end < 0 {
			return "", "", fmt.Errorf("invalid target %q: missing ]", spec)
		}
		host = spec[1:end]
		rest := spec[end+1:]
		if !strings.HasPrefix(rest, ":") {
			return "", "", fmt.Errorf("invalid target %q: missing port", spec)
		}
		ports = rest[1:]
	} else {
		idx := strings.LastIndex(spec, ":")
		if idx < 0 {
			return "", "", fmt.Errorf("invalid target %q: missing port", spec)
		}
		host = spec[:idx]
		ports = spec[idx+1:]
	}
	if strings.TrimSpace(host) == "" {
		return "", "", fmt.Errorf("invalid target %q: empty host", spec)
	}
	if strings.TrimSpace(ports) == "" {
		return "", "", fmt.Errorf("invalid target %q: empty port", spec)
	}
	return host, ports, nil
}

// ParseTarget expands a single target specification into concrete targets. The
// host component may be a hostname, an IP literal, or an IPv4 CIDR block; the
// port component may be a list and/or range such as "22,80,8000-8010".
//
// It fails if the specification expands to more than MaxTargets pairs.
func ParseTarget(spec string) ([]Target, error) {
	host, portSpec, err := splitHostPorts(spec)
	if err != nil {
		return nil, err
	}
	ports, err := ParsePorts(portSpec)
	if err != nil {
		return nil, err
	}

	var hosts []string
	if strings.Contains(host, "/") {
		hosts, err = ExpandCIDR(host)
		if err != nil {
			return nil, err
		}
	} else {
		hosts = []string{host}
	}

	// Computed in uint64: hosts*ports can reach ~4.3e9, which overflows a
	// 32-bit int, and the product must be checked before it is allocated.
	if n := uint64(len(hosts)) * uint64(len(ports)); n > MaxTargets {
		return nil, fmt.Errorf("target %q expands to %d host:port pairs, more than the %d supported", spec, n, MaxTargets)
	}

	targets := make([]Target, 0, len(hosts)*len(ports))
	for _, h := range hosts {
		for _, p := range ports {
			targets = append(targets, Target{Host: h, Port: p})
		}
	}
	return targets, nil
}

// ParseTargets expands and de-duplicates a list of target specifications while
// preserving first-seen order.
//
// It fails if the specifications expand to more than MaxTargets distinct pairs.
func ParseTargets(specs []string) ([]Target, error) {
	seen := make(map[Target]bool)
	var targets []Target
	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}
		expanded, err := ParseTarget(spec)
		if err != nil {
			return nil, err
		}
		for _, t := range expanded {
			if !seen[t] {
				if len(targets) >= MaxTargets {
					return nil, fmt.Errorf("too many targets: the given specifications expand to more than %d host:port pairs", MaxTargets)
				}
				seen[t] = true
				targets = append(targets, t)
			}
		}
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("no targets given")
	}
	return targets, nil
}
