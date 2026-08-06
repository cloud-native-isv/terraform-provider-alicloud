package alicloud

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var cmsObjectGroups = []string{
	"addon",
	"agg_task",
	"alert",
	"cloud_resource",
	"context",
	"context_store",
	"dataset",
	"delivery_task",
	"entity_store",
	"integration_policy",
	"memory",
	"memory_store",
	"pipeline",
	"prometheus",
	"service",
	"umodel",
	"workspace",
}

func TestCmsObjectGroupCount(t *testing.T) {
	if len(cmsObjectGroups) != 17 {
		t.Fatalf("expected 17 cms object groups, got %d", len(cmsObjectGroups))
	}
}

func TestCmsServiceFilesShouldNotRemainPlaceholders(t *testing.T) {
	for _, group := range cmsObjectGroups {
		file := "service_alicloud_cms_" + group + ".go"
		content, err := os.ReadFile(filepath.Clean(file))
		if err != nil {
			t.Fatalf("read %s failed: %v", file, err)
		}
		if strings.TrimSpace(string(content)) == "package alicloud" {
			t.Fatalf("expected %s to contain service capabilities, got placeholder", file)
		}
	}
}

// cmsImplementedGroups are the CMS 2.0 object groups implemented in gen-4
// (S2 minimal closed-loop subset). Their service files must call the
// cws-lib-go cms API instead of returning cmsServiceCapabilityBlocked.
var cmsImplementedGroups = []string{
	"workspace",
	"integration_policy",
	"addon",
	"prometheus",
	"alert",
	"cloud_resource",
	"entity_store",
}

func TestCmsImplementedGroupsMustNotReturnBlocked(t *testing.T) {
	for _, group := range cmsImplementedGroups {
		file := "service_alicloud_cms_" + group + ".go"
		content, err := os.ReadFile(filepath.Clean(file))
		if err != nil {
			t.Fatalf("read %s failed: %v", file, err)
		}
		text := string(content)
		if strings.Contains(text, "cmsServiceCapabilityBlocked") {
			t.Fatalf("expected implemented group %s to call the cws-lib-go cms API, still returns cmsServiceCapabilityBlocked", group)
		}
		if !strings.Contains(text, "cws-lib-go/lib/cloud/aliyun/api/cms") {
			t.Fatalf("expected implemented group %s to import the cws-lib-go cms API package", group)
		}
	}
}
