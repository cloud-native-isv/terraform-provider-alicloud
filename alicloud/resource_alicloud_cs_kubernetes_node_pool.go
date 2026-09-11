package alicloud

import (
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunAckAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/ack"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

const resourceAliCloudCSKubernetesNodePoolName = "alicloud_cs_kubernetes_node_pool"

// ackNodePoolDeleteForce matches the fork convention: node pools are deleted
// without the force flag; nodes in an abnormal pool state require manual
// intervention instead of a forced removal.
const ackNodePoolDeleteForce = false

func resourceAliCloudAckNodepool() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudAckNodepoolCreate,
		Read:   resourceAliCloudAckNodepoolRead,
		Update: resourceAliCloudAckNodepoolUpdate,
		Delete: resourceAliCloudAckNodepoolDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the cluster that owns the node pool.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the node pool.",
			},
			"instance_types": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    1,
				Description: "The instance types of the nodes. Only the first entry is effective: the cws-lib-go ACK API layer models a single instance type per node pool.",
			},
			"desired_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The desired number of nodes in the node pool.",
			},
			"min_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The minimum number of nodes when auto scaling is enabled.",
			},
			"max_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The maximum number of nodes when auto scaling is enabled.",
			},
			"auto_scaling": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether auto scaling is enabled for the node pool.",
			},
			"instance_charge_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "PostPaid",
				Description: "The charge type of the nodes: `PostPaid` or `PrePaid`, default to `PostPaid`.",
				ValidateFunc: validation.StringInSlice([]string{
					"PostPaid", "PrePaid",
				}, false),
			},
			"system_disk_category": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The system disk category of the nodes, e.g. `cloud_essd`.",
			},
			"system_disk_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "The system disk size (GiB) of the nodes.",
			},
			"system_disk_performance_level": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The performance level of the ESSD system disk, e.g. `PL1`.",
			},
			"runtime_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The container runtime name, e.g. `containerd`.",
			},
			"runtime_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The container runtime version.",
			},
			"image_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The node image type, e.g. `AliyunLinux3`.",
			},
			"vswitch_ids": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    1,
				Description: "The vSwitches of the node pool.",
			},
			"security_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The security group for the nodes. When omitted the cluster security group applies.",
			},
			"labels": {
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
							Optional: true,
						},
					},
				},
				Description: "The Kubernetes labels applied to the nodes.",
			},
			"taints": {
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
							Optional: true,
						},
						"effect": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"NoSchedule", "PreferNoSchedule", "NoExecute",
							}, false),
						},
					},
				},
				Description: "The Kubernetes taints applied to the nodes.",
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The tags of the node pool.",
			},
			"management": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_repair": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"auto_upgrade": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"auto_vulfix": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
				Description: "The managed node pool policies. A policy is enabled by setting its flag to true.",
			},

			// Computed attributes
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the node pool, e.g. `active`.",
			},
			"node_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The current number of nodes in the node pool.",
			},
		},
	}
}

// buildAckNodePoolConfig maps the resource schema onto the cws-lib-go
// AckNodePoolConfig input shared by create and modify. Exposed for offline
// table-driven tests.
func buildAckNodePoolConfig(d *schema.ResourceData) *aliyunAckAPI.AckNodePoolConfig {
	config := &aliyunAckAPI.AckNodePoolConfig{
		NodePoolName:       d.Get("name").(string),
		InstanceChargeType: d.Get("instance_charge_type").(string),
		DesiredSize:        int64(d.Get("desired_size").(int)),
		MinSize:            int64(d.Get("min_size").(int)),
		MaxSize:            int64(d.Get("max_size").(int)),
		AutoScaling:        d.Get("auto_scaling").(bool),
		ImageType:          d.Get("image_type").(string),
		SecurityGroupId:    d.Get("security_group_id").(string),
		Tags:               expandAckTags(d.Get("tags")),
	}

	// The cws-lib-go API layer models a single instance type per node pool;
	// the list-shaped schema keeps the upstream HCL contract.
	if instanceTypes := d.Get("instance_types").([]interface{}); len(instanceTypes) > 0 {
		config.InstanceType = instanceTypes[0].(string)
	}

	for _, v := range d.Get("vswitch_ids").([]interface{}) {
		config.VSwitchIds = append(config.VSwitchIds, v.(string))
	}

	systemDisk := &aliyunAckAPI.AckNodeDisk{
		Category:         d.Get("system_disk_category").(string),
		Size:             int64(d.Get("system_disk_size").(int)),
		PerformanceLevel: d.Get("system_disk_performance_level").(string),
	}
	if systemDisk.Category != "" || systemDisk.Size > 0 || systemDisk.PerformanceLevel != "" {
		config.SystemDisk = systemDisk
	}

	if name, ok := d.GetOk("runtime_name"); ok {
		config.Runtime = &aliyunAckAPI.AckNodeRuntime{
			Name:    name.(string),
			Version: d.Get("runtime_version").(string),
		}
	}

	if raw, ok := d.GetOk("labels"); ok {
		for _, v := range raw.([]interface{}) {
			label := v.(map[string]interface{})
			config.Labels = append(config.Labels, aliyunAckAPI.AckNodeLabel{
				Key:   label["key"].(string),
				Value: label["value"].(string),
			})
		}
	}
	if raw, ok := d.GetOk("taints"); ok {
		for _, v := range raw.([]interface{}) {
			taint := v.(map[string]interface{})
			config.Taints = append(config.Taints, aliyunAckAPI.AckNodeTaint{
				Key:    taint["key"].(string),
				Value:  taint["value"].(string),
				Effect: taint["effect"].(string),
			})
		}
	}

	if raw, ok := d.GetOk("management"); ok {
		blocks := raw.([]interface{})
		if len(blocks) > 0 {
			management := blocks[0].(map[string]interface{})
			config.Management = buildAckNodePoolManagement(management)
		}
	}
	return config
}

// buildAckNodePoolManagement maps the management schema block onto the API
// layer policies: a boolean flag true enables the corresponding policy.
func buildAckNodePoolManagement(management map[string]interface{}) *aliyunAckAPI.AckNodePoolManagement {
	result := &aliyunAckAPI.AckNodePoolManagement{}
	hasPolicy := false
	if repair, ok := management["auto_repair"].(bool); ok && repair {
		hasPolicy = true
		result.AutoRepairPolicy = &aliyunAckAPI.AckNodeAutoRepairPolicy{RestartNodes: true}
	}
	if upgrade, ok := management["auto_upgrade"].(bool); ok && upgrade {
		hasPolicy = true
		result.AutoUpgradePolicy = &aliyunAckAPI.AckNodeAutoUpgradePolicy{KubeletUpgrade: true}
	}
	if vulfix, ok := management["auto_vulfix"].(bool); ok && vulfix {
		hasPolicy = true
		result.AutoVulFixPolicy = &aliyunAckAPI.AckNodeAutoVulFixPolicy{RestartNodes: true, VulLevels: "asap"}
	}
	if !hasPolicy {
		return nil
	}
	return result
}

// setAckNodePoolAttributes copies the node pool detail into the resource
// state. Exposed for offline table-driven tests.
func setAckNodePoolAttributes(d *schema.ResourceData, pool *aliyunAckAPI.AckNodePool) error {
	if pool == nil {
		return nil
	}
	d.Set("cluster_id", pool.ClusterId)
	d.Set("name", pool.NodePoolName)
	d.Set("instance_types", []interface{}{pool.InstanceType})
	d.Set("desired_size", int(pool.DesiredSize))
	d.Set("min_size", int(pool.MinSize))
	d.Set("max_size", int(pool.MaxSize))
	d.Set("auto_scaling", pool.AutoScaling)
	d.Set("instance_charge_type", pool.InstanceChargeType)
	d.Set("image_type", pool.ImageType)
	d.Set("security_group_id", pool.SecurityGroupId)
	d.Set("tags", flattenAckTags(pool.Tags))
	d.Set("status", pool.Status)
	d.Set("node_count", int(pool.NodeCount))

	if pool.SystemDisk != nil {
		d.Set("system_disk_category", pool.SystemDisk.Category)
		d.Set("system_disk_size", int(pool.SystemDisk.Size))
		d.Set("system_disk_performance_level", pool.SystemDisk.PerformanceLevel)
	}
	if pool.Runtime != nil {
		d.Set("runtime_name", pool.Runtime.Name)
		d.Set("runtime_version", pool.Runtime.Version)
	}

	vswitchIds := make([]interface{}, 0, len(pool.VSwitchIds))
	for _, v := range pool.VSwitchIds {
		vswitchIds = append(vswitchIds, v)
	}
	d.Set("vswitch_ids", vswitchIds)

	labels := make([]interface{}, 0, len(pool.Labels))
	for _, label := range pool.Labels {
		labels = append(labels, map[string]interface{}{
			"key":   label.Key,
			"value": label.Value,
		})
	}
	d.Set("labels", labels)

	taints := make([]interface{}, 0, len(pool.Taints))
	for _, taint := range pool.Taints {
		taints = append(taints, map[string]interface{}{
			"key":    taint.Key,
			"value":  taint.Value,
			"effect": taint.Effect,
		})
	}
	d.Set("taints", taints)

	management := make([]interface{}, 0, 1)
	if pool.Management != nil {
		block := map[string]interface{}{
			"auto_repair": pool.Management.AutoRepairPolicy != nil,
			"auto_upgrade": func() bool {
				// The detail flavor only surfaces the kubelet upgrade switch.
				return pool.Management.AutoUpgradePolicy != nil && pool.Management.AutoUpgradePolicy.KubeletUpgrade
			}(),
			"auto_vulfix": pool.Management.AutoVulFixPolicy != nil,
		}
		management = append(management, block)
	}
	d.Set("management", management)
	return nil
}

func resourceAliCloudAckNodepoolCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ackService, err := NewAckService(client)
	if err != nil {
		return WrapError(err)
	}

	clusterId := d.Get("cluster_id").(string)
	pool, err := ackService.GetAPI().CreateClusterNodePool(clusterId, buildAckNodePoolConfig(d))
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, resourceAliCloudCSKubernetesNodePoolName, "CreateClusterNodePool", err)
	}
	d.SetId(fmt.Sprintf("%s:%s", clusterId, pool.NodePoolId))

	// Wait for the node pool to become active: 20m timeout, 5s poll.
	stateConf := buildAckStateConf(
		nil, []string{ackNodePoolStateActive},
		d.Timeout(schema.TimeoutCreate), 5*time.Second, 5*time.Second,
		ackService.AckNodePoolStateRefreshFunc(d.Id(), ackNodePoolFailStates))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudAckNodepoolRead(d, meta)
}

func resourceAliCloudAckNodepoolRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ackService, err := NewAckService(client)
	if err != nil {
		return WrapError(err)
	}

	clusterId, nodepoolId, err := ackParseTwoPartId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	pool, err := ackService.GetAPI().DescribeClusterNodePoolDetail(clusterId, nodepoolId)
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, resourceAliCloudCSKubernetesNodePoolName, "DescribeClusterNodePoolDetail", err)
	}

	return setAckNodePoolAttributes(d, pool)
}

func resourceAliCloudAckNodepoolUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ackService, err := NewAckService(client)
	if err != nil {
		return WrapError(err)
	}

	clusterId, nodepoolId, err := ackParseTwoPartId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	if _, err := ackService.GetAPI().ModifyClusterNodePool(clusterId, nodepoolId, buildAckNodePoolConfig(d)); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, resourceAliCloudCSKubernetesNodePoolName, "ModifyClusterNodePool", err)
	}

	stateConf := buildAckStateConf(
		nil, []string{ackNodePoolStateActive},
		d.Timeout(schema.TimeoutUpdate), 5*time.Second, 5*time.Second,
		ackService.AckNodePoolStateRefreshFunc(d.Id(), ackNodePoolFailStates))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudAckNodepoolRead(d, meta)
}

func resourceAliCloudAckNodepoolDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ackService, err := NewAckService(client)
	if err != nil {
		return WrapError(err)
	}

	clusterId, nodepoolId, err := ackParseTwoPartId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	if _, err := ackService.GetAPI().DeleteClusterNodepool(clusterId, nodepoolId, ackNodePoolDeleteForce); err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, resourceAliCloudCSKubernetesNodePoolName, "DeleteClusterNodepool", err)
	}

	stateConf := buildAckStateConf(
		nil, []string{ackClusterStateDeleted},
		d.Timeout(schema.TimeoutDelete), 5*time.Second, 5*time.Second,
		ackService.AckNodePoolDeleteStateRefreshFunc(d.Id()))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}
	return nil
}
