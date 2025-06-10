package alicloud

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

func TestAccAlicloudArmsIntegrationsDataSource(t *testing.T) {
	rand := acctest.RandInt()
	idsConf := dataSourceTestAccConfig{
		existConfig: testAccCheckAlicloudArmsIntegrationsDataSourceName(rand, map[string]string{
			"ids": `["${alicloud_arms_integration.default.id}"]`,
		}),
		fakeConfig: testAccCheckAlicloudArmsIntegrationsDataSourceName(rand, map[string]string{
			"ids": `["${alicloud_arms_integration.default.id}_fake"]`,
		}),
		existChangMap: map[string]string{
			"ids.#":                           "1",
			"names.#":                         "1",
			"integrations.#":                  "1",
			"integrations.0.id":               CHECKSET,
			"integrations.0.integration_id":   CHECKSET,
			"integrations.0.integration_name": CHECKSET,
			"integrations.0.integration_type": "webhooks",
			"integrations.0.status":           "Active",
			"integrations.0.create_time":      CHECKSET,
		},
		fakeChangMap: map[string]string{
			"ids.#":          "0",
			"names.#":        "0",
			"integrations.#": "0",
		},
	}
	nameRegexConf := dataSourceTestAccConfig{
		existConfig: testAccCheckAlicloudArmsIntegrationsDataSourceName(rand, map[string]string{
			"name_regex": `"${alicloud_arms_integration.default.integration_name}"`,
		}),
		fakeConfig: testAccCheckAlicloudArmsIntegrationsDataSourceName(rand, map[string]string{
			"name_regex": `"${alicloud_arms_integration.default.integration_name}_fake"`,
		}),
		existChangMap: map[string]string{
			"ids.#":                           "1",
			"names.#":                         "1",
			"integrations.#":                  "1",
			"integrations.0.integration_name": CHECKSET,
			"integrations.0.integration_type": "webhooks",
		},
		fakeChangMap: map[string]string{
			"ids.#":          "0",
			"names.#":        "0",
			"integrations.#": "0",
		},
	}
	integrationTypeConf := dataSourceTestAccConfig{
		existConfig: testAccCheckAlicloudArmsIntegrationsDataSourceName(rand, map[string]string{
			"integration_type": `"webhooks"`,
		}),
		fakeConfig: testAccCheckAlicloudArmsIntegrationsDataSourceName(rand, map[string]string{
			"integration_type": `"prometheus"`,
		}),
		existChangMap: map[string]string{
			"integrations.#":                  CHECKSET,
			"integrations.0.integration_type": "webhooks",
		},
		fakeChangMap: map[string]string{
			"integrations.#": "0",
		},
	}
	statusConf := dataSourceTestAccConfig{
		existConfig: testAccCheckAlicloudArmsIntegrationsDataSourceName(rand, map[string]string{
			"status": `"Active"`,
		}),
		fakeConfig: testAccCheckAlicloudArmsIntegrationsDataSourceName(rand, map[string]string{
			"status": `"Inactive"`,
		}),
		existChangMap: map[string]string{
			"integrations.#":        CHECKSET,
			"integrations.0.status": "Active",
		},
		fakeChangMap: map[string]string{
			"integrations.#": "0",
		},
	}
	allConf := dataSourceTestAccConfig{
		existConfig: testAccCheckAlicloudArmsIntegrationsDataSourceName(rand, map[string]string{
			"ids":              `["${alicloud_arms_integration.default.id}"]`,
			"name_regex":       `"${alicloud_arms_integration.default.integration_name}"`,
			"integration_type": `"webhooks"`,
			"status":           `"Active"`,
		}),
		fakeConfig: testAccCheckAlicloudArmsIntegrationsDataSourceName(rand, map[string]string{
			"ids":              `["${alicloud_arms_integration.default.id}_fake"]`,
			"name_regex":       `"${alicloud_arms_integration.default.integration_name}_fake"`,
			"integration_type": `"prometheus"`,
			"status":           `"Inactive"`,
		}),
		existChangMap: map[string]string{
			"ids.#":                           "1",
			"names.#":                         "1",
			"integrations.#":                  "1",
			"integrations.0.integration_name": CHECKSET,
			"integrations.0.integration_type": "webhooks",
			"integrations.0.status":           "Active",
		},
		fakeChangMap: map[string]string{
			"ids.#":          "0",
			"names.#":        "0",
			"integrations.#": "0",
		},
	}
	var existDataAlicloudArmsIntegrationsDataSourceNameMapFunc = func(rand int) map[string]string {
		return map[string]string{
			"ids.#":                           "1",
			"names.#":                         "1",
			"integrations.#":                  "1",
			"integrations.0.integration_id":   CHECKSET,
			"integrations.0.integration_name": CHECKSET,
			"integrations.0.integration_type": "webhooks",
			"integrations.0.description":      CHECKSET,
			"integrations.0.status":           "Active",
			"integrations.0.create_time":      CHECKSET,
			"integrations.0.config":           CHECKSET,
		}
	}
	var fakeDataAlicloudArmsIntegrationsDataSourceNameMapFunc = func(rand int) map[string]string {
		return map[string]string{
			"ids.#":          "0",
			"names.#":        "0",
			"integrations.#": "0",
		}
	}
	var alicloudArmsIntegrationsCheckInfo = dataSourceAttr{
		resourceId:   "data.alicloud_arms_integrations.default",
		existMapFunc: existDataAlicloudArmsIntegrationsDataSourceNameMapFunc,
		fakeMapFunc:  fakeDataAlicloudArmsIntegrationsDataSourceNameMapFunc,
	}

	preCheck := func() {
		testAccPreCheck(t)
	}
	alicloudArmsIntegrationsCheckInfo.dataSourceTestCheckWithPreCheck(t, rand, preCheck, idsConf, nameRegexConf, integrationTypeConf, statusConf, allConf)
}
func testAccCheckAlicloudArmsIntegrationsDataSourceName(rand int, attrMap map[string]string) string {
	var pairs []string
	for k, v := range attrMap {
		pairs = append(pairs, k+" = "+v)
	}

	config := fmt.Sprintf(`

variable "name" {
	default = "tf-testAccIntegration-%d"
}

resource "alicloud_arms_integration" "default" {
  integration_name = var.name
  integration_type = "webhooks"
  config = jsonencode({
    url    = "https://example.com/webhook"
    secret = "test-secret"
  })
  description = "Test integration created by Terraform"
  status      = "Active"
}

data "alicloud_arms_integrations" "default" {	
	%s
}
`, rand, strings.Join(pairs, " \n "))
	return config
}
