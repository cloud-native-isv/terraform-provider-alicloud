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

func resourceAliCloudNasMountTarget() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudNasMountTargetCreate,
		Read:   resourceAliCloudNasMountTargetRead,
		Update: resourceAliCloudNasMountTargetUpdate,
		Delete: resourceAliCloudNasMountTargetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(40 * time.Minute),
			Update: schema.DefaultTimeout(40 * time.Minute),
			Delete: schema.DefaultTimeout(40 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"access_group_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dual_stack": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"file_system_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mount_target_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"vswitch_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAliCloudNasMountTargetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapError(err)
	}

	// Get NAS API client
	nasAPI, err := nasService.getNasAPI()
	if err != nil {
		return WrapError(fmt.Errorf("failed to get NAS API client: %w", err))
	}

	// Prepare mount target object
	mountTarget := &aliyunNasAPI.MountTarget{
		NetworkType:     d.Get("network_type").(string),
		AccessGroupName: d.Get("access_group_name").(string),
	}

	if mountTarget.NetworkType == "" {
		mountTarget.NetworkType = string(Vpc)
	}

	if v, ok := d.GetOk("vpc_id"); ok {
		mountTarget.VpcId = v.(string)
	}

	if v, ok := d.GetOk("vswitch_id"); ok {
		mountTarget.VSwitchId = v.(string)
	}

	// If vpc_id is not provided but vswitch_id is, get vpc_id from vswitch
	if mountTarget.VpcId == "" && mountTarget.VSwitchId != "" {
		vpcService := VpcService{client}
		vsw, err := vpcService.DescribeVSwitchWithTeadsl(mountTarget.VSwitchId)
		if err != nil {
			return WrapError(err)
		}
		mountTarget.VpcId = vsw["VpcId"].(string)
	}

	fileSystemId := d.Get("file_system_id").(string)

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		createdMountTarget, createErr := nasAPI.CreateMountTarget(fileSystemId, mountTarget)
		if createErr != nil {
			if NeedRetry(createErr) || IsExpectedErrors(createErr, []string{"OperationDenied.InvalidState"}) {
				wait()
				return resource.RetryableError(createErr)
			}
			return resource.NonRetryableError(createErr)
		}
		mountTarget = createdMountTarget
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_mount_target", "CreateMountTarget", AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprintf("%s:%s", fileSystemId, mountTarget.MountTargetDomain))

	stateConf := BuildStateConf([]string{}, []string{"Active"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, nasService.NasMountTargetStateRefreshFunc(d.Id(), "Status", []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudNasMountTargetUpdate(d, meta)
}

func resourceAliCloudNasMountTargetRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapError(err)
	}

	mountTarget, err := nasService.DescribeNasMountTarget(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_nas_mount_target DescribeNasMountTarget Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("access_group_name", mountTarget.AccessGroupName)
	d.Set("network_type", mountTarget.NetworkType)
	d.Set("status", mountTarget.Status)
	d.Set("vswitch_id", mountTarget.VSwitchId)
	d.Set("vpc_id", mountTarget.VpcId)
	d.Set("mount_target_domain", mountTarget.MountTargetDomain)

	parts := strings.Split(d.Id(), ":")
	d.Set("file_system_id", parts[0])

	return nil
}

func resourceAliCloudNasMountTargetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapError(err)
	}

	update := false
	parts := strings.Split(d.Id(), ":")
	fileSystemId := parts[0]
	mountTargetDomain := parts[1]

	var accessGroupName, status string

	if !d.IsNewResource() && d.HasChange("access_group_name") {
		update = true
	}
	accessGroupName = d.Get("access_group_name").(string)

	if d.HasChange("status") {
		update = true
		status = d.Get("status").(string)
	}

	if update {
		// Get NAS API client
		nasAPI, err := nasService.getNasAPI()
		if err != nil {
			return WrapError(fmt.Errorf("failed to get NAS API client: %w", err))
		}

		wait := incrementalWait(3*time.Second, 5*time.Second)
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			modifyErr := nasAPI.ModifyMountTarget(fileSystemId, mountTargetDomain, accessGroupName, status)
			if modifyErr != nil {
				if NeedRetry(modifyErr) || IsExpectedErrors(modifyErr, []string{"OperationDenied.InvalidState"}) {
					wait()
					return resource.RetryableError(modifyErr)
				}
				return resource.NonRetryableError(modifyErr)
			}
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifyMountTarget", AlibabaCloudSdkGoERROR)
		}

		// Wait for update to complete
		stateConf := BuildStateConf([]string{}, []string{"Active"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, nasService.NasMountTargetStateRefreshFunc(d.Id(), "Status", []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	return resourceAliCloudNasMountTargetRead(d, meta)
}

func resourceAliCloudNasMountTargetDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapError(err)
	}

	parts := strings.Split(d.Id(), ":")

	fileSystemId := parts[0]
	mountTargetDomain := parts[1]

	// Get NAS API client
	nasAPI, err := nasService.getNasAPI()
	if err != nil {
		return WrapError(fmt.Errorf("failed to get NAS API client: %w", err))
	}

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		deleteErr := nasAPI.DeleteMountTarget(fileSystemId, mountTargetDomain)
		if deleteErr != nil {
			if IsExpectedErrors(deleteErr, []string{"Forbidden.NasNotFound"}) {
				return nil
			}
			if NeedRetry(deleteErr) || IsExpectedErrors(deleteErr, []string{"VolumeStatusForbidOperation", "OperationDenied.InvalidState"}) {
				wait()
				return resource.RetryableError(deleteErr)
			}
			return resource.NonRetryableError(deleteErr)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteMountTarget", AlibabaCloudSdkGoERROR)
	}

	// Wait for deletion to complete
	stateConf := BuildStateConf([]string{"Active", "Inactive"}, []string{""}, d.Timeout(schema.TimeoutDelete), 5*time.Second, nasService.NasMountTargetStateRefreshFunc(d.Id(), "Status", []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
