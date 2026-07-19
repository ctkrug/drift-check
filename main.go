// Command drift-check audits a polyglot monorepo and reports every place
// a pinned version file disagrees with what's installed or with CI.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ctkrug/drift-check/internal/ecosystem"
	"github.com/ctkrug/drift-check/internal/report"
)

const usageText = `drift-check [flags] [path]

Audits a polyglot monorepo and reports every place a pinned version file
disagrees with what's installed or with CI. Defaults to the current
directory. Exits non-zero when drift is found.

Flags:
  --json    output a machine-readable JSON report instead of text
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the CLI and returns the process exit code: 0 when no drift
// (or no pin files) is found, 1 when drift is found or a detector fails,
// 2 on a usage error.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("drift-check", flag.ContinueOnError)
	fs.SetOutput(stdout)
	fs.Usage = func() { fmt.Fprint(stdout, usageText) }
	jsonOutput := fs.Bool("json", false, "output a machine-readable JSON report")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}

	root := "."
	if fs.NArg() > 0 {
		root = fs.Arg(0)
	}

	detectors := []ecosystem.Detector{
		ecosystem.NewGoDetector(),
		ecosystem.NewNodeDetector(),
		ecosystem.NewPythonDetector(),
		ecosystem.NewRubyDetector(),
	}

	var results []*ecosystem.Result
	for _, d := range detectors {
		res, err := d.Detect(root)
		if err != nil {
			fmt.Fprintf(stderr, "drift-check: %s: %v\n", d.Name(), err)
			return 1
		}
		if res != nil {
			results = append(results, res)
		}
	}

	if len(results) == 0 {
		if *jsonOutput {
			report.WriteJSONEmpty(stdout)
		} else {
			fmt.Fprintln(stdout, "no version pin files found.")
		}
		return 0
	}

	var drifted int
	if *jsonOutput {
		drifted = report.WriteJSON(stdout, results)
	} else {
		drifted = report.Write(stdout, results)
	}
	if drifted > 0 {
		return 1
	}
	return 0
}
