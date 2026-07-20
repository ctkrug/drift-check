package ecosystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeRubyVersion(t *testing.T, dir, version string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".ruby-version"), []byte(version), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeGemfileLock(t *testing.T, dir, rubyVersion string) {
	t.Helper()
	content := "GEM\n  remote: https://rubygems.org/\n\nRUBY VERSION\n   ruby " + rubyVersion + "\n\nBUNDLED WITH\n   2.5.3\n"
	if err := os.WriteFile(filepath.Join(dir, "Gemfile.lock"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRubyDetector_NoPinFiles(t *testing.T) {
	dir := t.TempDir()
	res, err := NewRubyDetector().Detect(dir, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil result, got %+v", res)
	}
}

func TestRubyDetector_RubyVersionAndGemfileLockConflictNamedSeparately(t *testing.T) {
	dir := t.TempDir()
	writeRubyVersion(t, dir, "3.2.0")
	writeGemfileLock(t, dir, "3.3.0p0")

	res, err := NewRubyDetector().Detect(dir, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Pins) < 2 {
		t.Fatalf("expected both .ruby-version and Gemfile.lock pins, got %+v", res.Pins)
	}
	if res.Pins[0].Source != ".ruby-version" || res.Pins[1].Source != "Gemfile.lock" {
		t.Fatalf("expected distinct sources, got %+v", res.Pins)
	}
	if res.Pins[1].Version != "3.3.0" {
		t.Fatalf("expected Gemfile.lock patchlevel stripped to 3.3.0, got %q", res.Pins[1].Version)
	}
	if !res.Drift {
		t.Fatal("expected drift between .ruby-version and Gemfile.lock")
	}
	if !strings.Contains(res.Detail, ".ruby-version says 3.2.0") || !strings.Contains(res.Detail, "Gemfile.lock says 3.3.0") {
		t.Errorf("detail %q doesn't name both sources", res.Detail)
	}
}

func TestRubyDetector_OnlyGemfileLockStillReconciles(t *testing.T) {
	dir := t.TempDir()
	writeGemfileLock(t, dir, "3.3.0")
	writeWorkflow(t, dir, "ci.yml", `
jobs:
  test:
    steps:
      - uses: ruby/setup-ruby@v1
        with:
          ruby-version: "3.1.0"
`)

	res, err := NewRubyDetector().Detect(dir, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Fatal("expected a result from Gemfile.lock alone")
	}
	if res.Pins[0].Source != "Gemfile.lock" {
		t.Fatalf("expected Gemfile.lock as the only file pin, got %+v", res.Pins)
	}
	if !res.Drift {
		t.Fatal("expected drift against the CI pin")
	}
}

func TestRubyDetector_CIPinParsed(t *testing.T) {
	dir := t.TempDir()
	writeRubyVersion(t, dir, "3.3.0")
	writeWorkflow(t, dir, "ci.yml", `
jobs:
  test:
    steps:
      - uses: ruby/setup-ruby@v1
        with:
          ruby-version: "3.3.0"
`)

	res, err := NewRubyDetector().Detect(dir, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, p := range res.Pins {
		if strings.Contains(p.Source, "workflows") {
			found = true
			if p.Version != "3.3.0" {
				t.Errorf("ci pin version = %q, want 3.3.0", p.Version)
			}
		}
	}
	if !found {
		t.Fatalf("expected a ci pin among %+v", res.Pins)
	}
}

func TestParseGemfileLockRubyVersion_NoFile(t *testing.T) {
	v, err := parseGemfileLockRubyVersion(filepath.Join(t.TempDir(), "Gemfile.lock"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "" {
		t.Fatalf("expected empty version, got %q", v)
	}
}

func TestParseGemfileLockRubyVersion_NoStanza(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Gemfile.lock")
	if err := os.WriteFile(path, []byte("GEM\n  remote: https://rubygems.org/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	v, err := parseGemfileLockRubyVersion(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "" {
		t.Fatalf("expected empty version when no RUBY VERSION stanza, got %q", v)
	}
}

func TestRubyDetector_ReportsMissingInstalledToolchain(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	dir := t.TempDir()
	writeRubyVersion(t, dir, "3.3.0")

	res, err := NewRubyDetector().Detect(dir, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || !res.Drift {
		t.Fatalf("expected missing installed Ruby to cause drift, got %+v", res)
	}
	if got := res.Pins[len(res.Pins)-1]; got.Source != "installed" || got.Version != "not found" {
		t.Fatalf("installed pin = %+v, want installed/not found", got)
	}
}
