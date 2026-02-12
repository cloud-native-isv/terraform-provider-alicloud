package alicloud

import (
	"fmt"
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	slsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAlicloudLogtailConfig_basic(t *testing.T) {
	var config slsAPI.LogtailPipelineConfig

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAlicloudLogtailConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlicloudLogtailConfigBasic("tf-test-project", "tf-test-config"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlicloudLogtailConfigExists("alicloud_logtail_config.foo", &config),
					resource.TestCheckResourceAttr("alicloud_logtail_config.foo", "name", "tf-test-config"),
					resource.TestCheckResourceAttr("alicloud_logtail_config.foo", "project", "tf-test-project"),
				),
			},
		},
	})
}

func testAccCheckAlicloudLogtailConfigExists(n string, config *slsAPI.LogtailPipelineConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Logtail Config ID is set")
		}

		client := testAccProvider.Meta().(*connectivity.AliyunClient)
		slsService, err := NewSlsService(client)
		if err != nil {
			return err
		}

		found, err := slsService.DescribeSlsLogtailPipelineConfig(rs.Primary.ID)
		if err != nil {
			return err
		}

		*config = *found
		return nil
	}
}

func testAccCheckAlicloudLogtailConfigDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "alicloud_logtail_config" {
			continue
		}

		_, err := slsService.DescribeSlsLogtailPipelineConfig(rs.Primary.ID)
		if err != nil {
			if NotFoundError(err) {
				continue
			}
			return err
		}
		return fmt.Errorf("Logtail Config %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAlicloudLogtailConfigBasic(project, name string) string {
	return fmt.Sprintf(`
resource "alicloud_log_project" "foo" {
  name = "%s"
  description = "tf test project"
}

resource "alicloud_log_store" "foo" {
  project = alicloud_log_project.foo.name
  name = "tf-test-logstore"
}

resource "alicloud_logtail_config" "foo" {
  project = alicloud_log_project.foo.name
  name    = "%s"
  
  inputs {
    type = "file_input"
    config_json = "{\"logPath\":\"/var/log\"}"
  }

  flushers {
    type = "flusher_sls"
    config_json = <<EOF
    {
      "endpoint": "cn-hangzhou.log.aliyuncs.com",
      "logstore": "${alicloud_log_store.foo.name}"
    }
    EOF
  }
}
`, project, name)
}
