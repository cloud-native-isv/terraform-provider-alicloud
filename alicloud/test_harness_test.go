package alicloud

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var testAccProvider = Provider().(*schema.Provider)

var testAccProviderFactories = map[string]terraform.ResourceProviderFactory{
	"alicloud": func() (terraform.ResourceProvider, error) {
		return testAccProvider, nil
	},
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	requireAnyEnvironmentVariable(t, "access key", "ALICLOUD_ACCESS_KEY", "ALIBABA_CLOUD_ACCESS_KEY_ID", "ALIBABACLOUD_ACCESS_KEY_ID")
	requireAnyEnvironmentVariable(t, "secret key", "ALICLOUD_SECRET_KEY", "ALIBABA_CLOUD_ACCESS_KEY_SECRET", "ALIBABACLOUD_ACCESS_KEY_SECRET")
	requireAnyEnvironmentVariable(t, "region", "ALICLOUD_REGION", "ALIBABA_CLOUD_REGION")
}

func requireAnyEnvironmentVariable(t *testing.T, description string, names ...string) {
	t.Helper()
	for _, name := range names {
		if os.Getenv(name) != "" {
			return
		}
	}
	t.Fatalf("one of %s must be set for the acceptance test %s", strings.Join(names, ", "), description)
}

func TestTestAccProviderFactory(t *testing.T) {
	factory, ok := testAccProviderFactories["alicloud"]
	if !ok {
		t.Fatal("alicloud provider factory is missing")
	}
	provider, err := factory()
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}
	if provider != testAccProvider {
		t.Fatal("provider factory returned an unexpected provider")
	}
}
