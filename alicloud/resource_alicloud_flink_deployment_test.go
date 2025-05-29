package alicloud

import (
	"fmt"
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAliCloudFlinkDeployment_basic0(t *testing.T) {
	var v map[string]interface{}
	resourceId := "alicloud_flink_deployment.default"
	checkoutSupportedRegions(t, true, connectivity.FlinkSupportRegions)
	ra := resourceAttrInit(resourceId, AliCloudFlinkDeploymentMap0)
	rc := resourceCheckInitWithDescribeMethod(resourceId, &v, func() interface{} {
		return &FlinkService{client: testAccProvider.Meta().(*connectivity.AliyunClient)}
	}, "GetDeployment")
	rac := resourceAttrCheckInit(rc, ra)
	testAccCheck := rac.resourceAttrMapUpdateSet()
	rand := acctest.RandIntRange(10000, 99999)
	name := fmt.Sprintf("tf-testacc%sflinkdeployment%d", defaultRegionToTest, rand)
	testAccConfig := resourceTestAccConfigFunc(resourceId, name, AliCloudFlinkDeploymentBasicDependence0)
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
					"workspace_id": "${alicloud_flink_workspace.default.id}",
					"namespace_id": "default",
					"job_name":     name,
					"entry_class":  "com.example.WordCount",
					"jar_uri":      "oss://example-bucket/wordcount.jar",
					"parallelism":  "2",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"workspace_id": CHECKSET,
						"namespace_id": "default",
						"job_name":     name,
						"entry_class":  "com.example.WordCount",
						"jar_uri":      "oss://example-bucket/wordcount.jar",
						"parallelism":  "2",
					}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"parallelism": "4",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"parallelism": "4",
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

var AliCloudFlinkDeploymentMap0 = map[string]string{
	"deployment_id": CHECKSET,
	"status":        CHECKSET,
}

func AliCloudFlinkDeploymentBasicDependence0(name string) string {
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

resource "alicloud_flink_workspace" "default" {
  name               = var.name
  resource_group_id  = data.alicloud_resource_manager_resource_groups.default.ids.0
  zone_id           = data.alicloud_flink_zones.default.ids.0
  vpc_id            = alicloud_vpc.default.id
  vswitch_ids       = [alicloud_vswitch.default.id]
  
  storage {
    oss_bucket = alicloud_oss_bucket.default.bucket
  }
  
  resource {
    cpu    = 4
    memory = 8192
  }
}
`, name)
}
