package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/moveeeax/portping/internal/probe"
	"github.com/moveeeax/portping/internal/scan"
)

type options struct {
	timeout     time.Duration
	concurrency int
	count       int
	jsonOut     bool
	openOnly    bool
	file        string
	failClosed  bool
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("portping", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintf(stderr, "usage: portping [flags] target [target ...]\n\n")
		fmt.Fprintf(stderr, "A target is HOST:PORTS, where HOST may be a hostname, IP, or IPv4 CIDR\n")
		fmt.Fprintf(stderr, "and PORTS may be a list and/or range, e.g. 10.0.0.0/28:22,80,8000-8010\n\n")
		fmt.Fprintf(stderr, "flags:\n")
		fs.PrintDefaults()
	}

	var opt options
	fs.DurationVar(&opt.timeout, "timeout", 2*time.Second, "per-dial timeout")
	fs.IntVar(&opt.concurrency, "concurrency", 64, "number of concurrent workers")
	fs.IntVar(&opt.count, "count", 1, "dial attempts before a target is unreachable")
	fs.BoolVar(&opt.jsonOut, "json", false, "emit results as a JSON array")
	fs.BoolVar(&opt.openOnly, "open-only", false, "only report reachable targets")
	fs.StringVar(&opt.file, "file", "", "read targets from file (use - for stdin)")
	fs.BoolVar(&opt.failClosed, "fail-on-unreachable", true, "exit non-zero if any target is unreachable")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	specs, err := collectSpecs(fs.Args(), opt.file, stdin)
	if err != nil {
		fmt.Fprintf(stderr, "portping: %v\n", err)
		return 2
	}

	targets, err := scan.ParseTargets(specs)
	if err != nil {
		fmt.Fprintf(stderr, "portping: %v\n", err)
		return 2
	}

	prober := probe.New(opt.timeout, opt.count)
	results := scan.Run(context.Background(), targets, opt.concurrency, prober.Probe)
	summary := scan.Aggregate(results)

	shown := results
	if opt.openOnly {
		shown = filterOpen(results)
	}

	if opt.jsonOut {
		if err := renderJSON(stdout, shown); err != nil {
			fmt.Fprintf(stderr, "portping: %v\n", err)
			return 2
		}
	} else {
		renderTable(stdout, shown)
	}

	if opt.failClosed && !summary.AllOpen() {
		return 1
	}
	return 0
}
