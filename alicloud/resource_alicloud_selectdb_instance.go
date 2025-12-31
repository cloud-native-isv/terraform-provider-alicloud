package alicloud

import (
	"fmt"
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
			"instance_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The description of the SelectDB instance.",
			},

			// ======== Resource Organization ========
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group ID.",
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "A mapping of tags to assign to the resource.",
			},

			// ======== Database Engine Configuration ========
			"engine": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "selectdb",
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
				Computed:    true,
				ForceNew:    true,
				Description: "Multi-zone configuration for high availability. Required when deploy_scheme is multi_az.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The zone ID.",
						},
						"vswitch_ids": {
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "The VSwitch IDs in this zone.",
						},
						"cidr": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The CIDR block.",
						},
						"available_ip_count": {
							Type:        schema.TypeInt,
							Computed:    true,
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
				Default:     "single_az",
				Description: "The deployment scheme of the SelectDB instance.",
			},
			// ======== Billing Configuration ========
			"charge_type": {
				Type:         schema.TypeString,
				ValidateFunc: StringInSlice([]string{"Prepaid", "Postpaid"}, false),
				Required:     true,
				ForceNew:     true,
				Description:  "The payment type of the SelectDB instance. Valid values: Prepaid, Postpaid.",
			},
			"period": {
				Type:             schema.TypeString,
				ValidateFunc:     StringInSlice([]string{string(Year), string(Month)}, false),
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: selectdbPostPaidDiffSuppressFunc,
				Description:      "The billing period for Prepaid instances. Valid values: Month, Year.",
			},
			"period_time": {
				Type:             schema.TypeInt,
				ValidateFunc:     IntInSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 12, 24, 36}),
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: selectdbPostPaidDiffSuppressFunc,
				Description:      "The period time for Prepaid instances.",
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

			// ======== Database Account Configuration ========
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The admin username for the SelectDB instance. Once set, cannot be changed.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The admin password for the SelectDB instance. This is write-only and cannot be read back.",
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
					},
				},
			},

			// ======== Computed Information - Basic Instance ========
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the SelectDB instance.",
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
			"cluster_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The database cluster list of the instance.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The cluster ID.",
						},
						"cluster_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The cluster name.",
						},
						"cluster_class": {
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
						"memory_gb": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The memory size in GB.",
						},
						"cache_size_gb": {
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

	deployScheme := d.Get("deploy_scheme").(string)
	if deployScheme == "multi_az" {
		if v, ok := d.GetOk("multi_zone"); ok {
			list := v.([]interface{})
			if len(list) != 3 {
				return fmt.Errorf("The number of zones (multi_zone field) must be three when deploy_scheme is multi_az")
			}
		} else {
			return fmt.Errorf("multi_zone field is required when deploy_scheme is multi_az")
		}
	}

	// Build instance directly using Instance struct fields
	instance := &selectdb.Instance{
		Engine:   "selectdb",
		Name:     d.Get("instance_name").(string),
		RegionId: client.RegionId,
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

	if v, ok := d.GetOk("multi_zone"); ok {
		multiZones := v.([]interface{})
		for _, mz := range multiZones {
			mzMap := mz.(map[string]interface{})
			vswitchIds := make([]string, 0)
			if vIds, ok := mzMap["vswitch_ids"].([]interface{}); ok {
				for _, id := range vIds {
					vswitchIds = append(vswitchIds, id.(string))
				}
			}
			instance.MultiZone = append(instance.MultiZone, selectdb.MultiZone{
				ZoneId:     mzMap["zone_id"].(string),
				VSwitchIds: vswitchIds,
			})
		}
	}

	// Set payment configuration
	instance.ChargeType = d.Get("charge_type").(string)
	if instance.ChargeType == "Prepaid" {

		// Set period for subscription instances
		if v, ok := d.GetOk("period"); ok {
			instance.Period = v.(string)
		}
		if v, ok := d.GetOk("period_time"); ok {
			instance.UsedTime = int32(v.(int))
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
	instanceId, err := selectDBService.CreateSelectDBInstance(instance)
	if err != nil {
		return WrapError(err)
	}

	if instanceId == nil {
		return WrapErrorf(fmt.Errorf("create result is nil or instance ID is empty"), IdMsg, "alicloud_selectdb_instance")
	}

	d.SetId(*instanceId)

	// Wait for instance to be ready
	err = selectDBService.WaitForSelectDBInstanceCreated(d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// Set admin password if provided
	if username, ok := d.GetOk("username"); ok {
		if password, passwordOk := d.GetOk("password"); passwordOk {
			err = selectDBService.resetSelectDBInstancePassword(d.Id(), username.(string), password.(string))
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ResetSelectDBInstancePassword", AlibabaCloudSdkGoERROR)
			}
		}
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
	if d.HasChange("instance_name") {
		description := d.Get("instance_name").(string)

		err := selectDBService.ModifySelectDBInstance(d.Id(), "DBInstanceDescription", description)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifySelectDBInstanceDescription", AlibabaCloudSdkGoERROR)
		}
		d.SetPartial("instance_name")
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
				err = selectDBService.UpgradeSelectDBInstanceEngineVersion(d.Id(), upgradeTargetVersion)
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

		err := selectDBService.ModifySelectDBInstance(d.Id(), "MaintainTime", maintainTime)
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

		// Calculate the modifications needed
		modifications := calculateSecurityIPGroupChanges(oldGroups, newGroups)

		// Apply each modification
		for _, modification := range modifications {
			// Set common fields
			modification.DBInstanceId = d.Id()
			modification.RegionId = client.RegionId

			// Use retry for security IP modification
			action := "ModifySelectDBSecurityIPList"
			err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
				_, err := selectDBService.ModifySelectDBSecurityIPList(
					modification.DBInstanceId,
					modification.SecurityIPList,
					modification.GroupName,
					modification.ModifyMode,
					modification.RegionId,
					modification.ResourceOwnerId,
				)
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
				selectDBService.SelectDBSecurityIPListStateRefreshFunc(d.Id(), modification.GroupName, []string{}),
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

	// Handle password update
	if d.HasChange("password") {
		if password, ok := d.GetOk("password"); ok {
			username := d.Get("username").(string)
			if username == "" {
				return WrapErrorf(fmt.Errorf("username must be provided when setting password"), DefaultErrorMsg, d.Id(), "ResetSelectDBInstancePassword", AlibabaCloudSdkGoERROR)
			}

			err := selectDBService.resetSelectDBInstancePassword(d.Id(), username, password.(string))
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ResetSelectDBInstancePassword", AlibabaCloudSdkGoERROR)
			}
		}
		d.SetPartial("password")
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
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set basic information
	d.Set("instance_name", instance.Name)
	d.Set("engine", instance.Engine)
	d.Set("engine_version", instance.EngineVersion)
	d.Set("status", instance.Status)
	d.Set("zone_id", instance.ZoneId)
	d.Set("vpc_id", instance.VpcId)
	d.Set("vswitch_id", instance.VswitchId)
	// Set charge type
	d.Set("charge_type", instance.ChargeType)
	d.Set("deploy_scheme", instance.DeployScheme)
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

	// Set username (password is write-only, so we don't set it from API response)
	// Username is preserved from the Terraform state
	if username, ok := d.GetOk("username"); ok {
		d.Set("username", username.(string))
	}

	// Set multi-zone information (computed field)
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
	} else {
		d.Set("multi_zone", []map[string]interface{}{})
	} // Set tags
	tags := make(map[string]interface{})
	if len(instance.Tags) > 0 {
		for _, tag := range instance.Tags {
			if !ignoredTags(tag.Key, tag.Value) {
				tags[tag.Key] = tag.Value
			}
		}
	}
	d.Set("tags", tags)

	// Set cluster information
	clusterList := make([]map[string]interface{}, 0)
	if len(instance.DBClusterList) > 0 {
		// Find default cache size from BE cluster
		defaultCacheSize := 0
		defaultBeClusterId := instanceId + "-be"

		for _, cluster := range instance.DBClusterList {
			clusterInfo := map[string]interface{}{
				"cluster_id":           cluster.ClusterId,
				"cluster_name":         cluster.ClusterName,
				"cluster_class":        cluster.ClusterClass,
				"status":               cluster.Status,
				"charge_type":          cluster.ChargeType,
				"cpu_cores":            int(cluster.CpuCores),
				"memory_gb":            int(cluster.Memory),
				"cache_size_gb":        fmt.Sprintf("%d", cluster.CacheSize),
				"cache_storage_type":   cluster.CacheStorageType,
				"performance_level":    cluster.PerformanceLevel,
				"scaling_rules_enable": cluster.ScalingRulesEnable,
				"created_time":         cluster.CreatedTime,
				"modified_time":        cluster.ModifiedTime,
			}
			clusterList = append(clusterList, clusterInfo)

			// Set instance class and cache size from default BE cluster
			if cluster.ClusterId == defaultBeClusterId {
				d.Set("instance_class", cluster.ClusterClass)
				defaultCacheSize = int(cluster.CacheSize)
			}
		}

		if defaultCacheSize > 0 {
			d.Set("cache_size", defaultCacheSize)
		}
	}

	err = d.Set("cluster_list", clusterList)

	// Get network information and security IP lists
	securityIPGroups, err := selectDBService.DescribeSelectDBSecurityIPList(instanceId, client.RegionId, 0)
	if err != nil && !NotFoundError(err) {
		return WrapError(err)
	}

	// Set security IP groups
	if len(securityIPGroups) > 0 {
		securityIPGroupsData := make([]map[string]interface{}, 0)
		for _, group := range securityIPGroups {
			if len(group.SecurityIPList) > 0 {
				securityIPGroup := map[string]interface{}{
					"group_name":       group.GroupName,
					"security_ip_list": group.SecurityIPList,
				}
				securityIPGroupsData = append(securityIPGroupsData, securityIPGroup)
			}
		}
		d.Set("security_ip_groups", securityIPGroupsData)
	}

	// Set security IP lists for computed display
	if len(securityIPGroups) > 0 {
		securityIPLists := make([]map[string]interface{}, 0)
		for _, group := range securityIPGroups {
			securityIPList := map[string]interface{}{
				"group_name":         group.GroupName,
				"group_tag":          group.GroupTag,
				"security_ip_type":   group.SecurityIPType,
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
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// Delete the instance
	err = selectDBService.DeleteSelectDBInstance(d.Id())
	if err != nil {
		if NotFoundError(err) {
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

// SecurityIPModification represents a security IP list modification
type SecurityIPModification struct {
	GroupName       string
	SecurityIPList  string
	ModifyMode      string
	DBInstanceId    string
	RegionId        string
	ResourceOwnerId int64
}

// calculateSecurityIPGroupChanges calculates the modifications needed for security IP groups
// Returns a list of modifications with appropriate modify_mode values
func calculateSecurityIPGroupChanges(oldGroups, newGroups []interface{}) []*SecurityIPModification {
	var modifications []*SecurityIPModification

	// Create maps for easier lookup
	oldGroupsMap := make(map[string]map[string]interface{})
	newGroupsMap := make(map[string]map[string]interface{})

	// Build old groups map
	for _, oldGroupInterface := range oldGroups {
		oldGroup := oldGroupInterface.(map[string]interface{})
		groupName := oldGroup["group_name"].(string)
		if groupName == "" {
			groupName = "default"
		}
		oldGroupsMap[groupName] = oldGroup
	}

	// Build new groups map
	for _, newGroupInterface := range newGroups {
		newGroup := newGroupInterface.(map[string]interface{})
		groupName := newGroup["group_name"].(string)
		if groupName == "" {
			groupName = "default"
		}
		newGroupsMap[groupName] = newGroup
	}

	// Find groups to add or update
	for groupName, newGroup := range newGroupsMap {
		var modifyMode string
		oldGroup, existsInOld := oldGroupsMap[groupName]

		if !existsInOld {
			// Group doesn't exist in old, so add it
			modifyMode = "1" // Append/Add
		} else {
			// Group exists in both, compare security IP lists
			oldIPSet := oldGroup["security_ip_list"].(*schema.Set)
			newIPSet := newGroup["security_ip_list"].(*schema.Set)

			// Check if IP lists are different
			if !oldIPSet.Equal(newIPSet) {
				modifyMode = "0" // Cover/Update
			} else {
				// No change needed
				continue
			}
		}

		// Convert security IP list to comma-separated string
		securityIpList := newGroup["security_ip_list"].(*schema.Set).List()
		var ipStrings []string
		for _, ip := range securityIpList {
			ipStrings = append(ipStrings, ip.(string))
		}
		securityIps := strings.Join(ipStrings, ",")

		modifications = append(modifications, &SecurityIPModification{
			GroupName:      groupName,
			SecurityIPList: securityIps,
			ModifyMode:     modifyMode,
		})
	}

	// Find groups to delete (exists in old but not in new)
	for groupName, oldGroup := range oldGroupsMap {
		if _, existsInNew := newGroupsMap[groupName]; !existsInNew {
			// Group exists in old but not in new, so delete it
			// Convert security IP list to comma-separated string for deletion
			securityIpList := oldGroup["security_ip_list"].(*schema.Set).List()
			var ipStrings []string
			for _, ip := range securityIpList {
				ipStrings = append(ipStrings, ip.(string))
			}
			securityIps := strings.Join(ipStrings, ",")

			modifications = append(modifications, &SecurityIPModification{
				GroupName:      groupName,
				SecurityIPList: securityIps,
				ModifyMode:     "2", // Delete
			})
		}
	}

	return modifications
}
