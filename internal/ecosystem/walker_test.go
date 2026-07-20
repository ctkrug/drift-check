package ecosystem

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFindProjectRoots_FindsNestedPinsAndSkipsGeneratedTrees(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{
		"services/api/go.mod",
		"packages/web/.nvmrc",
		"vendor/ignored/go.mod",
		"node_modules/ignored/.python-version",
		".git/ignored/.ruby-version",
	} {
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("pin"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := FindProjectRoots(root)
	if err != nil {
		t.Fatalf("FindProjectRoots() error = %v", err)
	}
	want := []string{
		filepath.Join(root, "packages", "web"),
		filepath.Join(root, "services", "api"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FindProjectRoots() = %v, want %v", got, want)
	}
}

func TestFindProjectRoots_IncludesRootOnlyOnce(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := FindProjectRoots(root)
	if err != nil {
		t.Fatalf("FindProjectRoots() error = %v", err)
	}
	if want := []string{root}; !reflect.DeepEqual(got, want) {
		t.Fatalf("FindProjectRoots() = %v, want %v", got, want)
	}
}
