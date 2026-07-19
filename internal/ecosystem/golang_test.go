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
