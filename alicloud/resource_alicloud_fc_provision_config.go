// Package alicloud. This file is generated automatically. Please do not modify it manually, thank you!
package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	aliyunFCAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/fc/v3"
)

func resourceAliCloudFCProvisionConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFCProvisionConfigCreate,
		Read:   resourceAliCloudFCProvisionConfigRead,
		Update: resourceAliCloudFCProvisionConfigUpdate,
		Delete: resourceAliCloudFCProvisionConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"always_allocate_cpu": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"always_allocate_gpu": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"current": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"current_error": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"function_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"qualifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"scheduled_actions": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"schedule_expression": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"target": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"time_zone": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"end_time": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"start_time": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"target": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: IntBetween(0, 10000),
			},
			"target_tracking_policies": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"metric_target": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"time_zone": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"end_time": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"metric_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"start_time": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"min_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"max_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

// BuildProvisionConfigFromSchema builds ProvisionConfig from Terraform schema data
func BuildProvisionConfigFromSchema(d *schema.ResourceData) *aliyunFCAPI.ProvisionConfig {
	config := &aliyunFCAPI.ProvisionConfig{}

	if v, ok := d.GetOk("always_allocate_cpu"); ok {
		config.AlwaysAllocateCPU = tea.Bool(v.(bool))
	}

	if v, ok := d.GetOk("always_allocate_gpu"); ok {
		config.AlwaysAllocateGPU = tea.Bool(v.(bool))
	}

	if v, ok := d.GetOk("target"); ok {
		config.Target = tea.Int64(int64(v.(int)))
	}

	// Add scheduled actions
	if v, ok := d.GetOk("scheduled_actions"); ok {
		if actions := v.([]interface{}); len(actions) > 0 {
			config.ScheduledActions = make([]*aliyunFCAPI.ScheduledAction, len(actions))
			for i, action := range actions {
				actionMap := action.(map[string]interface{})
				scheduledAction := &aliyunFCAPI.ScheduledAction{}

				if name, ok := actionMap["name"].(string); ok && name != "" {
					scheduledAction.Name = tea.String(name)
				}

				if startTime, ok := actionMap["start_time"].(string); ok && startTime != "" {
					scheduledAction.StartTime = tea.String(startTime)
				}

				if endTime, ok := actionMap["end_time"].(string); ok && endTime != "" {
					scheduledAction.EndTime = tea.String(endTime)
				}

				if target, ok := actionMap["target"].(int); ok {
					scheduledAction.Target = tea.Int64(int64(target))
				}

				if schedule, ok := actionMap["schedule_expression"].(string); ok && schedule != "" {
					scheduledAction.Schedule = tea.String(schedule)
				}

				// Note: TimeZone field is not available in ScheduledAction in FC v3
				// We'll ignore it for now

				config.ScheduledActions[i] = scheduledAction
			}
		}
	}

	// Add target tracking policies
	if v, ok := d.GetOk("target_tracking_policies"); ok {
		if policies := v.([]interface{}); len(policies) > 0 {
			config.TargetTrackingPolicies = make([]*aliyunFCAPI.TargetTrackingPolicy, len(policies))
			for i, policy := range policies {
				policyMap := policy.(map[string]interface{})
				targetTrackingPolicy := &aliyunFCAPI.TargetTrackingPolicy{}

				if name, ok := policyMap["name"].(string); ok && name != "" {
					targetTrackingPolicy.Name = tea.String(name)
				}

				// Note: StartTime and EndTime fields are not available in TargetTrackingPolicy in FC v3
				// We'll ignore them for now

				if metricType, ok := policyMap["metric_type"].(string); ok && metricType != "" {
					targetTrackingPolicy.MetricType = tea.String(metricType)
				}

				if metricTarget, ok := policyMap["metric_target"].(float64); ok {
					targetTrackingPolicy.TargetValue = tea.Float64(metricTarget)
				}

				// Note: MinCapacity and MaxCapacity fields are not available in TargetTrackingPolicy in FC v3
				// We'll ignore them for now

				// Note: TimeZone field is not available in TargetTrackingPolicy in FC v3
				// We'll ignore it for now

				config.TargetTrackingPolicies[i] = targetTrackingPolicy
			}
		}
	}

	return config
}

// SetSchemaFromProvisionConfig sets terraform schema data from ProvisionConfig
func SetSchemaFromProvisionConfig(d *schema.ResourceData, config *aliyunFCAPI.ProvisionConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.AlwaysAllocateCPU != nil {
		d.Set("always_allocate_cpu", *config.AlwaysAllocateCPU)
	}

	if config.AlwaysAllocateGPU != nil {
		d.Set("always_allocate_gpu", *config.AlwaysAllocateGPU)
	}

	if config.Current != nil {
		d.Set("current", *config.Current)
	}

	if config.CurrentError != nil {
		d.Set("current_error", *config.CurrentError)
	}

	if config.FunctionArn != nil {
		d.Set("function_arn", *config.FunctionArn)
	}

	if config.Target != nil {
		d.Set("target", *config.Target)
	}

	// Set scheduled actions
	if config.ScheduledActions != nil {
		scheduledActionsMaps := make([]map[string]interface{}, 0)
		for _, action := range config.ScheduledActions {
			if action != nil {
				actionMap := make(map[string]interface{})

				if action.Name != nil {
					actionMap["name"] = *action.Name
				}

				if action.StartTime != nil {
					actionMap["start_time"] = *action.StartTime
				}

				if action.EndTime != nil {
					actionMap["end_time"] = *action.EndTime
				}

				if action.Target != nil {
					actionMap["target"] = *action.Target
				}

				if action.Schedule != nil {
					actionMap["schedule_expression"] = *action.Schedule
				}

				// Note: TimeZone field is not available in ScheduledAction in FC v3
				// We'll leave it as empty

				scheduledActionsMaps = append(scheduledActionsMaps, actionMap)
			}
		}
		d.Set("scheduled_actions", scheduledActionsMaps)
	}

	// Set target tracking policies
	if config.TargetTrackingPolicies != nil {
		targetTrackingPoliciesMaps := make([]map[string]interface{}, 0)
		for _, policy := range config.TargetTrackingPolicies {
			if policy != nil {
				policyMap := make(map[string]interface{})

				if policy.Name != nil {
					policyMap["name"] = *policy.Name
				}

				// Note: StartTime and EndTime fields are not available in TargetTrackingPolicy in FC v3
				// We'll leave them as empty

				if policy.MetricType != nil {
					policyMap["metric_type"] = *policy.MetricType
				}

				if policy.TargetValue != nil {
					policyMap["metric_target"] = *policy.TargetValue
				}

				// Note: MinCapacity and MaxCapacity fields are not available in TargetTrackingPolicy in FC v3
				// We'll leave them as empty

				// Note: TimeZone field is not available in TargetTrackingPolicy in FC v3
				// We'll leave it as empty

				targetTrackingPoliciesMaps = append(targetTrackingPoliciesMaps, policyMap)
			}
		}
		d.Set("target_tracking_policies", targetTrackingPoliciesMaps)
	}

	return nil
}

func resourceAliCloudFCProvisionConfigCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName := d.Get("function_name").(string)
	qualifier := "LATEST"
	if v, ok := d.GetOk("qualifier"); ok {
		qualifier = v.(string)
	}

	log.Printf("[DEBUG] Creating FC Provision Config for function: %s, qualifier: %s", functionName, qualifier)

	// Build provision config from schema
	config := BuildProvisionConfigFromSchema(d)

	// Create provision config using service layer
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err = fcService.CreateFCProvisionConfig(functionName, qualifier, config)
		if err != nil {
			if NeedRetry(err) {
				log.Printf("[WARN] FC Provision Config creation failed with retryable error: %s. Retrying...", err)
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			log.Printf("[ERROR] FC Provision Config creation failed: %s", err)
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_fc_provision_config", "CreateProvisionConfig", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID
	d.SetId(functionName)
	log.Printf("[DEBUG] FC Provision Config created successfully for function: %s", functionName)

	return resourceAliCloudFCProvisionConfigRead(d, meta)
}

func resourceAliCloudFCProvisionConfigRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName := d.Id()
	qualifier := "LATEST"
	if v, ok := d.GetOk("qualifier"); ok {
		qualifier = v.(string)
	}

	objectRaw, err := fcService.DescribeFCProvisionConfig(functionName, qualifier)
	if err != nil {
		if !d.IsNewResource() && IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_fc_provision_config DescribeFCProvisionConfig Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Use helper to set schema fields
	err = SetSchemaFromProvisionConfig(d, objectRaw)
	if err != nil {
		return WrapError(err)
	}

	d.Set("function_name", functionName)

	return nil
}

func resourceAliCloudFCProvisionConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName := d.Id()
	qualifier := "LATEST"
	if v, ok := d.GetOk("qualifier"); ok {
		qualifier = v.(string)
	}

	log.Printf("[DEBUG] Updating FC Provision Config for function: %s, qualifier: %s", functionName, qualifier)

	// Check if any field has changed
	if d.HasChange("always_allocate_cpu") || d.HasChange("always_allocate_gpu") ||
		d.HasChange("target") || d.HasChange("scheduled_actions") ||
		d.HasChange("target_tracking_policies") {
		// Build provision config from schema
		config := BuildProvisionConfigFromSchema(d)

		// Update provision config using service layer
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err := fcService.UpdateFCProvisionConfig(functionName, qualifier, config)
			if err != nil {
				if NeedRetry(err) {
					log.Printf("[WARN] FC Provision Config update failed with retryable error: %s. Retrying...", err)
					time.Sleep(5 * time.Second)
					return resource.RetryableError(err)
				}
				log.Printf("[ERROR] FC Provision Config update failed: %s", err)
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateProvisionConfig", AlibabaCloudSdkGoERROR)
		}

		log.Printf("[DEBUG] FC Provision Config updated successfully for function: %s", functionName)
	}

	return resourceAliCloudFCProvisionConfigRead(d, meta)
}

func resourceAliCloudFCProvisionConfigDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	fcService, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName := d.Id()
	qualifier := "LATEST"
	if v, ok := d.GetOk("qualifier"); ok {
		qualifier = v.(string)
	}

	log.Printf("[DEBUG] Deleting FC Provision Config for function: %s, qualifier: %s", functionName, qualifier)

	// Delete provision config using service layer
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := fcService.DeleteFCProvisionConfig(functionName, qualifier)
		if err != nil {
			if IsNotFoundError(err) {
				log.Printf("[DEBUG] FC Provision Config not found during deletion for function: %s", functionName)
				return nil
			}
			if NeedRetry(err) {
				log.Printf("[WARN] FC Provision Config deletion failed with retryable error: %s. Retrying...", err)
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			log.Printf("[ERROR] FC Provision Config deletion failed: %s", err)
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteProvisionConfig", AlibabaCloudSdkGoERROR)
	}

	log.Printf("[DEBUG] FC Provision Config deleted successfully for function: %s", functionName)

	return nil
}
