package alicloud

import (
	"testing"
)

func TestAdbpgBackupPolicySchemaAttributes(t *testing.T) {
	r := resourceAliCloudAdbpgBackupPolicy()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}

	s := r.Schema["db_instance_id"]
	if !s.Required || !s.ForceNew {
		t.Fatal("expected db_instance_id to be Required and ForceNew")
	}

	optionalComputed := []string{"backup_retention_period", "preferred_backup_period", "preferred_backup_time", "enable_recovery_point"}
	for _, field := range optionalComputed {
		s, ok := r.Schema[field]
		if !ok {
			t.Fatalf("expected %s field in schema", field)
		}
		if !s.Optional || !s.Computed {
			t.Fatalf("expected %s to be Optional and Computed", field)
		}
	}
}

func TestAdbpgBackupPolicyImporterExists(t *testing.T) {
	r := resourceAliCloudAdbpgBackupPolicy()
	if r.Importer == nil {
		t.Fatal("expected Importer to be set for terraform import support")
	}
}

func TestAdbpgBackupPolicyCRUDFunctionsNotNil(t *testing.T) {
	r := resourceAliCloudAdbpgBackupPolicy()
	if r.Create == nil || r.Read == nil || r.Update == nil || r.Delete == nil {
		t.Fatal("expected all CRUD functions to be set")
	}
}
