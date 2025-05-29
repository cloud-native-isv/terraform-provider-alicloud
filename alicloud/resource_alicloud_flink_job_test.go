package alicloud

import (
	"fmt"
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAliCloudFlinkJob_basic(t *testing.T) {
	var v map[string]interface{}
	resourceId := "alicloud_flink_job.default"
	ra := resourceAttrInit(resourceId, AliCloudFlinkJobMap0)
	rc := resourceCheckInitWithDescribeMethod(resourceId, &v, func() interface{} {
		client := testAccProvider.Meta().(*connectivity.AliyunClient)
		service, _ := NewFlinkService(client)
		return service
	}, "DescribeFlinkJob")
	rac := resourceAttrCheckInit(rc, ra)

	testAccCheck := rac.resourceAttrMapUpdateSet()
	rand := acctest.RandIntRange(10000, 99999)
	name := fmt.Sprintf("tf-testacc%sflinkjob%d", defaultRegionToTest, rand)
	testAccConfig := resourceTestAccConfigFunc(resourceId, name, AliCloudFlinkJobBasicDependence0)
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
					"workspace_id":             "${alicloud_flink_workspace.default.id}",
					"namespace":                "${alicloud_flink_namespace.default.name}",
					"deployment_id":            "${alicloud_flink_deployment.default.id}",
					"job_name":                 name,
					"allow_non_restored_state": false,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"workspace_id":             CHECKSET,
						"namespace":                CHECKSET,
						"deployment_id":            CHECKSET,
						"job_name":                 name,
						"allow_non_restored_state": "false",
					}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"parallelism": "2",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"parallelism": "2",
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

var AliCloudFlinkJobMap0 = map[string]string{
	"status": CHECKSET,
}

func AliCloudFlinkJobBasicDependence0(name string) string {
	return fmt.Sprintf(`
variable "name" {
	default = "%s"
}

resource "alicloud_flink_workspace" "default" {
  name              = var.name
  resource_group_id = data.alicloud_resource_manager_resource_groups.default.ids.0
  zone_id           = data.alicloud_zones.default.zones.0.id
  vpc_id            = alicloud_vpc.default.id
  vswitch_ids       = [alicloud_vswitch.default.id]
  security_group_id = alicloud_security_group.default.id

  resource {
    cpu    = 500
    memory = 2000
  }

  storage {
    oss_bucket = alicloud_oss_bucket.default.bucket
  }
}

resource "alicloud_flink_namespace" "default" {
  workspace_id = alicloud_flink_workspace.default.id
  name         = var.name
  cpu          = 200
  memory       = 1000
}

resource "alicloud_flink_deployment" "default" {
  workspace_id = alicloud_flink_workspace.default.id
  namespace    = alicloud_flink_namespace.default.name
  name         = var.name
  target       = "STREAMING"
  deployment_type = "STREAMING"
  
  artifact {
    artifact_type = "JAR"
    artifact_url  = "oss://test-bucket/test.jar"
    main_class    = "com.example.MainClass"
  }
}

data "alicloud_resource_manager_resource_groups" "default" {}

data "alicloud_zones" "default" {
  available_resource_creation = "VSwitch"
}

resource "alicloud_vpc" "default" {
  vpc_name   = var.name
  cidr_block = "172.16.0.0/16"
}

resource "alicloud_vswitch" "default" {
  vpc_id       = alicloud_vpc.default.id
  cidr_block   = "172.16.0.0/24"
  zone_id      = data.alicloud_zones.default.zones.0.id
  vswitch_name = var.name
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
