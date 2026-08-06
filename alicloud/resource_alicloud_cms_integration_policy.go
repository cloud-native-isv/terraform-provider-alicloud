package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	cmsapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/cms"
	commonapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// resourceAliCloudCmsIntegrationPolicy manages a CMS 2.0 integration
// policy through the cws-lib-go cms API layer (full CRUD).
func resourceAliCloudCmsIntegrationPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudCmsIntegrationPolicyCreate,
		Read:   resourceAliCloudCmsIntegrationPolicyRead,
		Update: resourceAliCloudCmsIntegrationPolicyUpdate,
		Delete: resourceAliCloudCmsIntegrationPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the integration policy.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the integration policy.",
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The policy definition version reported by the service.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the integration policy.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the integration policy.",
			},
			"update_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last update time of the integration policy.",
			},
		},
	}
}

func resourceAliCloudCmsIntegrationPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	policy := &cmsapi.CmsIntegrationPolicy{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}
	result, err := service.CreateCmsIntegrationPolicy(policy)
	if err != nil {
		return WrapError(err)
	}
	if result.PolicyId == "" {
		return fmt.Errorf("create CMS integration policy %q returned an empty policy id", policy.Name)
	}

	d.SetId(result.PolicyId)
	return resourceAliCloudCmsIntegrationPolicyRead(d, meta)
}

func resourceAliCloudCmsIntegrationPolicyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	policy, err := service.GetCmsIntegrationPolicy(d.Id())
	if err != nil {
		if commonapi.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	if policy.Name != "" {
		d.Set("name", policy.Name)
	}
	if policy.Description != "" {
		d.Set("description", policy.Description)
	}
	d.Set("version", policy.Version)
	d.Set("status", policy.Status)
	d.Set("create_time", policy.CreateTime)
	d.Set("update_time", policy.UpdateTime)
	return nil
}

func resourceAliCloudCmsIntegrationPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	if d.HasChange("name") || d.HasChange("description") {
		policy := &cmsapi.CmsIntegrationPolicy{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
		}
		if _, err := service.UpdateCmsIntegrationPolicy(d.Id(), policy); err != nil {
			return WrapError(err)
		}
	}
	return resourceAliCloudCmsIntegrationPolicyRead(d, meta)
}

func resourceAliCloudCmsIntegrationPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	if err := service.DeleteCmsIntegrationPolicy(d.Id()); err != nil {
		return WrapError(err)
	}
	return nil
}
