package main

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/cybercapybara/portping/internal/scan"
)

// collectSpecs gathers target specifications from CLI args and, when requested,
// from a file or stdin. If no args and no file are given, stdin is read.
func collectSpecs(args []string, file string, stdin io.Reader) ([]string, error) {
	specs := append([]string{}, args...)

	var src io.Reader
	switch {
	case file == "-":
		src = stdin
	case file != "":
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		src = f
	case len(args) == 0:
		src = stdin
	}

	if src != nil {
		scanner := bufio.NewScanner(src)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			specs = append(specs, strings.Fields(line)...)
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
	return specs, nil
}

func filterOpen(results []scan.Result) []scan.Result {
	out := make([]scan.Result, 0, len(results))
	for _, r := range results {
		if r.Open {
			out = append(out, r)
		}
	}
	return out
}
