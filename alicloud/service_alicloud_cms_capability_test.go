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
