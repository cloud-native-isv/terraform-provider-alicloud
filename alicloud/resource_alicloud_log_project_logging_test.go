package alicloud

import (
	"fmt"
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func init() {
	resource.AddTestSweepers("alicloud_log_project_logging", &resource.Sweeper{
		Name: "alicloud_log_project_logging",
		F:    testSweepLogProjectLogging,
		Dependencies: []string{
			"alicloud_log_project",
		},
	})
}

func testSweepLogProjectLogging(region string) error {
	// This sweeper is intentionally empty because the service logging configuration
	// will be cleaned up when the associated log project is deleted
	return nil
}

func TestAccAlicloudLogProjectLogging_basic(t *testing.T) {
	var v map[string]interface{}
	resourceId := "alicloud_log_project_logging.default"
	ra := resourceAttrInit(resourceId, AlicloudLogProjectLoggingMap)
	rc := resourceCheckInitWithDescribeMethod(resourceId, &v, func() interface{} {
		return &SlsServiceV2{testAccProvider.Meta().(*connectivity.AliyunClient)}
	}, "GetSlsLogging")
	rac := resourceAttrCheckInit(rc, ra)
	testAccCheck := rac.resourceAttrMapUpdateSet()
	rand := acctest.RandIntRange(10000, 99999)
	name := fmt.Sprintf("tf-testacc%sslsproject%d", defaultRegionToTest, rand)
	testAccConfig := resourceTestAccConfigFunc(resourceId, name, AlicloudLogProjectLoggingBasicDependence)
	
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
					"project": "${alicloud_log_project.example.project_name}",
					"logging_details": []map[string]interface{}{
						{
							"type":     "operation_log",
							"logstore": "internal-operation_log",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"project":           name,
						"logging_details.#": "1",
					}),
				),
			},
			{
				ResourceName:      resourceId,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"logging_details": []map[string]interface{}{
						{
							"type":     "operation_log",
							"logstore": "updated-operation_log",
						},
						{
							"type":     "consumer_group",
							"logstore": "internal-consumer_group",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"logging_details.#": "2",
					}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"logging_details": []map[string]interface{}{
						{
							"type":     "operation_log",
							"logstore": "internal-operation_log",
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"logging_details.#": "1",
					}),
				),
			},
		},
	})
}

var AlicloudLogProjectLoggingMap = map[string]string{
	"project": CHECKSET,
}

func AlicloudLogProjectLoggingBasicDependence(name string) string {
	return fmt.Sprintf(`
variable "name" {
    default = "%s"
}

resource "alicloud_log_project" "example" {
  project_name = var.name
  description  = "created by terraform"
}

`, name)
}