package ecosystem

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func writeNvmrc(t *testing.T, dir, version string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".nvmrc"), []byte(version), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestNodeDetector_NoNvmrc(t *testing.T) {
	dir := t.TempDir()
	res, err := NewNodeDetector().Detect(dir, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil result when .nvmrc is absent, got %+v", res)
	}
}

func TestNodeDetector_EmptyNvmrcTreatedAsAbsent(t *testing.T) {
	dir := t.TempDir()
	writeNvmrc(t, dir, "")

	res, err := NewNodeDetector().Detect(dir, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil result for empty .nvmrc, got %+v", res)
	}
}

func TestNodeDetector_NoDriftAgainstInstalled(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not on PATH")
	}
	installed, err := installedNodeVersion()
	if err != nil || installed == "" {
		t.Skip("could not resolve installed node version")
	}

	dir := t.TempDir()
	writeNvmrc(t, dir, installed)

	res, err := NewNodeDetector().Detect(dir, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Drift {
		t.Fatalf("expected no drift, got detail: %s", res.Detail)
	}
}

func TestNodeDetector_IncludesCIPinAndDetectsDrift(t *testing.T) {
	dir := t.TempDir()
	writeNvmrc(t, dir, "20.11.0")
	writeWorkflow(t, dir, "ci.yml", `
jobs:
  test:
    steps:
      - uses: actions/setup-node@v4
        with:
          node-version: "18.19.0"
`)

	res, err := NewNodeDetector().Detect(dir, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Pins) < 2 {
		t.Fatalf("expected at least 2 pins (.nvmrc, ci), got %+v", res.Pins)
	}
	if res.Pins[1].Version != "18.19.0" {
		t.Fatalf("unexpected ci pin: %+v", res.Pins[1])
	}
	if !res.Drift {
		t.Fatal("expected drift between .nvmrc=20.11.0 and ci=18.19.0")
	}
	if !strings.Contains(res.Detail, "20.11.0") || !strings.Contains(res.Detail, "18.19.0") {
		t.Errorf("detail %q missing expected versions", res.Detail)
	}
}

func TestNodeDetector_ReportsMissingInstalledToolchain(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	dir := t.TempDir()
	writeNvmrc(t, dir, "20.11.0")

	res, err := NewNodeDetector().Detect(dir, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || !res.Drift {
		t.Fatalf("expected missing installed Node to cause drift, got %+v", res)
	}
	if got := res.Pins[len(res.Pins)-1]; got.Source != "installed" || got.Version != "not found" {
		t.Fatalf("installed pin = %+v, want installed/not found", got)
	}
}
