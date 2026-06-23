package alicloud

import (
	"testing"
)

func TestAdbpgInstanceSchemaAttributes(t *testing.T) {
	r := resourceAliCloudAdbpgInstance()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}

	requiredFields := []string{"engine_version", "zone_id", "db_instance_class", "seg_node_num"}
	for _, field := range requiredFields {
		s, ok := r.Schema[field]
		if !ok {
			t.Fatalf("expected %s field in schema", field)
		}
		if !s.Required {
			t.Fatalf("expected %s to be Required", field)
		}
	}

	forceNewFields := []string{"engine_version", "db_instance_mode", "instance_network_type", "vpc_id", "vswitch_id", "zone_id", "pay_type", "storage_type", "master_node_num", "serverless_mode", "encryption_key", "encryption_type"}
	for _, field := range forceNewFields {
		s, ok := r.Schema[field]
		if !ok {
			t.Fatalf("expected %s field in schema", field)
		}
		if !s.ForceNew {
			t.Fatalf("expected %s to be ForceNew", field)
		}
	}

	computedFields := []string{"status", "connection_string", "port", "creation_time"}
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

func TestAdbpgInstanceImporterExists(t *testing.T) {
	r := resourceAliCloudAdbpgInstance()
	if r.Importer == nil {
		t.Fatal("expected Importer to be set for terraform import support")
	}
}

func TestAdbpgInstanceTimeouts(t *testing.T) {
	r := resourceAliCloudAdbpgInstance()
	if r.Timeouts == nil {
		t.Fatal("expected Timeouts to be set")
	}
	if r.Timeouts.Create == nil {
		t.Fatal("expected Create timeout")
	}
	if r.Timeouts.Update == nil {
		t.Fatal("expected Update timeout")
	}
	if r.Timeouts.Delete == nil {
		t.Fatal("expected Delete timeout")
	}
}

func TestAdbpgInstanceCRUDFunctionsNotNil(t *testing.T) {
	r := resourceAliCloudAdbpgInstance()
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

func TestAdbpgInstanceTagsSchema(t *testing.T) {
	r := resourceAliCloudAdbpgInstance()
	s, ok := r.Schema["tags"]
	if !ok {
		t.Fatal("expected tags field in schema")
	}
	if !s.Optional {
		t.Fatal("expected tags to be Optional")
	}
}
