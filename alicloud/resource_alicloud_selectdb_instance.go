package alicloud

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudSelectDBInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudSelectDBInstanceCreate,
		Read:   resourceAliCloudSelectDBInstanceRead,
		Update: resourceAliCloudSelectDBInstanceUpdate,
		Delete: resourceAliCloudSelectDBInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			// ======== Basic Instance Information ========
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The description of the SelectDB instance.",
			},

			// ======== Resource Organization ========
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The resource group ID.",
			},
			"tags": tagsSchema(),

			// ======== Database Engine Configuration ========
			"engine": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "SelectDB",
				Description: "The engine type of the SelectDB instance. Default is SelectDB.",
			},
			"engine_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The engine version of the SelectDB instance. Can be updated to upgrade the engine version.",
			},

			// ======== Network Configuration ========
			"zone_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The zone ID where the SelectDB instance is located.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The VPC ID where the SelectDB instance is located.",
			},
			"vswitch_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The VSwitch ID where the SelectDB instance is located.",
			},
			"multi_zone": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Description: "Multi-zone configuration for high availability.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The zone ID.",
						},
						"vswitch_ids": {
							Type:        schema.TypeList,
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "The VSwitch IDs in this zone.",
						},
						"cidr": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The CIDR block.",
						},
						"available_ip_count": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The available IP count.",
						},
					},
				},
			},

			// ======== Compute and Storage Resources ========
			"instance_class": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The instance class of the SelectDB instance.",
			},
			"cache_size": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The cache size of the SelectDB instance in GB.",
			},
			"deploy_scheme": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The deployment scheme of the SelectDB instance.",
			},

			// ======== Billing Configuration ========
			"payment_type": {
				Type:         schema.TypeString,
				ValidateFunc: StringInSlice([]string{"PayAsYouGo", "Subscription"}, false),
				Required:     true,
				ForceNew:     true,
				Description:  "The payment type of the SelectDB instance. Valid values: PayAsYouGo, Subscription.",
			},
			"period": {
				Type:             schema.TypeString,
				ValidateFunc:     StringInSlice([]string{string(Year), string(Month)}, false),
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: selectdbPostPaidDiffSuppressFunc,
				Description:      "The billing period for Subscription instances. Valid values: Month, Year.",
			},
			"period_time": {
				Type:             schema.TypeInt,
				ValidateFunc:     IntInSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 12, 24, 36}),
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: selectdbPostPaidDiffSuppressFunc,
				Description:      "The period time for Subscription instances.",
			},

			// ======== Maintenance Configuration ========
			"maintain_start_time": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The maintenance start time of the SelectDB instance in HH:MM format.",
			},
			"maintain_end_time": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The maintenance end time of the SelectDB instance in HH:MM format.",
			},

			// ======== Security Configuration ========
			"security_ip_groups": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Security IP groups for controlling access to the instance.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "default",
							Description: "The name of the security IP group. Default value: default.",
						},
						"security_ip_list": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of IP addresses or CIDR blocks allowed to access the SelectDB instance.",
						},
						"modify_mode": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "Cover",
							Description: "The modification mode. Valid values: Cover, Append, Delete. Default value: Cover.",
						},
					},
				},
			},

			// ======== Computed Information - Basic Instance ========
			"region_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The region ID where the instance is located.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the SelectDB instance.",
			},
			"charge_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The charge type of the instance.",
			},
			"category": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The category of the instance.",
			},
			"instance_used_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The instance used type.",
			},

			// ======== Computed Information - Connection ========
			"connection_string": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The connection string of the instance.",
			},
			"sub_domain": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The sub domain of the instance.",
			},

			// ======== Computed Information - Resource Configuration ========
			"resource_cpu": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The CPU cores of the instance.",
			},
			"resource_memory": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The memory size of the instance in GB.",
			},
			"storage_size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The storage size of the instance in GB.",
			},
			"storage_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The storage type of the instance.",
			},
			"object_store_size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The object store size of the instance in GB.",
			},
			"cluster_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The cluster count of the instance.",
			},
			"scale_min": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The minimum scale of the instance.",
			},
			"scale_max": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The maximum scale of the instance.",
			},
			"scale_replica": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The scale replica of the instance.",
			},

			// ======== Computed Information - Lock Status ========
			"lock_mode": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The lock mode of the instance.",
			},
			"lock_reason": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The lock reason of the instance.",
			},

			// ======== Computed Information - Time ========
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the instance.",
			},
			"gmt_created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The GMT creation time of the instance.",
			},
			"gmt_modified": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The GMT modification time of the instance.",
			},
			"expire_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The expiration time of the instance.",
			},

			// ======== Computed Information - Upgrade ========
			"can_upgrade_versions": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of engine versions that the instance can upgrade to.",
			},

			// ======== Computed Information - Network Details ========
			"instance_net_infos": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The network information of the instance.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The cluster ID.",
						},
						"connection_string": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The connection string.",
						},
						"ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The IP address.",
						},
						"net_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The network type.",
						},
						"user_visible": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the network is user visible.",
						},
						"vpc_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The VPC ID.",
						},
						"vpc_instance_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The VPC instance ID.",
						},
						"vswitch_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The VSwitch ID.",
						},
						"port_list": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The port list.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"port": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The port number.",
									},
									"protocol": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The protocol.",
									},
								},
							},
						},
					},
				},
			},

			// ======== Computed Information - Security ========
			"security_ip_lists": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The security IP lists of the instance.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The group name.",
						},
						"group_tag": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The group tag.",
						},
						"security_ip_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The security IP type.",
						},
						"security_ip_list": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "The security IP list.",
						},
						"whitelist_net_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The whitelist network type.",
						},
					},
				},
			},

			// ======== Computed Information - Database Clusters ========
			"db_cluster_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The database cluster list of the instance.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"db_cluster_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The cluster ID.",
						},
						"db_cluster_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The cluster name.",
						},
						"db_cluster_class": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The cluster class.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The cluster status.",
						},
						"charge_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The cluster charge type.",
						},
						"cpu_cores": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The CPU cores.",
						},
						"memory": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The memory size in GB.",
						},
						"cache_storage_size_gb": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The cache storage size in GB.",
						},
						"cache_storage_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The cache storage type.",
						},
						"performance_level": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The performance level.",
						},
						"scaling_rules_enable": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether scaling rules are enabled.",
						},
						"created_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The creation time.",
						},
						"modified_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The modification time.",
						},
					},
				},
			},
		},
	}
}

func resourceAliCloudSelectDBInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	selectDBService, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	// Build instance directly using Instance struct fields
	instance := &selectdb.Instance{
		Engine:      "SelectDB",
		Description: d.Get("description").(string),
		RegionId:    client.RegionId,
	}

	// Set basic configuration
	if v, ok := d.GetOk("engine_version"); ok {
		instance.EngineVersion = v.(string)
	}

	if v, ok := d.GetOk("zone_id"); ok {
		instance.ZoneId = v.(string)
	}

	if v, ok := d.GetOk("vpc_id"); ok {
		instance.VpcId = v.(string)
	}

	if v, ok := d.GetOk("vswitch_id"); ok {
		instance.VswitchId = v.(string)
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		instance.ResourceGroupId = v.(string)
	}

	if v, ok := d.GetOk("deploy_scheme"); ok {
		instance.DeployScheme = v.(string)
	}

	if v, ok := d.GetOk("instance_class"); ok {
		instance.InstanceClass = v.(string)
	}

	if v, ok := d.GetOk("cache_size"); ok {
		instance.CacheSize = int32(v.(int))
	}

	// Set payment configuration
	paymentType := d.Get("payment_type").(string)
	switch paymentType {
	case "PayAsYouGo":
		instance.ChargeType = "POSTPAY"
	case "Subscription":
		instance.ChargeType = "PREPAY"

		// Set period for subscription instances
		if v, ok := d.GetOk("period"); ok {
			instance.Period = v.(string)
		}
		if v, ok := d.GetOk("period_time"); ok {
			instance.UsedTime = int32(v.(int))
		}
	}

	// Set multi-zone configuration
	if v, ok := d.GetOk("multi_zone"); ok {
		multiZoneList := v.([]interface{})
		for _, item := range multiZoneList {
			multiZone := item.(map[string]interface{})
			mz := selectdb.MultiZone{
				ZoneId: multiZone["zone_id"].(string),
			}

			if vswitchIds, ok := multiZone["vswitch_ids"]; ok {
				vswitchIdList := vswitchIds.([]interface{})
				for _, vswitchId := range vswitchIdList {
					mz.VSwitchIds = append(mz.VSwitchIds, vswitchId.(string))
				}
			}

			if cidr, ok := multiZone["cidr"]; ok && cidr.(string) != "" {
				mz.Cidr = cidr.(string)
			}

			if availableIpCount, ok := multiZone["available_ip_count"]; ok {
				mz.AvailableIpCount = int64(availableIpCount.(int))
			}

			instance.MultiZone = append(instance.MultiZone, mz)
		}
	}

	// Set tags
	if v, ok := d.GetOk("tags"); ok {
		tags := v.(map[string]interface{})
		for key, value := range tags {
			instance.Tags = append(instance.Tags, selectdb.Tag{
				Key:   key,
				Value: value.(string),
			})
		}
	}

	// Create instance
	result, err := selectDBService.CreateSelectDBInstance(instance)
	if err != nil {
		return WrapError(err)
	}

	if result == nil || result.Id == "" {
		return WrapErrorf(fmt.Errorf("create result is nil or instance ID is empty"), IdMsg, "alicloud_selectdb_instance")
	}

	d.SetId(result.Id)

	// Wait for instance to be ready
	err = selectDBService.WaitForSelectDBInstanceCreated(d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudSelectDBInstanceRead(d, meta)
}

func resourceAliCloudSelectDBInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	selectDBService, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	d.Partial(true)

	// Handle description update
	if d.HasChange("description") {
		description := d.Get("description").(string)
		regionId := d.Get("region_id").(string)

		err := selectDBService.ModifySelectDBInstance(d.Id(), "DBInstanceDescription", description, regionId)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifySelectDBInstanceDescription", AlibabaCloudSdkGoERROR)
		}
		d.SetPartial("description")
	}

	// Handle engine version upgrade
	if d.HasChange("engine_version") {
		_, newVersionInterface := d.GetChange("engine_version")
		newVersion := newVersionInterface.(string)

		if newVersion != "" {
			// Get current instance to check available upgrade versions
			currentInstance, err := selectDBService.DescribeSelectDBInstance(d.Id())
			if err != nil {
				return WrapError(err)
			}

			// Validate if the version can be upgraded
			canUpgrade := false
			upgradeTargetVersion := newVersion

			// Check if newVersion is a major version (like "3.0" or "4.0")
			isNewVersionWithMajorVersion := (newVersion == "3.0" || newVersion == "4.0")

			for _, availableVersion := range currentInstance.CanUpgradeVersions {
				if isNewVersionWithMajorVersion {
					if len(availableVersion) >= len(newVersion) && availableVersion[:len(newVersion)] == newVersion {
						upgradeTargetVersion = availableVersion
						canUpgrade = true
						break
					}
				} else {
					if availableVersion == newVersion {
						upgradeTargetVersion = availableVersion
						canUpgrade = true
						break
					}
				}
			}

			if canUpgrade {
				regionId := d.Get("region_id").(string)

				err = selectDBService.UpgradeSelectDBInstanceEngineVersion(d.Id(), upgradeTargetVersion, regionId)
				if err != nil {
					return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpgradeSelectDBInstanceEngineVersion", AlibabaCloudSdkGoERROR)
				}

				// Wait for upgrade to complete
				err = selectDBService.WaitForSelectDBInstanceUpdated(d.Id(), d.Timeout(schema.TimeoutUpdate))
				if err != nil {
					return WrapErrorf(err, IdMsg, d.Id())
				}
			} else {
				return WrapErrorf(fmt.Errorf("invalid upgrade version for %s, cannot upgrade to %s", d.Id(), newVersion), DefaultErrorMsg, d.Id(), "UpgradeSelectDBInstanceEngineVersion", AlibabaCloudSdkGoERROR)
			}
		}

		d.SetPartial("engine_version")
	}

	// Handle maintenance time update
	if d.HasChange("maintain_start_time") || d.HasChange("maintain_end_time") {
		maintainTime := fmt.Sprintf("%s-%s", d.Get("maintain_start_time").(string), d.Get("maintain_end_time").(string))
		regionId := d.Get("region_id").(string)

		err := selectDBService.ModifySelectDBInstance(d.Id(), "MaintainTime", maintainTime, regionId)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifySelectDBInstanceMaintainTime", AlibabaCloudSdkGoERROR)
		}
		d.SetPartial("maintain_start_time")
		d.SetPartial("maintain_end_time")
	}

	// Handle security IP groups update
	if d.HasChange("security_ip_groups") {
		oldValue, newValue := d.GetChange("security_ip_groups")
		oldGroups := oldValue.(*schema.Set).List()
		newGroups := newValue.(*schema.Set).List()

		// Debug: print old and new groups changes
		fmt.Printf("[DEBUG] Security IP groups change - oldGroups: %+v, newGroups: %+v\n", oldGroups, newGroups)

		// Process each new group
		for _, newGroupInterface := range newGroups {
			newGroup := newGroupInterface.(map[string]interface{})
			groupName := newGroup["group_name"].(string)
			if groupName == "" {
				groupName = "default"
			}
			modifyMode := newGroup["modify_mode"].(string)
			if modifyMode == "" {
				modifyMode = "Cover"
			}
			securityIpList := newGroup["security_ip_list"].(*schema.Set).List()

			// Convert security IP list to comma-separated string
			var ipStrings []string
			for _, ip := range securityIpList {
				ipStrings = append(ipStrings, ip.(string))
			}
			securityIps := strings.Join(ipStrings, ",")

			// Build modification request
			modification := &selectdb.SecurityIPListModification{
				DBInstanceId:   d.Id(),
				GroupName:      groupName,
				SecurityIPList: securityIps,
				ModifyMode:     modifyMode,
				RegionId:       client.RegionId,
			}

			// Use retry for security IP modification
			action := "ModifySelectDBSecurityIPList"
			err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
				_, err := selectDBService.ModifySelectDBSecurityIPList(modification)
				if err != nil {
					if NeedRetry(err) {
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				return nil
			})

			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
			}

			// Wait for security IP list to be updated
			stateConf := BuildStateConf(
				[]string{},
				[]string{"Available"},
				d.Timeout(schema.TimeoutUpdate),
				5*time.Second,
				selectDBService.SelectDBSecurityIPListStateRefreshFunc(d.Id(), groupName, []string{}),
			)

			if _, err := stateConf.WaitForState(); err != nil {
				return WrapErrorf(err, IdMsg, d.Id())
			}
		}

		d.SetPartial("security_ip_groups")
	}

	// Handle tags update
	if d.HasChange("tags") {
		oraw, nraw := d.GetChange("tags")
		old := oraw.(map[string]interface{})
		new := nraw.(map[string]interface{})

		// Convert to string maps
		added := make(map[string]string)
		removed := make(map[string]string)

		// Find added tags
		for key, value := range new {
			if _, exists := old[key]; !exists {
				added[key] = value.(string)
			} else if old[key] != value {
				added[key] = value.(string)
			}
		}

		// Find removed tags
		for key, value := range old {
			if _, exists := new[key]; !exists {
				removed[key] = value.(string)
			}
		}

		if err := selectDBService.SetResourceTags(d.Id(), added, removed); err != nil {
			return WrapError(err)
		}
		d.SetPartial("tags")
	}

	// Wait for all modifications to complete
	err = selectDBService.WaitForSelectDBInstanceUpdated(d.Id(), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	d.Partial(false)
	return resourceAliCloudSelectDBInstanceRead(d, meta)
}

func resourceAliCloudSelectDBInstanceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	selectDBService, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Id()
	instance, err := selectDBService.DescribeSelectDBInstance(instanceId)
	if err != nil {
		if IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set basic information
	d.Set("description", instance.Description)
	d.Set("engine", instance.Engine)
	d.Set("engine_version", instance.EngineVersion)
	d.Set("status", instance.Status)
	d.Set("region_id", instance.RegionId)
	d.Set("zone_id", instance.ZoneId)
	d.Set("vpc_id", instance.VpcId)
	d.Set("vswitch_id", instance.VswitchId)

	// Set payment type
	if instance.ChargeType == "POSTPAY" {
		d.Set("payment_type", "PayAsYouGo")
	} else if instance.ChargeType == "PREPAY" {
		d.Set("payment_type", "Subscription")
	}

	// Set charge type
	d.Set("charge_type", instance.ChargeType)
	d.Set("category", instance.Category)
	d.Set("instance_used_type", instance.InstanceUsedType)
	d.Set("connection_string", instance.ConnectionString)
	d.Set("sub_domain", instance.SubDomain)

	// Set resource configuration
	d.Set("resource_cpu", instance.ResourceCpu)
	d.Set("resource_memory", instance.ResourceMemory)
	d.Set("storage_size", instance.StorageSize)
	d.Set("storage_type", instance.StorageType)
	d.Set("object_store_size", instance.ObjectStoreSize)
	d.Set("cluster_count", instance.ClusterCount)
	d.Set("scale_min", instance.ScaleMin)
	d.Set("scale_max", instance.ScaleMax)
	d.Set("scale_replica", instance.ScaleReplica)

	// Set lock information
	d.Set("lock_mode", instance.LockMode)
	d.Set("lock_reason", instance.LockReason)

	// Set time information
	d.Set("create_time", instance.CreateTime)
	d.Set("gmt_created", instance.GmtCreated)
	d.Set("gmt_modified", instance.GmtModified)
	d.Set("expire_time", instance.ExpireTime)
	d.Set("maintain_start_time", instance.MaintainStarttime)
	d.Set("maintain_end_time", instance.MaintainEndtime)

	// Set upgrade information
	d.Set("can_upgrade_versions", instance.CanUpgradeVersions)

	// Set resource group
	d.Set("resource_group_id", instance.ResourceGroupId)

	// Set multi-zone information
	if len(instance.MultiZone) > 0 {
		multiZoneList := make([]map[string]interface{}, 0)
		for _, mz := range instance.MultiZone {
			multiZone := map[string]interface{}{
				"zone_id":            mz.ZoneId,
				"vswitch_ids":        mz.VSwitchIds,
				"cidr":               mz.Cidr,
				"available_ip_count": mz.AvailableIpCount,
			}
			multiZoneList = append(multiZoneList, multiZone)
		}
		d.Set("multi_zone", multiZoneList)
	}

	// Set tags
	if len(instance.Tags) > 0 {
		tags := make(map[string]interface{})
		for _, tag := range instance.Tags {
			if !ignoredTags(tag.Key, tag.Value) {
				tags[tag.Key] = tag.Value
			}
		}
		d.Set("tags", tags)
	}

	// Set cluster information
	if len(instance.DBClusterList) > 0 {
		clusterList := make([]map[string]interface{}, 0)

		// Find default cache size from BE cluster
		defaultCacheSize := 0
		defaultBeClusterId := instanceId + "-be"

		for _, cluster := range instance.DBClusterList {
			clusterInfo := map[string]interface{}{
				"db_cluster_id":         cluster.ClusterId,
				"db_cluster_name":       cluster.ClusterName,
				"db_cluster_class":      cluster.ClusterClass,
				"status":                cluster.Status,
				"charge_type":           cluster.ChargeType,
				"cpu_cores":             cluster.CpuCores,
				"memory":                cluster.Memory,
				"cache_storage_size_gb": cluster.CacheStorageSizeGB,
				"cache_storage_type":    cluster.CacheStorageType,
				"performance_level":     cluster.PerformanceLevel,
				"scaling_rules_enable":  cluster.ScalingRulesEnable,
				"created_time":          cluster.CreatedTime,
				"modified_time":         cluster.ModifiedTime,
			}
			clusterList = append(clusterList, clusterInfo)

			// Set instance class and cache size from default BE cluster
			if cluster.ClusterId == defaultBeClusterId {
				d.Set("instance_class", cluster.ClusterClass)
				if cacheSize, err := strconv.Atoi(cluster.CacheStorageSizeGB); err == nil {
					defaultCacheSize = cacheSize
				}
			}
		}

		d.Set("db_cluster_list", clusterList)
		if defaultCacheSize > 0 {
			d.Set("cache_size", defaultCacheSize)
		}
	}

	// Get network information and security IP lists
	query := &selectdb.SecurityIPListQuery{
		DBInstanceId: instanceId,
		RegionId:     client.RegionId,
	}

	securityIPResult, err := selectDBService.DescribeSelectDBSecurityIPList(query)
	if err != nil && !IsNotFoundError(err) {
		return WrapError(err)
	}

	// Set security IP groups
	if securityIPResult != nil && len(securityIPResult.GroupItems) > 0 {
		securityIPGroups := make([]map[string]interface{}, 0)
		for _, group := range securityIPResult.GroupItems {
			if len(group.SecurityIPList) > 0 {
				securityIPGroup := map[string]interface{}{
					"group_name":       group.GroupName,
					"security_ip_list": group.SecurityIPList,
					"modify_mode":      "Cover", // Default mode
				}
				securityIPGroups = append(securityIPGroups, securityIPGroup)
			}
		}
		d.Set("security_ip_groups", securityIPGroups)
	}

	// Set security IP lists for computed display
	if securityIPResult != nil && len(securityIPResult.GroupItems) > 0 {
		securityIPLists := make([]map[string]interface{}, 0)
		for _, group := range securityIPResult.GroupItems {
			securityIPList := map[string]interface{}{
				"group_name":         group.GroupName,
				"group_tag":          group.GroupTag,
				"security_ip_type":   "", // This would need to be retrieved from the API
				"security_ip_list":   group.SecurityIPList,
				"whitelist_net_type": group.WhitelistNetType,
			}
			securityIPLists = append(securityIPLists, securityIPList)
		}
		d.Set("security_ip_lists", securityIPLists)
	} else {
		d.Set("security_ip_lists", []map[string]interface{}{})
	}

	// Get network information
	// This would require a separate API call in the real implementation
	// For now, we'll set empty lists to avoid nil pointer errors
	d.Set("instance_net_infos", []map[string]interface{}{})

	return nil
}

func resourceAliCloudSelectDBInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	selectDBService, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	// Wait for instance to be in a stable state before deletion
	err = selectDBService.WaitForSelectDBInstanceUpdated(d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil {
		// If instance is not found, it's already deleted
		if IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// Delete the instance
	err = selectDBService.DeleteSelectDBInstance(d.Id())
	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteDBInstance", AlibabaCloudSdkGoERROR)
	}

	// Wait for deletion to complete
	err = selectDBService.WaitForSelectDBInstanceDeleted(d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
