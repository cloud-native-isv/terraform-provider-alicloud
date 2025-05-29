package alicloud

import (
	"fmt"
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAliCloudFlinkWorkspace_basic0(t *testing.T) {
	var v map[string]interface{}
	resourceId := "alicloud_flink_workspace.default"
	checkoutSupportedRegions(t, true, connectivity.FlinkSupportRegions)
	ra := resourceAttrInit(resourceId, AliCloudFlinkWorkspaceMap0)
	rc := resourceCheckInitWithDescribeMethod(resourceId, &v, func() interface{} {
		return &FlinkService{client: testAccProvider.Meta().(*connectivity.AliyunClient)}
	}, "DescribeFlinkWorkspace")
	rac := resourceAttrCheckInit(rc, ra)
	testAccCheck := rac.resourceAttrMapUpdateSet()
	rand := acctest.RandIntRange(10000, 99999)
	name := fmt.Sprintf("tf-testacc%sflinkworkspace%d", defaultRegionToTest, rand)
	testAccConfig := resourceTestAccConfigFunc(resourceId, name, AliCloudFlinkWorkspaceBasicDependence0)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		IDRefreshName: resourceId,
		Providers:     testAccProviders,
		CheckDestroy:  rac.checkResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccConfig(map[string]interface{}{
					"name":              name,
					"resource_group_id": "${data.alicloud_resource_manager_resource_groups.default.ids.0}",
					"zone_id":           "${data.alicloud_flink_zones.default.ids.0}",
					"vpc_id":            "${alicloud_vpc.default.id}",
					"vswitch_ids":       []string{"${alicloud_vswitch.default.id}"},
					"storage": []map[string]interface{}{
						{
							"oss_bucket": "${alicloud_oss_bucket.default.bucket}",
						},
					},
					"resource": []map[string]interface{}{
						{
							"cpu":    "4",
							"memory": "8192",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"name":              name,
						"resource_group_id": CHECKSET,
						"zone_id":           CHECKSET,
						"vpc_id":            CHECKSET,
						"vswitch_ids.#":     "1",
						"storage.#":         "1",
						"resource.#":        "1",
					}),
				),
			},
			{
				ResourceName:      resourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

var AliCloudFlinkWorkspaceMap0 = map[string]string{
	"status": CHECKSET,
}

func AliCloudFlinkWorkspaceBasicDependence0(name string) string {
	return fmt.Sprintf(`
variable "name" {
  default = "%s"
}

data "alicloud_resource_manager_resource_groups" "default" {
  status = "OK"
}

data "alicloud_flink_zones" "default" {}

resource "alicloud_vpc" "default" {
  vpc_name   = var.name
  cidr_block = "10.4.0.0/16"
}

resource "alicloud_vswitch" "default" {
  vswitch_name = var.name
  cidr_block   = "10.4.0.0/24"
  vpc_id       = alicloud_vpc.default.id
  zone_id      = data.alicloud_flink_zones.default.ids.0
}

resource "alicloud_oss_bucket" "default" {
  bucket = var.name
}
`, name)
}

func TestAccAliCloudFlinkWorkspace_basic1(t *testing.T) {
	var v map[string]interface{}
	resourceId := "alicloud_flink_workspace.default"
	checkoutSupportedRegions(t, true, connectivity.FlinkSupportRegions)
	ra := resourceAttrInit(resourceId, AliCloudFlinkWorkspaceMap1)
	rc := resourceCheckInitWithDescribeMethod(resourceId, &v, func() interface{} {
		return &FlinkService{client: testAccProvider.Meta().(*connectivity.AliyunClient)}
	}, "DescribeFlinkWorkspace")
	rac := resourceAttrCheckInit(rc, ra)
	testAccCheck := rac.resourceAttrMapUpdateSet()
	rand := acctest.RandIntRange(10000, 99999)
	name := fmt.Sprintf("tf-testacc%sflinkworkspace%d", defaultRegionToTest, rand)
	testAccConfig := resourceTestAccConfigFunc(resourceId, name, AliCloudFlinkWorkspaceBasicDependence1)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		IDRefreshName: resourceId,
		Providers:     testAccProviders,
		CheckDestroy:  rac.checkResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccConfig(map[string]interface{}{
					"name":              name,
					"description":       "Test workspace created by Terraform",
					"resource_group_id": "${data.alicloud_resource_manager_resource_groups.default.ids.0}",
					"zone_id":           "${data.alicloud_flink_zones.default.ids.0}",
					"vpc_id":            "${alicloud_vpc.default.id}",
					"vswitch_ids":       []string{"${alicloud_vswitch.default.id}"},
					"security_group_id": "${alicloud_security_group.default.id}",
					"architecture_type": "X86",
					"charge_type":       "POST",
					"storage": []map[string]interface{}{
						{
							"oss_bucket": "${alicloud_oss_bucket.default.bucket}",
						},
					},
					"resource": []map[string]interface{}{
						{
							"cpu":    "8",
							"memory": "16384",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"name":              name,
						"description":       "Test workspace created by Terraform",
						"resource_group_id": CHECKSET,
						"zone_id":           CHECKSET,
						"vpc_id":            CHECKSET,
						"vswitch_ids.#":     "1",
						"security_group_id": CHECKSET,
						"architecture_type": "X86",
						"charge_type":       "POST",
						"storage.#":         "1",
						"resource.#":        "1",
					}),
				),
			},
			{
				ResourceName:      resourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

var AliCloudFlinkWorkspaceMap1 = map[string]string{
	"status": CHECKSET,
}

func AliCloudFlinkWorkspaceBasicDependence1(name string) string {
	return fmt.Sprintf(`
variable "name" {
  default = "%s"
}

data "alicloud_resource_manager_resource_groups" "default" {
  status = "OK"
}

data "alicloud_flink_zones" "default" {}

resource "alicloud_vpc" "default" {
  vpc_name   = var.name
  cidr_block = "10.4.0.0/16"
}

resource "alicloud_vswitch" "default" {
  vswitch_name = var.name
  cidr_block   = "10.4.0.0/24"
  vpc_id       = alicloud_vpc.default.id
  zone_id      = data.alicloud_flink_zones.default.ids.0
}

resource "alicloud_security_group" "default" {
  name   = var.name
  vpc_id = alicloud_vpc.default.id
}

resource "alicloud_oss_bucket" "default" {
  bucket = var.name
}
`, name)
}
