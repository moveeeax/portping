package scan

import (
	"fmt"
	"strings"
)

// splitHostPorts splits a "host:ports" specification into its host and port
// components.
func splitHostPorts(spec string) (host, ports string, err error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "", "", fmt.Errorf("empty target")
	}
	idx := strings.LastIndex(spec, ":")
	if idx < 0 {
		return "", "", fmt.Errorf("invalid target %q: missing port", spec)
	}
	host = spec[:idx]
	ports = spec[idx+1:]
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
