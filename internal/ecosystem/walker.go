package ecosystem

import (
	"io/fs"
	"path/filepath"
	"sort"
)

var pinFileNames = map[string]struct{}{
	"go.mod":          {},
	".nvmrc":          {},
	".python-version": {},
	".ruby-version":   {},
	"Gemfile.lock":    {},
}

var skippedDirectoryNames = map[string]struct{}{
	".git":         {},
	"node_modules": {},
	"vendor":       {},
}

// FindProjectRoots returns every directory below root containing a supported
// pin file. Dependency and VCS trees are skipped because their pins belong to
// checked-in dependencies rather than to the repository being audited.
func FindProjectRoots(root string) ([]string, error) {
	roots := make(map[string]struct{})
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != root {
				if _, skip := skippedDirectoryNames[entry.Name()]; skip {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if _, ok := pinFileNames[entry.Name()]; ok {
			roots[filepath.Dir(path)] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	found := make([]string, 0, len(roots))
	for path := range roots {
		found = append(found, path)
	}
	// filepath.WalkDir is lexical, but collecting in a map loses that order.
	// Sorting preserves stable report and JSON output across runs.
	sort.Strings(found)
	return found, nil
}
