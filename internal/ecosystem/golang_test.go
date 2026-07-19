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
	res, err := NewGoDetector().Detect(t.TempDir())
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

	res, err := NewGoDetector().Detect(dir)
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

	res, err := NewGoDetector().Detect(dir)
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

func TestGoDetector_NoWorkflowFileOmitsCIPin(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir, "go 1.24")

	res, err := NewGoDetector().Detect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, p := range res.Pins {
		if p.Source == "ci" || strings.Contains(p.Source, "workflows") {
			t.Fatalf("did not expect a ci pin, got %+v", res.Pins)
		}
	}
}

func TestVersionsAgree(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"1.24", "1.24.3", true},
		{"1.24.3", "1.24", true},
		{"1.24", "1.24", true},
		{"1.24", "1.23", false},
		{"1.2", "1.23", false},
		{"1.2.9", "1.2.9", true},
	}
	for _, c := range cases {
		if got := versionsAgree(c.a, c.b); got != c.want {
			t.Errorf("versionsAgree(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestReconcile_DetectsDrift(t *testing.T) {
	pins := []Pin{
		{Source: "go.mod", Version: "1.24"},
		{Source: "installed", Version: "1.22.2"},
	}
	drift, detail := reconcile(pins)
	if !drift {
		t.Fatal("expected drift to be detected")
	}
	if detail == "" {
		t.Fatal("expected a non-empty detail explaining the drift")
	}
}

func TestReconcile_ThreeWayDriftNamesEverySource(t *testing.T) {
	pins := []Pin{
		{Source: "go.mod", Version: "1.24"},
		{Source: "ci", Version: "1.23"},
		{Source: "installed", Version: "1.22"},
	}
	drift, detail := reconcile(pins)
	if !drift {
		t.Fatal("expected drift to be detected")
	}
	for _, want := range []string{"go.mod says 1.24", "ci says 1.23", "installed says 1.22"} {
		if !strings.Contains(detail, want) {
			t.Errorf("detail %q missing %q", detail, want)
		}
	}
}

func TestReconcile_NoDriftWhenCompatible(t *testing.T) {
	pins := []Pin{
		{Source: "go.mod", Version: "1.24"},
		{Source: "installed", Version: "1.24.3"},
	}
	drift, _ := reconcile(pins)
	if drift {
		t.Fatal("expected no drift for compatible versions")
	}
}
