package alicloud

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

func TestAccAliCloudCenTransitRouteTableAggregationsDataSource(t *testing.T) {
	rand := acctest.RandInt()
	idsConf := dataSourceTestAccConfig{
		existConfig: testAccCheckAliCloudCenTransitRouteTableAggregationsDataSourceName(rand, map[string]string{
			"ids": `["${alicloud_cen_transit_route_table_aggregation.default.id}"]`,
		}),
		fakeConfig: testAccCheckAliCloudCenTransitRouteTableAggregationsDataSourceName(rand, map[string]string{
			"ids": `["${alicloud_cen_transit_route_table_aggregation.default.id}_fake"]`,
		}),
	}
	nameRegexConf := dataSourceTestAccConfig{
		existConfig: testAccCheckAliCloudCenTransitRouteTableAggregationsDataSourceName(rand, map[string]string{
			"name_regex": `"${alicloud_cen_transit_route_table_aggregation.default.transit_route_table_aggregation_name}"`,
		}),
		fakeConfig: testAccCheckAliCloudCenTransitRouteTableAggregationsDataSourceName(rand, map[string]string{
			"name_regex": `"${alicloud_cen_transit_route_table_aggregation.default.transit_route_table_aggregation_name}_fake"`,
		}),
	}
	transitRouteTableAggregationCidrConf := dataSourceTestAccConfig{
		existConfig: testAccCheckAliCloudCenTransitRouteTableAggregationsDataSourceName(rand, map[string]string{
			"transit_route_table_aggregation_cidr": `"${alicloud_cen_transit_route_table_aggregation.default.transit_route_table_aggregation_cidr}"`,
		}),
		fakeConfig: testAccCheckAliCloudCenTransitRouteTableAggregationsDataSourceName(rand, map[string]string{
			"transit_route_table_aggregation_cidr": `"10.0.0.0/9"`,
		}),
	}
	statusConf := dataSourceTestAccConfig{
		existConfig: testAccCheckAliCloudCenTransitRouteTableAggregationsDataSourceName(rand, map[string]string{
			"status": `"AllConfigured"`,
		}),
		fakeConfig: testAccCheckAliCloudCenTransitRouteTableAggregationsDataSourceName(rand, map[string]string{
			"status": `"ConfigFailed"`,
		}),
	}
	allConf := dataSourceTestAccConfig{
		existConfig: testAccCheckAliCloudCenTransitRouteTableAggregationsDataSourceName(rand, map[string]string{
			"ids":                                  `["${alicloud_cen_transit_route_table_aggregation.default.id}"]`,
			"name_regex":                           `"${alicloud_cen_transit_route_table_aggregation.default.transit_route_table_aggregation_name}"`,
			"transit_route_table_aggregation_cidr": `"${alicloud_cen_transit_route_table_aggregation.default.transit_route_table_aggregation_cidr}"`,
			"status":                               `"AllConfigured"`,
		}),
		fakeConfig: testAccCheckAliCloudCenTransitRouteTableAggregationsDataSourceName(rand, map[string]string{
			"ids":                                  `["${alicloud_cen_transit_route_table_aggregation.default.id}_fake"]`,
			"name_regex":                           `"${alicloud_cen_transit_route_table_aggregation.default.transit_route_table_aggregation_name}_fake"`,
			"transit_route_table_aggregation_cidr": `"10.0.0.0/9"`,
			"status":                               `"ConfigFailed"`,
		}),
	}
	var existAliCloudCenTransitRouteTableAggregationsDataSourceNameMapFunc = func(rand int) map[string]string {
		return map[string]string{
			"ids.#":                                 "1",
			"names.#":                               "1",
			"transit_route_table_aggregations.#":    "1",
			"transit_route_table_aggregations.0.id": CHECKSET,
			"transit_route_table_aggregations.0.transit_route_table_id":                      CHECKSET,
			"transit_route_table_aggregations.0.transit_route_table_aggregation_cidr":        "10.0.0.0/8",
			"transit_route_table_aggregations.0.transit_route_table_aggregation_scope":       "VPC",
			"transit_route_table_aggregations.0.route_type":                                  "Static",
			"transit_route_table_aggregations.0.transit_route_table_aggregation_name":        CHECKSET,
			"transit_route_table_aggregations.0.transit_route_table_aggregation_description": CHECKSET,
			"transit_route_table_aggregations.0.status":                                      "AllConfigured",
		}
	}
	var fakeAliCloudCenTransitRouteTableAggregationsDataSourceNameMapFunc = func(rand int) map[string]string {
		return map[string]string{
			"ids.#":                              "0",
			"names.#":                            "0",
			"transit_route_table_aggregations.#": "0",
		}
	}
	var alicloudCenTransitRouteTableAggregationsCheckInfo = dataSourceAttr{
		resourceId:   "data.alicloud_cen_transit_route_table_aggregations.default",
		existMapFunc: existAliCloudCenTransitRouteTableAggregationsDataSourceNameMapFunc,
		fakeMapFunc:  fakeAliCloudCenTransitRouteTableAggregationsDataSourceNameMapFunc,
	}
	preCheck := func() {
		testAccPreCheck(t)
	}
	alicloudCenTransitRouteTableAggregationsCheckInfo.dataSourceTestCheckWithPreCheck(t, rand, preCheck, idsConf, nameRegexConf, transitRouteTableAggregationCidrConf, statusConf, allConf)
}

func testAccCheckAliCloudCenTransitRouteTableAggregationsDataSourceName(rand int, attrMap map[string]string) string {
	var pairs []string
	for k, v := range attrMap {
		pairs = append(pairs, k+" = "+v)
	}

	config := fmt.Sprintf(`
	variable "name" {
  		default = "tf-testAccCenTransitRouteTableAggregation-%d"
	}

	resource "alicloud_cen_instance" "default" {
  		cen_instance_name = var.name
	}

	resource "alicloud_cen_transit_router" "default" {
  		cen_id = alicloud_cen_instance.default.id
	}

	resource "alicloud_cen_transit_router_route_table" "default" {
  		transit_router_id = alicloud_cen_transit_router.default.transit_router_id
	}

	resource "alicloud_cen_transit_route_table_aggregation" "default" {
  		transit_route_table_id                      = alicloud_cen_transit_router_route_table.default.transit_router_route_table_id
  		transit_route_table_aggregation_cidr        = "10.0.0.0/8"
  		transit_route_table_aggregation_scope       = "VPC"
  		transit_route_table_aggregation_name        = var.name
  		transit_route_table_aggregation_description = var.name
	}

	data "alicloud_cen_transit_route_table_aggregations" "default" {
  		transit_route_table_id = alicloud_cen_transit_route_table_aggregation.default.transit_route_table_id
		%s
	}
`, rand, strings.Join(pairs, " \n "))
	return config
}
