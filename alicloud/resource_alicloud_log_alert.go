package alicloud

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAlicloudLogAlert() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudLogAlertCreate,
		Read:   resourceAlicloudLogAlertRead,
		Update: resourceAlicloudLogAlertUpdate,
		Delete: resourceAlicloudLogAlertDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"alert_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"alert_displayname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"alert_description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"condition": {
				Type:       schema.TypeString,
				Optional:   true,
				Deprecated: "Deprecated from 1.161.0+, use eval_condition in severity_configurations",
			},
			"dashboard": {
				Type:       schema.TypeString,
				Optional:   true,
				Deprecated: "Deprecated from 1.161.0+, use dashboardId in query_list",
			},
			"mute_until": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"throttling": {
				Type:       schema.TypeString,
				Optional:   true,
				Deprecated: "Deprecated from 1.161.0+, use repeat_interval in policy_configuration",
			},
			"notify_threshold": {
				Type:       schema.TypeInt,
				Optional:   true,
				Deprecated: "Deprecated from 1.161.0+, use threshold",
			},
			"threshold": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"no_data_fire": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"no_data_severity": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntInSlice([]int{2, 4, 6, 8, 10}),
			},
			"send_resolved": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"auto_annotation": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"query_list": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"chart_title": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"project": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"store": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"store_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"query": {
							Type:     schema.TypeString,
							Required: true,
						},
						"logstore": {
							Type:       schema.TypeString,
							Optional:   true,
							Deprecated: "Deprecated from 1.161.0+, use store",
						},
						"start": {
							Type:     schema.TypeString,
							Required: true,
						},
						"end": {
							Type:     schema.TypeString,
							Required: true,
						},
						"time_span_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "Custom",
						},
						"region": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"role_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dashboard_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"power_sql_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"auto", "enable", "disable"}, false),
						},
					},
				},
			},

			"notification_list": {
				Type:       schema.TypeList,
				Optional:   true,
				Deprecated: "Deprecated from 1.161.0+, use policy_configuration for notification",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"sms",
								"dingtalk",
								"email",
								"messageCenter"},
								false),
						},
						"content": {
							Type:     schema.TypeString,
							Required: true,
						},
						"service_uri": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mobile_list": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"email_list": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"labels": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"annotations": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"severity_configurations": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"severity": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntInSlice([]int{2, 4, 6, 8, 10}),
						},
						"eval_condition": {
							Type:     schema.TypeMap,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"join_configurations": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"cross_join", "inner_join", "left_join", "right_join", "full_join", "left_exclude", "right_exclude", "concat", "no_join"}, false),
						},
						"condition": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"policy_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alert_policy_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"action_policy_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"repeat_interval": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"group_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"fields": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"template_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"lang": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"sys", "user"}, false),
						},
						"tokens": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"annotations": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"schedule_interval": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Deprecated:    "Field 'schedule_interval' has been deprecated from provider version 1.176.0. New field 'schedule' instead.",
				ConflictsWith: []string{"schedule"},
			},
			"schedule_type": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Deprecated:    "Field 'schedule_type' has been deprecated from provider version 1.176.0. New field 'schedule' instead.",
				ConflictsWith: []string{"schedule"},
			},
			"schedule": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"FixedRate", "Cron", "Hourly", "Daily", "Weekly"}, false),
						},
						"interval": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"cron_expression": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"day_of_week": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"hour": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"delay": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"run_immediately": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"time_zone": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				ConflictsWith: []string{"schedule_type", "schedule_interval"},
			},
		},
	}
}

func resourceAlicloudLogAlertCreate(d *schema.ResourceData, meta interface{}) error {
	// TODO: Log alert management is not yet fully implemented in the new SLS API
	// This resource needs to be updated to use the new API methods when they become available
	return WrapError(fmt.Errorf("log alert management is temporarily unavailable during API migration"))
}

func resourceAlicloudLogAlertRead(d *schema.ResourceData, meta interface{}) error {
	// TODO: Log alert management is not yet fully implemented in the new SLS API
	d.SetId("")
	return nil
}

func resourceAlicloudLogAlertUpdate(d *schema.ResourceData, meta interface{}) error {
	// TODO: Log alert management is not yet fully implemented in the new SLS API
	return WrapError(fmt.Errorf("log alert management is temporarily unavailable during API migration"))
}

func resourceAlicloudLogAlertDelete(d *schema.ResourceData, meta interface{}) error {
	// TODO: Log alert management is not yet fully implemented in the new SLS API
	return nil
}
