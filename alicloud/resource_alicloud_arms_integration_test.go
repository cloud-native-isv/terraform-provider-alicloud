package alicloud

import (
	"fmt"
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAlicloudArmsIntegration_basic0(t *testing.T) {
	var v map[string]interface{}
	resourceId := "alicloud_arms_integration.default"
	ra := resourceAttrInit(resourceId, AlicloudArmsIntegrationMap0)
	rc := resourceCheckInitWithDescribeMethod(resourceId, &v, func() interface{} {
		return &ArmsService{testAccProvider.Meta().(*connectivity.AliyunClient)}
	}, "DescribeArmsIntegration")
	rac := resourceAttrCheckInit(rc, ra)
	testAccCheck := rac.resourceAttrMapUpdateSet()
	rand := acctest.RandIntRange(10000, 99999)
	name := fmt.Sprintf("tf-testacc%sarmsintegration%d", defaultRegionToTest, rand)
	testAccConfig := resourceTestAccConfigFunc(resourceId, name, AlicloudArmsIntegrationBasicDependence0)
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
					"integration_name": name,
					"integration_type": "webhooks",
					"config":           "{\"url\":\"https://example.com/webhook\",\"secret\":\"test-secret\"}",
					"description":      "Test integration created by Terraform",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"integration_name": name,
						"integration_type": "webhooks",
						"description":      "Test integration created by Terraform",
					}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"description": "Updated test integration description",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"description": "Updated test integration description",
					}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"status": "Inactive",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"status": "Inactive",
					}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"config": "{\"url\":\"https://example.com/updated-webhook\",\"secret\":\"updated-secret\"}",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"config": "{\"url\":\"https://example.com/updated-webhook\",\"secret\":\"updated-secret\"}",
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

func TestAccAlicloudArmsIntegration_basic1(t *testing.T) {
	var v map[string]interface{}
	resourceId := "alicloud_arms_integration.default"
	ra := resourceAttrInit(resourceId, AlicloudArmsIntegrationMap1)
	rc := resourceCheckInitWithDescribeMethod(resourceId, &v, func() interface{} {
		return &ArmsService{testAccProvider.Meta().(*connectivity.AliyunClient)}
	}, "DescribeArmsIntegration")
	rac := resourceAttrCheckInit(rc, ra)
	testAccCheck := rac.resourceAttrMapUpdateSet()
	rand := acctest.RandIntRange(10000, 99999)
	name := fmt.Sprintf("tf-testacc%sarmsintegration%d", defaultRegionToTest, rand)
	testAccConfig := resourceTestAccConfigFunc(resourceId, name, AlicloudArmsIntegrationBasicDependence1)
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
					"integration_name": name,
					"integration_type": "prometheus",
					"config":           "{\"url\":\"http://prometheus.example.com:9090\",\"username\":\"admin\",\"password\":\"password\"}",
					"description":      "Prometheus integration for monitoring",
					"status":           "Active",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"integration_name": name,
						"integration_type": "prometheus",
						"description":      "Prometheus integration for monitoring",
						"status":           "Active",
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

var AlicloudArmsIntegrationMap0 = map[string]string{
	"status": "Active",
}

var AlicloudArmsIntegrationMap1 = map[string]string{}

func AlicloudArmsIntegrationBasicDependence0(name string) string {
	return fmt.Sprintf(`
variable "name" {
    default = "%s"
}
`, name)
}

func AlicloudArmsIntegrationBasicDependence1(name string) string {
	return fmt.Sprintf(`
variable "name" {
    default = "%s"
}
`, name)
}
