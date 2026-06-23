package alicloud

import (
	"testing"
)

func TestAdbpgZonesDataSourceSchema(t *testing.T) {
	r := dataSourceAliCloudAdbpgZones()
	if r == nil {
		t.Fatal("expected non-nil data source")
	}

	if _, ok := r.Schema["multi"]; !ok {
		t.Fatal("expected multi filter field")
	}

	zones, ok := r.Schema["zones"]
	if !ok {
		t.Fatal("expected zones output field")
	}
	if !zones.Computed {
		t.Fatal("expected zones to be Computed")
	}

	ids, ok := r.Schema["ids"]
	if !ok {
		t.Fatal("expected ids field")
	}
	if !ids.Computed {
		t.Fatal("expected ids to be Computed")
	}
}
