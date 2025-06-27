package alicloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	flinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"namespace_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"connector_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"connector_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"jar_url": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"source": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"sink": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"lookup": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"supported_formats": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"dependencies": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
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
	connectorName := d.Get("connector_name").(string)

	connector := &flinkAPI.Connector{
		Name:        connectorName,
		Type:        d.Get("connector_type").(string),
		JarUrl:      d.Get("jar_url").(string),
		Description: d.Get("description").(string),
		Source:      d.Get("source").(bool),
		Sink:        d.Get("sink").(bool),
		Lookup:      d.Get("lookup").(bool),
	}

	// Handle supported formats
	if supportedFormats := d.Get("supported_formats").([]interface{}); len(supportedFormats) > 0 {
		connector.SupportedFormats = make([]string, len(supportedFormats))
		for i, format := range supportedFormats {
			connector.SupportedFormats[i] = format.(string)
		}
	}

	// Handle dependencies
	if dependencies := d.Get("dependencies").([]interface{}); len(dependencies) > 0 {
		connector.Dependencies = make([]string, len(dependencies))
		for i, dep := range dependencies {
			connector.Dependencies[i] = dep.(string)
		}
	}

	// Create connector
	_, err = flinkService.RegisterCustomConnector(workspaceId, namespaceName, connector)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_connector", "RegisterCustomConnector", AlibabaCloudSdkGoERROR)
	}

	d.SetId(workspaceId + ":" + namespaceName + ":" + connectorName)

	// Wait for connector to be available
	stateConf := BuildStateConf([]string{"CREATING"}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, flinkService.FlinkConnectorStateRefreshFunc(workspaceId, namespaceName, connectorName, []string{"FAILED"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudFlinkConnectorRead(d, meta)
}

func resourceAliCloudFlinkConnectorRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId, namespaceName, connectorName, err := parseConnectorResourceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	connector, err := flinkService.GetConnector(workspaceId, namespaceName, connectorName)
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_flink_connector GetConnector Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("workspace_id", workspaceId)
	d.Set("namespace_name", namespaceName)
	d.Set("connector_name", connector.Name)
	d.Set("connector_type", connector.Type)
	d.Set("jar_url", connector.JarUrl)
	d.Set("description", connector.Description)
	d.Set("source", connector.Source)
	d.Set("sink", connector.Sink)
	d.Set("lookup", connector.Lookup)
	d.Set("supported_formats", connector.SupportedFormats)
	d.Set("dependencies", connector.Dependencies)

	return nil
}

func resourceAliCloudFlinkConnectorUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId, namespaceName, connectorName, err := parseConnectorResourceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	update := false
	connector := &flinkAPI.Connector{
		Name: connectorName,
	}

	if d.HasChange("jar_url") {
		connector.JarUrl = d.Get("jar_url").(string)
		update = true
	}

	if d.HasChange("description") {
		connector.Description = d.Get("description").(string)
		update = true
	}

	if d.HasChange("source") {
		connector.Source = d.Get("source").(bool)
		update = true
	}

	if d.HasChange("sink") {
		connector.Sink = d.Get("sink").(bool)
		update = true
	}

	if d.HasChange("lookup") {
		connector.Lookup = d.Get("lookup").(bool)
		update = true
	}

	if d.HasChange("supported_formats") {
		if supportedFormats := d.Get("supported_formats").([]interface{}); len(supportedFormats) > 0 {
			connector.SupportedFormats = make([]string, len(supportedFormats))
			for i, format := range supportedFormats {
				connector.SupportedFormats[i] = format.(string)
			}
		}
		update = true
	}

	if d.HasChange("dependencies") {
		if dependencies := d.Get("dependencies").([]interface{}); len(dependencies) > 0 {
			connector.Dependencies = make([]string, len(dependencies))
			for i, dep := range dependencies {
				connector.Dependencies[i] = dep.(string)
			}
		}
		update = true
	}

	if update {
		_, err := flinkService.RegisterCustomConnector(workspaceId, namespaceName, connector)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "RegisterCustomConnector", AlibabaCloudSdkGoERROR)
		}

		// Wait for update to complete
		stateConf := BuildStateConf([]string{"MODIFYING"}, []string{"Available"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, flinkService.FlinkConnectorStateRefreshFunc(workspaceId, namespaceName, connectorName, []string{"FAILED"}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	return resourceAliCloudFlinkConnectorRead(d, meta)
}

func resourceAliCloudFlinkConnectorDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId, namespaceName, connectorName, err := parseConnectorResourceId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	err = flinkService.DeleteCustomConnector(workspaceId, namespaceName, connectorName)
	if err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteCustomConnector", AlibabaCloudSdkGoERROR)
	}

	// Wait for connector to be deleted
	stateConf := &resource.StateChangeConf{
		Pending: []string{"DELETING"},
		Target:  []string{""},
		Refresh: func() (interface{}, string, error) {
			connector, err := flinkService.GetConnector(workspaceId, namespaceName, connectorName)
			if err != nil {
				if NotFoundError(err) {
					return nil, "", nil
				}
				return nil, "", WrapError(err)
			}
			return connector, "DELETING", nil
		},
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}

// Helper function to parse connector resource ID
func parseConnectorResourceId(id string) (string, string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid connector resource ID format: %s, expected workspace_id:namespace_name:connector_name", id)
	}
	return parts[0], parts[1], parts[2], nil
}
