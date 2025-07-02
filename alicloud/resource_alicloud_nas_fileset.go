package alicloud

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudNasFileset() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudNasFilesetCreate,
		Read:   resourceAliCloudNasFilesetRead,
		Update: resourceAliCloudNasFilesetUpdate,
		Delete: resourceAliCloudNasFilesetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.All(validation.StringLenBetween(2, 128), validation.StringDoesNotMatch(regexp.MustCompile(`(^http://.*)|(^https://.*)`), "It cannot begin with \"http://\", \"https://\".")),
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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

func resourceAliCloudNasFilesetCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapError(err)
	}

	fileSystemId := d.Get("file_system_id").(string)
	fileSystemPath := d.Get("file_system_path").(string)
	description := ""
	if v, ok := d.GetOk("description"); ok {
		description = v.(string)
	}
	deletionProtection := d.Get("deletion_protection").(bool)

	// Create fileset using service layer
	fileset, err := nasService.CreateNasFileset(fileSystemId, fileSystemPath, description, deletionProtection)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_nas_fileset", "CreateNasFileset", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID in format: fileSystemId:filesetId
	d.SetId(fmt.Sprintf("%s:%s", fileSystemId, fileset.FsetId))

	// Wait for fileset to be created
	err = nasService.WaitForNasFileset(d.Id(), "Running", int(d.Timeout(schema.TimeoutCreate).Seconds()))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudNasFilesetRead(d, meta)
}

func resourceAliCloudNasFilesetRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapError(err)
	}

	// Get fileset using service layer
	fileset, err := nasService.DescribeNasFileset(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_nas_fileset DescribeNasFileset Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set resource attributes using strong types
	d.Set("file_system_id", fileset.FileSystemId)
	d.Set("fileset_id", fileset.FsetId)
	d.Set("description", fileset.Description)
	d.Set("file_system_path", fileset.FileSystemPath)
	d.Set("status", fileset.Status)
	d.Set("deletion_protection", fileset.DeletionProtection)

	return nil
}

func resourceAliCloudNasFilesetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	fileSystemId := parts[0]
	fsetId := parts[1]

	update := false
	description := ""
	deletionProtection := false

	if d.HasChange("description") {
		update = true
		if v, ok := d.GetOk("description"); ok {
			description = v.(string)
		}
	}

	if d.HasChange("deletion_protection") {
		update = true
		deletionProtection = d.Get("deletion_protection").(bool)
	}

	if update {
		// Update fileset using service layer
		err := nasService.UpdateNasFileset(fileSystemId, fsetId, description, deletionProtection)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateNasFileset", AlibabaCloudSdkGoERROR)
		}

		// Wait for update to complete
		err = nasService.WaitForNasFileset(d.Id(), "Running", int(d.Timeout(schema.TimeoutUpdate).Seconds()))
		if err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	return resourceAliCloudNasFilesetRead(d, meta)
}

func resourceAliCloudNasFilesetDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	fileSystemId := parts[0]
	fsetId := parts[1]

	// Delete fileset using service layer
	err = nasService.DeleteNasFileset(fileSystemId, fsetId)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteNasFileset", AlibabaCloudSdkGoERROR)
	}

	// Wait for fileset to be deleted
	err = nasService.WaitForNasFilesetDeleted(d.Id(), int(d.Timeout(schema.TimeoutDelete).Seconds()))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
