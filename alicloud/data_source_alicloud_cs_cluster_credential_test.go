package alicloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestCsClusterCredentialSchemaContract(t *testing.T) {
	dataSource := dataSourceAliCloudCSClusterCredential()
	if err := dataSource.InternalValidate(nil, false); err != nil {
		t.Fatalf("InternalValidate() error = %v", err)
	}

	clusterId := dataSource.Schema["cluster_id"]
	if !clusterId.Required {
		t.Fatalf("cluster_id must be Required: %#v", clusterId)
	}
	if dataSource.Schema["temporary_duration_minutes"].Type != schema.TypeInt {
		t.Fatalf("temporary_duration_minutes must be TypeInt: %#v", dataSource.Schema["temporary_duration_minutes"])
	}
	if dataSource.Schema["private_ip_address"].Type != schema.TypeBool {
		t.Fatalf("private_ip_address must be TypeBool: %#v", dataSource.Schema["private_ip_address"])
	}
}

// TestCsClusterCredentialConfigIsSensitive pins the security requirement:
// the kubeconfig config attribute is a cluster access credential and must be
// marked Sensitive so it never leaks into plan/apply output.
func TestCsClusterCredentialConfigIsSensitive(t *testing.T) {
	config := dataSourceAliCloudCSClusterCredential().Schema["config"]
	if config == nil {
		t.Fatalf("config attribute missing from schema")
	}
	if !config.Computed {
		t.Fatalf("config must be Computed: %#v", config)
	}
	if !config.Sensitive {
		t.Fatalf("config must be Sensitive: %#v", config)
	}
}

func TestCsClusterCredentialBuildKubeConfigOptions(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		config                   map[string]interface{}
		wantPrivateIpAddress     bool
		wantTemporaryDurationMin int64
	}{
		{
			name:   "defaults",
			config: map[string]interface{}{"cluster_id": "c1a2b3"},
		},
		{
			name:                     "temporary duration only",
			config:                   map[string]interface{}{"cluster_id": "c1a2b3", "temporary_duration_minutes": 120},
			wantTemporaryDurationMin: 120,
		},
		{
			name:                 "private ip only",
			config:               map[string]interface{}{"cluster_id": "c1a2b3", "private_ip_address": true},
			wantPrivateIpAddress: true,
		},
		{
			name:                     "both options",
			config:                   map[string]interface{}{"cluster_id": "c1a2b3", "private_ip_address": true, "temporary_duration_minutes": 60},
			wantPrivateIpAddress:     true,
			wantTemporaryDurationMin: 60,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, dataSourceAliCloudCSClusterCredential().Schema, tc.config)
			got := buildAckKubeConfigOptions(d)
			if got.PrivateIpAddress != tc.wantPrivateIpAddress {
				t.Fatalf("PrivateIpAddress = %t, want %t", got.PrivateIpAddress, tc.wantPrivateIpAddress)
			}
			if got.TemporaryDurationMinutes != tc.wantTemporaryDurationMin {
				t.Fatalf("TemporaryDurationMinutes = %d, want %d", got.TemporaryDurationMinutes, tc.wantTemporaryDurationMin)
			}
		})
	}
}
