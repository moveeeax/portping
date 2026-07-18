package main

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/cybercapybara/portping/internal/scan"
)

func renderTable(w io.Writer, results []scan.Result) {
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "TARGET\tSTATE\tLATENCY\tATTEMPTS\tDETAIL")
	for _, r := range results {
		state := "closed"
		latency := "-"
		detail := r.Err
		if r.Open {
			state = "open"
			latency = fmt.Sprintf("%.1fms", float64(r.Latency.Microseconds())/1000.0)
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d\t%s\n", r.Target.Addr(), state, latency, r.Attempts, detail)
	}
	tw.Flush()
}

type jsonResult struct {
	Host      string  `json:"host"`
	Port      int     `json:"port"`
	Open      bool    `json:"open"`
	LatencyMS float64 `json:"latency_ms"`
	Attempts  int     `json:"attempts"`
	Error     string  `json:"error,omitempty"`
}

func renderJSON(w io.Writer, results []scan.Result) error {
	out := make([]jsonResult, 0, len(results))
	for _, r := range results {
		jr := jsonResult{
			Host:     r.Target.Host,
			Port:     r.Target.Port,
			Open:     r.Open,
			Attempts: r.Attempts,
			Error:    r.Err,
		}
		if r.Open {
			jr.LatencyMS = float64(r.Latency.Microseconds()) / 1000.0
		}
		out = append(out, jr)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
