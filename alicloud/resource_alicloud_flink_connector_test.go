package alicloud

import (
	"fmt"
	"testing"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAlicloudFlinkConnector_basic(t *testing.T) {
	var v map[string]interface{}
	resourceId := "alicloud_flink_connector.default"
	ra := resourceAttrInit(resourceId, testAccAlicloudFlinkConnectorBasicMap)
	rc := resourceCheckInitWithDescribeMethod(resourceId, &v, func() interface{} {
		return &FlinkService{testAccProvider.Meta().(*connectivity.AliyunClient)}
	}, "DescribeFlinkConnector")
	rac := resourceAttrCheckInit(rc, ra)
	testAccCheck := rac.resourceAttrMapUpdateSet()
	rand := acctest.RandIntRange(1000000, 9999999)
	name := fmt.Sprintf("tf-testAcc%d", rand)
	testAccConfig := resourceTestAccConfigFunc(resourceId, name, resourceFlinkConnectorConfigDependence)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckWithRegions(t, true, connectivity.FlinkRegions)
		},
		IDRefreshName: resourceId,
		Providers:     testAccProviders,
		CheckDestroy:  rac.checkResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccConfig(map[string]interface{}{
					"workspace_id": "${alicloud_flink_workspace.default.id}",
					"namespace_id": "${alicloud_flink_namespace.default.name}",
					"name":         name,
					"type":         "kafka",
					"properties": []map[string]interface{}{
						{
							"key":         "connector.version",
							"value":       "universal",
							"description": "Connector version",
						},
						{
							"key":         "connector.properties.bootstrap.servers",
							"value":       "localhost:9092",
							"description": "Kafka bootstrap servers",
						},
					},
					"dependencies": []string{
						"flink-connector-kafka-1.13.6.jar",
						"kafka-clients-2.4.1.jar",
					},
					"supported_formats": []string{
						"json",
						"csv",
					},
					"source": true,
					"sink":   true,
					"lookup": false,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"workspace_id":        CHECKSET,
						"namespace_id":        CHECKSET,
						"name":                name,
						"type":                "kafka",
						"properties.#":        "2",
						"dependencies.#":      "2",
						"supported_formats.#": "2",
						"source":              "true",
						"sink":                "true",
						"lookup":              "false",
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

var testAccAlicloudFlinkConnectorBasicMap = map[string]string{
	"workspace_id":        CHECKSET,
	"namespace_id":        CHECKSET,
	"name":                CHECKSET,
	"type":                CHECKSET,
	"properties.#":        CHECKSET,
	"dependencies.#":      CHECKSET,
	"supported_formats.#": CHECKSET,
	"source":              CHECKSET,
	"sink":                CHECKSET,
	"lookup":              CHECKSET,
}

func resourceFlinkConnectorConfigDependence(name string) string {
	return fmt.Sprintf(`
variable "name" {
  default = "%s"
}

data "alicloud_regions" "default" {
  current = true
}

resource "alicloud_flink_workspace" "default" {
  workspace_name = var.name
  region         = data.alicloud_regions.default.regions.0.id
}

resource "alicloud_flink_namespace" "default" {
  namespace_name = var.name
  workspace      = alicloud_flink_workspace.default.id
}
`, name)
}

// Verify connector exists
func (s *FlinkService) DescribeFlinkConnector(id string) (object map[string]interface{}, err error) {
	client := s.client
	flinkService := FlinkService{s.client}
	object = make(map[string]interface{})

	parts, err := ParseResourceId(id, 3)
	if err != nil {
		return object, WrapError(err)
	}
	workspace := parts[0]
	namespace := parts[1]
	name := parts[2]

	conn, err := flinkService.GetConnector(
		&workspace,
		&namespace,
		&name)
	if err != nil {
		return object, WrapError(err)
	}

	// No connector found
	if conn == nil {
		return object, WrapErrorf(Error(GetNotFoundMessage("FlinkConnector", id)), NotFoundMsg, ProviderERROR)
	}

	object["workspace_id"] = workspace
	object["namespace_id"] = namespace
	if conn.Name != nil {
		object["name"] = *conn.Name
	}
	if conn.Type != nil {
		object["type"] = *conn.Type
	}
	if conn.Lookup != nil {
		object["lookup"] = *conn.Lookup
	}
	if conn.Source != nil {
		object["source"] = *conn.Source
	}
	if conn.Sink != nil {
		object["sink"] = *conn.Sink
	}

	// Set properties
	if conn.Properties != nil && len(conn.Properties) > 0 {
		properties := make([]map[string]interface{}, 0, len(conn.Properties))
		for _, property := range conn.Properties {
			prop := make(map[string]interface{})
			if property.Key != nil {
				prop["key"] = *property.Key
			}
			if property.Value != nil {
				prop["value"] = *property.Value
			}
			if property.Description != nil {
				prop["description"] = *property.Description
			}
			properties = append(properties, prop)
		}
		object["properties"] = properties
	}

	// Set dependencies
	if conn.Dependencies != nil && len(conn.Dependencies) > 0 {
		dependencies := make([]string, 0, len(conn.Dependencies))
		for _, dep := range conn.Dependencies {
			if dep != nil {
				dependencies = append(dependencies, *dep)
			}
		}
		object["dependencies"] = dependencies
	}

	// Set supported formats
	if conn.SupportedFormats != nil && len(conn.SupportedFormats) > 0 {
		formats := make([]string, 0, len(conn.SupportedFormats))
		for _, format := range conn.SupportedFormats {
			if format != nil {
				formats = append(formats, *format)
			}
		}
		object["supported_formats"] = formats
	}

	return object, nil
}
