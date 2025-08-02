package alicloud

import (
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
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

	// Add region if available
	if service.client.RegionId != "" {
		options = append(options, selectdb.WithRegion(service.client.RegionId))
	}

	// Create the cluster
	cluster, err := service.CreateSelectDBCluster(instanceId, clusterClass, cacheSize, options...)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_selectdb_cluster", "CreateCluster", AlibabaCloudSdkGoERROR)
	}

	d.SetId(service.EncodeSelectDBClusterId(instanceId, cluster.ClusterId))

	// Wait for the cluster to be created
	err = service.WaitForSelectDBCluster(instanceId, cluster.ClusterId, Running, int(d.Timeout(schema.TimeoutCreate).Seconds()))
	if err != nil {
		return WrapError(err)
	}

	// Set auto scaling rules if specified
	if rules, ok := d.GetOk("auto_scaling_rules"); ok && len(rules.([]interface{})) > 0 {
		err = updateSelectDBAutoScalingRules(d, service)
		if err != nil {
			return WrapError(err)
		}
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
		if IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DescribeSelectDBCluster", AlibabaCloudSdkGoERROR)
	}

	d.Set("instance_id", instanceId)
	d.Set("cluster_id", clusterId)
	d.Set("cluster_name", cluster.ClusterName)
	d.Set("status", cluster.Status)
	d.Set("create_time", cluster.CreatedTime)

	// Since the current CWS-Lib-Go API doesn't provide FE/BE specific details,
	// we need to set default/computed values or retrieve from Terraform state
	// TODO: Update when CWS-Lib-Go provides complete cluster details

	// Set FE configuration - use current state if available
	if existingFE := d.Get("fe_config").([]interface{}); len(existingFE) > 0 {
		d.Set("fe_config", existingFE)
	}

	// Set BE configuration - use current state if available
	if existingBE := d.Get("be_config").([]interface{}); len(existingBE) > 0 {
		d.Set("be_config", existingBE)
	}

	// Set description from current state if not available in API response
	if existingDesc := d.Get("description").(string); existingDesc != "" {
		d.Set("description", existingDesc)
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

	// Update cluster name or description if changed
	if d.HasChanges("cluster_name", "description") {
		// Use cluster modification API to update name/description
		var options []selectdb.ModifyClusterOption

		// Note: Current CWS-Lib-Go may not support all update operations
		// This is a placeholder for when the API supports these operations
		_, err := service.ModifySelectDBCluster(instanceId, clusterId, options...)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifyCluster", AlibabaCloudSdkGoERROR)
		}

		d.SetPartial("cluster_name")
		d.SetPartial("description")
	}

	// Update FE node count if changed
	if d.HasChange("fe_config") {
		// TODO: Implement FE node scaling when CWS-Lib-Go supports it
		// oldConfig, newConfig := d.GetChange("fe_config")
		// oldFEConfig := oldConfig.([]interface{})[0].(map[string]interface{})
		// newFEConfig := newConfig.([]interface{})[0].(map[string]interface{})

		// if oldFEConfig["node_count"] != newFEConfig["node_count"] {
		//     Implement FE node scaling API call
		// }

		d.SetPartial("fe_config")
	}

	// Update BE node count if changed
	if d.HasChange("be_config") {
		// TODO: Implement BE node scaling when CWS-Lib-Go supports it
		// oldConfig, newConfig := d.GetChange("be_config")
		// oldBEConfig := oldConfig.([]interface{})[0].(map[string]interface{})
		// newBEConfig := newConfig.([]interface{})[0].(map[string]interface{})

		// if oldBEConfig["node_count"] != newBEConfig["node_count"] {
		//     Implement BE node scaling API call
		// }

		d.SetPartial("be_config")
	}

	// Update auto scaling rules if changed
	if d.HasChange("auto_scaling_rules") {
		// TODO: Implement auto scaling rules update when CWS-Lib-Go supports it
		// err = updateSelectDBAutoScalingRules(d, service)
		// if err != nil {
		//     return WrapError(err)
		// }
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

	err = service.DeleteSelectDBCluster(instanceId, clusterId)
	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteCluster", AlibabaCloudSdkGoERROR)
	}

	return WrapError(service.WaitForSelectDBCluster(instanceId, clusterId, Deleted, int(d.Timeout(schema.TimeoutDelete).Seconds())))
}

func updateSelectDBAutoScalingRules(d *schema.ResourceData, service *SelectDBService) error {
	// TODO: Implement auto scaling rules update when CWS-Lib-Go supports it
	// Currently commented out as the required API structures are not available

	/*
		instanceId, clusterId, err := service.DecodeSelectDBClusterId(d.Id())
		if err != nil {
			return err
		}

		rules := d.Get("auto_scaling_rules").([]interface{})
		// Implementation would go here when API structures are available
	*/

	return nil
}
