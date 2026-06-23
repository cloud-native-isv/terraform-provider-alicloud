package alicloud

import (
	"testing"
)

func TestAdbpgSecurityIpArraySchemaAttributes(t *testing.T) {
	r := resourceAliCloudAdbpgSecurityIpArray()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}

	s := r.Schema["db_instance_id"]
	if !s.Required || !s.ForceNew {
		t.Fatal("expected db_instance_id to be Required and ForceNew")
	}

	s = r.Schema["db_instance_ip_array_name"]
	if !s.Optional || !s.ForceNew {
		t.Fatal("expected db_instance_ip_array_name to be Optional and ForceNew")
	}
	if s.Default != "default" {
		t.Fatalf("expected default value 'default', got %v", s.Default)
	}

	s = r.Schema["security_ip_list"]
	if !s.Required {
		t.Fatal("expected security_ip_list to be Required")
	}
}

func TestAdbpgSecurityIpArrayImporterExists(t *testing.T) {
	r := resourceAliCloudAdbpgSecurityIpArray()
	if r.Importer == nil {
		t.Fatal("expected Importer to be set for terraform import support")
	}
}

func TestAdbpgSecurityIpArrayCRUDFunctionsNotNil(t *testing.T) {
	r := resourceAliCloudAdbpgSecurityIpArray()
	if r.Create == nil || r.Read == nil || r.Update == nil || r.Delete == nil {
		t.Fatal("expected all CRUD functions to be set")
	}
}
