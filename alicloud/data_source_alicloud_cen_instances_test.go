package alicloud

import (
	"strings"
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"

	"fmt"
)

func TestAccAliCloudCenInstancesDataSource(t *testing.T) {
	rand := acctest.RandIntRange(1000000, 99999999)
	idsConf := dataSourceTestAccConfig{
		existConfig: testAccAliCloudCenInstancesDataSourceConfig(rand, map[string]string{
			"ids": `["${alicloud_cen_instance.default.id}"]`,
		}),
		fakeConfig: testAccAliCloudCenInstancesDataSourceConfig(rand, map[string]string{
			"ids": `["${alicloud_cen_instance.default.id}_fake"]`,
		}),
	}
	statusConf := dataSourceTestAccConfig{
		existConfig: testAccAliCloudCenInstancesDataSourceConfig(rand, map[string]string{
			"name_regex": `"${alicloud_cen_instance.default.name}"`,
			"status":     `"Active"`,
		}),
		fakeConfig: testAccAliCloudCenInstancesDataSourceConfig(rand, map[string]string{
			"name_regex": `"${alicloud_cen_instance.default.name}_fake"`,
			"status":     `"Deleting"`,
		}),
	}

	nameRegexConf := dataSourceTestAccConfig{
		existConfig: testAccAliCloudCenInstancesDataSourceConfig(rand, map[string]string{
			"name_regex": `"${alicloud_cen_instance.default.name}"`,
		}),
		fakeConfig: testAccAliCloudCenInstancesDataSourceConfig(rand, map[string]string{
			"name_regex": `"${alicloud_cen_instance.default.name}_fake"`,
		}),
	}

	tagsConf := dataSourceTestAccConfig{
		existConfig: testAccAliCloudCenInstancesDataSourceConfig(rand, map[string]string{
			"name_regex": `"${alicloud_cen_instance.default.name}"`,
			"tags": `{
							Created = "TF"
							For 	= "acceptance test"
					  }`,
		}),
		fakeConfig: testAccAliCloudCenInstancesDataSourceConfig(rand, map[string]string{
			"name_regex": `"${alicloud_cen_instance.default.name}"`,
			"tags": `{
							Created = "TF-fake"
							For 	= "acceptance test-fake"
					  }`,
		}),
	}

	allConf := dataSourceTestAccConfig{
		existConfig: testAccAliCloudCenInstancesDataSourceConfig(rand, map[string]string{
			"ids":        `["${alicloud_cen_instance.default.id}"]`,
			"name_regex": `"${alicloud_cen_instance.default.name}"`,
		}),
		fakeConfig: testAccAliCloudCenInstancesDataSourceConfig(rand, map[string]string{
			"ids":        `["${alicloud_cen_instance.default.id}"]`,
			"name_regex": `"${alicloud_cen_instance.default.name}_fake"`,
		}),
	}
	preCheck := func() {
		testAccPreCheckWithRegions(t, true, connectivity.CenNoSkipRegions)
	}
	CenInstancesCheckInfo.dataSourceTestCheckWithPreCheck(t, rand, preCheck, idsConf, statusConf, nameRegexConf, tagsConf, allConf)
}

func testAccAliCloudCenInstancesDataSourceConfig(rand int, attrMap map[string]string) string {
	var pairs []string
	for k, v := range attrMap {
		pairs = append(pairs, k+" = "+v)
	}

	config := fmt.Sprintf(`
		resource "alicloud_cen_instance" "default" {
			cen_instance_name = "tf-testAcc%sCenInstancesDataSourceCen-%d"
			description = "tf-testAccCenConfigDescription"
			tags 		= {
				Created = "TF"
				For 	= "acceptance test"
			}
		}

		data "alicloud_cen_instances" "default" {
			%s
		}
`, defaultRegionToTest, rand, strings.Join(pairs, "\n  "))
	return config
}

var existCenInstancesMapFunc = func(rand int) map[string]string {
	return map[string]string{
		"names.#":     "1",
		"instances.#": "1",
		"instances.0.cen_bandwidth_package_ids.#": "0",
		"instances.0.id":                CHECKSET,
		"instances.0.cen_id":            CHECKSET,
		"instances.0.description":       "tf-testAccCenConfigDescription",
		"instances.0.name":              fmt.Sprintf("tf-testAcc%sCenInstancesDataSourceCen-%d", defaultRegionToTest, rand),
		"instances.0.cen_instance_name": fmt.Sprintf("tf-testAcc%sCenInstancesDataSourceCen-%d", defaultRegionToTest, rand),
		"instances.0.protection_level":  "REDUCED",
		"instances.0.status":            "Active",
		"instances.0.tags.%":            "2",
		"instances.0.tags.Created":      "TF",
		"instances.0.tags.For":          "acceptance test",
		"instances.0.create_time":       CHECKSET,
	}
}

var fakeCenInstancesMapFunc = func(rand int) map[string]string {
	return map[string]string{
		"names.#":     "0",
		"instances.#": "0",
	}
}

var existCenInstancesMultiMapFunc = func(rand int) map[string]string {
	return map[string]string{
		"names.#":                 "5",
		"instances.#":             "5",
		"instances.0.name":        fmt.Sprintf("tf-testAcc%sCenInstancesDataSourceCen-%d", defaultRegionToTest, rand),
		"instances.0.description": "tf-testAccCenConfigDescription",
	}
}

var CenInstancesCheckInfo = dataSourceAttr{
	resourceId:   "data.alicloud_cen_instances.default",
	existMapFunc: existCenInstancesMapFunc,
	fakeMapFunc:  fakeCenInstancesMapFunc,
}
var CenInstancesCheckInfoMulti = dataSourceAttr{
	resourceId:   "data.alicloud_cen_instances.default",
	existMapFunc: existCenInstancesMultiMapFunc,
	fakeMapFunc:  fakeCenInstancesMapFunc,
}
