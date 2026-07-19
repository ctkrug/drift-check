package ecosystem

import (
	"os"
	"path/filepath"
	"strings"
)

// PythonDetector reconciles .python-version against the installed
// interpreter.
//
// TODO(BACKLOG.md epic 2): full reconciliation against `python --version`
// and CI pins. For now it only detects presence of a pin file so the
// scaffold has a real, if partial, implementation to build on.
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
	return &Result{
		Ecosystem: d.Name(),
		Pins:      []Pin{{Source: ".python-version", Version: version}},
	}, nil
}
