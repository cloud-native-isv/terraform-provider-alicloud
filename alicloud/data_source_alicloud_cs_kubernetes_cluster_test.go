package alicloud

import (
	"reflect"
	"testing"

	aliyunAckAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/ack"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestCsClusterDataSourceSchemaContract(t *testing.T) {
	dataSource := dataSourceAliCloudCSKubernetesCluster()
	if err := dataSource.InternalValidate(nil, false); err != nil {
		t.Fatalf("InternalValidate() error = %v", err)
	}

	clusterId := dataSource.Schema["cluster_id"]
	if !clusterId.Required || clusterId.Computed || clusterId.Optional {
		t.Fatalf("cluster_id must be Required-only: %#v", clusterId)
	}
	for _, name := range []string{
		"name", "state", "cluster_spec", "cluster_type", "profile", "version",
		"region_id", "zone_id", "master_url", "vpc_id", "vswitch_ids",
		"container_cidr", "service_cidr", "subnet_cidr", "security_group_id",
		"resource_group_id", "deletion_protection", "node_count",
		"worker_ram_role_name", "timezone", "tags", "maintenance_window",
	} {
		field, ok := dataSource.Schema[name]
		if !ok || !field.Computed {
			t.Fatalf("field %q must be Computed: %#v", name, field)
		}
	}

	window := dataSource.Schema["maintenance_window"]
	if window.Type != schema.TypeList {
		t.Fatalf("maintenance_window must be a list: %#v", window)
	}
	windowFields := window.Elem.(*schema.Resource).Schema
	for _, name := range []string{"enable", "start_time", "duration", "weekly_period"} {
		if _, ok := windowFields[name]; !ok {
			t.Fatalf("maintenance_window block missing field %q", name)
		}
	}
}

func TestCsClusterFlattenMaintenanceWindow(t *testing.T) {
	for _, tc := range []struct {
		name   string
		window *aliyunAckAPI.AckMaintenanceWindow
		want   []interface{}
	}{
		{
			name:   "nil window flattens to empty list",
			window: nil,
			want:   []interface{}{},
		},
		{
			name: "full window",
			window: &aliyunAckAPI.AckMaintenanceWindow{
				Enable:       true,
				StartTime:    "03:00:00Z",
				Duration:     "3h",
				WeeklyPeriod: []string{"Monday", "Tuesday"},
			},
			want: []interface{}{map[string]interface{}{
				"enable":        true,
				"start_time":    "03:00:00Z",
				"duration":      "3h",
				"weekly_period": []interface{}{"Monday", "Tuesday"},
			}},
		},
		{
			name: "window without weekly period",
			window: &aliyunAckAPI.AckMaintenanceWindow{
				Enable:    false,
				StartTime: "03:00:00Z",
				Duration:  "3h",
			},
			want: []interface{}{map[string]interface{}{
				"enable":        false,
				"start_time":    "03:00:00Z",
				"duration":      "3h",
				"weekly_period": []interface{}{},
			}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := flattenAckMaintenanceWindow(tc.window)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("flattenAckMaintenanceWindow() = %#v, want %#v", got, tc.want)
			}
		})
	}
}
