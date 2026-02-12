package alicloud

import (
	"testing"
)

func TestResourceAlicloudLogtailConfig_SchemaBasics(t *testing.T) {
	r := resourceAliCloudLogtailConfig()
	if r.Schema["project"] == nil || !r.Schema["project"].Required {
		t.Fatalf("project should be required")
	}
	if r.Schema["name"] == nil || !r.Schema["name"].Required {
		t.Fatalf("name should be required")
	}
	if r.Schema["inputs"] == nil || !r.Schema["inputs"].Required {
		t.Fatalf("inputs should be required")
	}
	if r.Schema["flushers"] == nil || !r.Schema["flushers"].Required {
		t.Fatalf("flushers should be required")
	}
}

func TestResourceAlicloudLogtailConfig_NameValidation(t *testing.T) {
	r := resourceAliCloudLogtailConfig()
	v := r.Schema["name"].ValidateFunc
	if v == nil {
		t.Fatalf("name ValidateFunc should not be nil")
	}
	_, errs := v("kangaroo-pai-file-fabricmanager-proxy", "name")
	if len(errs) > 0 {
		t.Fatalf("expected valid name, got errors: %v", errs)
	}
	_, errs = v("INVALID-NAME", "name")
	if len(errs) == 0 {
		t.Fatalf("expected invalid name to fail validation")
	}
}
