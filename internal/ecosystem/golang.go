package ecosystem

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// GoDetector reconciles the `go` directive in go.mod against the Go
// toolchain actually on PATH.
type GoDetector struct{}

func NewGoDetector() *GoDetector { return &GoDetector{} }

func (d *GoDetector) Name() string { return "Go" }

var goModDirectiveRe = regexp.MustCompile(`^go\s+(\d+\.\d+(?:\.\d+)?)`)

func (d *GoDetector) Detect(root string) (*Result, error) {
	modPath := filepath.Join(root, "go.mod")
	f, err := os.Open(modPath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pinned, err := parseGoModVersion(f)
	if err != nil {
		return nil, err
	}
	if pinned == "" {
		return nil, nil
	}

	pins := []Pin{{Source: "go.mod", Version: pinned}}

	if installed, err := installedGoVersion(); err == nil && installed != "" {
		pins = append(pins, Pin{Source: "installed", Version: installed})
	}

	res := &Result{Ecosystem: d.Name(), Pins: pins}
	res.Drift, res.Detail = reconcile(pins)
	return res, nil
}

// parseGoModVersion extracts the version from a `go X.Y` directive line.
func parseGoModVersion(f *os.File) (string, error) {
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if m := goModDirectiveRe.FindStringSubmatch(line); m != nil {
			return m[1], nil
		}
	}
	return "", scanner.Err()
}

var goVersionOutputRe = regexp.MustCompile(`go(\d+\.\d+(?:\.\d+)?)`)

// installedGoVersion shells out to `go version` and extracts the version.
func installedGoVersion() (string, error) {
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		return "", err
	}
	m := goVersionOutputRe.FindStringSubmatch(string(out))
	if m == nil {
		return "", nil
	}
	return m[1], nil
}

// reconcile compares a set of pins for the same ecosystem and reports
// whether they disagree. Versions are compared as dotted prefixes: "1.24"
// matches "1.24.3" (a go.mod directive doesn't pin a patch version).
func reconcile(pins []Pin) (drift bool, detail string) {
	if len(pins) < 2 {
		return false, ""
	}
	base := pins[0]
	for _, p := range pins[1:] {
		if !versionsAgree(base.Version, p.Version) {
			return true, base.Source + " says " + base.Version + ", " +
				p.Source + " says " + p.Version
		}
	}
	return false, ""
}

// versionsAgree reports whether two version strings are compatible,
// comparing only as many dotted components as the shorter one has (e.g.
// "1.24" agrees with "1.24.3", but not with "1.23").
func versionsAgree(a, b string) bool {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	n := len(pa)
	if len(pb) < n {
		n = len(pb)
	}
	for i := 0; i < n; i++ {
		if pa[i] != pb[i] {
			return false
		}
	}
	return true
}
