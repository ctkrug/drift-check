package report

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ctkrug/drift-check/internal/ecosystem"
)

func TestWriteJSON_ValidAndContainsExpectedFields(t *testing.T) {
	results := []*ecosystem.Result{
		{
			Ecosystem: "Go",
			Pins: []ecosystem.Pin{
				{Source: "go.mod", Version: "1.24"},
				{Source: "installed", Version: "1.22"},
			},
			Drift:  true,
			Detail: "go.mod says 1.24, installed says 1.22",
		},
	}

	var buf bytes.Buffer
	drifted := WriteJSON(&buf, results)
	if drifted != 1 {
		t.Fatalf("drifted = %d, want 1", drifted)
	}

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if decoded["drift"] != true {
		t.Errorf("top-level drift = %v, want true", decoded["drift"])
	}
	ecosystems, ok := decoded["ecosystems"].([]any)
	if !ok || len(ecosystems) != 1 {
		t.Fatalf("expected 1 ecosystem entry, got %v", decoded["ecosystems"])
	}
	entry := ecosystems[0].(map[string]any)
	if entry["ecosystem"] != "Go" {
		t.Errorf("ecosystem = %v, want Go", entry["ecosystem"])
	}
	pins, ok := entry["pins"].([]any)
	if !ok || len(pins) != 2 {
		t.Fatalf("expected 2 pins, got %v", entry["pins"])
	}
}

func TestWriteJSON_NoDriftReportsFalse(t *testing.T) {
	results := []*ecosystem.Result{
		{Ecosystem: "Go", Pins: []ecosystem.Pin{{Source: "go.mod", Version: "1.24"}}, Drift: false},
	}

	var buf bytes.Buffer
	drifted := WriteJSON(&buf, results)
	if drifted != 0 {
		t.Fatalf("drifted = %d, want 0", drifted)
	}

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if decoded["drift"] != false {
		t.Errorf("top-level drift = %v, want false", decoded["drift"])
	}
}

func TestWriteJSON_SkipsNilResults(t *testing.T) {
	results := []*ecosystem.Result{nil, {Ecosystem: "Go", Pins: []ecosystem.Pin{{Source: "go.mod", Version: "1.24"}}}}

	var buf bytes.Buffer
	WriteJSON(&buf, results)

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	ecosystems := decoded["ecosystems"].([]any)
	if len(ecosystems) != 1 {
		t.Fatalf("expected nil result skipped, got %d entries", len(ecosystems))
	}
}

func TestWriteJSONEmpty_ValidWithMessage(t *testing.T) {
	var buf bytes.Buffer
	WriteJSONEmpty(&buf)

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if decoded["message"] != "no version pin files found" {
		t.Errorf("message = %v, want the no-pin-files message", decoded["message"])
	}
	if decoded["drift"] != false {
		t.Errorf("drift = %v, want false", decoded["drift"])
	}
	ecosystems, ok := decoded["ecosystems"].([]any)
	if !ok || len(ecosystems) != 0 {
		t.Fatalf("expected empty ecosystems array, got %v", decoded["ecosystems"])
	}
}
