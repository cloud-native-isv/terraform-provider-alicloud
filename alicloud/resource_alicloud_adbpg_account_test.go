package alicloud

import (
	"testing"
)

func TestAdbpgAccountSchemaAttributes(t *testing.T) {
	r := resourceAliCloudAdbpgAccount()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}

	requiredFields := []string{"db_instance_id", "account_name", "account_password"}
	for _, field := range requiredFields {
		s, ok := r.Schema[field]
		if !ok {
			t.Fatalf("expected %s field in schema", field)
		}
		if !s.Required {
			t.Fatalf("expected %s to be Required", field)
		}
	}

	forceNewFields := []string{"db_instance_id", "account_name", "account_type"}
	for _, field := range forceNewFields {
		s, ok := r.Schema[field]
		if !ok {
			t.Fatalf("expected %s field in schema", field)
		}
		if !s.ForceNew {
			t.Fatalf("expected %s to be ForceNew", field)
		}
	}

	s := r.Schema["account_password"]
	if !s.Sensitive {
		t.Fatal("expected account_password to be Sensitive")
	}
}

func TestAdbpgAccountImporterExists(t *testing.T) {
	r := resourceAliCloudAdbpgAccount()
	if r.Importer == nil {
		t.Fatal("expected Importer to be set for terraform import support")
	}
}

func TestAdbpgAccountCRUDFunctionsNotNil(t *testing.T) {
	r := resourceAliCloudAdbpgAccount()
	if r.Create == nil {
		t.Fatal("expected Create function")
	}
	if r.Read == nil {
		t.Fatal("expected Read function")
	}
	if r.Update == nil {
		t.Fatal("expected Update function")
	}
	if r.Delete == nil {
		t.Fatal("expected Delete function")
	}
}
