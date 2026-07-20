package ecosystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeWorkflow(t *testing.T, dir, name, content string) string {
	t.Helper()
	wfDir := filepath.Join(dir, ".github", "workflows")
	if err := os.MkdirAll(wfDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(wfDir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestFindWorkflowPins_ExtractsQuotedVersion(t *testing.T) {
	dir := t.TempDir()
	writeWorkflow(t, dir, "ci.yml", `
name: CI
on: push
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
      - run: go build ./...
`)

	pins := findWorkflowPins(dir, "actions/setup-go", "go-version")
	if len(pins) != 1 {
		t.Fatalf("expected 1 pin, got %d: %+v", len(pins), pins)
	}
	if pins[0].version != "1.23" {
		t.Errorf("version = %q, want 1.23", pins[0].version)
	}
	if pins[0].source != ".github/workflows/ci.yml" {
		t.Errorf("source = %q, want .github/workflows/ci.yml", pins[0].source)
	}
}

func TestFindWorkflowPins_ExtractsUnquotedVersion(t *testing.T) {
	dir := t.TempDir()
	writeWorkflow(t, dir, "ci.yml", `
jobs:
  test:
    steps:
      - uses: actions/setup-go@v5
        with:
          go-version: 1.22
`)

	pins := findWorkflowPins(dir, "actions/setup-go", "go-version")
	if len(pins) != 1 || pins[0].version != "1.22" {
		t.Fatalf("unexpected pins: %+v", pins)
	}
}

func TestFindWorkflowPins_NoSetupStepProducesNoPin(t *testing.T) {
	dir := t.TempDir()
	writeWorkflow(t, dir, "ci.yml", `
jobs:
  test:
    steps:
      - uses: actions/checkout@v4
      - run: echo hi
`)

	pins := findWorkflowPins(dir, "actions/setup-go", "go-version")
	if len(pins) != 0 {
		t.Fatalf("expected no pins, got %+v", pins)
	}
}

func TestFindWorkflowPins_NoWorkflowsDir(t *testing.T) {
	pins := findWorkflowPins(t.TempDir(), "actions/setup-go", "go-version")
	if pins != nil {
		t.Fatalf("expected nil pins, got %+v", pins)
	}
}

func TestFindWorkflowPins_IgnoresOtherActionsWithField(t *testing.T) {
	dir := t.TempDir()
	writeWorkflow(t, dir, "ci.yml", `
jobs:
  test:
    steps:
      - uses: actions/setup-node@v4
        with:
          go-version: "99.99"
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24"
`)

	pins := findWorkflowPins(dir, "actions/setup-go", "go-version")
	if len(pins) != 1 || pins[0].version != "1.24" {
		t.Fatalf("expected only the setup-go pin, got %+v", pins)
	}
}

func TestFindWorkflowPins_DoesNotMatchActionPrefixLookalike(t *testing.T) {
	dir := t.TempDir()
	writeWorkflow(t, dir, "ci.yml", `
jobs:
  test:
    steps:
      - uses: actions/setup-goose@v1
        with:
          go-version: "99.99"
`)

	pins := findWorkflowPins(dir, "actions/setup-go", "go-version")
	if len(pins) != 0 {
		t.Fatalf("expected no pin from lookalike action, got %+v", pins)
	}
}

func TestFindWorkflowPins_MultipleFilesBothScanned(t *testing.T) {
	dir := t.TempDir()
	writeWorkflow(t, dir, "ci.yml", `
jobs:
  test:
    steps:
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
`)
	writeWorkflow(t, dir, "release.yml", `
jobs:
  release:
    steps:
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24"
`)

	pins := findWorkflowPins(dir, "actions/setup-go", "go-version")
	if len(pins) != 2 {
		t.Fatalf("expected 2 pins across 2 files, got %+v", pins)
	}
}

func TestFindWorkflowPins_MultipleStepsInOneFile(t *testing.T) {
	dir := t.TempDir()
	writeWorkflow(t, dir, "ci.yml", `
jobs:
  test-old:
    steps:
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
  test-new:
    steps:
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24"
`)

	pins := findWorkflowPins(dir, "actions/setup-go", "go-version")
	if len(pins) != 2 || pins[0].version != "1.23" || pins[1].version != "1.24" {
		t.Fatalf("expected both setup-go pins, got %+v", pins)
	}
}

func TestFindWorkflowPins_MalformedFileSkippedWithoutCrashing(t *testing.T) {
	dir := t.TempDir()
	// A single line far larger than bufio.Scanner's default token size
	// trips a scan error, standing in for "unparseable" input here.
	huge := strings.Repeat("a", 1<<20)
	writeWorkflow(t, dir, "broken.yml", huge)
	writeWorkflow(t, dir, "ci.yml", `
jobs:
  test:
    steps:
      - uses: actions/setup-go@v5
        with:
          go-version: "1.24"
`)

	pins := findWorkflowPins(dir, "actions/setup-go", "go-version")
	if len(pins) != 1 || pins[0].version != "1.24" {
		t.Fatalf("expected the valid file's pin despite the broken one, got %+v", pins)
	}
}
