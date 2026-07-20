package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCompiledBinary_FixtureMonorepoMatchesGoldenReport(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "drift-check")
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build binary: %v\n%s", err, out)
	}

	tools := t.TempDir()
	writeToolchain(t, tools, "go", "go version go1.24.3 linux/amd64")
	writeToolchain(t, tools, "node", "v20.11.0")
	writeToolchain(t, tools, "python3", "Python 3.12.1")
	writeToolchain(t, tools, "ruby", "ruby 3.3.0p0")
	t.Setenv("PATH", tools)

	fixture := filepath.Join("testdata", "monorepo")
	cmd := exec.Command(bin, fixture)
	out, err := cmd.CombinedOutput()
	if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 1 {
		t.Fatalf("exit error = %v, want code 1; output:\n%s", err, out)
	}

	want, err := os.ReadFile(filepath.Join(fixture, "report.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != string(want) {
		t.Errorf("compiled report mismatch (-want +got):\n--- want\n%s--- got\n%s", want, out)
	}
}

func writeToolchain(t *testing.T, dir, name, output string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho '"+output+"'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}
