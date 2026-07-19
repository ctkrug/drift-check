// Command drift-check audits a polyglot monorepo and reports every place
// a pinned version file disagrees with what's installed or with CI.
package main

import (
	"fmt"
	"os"

	"github.com/ctkrug/drift-check/internal/ecosystem"
	"github.com/ctkrug/drift-check/internal/report"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help":
			printUsage()
			return
		default:
			root = os.Args[1]
		}
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
			fmt.Fprintf(os.Stderr, "drift-check: %s: %v\n", d.Name(), err)
			os.Exit(1)
		}
		if res != nil {
			results = append(results, res)
		}
	}

	if len(results) == 0 {
		fmt.Println("no version pin files found.")
		return
	}

	drifted := report.Write(os.Stdout, results)
	if drifted > 0 {
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`drift-check [path]

Audits a polyglot monorepo and reports every place a pinned version file
disagrees with what's installed or with CI. Defaults to the current
directory. Exits non-zero when drift is found.`)
}
