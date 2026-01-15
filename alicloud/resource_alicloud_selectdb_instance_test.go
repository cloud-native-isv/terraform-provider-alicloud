package alicloud

import (
	"fmt"
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestAccAliCloudSelectDBInstance_basic(t *testing.T) {
	var instanceId string

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAliCloudSelectDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAliCloudSelectDBInstanceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliCloudSelectDBInstanceExists("alicloud_selectdb_instance.default", &instanceId),
					resource.TestCheckResourceAttr("alicloud_selectdb_instance.default", "instance_name", "tf-test-selectdb"),
					resource.TestCheckResourceAttr("alicloud_selectdb_instance.default", "engine", "selectdb"),
					resource.TestCheckResourceAttr("alicloud_selectdb_instance.default", "charge_type", "PostPaid"),
				),
			},
		},
	})
}

func testAccCheckAliCloudSelectDBInstanceExists(n string, instanceId *string) resource.TestCheckFunc {
	return func(s *schema.TerraformState) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SelectDB Instance ID is set")
		}

		client := testAccProvider.Meta().(*connectivity.AliyunClient)
		service, err := NewSelectDBService(client)
		if err != nil {
			return err
		}

		instance, err := service.DescribeSelectDBInstance(rs.Primary.ID)
		if err != nil {
			return err
		}

		if instance == nil {
			return fmt.Errorf("SelectDB Instance not found")
		}

		*instanceId = rs.Primary.ID
		return nil
	}
}

func testAccCheckAliCloudSelectDBInstanceDestroy(s *schema.TerraformState) error {
	client := testAccProvider.Meta().(*connectivity.AliyunClient)
	service, err := NewSelectDBService(client)
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "alicloud_selectdb_instance" {
			continue
		}

		instance, err := service.DescribeSelectDBInstance(rs.Primary.ID)
		if err != nil {
			if NotFoundError(err) {
				continue
			}
			return err
		}

		if instance != nil {
			return fmt.Errorf("SelectDB Instance still exists")
		}
	}

	return nil
}

const testAccAliCloudSelectDBInstanceConfig_basic = `
variable "name" {
  default = "tf-test-selectdb"
}

data "alicloud_zones" "default" {
  available_resource_creation = "SelectDB"
}

data "alicloud_vpcs" "default" {
  name_regex = "^default-NODELETING$"
}

data "alicloud_vswitches" "default" {
  vpc_id = data.alicloud_vpcs.default.ids.0
  zone_id = data.alicloud_zones.default.zones.0.id
}

resource "alicloud_selectdb_instance" "default" {
  instance_name       = var.name
  engine              = "selectdb"
  engine_version      = "3.0"
  zone_id             = data.alicloud_zones.default.zones.0.id
  vpc_id              = data.alicloud_vpcs.default.ids.0
  vswitch_id          = data.alicloud_vswitches.default.ids.0
  instance_class      = "selectdb.xlarge"
  cache_size          = 200
  charge_type         = "PostPaid"
  username            = "admin_test"
  password            = "Test1234!"
}
`
