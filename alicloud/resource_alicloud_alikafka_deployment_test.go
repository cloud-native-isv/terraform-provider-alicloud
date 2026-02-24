package alicloud

import (
	"testing"
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
			name: "valid empty inner lists",
			input: []interface{}{
				[]interface{}{},
				[]interface{}{},
			},
			expectError: false,
			expectText:  `[[],[]]`,
		},
		{
			name:        "invalid length (empty)",
			input:       []interface{}{},
			expectError: true,
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
