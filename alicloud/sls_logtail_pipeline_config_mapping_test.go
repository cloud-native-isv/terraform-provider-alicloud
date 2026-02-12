package alicloud

import (
	"testing"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
)

func TestNormalizeLogtailConfigJson(t *testing.T) {
	cases := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{
			input:    `{"b": 2, "a": 1}`,
			expected: `{"a":1,"b":2}`,
			wantErr:  false,
		},
		{
			input:    `   { "x":  "y" }  `,
			expected: `{"x":"y"}`,
			wantErr:  false,
		},
		{
			input:    ``,
			expected: ``,
			wantErr:  false,
		},
		{
			input:    `invalid`,
			expected: `invalid`,
			wantErr:  true,
		},
	}

	for _, tc := range cases {
		out, err := NormalizeLogtailConfigJson(tc.input)
		if tc.wantErr && err == nil {
			t.Errorf("NormalizeLogtailConfigJson(%q) wanted error, got nil", tc.input)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("NormalizeLogtailConfigJson(%q) got error: %v", tc.input, err)
		}
		if !tc.wantErr && out != tc.expected {
			t.Errorf("NormalizeLogtailConfigJson(%q) = %q, want %q", tc.input, out, tc.expected)
		}
	}
}

func TestValidateLogtailConfigJsonObject(t *testing.T) {
	cases := []struct {
		input string
		valid bool
	}{
		{`{"a": 1}`, true},
		{`{}`, true},
		{`[]`, false},
		{`"string"`, false},
		{`123`, false},
		{``, true}, // Check implementation if empty string is valid
		{`invalid`, false},
	}

	for _, tc := range cases {
		err := ValidateLogtailConfigJsonObject(tc.input)
		if tc.valid && err != nil {
			t.Errorf("ValidateLogtailConfigJsonObject(%q) got error: %v", tc.input, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("ValidateLogtailConfigJsonObject(%q) wanted error, got nil", tc.input)
		}
	}
}

func TestSlsLogtailPipelineConfigMapping(t *testing.T) {
	// Test Domain -> Lib -> Domain roundtrip

	domainConfig := &SlsLogtailPipelineConfig{
		Project: "test-project",
		Name:    "test-config",
		Inputs: []SlsLogtailPipelineConfigPlugin{
			{Type: "input_file", ConfigJson: `{"logPath":"/var/log"}`},
		},
		Flushers: []SlsLogtailPipelineConfigPlugin{
			{Type: "flusher_sls", ConfigJson: `{"endpoint":"cn-hangzhou"}`},
		},
		GlobalJson: `{"k1":"v1"}`,
		TaskJson:   `{"k2":"v2"}`,
		LogSample:  "sample",
	}

	// To Lib
	var libConfig *sls.LogtailPipelineConfig
	var err error
	libConfig, err = domainConfig.ToLibConfig()
	if err != nil {
		t.Fatalf("ToLibConfig failed: %v", err)
	}

	if libConfig.ConfigName != "test-config" {
		t.Errorf("ConfigName = %s, want test-config", libConfig.ConfigName)
	}
	if len(libConfig.Inputs) != 1 {
		t.Fatalf("Inputs len = %d, want 1", len(libConfig.Inputs))
	}
	if libConfig.Inputs[0]["type"] != "input_file" {
		t.Errorf("Inputs[0].type = %s, want input_file", libConfig.Inputs[0]["type"])
	}
	if libConfig.Inputs[0]["logPath"] != "/var/log" {
		t.Errorf("Inputs[0].logPath = %s, want /var/log", libConfig.Inputs[0]["logPath"])
	}

	// From Lib
	// Note: We need to set up the lib structure as if it came from API (inputs map containing type)
	newDomainConfig := FromLibConfig(libConfig, "test-project")

	if newDomainConfig.Project != "test-project" {
		t.Errorf("Project = %s, want test-project", newDomainConfig.Project)
	}
	if newDomainConfig.Name != "test-config" {
		t.Errorf("Name = %s, want test-config", newDomainConfig.Name)
	}
	if len(newDomainConfig.Inputs) != 1 {
		t.Fatalf("New inputs len = %d, want 1", len(newDomainConfig.Inputs))
	}
	if newDomainConfig.Inputs[0].Type != "input_file" {
		t.Errorf("New inputs[0].Type = %s, want input_file", newDomainConfig.Inputs[0].Type)
	}

	// Expect config_json to NOT contain type because we strip it
	expectedJson := `{"logPath":"/var/log"}`
	normalized, _ := NormalizeLogtailConfigJson(newDomainConfig.Inputs[0].ConfigJson)
	if normalized != expectedJson {
		t.Errorf("New inputs[0].ConfigJson = %s, want %s", normalized, expectedJson)
	}
}

func TestSlsLogtailPipelineConfigPlugin_ToMap(t *testing.T) {
	// Case: ConfigJson overrides type - expecting code to enforce type from struct
	p := SlsLogtailPipelineConfigPlugin{
		Type:       "real_type",
		ConfigJson: `{"type":"fake_type", "k":"v"}`,
	}

	m, err := p.ToMap()
	if err != nil {
		t.Fatalf("ToMap failed: %v", err)
	}

	if m["type"] != "real_type" {
		t.Errorf("ToMap should prioritize struct type. Got %s, want real_type", m["type"])
	}
	if m["k"] != "v" {
		t.Errorf("ToMap lost config keys. Got %s", m["k"])
	}
}

func TestSlsLogtailPipelineConfigMapping_ProductionYamlShape(t *testing.T) {
	libCfg := &sls.LogtailPipelineConfig{
		ConfigName: "kangaroo-pai-file-fabricmanager-proxy",
		Inputs: []map[string]interface{}{
			{
				"Type":                           "input_file",
				"AllowingIncludedByMultiConfigs": true,
				"EnableContainerDiscovery":       false,
				"FileEncoding":                   "utf8",
				"FilePaths": []interface{}{
					"/logtail_host/var/log/fabricmanager-proxy/**/fabric.log",
				},
				"MaxDirSearchDepth": float64(10),
				"TailSizeKB":        float64(10485760),
			},
		},
		Processors: []map[string]interface{}{
			{
				"Type":      "processor_parse_json_native",
				"SourceKey": "content",
			},
		},
		Flushers: []map[string]interface{}{
			{
				"Type":          "flusher_sls",
				"Endpoint":      "cn-shanghai-b-intranet.log.aliyuncs.com",
				"Logstore":      "kangaroo-pai-fabricmanager-proxy",
				"Region":        "cn-shanghai-b",
				"TelemetryType": "logs",
			},
		},
		Aggregators:    []map[string]interface{}{},
		Global:         map[string]interface{}{"TopicType": "default"},
		Task:           map[string]interface{}{},
		LogSample:      "",
		CreateTime:     1770809720,
		LastModifyTime: 1770869527,
	}

	domain := FromLibConfig(libCfg, "test-project")
	if domain.Name != "kangaroo-pai-file-fabricmanager-proxy" {
		t.Fatalf("unexpected name: %s", domain.Name)
	}
	if len(domain.Inputs) != 1 || domain.Inputs[0].Type != "input_file" {
		t.Fatalf("unexpected input type: %+v", domain.Inputs)
	}
	if len(domain.Processors) != 1 || domain.Processors[0].Type != "processor_parse_json_native" {
		t.Fatalf("unexpected processor type: %+v", domain.Processors)
	}
	if len(domain.Flushers) != 1 || domain.Flushers[0].Type != "flusher_sls" {
		t.Fatalf("unexpected flusher type: %+v", domain.Flushers)
	}
	if len(domain.Aggregators) != 0 {
		t.Fatalf("aggregators should be empty")
	}
	if domain.GlobalJson != `{"TopicType":"default"}` {
		t.Fatalf("unexpected global_json: %s", domain.GlobalJson)
	}
	if domain.TaskJson != `{}` {
		t.Fatalf("unexpected task_json: %s", domain.TaskJson)
	}
	if domain.CreateTime != 1770809720 || domain.LastModifyTime != 1770869527 {
		t.Fatalf("unexpected timestamps: %d/%d", domain.CreateTime, domain.LastModifyTime)
	}

	// round-trip back to lib model
	newLib, err := domain.ToLibConfig()
	if err != nil {
		t.Fatalf("ToLibConfig failed: %v", err)
	}
	if newLib.Inputs[0]["type"] != "input_file" {
		t.Fatalf("unexpected round-trip input type: %v", newLib.Inputs[0]["type"])
	}
	if _, exists := newLib.Inputs[0]["Type"]; exists {
		t.Fatalf("round-trip should not keep uppercase Type key")
	}
}
