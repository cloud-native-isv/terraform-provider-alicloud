package alicloud

import (
	"fmt"
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAlicloudOtsInstance_vcu(t *testing.T) {
	var instanceId string
	rName := "tf-test-ots-vcu-" + acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAlicloudOtsInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAlicloudOtsInstanceConfig_vcu(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlicloudOtsInstanceExists("alicloud_ots_instance.default", &instanceId),
					resource.TestCheckResourceAttr("alicloud_ots_instance.default", "instance_specification", "VCU"),
					resource.TestCheckResourceAttr("alicloud_ots_instance.default", "elastic_vcu_upper_limit", "1"),
				),
			},
			{
				Config: testAccAlicloudOtsInstanceConfig_updateVcuLimit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlicloudOtsInstanceExists("alicloud_ots_instance.default", &instanceId),
					resource.TestCheckResourceAttr("alicloud_ots_instance.default", "instance_specification", "VCU"),
					resource.TestCheckResourceAttr("alicloud_ots_instance.default", "elastic_vcu_upper_limit", "2"),
				),
			},
		},
	})
}

func testAccCheckAlicloudOtsInstanceExists(n string, instanceId *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No OTS Instance ID is set")
		}

		client := testAccProvider.Meta().(*connectivity.AliyunClient)
		otsService, err := NewOtsService(client)
		if err != nil {
			return err
		}

		instance, err := otsService.DescribeOtsInstance(rs.Primary.ID)
		if err != nil {
			return err
		}

		if instance == nil {
			return fmt.Errorf("OTS Instance not found")
		}

		*instanceId = rs.Primary.ID
		return nil
	}
}

func testAccCheckAlicloudOtsInstanceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "alicloud_ots_instance" {
			continue
		}

		instance, err := otsService.DescribeOtsInstance(rs.Primary.ID)
		if err != nil {
			if NotFoundError(err) {
				continue
			}
			return err
		}

		if instance != nil {
			return fmt.Errorf("OTS Instance still exists")
		}
	}

	return nil
}

func testAccAlicloudOtsInstanceConfig_vcu(name string) string {
	return fmt.Sprintf(`
variable "name" {
  default = "%s"
}

resource "alicloud_ots_instance" "default" {
  name                   = var.name
  instance_specification = "VCU"
  description            = "tf-test-ots-vcu-description"
  elastic_vcu_upper_limit = 1
  network_type_acl       = ["INTERNET", "VPC", "CLASSIC"]
}
`, name)
}

func testAccAlicloudOtsInstanceConfig_updateVcuLimit(name string) string {
	return fmt.Sprintf(`
variable "name" {
  default = "%s"
}

resource "alicloud_ots_instance" "default" {
  name                   = var.name
  instance_specification = "VCU"
  description            = "tf-test-ots-vcu-description"
  elastic_vcu_upper_limit = 2
  network_type_acl       = ["INTERNET", "VPC", "CLASSIC"]
}
`, name)
}
