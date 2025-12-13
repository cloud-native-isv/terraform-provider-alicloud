package alicloud

import (
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudSelectDBPublicConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudSelectDBPublicConnectionCreate,
		Read:   resourceAliCloudSelectDBPublicConnectionRead,
		Delete: resourceAliCloudSelectDBPublicConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the SelectDB instance for which to allocate the public connection.",
			},
			"connection_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of all connections for the SelectDB instance.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"net_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Network type of the connection (VPC or Public).",
						},
						"protocol": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Protocol type of the connection (MySQLPort or HttpPort).",
						},
						"hostname": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Hostname of the connection.",
						},
						"port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Port number of the connection.",
						},
					},
				},
			},
		},
	}
}

func resourceAliCloudSelectDBPublicConnectionCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	// Create public connection object from schema data
	connection := ConvertToPublicConnection(d, service)

	// Use resource.Retry for creation to handle temporary failures
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := service.AllocateSelectDBPublicConnection(connection)
		if err != nil {
			if NeedRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_selectdb_public_connection", "AllocatePublicConnection", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID using instance ID
	d.SetId(connection.DBInstanceId)

	// Wait for public connection to be available
	err = service.WaitForSelectDBPublicConnection(d.Id(), Available, int(d.Timeout(schema.TimeoutCreate).Seconds()))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudSelectDBPublicConnectionRead(d, meta)
}

func resourceAliCloudSelectDBPublicConnectionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Id()

	// Get public connection information for the instance
	publicConnection, err := service.DescribeSelectDBPublicConnection(instanceId)
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_selectdb_public_connection not found!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DescribeSelectDBPublicConnection", AlibabaCloudSdkGoERROR)
	}

	if publicConnection == nil {
		if !d.IsNewResource() {
			log.Printf("[DEBUG] SelectDB public connection for instance %s not found, removing from state", instanceId)
			d.SetId("")
			return nil
		}
		return WrapErrorf(Error(GetNotFoundMessage("SelectDB Public Connection", instanceId)), NotFoundMsg, ProviderERROR)
	}

	// Set basic attributes
	d.Set("instance_id", instanceId)

	// Get all connections (VPC and Public, MySQL and HTTP)
	allConnections, err := service.GetSelectDBAllConnections(instanceId)
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] No connections found for instance %s, removing from state", instanceId)
			d.SetId("")
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "GetSelectDBAllConnections", AlibabaCloudSdkGoERROR)
	}

	// Convert connections to map format for Terraform
	var connectionList []map[string]interface{}
	for _, conn := range allConnections {
		connectionMap := map[string]interface{}{
			"net_type": conn.NetType,
			"protocol": conn.Protocol,
			"hostname": conn.Hostname,
			"port":     int(conn.Port),
		}
		connectionList = append(connectionList, connectionMap)
	}

	// Set the connection list
	if err := d.Set("connection_list", connectionList); err != nil {
		return WrapError(err)
	}

	return nil
}

func resourceAliCloudSelectDBPublicConnectionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Id()

	// Get the current public connection to use for release operation
	publicConnection, err := service.DescribeSelectDBPublicConnection(instanceId)
	if err != nil {
		if NotFoundError(err) {
			return nil // No public connection to release
		}
		return WrapError(err)
	}

	if publicConnection == nil {
		// No public connection to release
		return nil
	}

	// Use the public connection object directly for release
	connection := &selectdb.PublicConnection{
		DBInstanceId: instanceId,
		NetType:      publicConnection.NetType,
		Protocol:     publicConnection.Protocol,
		Hostname:     publicConnection.Hostname,
		Port:         publicConnection.Port,
		RegionId:     service.GetRegionId(),
	}

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := service.ReleaseSelectDBPublicConnection(connection)
		if err != nil {
			if NotFoundError(err) {
				return nil
			}
			if NeedRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ReleasePublicConnection", AlibabaCloudSdkGoERROR)
	}

	// Wait for the public connection to be deleted
	err = service.WaitForSelectDBPublicConnection(instanceId, Deleted, int(d.Timeout(schema.TimeoutDelete).Seconds()))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
