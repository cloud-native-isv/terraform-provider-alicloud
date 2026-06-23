package alicloud

import (
	"testing"
)

func TestAdbpgDatabaseSchemaAttributes(t *testing.T) {
	r := resourceAliCloudAdbpgDatabase()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}

	requiredFields := []string{"db_instance_id", "db_name"}
	for _, field := range requiredFields {
		s, ok := r.Schema[field]
		if !ok {
			t.Fatalf("expected %s field in schema", field)
		}
		if !s.Required {
			t.Fatalf("expected %s to be Required", field)
		}
	}

	allForceNew := []string{"db_instance_id", "db_name", "db_description", "character_name"}
	for _, field := range allForceNew {
		s, ok := r.Schema[field]
		if !ok {
			t.Fatalf("expected %s field in schema", field)
		}
		if !s.ForceNew {
			t.Fatalf("expected %s to be ForceNew (no update API)", field)
		}
	}
}

func TestAdbpgDatabaseImporterExists(t *testing.T) {
	r := resourceAliCloudAdbpgDatabase()
	if r.Importer == nil {
		t.Fatal("expected Importer to be set for terraform import support")
	}
}

func TestAdbpgDatabaseNoUpdateFunction(t *testing.T) {
	r := resourceAliCloudAdbpgDatabase()
	if r.Update != nil {
		t.Fatal("expected no Update function (all fields ForceNew)")
	}
}
