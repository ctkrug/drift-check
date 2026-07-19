package ecosystem

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func writePythonVersion(t *testing.T, dir, version string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".python-version"), []byte(version), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPythonDetector_NoPin(t *testing.T) {
	res, err := NewPythonDetector().Detect(t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil result when .python-version is absent, got %+v", res)
	}
}

func TestPythonDetector_EmptyFileTreatedAsAbsent(t *testing.T) {
	dir := t.TempDir()
	writePythonVersion(t, dir, "")

	res, err := NewPythonDetector().Detect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil result for empty .python-version, got %+v", res)
	}
}

func TestPythonDetector_PrefixAgreesWithPatchInstall(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not on PATH")
	}
	installed, err := installedPythonVersion()
	if err != nil || installed == "" {
		t.Skip("could not resolve installed python version")
	}
	parts := strings.Split(installed, ".")
	if len(parts) < 2 {
		t.Skip("installed version too short to test prefix agreement")
	}
	majorMinor := parts[0] + "." + parts[1]

	dir := t.TempDir()
	writePythonVersion(t, dir, majorMinor)

	res, err := NewPythonDetector().Detect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Drift {
		t.Fatalf("expected %s to agree with installed %s, got drift: %s", majorMinor, installed, res.Detail)
	}
}

func TestPythonDetector_IncludesCIPinAndDetectsDrift(t *testing.T) {
	dir := t.TempDir()
	writePythonVersion(t, dir, "3.12.1")
	writeWorkflow(t, dir, "ci.yml", `
jobs:
  test:
    steps:
      - uses: actions/setup-python@v5
        with:
          python-version: "3.11.7"
`)

	res, err := NewPythonDetector().Detect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Pins) < 2 || res.Pins[1].Version != "3.11.7" {
		t.Fatalf("unexpected pins: %+v", res.Pins)
	}
	if !res.Drift {
		t.Fatal("expected drift between .python-version=3.12.1 and ci=3.11.7")
	}
}
