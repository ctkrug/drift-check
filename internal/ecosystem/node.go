package ecosystem

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// NodeDetector reconciles the version pinned in .nvmrc against what's
// installed on PATH and what CI pins via actions/setup-node.
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

	pins := []Pin{{Source: ".nvmrc", Version: version}}

	if ciPins := findWorkflowPins(root, "actions/setup-node", "node-version"); len(ciPins) > 0 {
		pins = append(pins, Pin{Source: ciPins[0].source, Version: ciPins[0].version})
	}

	if installed, err := installedNodeVersion(); err == nil && installed != "" {
		pins = append(pins, Pin{Source: "installed", Version: installed})
	}

	res := &Result{Ecosystem: d.Name(), Pins: pins}
	res.Drift, res.Detail = reconcile(pins)
	return res, nil
}

var nodeVersionOutputRe = regexp.MustCompile(`v?(\d+\.\d+(?:\.\d+)?)`)

// installedNodeVersion shells out to `node -v` and extracts the version,
// stripping the leading "v" node prints (e.g. "v20.11.0").
func installedNodeVersion() (string, error) {
	out, err := exec.Command("node", "-v").Output()
	if err != nil {
		return "", err
	}
	m := nodeVersionOutputRe.FindStringSubmatch(string(out))
	if m == nil {
		return "", nil
	}
	return m[1], nil
}
