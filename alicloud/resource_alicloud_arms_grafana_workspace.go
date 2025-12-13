// Package alicloud. This file is generated automatically. Please do not modify it manually, thank you!
package alicloud

import (
	"log"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudArmsGrafanaWorkspace() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudArmsGrafanaWorkspaceCreate,
		Read:   resourceAliCloudArmsGrafanaWorkspaceRead,
		Update: resourceAliCloudArmsGrafanaWorkspaceUpdate,
		Delete: resourceAliCloudArmsGrafanaWorkspaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"account_number": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"aliyun_lang": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"auto_renew": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_account_number": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"duration": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"grafana_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"grafana_workspace_edition": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"grafana_workspace_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"pricing_cycle": {
				Type:     schema.TypeString,
				Optional: true,
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAliCloudArmsGrafanaWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Build parameters
	grafanaWorkspaceName := d.Get("grafana_workspace_name").(string)
	grafanaWorkspaceEdition := d.Get("grafana_workspace_edition").(string)
	description := ""
	if v, ok := d.GetOk("description"); ok {
		description = v.(string)
	}
	resourceGroupId := ""
	if v, ok := d.GetOk("resource_group_id"); ok {
		resourceGroupId = v.(string)
	}
	grafanaVersion := ""
	if v, ok := d.GetOk("grafana_version"); ok {
		grafanaVersion = v.(string)
	}
	password := ""
	if v, ok := d.GetOk("password"); ok {
		password = v.(string)
	}
	aliyunLang := ""
	if v, ok := d.GetOk("aliyun_lang"); ok {
		aliyunLang = v.(string)
	}

	// TODO: Add support for tags when needed
	var tags []aliyunArmsAPI.GrafanaWorkspaceTag

	// Create the workspace using Service layer
	var workspace *aliyunArmsAPI.GrafanaWorkspace
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		result, err := service.CreateGrafanaWorkspace(grafanaWorkspaceName, grafanaWorkspaceEdition, description, resourceGroupId, grafanaVersion, password, aliyunLang, tags)
		if err != nil {
			if NeedRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		workspace = result
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_grafana_workspace", "CreateGrafanaWorkspace", AlibabaCloudSdkGoERROR)
	}

	d.SetId(tea.StringValue(workspace.GrafanaWorkspaceId))

	// Wait for creation to complete
	err = service.WaitForArmsGrafanaWorkspaceCreated(d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudArmsGrafanaWorkspaceRead(d, meta)
}

func resourceAliCloudArmsGrafanaWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	workspace, err := service.DescribeArmsGrafanaWorkspace(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_arms_grafana_workspace DescribeArmsGrafanaWorkspace Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set all fields using strong types
	d.Set("create_time", timeToString(workspace.GmtCreate))
	d.Set("description", workspace.Description)
	d.Set("grafana_version", workspace.GrafanaVersion)
	d.Set("grafana_workspace_edition", workspace.GrafanaWorkspaceEdition)
	d.Set("grafana_workspace_name", workspace.GrafanaWorkspaceName)
	d.Set("region_id", workspace.RegionId)
	d.Set("resource_group_id", workspace.ResourceGroupId)
	d.Set("status", workspace.Status)

	return nil
}

func resourceAliCloudArmsGrafanaWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	d.Partial(true)

	// Check for basic workspace updates
	if !d.IsNewResource() && (d.HasChange("grafana_workspace_name") || d.HasChange("description")) {
		grafanaWorkspaceName := d.Get("grafana_workspace_name").(string)
		description := ""
		if v, ok := d.GetOk("description"); ok {
			description = v.(string)
		}
		aliyunLang := ""
		if v, ok := d.GetOk("aliyun_lang"); ok {
			aliyunLang = v.(string)
		}

		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err := service.UpdateGrafanaWorkspace(d.Id(), grafanaWorkspaceName, description, aliyunLang)
			if err != nil {
				if NeedRetry(err) {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateGrafanaWorkspace", AlibabaCloudSdkGoERROR)
		}
	}

	d.Partial(false)
	return resourceAliCloudArmsGrafanaWorkspaceRead(d, meta)
}

func resourceAliCloudArmsGrafanaWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewArmsService(client)
	if err != nil {
		return WrapError(err)
	}

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := service.DeleteGrafanaWorkspace(d.Id())
		if err != nil {
			if NeedRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteGrafanaWorkspace", AlibabaCloudSdkGoERROR)
	}

	// Wait for deletion to complete
	err = service.WaitForArmsGrafanaWorkspaceDeleted(d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
