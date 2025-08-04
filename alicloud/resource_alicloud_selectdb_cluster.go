package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudSelectDBCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudSelectDBClusterCreate,
		Read:   resourceAliCloudSelectDBClusterRead,
		Update: resourceAliCloudSelectDBClusterUpdate,
		Delete: resourceAliCloudSelectDBClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			// ======== Basic Cluster Information ========
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the SelectDB instance.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the SelectDB cluster.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the SelectDB cluster.",
			},

			// ======== Cluster Configuration ========
			"cluster_class": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The cluster class (specification) of the SelectDB cluster.",
			},
			"cache_size": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "200GB",
				Description: "The cache size of the SelectDB cluster (e.g., '200GB').",
			},

			// ======== Cluster Scaling Configuration ========
			"auto_scaling_rules": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the auto scaling rule.",
						},
						"rule_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"BE_SCALING", "FE_SCALING"}, false),
							Description:  "The type of the auto scaling rule.",
						},
						"min_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
							Description:  "The minimum node count for auto scaling.",
						},
						"max_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
							Description:  "The maximum node count for auto scaling.",
						},
						"trigger_threshold": {
							Type:         schema.TypeFloat,
							Required:     true,
							ValidateFunc: validation.FloatBetween(0.1, 0.9),
							Description:  "The threshold to trigger scaling action.",
						},
						"metric_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"CPU", "MEMORY", "DISK"}, false),
							Description:  "The metric type for scaling rule.",
						},
					},
				},
				Description: "The auto scaling rules for the cluster.",
			},

			// ======== Computed Information ========
			"cluster_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the SelectDB cluster.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the SelectDB cluster.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the SelectDB cluster.",
			},
		},
	}
}

func resourceAliCloudSelectDBClusterCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("instance_id").(string)

	// Get cluster configuration
	clusterClass := d.Get("cluster_class").(string)
	cacheSize := d.Get("cache_size").(string)

	var options []selectdb.ClusterCreateOption

	// Add description if specified
	if description := d.Get("description").(string); description != "" {
		options = append(options, selectdb.WithClusterDescription(description))
	}
	if description := d.Get("description").(string); description != "" {
		options = append(options, selectdb.WithClusterDescription(description))
	}

	// Add engine settings
	options = append(options, selectdb.WithEngine("selectdb"))
	options = append(options, selectdb.WithEngineVersion("2.1"))

	// Add charge type (default to PostPaid)
	options = append(options, selectdb.WithChargeType("PostPaid"))

	// Add region if available
	if service.client.RegionId != "" {
		options = append(options, selectdb.WithRegion(service.client.RegionId))
	}

	var cluster *selectdb.Cluster
	// Use resource.Retry for creation to handle temporary failures
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		result, err := service.CreateSelectDBCluster(instanceId, clusterClass, cacheSize, options...)
		if err != nil {
			if NeedRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		cluster = result
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_selectdb_cluster", "CreateCluster", AlibabaCloudSdkGoERROR)
	}

	d.SetId(service.EncodeSelectDBClusterId(instanceId, cluster.ClusterId))

	// Wait for the cluster to be created
	err = service.WaitForSelectDBClusterCreated(instanceId, cluster.ClusterId, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudSelectDBClusterRead(d, meta)
}

func resourceAliCloudSelectDBClusterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId, clusterId, err := service.DecodeSelectDBClusterId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	cluster, err := service.DescribeSelectDBCluster(instanceId, clusterId)
	if err != nil {
		if !d.IsNewResource() && IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_selectdb_cluster DescribeSelectDBCluster Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DescribeSelectDBCluster", AlibabaCloudSdkGoERROR)
	}

	d.Set("instance_id", instanceId)
	d.Set("cluster_id", clusterId)

	// Set cluster basic information
	if cluster.ClusterName != "" {
		d.Set("name", cluster.ClusterName)
	}

	if cluster.Status != "" {
		d.Set("status", cluster.Status)
	}

	if cluster.CreatedTime != "" {
		d.Set("create_time", cluster.CreatedTime)
	}

	// Get cluster configuration to extract more detailed information
	config, err := service.DescribeSelectDBClusterConfig(clusterId, instanceId)
	if err == nil && config != nil {
		// Extract configuration information if available
		if len(config.Params) > 0 {
			// Parse cluster configuration parameters
			// This is a simplified mapping - in practice you might want to
			// parse specific parameters based on their names
			for _, param := range config.Params {
				if param.Name == "cluster_description" && param.Value != "" {
					d.Set("description", param.Value)
				}
			}
		}
	}

	// Set cluster configuration
	if cluster.ClusterClass != "" {
		d.Set("cluster_class", cluster.ClusterClass)
	}

	if cluster.CacheStorageSizeGB != "" {
		d.Set("cache_size", cluster.CacheStorageSizeGB)
	}

	// Set description from cluster information or preserve from state
	if cluster.ClusterName != "" && d.Get("description").(string) == "" {
		// If no description is set and we have cluster info, try to get it from elsewhere
		if existingDesc := d.Get("description").(string); existingDesc != "" {
			d.Set("description", existingDesc)
		}
	}

	return nil
}

func resourceAliCloudSelectDBClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId, clusterId, err := service.DecodeSelectDBClusterId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	d.Partial(true)

	// Update cluster class if changed
	if d.HasChange("cluster_class") {
		newClusterClass := d.Get("cluster_class").(string)

		// Modify cluster class using the API
		var options []selectdb.ModifyClusterOption
		options = append(options, selectdb.WithClusterClass(newClusterClass))

		_, err := service.ModifySelectDBCluster(instanceId, clusterId, options...)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifyCluster", AlibabaCloudSdkGoERROR)
		}

		// Wait for modification to complete
		err = service.WaitForSelectDBClusterUpdated(instanceId, clusterId, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}

		d.SetPartial("cluster_class")
	}

	// Update cache size if changed
	if d.HasChange("cache_size") {
		newCacheSize := d.Get("cache_size").(string)

		// Modify cache size using the API
		var options []selectdb.ModifyClusterOption
		options = append(options, selectdb.WithCacheSize(newCacheSize))

		_, err := service.ModifySelectDBCluster(instanceId, clusterId, options...)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifyCluster", AlibabaCloudSdkGoERROR)
		}

		// Wait for modification to complete
		err = service.WaitForSelectDBClusterUpdated(instanceId, clusterId, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}

		d.SetPartial("cache_size")
	}

	// Update cluster description if changed
	if d.HasChange("description") {
		newDescription := d.Get("description").(string)
		if newDescription != "" {
			// Use cluster configuration modification to update description
			// This is a workaround since there's no direct API for description update
			parameters := fmt.Sprintf(`{"cluster_description": "%s"}`, newDescription)
			_, err := service.selectdbAPI.ModifyClusterConfig(clusterId, instanceId, parameters)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifyClusterConfig", AlibabaCloudSdkGoERROR)
			}
		}
		d.SetPartial("description")
	}

	// Update cluster name - Note: SelectDB may not support cluster name changes
	// This is kept for future API support
	if d.HasChange("cluster_name") {
		// Currently, cluster name changes are typically not supported
		// Log a warning and continue
		log.Printf("[WARN] Cluster name changes are not supported for SelectDB clusters")
		d.SetPartial("cluster_name")
	}

	// Auto scaling rules update - placeholder for future implementation
	if d.HasChange("auto_scaling_rules") {
		// Note: Auto scaling rules management is not yet implemented in the API
		// This is a placeholder for when the API supports these operations
		log.Printf("[WARN] Auto scaling rules updates are not yet implemented")
		d.SetPartial("auto_scaling_rules")
	}

	d.Partial(false)
	return resourceAliCloudSelectDBClusterRead(d, meta)
}

func resourceAliCloudSelectDBClusterDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId, clusterId, err := service.DecodeSelectDBClusterId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := service.DeleteSelectDBCluster(instanceId, clusterId)
		if err != nil {
			if IsNotFoundError(err) {
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
		if IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteCluster", AlibabaCloudSdkGoERROR)
	}

	// Wait for the cluster to be deleted
	err = service.WaitForSelectDBClusterDeleted(instanceId, clusterId, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}

// Helper functions for data conversion

// convertInterfaceToStringSlice converts []interface{} to []string
func convertInterfaceToStringSlice(v interface{}) []string {
	if v == nil {
		return []string{}
	}
	vList := v.([]interface{})
	result := make([]string, len(vList))
	for i, val := range vList {
		if val != nil {
			result[i] = val.(string)
		}
	}
	return result
}

// convertStringSliceToInterface converts []string to []interface{}
func convertStringSliceToInterface(slice []string) []interface{} {
	result := make([]interface{}, len(slice))
	for i, val := range slice {
		result[i] = val
	}
	return result
}

// validateSelectDBClusterConfig validates cluster configuration parameters
func validateSelectDBClusterConfig(feConfig map[string]interface{}) error {
	// Validate FE config
	if nodeCount, ok := feConfig["node_count"].(int); ok && nodeCount < 1 {
		return fmt.Errorf("FE node count must be at least 1")
	}

	if nodeType, ok := feConfig["node_type"].(string); ok && nodeType == "" {
		return fmt.Errorf("FE node type cannot be empty")
	}

	return nil
}
