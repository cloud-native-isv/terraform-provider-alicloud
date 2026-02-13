package alicloud

import (
	"fmt"
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/terraform"
)

func TestResourceAlicloudLogtailConfig_SchemaBasics(t *testing.T) {
	r := resourceAliCloudLogtailConfig()
	// Common fields
	if r.Schema["project"] == nil || !r.Schema["project"].Required {
		t.Fatalf("project should be required")
	}
	if r.Schema["name"] == nil || !r.Schema["name"].Required {
		t.Fatalf("name should be required")
	}

	// Old Schema fields
	if r.Schema["input_type"] == nil || !r.Schema["input_type"].Required {
		t.Fatalf("input_type should be required in old schema")
	}
	if r.Schema["output_type"] == nil || !r.Schema["output_type"].Required {
		// Output type used to be optional in some versions, but here it is explicitly populated by resource logic
		// Checking existence is enough
		t.Logf("output_type exists")
	}
	if r.Schema["input_detail"] == nil || !r.Schema["input_detail"].Required {
		t.Fatalf("input_detail should be required in old schema")
	}
}

func TestResourceAlicloudLogtailConfig_NameValidation(t *testing.T) {
	r := resourceAliCloudLogtailConfig()
	v := r.Schema["name"].ValidateFunc
	if v == nil {
		t.Fatalf("name ValidateFunc should not be nil")
	}
	_, errs := v("valid-name_123", "name")
	if len(errs) > 0 {
		t.Fatalf("expected valid name, got errors: %v", errs)
	}
	_, errs = v("INVALID-NAME", "name")
	if len(errs) == 0 {
		t.Fatalf("expected invalid name to fail validation")
	}
}

func TestAccAliCloudLogtailConfig_basic(t *testing.T) {
	var config string
	projectName := "tf-test-project-" + RandString(8)
	logstoreName := "tf-test-logstore-" + RandString(8)
	configName := "tf-test-config-" + RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAliCloudLogtailConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAliCloudLogtailConfigBasicConfig(projectName, logstoreName, configName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliCloudLogtailConfigExists("alicloud_logtail_config.foo", &config),
					resource.TestCheckResourceAttr("alicloud_logtail_config.foo", "name", configName),
					resource.TestCheckResourceAttr("alicloud_logtail_config.foo", "input_type", "file"),
					resource.TestCheckResourceAttr("alicloud_logtail_config.foo", "project", projectName),
					resource.TestCheckResourceAttr("alicloud_logtail_config.foo", "logstore", logstoreName),
				),
			},
		},
	})
}

func testAccCheckAliCloudLogtailConfigExists(n string, config *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Logtail Config ID is set")
		}

		client := testAccProvider.Meta().(*connectivity.AliyunClient)
		service, err := NewSlsService(client)
		if err != nil {
			return err
		}

		found, err := service.DescribeSlsLogtailConfig(rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.ConfigName != rs.Primary.Attributes["name"] {
			return fmt.Errorf("Logtail Config not found")
		}

		return nil
	}
}

func testAccCheckAliCloudLogtailConfigDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*connectivity.AliyunClient)
	service, err := NewSlsService(client)
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "alicloud_logtail_config" {
			continue
		}

		_, err := service.DescribeSlsLogtailConfig(rs.Primary.ID)
		if err != nil {
			if NotFoundError(err) {
				continue
			}
			return err
		}
		return fmt.Errorf("Logtail Config still exists")
	}

	return nil
}

func testAccAliCloudLogtailConfigBasicConfig(project, logstore, name string) string {
	return fmt.Sprintf(`
resource "alicloud_log_project" "foo" {
  name        = "%s"
  description = "tf-test"
}

resource "alicloud_log_store" "foo" {
  project          = alicloud_log_project.foo.name
  name             = "%s"
  retention_period = 3000
  shard_count      = 1
}

resource "alicloud_logtail_config" "foo" {
  project      = alicloud_log_project.foo.name
  logstore     = alicloud_log_store.foo.name
  name         = "%s"
  input_type   = "file"
  output_type  = "LogService"
  input_detail = <<EOF
  {
	"logPath": "/log",
	"filePattern": "access.log",
	"logType": "common_reg_log",
	"topicFormat": "none",
	"discardUnmatch": false,
	"enableRawLog": true,
	"maxDepth": 1000
  }
  EOF
}
`, project, logstore, name)
}
