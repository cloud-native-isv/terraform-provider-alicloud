package alicloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAlicloudSelectDBClusters_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlicloudSelectDBClustersConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.alicloud_selectdb_clusters.default", "clusters.#", "1"),
					resource.TestCheckResourceAttrSet("data.alicloud_selectdb_clusters.default", "clusters.0.cluster_id"),
					resource.TestCheckResourceAttrSet("data.alicloud_selectdb_clusters.default", "clusters.0.cluster_description"),
				),
			},
		},
	})
}

const testAccAlicloudSelectDBClustersConfig_basic = `
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

data "alicloud_selectdb_clusters" "default" {
  ids = [alicloud_selectdb_cluster.default.id]
}
`
