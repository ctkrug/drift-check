package ecosystem

import (
	"os"
	"path/filepath"
	"strings"
)

// NodeDetector reconciles .nvmrc / package.json engines against the
// installed Node version.
//
// TODO(BACKLOG.md epic 2): full reconciliation against `node -v` and CI
// pins. For now it only detects presence of a pin file so the scaffold
// has a real, if partial, implementation to build on.
type NodeDetector struct{}

func NewNodeDetector() *NodeDetector { return &NodeDetector{} }

func (d *NodeDetector) Name() string { return "Node" }

func (d *NodeDetector) Detect(root string) (*Result, error) {
	nvmrc := filepath.Join(root, ".nvmrc")
	data, err := os.ReadFile(nvmrc)
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
		Pins:      []Pin{{Source: ".nvmrc", Version: version}},
	}, nil
}
