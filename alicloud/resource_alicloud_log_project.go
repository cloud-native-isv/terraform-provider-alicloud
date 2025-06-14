// Package alicloud. This file is generated automatically. Please do not modify it manually, thank you!
package alicloud

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudSlsProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudSlsProjectCreate,
		Read:   resourceAliCloudSlsProjectRead,
		Update: resourceAliCloudSlsProjectUpdate,
		Delete: resourceAliCloudSlsProjectDelete,
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
				ValidateFunc: StringMatch(regexp.MustCompile("^[0-9a-zA-Z_-]+$"), "The name of the log project. It is the only in one Alicloud account. The project name is globally unique in Alibaba Cloud and cannot be modified after it is created. The naming rules are as follows:- The project name must be globally unique. - The name can contain only lowercase letters, digits, and hyphens (-). - It must start and end with a lowercase letter or number. - The value contains 3 to 63 characters."),
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
			"policy": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchema(),
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Deprecated:   "Field 'name' has been deprecated since provider version 1.223.0. New field 'project_name' instead.",
				ForceNew:     true,
				ValidateFunc: StringMatch(regexp.MustCompile("^[0-9a-zA-Z_-]+$"), "The name of the log project. It is the only in one Alicloud account. The project name is globally unique in Alibaba Cloud and cannot be modified after it is created. The naming rules are as follows:- The project name must be globally unique. - The name can contain only lowercase letters, digits, and hyphens (-). - It must start and end with a lowercase letter or number. - The value contains 3 to 63 characters."),
			},
		},
	}
}

func resourceAliCloudSlsProjectCreate(d *schema.ResourceData, meta interface{}) error {
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

	// Prepare create project request
	createRequest := map[string]interface{}{
		"projectName": projectName,
	}

	if v, ok := d.GetOk("description"); ok {
		createRequest["description"] = v.(string)
	}
	if v, ok := d.GetOk("resource_group_id"); ok {
		createRequest["resourceGroupId"] = v.(string)
	}

	// Create project using SlsService
	err = slsService.CreateProject(createRequest)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_project", "CreateProject", AlibabaCloudSdkGoERROR)
	}

	d.SetId(projectName)

	// Use StateRefreshFunc to wait for project creation completion
	stateConf := BuildStateConf([]string{}, []string{"Normal"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, slsService.SlsProjectStateRefreshFunc(d.Id(), "status", []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudSlsProjectRead(d, meta)
}

func resourceAliCloudSlsProjectRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_project", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	// Get project details
	objectRaw, err := slsService.DescribeSlsProject(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_log_project DescribeSlsProject Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set basic project attributes
	d.Set("create_time", objectRaw["createTime"])
	d.Set("description", objectRaw["description"])
	d.Set("resource_group_id", objectRaw["resourceGroupId"])
	d.Set("status", objectRaw["status"])
	d.Set("project_name", objectRaw["projectName"])

	// Get and set tags
	tagObjectRaw, err := slsService.DescribeListTagResources(d.Id())
	if err != nil {
		return WrapError(err)
	}

	tagsMaps := tagObjectRaw["tagResources"]
	d.Set("tags", tagsToMap(tagsMaps))

	// Set project policy if exists
	if policy, exists := objectRaw["policy"]; exists && policy != "" && policy != "{}" {
		d.Set("policy", policy)
	}

	// Set deprecated name field for backward compatibility
	d.Set("name", d.Get("project_name"))

	return nil
}

func resourceAliCloudSlsProjectUpdate(d *schema.ResourceData, meta interface{}) error {
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

	// Update tags
	if d.HasChange("tags") {
		if err := slsService.SetResourceTags(d, "PROJECT"); err != nil {
			return WrapError(err)
		}
		d.SetPartial("tags")
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

	d.Partial(false)
	return resourceAliCloudSlsProjectRead(d, meta)
}

func resourceAliCloudSlsProjectDelete(d *schema.ResourceData, meta interface{}) error {
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
	stateConf := BuildStateConf([]string{"Normal"}, []string{}, d.Timeout(schema.TimeoutDelete), 5*time.Second, slsService.SlsProjectStateRefreshFunc(d.Id(), "status", []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
