// Command drift-check audits a polyglot monorepo and reports every place
// a pinned version file disagrees with what's installed or with CI.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
	if fs.NArg() > 1 {
		fmt.Fprint(stdout, usageText)
		return 2
	}
	if fs.NArg() == 1 {
		root = fs.Arg(0)
	}

	detectors := []ecosystem.Detector{
		ecosystem.NewGoDetector(),
		ecosystem.NewNodeDetector(),
		ecosystem.NewPythonDetector(),
		ecosystem.NewRubyDetector(),
	}

	projectRoots, err := ecosystem.FindProjectRoots(root)
	if err != nil {
		fmt.Fprintf(stderr, "drift-check: %s\n", err)
		return 1
	}

	var results []*ecosystem.Result
	for _, projectRoot := range projectRoots {
		for _, d := range detectors {
			res, err := d.Detect(projectRoot, root)
			if err != nil {
				fmt.Fprintf(stderr, "drift-check: %s: %v\n", d.Name(), err)
				return 1
			}
			if res != nil {
				prefixResultSources(res, root, projectRoot)
				results = append(results, res)
			}
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

// prefixResultSources identifies pins belonging to nested projects without
// obscuring the shared local "installed" runtime source.
func prefixResultSources(res *ecosystem.Result, root, projectRoot string) {
	rel, err := filepath.Rel(root, projectRoot)
	if err != nil || rel == "." {
		return
	}
	prefix := filepath.ToSlash(rel)
	for i := range res.Pins {
		if res.Pins[i].Source != "installed" &&
			!strings.HasPrefix(res.Pins[i].Source, ".github/workflows/") {
			res.Pins[i].Source = prefix + "/" + res.Pins[i].Source
		}
	}
	if res.Drift {
		parts := make([]string, len(res.Pins))
		for i, pin := range res.Pins {
			parts[i] = pin.Source + " says " + pin.Version
		}
		res.Detail = strings.Join(parts, ", ")
	}
}
