package alicloud

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAlicloudNasFileset() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudNasFilesetCreate,
		Read:   resourceAlicloudNasFilesetRead,
		Update: resourceAlicloudNasFilesetUpdate,
		Delete: resourceAlicloudNasFilesetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.All(validation.StringLenBetween(2, 128), validation.StringDoesNotMatch(regexp.MustCompile(`(^http://.*)|(^https://.*)`), "It cannot begin with \"http://\", \"https://\".")),
			},
			"dry_run": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"file_system_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"file_system_path": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"fileset_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAlicloudNasFilesetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Create credentials for CWS-Lib-Go API
	credentials := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	// Create NAS API client
	nasAPI, err := nas.NewNasAPI(credentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_fileset", "NewNasAPI", AlibabaCloudSdkGoERROR)
	}

	// Create fileset using strong-typed API with individual parameters
	fileset, err := nasAPI.CreateFileset(
		d.Get("file_system_id").(string),
		d.Get("file_system_path").(string),
		func() string {
			if v, ok := d.GetOk("description"); ok {
				return v.(string)
			}
			return ""
		}(),
		false, // deletionProtection - default to false
	)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_fileset", "CreateFileset", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID in format: fileSystemId:filesetId
	d.SetId(fmt.Sprint(d.Get("file_system_id").(string), ":", fileset.FsetId))

	// Wait for fileset to be created using StateRefreshFunc
	stateConf := BuildStateConf([]string{}, []string{"CREATED"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, func() (interface{}, string, error) {
		parts, err := ParseResourceId(d.Id(), 2)
		if err != nil {
			return nil, "", WrapError(err)
		}

		fileSystemId := parts[0]
		fsetId := parts[1]

		object, err := nasAPI.GetFileset(fileSystemId, fsetId)
		if err != nil {
			if IsExpectedErrors(err, []string{"InvalidFileset.NotFound", "Forbidden.FilesetNotFound"}) {
				return nil, "", WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
			}
			return nil, "", WrapError(err)
		}

		return object, object.Status, nil
	})

	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAlicloudNasFilesetRead(d, meta)
}
func resourceAlicloudNasFilesetRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Create credentials for CWS-Lib-Go API
	credentials := &common.Credentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	// Create NAS API client
	nasAPI, err := nas.NewNasAPI(credentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_fileset", "NewNasAPI", AlibabaCloudSdkGoERROR)
	}

	// Parse resource ID to extract file system ID and fileset ID
	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	fileSystemId := parts[0]
	fsetId := parts[1]

	// Get fileset using strong-typed API
	object, err := nasAPI.GetFileset(fileSystemId, fsetId)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidFileset.NotFound", "Forbidden.FilesetNotFound"}) {
			log.Printf("[DEBUG] Resource alicloud_nas_fileset not found, removing from state: %s", err)
			d.SetId("")
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "GetFileset", AlibabaCloudSdkGoERROR)
	}

	// Set resource attributes using strong types
	d.Set("file_system_id", object.FileSystemId)
	d.Set("fileset_id", object.FsetId)
	d.Set("description", object.Description)
	d.Set("file_system_path", object.FileSystemPath)
	d.Set("status", object.Status)

	return nil
}
func resourceAlicloudNasFilesetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var response map[string]interface{}
	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}
	request := map[string]interface{}{
		"FileSystemId": parts[0],
		"FsetId":       parts[1],
	}
	if d.HasChange("description") {
		if v, ok := d.GetOk("description"); ok {
			request["Description"] = v
		}
	}
	if v, ok := d.GetOkExists("dry_run"); ok {
		request["DryRun"] = v
	}
	action := "ModifyFileset"
	request["ClientToken"] = buildClientToken("ModifyFileset")
	wait := incrementalWait(3*time.Second, 3*time.Second)
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
	return resourceAlicloudNasFilesetRead(d, meta)
}
func resourceAlicloudNasFilesetDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}
	action := "DeleteFileset"
	var response map[string]interface{}
	request := map[string]interface{}{
		"FileSystemId": parts[0],
		"FsetId":       parts[1],
	}

	if v, ok := d.GetOkExists("dry_run"); ok {
		request["DryRun"] = v
	}
	request["ClientToken"] = buildClientToken("DeleteFileset")
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
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
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapError(err)
	}

	stateConf := BuildStateConf([]string{}, []string{}, d.Timeout(schema.TimeoutCreate), 5*time.Second, nasService.NasFilesetStateRefreshFunc(d.Id(), []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
