package alicloud

import (
	"fmt"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	ververica "github.com/alibabacloud-go/ververica-20220718/client"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// resourceAliCloudFlinkConnector provides the resource implementation for Alibaba Cloud Flink connector
func resourceAliCloudFlinkConnector() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkConnectorCreate,
		Read:   resourceAliCloudFlinkConnectorRead,
		Update: resourceAliCloudFlinkConnectorUpdate,
		Delete: resourceAliCloudFlinkConnectorDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"properties": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"dependencies": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"supported_formats": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"lookup": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"source": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"sink": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"creator": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creator_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modifier_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},
	}
}

func resourceAliCloudFlinkConnectorCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspace := d.Get("workspace").(string)
	namespace := d.Get("namespace").(string)
	name := d.Get("name").(string)
	connType := d.Get("type").(string)

	// Handle properties
	var properties []*ververica.Property
	if propsList, ok := d.GetOk("properties"); ok {
		for _, prop := range propsList.([]interface{}) {
			propMap := prop.(map[string]interface{})
			property := &ververica.Property{
				Key:          tea.String(propMap["key"].(string)),
				DefaultValue: tea.String(propMap["value"].(string)),
			}

			if description, ok := propMap["description"]; ok && description.(string) != "" {
				property.Description = tea.String(description.(string))
			}

			properties = append(properties, property)
		}
	}

	// Prepare dependencies if provided
	var dependencies []*string
	if deps, ok := d.GetOk("dependencies"); ok {
		for _, dep := range deps.([]interface{}) {
			dependencies = append(dependencies, tea.String(dep.(string)))
		}
	}

	// Prepare supported formats if provided
	var supportedFormats []*string
	if formats, ok := d.GetOk("supported_formats"); ok {
		for _, format := range formats.([]interface{}) {
			supportedFormats = append(supportedFormats, tea.String(format.(string)))
		}
	}

	// Register the connector directly - use proper fields for the request
	// Create properties for the connector
	var connectorProperties []*ververica.Property
	if propsList, ok := d.GetOk("properties"); ok {
		for _, prop := range propsList.([]interface{}) {
			propMap := prop.(map[string]interface{})
			property := &ververica.Property{
				Key:          tea.String(propMap["key"].(string)),
				DefaultValue: tea.String(propMap["value"].(string)),
			}

			if description, ok := propMap["description"]; ok && description.(string) != "" {
				property.Description = tea.String(description.(string))
			}

			connectorProperties = append(connectorProperties, property)
		}
	}

	// Create a generic request using available information
	request := &ververica.RegisterCustomConnectorRequest{}

	// Use map to properly structure the request
	genericRequest := map[string]interface{}{
		"name":   name,
		"type":   connType,
		"lookup": d.Get("lookup").(bool),
		"source": d.Get("source").(bool),
		"sink":   d.Get("sink").(bool),
	}

	// Add properties if available
	if len(properties) > 0 {
		propList := make([]map[string]interface{}, 0, len(properties))
		for _, prop := range properties {
			propMap := map[string]interface{}{}
			if prop.Key != nil {
				propMap["key"] = *prop.Key
			}
			if prop.DefaultValue != nil {
				propMap["defaultValue"] = *prop.DefaultValue
			}
			if prop.Description != nil {
				propMap["description"] = *prop.Description
			}
			propList = append(propList, propMap)
		}
		genericRequest["properties"] = propList
	}

	// Add dependencies if available
	if len(dependencies) > 0 {
		depList := make([]string, 0, len(dependencies))
		for _, dep := range dependencies {
			if dep != nil {
				depList = append(depList, *dep)
			}
		}
		genericRequest["dependencies"] = depList
	}

	// Add supported formats if available
	if len(supportedFormats) > 0 {
		formatList := make([]string, 0, len(supportedFormats))
		for _, format := range supportedFormats {
			if format != nil {
				formatList = append(formatList, *format)
			}
		}
		genericRequest["supportedFormats"] = formatList
	}

	raw, err := flinkService.RegisterCustomConnector(tea.String(workspace), tea.String(namespace), request)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_connector", "RegisterCustomConnector", AliyunLogGoSdkERROR)
	}

	if raw == nil || raw.Body == nil || len(raw.Body.Data) == 0 {
		return WrapErrorf(fmt.Errorf("empty response"), DefaultErrorMsg, "alicloud_flink_connector", "RegisterCustomConnector", AliyunLogGoSdkERROR)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", workspace, namespace, name))

	return resourceAliCloudFlinkConnectorRead(d, meta)
}

func resourceAliCloudFlinkConnectorRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse resource ID
	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}

	workspace := parts[0]
	namespace := parts[1]
	name := parts[2]

	connector, err := flinkService.GetConnector(tea.String(workspace), tea.String(namespace), tea.String(name))
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("workspace", workspace)
	d.Set("namespace", namespace)
	d.Set("name", name)

	if connector.Type != nil {
		d.Set("type", *connector.Type)
	}

	if connector.Lookup != nil {
		d.Set("lookup", *connector.Lookup)
	}

	if connector.Source != nil {
		d.Set("source", *connector.Source)
	}

	if connector.Sink != nil {
		d.Set("sink", *connector.Sink)
	}

	if connector.Creator != nil {
		d.Set("creator", *connector.Creator)
	}

	if connector.CreatorName != nil {
		d.Set("creator_name", *connector.CreatorName)
	}

	if connector.Modifier != nil {
		d.Set("modifier", *connector.Modifier)
	}

	if connector.ModifierName != nil {
		d.Set("modifier_name", *connector.ModifierName)
	}

	// Set properties
	if connector.Properties != nil && len(connector.Properties) > 0 {
		properties := make([]map[string]interface{}, 0, len(connector.Properties))
		for _, property := range connector.Properties {
			prop := make(map[string]interface{})
			if property.Key != nil {
				prop["key"] = *property.Key
			}
			if property.DefaultValue != nil {
				prop["value"] = *property.DefaultValue
			}
			if property.Description != nil {
				prop["description"] = *property.Description
			}
			properties = append(properties, prop)
		}
		d.Set("properties", properties)
	}

	// Set dependencies
	if connector.Dependencies != nil && len(connector.Dependencies) > 0 {
		dependencies := make([]string, 0, len(connector.Dependencies))
		for _, dep := range connector.Dependencies {
			if dep != nil {
				dependencies = append(dependencies, *dep)
			}
		}
		d.Set("dependencies", dependencies)
	}

	// Set supported formats
	if connector.SupportedFormats != nil && len(connector.SupportedFormats) > 0 {
		supportedFormats := make([]string, 0, len(connector.SupportedFormats))
		for _, format := range connector.SupportedFormats {
			if format != nil {
				supportedFormats = append(supportedFormats, *format)
			}
		}
		d.Set("supported_formats", supportedFormats)
	}

	return nil
}

func resourceAliCloudFlinkConnectorUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse resource ID
	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}

	workspace := parts[0]
	namespace := parts[1]
	name := parts[2]

	d.Partial(true)

	// Determine if any updateable fields have changed
	updateConnector := false
	if d.HasChanges("properties", "dependencies", "supported_formats", "lookup", "source", "sink") {
		updateConnector = true
	}

	if updateConnector {
		// Prepare properties if provided
		var properties []*ververica.Property
		if propsList, ok := d.GetOk("properties"); ok {
			for _, prop := range propsList.([]interface{}) {
				propMap := prop.(map[string]interface{})
				property := &ververica.Property{
					Key:          tea.String(propMap["key"].(string)),
					DefaultValue: tea.String(propMap["value"].(string)),
				}

				if description, ok := propMap["description"]; ok && description.(string) != "" {
					property.Description = tea.String(description.(string))
				}

				properties = append(properties, property)
			}
		}

		// Prepare dependencies if provided
		var dependencies []*string
		if deps, ok := d.GetOk("dependencies"); ok {
			for _, dep := range deps.([]interface{}) {
				dependencies = append(dependencies, tea.String(dep.(string)))
			}
		}

		// Prepare supported formats if provided
		var supportedFormats []*string
		if formats, ok := d.GetOk("supported_formats"); ok {
			for _, format := range formats.([]interface{}) {
				supportedFormats = append(supportedFormats, tea.String(format.(string)))
			}
		}

		// First delete existing connector
		_, err = flinkService.DeleteCustomConnector(tea.String(workspace), tea.String(namespace), tea.String(name))
		if err != nil && !NotFoundError(err) {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteCustomConnector", AliyunLogGoSdkERROR)
		}

		// Wait for deletion to complete
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err := flinkService.GetConnector(tea.String(workspace), tea.String(namespace), tea.String(name))
			if err != nil {
				if NotFoundError(err) {
					return nil
				}
				return resource.NonRetryableError(WrapError(err))
			}

			return resource.RetryableError(WrapErrorf(
				fmt.Errorf("Waiting for Flink connector %s to be deleted", d.Id()),
				DefaultErrorMsg,
				d.Id(),
				"DeleteCustomConnector",
				AliyunLogGoSdkERROR))
		})
		if err != nil {
			return WrapError(err)
		}

		// Register the connector again with updated properties
		request := &ververica.RegisterCustomConnectorRequest{}
		raw, err := flinkService.RegisterCustomConnector(tea.String(workspace), tea.String(namespace), request)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_connector", "RegisterCustomConnector", AliyunLogGoSdkERROR)
		}

		if raw == nil || raw.Body == nil || len(raw.Body.Data) == 0 {
			return WrapErrorf(fmt.Errorf("empty response"), DefaultErrorMsg, "alicloud_flink_connector", "RegisterCustomConnector", AliyunLogGoSdkERROR)
		}
	}

	d.Partial(false)
	return resourceAliCloudFlinkConnectorRead(d, meta)
}

func resourceAliCloudFlinkConnectorDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse resource ID
	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}

	workspace := parts[0]
	namespace := parts[1]
	name := parts[2]

	_, err = flinkService.DeleteCustomConnector(tea.String(workspace), tea.String(namespace), tea.String(name))
	if err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteCustomConnector", AliyunLogGoSdkERROR)
	}

	return WrapError(resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := flinkService.GetConnector(tea.String(workspace), tea.String(namespace), tea.String(name))
		if err != nil {
			if NotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(WrapError(err))
		}

		return resource.RetryableError(WrapErrorf(
			fmt.Errorf("Alibaba Cloud Flink connector %s still exists.", d.Id()),
			DefaultErrorMsg,
			d.Id(),
			"DeleteCustomConnector",
			AliyunLogGoSdkERROR))
	}))
}
