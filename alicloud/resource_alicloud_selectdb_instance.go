package alicloud

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
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
			// Updatable fields based on Instance struct
			"db_instance_description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The description of the SelectDB instance.",
			},
			"engine_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The engine version of the SelectDB instance. Can be updated to upgrade the engine version.",
			},
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

			// Force new fields - cannot be updated after creation
			"db_instance_class": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The instance class of the SelectDB instance.",
			},
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
			"cache_size": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The cache size of the SelectDB instance in GB.",
			},
			"engine": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "SelectDB",
				Description: "The engine type of the SelectDB instance. Default is SelectDB.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The resource group ID.",
			},
			"deploy_scheme": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The deployment scheme of the SelectDB instance.",
			},
			"security_ip_list": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The security IP list for accessing the instance.",
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

			// Deprecated fields for backward compatibility
			"engine_minor_version": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"engine_version"},
				Deprecated:    "Field `engine_minor_version` has been deprecated. Use `engine_version` instead.",
				Description:   "Deprecated: Use engine_version instead.",
			},
			"upgraded_engine_minor_version": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"engine_version", "engine_minor_version"},
				Deprecated:    "Field `upgraded_engine_minor_version` has been deprecated. Use `engine_version` instead.",
				Description:   "Deprecated: Use engine_version instead.",
			},
			"maintenance_window": {
				Type:        schema.TypeString,
				Optional:    true,
				Deprecated:  "Field `maintenance_window` has been deprecated. Use `maintain_start_time` and `maintain_end_time` instead.",
				Description: "Deprecated: Use maintain_start_time and maintain_end_time instead.",
			},
			"enable_public_network": {
				Type:        schema.TypeBool,
				Optional:    true,
				Deprecated:  "Field `enable_public_network` has been deprecated. Use separate connection management resources instead.",
				Description: "Deprecated: Use separate connection management resources instead.",
			},
			"admin_pass": {
				Type:        schema.TypeString,
				Sensitive:   true,
				Optional:    true,
				Deprecated:  "Field `admin_pass` has been deprecated. Use separate account management resources instead.",
				Description: "Deprecated: Use separate account management resources instead.",
			},
			"desired_security_ip_lists": {
				Type:        schema.TypeList,
				Optional:    true,
				Deprecated:  "Field `desired_security_ip_lists` has been deprecated. Use separate security IP management resources instead.",
				Description: "Deprecated: Use separate security IP management resources instead.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"security_ip_list": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			// Tags support
			"tags": tagsSchema(),

			// Computed values - instance information
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

			// Resource configuration
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

			// Lock information
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

			// Time information
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

			// Upgrade information
			"can_upgrade_versions": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of engine versions that the instance can upgrade to.",
			},

			// Instance network information
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

			// Security IP lists
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

			// DB cluster list
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

			// Deprecated computed fields for backward compatibility
			"cpu_prepaid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Deprecated:  "Field `cpu_prepaid` has been deprecated. Use `resource_cpu` instead.",
				Description: "Deprecated: Use resource_cpu instead.",
			},
			"memory_prepaid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Deprecated:  "Field `memory_prepaid` has been deprecated. Use `resource_memory` instead.",
				Description: "Deprecated: Use resource_memory instead.",
			},
			"cache_size_prepaid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Deprecated:  "Field `cache_size_prepaid` has been deprecated. Use `storage_size` instead.",
				Description: "Deprecated: Use storage_size instead.",
			},
			"cluster_count_prepaid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Deprecated:  "Field `cluster_count_prepaid` has been deprecated. Use `cluster_count` instead.",
				Description: "Deprecated: Use cluster_count instead.",
			},
			"cpu_postpaid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Deprecated:  "Field `cpu_postpaid` has been deprecated. Use `resource_cpu` instead.",
				Description: "Deprecated: Use resource_cpu instead.",
			},
			"memory_postpaid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Deprecated:  "Field `memory_postpaid` has been deprecated. Use `resource_memory` instead.",
				Description: "Deprecated: Use resource_memory instead.",
			},
			"cache_size_postpaid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Deprecated:  "Field `cache_size_postpaid` has been deprecated. Use `storage_size` instead.",
				Description: "Deprecated: Use storage_size instead.",
			},
			"cluster_count_postpaid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Deprecated:  "Field `cluster_count_postpaid` has been deprecated. Use `cluster_count` instead.",
				Description: "Deprecated: Use cluster_count instead.",
			},
			"gmt_expired": {
				Type:        schema.TypeString,
				Computed:    true,
				Deprecated:  "Field `gmt_expired` has been deprecated. Use `expire_time` instead.",
				Description: "Deprecated: Use expire_time instead.",
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
		Description: d.Get("db_instance_description").(string),
		RegionId:    client.RegionId,
	}

	// Set basic configuration
	if v, ok := d.GetOk("engine_version"); ok {
		instance.EngineVersion = v.(string)
	} else if v, ok := d.GetOk("engine_minor_version"); ok {
		instance.EngineVersion = v.(string)
	} else if v, ok := d.GetOk("upgraded_engine_minor_version"); ok {
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

	// Set payment configuration
	paymentType := d.Get("payment_type").(string)
	if paymentType == "PayAsYouGo" {
		instance.ChargeType = "POSTPAY"
	} else if paymentType == "Subscription" {
		instance.ChargeType = "PREPAY"
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
	stateConf := BuildStateConf([]string{"CREATING", "RESOURCE_PREPARING"}, []string{"ACTIVATION"},
		d.Timeout(schema.TimeoutCreate), 1*time.Minute,
		selectDBService.SelectDBInstanceStateRefreshFunc(d.Id(), []string{"DELETING", "FAILED"}))

	if _, err := stateConf.WaitForState(); err != nil {
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
	if d.HasChange("db_instance_description") {
		description := d.Get("db_instance_description").(string)
		regionId := d.Get("region_id").(string)

		err := selectDBService.ModifySelectDBInstance(d.Id(), "DBInstanceDescription", description, regionId)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifySelectDBInstanceDescription", AlibabaCloudSdkGoERROR)
		}
		d.SetPartial("db_instance_description")
	}

	// Handle engine version upgrade
	if d.HasChange("engine_version") || d.HasChange("engine_minor_version") || d.HasChange("upgraded_engine_minor_version") {
		var newVersion string

		if d.HasChange("engine_version") {
			_, newVersionInterface := d.GetChange("engine_version")
			newVersion = newVersionInterface.(string)
		} else if d.HasChange("engine_minor_version") {
			_, newVersionInterface := d.GetChange("engine_minor_version")
			newVersion = newVersionInterface.(string)
		} else if d.HasChange("upgraded_engine_minor_version") {
			_, newVersionInterface := d.GetChange("upgraded_engine_minor_version")
			newVersion = newVersionInterface.(string)
		}

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
				stateConf := BuildStateConf([]string{"MODULE_UPGRADING"}, []string{"ACTIVATION"},
					d.Timeout(schema.TimeoutUpdate), 10*time.Second,
					selectDBService.SelectDBInstanceStateRefreshFunc(d.Id(), []string{"DELETING", "FAILED"}))

				if _, err := stateConf.WaitForState(); err != nil {
					return WrapErrorf(err, IdMsg, d.Id())
				}
			} else {
				return WrapErrorf(fmt.Errorf("invalid upgrade version for %s, cannot upgrade to %s", d.Id(), newVersion), DefaultErrorMsg, d.Id(), "UpgradeSelectDBInstanceEngineVersion", AlibabaCloudSdkGoERROR)
			}
		}

		d.SetPartial("engine_version")
		d.SetPartial("engine_minor_version")
		d.SetPartial("upgraded_engine_minor_version")
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

	// Handle tags update
	if d.HasChange("tags") {
		// TODO: Implement tag management for SelectDB instances
		// added, removed := parsingTags(d)
		// if err := selectDBService.SetResourceTags(d.Id(), added, removed); err != nil {
		// 	return WrapError(err)
		// }
		d.SetPartial("tags")
	}

	// Wait for all modifications to complete
	stateConf := BuildStateConf([]string{"RESOURCE_PREPARING", "CREATING", "CLASS_CHANGING", "MODULE_UPGRADING"},
		[]string{"ACTIVATION"}, d.Timeout(schema.TimeoutUpdate), 10*time.Second,
		selectDBService.SelectDBInstanceStateRefreshFunc(d.Id(), []string{"DELETING", "FAILED"}))

	if _, err := stateConf.WaitForState(); err != nil {
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
	d.Set("db_instance_description", instance.Description)
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

	// Set deprecated fields for backward compatibility
	d.Set("engine_minor_version", instance.EngineVersion)
	d.Set("upgraded_engine_minor_version", instance.EngineVersion)
	d.Set("gmt_expired", instance.ExpireTime)

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
				"db_cluster_id":         cluster.DbClusterId,
				"db_cluster_name":       cluster.DbClusterName,
				"db_cluster_class":      cluster.DbClusterClass,
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
			if cluster.DbClusterId == defaultBeClusterId {
				d.Set("db_instance_class", cluster.DbClusterClass)
				if cacheSize, err := strconv.Atoi(cluster.CacheStorageSizeGB); err == nil {
					defaultCacheSize = cacheSize
				}
			}
		}

		d.Set("db_cluster_list", clusterList)
		if defaultCacheSize > 0 {
			d.Set("cache_size", defaultCacheSize)
		}

		// Set deprecated computed fields for backward compatibility
		cpuPrepaid := 0
		cpuPostpaid := 0
		memPrepaid := 0
		memPostpaid := 0
		cachePrepaid := 0
		cachePostpaid := 0
		clusterPrepaidCount := 0
		clusterPostpaidCount := 0

		for _, cluster := range instance.DBClusterList {
			if cluster.ChargeType == "Postpaid" {
				cpuPostpaid += int(cluster.CpuCores)
				memPostpaid += int(cluster.Memory)
				if cacheSize, err := strconv.Atoi(cluster.CacheStorageSizeGB); err == nil {
					cachePostpaid += cacheSize
				}
				clusterPostpaidCount++
			} else if cluster.ChargeType == "Prepaid" {
				cpuPrepaid += int(cluster.CpuCores)
				memPrepaid += int(cluster.Memory)
				if cacheSize, err := strconv.Atoi(cluster.CacheStorageSizeGB); err == nil {
					cachePrepaid += cacheSize
				}
				clusterPrepaidCount++
			}
		}

		d.Set("cpu_prepaid", cpuPrepaid)
		d.Set("memory_prepaid", memPrepaid)
		d.Set("cache_size_prepaid", cachePrepaid)
		d.Set("cluster_count_prepaid", clusterPrepaidCount)
		d.Set("cpu_postpaid", cpuPostpaid)
		d.Set("memory_postpaid", memPostpaid)
		d.Set("cache_size_postpaid", cachePostpaid)
		d.Set("cluster_count_postpaid", clusterPostpaidCount)
	}

	// Get network information
	// This would require a separate API call in the real implementation
	// For now, we'll set empty lists to avoid nil pointer errors
	d.Set("instance_net_infos", []map[string]interface{}{})
	d.Set("security_ip_lists", []map[string]interface{}{})

	return nil
}

func resourceAliCloudSelectDBInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	selectDBService, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	// Wait for instance to be in a stable state before deletion
	stateConf := BuildStateConf([]string{"RESOURCE_PREPARING", "CREATING", "CLASS_CHANGING", "MODULE_UPGRADING"},
		[]string{"ACTIVATION"}, d.Timeout(schema.TimeoutDelete), 10*time.Second,
		selectDBService.SelectDBInstanceStateRefreshFunc(d.Id(), []string{"DELETING", "FAILED"}))

	if _, err := stateConf.WaitForState(); err != nil {
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
	stateConf = BuildStateConf([]string{"DELETING", "ACTIVATION"}, []string{""},
		d.Timeout(schema.TimeoutDelete), 5*time.Second,
		selectDBService.SelectDBInstanceStateRefreshFunc(d.Id(), []string{}))

	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
