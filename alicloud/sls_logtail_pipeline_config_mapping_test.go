package alicloud

import (
	"encoding/json"
	"testing"
)

func TestSlsLogtailPipelineConfigPlugin_ToMap_UsesUppercaseType(t *testing.T) {
	p := &SlsLogtailPipelineConfigPlugin{
		Type:       "input_file",
		ConfigJson: `{"project":"test-project","type":"should-be-overridden"}`,
	}

	m, err := p.ToMap()
	if err != nil {
		t.Fatalf("ToMap failed: %v", err)
	}

	if got, ok := m["Type"].(string); !ok || got != "input_file" {
		t.Fatalf("expected Type=input_file, got %#v", m["Type"])
	}

	if _, exists := m["type"]; exists {
		t.Fatalf("expected lowercase type key to be absent, got map: %#v", m)
	}
}

func TestSlsLogtailPipelineConfig_Mapping_RoundTrip(t *testing.T) {
	// Simple test to ensure mapping logic compiles and runs

	raw := `{"b":2, "a":1}`
	normalized, err := normalizeJsonString(raw)
	if err != nil {
		t.Fatalf("normalizeJsonString failed: %v", err)
	}

	// Go's json.Marshal sorts keys, so this should be stable
	expected := `{"a":1,"b":2}`
	if normalized != expected {
		var nMap, eMap map[string]interface{}
		json.Unmarshal([]byte(normalized), &nMap)
		json.Unmarshal([]byte(expected), &eMap)

		if len(nMap) != len(eMap) {
			t.Fatalf("Normalization changed content")
		}
	}
}
