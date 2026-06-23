package alicloud

import (
	"testing"
)

func TestAdbpgConnectionSchemaAttributes(t *testing.T) {
	r := resourceAliCloudAdbpgConnection()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}

	requiredFields := []string{"db_instance_id", "connection_string_prefix"}
	for _, field := range requiredFields {
		s, ok := r.Schema[field]
		if !ok {
			t.Fatalf("expected %s field in schema", field)
		}
		if !s.Required {
			t.Fatalf("expected %s to be Required", field)
		}
		if !s.ForceNew {
			t.Fatalf("expected %s to be ForceNew", field)
		}
	}

	computedFields := []string{"connection_string", "ip_address", "port"}
	for _, field := range computedFields {
		s, ok := r.Schema[field]
		if !ok {
			t.Fatalf("expected %s field in schema", field)
		}
		if !s.Computed {
			t.Fatalf("expected %s to be Computed", field)
		}
	}
}

func TestAdbpgConnectionImporterExists(t *testing.T) {
	r := resourceAliCloudAdbpgConnection()
	if r.Importer == nil {
		t.Fatal("expected Importer to be set for terraform import support")
	}
}

func TestAdbpgConnectionNoUpdateFunction(t *testing.T) {
	r := resourceAliCloudAdbpgConnection()
	if r.Update != nil {
		t.Fatal("expected no Update function (allocate/release only)")
	}
}
