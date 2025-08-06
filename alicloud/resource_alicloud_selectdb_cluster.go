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
			"cluster_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the SelectDB cluster. This field is computed from the API response.",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The description of the SelectDB cluster.",
			},
			"zone_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The zone ID where the SelectDB cluster will be created.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the VPC where the SelectDB cluster will be created.",
			},
			"vswitch_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the VSwitch where the SelectDB cluster will be created.",
			},

			// ======== Cluster Configuration ========
			"cluster_class": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The cluster class (specification) of the SelectDB cluster.",
			},
			"cache_size": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The cache size of the SelectDB cluster in GB (e.g., 200 for 200GB).",
			},
			"engine": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "selectdb",
				Description: "The database engine type of the SelectDB cluster.",
			},
			"engine_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "4.0",
				Description: "The database engine version of the SelectDB cluster.",
			},
			"charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "PostPaid",
				ValidateFunc: validation.StringInSlice([]string{"PostPaid", "PrePaid"}, false),
				Description:  "The billing method of the SelectDB cluster.",
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

			// ======== Cluster Configuration Parameters ========
			"params": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Configuration parameters for the SelectDB cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Parameter name.",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Parameter value.",
						},
						"default_value": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Default value of the parameter.",
						},
						"comment": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Parameter comment or description.",
						},
						"is_dynamic": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Whether the parameter is dynamic (can be changed without restart).",
						},
						"is_user_modifiable": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Whether the parameter can be modified by users.",
						},
						"optional": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Optional values or range for the parameter.",
						},
						"param_category": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Category of the parameter.",
						},
					},
				},
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
	zoneId := d.Get("zone_id").(string)
	vpcId := d.Get("vpc_id").(string)
	vswitchId := d.Get("vswitch_id").(string)

	// Create cluster object with all required fields
	cacheSizeGB := d.Get("cache_size").(int)
	cluster := &selectdb.Cluster{
		InstanceId:    instanceId,
		ZoneId:        zoneId,
		VpcId:         vpcId,
		VSwitchId:     vswitchId,
		ClusterClass:  d.Get("cluster_class").(string),
		CacheSize:     int32(cacheSizeGB),
		Engine:        d.Get("engine").(string),
		EngineVersion: d.Get("engine_version").(string),
		ChargeType:    d.Get("charge_type").(string),
	}

	// Set cluster description - always set this field, even if empty
	cluster.ClusterDescription = d.Get("description").(string)

	var result *selectdb.Cluster
	// Use resource.Retry for creation to handle temporary failures
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		createdCluster, err := service.CreateSelectDBCluster(cluster)
		if err != nil {
			if NeedRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		result = createdCluster
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_selectdb_cluster", "CreateSelectCluster", AlibabaCloudSdkGoERROR)
	}

	d.SetId(service.EncodeSelectDBClusterId(instanceId, result.ClusterId))

	// Wait for the cluster to be created
	err = service.WaitForSelectDBClusterCreated(instanceId, result.ClusterId, d.Timeout(schema.TimeoutCreate))
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

	if cluster.ZoneId != "" {
		d.Set("zone_id", cluster.ZoneId)
	}

	if cluster.VpcId != "" {
		d.Set("vpc_id", cluster.VpcId)
	}

	if cluster.VSwitchId != "" {
		d.Set("vswitch_id", cluster.VSwitchId)
	}

	if cluster.Status != "" {
		d.Set("status", cluster.Status)
	}

	if cluster.CreatedTime != "" {
		d.Set("create_time", cluster.CreatedTime)
	}

	// Set cluster description directly from cluster object
	d.Set("description", cluster.ClusterDescription)

	// Get cluster configuration parameters
	params, err := service.DescribeSelectDBClusterConfig(clusterId, instanceId)
	if err == nil && len(params) > 0 {
		// Convert parameters to schema format
		paramsList := make([]map[string]interface{}, 0)
		for _, param := range params {
			paramMap := map[string]interface{}{
				"name":               param.Name,
				"value":              param.Value,
				"default_value":      param.DefaultValue,
				"comment":            param.Comment,
				"is_dynamic":         param.IsDynamic,
				"is_user_modifiable": param.IsUserModifiable,
				"optional":           param.Optional,
				"param_category":     param.ParamCategory,
			}
			paramsList = append(paramsList, paramMap)
		}
		d.Set("params", paramsList)
	} else {
		// Set empty list if no parameters or error occurred
		d.Set("params", []map[string]interface{}{})
	}

	// Set cluster configuration
	if cluster.ClusterClass != "" {
		d.Set("cluster_class", cluster.ClusterClass)
	}

	if cluster.CacheSize > 0 {
		// Set cache size directly from int32 field
		d.Set("cache_size", int(cluster.CacheSize))
	}

	// Set engine configuration from cluster or preserve from state
	if cluster.Engine != "" {
		d.Set("engine", cluster.Engine)
	}

	if cluster.EngineVersion != "" {
		d.Set("engine_version", cluster.EngineVersion)
	}

	if cluster.ChargeType != "" {
		d.Set("charge_type", cluster.ChargeType)
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

	// Check if any modifiable fields have changed
	if d.HasChange("cluster_class") || d.HasChange("cache_size") || d.HasChange("engine") {
		// Get current cluster information
		currentCluster, err := service.DescribeSelectDBCluster(instanceId, clusterId)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DescribeSelectDBCluster", AlibabaCloudSdkGoERROR)
		}

		// Create updated cluster object with modified fields
		updatedCluster := &selectdb.Cluster{
			InstanceId: instanceId,
			ClusterId:  clusterId,
		}

		// Update cluster class if changed
		if d.HasChange("cluster_class") {
			updatedCluster.ClusterClass = d.Get("cluster_class").(string)
		} else {
			updatedCluster.ClusterClass = currentCluster.ClusterClass
		}

		// Update cache size if changed
		if d.HasChange("cache_size") {
			cacheSizeGB := d.Get("cache_size").(int)
			updatedCluster.CacheSize = int32(cacheSizeGB)
		} else {
			updatedCluster.CacheSize = currentCluster.CacheSize
		}

		// Update engine if changed
		if d.HasChange("engine") {
			updatedCluster.Engine = d.Get("engine").(string)
		} else {
			updatedCluster.Engine = currentCluster.Engine
		}

		// Perform the modification
		_, err = service.ModifySelectDBCluster(updatedCluster)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifyCluster", AlibabaCloudSdkGoERROR)
		}

		// Wait for modification to complete
		err = service.WaitForSelectDBClusterUpdated(instanceId, clusterId, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}

		if d.HasChange("cluster_class") {
			d.SetPartial("cluster_class")
		}
		if d.HasChange("cache_size") {
			d.SetPartial("cache_size")
		}
		if d.HasChange("engine") {
			d.SetPartial("engine")
		}
	}

	// Update cluster description if changed
	if d.HasChange("description") {
		newDescription := d.Get("description").(string)
		if newDescription != "" {
			// Use ModifySelectDBBEClusterAttribute to update description
			// The attributeType should be "DBInstanceDescription" for updating cluster description
			err := service.ModifySelectDBBEClusterAttribute(clusterId, instanceId, "DBInstanceDescription", newDescription)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifySelectDBBEClusterAttribute", AlibabaCloudSdkGoERROR)
			}
		}
		d.SetPartial("description")
	}

	// Auto scaling rules update - placeholder for future implementation
	if d.HasChange("auto_scaling_rules") {
		// Note: Auto scaling rules management is not yet implemented in the API
		// This is a placeholder for when the API supports these operations
		log.Printf("[WARN] Auto scaling rules updates are not yet implemented")
		d.SetPartial("auto_scaling_rules")
	}

	// Update cluster configuration parameters if changed
	if d.HasChange("params") {
		oldParams, newParams := d.GetChange("params")
		oldParamsSet := oldParams.(*schema.Set)
		newParamsSet := newParams.(*schema.Set)

		// Get parameters that need to be updated (only those that changed)
		var paramsToUpdate []selectdb.ClusterConfigParam

		// Convert new parameters to map for easy lookup
		newParamsMap := make(map[string]string)
		for _, paramInterface := range newParamsSet.List() {
			param := paramInterface.(map[string]interface{})
			name := param["name"].(string)
			value := param["value"].(string)
			newParamsMap[name] = value
		}

		// Convert old parameters to map for comparison
		oldParamsMap := make(map[string]string)
		for _, paramInterface := range oldParamsSet.List() {
			param := paramInterface.(map[string]interface{})
			name := param["name"].(string)
			value := param["value"].(string)
			oldParamsMap[name] = value
		}

		// Find parameters that have changed or are new
		for name, newValue := range newParamsMap {
			if oldValue, exists := oldParamsMap[name]; !exists || oldValue != newValue {
				paramsToUpdate = append(paramsToUpdate, selectdb.ClusterConfigParam{
					Name:  name,
					Value: newValue,
				})
			}
		}

		// Update parameters if there are any changes
		if len(paramsToUpdate) > 0 {
			err := service.UpdateSelectDBClusterConfig(clusterId, instanceId, paramsToUpdate)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateSelectDBClusterConfig", AlibabaCloudSdkGoERROR)
			}

			// Wait for configuration update to complete
			err = service.WaitForSelectDBClusterUpdated(instanceId, clusterId, d.Timeout(schema.TimeoutUpdate))
			if err != nil {
				return WrapErrorf(err, IdMsg, d.Id())
			}
		}

		d.SetPartial("params")
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
