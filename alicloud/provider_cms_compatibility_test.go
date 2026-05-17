package alicloud

import (
	"os"
	"strings"
	"testing"
)

func TestProviderCMSRegistrationCompatibility(t *testing.T) {
	content, err := os.ReadFile("provider.go")
	if err != nil {
		t.Fatalf("read provider.go failed: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, `"alicloud_cms_service"`) {
		t.Fatalf("expected provider registration to include alicloud_cms_service")
	}
}

func TestProviderCMSPlaceholderFilesRemainUnregistered(t *testing.T) {
	content, err := os.ReadFile("provider.go")
	if err != nil {
		t.Fatalf("read provider.go failed: %v", err)
	}
	text := string(content)

	for _, name := range []string{
		"alicloud_cms_addon",
		"alicloud_cms_agg_task",
		"alicloud_cms_context_store",
		"alicloud_cms_workspace",
		"alicloud_cms_umodel",
		"alicloud_cms_pipeline",
	} {
		if strings.Contains(text, `"`+name+`"`) {
			t.Fatalf("expected placeholder cms capability %s to remain unregistered", name)
		}
	}
}
