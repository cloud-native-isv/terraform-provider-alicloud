package alicloud

import (
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api"
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
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"namespace_name": {
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

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)
	name := d.Get("name").(string)
	connectorType := d.Get("type").(string)

	connector := &aliyunAPI.Connector{
		Name: name,
		Type: connectorType,
	}

	// Handle properties
	if properties, ok := d.GetOk("properties"); ok {
		propertyList := properties.([]interface{})
		connector.Properties = make([]*aliyunAPI.Property, len(propertyList))
		for i, prop := range propertyList {
			propMap := prop.(map[string]interface{})
			connector.Properties[i] = &aliyunAPI.Property{
				Key:         propMap["key"].(string),
				Description: propMap["description"].(string),
			}
		}
	}

	// Handle dependencies
	if dependencies, ok := d.GetOk("dependencies"); ok {
		depList := dependencies.([]interface{})
		connector.Dependencies = make([]string, len(depList))
		for i, dep := range depList {
			connector.Dependencies[i] = dep.(string)
		}
	}

	// Handle supported formats
	if supportedFormats, ok := d.GetOk("supported_formats"); ok {
		formatList := supportedFormats.([]interface{})
		connector.SupportedFormats = make([]string, len(formatList))
		for i, format := range formatList {
			connector.SupportedFormats[i] = format.(string)
		}
	}

	// Handle flags
	connector.Lookup = d.Get("lookup").(bool)
	connector.Source = d.Get("source").(bool)
	connector.Sink = d.Get("sink").(bool)

	// Register connector
	var response *aliyunAPI.Connector
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		resp, err := flinkService.RegisterCustomConnector(workspaceId, namespaceName, connector)
		if err != nil {
			if IsExpectedErrors(err, []string{"ThrottlingException", "OperationConflict"}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		response = resp
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_connector", "RegisterCustomConnector", AlibabaCloudSdkGoERROR)
	}

	if response == nil || response.Name == "" {
		return WrapError(Error("Failed to get connector name from response"))
	}

	// Set composite ID: workspace:namespace:connector_name
	d.SetId(workspaceId + ":" + namespaceName + ":" + response.Name)

	// Wait for connector registration to complete using StateRefreshFunc
	stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, flinkService.FlinkConnectorStateRefreshFunc(workspaceId, namespaceName, response.Name, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// 最后调用Read同步状态
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

	workspaceId := parts[0]
	namespaceName := parts[1]
	name := parts[2]

	connector, err := flinkService.GetConnector(workspaceId, namespaceName, name)
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("workspace_id", workspaceId)
	d.Set("namespace_name", namespaceName)
	d.Set("name", name)
	d.Set("type", connector.Type)
	d.Set("creator", connector.Creator)
	d.Set("description", connector.Description)

	// Set dependencies
	if connector.Dependencies != nil && len(connector.Dependencies) > 0 {
		d.Set("dependencies", connector.Dependencies)
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

	workspaceId := parts[0]
	namespaceName := parts[1]
	name := parts[2]

	d.Partial(true)

	// Determine if any updateable fields have changed
	updateConnector := false
	if d.HasChanges("properties", "dependencies", "supported_formats", "lookup", "source", "sink") {
		updateConnector = true
	}

	if updateConnector {
		// First delete existing connector
		err = flinkService.DeleteCustomConnector(workspaceId, namespaceName, name)
		if err != nil && !NotFoundError(err) {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteCustomConnector", AliyunLogGoSdkERROR)
		}

		// Wait for deletion to complete
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err := flinkService.GetConnector(workspaceId, namespaceName, name)
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

		// Create updated connector object
		connector := &aliyunAPI.Connector{
			Name: name,
			Type: d.Get("type").(string),
		}

		// Handle dependencies if provided
		if deps, ok := d.GetOk("dependencies"); ok {
			dependencies := make([]string, 0)
			for _, dep := range deps.([]interface{}) {
				dependencies = append(dependencies, dep.(string))
			}
			connector.Dependencies = dependencies
		}

		// Register the connector again with updated properties
		_, err := flinkService.RegisterCustomConnector(workspaceId, namespaceName, connector)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_connector", "RegisterCustomConnector", AliyunLogGoSdkERROR)
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

	workspaceId := parts[0]
	namespaceName := parts[1]
	name := parts[2]

	err = flinkService.DeleteCustomConnector(workspaceId, namespaceName, name)
	if err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteCustomConnector", AliyunLogGoSdkERROR)
	}

	return WrapError(resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := flinkService.GetConnector(workspaceId, namespaceName, name)
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
