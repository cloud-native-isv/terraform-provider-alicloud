package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudResourceManagerRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudResourceManagerRoleCreate,
		Read:   resourceAliCloudResourceManagerRoleRead,
		Update: resourceAliCloudResourceManagerRoleUpdate,
		Delete: resourceAliCloudResourceManagerRoleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"assume_role_policy_document": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					equal, _ := compareJsonTemplateAreEquivalent(old, new)
					return equal
				},
			},
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
				Removed:  "Field 'create_date' has been removed from provider version 1.114.0.",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"max_session_duration": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3600,
			},
			"role_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"update_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAliCloudResourceManagerRoleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var response map[string]interface{}
	action := "CreateRole"
	request := make(map[string]interface{})
	var err error
	request["AssumeRolePolicyDocument"] = d.Get("assume_role_policy_document")
	if v, ok := d.GetOk("description"); ok {
		request["Description"] = v
	}

	if v, ok := d.GetOk("max_session_duration"); ok {
		request["MaxSessionDuration"] = v
	}

	request["RoleName"] = d.Get("role_name")
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err = client.RpcPost("ResourceManager", "2020-03-31", action, nil, request, false)
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
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_resource_manager_role", action, AlibabaCloudSdkGoERROR)
	}
	responseRole := response["Role"].(map[string]interface{})
	d.SetId(fmt.Sprint(responseRole["RoleName"]))

	return resourceAliCloudResourceManagerRoleRead(d, meta)
}
func resourceAliCloudResourceManagerRoleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	resourcemanagerService := ResourcemanagerService{client}
	object, err := resourcemanagerService.DescribeResourceManagerRole(d.Id())
	if err != nil {
		if IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_resource_manager_role resourcemanagerService.DescribeResourceManagerRole Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("role_name", d.Id())
	d.Set("arn", object["Arn"])
	d.Set("assume_role_policy_document", object["AssumeRolePolicyDocument"])
	d.Set("description", object["Description"])
	d.Set("max_session_duration", object["MaxSessionDuration"])
	d.Set("role_id", object["RoleId"])
	d.Set("update_date", object["UpdateDate"])
	return nil
}
func resourceAliCloudResourceManagerRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	var err error
	var response map[string]interface{}
	update := false
	request := map[string]interface{}{
		"RoleName": d.Id(),
	}
	if d.HasChange("assume_role_policy_document") {
		update = true
	}
	request["NewAssumeRolePolicyDocument"] = d.Get("assume_role_policy_document")
	if d.HasChange("max_session_duration") {
		update = true
		request["NewMaxSessionDuration"] = d.Get("max_session_duration")
	}
	if update {
		action := "UpdateRole"
		wait := incrementalWait(3*time.Second, 3*time.Second)
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			response, err = client.RpcPost("ResourceManager", "2020-03-31", action, nil, request, false)
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
	}
	return resourceAliCloudResourceManagerRoleRead(d, meta)
}
func resourceAliCloudResourceManagerRoleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	action := "DeleteRole"
	var response map[string]interface{}
	var err error
	request := map[string]interface{}{
		"RoleName": d.Id(),
	}

	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		response, err = client.RpcPost("ResourceManager", "2020-03-31", action, nil, request, false)
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
		if IsExpectedErrors(err, []string{"EntityNotExist.Role"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
	}
	return nil
}
