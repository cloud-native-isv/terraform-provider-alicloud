package alicloud

import (
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

	// Use strong type from NAS service
	fileSystem, err := nasService.DescribeNasFileSystem(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_nas_file_system DescribeNasFileSystem Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set all basic attributes using strong types
	d.Set("capacity", fileSystem.Capacity)
	d.Set("create_time", fileSystem.CreateTime)
	d.Set("description", fileSystem.Description)
	d.Set("encrypt_type", fileSystem.EncryptType)
	d.Set("file_system_type", fileSystem.FileSystemType)
	d.Set("kms_key_id", fileSystem.KMSKeyId)
	d.Set("protocol_type", fileSystem.ProtocolType)
	d.Set("region_id", fileSystem.RegionId)
	d.Set("resource_group_id", fileSystem.ResourceGroupId)
	d.Set("status", fileSystem.Status)
	d.Set("storage_type", fileSystem.StorageType)
	d.Set("zone_id", fileSystem.ZoneId)

	// Always set vpc_id and vswitch_id, even if empty, to maintain consistency
	if fileSystem.VpcId != "" {
		d.Set("vpc_id", fileSystem.VpcId)
	} else {
		d.Set("vpc_id", "")
	}

	if fileSystem.VSwitchId != "" {
		d.Set("vswitch_id", fileSystem.VSwitchId)
	} else {
		d.Set("vswitch_id", "")
	}

	// Handle options block - Note: Options is not available in the strong type yet
	// Keep options empty for now until CWS-Lib-Go adds support
	optionsMaps := make([]map[string]interface{}, 0)
	if err := d.Set("options", optionsMaps); err != nil {
		return err
	}

	// Handle conditional SMB ACL for standard + SMB
	if fileSystem.FileSystemType == "standard" && fileSystem.ProtocolType == "SMB" {
		objectRaw, err := nasService.DescribeFileSystemDescribeSmbAcl(d.Id())
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
	if fileSystem.FileSystemType == "standard" {
		objectRaw, err := nasService.DescribeFileSystemGetRecycleBinAttribute(d.Id())
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

	// Handle tags using strong types from file system
	tagsMaps := make(map[string]interface{})
	if fileSystem.Tags != nil {
		for _, tag := range fileSystem.Tags {
			if tag.Key != "" {
				tagsMaps[tag.Key] = tag.Value
			}
		}
	}
	d.Set("tags", tagsMaps)

	// Handle conditional NFS ACL for standard + NFS
	if fileSystem.FileSystemType == "standard" && fileSystem.ProtocolType == "NFS" {
		objectRaw, err := nasService.DescribeFileSystemDescribeNfsAcl(d.Id())
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

	if d.HasChange("description") {
		description := d.Get("description").(string)
		return nasService.ModifyFileSystem(d.Id(), description)
	}

	// Note: Options is not yet supported in CWS-Lib-Go, so we skip options updates for now
	// TODO: Add options support when CWS-Lib-Go implements it

	return nil
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

	capacity := int64(d.Get("capacity").(int))
	if err := nasService.UpgradeFileSystem(d.Id(), capacity); err != nil {
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

	// For now, use RPC call directly since ChangeResourceGroup is not available in CWS-Lib-Go yet
	// TODO: Replace with CWS-Lib-Go method when available
	action := "ModifyFileSystem"
	request := map[string]interface{}{
		"FileSystemId":    d.Id(),
		"ResourceGroupId": d.Get("resource_group_id").(string),
	}

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		response, err := client.RpcPost("NAS", "2017-06-26", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})

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
			reservedDays := int64(3) // Default value
			if v, ok := d.GetOkExists("recycle_bin.0.reserved_days"); ok {
				reservedDays = int64(v.(int))
			}
			if err := nasService.EnableRecycleBin(d.Id(), reservedDays); err != nil {
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
			reservedDays := int64(d.Get("recycle_bin.0.reserved_days").(int))
			// Note: UpdateRecycleBinAttribute signature expects (fileSystemId, reservedDays, status)
			// but implementation only uses fileSystemId and reservedDays
			if err := nasService.UpdateRecycleBinAttribute(d.Id(), reservedDays, status); err != nil {
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
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapError(err)
	}

	// Delete file system using NAS service
	err = nasService.DeleteNasFileSystem(d.Id())
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidFileSystem.NotFound"}) || NotFoundError(err) {
			return nil
		}
		return WrapError(err)
	}

	// Wait for deletion completion
	stateConf := BuildStateConf([]string{"Running", "Stopped"}, []string{""}, d.Timeout(schema.TimeoutDelete), 5*time.Second, nasService.DescribeNasFileSystemStateRefreshFunc(d.Id(), "", []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
