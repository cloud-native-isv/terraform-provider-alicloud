package alicloud

import (
	"testing"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"
)

func TestFormatSelectedZonesReq(t *testing.T) {
	cases := []struct {
		name        string
		input       []interface{}
		expectError bool
		expectText  string
	}{
		{
			name: "valid input",
			input: []interface{}{
				[]interface{}{"zoneh", "zonef"},
				[]interface{}{"zonek"},
			},
			expectError: false,
			expectText:  `[["zoneh","zonef"],["zonek"]]`,
		},
		{
			name: "valid input single",
			input: []interface{}{
				[]interface{}{"zoneh"},
				[]interface{}{},
			},
			expectError: false,
			expectText:  `[["zoneh"],[]]`,
		},
		{
			name: "valid input with inner []string",
			input: []interface{}{
				[]string{"zoneh", "zonef"},
				[]string{"zonek"},
			},
			expectError: false,
			expectText:  `[["zoneh","zonef"],["zonek"]]`,
		},
		{
			name: "valid empty inner lists",
			input: []interface{}{
				[]interface{}{},
				[]interface{}{},
			},
			expectError: false,
			expectText:  `[[],[]]`,
		},
		{
			name:        "empty input returns empty string",
			input:       []interface{}{},
			expectError: false,
			expectText:  "",
		},
		{
			name:        "nil input returns empty string",
			input:       nil,
			expectError: false,
			expectText:  "",
		},
		{
			name: "invalid length (too many)",
			input: []interface{}{
				[]interface{}{"zone1"},
				[]interface{}{"zone2"},
				[]interface{}{"zone3"},
			},
			expectError: true,
		},
		{
			name: "invalid inner element type",
			input: []interface{}{
				[]interface{}{"zone1"},
				"not a list",
			},
			expectError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := formatSelectedZonesReq(tc.input)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if result != tc.expectText {
					t.Errorf("Expected text %s but got %s", tc.expectText, result)
				}
			}
		})
	}
}

func TestFormatSelectedZonesReqWithInvalidTopLevelType(t *testing.T) {
	_, err := formatSelectedZonesReq("invalid")
	if err == nil {
		t.Fatalf("Expected error but got nil")
	}
}

func TestIsJSONStringObjectSubset(t *testing.T) {
	cases := []struct {
		name     string
		subset   string
		superset string
		expect   bool
	}{
		{
			name:     "subset should be true",
			subset:   `{"a":"1","b":"2"}`,
			superset: `{"a":"1","b":"2","c":"3"}`,
			expect:   true,
		},
		{
			name:     "different value should be false",
			subset:   `{"a":"1"}`,
			superset: `{"a":"2","b":"2"}`,
			expect:   false,
		},
		{
			name:     "missing key should be false",
			subset:   `{"d":"1"}`,
			superset: `{"a":"2","b":"2"}`,
			expect:   false,
		},
		{
			name:     "invalid json should be false",
			subset:   `{a:1}`,
			superset: `{"a":"1"}`,
			expect:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := isJSONStringObjectSubset(tc.subset, tc.superset)
			if actual != tc.expect {
				t.Fatalf("expect %v, got %v", tc.expect, actual)
			}
		})
	}
}

func TestConfigReadPreserveStateForSubsetOrSuperset(t *testing.T) {
	cases := []struct {
		name   string
		state  string
		remote string
		expect bool
	}{
		{
			name:   "remote subset of state",
			state:  `{"a":"1","b":"2","c":"3"}`,
			remote: `{"a":"1","b":"2"}`,
			expect: true,
		},
		{
			name:   "remote superset of state",
			state:  `{"a":"1","b":"2"}`,
			remote: `{"a":"1","b":"2","c":"3"}`,
			expect: true,
		},
		{
			name:   "value changed should not preserve state",
			state:  `{"a":"1","b":"2"}`,
			remote: `{"a":"1","b":"9","c":"3"}`,
			expect: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			preserve := isJSONStringObjectSubset(tc.remote, tc.state) || isJSONStringObjectSubset(tc.state, tc.remote)
			if preserve != tc.expect {
				t.Fatalf("expect %v, got %v", tc.expect, preserve)
			}
		})
	}
}

func TestInferDeployModuleFromDeployType(t *testing.T) {
	eip := kafka.KafkaDeployType(4)
	vpc := kafka.KafkaDeployType(5)
	unknown := kafka.KafkaDeployType(9)

	tests := []struct {
		name       string
		deployType *kafka.KafkaDeployType
		expect     string
	}{
		{name: "nil", deployType: nil, expect: ""},
		{name: "eip", deployType: &eip, expect: "eip"},
		{name: "vpc", deployType: &vpc, expect: "vpc"},
		{name: "unknown", deployType: &unknown, expect: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := inferDeployModuleFromDeployType(tc.deployType)
			if got != tc.expect {
				t.Fatalf("expect %q, got %q", tc.expect, got)
			}
		})
	}
}

func TestInferSelectedZonesState(t *testing.T) {
	tests := []struct {
		name   string
		zoneId string
		expect [][]string
	}{
		{name: "empty zone", zoneId: "", expect: [][]string{{}, {}}},
		{name: "single zone", zoneId: "zoneb", expect: [][]string{{"zoneb"}, {}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := inferSelectedZonesState(tc.zoneId)
			if len(got) != len(tc.expect) || len(got[0]) != len(tc.expect[0]) || len(got[1]) != len(tc.expect[1]) {
				t.Fatalf("expect %v, got %v", tc.expect, got)
			}
			for i := range tc.expect {
				for j := range tc.expect[i] {
					if got[i][j] != tc.expect[i][j] {
						t.Fatalf("expect %v, got %v", tc.expect, got)
					}
				}
			}
		})
	}
}

func TestResolveServiceVersionForState(t *testing.T) {
	tests := []struct {
		name          string
		remoteVersion string
		stateVersion  string
		expect        string
	}{
		{name: "use remote when available", remoteVersion: "3.8.0", stateVersion: "2.6.2", expect: "3.8.0"},
		{name: "fallback to state when remote empty", remoteVersion: "", stateVersion: "2.6.2", expect: "2.6.2"},
		{name: "both empty", remoteVersion: "", stateVersion: "", expect: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveServiceVersionForState(tc.remoteVersion, tc.stateVersion)
			if got != tc.expect {
				t.Fatalf("expect %q, got %q", tc.expect, got)
			}
		})
	}
}
