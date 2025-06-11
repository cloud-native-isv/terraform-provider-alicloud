package alicloud

import (
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	armsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAlicloudArmsAlertNotificationPolicies() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudArmsAlertNotificationPoliciesRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
				ForceNew:     true,
			},
			"names": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"notification_policy_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"state": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"policies": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"notification_policy_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"notification_policy_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"send_recover_message": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"repeat_interval": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"escalation_policy_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"create_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"update_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"group_rule": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"group_wait": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"group_interval": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"grouping_fields": {
										Type:     schema.TypeList,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"matching_rules": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"matching_conditions": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"operator": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"value": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"notify_rule": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"notify_channels": {
										Type:     schema.TypeList,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"notify_objects": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"notify_object_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"notify_object_name": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"notify_object_type": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceAlicloudArmsAlertNotificationPoliciesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Initialize ARMS API client with credentials
	credentials := &armsAPI.ArmsCredentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	armsClient, err := armsAPI.NewArmsAPI(credentials)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_arms_alert_notification_policies", "InitializeArmsAPI", "Failed to initialize ARMS API client")
	}

	// Build filter parameters from input
	var notificationPolicyNameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		r, err := regexp.Compile(v.(string))
		if err != nil {
			return WrapError(err)
		}
		notificationPolicyNameRegex = r
	}

	idsMap := make(map[string]string)
	if v, ok := d.GetOk("ids"); ok {
		for _, vv := range v.([]interface{}) {
			if vv == nil {
				continue
			}
			idsMap[vv.(string)] = vv.(string)
		}
	}

	// Call ARMS API to list notification policies
	var allPolicies []*armsAPI.AlertNotificationPolicy
	page := int64(1)
	size := int64(100) // Use reasonable page size

	for {
		// List notification policies with detail enabled
		policies, err := armsClient.ListAlertNotificationPolicies(page, size, true)
		if err != nil {
			return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_arms_alert_notification_policies", "ListAlertNotificationPolicies", "Failed to list notification policies")
		}

		// Apply filters
		for _, policy := range policies {
			// Apply name filter if specified
			if v, ok := d.GetOk("notification_policy_name"); ok && v.(string) != "" {
				if policy.Name != v.(string) {
					continue
				}
			}

			// Apply name regex filter if specified
			if notificationPolicyNameRegex != nil && !notificationPolicyNameRegex.MatchString(policy.Name) {
				continue
			}

			// Apply ID filter if specified
			if len(idsMap) > 0 {
				if _, ok := idsMap[policy.Id]; !ok {
					continue
				}
			}

			// Apply state filter if specified
			if v, ok := d.GetOkExists("state"); ok {
				stateFilter := v.(bool)
				if policy.State != "" {
					// Convert string state to boolean
					policyState := policy.State == "true" || policy.State == "ENABLE"
					if policyState != stateFilter {
						continue
					}
				}
			}

			allPolicies = append(allPolicies, policy)
		}

		// Check if we've retrieved all results
		if len(policies) < int(size) {
			break
		}
		page++
	}

	// Build output data
	ids := make([]string, 0)
	names := make([]interface{}, 0)
	s := make([]map[string]interface{}, 0)

	for _, policy := range allPolicies {
		mapping := map[string]interface{}{
			"id":                       policy.Id,
			"notification_policy_id":   policy.Id,
			"notification_policy_name": policy.Name,
			"send_recover_message":     policy.SendRecoverMessage,
			"repeat_interval":          int(policy.RepeatInterval),
			"escalation_policy_id":     policy.EscalationPolicyId,
			"state":                    policy.State == "true" || policy.State == "ENABLE",
			"create_time":              policy.CreateTime,
			"update_time":              policy.UpdateTime,
		}

		// Add group rule if available
		if policy.GroupRule != nil {
			groupRule := map[string]interface{}{
				"group_wait":      int(policy.GroupRule.GroupWait),
				"group_interval":  int(policy.GroupRule.GroupInterval),
				"grouping_fields": policy.GroupRule.GroupingFields,
			}
			mapping["group_rule"] = []interface{}{groupRule}
		}

		// Add matching rules if available
		if policy.MatchingRules != nil && len(policy.MatchingRules) > 0 {
			matchingRules := make([]interface{}, 0)
			for _, rule := range policy.MatchingRules {
				matchingConditions := make([]interface{}, 0)
				for _, condition := range rule.MatchingConditions {
					matchingConditions = append(matchingConditions, map[string]interface{}{
						"key":      condition.Key,
						"operator": condition.Operator,
						"value":    condition.Value,
					})
				}
				matchingRules = append(matchingRules, map[string]interface{}{
					"matching_conditions": matchingConditions,
				})
			}
			mapping["matching_rules"] = matchingRules
		}

		// Add notify rule if available
		if policy.NotifyRule != nil && len(policy.NotifyRule) > 0 {
			notifyRules := make([]interface{}, 0)
			for _, rule := range policy.NotifyRule {
				notifyObjects := make([]interface{}, 0)
				for _, obj := range rule.NotifyObjects {
					notifyObjects = append(notifyObjects, map[string]interface{}{
						"notify_object_id":   obj.NotifyObjectId,
						"notify_object_name": obj.NotifyObjectName,
						"notify_object_type": obj.NotifyObjectType,
					})
				}
				notifyRules = append(notifyRules, map[string]interface{}{
					"notify_channels": rule.NotifyChannels,
					"notify_objects":  notifyObjects,
				})
			}
			mapping["notify_rule"] = notifyRules
		}

		ids = append(ids, policy.Id)
		names = append(names, policy.Name)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}

	if err := d.Set("policies", s); err != nil {
		return WrapError(err)
	}
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
