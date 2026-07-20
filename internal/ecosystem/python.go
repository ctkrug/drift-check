package ecosystem

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// PythonDetector reconciles the version pinned in .python-version against
// what's installed on PATH and what CI pins via actions/setup-python.
type PythonDetector struct{}

func NewPythonDetector() *PythonDetector { return &PythonDetector{} }

func (d *PythonDetector) Name() string { return "Python" }

func (d *PythonDetector) Detect(root string) (*Result, error) {
	pv := filepath.Join(root, ".python-version")
	data, err := os.ReadFile(pv)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	version := strings.TrimSpace(string(data))
	if version == "" {
		return nil, nil
	}

	pins := []Pin{{Source: ".python-version", Version: version}}

	if ciPins := findWorkflowPins(root, "actions/setup-python", "python-version"); len(ciPins) > 0 {
		pins = append(pins, Pin{Source: ciPins[0].source, Version: ciPins[0].version})
	}

	installed, err := installedPythonVersion()
	pins, err = appendInstalledPin(pins, installed, err)
	if err != nil {
		return nil, err
	}

	res := &Result{Ecosystem: d.Name(), Pins: pins}
	res.Drift, res.Detail = reconcile(pins)
	return res, nil
}

var pythonVersionOutputRe = regexp.MustCompile(`(\d+\.\d+(?:\.\d+)?)`)

// installedPythonVersion shells out to `python3 --version`, falling back to
// `python --version`, and extracts the version. Python 2 and old 3.x wrote
// the version to stderr, so both streams are checked.
func installedPythonVersion() (string, error) {
	var lastErr error
	for _, bin := range []string{"python3", "python"} {
		out, err := exec.Command(bin, "--version").CombinedOutput()
		if err != nil {
			if !errors.Is(err, exec.ErrNotFound) {
				lastErr = err
			}
			continue
		}
		if m := pythonVersionOutputRe.FindStringSubmatch(string(out)); m != nil {
			return m[1], nil
		}
	}
	if lastErr != nil {
		return "", lastErr
	}
	return "", exec.ErrNotFound
}
