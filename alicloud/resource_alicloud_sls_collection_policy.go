// Package alicloud. This file is generated automatically. Please do not modify it manually, thank you!
package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudSlsCollectionPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudSlsCollectionPolicyCreate,
		Read:   resourceAliCloudSlsCollectionPolicyRead,
		Update: resourceAliCloudSlsCollectionPolicyUpdate,
		Delete: resourceAliCloudSlsCollectionPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"centralize_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dest_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"dest_logstore": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dest_region": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dest_project": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"centralize_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"data_code": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"data_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_region": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"data_project": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"policy_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_tags": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"regions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"instance_ids": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"resource_mode": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: StringInSlice([]string{"all", "instanceMode", "attributeMode"}, false),
						},
					},
				},
			},
			"policy_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"product_code": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_directory": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_group_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"members": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func resourceAliCloudSlsCollectionPolicyCreate(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*connectivity.AliyunClient)

	action := fmt.Sprintf("/collectionpolicy")
	var request map[string]interface{}
	var response map[string]interface{}
	query := make(map[string]*string)
	body := make(map[string]interface{})
	hostMap := make(map[string]*string)
	var err error
	request = make(map[string]interface{})
	if v, ok := d.GetOk("policy_name"); ok {
		request["policyName"] = v
	}

	if v, ok := d.GetOkExists("centralize_enabled"); ok {
		request["centralizeEnabled"] = v
	}
	request["enabled"] = d.Get("enabled")
	request["productCode"] = d.Get("product_code")
	request["dataCode"] = d.Get("data_code")
	objectDataLocalMap := make(map[string]interface{})

	if v := d.Get("policy_config"); v != nil {
		resourceMode1, _ := jsonpath.Get("$[0].resource_mode", d.Get("policy_config"))
		if resourceMode1 != nil && resourceMode1 != "" {
			objectDataLocalMap["resourceMode"] = resourceMode1
		}
		resourceTags1, _ := jsonpath.Get("$[0].resource_tags", d.Get("policy_config"))
		if resourceTags1 != nil && resourceTags1 != "" {
			objectDataLocalMap["resourceTags"] = resourceTags1
		}
		regions1, _ := jsonpath.Get("$[0].regions", v)
		if regions1 != nil && regions1 != "" {
			objectDataLocalMap["regions"] = regions1
		}
		instanceIds1, _ := jsonpath.Get("$[0].instance_ids", v)
		if instanceIds1 != nil && instanceIds1 != "" {
			objectDataLocalMap["instanceIds"] = instanceIds1
		}

		request["policyConfig"] = objectDataLocalMap
	}

	objectDataLocalMap1 := make(map[string]interface{})

	if v := d.Get("centralize_config"); !IsNil(v) {
		destRegion1, _ := jsonpath.Get("$[0].dest_region", d.Get("centralize_config"))
		if destRegion1 != nil && destRegion1 != "" {
			objectDataLocalMap1["destRegion"] = destRegion1
		}
		destProject1, _ := jsonpath.Get("$[0].dest_project", d.Get("centralize_config"))
		if destProject1 != nil && destProject1 != "" {
			objectDataLocalMap1["destProject"] = destProject1
		}
		destLogstore1, _ := jsonpath.Get("$[0].dest_logstore", d.Get("centralize_config"))
		if destLogstore1 != nil && destLogstore1 != "" {
			objectDataLocalMap1["destLogstore"] = destLogstore1
		}
		destTtl, _ := jsonpath.Get("$[0].dest_ttl", d.Get("centralize_config"))
		if destTtl != nil && destTtl != "" {
			objectDataLocalMap1["destTTL"] = destTtl
		}

		request["centralizeConfig"] = objectDataLocalMap1
	}

	objectDataLocalMap2 := make(map[string]interface{})

	if v := d.Get("data_config"); !IsNil(v) {
		dataRegion1, _ := jsonpath.Get("$[0].data_region", d.Get("data_config"))
		if dataRegion1 != nil && dataRegion1 != "" {
			objectDataLocalMap2["dataRegion"] = dataRegion1
		}

		request["dataConfig"] = objectDataLocalMap2
	}

	objectDataLocalMap3 := make(map[string]interface{})

	if v := d.Get("resource_directory"); !IsNil(v) {
		accountGroupType1, _ := jsonpath.Get("$[0].account_group_type", d.Get("resource_directory"))
		if accountGroupType1 != nil && accountGroupType1 != "" {
			objectDataLocalMap3["accountGroupType"] = accountGroupType1
		}
		members1, _ := jsonpath.Get("$[0].members", v)
		if members1 != nil && members1 != "" {
			objectDataLocalMap3["members"] = members1
		}

		request["resourceDirectory"] = objectDataLocalMap3
	}

	body = request
	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err = client.Do("Sls", roaParam("POST", "2020-12-30", "UpsertCollectionPolicy", action), query, body, nil, hostMap, false)
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
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_collection_policy", action, AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprint(request["policyName"]))

	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_collection_policy", "NewSlsService", AlibabaCloudSdkGoERROR)
	}
	stateConf := BuildStateConf([]string{}, []string{"#CHECKSET"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, slsService.SlsCollectionPolicyStateRefreshFunc(d.Id(), []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudSlsCollectionPolicyRead(d, meta)
}

func resourceAliCloudSlsCollectionPolicyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_collection_policy", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	policy, err := slsService.DescribeSlsCollectionPolicy(d.Id())
	if err != nil {
		if !d.IsNewResource() && IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_sls_collection_policy DescribeSlsCollectionPolicy Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set basic attributes
	d.Set("centralize_enabled", policy.CentralizeEnabled)
	d.Set("data_code", policy.DataCode)
	d.Set("enabled", policy.Enabled)
	d.Set("product_code", policy.ProductCode)
	d.Set("policy_name", policy.PolicyName)

	// Handle centralize config
	centralizeConfigMaps := make([]map[string]interface{}, 0)
	if policy.CentralizeConfig != nil {
		centralizeConfigMap := make(map[string]interface{})
		centralizeConfigMap["dest_logstore"] = policy.CentralizeConfig.DestLogstore
		centralizeConfigMap["dest_project"] = policy.CentralizeConfig.DestProject
		centralizeConfigMap["dest_region"] = policy.CentralizeConfig.DestRegion
		centralizeConfigMap["dest_ttl"] = policy.CentralizeConfig.DestTTL
		centralizeConfigMaps = append(centralizeConfigMaps, centralizeConfigMap)

		if err := d.Set("centralize_config", centralizeConfigMaps); err != nil {
			return err
		}
	}

	// Handle data config
	dataConfigMaps := make([]map[string]interface{}, 0)
	if policy.DataConfig != nil {
		dataConfigMap := make(map[string]interface{})
		dataConfigMap["data_project"] = policy.DataConfig.DataProject
		dataConfigMap["data_region"] = policy.DataConfig.DataRegion
		dataConfigMaps = append(dataConfigMaps, dataConfigMap)

		if err := d.Set("data_config", dataConfigMaps); err != nil {
			return err
		}
	}

	// Handle policy config
	policyConfigMaps := make([]map[string]interface{}, 0)
	if policy.PolicyConfig != nil {
		policyConfigMap := make(map[string]interface{})
		policyConfigMap["resource_mode"] = policy.PolicyConfig.ResourceMode
		policyConfigMap["resource_tags"] = policy.PolicyConfig.ResourceTags
		policyConfigMap["instance_ids"] = policy.PolicyConfig.InstanceIds
		policyConfigMap["regions"] = policy.PolicyConfig.Regions
		policyConfigMaps = append(policyConfigMaps, policyConfigMap)

		if err := d.Set("policy_config", policyConfigMaps); err != nil {
			return err
		}
	}

	// Handle resource directory
	resourceDirectoryMaps := make([]map[string]interface{}, 0)
	if policy.ResourceDirectory != nil && policy.ResourceDirectory.AccountGroupType != "" {
		resourceDirectoryMap := make(map[string]interface{})
		resourceDirectoryMap["account_group_type"] = policy.ResourceDirectory.AccountGroupType
		resourceDirectoryMap["members"] = policy.ResourceDirectory.Members
		resourceDirectoryMaps = append(resourceDirectoryMaps, resourceDirectoryMap)

		if err := d.Set("resource_directory", resourceDirectoryMaps); err != nil {
			return err
		}
	}

	d.Set("policy_name", d.Id())

	return nil
}

func resourceAliCloudSlsCollectionPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_collection_policy", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	var request map[string]interface{}
	var response map[string]interface{}
	var query map[string]*string
	var body map[string]interface{}
	update := false
	action := fmt.Sprintf("/collectionpolicy")
	request = make(map[string]interface{})
	query = make(map[string]*string)
	body = make(map[string]interface{})
	hostMap := make(map[string]*string)
	request["policyName"] = d.Id()

	if d.HasChange("centralize_enabled") {
		update = true
	}
	if v, ok := d.GetOkExists("centralize_enabled"); ok || d.HasChange("centralize_enabled") {
		request["centralizeEnabled"] = v
	}
	if d.HasChange("enabled") {
		update = true
	}
	request["enabled"] = d.Get("enabled")
	if d.HasChange("product_code") {
		update = true
	}
	request["productCode"] = d.Get("product_code")
	if d.HasChange("data_code") {
		update = true
	}
	request["dataCode"] = d.Get("data_code")
	if d.HasChange("policy_config") {
		update = true
	}
	objectDataLocalMap := make(map[string]interface{})

	if v := d.Get("policy_config"); v != nil {
		resourceMode1, _ := jsonpath.Get("$[0].resource_mode", v)
		if resourceMode1 != nil && (d.HasChange("policy_config.0.resource_mode") || resourceMode1 != "") {
			objectDataLocalMap["resourceMode"] = resourceMode1
		}
		resourceTags1, _ := jsonpath.Get("$[0].resource_tags", v)
		if resourceTags1 != nil && (d.HasChange("policy_config.0.resource_tags") || resourceTags1 != "") {
			objectDataLocalMap["resourceTags"] = resourceTags1
		}
		regions1, _ := jsonpath.Get("$[0].regions", d.Get("policy_config"))
		if regions1 != nil && (d.HasChange("policy_config.0.regions") || regions1 != "") {
			objectDataLocalMap["regions"] = regions1
		}
		instanceIds1, _ := jsonpath.Get("$[0].instance_ids", d.Get("policy_config"))
		if instanceIds1 != nil && (d.HasChange("policy_config.0.instance_ids") || instanceIds1 != "") {
			objectDataLocalMap["instanceIds"] = instanceIds1
		}

		request["policyConfig"] = objectDataLocalMap
	}

	if d.HasChange("centralize_config") {
		update = true
	}
	objectDataLocalMap1 := make(map[string]interface{})

	if v := d.Get("centralize_config"); !IsNil(v) || d.HasChange("centralize_config") {
		destRegion1, _ := jsonpath.Get("$[0].dest_region", v)
		if destRegion1 != nil && (d.HasChange("centralize_config.0.dest_region") || destRegion1 != "") {
			objectDataLocalMap1["destRegion"] = destRegion1
		}
		destProject1, _ := jsonpath.Get("$[0].dest_project", v)
		if destProject1 != nil && (d.HasChange("centralize_config.0.dest_project") || destProject1 != "") {
			objectDataLocalMap1["destProject"] = destProject1
		}
		destLogstore1, _ := jsonpath.Get("$[0].dest_logstore", v)
		if destLogstore1 != nil && (d.HasChange("centralize_config.0.dest_logstore") || destLogstore1 != "") {
			objectDataLocalMap1["destLogstore"] = destLogstore1
		}
		destTtl, _ := jsonpath.Get("$[0].dest_ttl", v)
		if destTtl != nil && (d.HasChange("centralize_config.0.dest_ttl") || destTtl != "") {
			objectDataLocalMap1["destTTL"] = destTtl
		}

		request["centralizeConfig"] = objectDataLocalMap1
	}

	if d.HasChange("resource_directory") {
		update = true
	}
	objectDataLocalMap2 := make(map[string]interface{})

	if v := d.Get("resource_directory"); !IsNil(v) || d.HasChange("resource_directory") {
		accountGroupType1, _ := jsonpath.Get("$[0].account_group_type", v)
		if accountGroupType1 != nil && (d.HasChange("resource_directory.0.account_group_type") || accountGroupType1 != "") {
			objectDataLocalMap2["accountGroupType"] = accountGroupType1
		}
		members1, _ := jsonpath.Get("$[0].members", d.Get("resource_directory"))
		if members1 != nil && (d.HasChange("resource_directory.0.members") || members1 != "") {
			objectDataLocalMap2["members"] = members1
		}

		request["resourceDirectory"] = objectDataLocalMap2
	}

	if d.HasChange("data_config") {
		update = true
	}
	objectDataLocalMap3 := make(map[string]interface{})

	if v := d.Get("data_config"); !IsNil(v) || d.HasChange("data_config") {
		dataRegion1, _ := jsonpath.Get("$[0].data_region", v)
		if dataRegion1 != nil && (d.HasChange("data_config.0.data_region") || dataRegion1 != "") {
			objectDataLocalMap3["dataRegion"] = dataRegion1
		}

		request["dataConfig"] = objectDataLocalMap3
	}

	body = request
	if update {
		wait := incrementalWait(3*time.Second, 5*time.Second)
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			response, err = client.Do("Sls", roaParam("POST", "2020-12-30", "UpsertCollectionPolicy", action), query, body, nil, hostMap, false)
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
		stateConf := BuildStateConf([]string{}, []string{"#CHECKSET"}, d.Timeout(schema.TimeoutUpdate), 10*time.Second, slsService.SlsCollectionPolicyStateRefreshFunc(d.Id(), []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	return resourceAliCloudSlsCollectionPolicyRead(d, meta)
}

func resourceAliCloudSlsCollectionPolicyDelete(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_collection_policy", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	policyName := d.Id()
	action := fmt.Sprintf("/collectionpolicy/%s", policyName)
	var request map[string]interface{}
	var response map[string]interface{}
	query := make(map[string]*string)
	hostMap := make(map[string]*string)
	request = make(map[string]interface{})
	request["policyName"] = d.Id()

	if v, ok := d.GetOk("product_code"); ok {
		query["productCode"] = StringPointer(v.(string))
	}

	if v, ok := d.GetOk("data_code"); ok {
		query["dataCode"] = StringPointer(v.(string))
	}

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		response, err = client.Do("Sls", roaParam("DELETE", "2020-12-30", "DeleteCollectionPolicy", action), query, nil, nil, hostMap, false)
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
		if IsExpectedErrors(err, []string{"PolicyNotExist"}) || IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
	}

	stateConf := BuildStateConf([]string{}, []string{}, d.Timeout(schema.TimeoutDelete), 5*time.Second, slsService.SlsCollectionPolicyStateRefreshFunc(d.Id(), []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
