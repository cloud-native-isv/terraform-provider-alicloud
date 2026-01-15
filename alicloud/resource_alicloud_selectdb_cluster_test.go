package alicloud

import (
	"fmt"
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestAccAliCloudSelectDBCluster_basic(t *testing.T) {
	var clusterId string

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAliCloudSelectDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAliCloudSelectDBClusterConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliCloudSelectDBClusterExists("alicloud_selectdb_cluster.default", &clusterId),
					resource.TestCheckResourceAttr("alicloud_selectdb_cluster.default", "description", "tf-test-selectdb-cluster"),
					resource.TestCheckResourceAttr("alicloud_selectdb_cluster.default", "engine", "selectdb"),
					resource.TestCheckResourceAttr("alicloud_selectdb_cluster.default", "charge_type", "PostPaid"),
				),
			},
		},
	})
}

func testAccCheckAliCloudSelectDBClusterExists(n string, clusterId *string) resource.TestCheckFunc {
	return func(s *schema.TerraformState) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SelectDB Cluster ID is set")
		}

		client := testAccProvider.Meta().(*connectivity.AliyunClient)
		service, err := NewSelectDBService(client)
		if err != nil {
			return err
		}

		instanceId, cId, err := service.DecodeSelectDBClusterId(rs.Primary.ID)
		if err != nil {
			return err
		}

		cluster, err := service.DescribeSelectDBCluster(instanceId, cId)
		if err != nil {
			return err
		}

		if cluster == nil {
			return fmt.Errorf("SelectDB Cluster not found")
		}

		*clusterId = rs.Primary.ID
		return nil
	}
}

func testAccCheckAliCloudSelectDBClusterDestroy(s *schema.TerraformState) error {
	client := testAccProvider.Meta().(*connectivity.AliyunClient)
	service, err := NewSelectDBService(client)
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "alicloud_selectdb_cluster" {
			continue
		}

		instanceId, cId, err := service.DecodeSelectDBClusterId(rs.Primary.ID)
		if err != nil {
			return err
		}

		cluster, err := service.DescribeSelectDBCluster(instanceId, cId)
		if err != nil {
			if NotFoundError(err) {
				continue
			}
			return err
		}

		if cluster != nil {
			return fmt.Errorf("SelectDB Cluster still exists")
		}
	}

	return nil
}

const testAccAliCloudSelectDBClusterConfig_basic = `
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

resource "alicloud_selectdb_cluster" "default" {
  instance_id         = alicloud_selectdb_instance.default.id
  description         = "tf-test-selectdb-cluster"
  zone_id             = data.alicloud_zones.default.zones.0.id
  vpc_id              = data.alicloud_vpcs.default.ids.0
  vswitch_id          = data.alicloud_vswitches.default.ids.0
  cluster_class       = "selectdb.xlarge"
  cache_size          = 200
  engine              = "selectdb"
  engine_version      = "3.0"
  charge_type         = "PostPaid"
}
`
