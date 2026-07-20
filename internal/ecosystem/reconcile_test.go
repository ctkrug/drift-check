package ecosystem

import (
	"strings"
	"testing"
)

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
	for _, test := range cases {
		if got := versionsAgree(test.a, test.b); got != test.want {
			t.Errorf("versionsAgree(%q, %q) = %v, want %v", test.a, test.b, got, test.want)
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

func TestReconcile_CompareEveryPinPair(t *testing.T) {
	pins := []Pin{
		{Source: "go.mod", Version: "1.24"},
		{Source: "ci", Version: "1.24.1"},
		{Source: "installed", Version: "1.24.2"},
	}

	drift, detail := reconcile(pins)
	if !drift {
		t.Fatal("expected drift between the two exact patch versions")
	}
	for _, want := range []string{"ci says 1.24.1", "installed says 1.24.2"} {
		if !strings.Contains(detail, want) {
			t.Errorf("detail %q missing %q", detail, want)
		}
	}
}
