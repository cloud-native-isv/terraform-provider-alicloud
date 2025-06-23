package alicloud

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudNasFileSystem() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudNasFileSystemCreate,
		Read:   resourceAliCloudNasFileSystemRead,
		Update: resourceAliCloudNasFileSystemUpdate,
		Delete: resourceAliCloudNasFileSystemDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"capacity": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"encrypt_type": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: IntInSlice([]int{0, 1, 2}),
			},
			"file_system_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: StringInSlice([]string{"standard", "extreme", "cpfs"}, false),
			},
			"keytab": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"keytab_md5": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"nfs_acl": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_oplock": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"protocol_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: StringInSlice([]string{"NFS", "SMB", "cpfs"}, false),
			},
			"recycle_bin": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: StringInSlice([]string{"Enable", "Disable"}, false),
						},
						"reserved_days": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: IntBetween(0, 180),
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if v, ok := d.GetOk("recycle_bin.0.status"); ok && v.(string) == "Enable" {
									return false
								}
								return true
							},
						},
						"size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"secondary_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"enable_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"region_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"smb_acl": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"home_dir_path": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"enable_anonymous_access": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"super_admin_sid": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"reject_unencrypted_access": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"encrypt_data": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"snapshot_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: StringInSlice([]string{"Performance", "Capacity", "standard", "advance", "advance_100", "advance_200", "Premium"}, false),
			},
			"tags": tagsSchema(),
			"vswitch_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAliCloudNasFileSystemCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapError(err)
	}

	// Build FileSystem struct for service layer
	fileSystem := &nas.FileSystem{
		Description:     d.Get("description").(string),
		StorageType:     d.Get("storage_type").(string),
		ProtocolType:    d.Get("protocol_type").(string),
		ResourceGroupId: d.Get("resource_group_id").(string),
		ZoneId:          d.Get("zone_id").(string),
	}

	if v, ok := d.GetOk("file_system_type"); ok {
		fileSystem.FileSystemType = v.(string)
	}

	if v, ok := d.GetOk("vpc_id"); ok {
		fileSystem.VpcId = v.(string)
	}

	if v, ok := d.GetOk("vswitch_id"); ok {
		fileSystem.VSwitchId = v.(string)
	}

	if v, ok := d.GetOkExists("capacity"); ok {
		fileSystem.Capacity = int64(v.(int))
	}

	if v, ok := d.GetOkExists("encrypt_type"); ok {
		fileSystem.EncryptType = int32(v.(int))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		fileSystem.KMSKeyId = v.(string)
	}

	// Handle snapshot-based file system creation
	if v, ok := d.GetOk("snapshot_id"); ok {
		fileSystem.SnapshotId = v.(string)
	}

	// Create file system using service layer
	createdFileSystem, err := nasService.CreateNasFileSystem(fileSystem)
	if err != nil {
		return err
	}

	d.SetId(createdFileSystem.FileSystemId)

	stateConf := BuildStateConf([]string{}, []string{"Running"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, nasService.NasFileSystemStateRefreshFunc(d.Id(), "Status", []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudNasFileSystemRead(d, meta)
}

func resourceAliCloudNasFileSystemRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapError(err)
	}

	objectRaw, err := nasService.DescribeNasFileSystem(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_nas_file_system DescribeNasFileSystem Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set all basic attributes first
	d.Set("capacity", objectRaw["Capacity"])
	d.Set("create_time", objectRaw["CreateTime"])
	d.Set("description", objectRaw["Description"])
	d.Set("encrypt_type", objectRaw["EncryptType"])
	d.Set("file_system_type", objectRaw["FileSystemType"])
	d.Set("kms_key_id", objectRaw["KMSKeyId"])
	d.Set("protocol_type", objectRaw["ProtocolType"])
	d.Set("region_id", objectRaw["RegionId"])
	d.Set("resource_group_id", objectRaw["ResourceGroupId"])
	d.Set("status", objectRaw["Status"])
	d.Set("storage_type", objectRaw["StorageType"])
	d.Set("zone_id", objectRaw["ZoneId"])

	// Always set vpc_id and vswitch_id, even if empty, to maintain consistency
	if vpcId, ok := objectRaw["VpcId"]; ok {
		if vpcId != nil && fmt.Sprint(vpcId) != "" {
			d.Set("vpc_id", vpcId)
		} else {
			d.Set("vpc_id", "")
		}
	} else {
		d.Set("vpc_id", "")
	}

	// Check for VSwitchId first, fallback to QuorumVswId if needed
	if vswitchId, ok := objectRaw["VSwitchId"]; ok && vswitchId != nil && fmt.Sprint(vswitchId) != "" {
		d.Set("vswitch_id", vswitchId)
	} else if quorumVswId, ok := objectRaw["QuorumVswId"]; ok && quorumVswId != nil && fmt.Sprint(quorumVswId) != "" {
		d.Set("vswitch_id", quorumVswId)
	} else {
		d.Set("vswitch_id", "")
	}

	// Handle options block - always set even if empty
	optionsMaps := make([]map[string]interface{}, 0)
	if objectRaw["Options"] != nil {
		optionsRaw := objectRaw["Options"].(map[string]interface{})
		if len(optionsRaw) > 0 {
			optionsMap := make(map[string]interface{})
			optionsMap["enable_oplock"] = optionsRaw["EnableOplock"]
			optionsMaps = append(optionsMaps, optionsMap)
		}
	}
	if err := d.Set("options", optionsMaps); err != nil {
		return err
	}

	// Use values from API response for conditional checks instead of state
	fileSystemType := fmt.Sprint(objectRaw["FileSystemType"])
	protocolType := fmt.Sprint(objectRaw["ProtocolType"])

	// Handle conditional SMB ACL for standard + SMB
	if (fileSystemType == "standard") && (protocolType == "SMB") {
		objectRaw, err = nasService.DescribeFileSystemDescribeSmbAcl(d.Id())
		if err != nil && !NotFoundError(err) {
			return WrapError(err)
		}

		smbAclMaps := make([]map[string]interface{}, 0)
		if err == nil && objectRaw != nil {
			smbAclMap := make(map[string]interface{})
			smbAclMap["enable_anonymous_access"] = objectRaw["EnableAnonymousAccess"]
			smbAclMap["enabled"] = objectRaw["Enabled"]
			smbAclMap["encrypt_data"] = objectRaw["EncryptData"]
			smbAclMap["home_dir_path"] = objectRaw["HomeDirPath"]
			smbAclMap["reject_unencrypted_access"] = objectRaw["RejectUnencryptedAccess"]
			smbAclMap["super_admin_sid"] = objectRaw["SuperAdminSid"]
			smbAclMaps = append(smbAclMaps, smbAclMap)
		}
		if err := d.Set("smb_acl", smbAclMaps); err != nil {
			return err
		}
	} else {
		// For non-SMB or non-standard, ensure smb_acl is set to empty
		if err := d.Set("smb_acl", []map[string]interface{}{}); err != nil {
			return err
		}
	}

	// Handle conditional Recycle Bin for standard
	if fileSystemType == "standard" {
		objectRaw, err = nasService.DescribeFileSystemGetRecycleBinAttribute(d.Id())
		if err != nil && !NotFoundError(err) {
			return WrapError(err)
		}

		recycleBinMaps := make([]map[string]interface{}, 0)
		if err == nil && objectRaw != nil {
			recycleBinMap := make(map[string]interface{})
			recycleBinMap["enable_time"] = objectRaw["EnableTime"]
			recycleBinMap["reserved_days"] = objectRaw["ReservedDays"]
			recycleBinMap["secondary_size"] = objectRaw["SecondarySize"]
			recycleBinMap["size"] = objectRaw["Size"]
			recycleBinMap["status"] = objectRaw["Status"]
			recycleBinMaps = append(recycleBinMaps, recycleBinMap)
		}
		if err := d.Set("recycle_bin", recycleBinMaps); err != nil {
			return err
		}
	} else {
		// For non-standard, ensure recycle_bin is set to empty
		if err := d.Set("recycle_bin", []map[string]interface{}{}); err != nil {
			return err
		}
	}

	// Handle tags - always fetch and set
	objectRaw, err = nasService.DescribeFileSystemListTagResources(d.Id())
	if err != nil && !NotFoundError(err) {
		return WrapError(err)
	}

	tagsMaps := make(map[string]interface{})
	if err == nil && objectRaw != nil {
		if tagsRaw, err := jsonpath.Get("$.TagResources.TagResource", objectRaw); err == nil {
			tagsMaps = tagsToMap(tagsRaw)
		}
	}
	d.Set("tags", tagsMaps)

	// Handle conditional NFS ACL for standard + NFS
	if (fileSystemType == "standard") && (protocolType == "NFS") {
		objectRaw, err = nasService.DescribeFileSystemDescribeNfsAcl(d.Id())
		if err != nil && !NotFoundError(err) {
			return WrapError(err)
		}

		nfsAclMaps := make([]map[string]interface{}, 0)
		if err == nil && objectRaw != nil {
			nfsAclMap := make(map[string]interface{})
			nfsAclMap["enabled"] = objectRaw["Enabled"]
			nfsAclMaps = append(nfsAclMaps, nfsAclMap)
		}
		if err := d.Set("nfs_acl", nfsAclMaps); err != nil {
			return err
		}
	} else {
		// For non-NFS or non-standard, ensure nfs_acl is set to empty
		if err := d.Set("nfs_acl", []map[string]interface{}{}); err != nil {
			return err
		}
	}

	return nil
}

func resourceAliCloudNasFileSystemUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapError(err)
	}

	d.Partial(true)

	// Update basic file system properties (description, options)
	if err := updateFileSystemBasicProperties(d, nasService); err != nil {
		return err
	}

	// Update capacity (for non-standard file systems)
	if err := updateFileSystemCapacity(d, nasService); err != nil {
		return err
	}

	// Update resource group
	if err := updateFileSystemResourceGroup(d, client); err != nil {
		return err
	}

	// Update recycle bin configuration (for standard file systems)
	if err := updateFileSystemRecycleBin(d, nasService); err != nil {
		return err
	}

	// Update NFS ACL (for standard + NFS file systems)
	if err := updateFileSystemNfsAcl(d, nasService); err != nil {
		return err
	}

	// Update SMB ACL (for standard + SMB file systems)
	if err := updateFileSystemSmbAcl(d, nasService); err != nil {
		return err
	}

	// Update tags
	if d.HasChange("tags") {
		if err := nasService.SetResourceTags(d, "filesystem"); err != nil {
			return WrapError(err)
		}
	}

	d.Partial(false)
	return resourceAliCloudNasFileSystemRead(d, meta)
}

// updateFileSystemBasicProperties handles updates to description and options
func updateFileSystemBasicProperties(d *schema.ResourceData, nasService *NasService) error {
	if !d.HasChanges("description", "options") {
		return nil
	}

	request := map[string]interface{}{
		"FileSystemId": d.Id(),
	}

	if d.HasChange("description") {
		request["Description"] = d.Get("description")
	}

	if d.HasChange("options") {
		if v := d.Get("options"); v != nil {
			optionsMap := make(map[string]interface{})
			if enableOplock, err := jsonpath.Get("$[0].enable_oplock", v); err == nil && enableOplock != nil {
				optionsMap["EnableOplock"] = enableOplock
			}

			if len(optionsMap) > 0 {
				optionsJson, err := json.Marshal(optionsMap)
				if err != nil {
					return WrapError(err)
				}
				request["Options"] = string(optionsJson)
			}
		}
	}

	return nasService.ModifyFileSystem(request)
}

// updateFileSystemCapacity handles capacity upgrades for non-standard file systems
func updateFileSystemCapacity(d *schema.ResourceData, nasService *NasService) error {
	if !d.HasChange("capacity") {
		return nil
	}

	// Only non-standard file systems support capacity upgrades
	fileSystemType := d.Get("file_system_type").(string)
	if fileSystemType == "standard" {
		return nil
	}

	request := map[string]interface{}{
		"FileSystemId": d.Id(),
		"Capacity":     d.Get("capacity"),
		"DryRun":       "false",
		"ClientToken":  buildClientToken("UpgradeFileSystem"),
	}

	if err := nasService.UpgradeFileSystem(request); err != nil {
		return err
	}

	// Wait for upgrade completion
	stateConf := BuildStateConf([]string{}, []string{"Running"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, nasService.NasFileSystemStateRefreshFunc(d.Id(), "Status", []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}

// updateFileSystemResourceGroup handles resource group changes
func updateFileSystemResourceGroup(d *schema.ResourceData, client *connectivity.AliyunClient) error {
	if !d.HasChange("resource_group_id") {
		return nil
	}

	request := map[string]interface{}{
		"ResourceId":         d.Id(),
		"RegionId":           client.RegionId,
		"NewResourceGroupId": d.Get("resource_group_id"),
		"ResourceType":       "filesystem",
	}

	action := "ChangeResourceGroup"
	var response map[string]interface{}
	var err error

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		response, err = client.RpcPost("NAS", "2017-06-26", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
	}

	return nil
}

// updateFileSystemRecycleBin handles recycle bin configuration for standard file systems
func updateFileSystemRecycleBin(d *schema.ResourceData, nasService *NasService) error {
	// Only standard file systems support recycle bin
	if d.Get("file_system_type").(string) != "standard" {
		return nil
	}

	if d.HasChange("recycle_bin.0.status") {
		status := d.Get("recycle_bin.0.status").(string)
		if status == "Enable" {
			request := map[string]interface{}{
				"FileSystemId": d.Id(),
			}
			if v, ok := d.GetOkExists("recycle_bin.0.reserved_days"); ok {
				request["ReservedDays"] = v
			}
			if err := nasService.EnableRecycleBin(request); err != nil {
				return err
			}
		} else if status == "Disable" {
			if err := nasService.DisableAndCleanRecycleBin(d.Id()); err != nil {
				return err
			}
		}
	}

	if d.HasChange("recycle_bin.0.reserved_days") {
		status := d.Get("recycle_bin.0.status").(string)
		if status == "Enable" {
			request := map[string]interface{}{
				"FileSystemId": d.Id(),
				"ReservedDays": d.Get("recycle_bin.0.reserved_days"),
			}
			if err := nasService.UpdateRecycleBinAttribute(request); err != nil {
				return err
			}
		}
	}

	return nil
}

// updateFileSystemNfsAcl handles NFS ACL configuration for standard + NFS file systems
func updateFileSystemNfsAcl(d *schema.ResourceData, nasService *NasService) error {
	// Only standard + NFS file systems support NFS ACL
	if d.Get("file_system_type").(string) != "standard" || d.Get("protocol_type").(string) != "NFS" {
		return nil
	}

	if d.HasChange("nfs_acl.0.enabled") {
		enabled := d.Get("nfs_acl.0.enabled").(bool)
		if enabled {
			if err := nasService.EnableNfsAcl(d.Id()); err != nil {
				return err
			}
		} else {
			if err := nasService.DisableNfsAcl(d.Id()); err != nil {
				return err
			}
		}
	}

	return nil
}

// updateFileSystemSmbAcl handles SMB ACL configuration for standard + SMB file systems
func updateFileSystemSmbAcl(d *schema.ResourceData, nasService *NasService) error {
	// Only standard + SMB file systems support SMB ACL
	if d.Get("file_system_type").(string) != "standard" || d.Get("protocol_type").(string) != "SMB" {
		return nil
	}

	if d.HasChange("smb_acl.0.enabled") {
		enabled := d.Get("smb_acl.0.enabled").(bool)
		if enabled {
			request := map[string]interface{}{
				"FileSystemId": d.Id(),
				"Keytab":       d.Get("keytab"),
				"KeytabMd5":    d.Get("keytab_md5"),
			}
			if err := nasService.EnableSmbAcl(request); err != nil {
				return err
			}
		} else {
			if err := nasService.DisableSmbAcl(d.Id()); err != nil {
				return err
			}
		}
	}

	// Update SMB ACL properties if ACL is enabled
	if d.Get("smb_acl.0.enabled").(bool) {
		if d.HasChanges("smb_acl.0.super_admin_sid", "smb_acl.0.home_dir_path",
			"smb_acl.0.enable_anonymous_access", "smb_acl.0.encrypt_data",
			"smb_acl.0.reject_unencrypted_access") {

			request := map[string]interface{}{
				"FileSystemId": d.Id(),
				"Keytab":       d.Get("keytab"),
				"KeytabMd5":    d.Get("keytab_md5"),
			}

			if v, err := jsonpath.Get("$[0].super_admin_sid", d.Get("smb_acl")); err == nil && v != nil {
				request["SuperAdminSid"] = v
			}
			if v, err := jsonpath.Get("$[0].home_dir_path", d.Get("smb_acl")); err == nil && v != nil {
				request["HomeDirPath"] = v
			}
			if v, err := jsonpath.Get("$[0].enable_anonymous_access", d.Get("smb_acl")); err == nil && v != nil {
				request["EnableAnonymousAccess"] = v
			}
			if v, err := jsonpath.Get("$[0].encrypt_data", d.Get("smb_acl")); err == nil && v != nil {
				request["EncryptData"] = v
			}
			if v, err := jsonpath.Get("$[0].reject_unencrypted_access", d.Get("smb_acl")); err == nil && v != nil {
				request["RejectUnencryptedAccess"] = v
			}

			if err := nasService.ModifySmbAcl(request); err != nil {
				return err
			}
		}
	}

	return nil
}
func resourceAliCloudNasFileSystemDelete(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*connectivity.AliyunClient)
	action := "DeleteFileSystem"
	var request map[string]interface{}
	var response map[string]interface{}
	query := make(map[string]interface{})
	var err error
	request = make(map[string]interface{})
	request["FileSystemId"] = d.Id()

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		response, err = client.RpcPost("NAS", "2017-06-26", action, query, request, true)

		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	addDebug(action, response, request)

	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidFileSystem.NotFound"}) || NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
	}

	nasService, err := NewNasService(client)
	if err != nil {
		return WrapError(err)
	}

	stateConf := BuildStateConf([]string{}, []string{""}, d.Timeout(schema.TimeoutDelete), 5*time.Second, nasService.DescribeAsyncNasFileSystemStateRefreshFunc(d, response, "$.FileSystems", []string{}))
	if jobDetail, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id(), jobDetail)
	}

	return nil
}
