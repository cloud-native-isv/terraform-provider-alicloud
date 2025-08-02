package alicloud

import (
	"fmt"
	"log"
	"strconv"
	"strings"
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
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the SelectDB instance.",
			},
			"cluster_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the SelectDB cluster.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the SelectDB cluster.",
			},
			"fe_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"node_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
							Description:  "The number of FE nodes.",
						},
						"node_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The type of FE nodes.",
						},
						"resource_group": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The resource group of FE nodes.",
						},
					},
				},
				Description: "The configuration of FE nodes.",
			},
			"be_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"node_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
							Description:  "The number of BE nodes.",
						},
						"node_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The type of BE nodes.",
						},
						"resource_group": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The resource group of BE nodes.",
						},
						"disk_size": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      200,
							ValidateFunc: validation.IntBetween(100, 2000),
							Description:  "The disk size of BE nodes in GB.",
						},
						"disk_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntBetween(1, 10),
							Description:  "The number of disks per BE node.",
						},
					},
				},
				Description: "The configuration of BE nodes.",
			},
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

	// Prepare FE config
	feConfigList := d.Get("fe_config").([]interface{})
	feConfig := feConfigList[0].(map[string]interface{})

	// Prepare BE config
	beConfigList := d.Get("be_config").([]interface{})
	beConfig := beConfigList[0].(map[string]interface{})

	// Prepare cluster creation options
	clusterClass := feConfig["node_type"].(string)
	cacheSize := fmt.Sprintf("%dGB", beConfig["disk_size"].(int))

	var options []selectdb.ClusterCreateOption

	// Add description if specified
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
	err = service.WaitForSelectDBCluster(instanceId, cluster.ClusterId, Running, int(d.Timeout(schema.TimeoutCreate).Seconds()))
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
		d.Set("cluster_name", cluster.ClusterName)
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

	// Set FE configuration - preserve existing configuration from state
	// since the current API doesn't provide detailed FE/BE node information
	if existingFE := d.Get("fe_config").([]interface{}); len(existingFE) > 0 {
		feConfig := existingFE[0].(map[string]interface{})

		// Update with any available information from cluster
		if cluster.ClusterClass != "" {
			feConfig["node_type"] = cluster.ClusterClass
		}

		d.Set("fe_config", []interface{}{feConfig})
	}

	// Set BE configuration - preserve existing configuration from state
	if existingBE := d.Get("be_config").([]interface{}); len(existingBE) > 0 {
		beConfig := existingBE[0].(map[string]interface{})

		// Update with any available information from cluster
		if cluster.CacheStorageSizeGB != "" {
			// Parse cache size (remove "GB" suffix if present)
			cacheSizeStr := strings.TrimSuffix(cluster.CacheStorageSizeGB, "GB")
			if cacheSize, parseErr := strconv.Atoi(cacheSizeStr); parseErr == nil {
				beConfig["disk_size"] = cacheSize
			}
		}

		d.Set("be_config", []interface{}{beConfig})
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

	// Update cluster class if FE config changed
	if d.HasChange("fe_config") {
		oldConfig, newConfig := d.GetChange("fe_config")
		oldFEConfig := oldConfig.([]interface{})[0].(map[string]interface{})
		newFEConfig := newConfig.([]interface{})[0].(map[string]interface{})

		if oldFEConfig["node_type"] != newFEConfig["node_type"] {
			// Modify cluster class using the API
			var options []selectdb.ModifyClusterOption
			options = append(options, selectdb.WithClusterClass(newFEConfig["node_type"].(string)))

			_, err := service.ModifySelectDBCluster(instanceId, clusterId, options...)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifyCluster", AlibabaCloudSdkGoERROR)
			}

			// Wait for modification to complete
			err = service.WaitForSelectDBCluster(instanceId, clusterId, Running, int(d.Timeout(schema.TimeoutUpdate).Seconds()))
			if err != nil {
				return WrapErrorf(err, IdMsg, d.Id())
			}
		}

		d.SetPartial("fe_config")
	}

	// Update cache size if BE config changed
	if d.HasChange("be_config") {
		oldConfig, newConfig := d.GetChange("be_config")
		oldBEConfig := oldConfig.([]interface{})[0].(map[string]interface{})
		newBEConfig := newConfig.([]interface{})[0].(map[string]interface{})

		if oldBEConfig["disk_size"] != newBEConfig["disk_size"] {
			// Modify cache size using the API
			var options []selectdb.ModifyClusterOption
			cacheSize := fmt.Sprintf("%dGB", newBEConfig["disk_size"].(int))
			options = append(options, selectdb.WithCacheSize(cacheSize))

			_, err := service.ModifySelectDBCluster(instanceId, clusterId, options...)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifyCluster", AlibabaCloudSdkGoERROR)
			}

			// Wait for modification to complete
			err = service.WaitForSelectDBCluster(instanceId, clusterId, Running, int(d.Timeout(schema.TimeoutUpdate).Seconds()))
			if err != nil {
				return WrapErrorf(err, IdMsg, d.Id())
			}
		}

		d.SetPartial("be_config")
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
		err := service.DeleteSelectDBCluster(instanceId, clusterId, service.client.RegionId)
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
	stateConf := &resource.StateChangeConf{
		Pending: []string{selectdb.ClusterStatusDeleting, selectdb.ClusterStatusRunning, selectdb.ClusterStatusStopped},
		Target:  []string{""},
		Refresh: func() (interface{}, string, error) {
			cluster, err := service.DescribeSelectDBCluster(instanceId, clusterId)
			if err != nil {
				if IsNotFoundError(err) {
					return nil, "", nil
				}
				return nil, "", WrapError(err)
			}
			return cluster, cluster.Status, nil
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
func validateSelectDBClusterConfig(feConfig, beConfig map[string]interface{}) error {
	// Validate FE config
	if nodeCount, ok := feConfig["node_count"].(int); ok && nodeCount < 1 {
		return fmt.Errorf("FE node count must be at least 1")
	}

	if nodeType, ok := feConfig["node_type"].(string); ok && nodeType == "" {
		return fmt.Errorf("FE node type cannot be empty")
	}

	// Validate BE config
	if nodeCount, ok := beConfig["node_count"].(int); ok && nodeCount < 1 {
		return fmt.Errorf("BE node count must be at least 1")
	}

	if nodeType, ok := beConfig["node_type"].(string); ok && nodeType == "" {
		return fmt.Errorf("BE node type cannot be empty")
	}

	if diskSize, ok := beConfig["disk_size"].(int); ok && (diskSize < 100 || diskSize > 2000) {
		return fmt.Errorf("BE disk size must be between 100 and 2000 GB")
	}

	if diskCount, ok := beConfig["disk_count"].(int); ok && (diskCount < 1 || diskCount > 10) {
		return fmt.Errorf("BE disk count must be between 1 and 10")
	}

	return nil
}
