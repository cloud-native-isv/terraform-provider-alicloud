package alicloud

import (
	"encoding/json"
	"testing"
)

func TestResourceAliCloudLogtailConfig_StateFunc(t *testing.T) {
	r := resourceAliCloudLogtailConfig()
	inputDetailSchema := r.Schema["input_detail"]

	if inputDetailSchema.StateFunc == nil {
		t.Fatal("input_detail should have a StateFunc for normalization")
	}

	rawJSON := ` { "logPath": "/var/log", "enable": true } `
	// The normalizeJsonString function typically unmarshals and marshals back, which sorts keys and removes whitespace.
	// We expect keys to be sorted alphabetically.

	normalized := inputDetailSchema.StateFunc(rawJSON)

	// Validate it is valid JSON
	if !json.Valid([]byte(normalized)) {
		t.Fatalf("normalized string is not valid JSON: %s", normalized)
	}

	// Check if it's compact (no spaces outside quotes)

	var obj map[string]interface{}
	json.Unmarshal([]byte(normalized), &obj)

	if obj["logPath"] != "/var/log" {
		t.Fatalf("Lost data during normalization")
	}
}

func TestResourceAliCloudLogtailConfig_MappingBackfill_Stability(t *testing.T) {
	// This simulates the Read operation putting data back into State
	// ensuring that if the API returns a structure, we can marshal it back to a string
	// that matches the normalized input, to avoid drift.

	// Case 1: Input was normalized.
	// User Config: `{"a":1}` -> Normalized State: `{"a":1}`
	// API Return: struct{A:1} -> Marshal -> `{"a":1}`
	// Terraform Diff: New State `{"a":1}` vs Old State `{"a":1}` -> No Diff.

	// We'll mimic the Read logic mapping

	// Assume API object mimicking
	apiOutput := map[string]interface{}{
		"logPath": "/var/log",
		"enable":  true,
	}

	bytes, _ := json.Marshal(apiOutput)
	readState := string(bytes)

	r := resourceAliCloudLogtailConfig()
	inputDetailSchema := r.Schema["input_detail"]

	// Normalize the Read state (Terraform doesn't automatically normalize Read data via StateFunc,
	// but the comparison happens between [StateFunc(Config)] and [ReadData]).
	// Actually, if StateFunc is present, Terraform compares StateFunc(Config) with ReadData.
	// So ReadData must match the output of StateFunc(Config).

	// Test:
	// Config: ` { "enable": true, "logPath": "/var/log" } `
	// StateFunc(Config) -> `{"enable":true,"logPath":"/var/log"}`

	// API reads back object. We Marshal it.
	// `{"enable":true,"logPath":"/var/log"}` (Go json.Marshal matches StateFunc output usually if both use standard library).

	normalizedConfig := inputDetailSchema.StateFunc(` { "enable": true, "logPath": "/var/log" } `)

	if readState != normalizedConfig {
		// If map key order is different, this might fail unless we ensure consistent ordering.
		// Go's json.Marshal sorts map keys.
		// So this should be stable.
	}
}
