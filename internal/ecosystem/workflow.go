package ecosystem

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// workflowPin is a version pin extracted from one GitHub Actions workflow
// file's use of a "setup-<ecosystem>" action.
type workflowPin struct {
	source  string // relative path to the workflow file, e.g. ".github/workflows/ci.yml"
	version string
}

// findWorkflowPins scans every .github/workflows/*.yml(.yaml) file under
// root for steps that use actionPrefix (e.g. "actions/setup-go") and
// extracts the value of inputKey (e.g. "go-version") from that step's
// `with:` block. A workflow file that can't be parsed is skipped with a
// warning to stderr rather than failing the whole scan — one broken
// workflow shouldn't hide pins found in the others.
func findWorkflowPins(root, actionPrefix, inputKey string) []workflowPin {
	dir := filepath.Join(root, ".github", "workflows")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var pins []workflowPin
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			continue
		}
		path := filepath.Join(dir, name)
		version, err := parseWorkflowFile(path, actionPrefix, inputKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "drift-check: warning: skipping %s: %v\n", path, err)
			continue
		}
		if version == "" {
			continue
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}
		pins = append(pins, workflowPin{source: filepath.ToSlash(rel), version: version})
	}
	return pins
}

// parseWorkflowFile is a narrow, line-oriented scanner for the specific
// shape a GitHub Actions step takes — not a general YAML parser. It looks
// for a "uses: <actionPrefix>..." line, then the sibling "with:" block that
// follows it, and returns the value of inputKey from within that block.
func parseWorkflowFile(path, actionPrefix, inputKey string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	stepActive := false
	stepIndent := 0
	withActive := false
	withIndent := 0

	for scanner.Scan() {
		raw := scanner.Text()
		if strings.TrimSpace(raw) == "" || strings.HasPrefix(strings.TrimSpace(raw), "#") {
			continue
		}
		indent, rest := stripListMarker(raw)

		if withActive && indent <= withIndent {
			withActive = false
		}
		if stepActive && indent < stepIndent {
			stepActive = false
		}

		if key, value, ok := splitKV(rest); ok && key == "uses" {
			stepActive = isActionReference(value, actionPrefix)
			stepIndent = indent
			withActive = false
			continue
		}

		if stepActive && indent == stepIndent && strings.TrimSpace(rest) == "with:" {
			withActive = true
			withIndent = indent
			continue
		}

		if withActive {
			if key, value, ok := splitKV(rest); ok && key == inputKey {
				return value, nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", nil
}

// isActionReference accepts a setup action only when actionPrefix is the
// complete action name followed by its required @ref delimiter. A raw prefix
// would incorrectly match unrelated actions such as actions/setup-goose.
func isActionReference(value, actionPrefix string) bool {
	return strings.HasPrefix(value, actionPrefix) &&
		len(value) > len(actionPrefix) && value[len(actionPrefix)] == '@'
}

// stripListMarker returns the effective indent of a YAML line's content and
// the content itself, treating a leading "- " list marker as part of the
// indent (so "  - uses: x" and "    uses: x" report the same content
// indent for "uses:").
func stripListMarker(line string) (indent int, rest string) {
	i := 0
	for i < len(line) && line[i] == ' ' {
		i++
	}
	rest = line[i:]
	indent = i
	if strings.HasPrefix(rest, "- ") {
		j := 1
		for j < len(rest) && rest[j] == ' ' {
			j++
		}
		indent += j
		rest = rest[j:]
	}
	return indent, rest
}

// splitKV splits a "key: value" line, stripping surrounding quotes and
// trailing comments from the value.
func splitKV(line string) (key, value string, ok bool) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:idx])
	if key == "" {
		return "", "", false
	}
	value = unquoteValue(strings.TrimSpace(line[idx+1:]))
	return key, value, true
}

// unquoteValue strips surrounding quotes or a trailing "# comment" from a
// scalar YAML value.
func unquoteValue(s string) string {
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') {
		if idx := strings.IndexByte(s[1:], s[0]); idx >= 0 {
			return s[1 : idx+1]
		}
	}
	if idx := strings.Index(s, " #"); idx >= 0 {
		s = strings.TrimSpace(s[:idx])
	}
	return s
}
