package main

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/moveeeax/portping/internal/scan"
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
