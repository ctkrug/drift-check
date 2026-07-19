// Package report renders reconciled ecosystem.Result values as either a
// human-readable drift report or machine-readable JSON.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ctkrug/drift-check/internal/ecosystem"
)

// row holds one ecosystem's rendered line, pre-computed so column widths
// can be measured against every row's actual content before printing.
type row struct {
	ecosystem string
	sources   string
	status    string
}

// Write renders results to w, one line per ecosystem, and returns the
// number of ecosystems with drift. Column widths are sized to the widest
// value actually present in this run, so a long CI workflow path or
// version string in one ecosystem's pins doesn't misalign the status
// column for the others.
func Write(w io.Writer, results []*ecosystem.Result) int {
	drifted := 0
	var rows []row
	ecoWidth, sourcesWidth := 0, 0

	for _, r := range results {
		if r == nil {
			continue
		}
		status := "✓"
		if r.Drift {
			status = "⚠ DRIFT"
			drifted++
		}

		sourceParts := make([]string, len(r.Pins))
		for i, p := range r.Pins {
			sourceParts[i] = fmt.Sprintf("%s=%s", p.Source, p.Version)
		}
		sources := strings.Join(sourceParts, "  ")

		rows = append(rows, row{ecosystem: r.Ecosystem, sources: sources, status: status})
		if len(r.Ecosystem) > ecoWidth {
			ecoWidth = len(r.Ecosystem)
		}
		if len(sources) > sourcesWidth {
			sourcesWidth = len(sources)
		}
	}

	for _, l := range rows {
		fmt.Fprintf(w, "%-*s  %-*s  %s\n", ecoWidth, l.ecosystem, sourcesWidth, l.sources, l.status)
	}

	if drifted == 0 {
		fmt.Fprintf(w, "\nno drift found across %d ecosystem(s).\n", nonNil(results))
	} else {
		fmt.Fprintf(w, "\n%d drift(s) found across %d ecosystem(s).\n", drifted, nonNil(results))
	}
	return drifted
}

// jsonPin, jsonEcosystem, and jsonReport define the --json output shape.
type jsonPin struct {
	Source  string `json:"source"`
	Version string `json:"version"`
}

type jsonEcosystem struct {
	Ecosystem string    `json:"ecosystem"`
	Pins      []jsonPin `json:"pins"`
	Drift     bool      `json:"drift"`
	Detail    string    `json:"detail,omitempty"`
}

type jsonReport struct {
	Ecosystems []jsonEcosystem `json:"ecosystems"`
	Drift      bool            `json:"drift"`
	Message    string          `json:"message,omitempty"`
}

// WriteJSON renders results as a single JSON object to w and returns the
// number of ecosystems with drift. Exit-code behavior for the caller is
// identical to Write: non-zero when the returned count is greater than 0.
func WriteJSON(w io.Writer, results []*ecosystem.Result) int {
	out := jsonReport{Ecosystems: []jsonEcosystem{}}
	drifted := 0
	for _, r := range results {
		if r == nil {
			continue
		}
		pins := make([]jsonPin, len(r.Pins))
		for i, p := range r.Pins {
			pins[i] = jsonPin{Source: p.Source, Version: p.Version}
		}
		out.Ecosystems = append(out.Ecosystems, jsonEcosystem{
			Ecosystem: r.Ecosystem,
			Pins:      pins,
			Drift:     r.Drift,
			Detail:    r.Detail,
		})
		if r.Drift {
			drifted++
		}
	}
	out.Drift = drifted > 0

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(out) //nolint:errcheck // writer errors aren't actionable here
	return drifted
}

// WriteJSONEmpty renders the "no pin files found" case as JSON, so --json
// callers get valid JSON in every case rather than a plain-text fallback.
func WriteJSONEmpty(w io.Writer) {
	out := jsonReport{Ecosystems: []jsonEcosystem{}, Message: "no version pin files found"}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(out) //nolint:errcheck // writer errors aren't actionable here
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
