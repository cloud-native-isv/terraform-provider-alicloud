package alicloud

import (
	"reflect"
	"testing"
	"time"

	aliyunAckAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/ack"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestCsKubernetesAddonSchemaContract(t *testing.T) {
	resource := resourceAliCloudCSKubernetesAddon()
	if err := resource.InternalValidate(nil, true); err != nil {
		t.Fatalf("InternalValidate() error = %v", err)
	}

	for _, name := range []string{"cluster_id", "name"} {
		field, ok := resource.Schema[name]
		if !ok || !field.Required || !field.ForceNew {
			t.Fatalf("field %q must be Required and ForceNew: %#v", name, field)
		}
	}
	for _, name := range []string{"version", "config", "state"} {
		if _, ok := resource.Schema[name]; !ok {
			t.Fatalf("field %q missing from schema", name)
		}
	}
	if !resource.Schema["version"].Computed || !resource.Schema["version"].Optional {
		t.Fatalf("version must be Optional+Computed: %#v", resource.Schema["version"])
	}
	if !resource.Schema["state"].Computed {
		t.Fatalf("state must be Computed: %#v", resource.Schema["state"])
	}
	if resource.Timeouts == nil || resource.Timeouts.Create == nil || *resource.Timeouts.Create != 10*time.Minute {
		t.Fatalf("create timeout must be 10m: %#v", resource.Timeouts)
	}
	// Uninstall always cleans up cloud resources owned by the addon.
	if !ackAddonCleanupCloudResources {
		t.Fatalf("ackAddonCleanupCloudResources must be true")
	}
}

func TestCsKubernetesAddonBuildInstallOptions(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config map[string]interface{}
		want   *aliyunAckAPI.AckAddonInstallOptions
	}{
		{
			name: "name only, version and config default to empty",
			config: map[string]interface{}{
				"cluster_id": "c1a2b3",
				"name":       "terway-eniip",
			},
			want: &aliyunAckAPI.AckAddonInstallOptions{Name: "terway-eniip"},
		},
		{
			name: "pinned version with config",
			config: map[string]interface{}{
				"cluster_id": "c1a2b3",
				"name":       "flannel",
				"version":    "v1.2.4",
				"config":     `{"PodVswitchIds":["vsw-1"]}`,
			},
			want: &aliyunAckAPI.AckAddonInstallOptions{
				Name:    "flannel",
				Version: "v1.2.4",
				Config:  `{"PodVswitchIds":["vsw-1"]}`,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceAliCloudCSKubernetesAddon().Schema, tc.config)
			got := buildAckAddonInstallOptions(d)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("buildAckAddonInstallOptions() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

// TestCsKubernetesAddonCompositeIdFormat pins the id convention
// "clusterId:addonName" used by Create and parsed back by Read/Delete.
func TestCsKubernetesAddonCompositeIdFormat(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAliCloudCSKubernetesAddon().Schema, map[string]interface{}{
		"cluster_id": "c1a2b3",
		"name":       "terway-eniip",
	})
	d.SetId("c1a2b3:terway-eniip")

	clusterId, name, err := ackParseTwoPartId(d.Id())
	if err != nil {
		t.Fatalf("ackParseTwoPartId(%q) error = %v", d.Id(), err)
	}
	if clusterId != "c1a2b3" || name != "terway-eniip" {
		t.Fatalf("ackParseTwoPartId(%q) = (%q, %q)", d.Id(), clusterId, name)
	}
}
