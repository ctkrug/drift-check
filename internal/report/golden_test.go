package report

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctkrug/drift-check/internal/ecosystem"
)

// TestWrite_AllFourEcosystemsAligned guards against column misalignment
// when all four ecosystems are present together, including one with a
// long CI workflow path as a source label.
func TestWrite_AllFourEcosystemsAligned(t *testing.T) {
	results := []*ecosystem.Result{
		{
			Ecosystem: "Go",
			Pins: []ecosystem.Pin{
				{Source: "go.mod", Version: "1.24"},
				{Source: ".github/workflows/ci.yml", Version: "1.23"},
				{Source: "installed", Version: "1.22.2"},
			},
			Drift:  true,
			Detail: "go.mod says 1.24, .github/workflows/ci.yml says 1.23, installed says 1.22.2",
		},
		{
			Ecosystem: "Node",
			Pins: []ecosystem.Pin{
				{Source: ".nvmrc", Version: "20.11.0"},
				{Source: "installed", Version: "20.11.0"},
			},
			Drift: false,
		},
		{
			Ecosystem: "Python",
			Pins: []ecosystem.Pin{
				{Source: ".python-version", Version: "3.12.1"},
				{Source: ".github/workflows/ci.yml", Version: "3.11.7"},
				{Source: "installed", Version: "3.12.1"},
			},
			Drift:  true,
			Detail: ".python-version says 3.12.1, .github/workflows/ci.yml says 3.11.7, installed says 3.12.1",
		},
		{
			Ecosystem: "Ruby",
			Pins: []ecosystem.Pin{
				{Source: ".ruby-version", Version: "3.3.0"},
				{Source: "Gemfile.lock", Version: "3.3.0"},
			},
			Drift: false,
		},
	}

	var buf bytes.Buffer
	Write(&buf, results)

	golden := filepath.Join("testdata", "golden_report.txt")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(golden, buf.Bytes(), 0o644); err != nil {
			t.Fatalf("writing golden file: %v", err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("reading golden file: %v", err)
	}
	if buf.String() != string(want) {
		t.Errorf("report output doesn't match golden file %s\n--- got ---\n%s\n--- want ---\n%s", golden, buf.String(), string(want))
	}
}
