package ecosystem

import (
	"os"
	"path/filepath"
	"strings"
)

// RubyDetector reconciles .ruby-version against the installed interpreter
// and Gemfile.lock.
//
// TODO(BACKLOG.md epic 2): full reconciliation against `ruby -v`, CI pins,
// and the RUBY VERSION stanza in Gemfile.lock. For now it only detects
// presence of a pin file so the scaffold has a real, if partial,
// implementation to build on.
type RubyDetector struct{}

func NewRubyDetector() *RubyDetector { return &RubyDetector{} }

func (d *RubyDetector) Name() string { return "Ruby" }

func (d *RubyDetector) Detect(root string) (*Result, error) {
	rv := filepath.Join(root, ".ruby-version")
	data, err := os.ReadFile(rv)
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
		Pins:      []Pin{{Source: ".ruby-version", Version: version}},
	}, nil
}
