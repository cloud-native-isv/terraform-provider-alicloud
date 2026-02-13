package alicloud

import (
	"fmt"
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestResourceAlicloudLogtailPipelineConfig_SchemaBasics(t *testing.T) {
	r := resourceAliCloudLogtailPipelineConfig()
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

func TestResourceAlicloudLogtailPipelineConfig_NameValidation(t *testing.T) {
	r := resourceAliCloudLogtailPipelineConfig()
	v := r.Schema["name"].ValidateFunc
	if v == nil {
		t.Fatalf("name ValidateFunc should not be nil")
	}
	_, errs := v("valid-name_123", "name")
	if len(errs) > 0 {
		t.Fatalf("expected valid name, got errors: %v", errs)
	}
	_, errs = v("INVALID NAME", "name")
	if len(errs) == 0 {
		t.Fatalf("expected invalid name to fail validation")
	}
}

func testAccCheckAliCloudLogtailPipelineConfigExists(n string, config *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Logtail Pipeline Config ID is set")
		}

		client := testAccProvider.Meta().(*connectivity.AliyunClient)
		service, err := NewSlsService(client)
		if err != nil {
			return err
		}

		_, err = service.DescribeSlsLogtailPipelineConfig(rs.Primary.ID)
		if err != nil {
			return err
		}

		return nil
	}
}
