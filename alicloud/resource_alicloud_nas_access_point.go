// Package alicloud. This file is generated automatically. Please do not modify it manually, thank you!
package alicloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunNasAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudNasAccessPoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudNasAccessPointCreate,
		Read:   resourceAliCloudNasAccessPointRead,
		Update: resourceAliCloudNasAccessPointUpdate,
		Delete: resourceAliCloudNasAccessPointDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"access_group": {
				Type:     schema.TypeString,
				Required: true,
			},
			"access_point_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"access_point_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled_ram": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"file_system_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"posix_user": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"posix_group_id": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"posix_user_id": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"posix_secondary_group_ids": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeInt},
						},
					},
				},
			},
			"root_path": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"root_path_permission": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"owner_user_id": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"permission": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"owner_group_id": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vswitch_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAliCloudNasAccessPointCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_access_point", "NewNasService", AlibabaCloudSdkGoERROR)
	}

	// Prepare parameters for the service call
	fileSystemId := d.Get("file_system_id").(string)
	accessPointName := ""
	if v, ok := d.GetOk("access_point_name"); ok {
		accessPointName = v.(string)
	}

	accessGroup := d.Get("access_group").(string)
	rootPath := "/"
	if v, ok := d.GetOk("root_path"); ok {
		rootPath = v.(string)
	}

	enabledRam := false
	if v, ok := d.GetOkExists("enabled_ram"); ok {
		enabledRam = v.(bool)
	}

	vpcId := d.Get("vpc_id").(string)
	vSwitchId := d.Get("vswitch_id").(string)

	// Default owner IDs
	var ownerUid, ownerGid int64 = 1, 1
	permission := "0777"

	// Get root path permission if specified
	if v, ok := d.GetOk("root_path_permission"); ok {
		rootPathPermissions := v.([]interface{})
		if len(rootPathPermissions) > 0 {
			rootPathPermission := rootPathPermissions[0].(map[string]interface{})
			if ownerUserId, ok := rootPathPermission["owner_user_id"]; ok && ownerUserId != "" {
				ownerUid = int64(ownerUserId.(int))
			}
			if ownerGroupId, ok := rootPathPermission["owner_group_id"]; ok && ownerGroupId != "" {
				ownerGid = int64(ownerGroupId.(int))
			}
			if perm, ok := rootPathPermission["permission"]; ok && perm != "" {
				permission = perm.(string)
			}
		}
	}

	// Get POSIX user if specified
	var posixUser *aliyunNasAPI.PosixUser
	if v, ok := d.GetOk("posix_user"); ok {
		posixUsers := v.([]interface{})
		if len(posixUsers) > 0 {
			posixUserMap := posixUsers[0].(map[string]interface{})
			posixUser = &aliyunNasAPI.PosixUser{}

			if posixUserId, ok := posixUserMap["posix_user_id"]; ok && posixUserId != "" {
				posixUser.Uid = int64(posixUserId.(int))
			}
			if posixGroupId, ok := posixUserMap["posix_group_id"]; ok && posixGroupId != "" {
				posixUser.Gid = int64(posixGroupId.(int))
			}
		}
	}

	// Create access point using service layer
	accessPoint, err := nasService.CreateNasAccessPoint(
		fileSystemId,
		accessPointName,
		accessGroup,
		rootPath,
		enabledRam,
		vpcId,
		vSwitchId,
		ownerUid,
		ownerGid,
		permission,
		posixUser,
	)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_access_point", "CreateAccessPoint", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID
	d.SetId(fmt.Sprint(fileSystemId, ":", accessPoint.AccessPointId))

	// Wait for access point to become active using StateRefreshFunc
	stateConf := BuildStateConf([]string{"creating", "pending"}, []string{"active"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, nasService.NasAccessPointStateRefreshFunc(d.Id(), []string{"failed", "error"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// Finally call Read to sync complete state
	return resourceAliCloudNasAccessPointRead(d, meta)
}

func resourceAliCloudNasAccessPointRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "NewNasService", AlibabaCloudSdkGoERROR)
	}

	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return WrapErrorf(fmt.Errorf("invalid resource ID format"), DefaultErrorMsg, d.Id(), "ParseResourceId", AlibabaCloudSdkGoERROR)
	}
	fileSystemId := parts[0]
	accessPointId := parts[1]

	// Use service layer to get access point details
	accessPoint, err := nasService.DescribeNasAccessPoint(fileSystemId, accessPointId)
	if err != nil {
		if !d.IsNewResource() && IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_nas_access_point DescribeNasAccessPoint Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set attributes from the strongly-typed response
	d.Set("access_group", accessPoint.AccessGroup)
	d.Set("access_point_name", accessPoint.AccessPointName)
	d.Set("create_time", accessPoint.CreateTime)
	d.Set("enabled_ram", accessPoint.EnabledRam)
	d.Set("root_path", accessPoint.RootPath)
	d.Set("status", accessPoint.Status)
	d.Set("vswitch_id", accessPoint.VSwitchId)
	d.Set("vpc_id", accessPoint.VpcId)
	d.Set("access_point_id", accessPoint.AccessPointId)
	d.Set("file_system_id", accessPoint.FileSystemId)

	// Handle PosixUser
	posixUserMaps := make([]map[string]interface{}, 0)
	if accessPoint.PosixUser != nil {
		posixUserMap := map[string]interface{}{
			"posix_group_id":            int(accessPoint.PosixUser.Gid),
			"posix_user_id":             int(accessPoint.PosixUser.Uid),
			"posix_secondary_group_ids": convertInt64SliceToIntSlice(accessPoint.PosixUser.SecondaryGids),
		}
		posixUserMaps = append(posixUserMaps, posixUserMap)
	}
	d.Set("posix_user", posixUserMaps)

	// Handle RootPathPermission
	rootPathPermissionMaps := make([]map[string]interface{}, 0)
	if accessPoint.RootPathPermission != nil {
		rootPathPermissionMap := map[string]interface{}{
			"owner_group_id": int(accessPoint.RootPathPermission.OwnerGid),
			"owner_user_id":  int(accessPoint.RootPathPermission.OwnerUid),
			"permission":     accessPoint.RootPathPermission.Permission,
		}
		rootPathPermissionMaps = append(rootPathPermissionMaps, rootPathPermissionMap)
	}
	d.Set("root_path_permission", rootPathPermissionMaps)

	return nil
}

// Helper function to convert []int64 to []int for Terraform schema compatibility
func convertInt64SliceToIntSlice(slice []int64) []int {
	result := make([]int, len(slice))
	for i, v := range slice {
		result[i] = int(v)
	}
	return result
}

func resourceAliCloudNasAccessPointUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "NewNasService", AlibabaCloudSdkGoERROR)
	}

	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return WrapErrorf(fmt.Errorf("invalid resource ID format"), DefaultErrorMsg, d.Id(), "ParseResourceId", AlibabaCloudSdkGoERROR)
	}
	fileSystemId := parts[0]
	accessPointId := parts[1]

	update := false

	// Check if any field has changed
	if d.HasChange("access_point_name") || d.HasChange("access_group") || d.HasChange("enabled_ram") {
		update = true
	}

	if update {
		// Prepare parameters for the service call
		accessPointName := ""
		if v, ok := d.GetOk("access_point_name"); ok {
			accessPointName = v.(string)
		}

		accessGroup := d.Get("access_group").(string)

		enabledRam := false
		if v, ok := d.GetOkExists("enabled_ram"); ok {
			enabledRam = v.(bool)
		}

		// Use service layer method
		err = nasService.UpdateNasAccessPoint(fileSystemId, accessPointId, accessPointName, accessGroup, enabledRam)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifyAccessPoint", AlibabaCloudSdkGoERROR)
		}
	}

	return resourceAliCloudNasAccessPointRead(d, meta)
}

func resourceAliCloudNasAccessPointDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "NewNasService", AlibabaCloudSdkGoERROR)
	}

	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return WrapErrorf(fmt.Errorf("invalid resource ID format"), DefaultErrorMsg, d.Id(), "ParseResourceId", AlibabaCloudSdkGoERROR)
	}
	fileSystemId := parts[0]
	accessPointId := parts[1]

	// Use service layer method with retry logic
	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := nasService.DeleteNasAccessPoint(fileSystemId, accessPointId)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteAccessPoint", AlibabaCloudSdkGoERROR)
	}

	// Wait for deletion to complete
	stateConf := BuildStateConf([]string{}, []string{}, d.Timeout(schema.TimeoutDelete), 5*time.Second, nasService.NasAccessPointStateRefreshFunc(d.Id(), []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
