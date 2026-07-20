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

func (d *GoDetector) Detect(projectRoot, repositoryRoot string) (*Result, error) {
	modPath := filepath.Join(projectRoot, "go.mod")
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

	for _, ciPin := range findWorkflowPins(repositoryRoot, "actions/setup-go", "go-version") {
		pins = append(pins, Pin{Source: ciPin.source, Version: ciPin.version})
	}

	installed, err := installedGoVersion()
	pins, err = appendInstalledPin(pins, installed, err)
	if err != nil {
		return nil, err
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
