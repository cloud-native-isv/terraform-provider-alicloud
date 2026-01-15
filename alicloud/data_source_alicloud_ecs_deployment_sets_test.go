package alicloud

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

func TestAccAliCloudECSDeploymentSetsDataSource(t *testing.T) {
	rand := acctest.RandInt()
	idsConf := dataSourceTestAccConfig{
		existConfig: testAccCheckAliCloudEcsDeploymentSetsDataSourceName(rand, map[string]string{
			"ids": `["${alicloud_ecs_deployment_set.default.id}"]`,
		}),
		fakeConfig: testAccCheckAliCloudEcsDeploymentSetsDataSourceName(rand, map[string]string{
			"ids": `["${alicloud_ecs_deployment_set.default.id}_fake"]`,
		}),
	}
	deploymentSetNameConf := dataSourceTestAccConfig{
		existConfig: testAccCheckAliCloudEcsDeploymentSetsDataSourceName(rand, map[string]string{
			"ids":                 `["${alicloud_ecs_deployment_set.default.id}"]`,
			"deployment_set_name": `"${alicloud_ecs_deployment_set.default.deployment_set_name}"`,
		}),
		fakeConfig: testAccCheckAliCloudEcsDeploymentSetsDataSourceName(rand, map[string]string{
			"ids":                 `["${alicloud_ecs_deployment_set.default.id}"]`,
			"deployment_set_name": `"${alicloud_ecs_deployment_set.default.deployment_set_name}_fake"`,
		}),
	}
	strategyConf := dataSourceTestAccConfig{
		existConfig: testAccCheckAliCloudEcsDeploymentSetsDataSourceName(rand, map[string]string{
			"ids":      `["${alicloud_ecs_deployment_set.default.id}"]`,
			"strategy": `"Availability"`,
		}),
		fakeConfig: testAccCheckAliCloudEcsDeploymentSetsDataSourceName(rand, map[string]string{
			"ids":      `["${alicloud_ecs_deployment_set.default.id}_fake"]`,
			"strategy": `"Availability"`,
		}),
	}
	nameRegexConf := dataSourceTestAccConfig{
		existConfig: testAccCheckAliCloudEcsDeploymentSetsDataSourceName(rand, map[string]string{
			"name_regex": `"${alicloud_ecs_deployment_set.default.deployment_set_name}"`,
		}),
		fakeConfig: testAccCheckAliCloudEcsDeploymentSetsDataSourceName(rand, map[string]string{
			"name_regex": `"${alicloud_ecs_deployment_set.default.deployment_set_name}_fake"`,
		}),
	}
	allConf := dataSourceTestAccConfig{
		existConfig: testAccCheckAliCloudEcsDeploymentSetsDataSourceName(rand, map[string]string{
			"deployment_set_name": `"${alicloud_ecs_deployment_set.default.deployment_set_name}"`,
			"ids":                 `["${alicloud_ecs_deployment_set.default.id}"]`,
			"name_regex":          `"${alicloud_ecs_deployment_set.default.deployment_set_name}"`,
			"strategy":            `"Availability"`,
		}),
		fakeConfig: testAccCheckAliCloudEcsDeploymentSetsDataSourceName(rand, map[string]string{
			"deployment_set_name": `"${alicloud_ecs_deployment_set.default.deployment_set_name}_fake"`,
			"ids":                 `["${alicloud_ecs_deployment_set.default.id}_fake"]`,
			"name_regex":          `"${alicloud_ecs_deployment_set.default.deployment_set_name}_fake"`,
			"strategy":            `"Availability"`,
		}),
	}
	var existDataAliCloudAlbAclsSourceNameMapFunc = func(rand int) map[string]string {
		return map[string]string{
			"ids.#":                      "1",
			"names.#":                    "1",
			"sets.#":                     "1",
			"sets.0.deployment_set_name": fmt.Sprintf("tf-testAccDeploymentSet-%d", rand),
			"sets.0.strategy":            "Availability",
			"sets.0.domain":              "Default",
			"sets.0.granularity":         "Host",
		}
	}
	var fakeAliCloudEcsDeploymentSetsDataSourceNameMapFunc = func(rand int) map[string]string {
		return map[string]string{
			"ids.#":   "0",
			"names.#": "0",
		}
	}
	var alicloudEcsDeploymentSetsCheckInfo = dataSourceAttr{
		resourceId:   "data.alicloud_ecs_deployment_sets.default",
		existMapFunc: existDataAliCloudAlbAclsSourceNameMapFunc,
		fakeMapFunc:  fakeAliCloudEcsDeploymentSetsDataSourceNameMapFunc,
	}
	alicloudEcsDeploymentSetsCheckInfo.dataSourceTestCheck(t, rand, idsConf, deploymentSetNameConf, strategyConf, nameRegexConf, allConf)
}
func testAccCheckAliCloudEcsDeploymentSetsDataSourceName(rand int, attrMap map[string]string) string {
	var pairs []string
	for k, v := range attrMap {
		pairs = append(pairs, k+" = "+v)
	}

	config := fmt.Sprintf(`

variable "name" {	
	default = "tf-testAccDeploymentSet-%d"
}

resource "alicloud_ecs_deployment_set" "default" {
  strategy            = "Availability"
  domain              = "Default"
  granularity         = "Host"
  deployment_set_name = var.name
  description         = var.name
}

data "alicloud_ecs_deployment_sets" "default" {	
	%s
}
`, rand, strings.Join(pairs, " \n "))
	return config
}
