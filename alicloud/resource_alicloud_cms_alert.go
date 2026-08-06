package alicloud

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	cmsapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/cms"
	commonapi "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// resourceAliCloudCmsAlert manages a CMS 2.0 notification policy
// (NotifyPolicy) inside a CMS workspace through the cws-lib-go cms API
// layer. Alert notifications converge on NotifyPolicy (goal C4).
//
// The NotifyPolicy payload sections (notify_strategy, subscription,
// response_plan) are deeply nested typed structures; they are carried in
// configuration as JSON documents and expanded into the cws-lib-go typed
// structs before every call, so no map[string]any crosses the service
// boundary. A JSON attribute is only written back during Read when it is
// managed in configuration, which avoids spurious diffs caused by
// server-side defaults.
//
// The composite resource ID is "<workspace>:<policy_uuid>".
func resourceAliCloudCmsAlert() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudCmsAlertCreate,
		Read:   resourceAliCloudCmsAlertRead,
		Update: resourceAliCloudCmsAlertUpdate,
		Delete: resourceAliCloudCmsAlertDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"workspace": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The CMS workspace the notification policy belongs to.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the notification policy.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the notification policy.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the notification policy is enabled.",
			},
			"notify_strategy": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The notify strategy section of the policy as a JSON document.",
			},
			"subscription": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The subscription section of the policy as a JSON document.",
			},
			"response_plan": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The response plan section of the policy as a JSON document.",
			},
			"version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The optimistic-lock version of the policy.",
			},
			"user_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The owning account of the policy.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the policy.",
			},
			"update_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last update time of the policy.",
			},
		},
	}
}

func cmsAlertId(workspace, policyID string) string {
	return workspace + ":" + policyID
}

func parseCmsAlertId(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid CMS alert id %q, expected <workspace>:<policy_uuid>", id)
	}
	return parts[0], parts[1], nil
}

// expandCmsAlertPolicy builds the typed NotifyPolicy payload from the
// Terraform configuration, expanding the JSON document attributes into
// cws-lib-go typed structs.
func expandCmsAlertPolicy(d *schema.ResourceData) (*cmsapi.CmsNotifyPolicy, error) {
	policy := &cmsapi.CmsNotifyPolicy{
		Workspace:   d.Get("workspace").(string),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Version:     int32(d.Get("version").(int)),
	}
	if raw := d.Get("notify_strategy").(string); raw != "" {
		strategy := &cmsapi.CmsNotifyStrategy{}
		if err := json.Unmarshal([]byte(raw), strategy); err != nil {
			return nil, fmt.Errorf("notify_strategy is not a valid JSON document: %w", err)
		}
		policy.NotifyStrategy = strategy
	}
	if raw := d.Get("subscription").(string); raw != "" {
		subscription := &cmsapi.CmsNotifyPolicySubscription{}
		if err := json.Unmarshal([]byte(raw), subscription); err != nil {
			return nil, fmt.Errorf("subscription is not a valid JSON document: %w", err)
		}
		policy.Subscription = subscription
	}
	if raw := d.Get("response_plan").(string); raw != "" {
		responsePlan := &cmsapi.CmsResponsePlan{}
		if err := json.Unmarshal([]byte(raw), responsePlan); err != nil {
			return nil, fmt.Errorf("response_plan is not a valid JSON document: %w", err)
		}
		policy.ResponsePlan = responsePlan
	}
	return policy, nil
}

// syncCmsAlertDocument writes a JSON document attribute back to state only
// when the attribute is managed in configuration and the remote document
// differs semantically from the configured one.
func syncCmsAlertDocument(d *schema.ResourceData, attr string, value interface{}) error {
	stateRaw := d.Get(attr).(string)
	if stateRaw == "" {
		return nil
	}
	var stateValue interface{}
	if err := json.Unmarshal([]byte(stateRaw), &stateValue); err == nil {
		if remoteRaw, err := json.Marshal(value); err == nil {
			var remoteValue interface{}
			if err := json.Unmarshal(remoteRaw, &remoteValue); err == nil && reflect.DeepEqual(stateValue, remoteValue) {
				return nil
			}
		}
	}
	remoteRaw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("serialize remote %s: %w", attr, err)
	}
	d.Set(attr, string(remoteRaw))
	return nil
}

func resourceAliCloudCmsAlertCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	workspace := d.Get("workspace").(string)
	policy, err := expandCmsAlertPolicy(d)
	if err != nil {
		return WrapError(err)
	}

	result, err := service.CreateCmsNotifyPolicy(policy)
	if err != nil {
		return WrapError(err)
	}
	if result.Uuid == "" {
		return fmt.Errorf("create CMS notify policy %q in workspace %q returned an empty uuid", policy.Name, workspace)
	}

	d.SetId(cmsAlertId(workspace, result.Uuid))

	// The create payload carries no enable flag, so apply the desired
	// enabled state through the dedicated enable/disable actions.
	if !d.Get("enabled").(bool) {
		if err := service.DisableCmsNotifyPolicy(workspace, result.Uuid); err != nil {
			return WrapError(err)
		}
	}
	return resourceAliCloudCmsAlertRead(d, meta)
}

func resourceAliCloudCmsAlertRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	workspace, policyID, err := parseCmsAlertId(d.Id())
	if err != nil {
		return WrapError(err)
	}
	policy, err := service.GetCmsNotifyPolicy(workspace, policyID)
	if err != nil {
		if commonapi.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("workspace", firstNonEmptyString(policy.Workspace, workspace))
	if policy.Name != "" {
		d.Set("name", policy.Name)
	}
	if policy.Description != "" {
		d.Set("description", policy.Description)
	}
	d.Set("enabled", policy.Enabled)
	d.Set("version", int(policy.Version))
	d.Set("user_id", policy.UserId)
	d.Set("create_time", policy.CreateTime)
	d.Set("update_time", policy.UpdateTime)

	if policy.NotifyStrategy != nil {
		if err := syncCmsAlertDocument(d, "notify_strategy", policy.NotifyStrategy); err != nil {
			return WrapError(err)
		}
	}
	if policy.Subscription != nil {
		if err := syncCmsAlertDocument(d, "subscription", policy.Subscription); err != nil {
			return WrapError(err)
		}
	}
	if policy.ResponsePlan != nil {
		if err := syncCmsAlertDocument(d, "response_plan", policy.ResponsePlan); err != nil {
			return WrapError(err)
		}
	}
	return nil
}

func resourceAliCloudCmsAlertUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	workspace, policyID, err := parseCmsAlertId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("notify_strategy") ||
		d.HasChange("subscription") || d.HasChange("response_plan") {
		policy, err := expandCmsAlertPolicy(d)
		if err != nil {
			return WrapError(err)
		}
		policy.Uuid = policyID
		policy.Workspace = workspace
		if _, err := service.UpdateCmsNotifyPolicy(policy); err != nil {
			return WrapError(err)
		}
	}

	if d.HasChange("enabled") {
		if d.Get("enabled").(bool) {
			if err := service.EnableCmsNotifyPolicy(workspace, policyID); err != nil {
				return WrapError(err)
			}
		} else {
			if err := service.DisableCmsNotifyPolicy(workspace, policyID); err != nil {
				return WrapError(err)
			}
		}
	}
	return resourceAliCloudCmsAlertRead(d, meta)
}

func resourceAliCloudCmsAlertDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewCmsService(client)
	if err != nil {
		return WrapError(err)
	}

	workspace, policyID, err := parseCmsAlertId(d.Id())
	if err != nil {
		return WrapError(err)
	}
	if err := service.DeleteCmsNotifyPolicy(workspace, policyID); err != nil {
		return WrapError(err)
	}
	return nil
}
