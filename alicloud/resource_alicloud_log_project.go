// Package alicloud. This file is generated automatically. Please do not modify it manually, thank you!
package alicloud

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func isProjectTransferAccelerationNotSupportedError(err error) bool {
	return IsExpectedErrors(err, []string{"NotSupported", "The operation is not supported in this region."})
}

func setProjectTransferAccelerationEnabled(slsService *SlsService, projectName string, currentEnabled bool, targetEnabled bool) error {
	if currentEnabled == targetEnabled {
		log.Printf("[DEBUG] Project %s transfer acceleration is already %t, skip setting it", projectName, targetEnabled)
		return nil
	}

	if targetEnabled {
		if err := slsService.EnableProjectTransferAcceleration(projectName); err != nil {
			return err
		}
		return nil
	}

	if err := slsService.DisableProjectTransferAcceleration(projectName); err != nil {
		if isProjectTransferAccelerationNotSupportedError(err) {
			log.Printf("[WARN] Project %s DisableProjectTransferAcceleration returned NotSupported. Treating transfer_acceleration_enabled=false as applied: %s", projectName, err)
			return nil
		}
		return err
	}

	return nil
}

func resourceAliCloudLogProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudLogProjectCreate,
		Read:   resourceAliCloudLogProjectRead,
		Update: resourceAliCloudLogProjectUpdate,
		Delete: resourceAliCloudLogProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"project_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"project_name", "name"},
				ForceNew:     true,
				ValidateFunc: StringMatch(regexp.MustCompile("^[0-9a-zA-Z_-]+$"), "The name of the log project. It is the only in one AliCloud account. The project name is globally unique in Alibaba Cloud and cannot be modified after it is created. The naming rules are as follows:- The project name must be globally unique. - The name can contain only lowercase letters, digits, and hyphens (-). - It must start and end with a lowercase letter or number. - The value contains 3 to 63 characters."),
			},
			"name": {
				Type:       schema.TypeString,
				Optional:   true,
				Computed:   true,
				Deprecated: "Field 'name' has been deprecated since provider version 1.215.0. New field 'project_name' instead.",
				ForceNew:   true,
			},
			"resource_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modify_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_redundancy_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  string(aliyunSlsAPI.DataRedundancyTypeZRS),
			},
			"recycle_bin_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"transfer_acceleration_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"quota": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"policy": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAliCloudLogProjectCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_project", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	// Get project name from either project_name or name field
	var projectName string
	if v, ok := d.GetOk("project_name"); ok {
		projectName = v.(string)
	} else if v, ok := d.GetOk("name"); ok {
		projectName = v.(string)
	} else {
		return WrapError(fmt.Errorf("either project_name or name must be specified"))
	}

	// Build LogProject struct for creation
	logProject := &aliyunSlsAPI.LogProject{
		ProjectName: projectName,
	}

	// Set optional fields if provided
	if v, ok := d.GetOk("description"); ok {
		logProject.Description = v.(string)
	}
	if v, ok := d.GetOk("resource_group_id"); ok {
		logProject.ResourceGroupId = v.(string)
	}

	// Set data redundancy type if provided. Validation is handled by the Terraform module layer.
	if v, ok := d.GetOk("data_redundancy_type"); ok {
		dataRedundancyType := v.(string)
		logProject.DataRedundancyType = aliyunSlsAPI.DataRedundancyType(dataRedundancyType)
	}

	if v, ok := d.GetOk("recycle_bin_enabled"); ok {
		logProject.RecycleBinEnabled = v.(bool)
	}
	if v, ok := d.GetOk("quota"); ok {
		quotaMap := v.(map[string]interface{})
		if len(quotaMap) > 0 {
			logProject.Quota = quotaMap
		}
	}

	// Create project using SlsService
	err = slsService.CreateProject(logProject)
	if err != nil {
		// Check if the error is due to project already existing
		if IsAlreadyExistError(err) {
			log.Printf("[INFO] Project %s already exists, importing existing resource into Terraform state", projectName)
			d.SetId(projectName)
			return resourceAliCloudLogProjectRead(d, meta)
		}

		// For any other error, return the original error
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_project", "CreateProject", AlibabaCloudSdkGoERROR)
	}

	// Set the resource ID
	d.SetId(projectName)

	if err := setProjectTransferAccelerationEnabled(slsService, projectName, false, d.Get("transfer_acceleration_enabled").(bool)); err != nil {
		return err
	}

	// For newly created projects, we can directly call Read since SLS projects are immediately available
	return resourceAliCloudLogProjectRead(d, meta)
}

func resourceAliCloudLogProjectRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_project", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	// Get project details
	project, err := slsService.DescribeLogProject(d.Id())
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	log.Printf("[INFO] read Project %s, save resource into Terraform state", d.Id())

	// Set basic project attributes
	d.Set("create_time", project.CreateTime)
	d.Set("description", project.Description)
	d.Set("resource_group_id", project.ResourceGroupId)
	d.Set("status", project.Status)
	d.Set("project_name", project.ProjectName)
	d.Set("name", project.ProjectName)
	d.Set("owner", project.Owner)
	d.Set("region", project.Region)
	d.Set("location", project.Location)
	d.Set("last_modify_time", project.LastModifyTime)
	d.Set("data_redundancy_type", project.DataRedundancyType)
	d.Set("recycle_bin_enabled", project.RecycleBinEnabled)
	d.Set("transfer_acceleration_enabled", d.Get("transfer_acceleration_enabled"))
	d.Set("quota", project.Quota)

	return nil
}

func resourceAliCloudLogProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_project", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	d.Partial(true)

	// Update project description
	if d.HasChange("description") {
		updateRequest := map[string]interface{}{
			"projectName": d.Id(),
			"description": d.Get("description"),
		}

		err := slsService.UpdateProject(updateRequest)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateProject", AlibabaCloudSdkGoERROR)
		}
		d.SetPartial("description")
	}

	// Update recycle bin setting
	if d.HasChange("recycle_bin_enabled") {
		updateRequest := map[string]interface{}{
			"projectName":       d.Id(),
			"recycleBinEnabled": d.Get("recycle_bin_enabled"),
		}

		err := slsService.UpdateProject(updateRequest)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateProject", AlibabaCloudSdkGoERROR)
		}
		d.SetPartial("recycle_bin_enabled")
	}

	// Update resource group
	if d.HasChange("resource_group_id") {
		resourceGroupRequest := map[string]interface{}{
			"resourceId":      d.Id(),
			"resourceGroupId": d.Get("resource_group_id"),
			"resourceType":    "PROJECT",
		}

		err := slsService.ChangeResourceGroup(resourceGroupRequest)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ChangeResourceGroup", AlibabaCloudSdkGoERROR)
		}
		d.SetPartial("resource_group_id")
	}

	// Update project policy
	if d.HasChange("policy") {
		policy := ""
		if v, ok := d.GetOk("policy"); ok {
			policy = v.(string)
		}

		err := slsService.UpdateProjectPolicy(d.Id(), policy)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateProjectPolicy", AlibabaCloudSdkGoERROR)
		}
		d.SetPartial("policy")
	}

	if d.HasChange("transfer_acceleration_enabled") {
		oldValue, newValue := d.GetChange("transfer_acceleration_enabled")
		if err := setProjectTransferAccelerationEnabled(slsService, d.Id(), oldValue.(bool), newValue.(bool)); err != nil {
			return err
		}
		d.SetPartial("transfer_acceleration_enabled")
	}

	d.Partial(false)
	return resourceAliCloudLogProjectRead(d, meta)
}

func resourceAliCloudLogProjectDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_project", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	err = slsService.DeleteProject(d.Id())
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteProject", AlibabaCloudSdkGoERROR)
	}

	// Use StateRefreshFunc to wait for project deletion completion
	stateConf := BuildStateConf([]string{"Normal"}, []string{}, d.Timeout(schema.TimeoutDelete), 5*time.Second, slsService.LogProjectStateRefreshFunc(d.Id(), "status", []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
