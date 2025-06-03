package alicloud

import (
	"fmt"
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

func TestAccAlicloudFlinkEnginesDataSource(t *testing.T) {
	rand := acctest.RandInt()
	resourceId := "data.alicloud_flink_engines.default"
	name := fmt.Sprintf("tf-testAccFlinkEngines%d", rand)
	testAccConfig := dataSourceTestAccConfigFunc(resourceId, name, dataSourceFlinkEnginesConfigDependence)

	checkoutSupportedRegions(t, true, connectivity.FlinkSupportRegions)
	idsConf := dataSourceTestAccConfig{
		existConfig: testAccConfig(map[string]interface{}{
			"workspace_id": "${alicloud_flink_workspace.default.id}",
		}),
		fakeConfig: testAccConfig(map[string]interface{}{
			"workspace_id": "${alicloud_flink_workspace.default.id}",
			"name_regex":   "fake_regex",
		}),
	}

	var existFlinkEnginesMapFunc = func(rand int) map[string]string {
		return map[string]string{
			"ids.#":                                 "CHECKSET",
			"names.#":                               "CHECKSET",
			"engines.#":                             "CHECKSET",
			"engines.0.engine_version":              "CHECKSET",
			"engines.0.display_name":                "CHECKSET",
			"engines.0.supported_python_versions.#": "CHECKSET",
			"engines.0.supported_scala_versions.#":  "CHECKSET",
			"engines.0.supported_java_versions.#":   "CHECKSET",
			"engines.0.features.%":                  "CHECKSET",
		}
	}

	var fakeFlinkEnginesMapFunc = func(rand int) map[string]string {
		return map[string]string{
			"ids.#":     "0",
			"names.#":   "0",
			"engines.#": "0",
		}
	}

	var FlinkEnginesCheckInfo = dataSourceAttr{
		resourceId:   resourceId,
		existMapFunc: existFlinkEnginesMapFunc,
		fakeMapFunc:  fakeFlinkEnginesMapFunc,
	}

	FlinkEnginesCheckInfo.dataSourceTestCheck(t, rand, idsConf)
}

func dataSourceFlinkEnginesConfigDependence(name string) string {
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
  name              = var.name
  resource_group_id = data.alicloud_resource_manager_resource_groups.default.ids.0
  zone_id           = data.alicloud_flink_zones.default.ids.0
  vpc_id            = alicloud_vpc.default.id
  vswitch_ids       = [alicloud_vswitch.default.id]
  
  storage {
    oss_bucket = alicloud_oss_bucket.default.bucket
  }
  
  resource {
    cpu    = "4"
    memory = "8192"
  }
}
`, name)
}
