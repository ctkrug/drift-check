package ecosystem

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// RubyDetector reconciles the version pinned in .ruby-version and
// Gemfile.lock's RUBY VERSION stanza against what's installed on PATH and
// what CI pins via ruby/setup-ruby. .ruby-version and Gemfile.lock are
// kept as distinct pins so a mismatch between the two is named directly,
// not conflated into one source.
type RubyDetector struct{}

func NewRubyDetector() *RubyDetector { return &RubyDetector{} }

func (d *RubyDetector) Name() string { return "Ruby" }

func (d *RubyDetector) Detect(root string) (*Result, error) {
	var pins []Pin

	rubyVersionFile := filepath.Join(root, ".ruby-version")
	data, err := os.ReadFile(rubyVersionFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if v := strings.TrimSpace(string(data)); v != "" {
		pins = append(pins, Pin{Source: ".ruby-version", Version: v})
	}

	gemfileVersion, err := parseGemfileLockRubyVersion(filepath.Join(root, "Gemfile.lock"))
	if err != nil {
		return nil, err
	}
	if gemfileVersion != "" {
		pins = append(pins, Pin{Source: "Gemfile.lock", Version: gemfileVersion})
	}

	if len(pins) == 0 {
		return nil, nil
	}

	if ciPins := findWorkflowPins(root, "ruby/setup-ruby", "ruby-version"); len(ciPins) > 0 {
		pins = append(pins, Pin{Source: ciPins[0].source, Version: ciPins[0].version})
	}

	installed, err := installedRubyVersion()
	pins, err = appendInstalledPin(pins, installed, err)
	if err != nil {
		return nil, err
	}

	res := &Result{Ecosystem: d.Name(), Pins: pins}
	res.Drift, res.Detail = reconcile(pins)
	return res, nil
}

var gemfileLockRubyVersionRe = regexp.MustCompile(`ruby\s+(\d+\.\d+\.\d+)`)

// parseGemfileLockRubyVersion extracts the version from Gemfile.lock's
// "RUBY VERSION" stanza (e.g. "ruby 3.3.0p0" -> "3.3.0"). Returns "" with
// no error when the file or stanza is absent.
func parseGemfileLockRubyVersion(path string) (string, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inStanza := false
	for scanner.Scan() {
		trimmed := strings.TrimSpace(scanner.Text())
		if trimmed == "RUBY VERSION" {
			inStanza = true
			continue
		}
		if !inStanza {
			continue
		}
		if trimmed == "" {
			break
		}
		if m := gemfileLockRubyVersionRe.FindStringSubmatch(trimmed); m != nil {
			return m[1], nil
		}
	}
	return "", scanner.Err()
}

var rubyVersionOutputRe = regexp.MustCompile(`ruby\s+(\d+\.\d+(?:\.\d+)?)`)

// installedRubyVersion shells out to `ruby -v` and extracts the version.
func installedRubyVersion() (string, error) {
	out, err := exec.Command("ruby", "-v").Output()
	if err != nil {
		return "", err
	}
	m := rubyVersionOutputRe.FindStringSubmatch(string(out))
	if m == nil {
		return "", nil
	}
	return m[1], nil
}
