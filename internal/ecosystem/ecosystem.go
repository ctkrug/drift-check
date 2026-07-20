// Package ecosystem defines the common shape every language detector
// implements, and the reconciled result the report package renders.
package ecosystem

import (
	"errors"
	"os/exec"
)

// Pin is a single version claim from one source (a pin file, CI config,
// or the installed toolchain).
type Pin struct {
	Source  string // e.g. "go.mod", "installed", ".github/workflows/ci.yml"
	Version string
}

// Result is one ecosystem's reconciled view: every pin found for it, and
// whether they agree.
type Result struct {
	Ecosystem string
	Pins      []Pin
	Drift     bool
	Detail    string // human-readable explanation, empty when no drift
}

// Detector inspects a repo root for one ecosystem's version pins and
// resolves what's actually installed, returning nil (not an error) when
// the ecosystem isn't present in the repo at all.
type Detector interface {
	// Name is the ecosystem's display name, e.g. "Go", "Node".
	Name() string
	// Detect scans root and returns a reconciled Result, or nil if this
	// ecosystem has no pin files present in root.
	Detect(root string) (*Result, error)
}

const missingToolchainVersion = "not found"

// appendInstalledPin turns an absent executable into an explicit pin so an
// audit can distinguish an unavailable toolchain from one it never checked.
// Other command failures remain errors: silently ignoring them could produce
// a misleading clean report.
func appendInstalledPin(pins []Pin, version string, err error) ([]Pin, error) {
	if err == nil {
		if version != "" {
			pins = append(pins, Pin{Source: "installed", Version: version})
		}
		return pins, nil
	}
	if errors.Is(err, exec.ErrNotFound) {
		return append(pins, Pin{Source: "installed", Version: missingToolchainVersion}), nil
	}
	return nil, err
}
