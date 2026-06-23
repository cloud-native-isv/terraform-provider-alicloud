package alicloud

import (
	"testing"
)

func TestAdbpgInstancesDataSourceSchema(t *testing.T) {
	r := dataSourceAliCloudAdbpgInstances()
	if r == nil {
		t.Fatal("expected non-nil data source")
	}

	filterFields := []string{"ids", "description_regex", "status", "resource_group_id", "tags", "output_file"}
	for _, field := range filterFields {
		if _, ok := r.Schema[field]; !ok {
			t.Fatalf("expected %s filter field in schema", field)
		}
	}

	instances, ok := r.Schema["instances"]
	if !ok {
		t.Fatal("expected instances output field")
	}
	if !instances.Computed {
		t.Fatal("expected instances to be Computed")
	}
}
