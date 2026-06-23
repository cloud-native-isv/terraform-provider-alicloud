package alicloud

import (
	"testing"
)

func TestAdbpgSslSchemaAttributes(t *testing.T) {
	r := resourceAliCloudAdbpgSsl()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}

	s := r.Schema["db_instance_id"]
	if !s.Required || !s.ForceNew {
		t.Fatal("expected db_instance_id to be Required and ForceNew")
	}

	s = r.Schema["ssl_enabled"]
	if !s.Required {
		t.Fatal("expected ssl_enabled to be Required")
	}

	s = r.Schema["ssl_expired"]
	if !s.Computed {
		t.Fatal("expected ssl_expired to be Computed")
	}
}

func TestAdbpgSslImporterExists(t *testing.T) {
	r := resourceAliCloudAdbpgSsl()
	if r.Importer == nil {
		t.Fatal("expected Importer to be set for terraform import support")
	}
}

func TestAdbpgSslCRUDFunctionsNotNil(t *testing.T) {
	r := resourceAliCloudAdbpgSsl()
	if r.Create == nil || r.Read == nil || r.Update == nil || r.Delete == nil {
		t.Fatal("expected all CRUD functions to be set")
	}
}
