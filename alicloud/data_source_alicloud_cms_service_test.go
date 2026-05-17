package alicloud

import (
	"os"
	"strings"
	"testing"
)

func TestDataSourceAliCloudCmsServiceSchema(t *testing.T) {
	r := dataSourceAliCloudCmsService()
	if r == nil {
		t.Fatalf("expected non-nil data source")
	}
	if _, ok := r.Schema["enable"]; !ok {
		t.Fatalf("expected enable field")
	}
	if _, ok := r.Schema["status"]; !ok {
		t.Fatalf("expected status field")
	}
	if got := r.Schema["enable"].Default; got != "Off" {
		t.Fatalf("expected enable default Off, got %v", got)
	}
}

func TestDataSourceAliCloudCmsServiceLayeringAudit(t *testing.T) {
	content, err := os.ReadFile("data_source_alicloud_cms_service.go")
	if err != nil {
		t.Fatalf("read data source file failed: %v", err)
	}
	text := string(content)
	if strings.Contains(text, "RpcPost(") {
		t.Fatalf("expected data source not to call RpcPost directly")
	}
	if strings.Count(text, "package alicloud") != 1 {
		t.Fatalf("expected exactly one package declaration")
	}
	if !strings.Contains(text, "CmsServiceHasNotBeenOpened") {
		t.Fatalf("expected not-opened compatibility ID")
	}
	if !strings.Contains(text, "CmsServiceHasBeenOpened") {
		t.Fatalf("expected opened compatibility ID")
	}
	if !strings.Contains(text, `d.Set("status", "Opened")`) {
		t.Fatalf("expected opened status compatibility behavior")
	}
}
