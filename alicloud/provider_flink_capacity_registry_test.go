package alicloud

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestFlinkCapacityAllocationProviderRegistry(t *testing.T) {
	provider := Provider().(*schema.Provider)

	if resource := provider.ResourcesMap["alicloud_flink_workspace_capacity_allocation"]; resource == nil {
		t.Fatal("alicloud_flink_workspace_capacity_allocation is not registered")
	}
	if _, exists := provider.ResourcesMap["alicloud_flink_capacity_coordinator"]; exists {
		t.Fatal("alicloud_flink_capacity_coordinator must not remain registered")
	}

	for name, resourceName := range map[string]string{
		"workspace":         "alicloud_flink_workspace",
		"namespace":         "alicloud_flink_namespace",
		"deployment target": "alicloud_flink_deployment_target",
	} {
		resource := provider.ResourcesMap[resourceName]
		if resource == nil {
			t.Fatalf("%s resource %q is not registered", name, resourceName)
		}
		for _, field := range []string{"capacity_management", "bootstrap_capacity"} {
			if _, exists := resource.Schema[field]; exists {
				t.Fatalf("%s still exposes %s", name, field)
			}
		}
	}
}

func TestFlinkProviderResourceNamesRemainAuthoritative(t *testing.T) {
	provider := Provider().(*schema.Provider)
	want := map[string]struct{}{
		"alicloud_flink_connector":                     {},
		"alicloud_flink_deployment":                    {},
		"alicloud_flink_deployment_draft":              {},
		"alicloud_flink_deployment_folder":             {},
		"alicloud_flink_deployment_target":             {},
		"alicloud_flink_job":                           {},
		"alicloud_flink_member":                        {},
		"alicloud_flink_namespace":                     {},
		"alicloud_flink_session_cluster":               {},
		"alicloud_flink_udf":                           {},
		"alicloud_flink_variable":                      {},
		"alicloud_flink_workspace":                     {},
		"alicloud_flink_workspace_capacity_allocation": {},
	}
	got := make(map[string]struct{}, len(want))
	for name := range provider.ResourcesMap {
		if strings.HasPrefix(name, "alicloud_flink_") {
			got[name] = struct{}{}
		}
	}
	for name := range want {
		if _, exists := got[name]; !exists {
			t.Errorf("missing Flink resource %q", name)
		}
	}
	for name := range got {
		if _, exists := want[name]; !exists {
			t.Errorf("unexpected Flink resource %q", name)
		}
	}
}
