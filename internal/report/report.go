// Package report renders reconciled ecosystem.Result values as a
// human-readable drift report.
package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/ctkrug/drift-check/internal/ecosystem"
)

// Write renders results to w, one line per ecosystem, and returns the
// number of ecosystems with drift.
func Write(w io.Writer, results []*ecosystem.Result) int {
	drifted := 0
	for _, r := range results {
		if r == nil {
			continue
		}
		status := "✓"
		if r.Drift {
			status = "⚠ DRIFT"
			drifted++
		}

		sources := make([]string, len(r.Pins))
		for i, p := range r.Pins {
			sources[i] = fmt.Sprintf("%s=%s", p.Source, p.Version)
		}
		fmt.Fprintf(w, "%-8s %-40s %s\n", r.Ecosystem, strings.Join(sources, "  "), status)
	}

	if drifted == 0 {
		fmt.Fprintf(w, "\nno drift found across %d ecosystem(s).\n", nonNil(results))
	} else {
		fmt.Fprintf(w, "\n%d drift(s) found across %d ecosystem(s).\n", drifted, nonNil(results))
	}
	return drifted
}

func nonNil(results []*ecosystem.Result) int {
	n := 0
	for _, r := range results {
		if r != nil {
			n++
		}
	}
	return n
}
