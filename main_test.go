package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var testGoVersionRe = regexp.MustCompile(`go(\d+\.\d+(?:\.\d+)?)`)

// installedGoVersionForTest shells out to `go version`, independent of the
// ecosystem package's unexported equivalent, so the no-drift wow-moment
// test can build a fixture that matches whatever toolchain runs the suite.
func installedGoVersionForTest() (string, error) {
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		return "", err
	}
	m := testGoVersionRe.FindStringSubmatch(string(out))
	if m == nil {
		return "", err
	}
	return m[1], nil
}

func writeFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRun_NoPinFiles_ExitsZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{t.TempDir()}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "no version pin files found") {
		t.Errorf("stdout = %q, want a message about no pin files", stdout.String())
	}
}

func TestRun_NoDrift_ExitsZero(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/x\n\ngo 1.22\n")

	var stdout, stderr bytes.Buffer
	code := run([]string{dir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestRun_Drift_ExitsOne(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/x\n\ngo 1.99\n")

	var stdout, stderr bytes.Buffer
	code := run([]string{dir}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1; stdout=%s", code, stdout.String())
	}
}

func TestRun_JSONAndTextExitCodesMatch(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/x\n\ngo 1.99\n")

	var textOut, jsonOut, stderr bytes.Buffer
	textCode := run([]string{dir}, &textOut, &stderr)
	jsonCode := run([]string{"--json", dir}, &jsonOut, &stderr)

	if textCode != jsonCode {
		t.Fatalf("text exit = %d, json exit = %d, want equal", textCode, jsonCode)
	}
	if !strings.Contains(jsonOut.String(), `"drift": true`) {
		t.Errorf("json output missing drift:true: %s", jsonOut.String())
	}
}

func TestRun_HelpExitsZero(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout.String(), "drift-check") {
		t.Errorf("expected usage text, got %q", stdout.String())
	}
}

// TestRun_WowMoment reproduces the headline demo from docs/BACKLOG.md: a
// go.mod, a CI workflow, and the installed toolchain each naming a
// different Go version, all three surfaced by name in one report.
func TestRun_WowMoment_ThreeWayGoDrift(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/monorepo\n\ngo 1.24\n")
	writeFile(t, dir, ".github/workflows/ci.yml", `
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

	var stdout, stderr bytes.Buffer
	code := run([]string{dir}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit code = %d, want 1 (drift found); stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"1.24", "1.23", "DRIFT"} {
		if !strings.Contains(out, want) {
			t.Errorf("report missing %q:\n%s", want, out)
		}
	}
}

// TestRun_WowMoment_NoDriftWhenAllAgree is the flip side of the wow-moment
// demo: when go.mod, CI, and the installed toolchain all agree, the report
// shows no drift and exits 0.
func TestRun_WowMoment_NoDriftWhenAllAgree(t *testing.T) {
	installed, err := installedGoVersionForTest()
	if err != nil {
		t.Skip("no go toolchain on PATH to compare against")
	}

	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/monorepo\n\ngo "+installed+"\n")
	writeFile(t, dir, ".github/workflows/ci.yml", `
jobs:
  test:
    steps:
      - uses: actions/setup-go@v5
        with:
          go-version: "`+installed+`"
`)

	var stdout, stderr bytes.Buffer
	code := run([]string{dir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%s", code, stdout.String())
	}
	if strings.Contains(stdout.String(), "DRIFT") {
		t.Errorf("expected no drift, got:\n%s", stdout.String())
	}
}
