package ecosystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeGoMod(t *testing.T, dir, directive string) {
	t.Helper()
	content := "module example.com/x\n\n" + directive + "\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestGoDetector_NoGoMod(t *testing.T) {
	dir := t.TempDir()
	res, err := NewGoDetector().Detect(dir, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil result when go.mod is absent, got %+v", res)
	}
}

func TestGoDetector_ParsesDirective(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir, "go 1.24")

	res, err := NewGoDetector().Detect(dir, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Fatal("expected a result, got nil")
	}
	if res.Pins[0].Source != "go.mod" || res.Pins[0].Version != "1.24" {
		t.Fatalf("unexpected pin: %+v", res.Pins[0])
	}
}

func TestGoDetector_ThreeWayIncludesCIPin(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir, "go 1.24")
	writeWorkflow(t, dir, "ci.yml", `
jobs:
  test:
    steps:
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
`)

	res, err := NewGoDetector().Detect(dir, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Pins) != 3 {
		t.Fatalf("expected 3 pins (go.mod, ci, installed), got %d: %+v", len(res.Pins), res.Pins)
	}
	if res.Pins[1].Source != ".github/workflows/ci.yml" || res.Pins[1].Version != "1.23" {
		t.Fatalf("unexpected ci pin: %+v", res.Pins[1])
	}
	if !res.Drift {
		t.Fatal("expected drift: go.mod=1.24, ci=1.23")
	}
	if !strings.Contains(res.Detail, "go.mod says 1.24") || !strings.Contains(res.Detail, "1.23") {
		t.Errorf("detail %q missing expected sources", res.Detail)
	}
}

func TestGoDetector_UsesRepositoryWorkflowsForNestedProject(t *testing.T) {
	repositoryRoot := t.TempDir()
	projectRoot := filepath.Join(repositoryRoot, "services", "api")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	writeGoMod(t, projectRoot, "go 1.24")
	writeWorkflow(t, repositoryRoot, "ci.yml", `
jobs:
  test:
    steps:
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
`)

	res, err := NewGoDetector().Detect(projectRoot, repositoryRoot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Pins) < 2 || res.Pins[1].Source != ".github/workflows/ci.yml" || res.Pins[1].Version != "1.23" {
		t.Fatalf("nested project did not include repository workflow pin: %+v", res.Pins)
	}
}

func TestGoDetector_IncludesEveryWorkflowPin(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir, "go 1.24")
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
          go-version: "1.22"
`)

	res, err := NewGoDetector().Detect(dir, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]string{
		".github/workflows/ci.yml":      "1.23",
		".github/workflows/release.yml": "1.22",
	}
	for _, pin := range res.Pins {
		if version, ok := want[pin.Source]; ok && pin.Version == version {
			delete(want, pin.Source)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing workflow pins %v from %+v", want, res.Pins)
	}
}

func TestGoDetector_NoWorkflowFileOmitsCIPin(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir, "go 1.24")

	res, err := NewGoDetector().Detect(dir, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range res.Pins {
		if p.Source == "ci" || strings.Contains(p.Source, "workflows") {
			t.Fatalf("did not expect a ci pin, got %+v", res.Pins)
		}
	}
}

func TestGoDetector_ReportsMissingInstalledToolchain(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	dir := t.TempDir()
	writeGoMod(t, dir, "go 1.24")

	res, err := NewGoDetector().Detect(dir, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || !res.Drift {
		t.Fatalf("expected missing installed Go to cause drift, got %+v", res)
	}
	if got := res.Pins[len(res.Pins)-1]; got.Source != "installed" || got.Version != "not found" {
		t.Fatalf("installed pin = %+v, want installed/not found", got)
	}
}
