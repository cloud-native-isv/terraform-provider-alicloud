package alicloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudSelectDBClusterBinding() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudSelectDBClusterBindingCreate,
		Read:   resourceAliCloudSelectDBClusterBindingRead,
		Delete: resourceAliCloudSelectDBClusterBindingDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			// ======== Required Parameters ========
			"cluster_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the SelectDB cluster to bind.",
			},
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the SelectDB instance.",
			},

			// ======== Optional Parameters ========
			"cluster_id_bak": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The backup cluster ID for binding.",
			},

			// ======== Computed Information ========
			"instance_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the SelectDB instance.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the cluster binding.",
			},
		},
	}
}

func resourceAliCloudSelectDBClusterBindingCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	clusterId := d.Get("cluster_id").(string)
	instanceId := d.Get("instance_id").(string)

	var clusterIdBak []string
	if bakId := d.Get("cluster_id_bak").(string); bakId != "" {
		clusterIdBak = append(clusterIdBak, bakId)
	}

	// Use resource.Retry for creation to handle temporary failures
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := service.CreateSelectDBClusterBinding(clusterId, instanceId, clusterIdBak...)
		if err != nil {
			if NeedRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_selectdb_cluster_binding", "CreateClusterBinding", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID using cluster:instance format
	d.SetId(fmt.Sprintf("%s:%s", clusterId, instanceId))

	return resourceAliCloudSelectDBClusterBindingRead(d, meta)
}

func resourceAliCloudSelectDBClusterBindingRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse resource ID
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return WrapError(fmt.Errorf("invalid resource ID format: %s", d.Id()))
	}
	clusterId := parts[0]
	instanceId := parts[1]

	// Since there's no direct API to describe cluster binding,
	// we verify the binding exists by checking if the cluster and instance exist
	// and are accessible together
	cluster, err := service.DescribeSelectDBCluster(instanceId, clusterId)
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_selectdb_cluster_binding not found!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DescribeSelectDBCluster", AlibabaCloudSdkGoERROR)
	}

	// Set computed attributes
	d.Set("cluster_id", clusterId)
	d.Set("instance_id", instanceId)

	if cluster.Status != "" {
		d.Set("status", cluster.Status)
	}

	// Try to get instance information for instance name
	instance, err := service.DescribeSelectDBInstance(instanceId)
	if err == nil && instance != nil {
		if instance.Name != "" {
			d.Set("instance_name", instance.Name)
		}
	}

	return nil
}

func resourceAliCloudSelectDBClusterBindingDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse resource ID
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return WrapError(fmt.Errorf("invalid resource ID format: %s", d.Id()))
	}
	clusterId := parts[0]
	instanceId := parts[1]

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := service.DeleteSelectDBClusterBinding(clusterId, instanceId)
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
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteClusterBinding", AlibabaCloudSdkGoERROR)
	}

	// Wait for the cluster binding to be deleted
	err = service.WaitForSelectDBClusterBindingDeleted(clusterId, instanceId, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
