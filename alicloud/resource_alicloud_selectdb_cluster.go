package alicloud

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudSelectDBCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudSelectDBClusterCreate,
		Read:   resourceAliCloudSelectDBClusterRead,
		Update: resourceAliCloudSelectDBClusterUpdate,
		Delete: resourceAliCloudSelectDBClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"db_instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"payment_type": {
				Type:         schema.TypeString,
				ValidateFunc: StringInSlice([]string{"PayAsYouGo", "Subscription"}, false),
				Required:     true,
				ForceNew:     true,
			},
			"db_cluster_class": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cache_size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"db_cluster_description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"desired_params": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"desired_status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: StringInSlice([]string{"STOPPING", "STARTING", "RESTART"}, false),
			},
			"elastic_rules_enable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"elastic_rules": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"execution_period": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: StringInSlice([]string{"Day", "Week"}, false),
						},
						"elastic_rule_start_time": {
							Type:     schema.TypeString,
							Required: true,
						},
						"cluster_class": {
							Type:     schema.TypeString,
							Required: true,
						},
						"rule_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},

			// computed
			"db_cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"region_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"param_change_logs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"old_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"new_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"gmt_created": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"gmt_modified": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"config_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"is_applied": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceAliCloudSelectDBClusterCreate(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*connectivity.AliyunClient)
	// selectDBService, err := NewSelectDBService(client)
	// if err != nil {
	// 	return WrapError(err)
	// }

	// request, err := buildSelectDBCreateClusterRequest(d, meta)
	// if err != nil {
	// 	return WrapError(err)
	// }
	// action := "CreateDBCluster"
	// response, err := selectDBService.RequestProcessForSelectDB(request, action, "POST")
	// if err != nil {
	// 	return WrapError(err)
	// }
	// if resp, err := jsonpath.Get("$.Data", response); err != nil || resp == nil {
	// 	return WrapErrorf(err, IdMsg, "alicloud_selectdb_db_clusters")
	// } else {
	// 	clusterId := resp.(map[string]interface{})["ClusterId"].(string)
	// 	d.SetId(fmt.Sprint(d.Get("db_instance_id").(string) + ":" + clusterId))
	// }

	// stateConf := BuildStateConf([]string{"RESOURCE_PREPARING", "CREATING"}, []string{"ACTIVATION"}, d.Timeout(schema.TimeoutCreate), 20*time.Second, selectDBService.SelectDBClusterStateRefreshFunc(d.Id(), []string{"DELETING"}))
	// if _, err := stateConf.WaitForState(); err != nil {
	// 	return WrapErrorf(err, IdMsg, d.Id())
	// }
	// return resourceAliCloudSelectDBClusterUpdate(d, meta)
	return nil
}

func resourceAliCloudSelectDBClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*connectivity.AliyunClient)
	// selectDBService, err := NewSelectDBService(client)
	// if err != nil {
	// 	return WrapError(err)
	// }
	// d.Partial(true)

	// cacheSizeModified := false
	// if !d.IsNewResource() && (d.HasChange("db_cluster_class")) {
	// 	_, newClass := d.GetChange("db_cluster_class")
	// 	cache_size := 0
	// 	if d.HasChange("cache_size") {
	// 		_, newCacheSize := d.GetChange("cache_size")
	// 		cache_size = newCacheSize.(int)
	// 		cacheSizeModified = true
	// 	}
	// 	_, err := selectDBService.ModifySelectDBCluster(d.Id(), newClass.(string), cache_size)
	// 	if err != nil {
	// 		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifyDBCluster", AlibabaCloudSdkGoERROR)
	// 	}
	// 	stateConf := BuildStateConf([]string{"RESOURCE_PREPARING", "CLASS_CHANGING"}, []string{"ACTIVATION"}, d.Timeout(schema.TimeoutUpdate), 1*time.Minute, selectDBService.SelectDBClusterStateRefreshFunc(d.Id(), []string{"DELETING"}))
	// 	if _, err := stateConf.WaitForState(); err != nil {
	// 		return WrapErrorf(err, IdMsg, d.Id())
	// 	}
	// 	d.SetPartial("db_cluster_class")
	// 	if d.HasChange("cache_size") {
	// 		d.SetPartial("cache_size")
	// 	}
	// }

	// if !d.IsNewResource() && d.HasChange("cache_size") && !cacheSizeModified {
	// 	_, newCacheSize := d.GetChange("cache_size")
	// 	db_cluster_class := d.Get("db_cluster_class").(string)
	// 	if d.HasChange("db_cluster_class") {
	// 		_, newClass := d.GetChange("db_cluster_class")
	// 		db_cluster_class = newClass.(string)
	// 	}
	// 	_, err := selectDBService.ModifySelectDBCluster(d.Id(), db_cluster_class, newCacheSize.(int))
	// 	if err != nil {
	// 		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifyDBCluster", AlibabaCloudSdkGoERROR)
	// 	}
	// 	stateConf := BuildStateConf([]string{"RESOURCE_PREPARING", "CLASS_CHANGING"}, []string{"ACTIVATION"}, d.Timeout(schema.TimeoutUpdate), 1*time.Minute, selectDBService.SelectDBClusterStateRefreshFunc(d.Id(), []string{"DELETING"}))
	// 	if _, err := stateConf.WaitForState(); err != nil {
	// 		return WrapErrorf(err, IdMsg, d.Id())
	// 	}
	// 	d.SetPartial("cache_size")
	// 	if d.HasChange("db_cluster_class") {
	// 		d.SetPartial("db_cluster_class")
	// 	}
	// }

	// if !d.IsNewResource() && d.HasChange("db_cluster_description") {
	// 	_, newDesc := d.GetChange("db_cluster_description")
	// 	_, err := selectDBService.ModifySelectDBClusterDescription(d.Id(), newDesc.(string))
	// 	if err != nil {
	// 		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifyBEClusterAttribute", AlibabaCloudSdkGoERROR)
	// 	}
	// 	d.SetPartial("db_cluster_description")
	// }

	// if !d.IsNewResource() && d.HasChange("elastic_rules_enable") {
	// 	enable := d.Get("elastic_rules_enable").(bool)
	// 	_, err := selectDBService.EnDisableScalingRules(d.Id(), enable)
	// 	if err != nil {
	// 		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "EnDisableScalingRules", AlibabaCloudSdkGoERROR)
	// 	}
	// 	d.SetPartial("elastic_rules_enable")
	// }

	// if !d.IsNewResource() && d.HasChange("elastic_rules") {
	// 	// Get existing rules to identify what needs to be created, modified, or deleted
	// 	existingRules, err := selectDBService.DescribeElasticRules(d.Id())
	// 	if err != nil {
	// 		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DescribeElasticRules", AlibabaCloudSdkGoERROR)
	// 	}

	// 	// Map existing rules by their properties for easy lookup
	// 	existingRulesMap := make(map[string]map[string]interface{})
	// 	for _, rule := range existingRules {
	// 		ruleItem := rule.(map[string]interface{})
	// 		key := fmt.Sprintf("%s:%s:%s",
	// 			ruleItem["ExecutionPeriod"],
	// 			ruleItem["ElasticRuleStartTime"],
	// 			ruleItem["ClusterClass"])
	// 		existingRulesMap[key] = ruleItem
	// 	}

	// 	// Process new rules configuration
	// 	o, n := d.GetChange("elastic_rules")
	// 	oldRules := o.([]interface{})
	// 	newRules := n.([]interface{})

	// 	// Track deleted rules by key
	// 	oldRuleKeys := make(map[string]bool)
	// 	for _, oldRule := range oldRules {
	// 		if oldRule == nil {
	// 			continue
	// 		}
	// 		oldRuleMap := oldRule.(map[string]interface{})
	// 		key := fmt.Sprintf("%s:%s:%s",
	// 			oldRuleMap["execution_period"],
	// 			oldRuleMap["elastic_rule_start_time"],
	// 			oldRuleMap["cluster_class"])
	// 		oldRuleKeys[key] = true
	// 	}

	// 	// Process new rules - create or modify
	// 	for _, newRule := range newRules {
	// 		if newRule == nil {
	// 			continue
	// 		}
	// 		newRuleMap := newRule.(map[string]interface{})
	// 		key := fmt.Sprintf("%s:%s:%s",
	// 			newRuleMap["execution_period"],
	// 			newRuleMap["elastic_rule_start_time"],
	// 			newRuleMap["cluster_class"])

	// 		// If rule exists, may need to update
	// 		if existingRule, exists := existingRulesMap[key]; exists {
	// 			ruleId := existingRule["RuleId"].(float64)
	// 			// Check if rule needs modification
	// 			if newRuleMap["cluster_class"].(string) != existingRule["ClusterClass"].(string) {
	// 				_, err := selectDBService.ModifyElasticRule(d.Id(), int(ruleId),
	// 					newRuleMap["execution_period"].(string),
	// 					newRuleMap["elastic_rule_start_time"].(string),
	// 					newRuleMap["cluster_class"].(string))
	// 				if err != nil {
	// 					return WrapErrorf(err, DefaultErrorMsg, d.Id(), "ModifyElasticRule", AlibabaCloudSdkGoERROR)
	// 				}
	// 			}
	// 			delete(oldRuleKeys, key)
	// 		} else {
	// 			// Create new rule
	// 			_, err := selectDBService.CreateElasticRule(d.Id(),
	// 				newRuleMap["execution_period"].(string),
	// 				newRuleMap["elastic_rule_start_time"].(string),
	// 				newRuleMap["cluster_class"].(string))
	// 			if err != nil {
	// 				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "CreateElasticRule", AlibabaCloudSdkGoERROR)
	// 			}
	// 		}
	// 	}

	// 	// Delete rules that no longer exist in the new configuration
	// 	for key := range oldRuleKeys {
	// 		if existingRule, exists := existingRulesMap[key]; exists {
	// 			ruleId := existingRule["RuleId"].(float64)
	// 			_, err := selectDBService.DeleteElasticRule(d.Id(), int(ruleId))
	// 			if err != nil {
	// 				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteElasticRule", AlibabaCloudSdkGoERROR)
	// 			}
	// 		}
	// 	}

	// 	d.SetPartial("elastic_rules")
	// }

	// if !d.IsNewResource() && d.HasChange("desired_status") {
	// 	_, newStatus := d.GetChange("desired_status")
	// 	oldStatus := d.Get("status")
	// 	if oldStatus.(string) != "" && newStatus.(string) != "" {
	// 		_, err := selectDBService.UpdateSelectDBClusterStatus(d.Id(), newStatus.(string))
	// 		if err != nil {
	// 			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateSelectDBClusterStatus", AlibabaCloudSdkGoERROR)
	// 		}
	// 		newStatusFinal := convertSelectDBClusterStatusActionFinal(newStatus.(string))
	// 		if newStatusFinal == "" {
	// 			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateSelectDBClusterStatus", AlibabaCloudSdkGoERROR)
	// 		}

	// 		// Wait for status change to complete
	// 		stateConf := BuildStateConf([]string{"STATUS_CHANGING", "RESOURCE_PREPARING"}, []string{newStatusFinal}, d.Timeout(schema.TimeoutUpdate), 1*time.Minute, selectDBService.SelectDBClusterStateRefreshFunc(d.Id(), []string{"DELETING"}))
	// 		if _, err := stateConf.WaitForState(); err != nil {
	// 			return WrapErrorf(err, IdMsg, d.Id())
	// 		}
	// 		d.SetPartial("desired_status")
	// 	}
	// }

	// if d.HasChange("desired_params") {
	// 	oldConfig, newConfig := d.GetChange("desired_params")
	// 	oldConfigMap := oldConfig.([]interface{})
	// 	newConfigMap := newConfig.([]interface{})
	// 	oldConfigMapIndex := make(map[string]string)
	// 	for _, v := range oldConfigMap {
	// 		item := v.(map[string]interface{})
	// 		oldConfigMapIndex[item["name"].(string)] = item["value"].(string)
	// 	}
	// 	newConfigMapIndex := make(map[string]string)
	// 	for _, v := range newConfigMap {
	// 		item := v.(map[string]interface{})
	// 		newConfigMapIndex[item["name"].(string)] = item["value"].(string)
	// 	}

	// 	diffConfig := make(map[string]string)
	// 	for k, v := range newConfigMapIndex {
	// 		if oldConfigMapIndex[k] != v {
	// 			diffConfig[k] = v
	// 		}
	// 	}

	// 	if _, err := selectDBService.UpdateSelectDBClusterConfig(d.Id(), diffConfig); err != nil {
	// 		return WrapError(err)
	// 	}
	// 	d.SetPartial("desired_params")

	// 	stateConf := BuildStateConf([]string{"RESTARTING", "MODIFY_PARAM"}, []string{"ACTIVATION"}, d.Timeout(schema.TimeoutUpdate), 10*time.Second, selectDBService.SelectDBClusterStateRefreshFunc(d.Id(), []string{}))
	// 	if _, err := stateConf.WaitForState(); err != nil {
	// 		return WrapErrorf(err, IdMsg, d.Id())
	// 	}

	// }

	// d.Partial(false)
	// return resourceAliCloudSelectDBClusterRead(d, meta)
	return nil
}

func resourceAliCloudSelectDBClusterRead(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*connectivity.AliyunClient)
	// selectDBService, err := NewSelectDBService(client)
	// if err != nil {
	// 	return WrapError(err)
	// }

	// clusterResp, err := selectDBService.DescribeSelectDBCluster(d.Id())
	// if err != nil {
	// 	return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_selectdb_db_cluster", AlibabaCloudSdkGoERROR)
	// }
	// cpu, _ := clusterResp["CpuCores"].(json.Number).Int64()
	// memory, _ := clusterResp["Memory"].(json.Number).Int64()
	// cache, _ := clusterResp["CacheStorageSizeGB"].(json.Number).Int64()

	// d.Set("status", clusterResp["Status"])
	// d.Set("create_time", clusterResp["CreatedTime"])
	// d.Set("db_cluster_description", clusterResp["DbClusterName"])
	// d.Set("payment_type", convertChargeTypeToPaymentType(clusterResp["ChargeType"]))
	// d.Set("db_instance_id", clusterResp["DbInstanceName"])
	// d.Set("db_cluster_class", clusterResp["DbClusterClass"])
	// d.Set("cpu", cpu)
	// d.Set("memory", memory)
	// d.Set("cache_size", cache)

	// d.Set("engine", fmt.Sprint(clusterResp["Engine"]))
	// d.Set("engine_version", fmt.Sprint(clusterResp["EngineVersion"]))
	// d.Set("vpc_id", fmt.Sprint(clusterResp["VpcId"]))
	// d.Set("zone_id", fmt.Sprint(clusterResp["ZoneId"]))
	// d.Set("region_id", fmt.Sprint(clusterResp["RegionId"]))

	// // Read elastic rules status
	// if _, exists := clusterResp["ScalingRulesEnable"]; exists {
	// 	d.Set("elastic_rules_enable", clusterResp["ScalingRulesEnable"].(bool))
	// }

	// // Read elastic rules if enabled
	// if d.Get("elastic_rules_enable").(bool) {
	// 	elasticRules, err := selectDBService.DescribeElasticRules(d.Id())
	// 	if err != nil {
	// 		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_selectdb_db_cluster_elastic_rules", AlibabaCloudSdkGoERROR)
	// 	}

	// 	rulesMapping := make([]map[string]interface{}, 0)
	// 	for _, rule := range elasticRules {
	// 		ruleItem := rule.(map[string]interface{})
	// 		ruleId := int(ruleItem["RuleId"].(float64))
	// 		rulesMapping = append(rulesMapping, map[string]interface{}{
	// 			"execution_period":        ruleItem["ExecutionPeriod"],
	// 			"elastic_rule_start_time": ruleItem["ElasticRuleStartTime"],
	// 			"cluster_class":           ruleItem["ClusterClass"],
	// 			"rule_id":                 ruleId,
	// 		})
	// 	}
	// 	d.Set("elastic_rules", rulesMapping)
	// }

	// configChangeArrayList, err := selectDBService.DescribeSelectDBClusterConfigChangeLog(d.Id())
	// if err != nil {
	// 	return WrapError(err)
	// }
	// configChangeArray := make([]map[string]interface{}, 0)
	// for _, v := range configChangeArrayList {
	// 	m1 := v.(map[string]interface{})
	// 	ConfigId, _ := m1["Id"].(json.Number).Int64()

	// 	temp1 := map[string]interface{}{
	// 		"name":         m1["Name"].(string),
	// 		"old_value":    m1["OldValue"].(string),
	// 		"new_value":    m1["NewValue"].(string),
	// 		"is_applied":   m1["IsApplied"].(bool),
	// 		"gmt_created":  m1["GmtCreated"].(string),
	// 		"gmt_modified": m1["GmtModified"].(string),
	// 		"config_id":    ConfigId,
	// 	}
	// 	configChangeArray = append(configChangeArray, temp1)
	// }
	// d.Set("param_change_logs", configChangeArray)
	return nil
}

func resourceAliCloudSelectDBClusterDelete(d *schema.ResourceData, meta interface{}) error {
	// client := meta.(*connectivity.AliyunClient)
	// selectDBService, err := NewSelectDBService(client)
	// if err != nil {
	// 	return WrapError(err)
	// }

	// stateConf := BuildStateConf([]string{"RESOURCE_PREPARING", "CLASS_CHANGING", "CREATING", "STOPPING", "STARTING", "RESTARTING", "RESTART", "MODIFY_PARAM"},
	// 	[]string{"ACTIVATION"}, d.Timeout(schema.TimeoutUpdate), 10*time.Second, selectDBService.SelectDBClusterStateRefreshFunc(d.Id(), []string{}))

	// if _, err := stateConf.WaitForState(); err != nil {
	// 	return WrapErrorf(err, IdMsg, d.Id())
	// }

	// _, err := selectDBService.DescribeSelectDBCluster(d.Id())
	// if err != nil {
	// 	if IsNotFoundError(err) {
	// 		return nil
	// 	}
	// 	return WrapError(err)
	// }

	// _, err = selectDBService.DeleteSelectDBCluster(d.Id())
	// if err != nil {
	// 	return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteDBCluster", AlibabaCloudSdkGoERROR)
	// }

	// instance_id := d.Get("db_instance_id").(string)
	// // cluster deleting cannot be checked, use instance from class changing to active instead.
	// // cluster deleting = related instance update
	// stateConf = BuildStateConf([]string{"CLASS_CHANGING"}, []string{"ACTIVATION"}, d.Timeout(schema.TimeoutDelete), 10*time.Second, selectDBService.SelectDBInstanceStateRefreshFunc(instance_id, []string{"DELETING"}))
	// if _, err := stateConf.WaitForState(); err != nil {
	// 	return WrapErrorf(err, IdMsg, d.Id())
	// }
	return nil
}

func buildSelectDBCreateClusterRequest(d *schema.ResourceData, meta interface{}) (map[string]interface{}, error) {
	// client := meta.(*connectivity.AliyunClient)
	// selectDBService, err := NewSelectDBService(client)
	// if err != nil {
	// 	return WrapError(err)
	// }

	// instanceResp, err := selectDBService.DescribeSelectDBInstance(d.Get("db_instance_id").(string))
	// if err != nil {
	// 	return nil, WrapErrorf(err, DefaultErrorMsg, d.Id())
	// }

	// vswitchId := ""
	// netResp, err := selectDBService.DescribeSelectDBInstanceNetInfo(d.Get("db_instance_id").(string))
	// if err != nil {
	// 	return nil, WrapErrorf(err, DefaultErrorMsg, d.Get("db_instance_id").(string))
	// }
	// resultClusterNet, _ := netResp["DBInstanceNetInfos"].([]interface{})
	// for _, v := range resultClusterNet {
	// 	item := v.(map[string]interface{})["VswitchId"].(string)
	// 	if item != "" {
	// 		vswitchId = item
	// 		break
	// 	}
	// }

	// cache_size, exist := d.GetOkExists("cache_size")
	// if !exist {
	// 	return nil, WrapErrorf(err, DefaultErrorMsg, d.Id())
	// }

	// request := map[string]interface{}{
	// 	"DBInstanceId":         d.Get("db_instance_id").(string),
	// 	"Engine":               "SelectDB",
	// 	"EngineVersion":        instanceResp["EngineVersion"],
	// 	"DBClusterClass":       d.Get("db_cluster_class").(string),
	// 	"RegionId":             client.RegionId,
	// 	"ZoneId":               instanceResp["ZoneId"],
	// 	"VpcId":                instanceResp["VpcId"],
	// 	"VSwitchId":            vswitchId,
	// 	"CacheSize":            cache_size.(int),
	// 	"DBClusterDescription": Trim(d.Get("db_cluster_description").(string)),
	// }

	// payType := convertPaymentTypeToChargeType(d.Get("payment_type"))

	// if payType == string(PostPaid) {
	// 	request["ChargeType"] = string("Postpaid")
	// } else if payType == string(PrePaid) {
	// 	period_time, _ := d.GetOkExists("period_time")
	// 	request["ChargeType"] = string("Prepaid")
	// 	request["Period"] = d.Get("period").(string)
	// 	request["UsedTime"] = strconv.Itoa(period_time.(int))
	// }

	return nil, nil
}

func convertSelectDBClusterStatusActionFinal(source string) string {
	action := ""
	switch source {
	case "STOPPING", "STOPPED":
		action = "STOPPED"
	case "STARTING":
		action = "ACTIVATION"
	case "RESTART", "RESTARTING":
		action = "ACTIVATION"
	}
	return action
}
