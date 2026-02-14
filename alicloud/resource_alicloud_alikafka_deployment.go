package alicloud

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudAlikafkaDeployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudAlikafkaDeploymentCreate,
		Read:   resourceAliCloudAlikafkaDeploymentRead,
		Update: resourceAliCloudAlikafkaDeploymentUpdate,
		Delete: resourceAliCloudAlikafkaDeploymentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"vswitch_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cross_zone": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     true,
				Description: "Specifies whether to deploy the instance across zones. true: Deploy the instance across zones. false: Do not deploy the instance across zones. Default value: true.",
			},
			"deploy_module": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "vpc",
				Description: "The deployment module of the instance. valid: vpc, eip",
			},
			"security_group": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"service_version": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"config": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"selected_zones": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeList,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				Description: "The JSON string of selected zones for the instance. Format: [[\"zone1\", \"zone2\"], [\"zone3\"]]",
			},
			"vswitch_ids": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"eip_max": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func formatSelectedZonesReq(configured []interface{}) (string, error) {
	if len(configured) != 2 {
		return "", fmt.Errorf("the selected_zones must be a two-dimensional array with exactly two elements")
	}

	var zones [][]string
	for _, v := range configured {
		list, ok := v.([]interface{})
		if !ok {
			return "", fmt.Errorf("the element of selected_zones must be a list")
		}

		items := []string{}
		for _, item := range list {
			str, ok := item.(string)
			if !ok {
				return "", fmt.Errorf("the element of inner zone list must be a string")
			}
			items = append(items, str)
		}
		zones = append(zones, items)
	}

	data, err := json.Marshal(zones)
	if err != nil {
		return "", fmt.Errorf("failed to marshal selected_zones: %s", err)
	}

	result := string(data)
	log.Printf("[DEBUG] formatSelectedZonesReq result: %s", result)
	return result, nil
}

func isJSONStringObjectSubset(subsetJSON, supersetJSON string) bool {
	var subset map[string]interface{}
	if err := json.Unmarshal([]byte(subsetJSON), &subset); err != nil {
		return false
	}

	var superset map[string]interface{}
	if err := json.Unmarshal([]byte(supersetJSON), &superset); err != nil {
		return false
	}

	for k, subsetVal := range subset {
		supersetVal, ok := superset[k]
		if !ok {
			return false
		}

		if !reflect.DeepEqual(subsetVal, supersetVal) {
			return false
		}
	}

	return true
}

func resourceAliCloudAlikafkaDeploymentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("instance_id").(string)
	vswitchId := d.Get("vswitch_id").(string)

	// Create options map for deployment
	options := make(map[string]interface{})

	if v, ok := d.GetOk("vpc_id"); ok {
		options["vpc_id"] = v.(string)
	}

	if v, ok := d.GetOk("zone_id"); ok {
		options["zone_id"] = v.(string)
	}

	if v, ok := d.GetOk("name"); ok {
		options["name"] = v.(string)
	}

	if v, ok := d.GetOk("cross_zone"); ok {
		options["cross_zone"] = v.(bool)
	}

	if v, ok := d.GetOk("deploy_module"); ok {
		options["deploy_module"] = v.(string)
	}

	if v, ok := d.GetOk("security_group"); ok {
		options["security_group"] = v.(string)
	}

	if v, ok := d.GetOk("service_version"); ok {
		options["service_version"] = v.(string)
	}

	if v, ok := d.GetOk("config"); ok {
		options["config"] = v.(string)
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		options["kms_key_id"] = v.(string)
	}

	if v, ok := d.GetOk("selected_zones"); ok {
		zonesStr, err := formatSelectedZonesReq(v.([]interface{}))
		if err != nil {
			return fmt.Errorf("formatting selected_zones failed: %s", err)
		}
		options["selected_zones"] = zonesStr
		log.Printf("[DEBUG] Resource alicloud_alikafka_deployment SelectedZones=%s", zonesStr)
	}

	if v, ok := d.GetOk("vswitch_ids"); ok {
		vswitchList := expandStringList(v.([]interface{}))
		if len(vswitchList) > 0 {
			options["vswitch_ids"] = vswitchList
		}
	}

	// Use CWS-Lib-Go API to start the instance
	config := kafka.StartInstanceConfig{
		InstanceId: instanceId,
		RegionId:   client.RegionId,
		VpcId:      options["vpc_id"].(string),
		VSwitchId:  vswitchId,
	}
	if v, ok := options["zone_id"].(string); ok {
		config.ZoneId = v
	}
	if v, ok := options["deploy_module"].(string); ok {
		config.DeployModule = v
	}
	if v, ok := options["is_eip_inner"].(bool); ok && v {
		config.IsEipInner = tea.Bool(v)
	}
	if v, ok := options["is_set_user_and_password"].(bool); ok && v {
		config.IsSetUserAndPassword = tea.Bool(v)
	}
	if v, ok := options["username"].(string); ok {
		config.Username = v
	}
	if v, ok := options["password"].(string); ok {
		config.Password = v
	}
	if v, ok := options["name"].(string); ok {
		config.Name = v
	}
	if v, ok := options["cross_zone"].(bool); ok && v {
		config.CrossZone = tea.Bool(v)
	}
	if v, ok := options["security_group"].(string); ok {
		config.SecurityGroup = v
	}
	if v, ok := options["service_version"].(string); ok {
		config.ServiceVersion = v
	}
	if v, ok := options["config"].(string); ok {
		config.Config = v
	}
	if v, ok := options["kms_key_id"].(string); ok {
		config.KMSKeyId = v
	}
	if v, ok := options["notifier"].(string); ok {
		config.Notifier = v
	}
	if v, ok := options["user_phone_num"].(string); ok {
		config.UserPhoneNum = v
	}
	if v, ok := options["selected_zones"].(string); ok {
		config.SelectedZones = v
	}
	if v, ok := options["is_force_selected_zones"].(bool); ok && v {
		config.IsForceSelectedZones = tea.Bool(v)
	}
	if v, ok := options["vswitch_ids"].([]string); ok {
		config.VSwitchIds = v
	}

	err = kafkaService.StartAlikafkaInstance(config)
	if err != nil {
		return WrapError(err)
	}

	addDebug("StartAlikafkaInstance", "Success", instanceId)

	d.SetId(instanceId)

	// Wait for deployment(start) to complete: instance should be running
	err = kafkaService.WaitForAliKafkaInstanceStarting(d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudAlikafkaDeploymentRead(d, meta)
}

func resourceAliCloudAlikafkaDeploymentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	object, err := kafkaService.DescribeAlikafkaInstance(d.Id())
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("instance_id", object.InstanceId)
	d.Set("name", tea.StringValue(object.Name))
	d.Set("vpc_id", object.VpcId)
	d.Set("vswitch_id", object.VSwitchId)
	if object.ZoneId != "" {
		d.Set("zone_id", object.ZoneId)
	}
	d.Set("security_group", object.SecurityGroup)

	stateConfig := d.Get("config").(string)
	remoteConfig := object.Config
	if remoteConfig != "" {
		// Some Kafka APIs return a subset of config keys and omit defaults.
		// Keep user/state config when remote values are only a subset to avoid false ForceNew drift.
		if stateConfig != "" && isJSONStringObjectSubset(remoteConfig, stateConfig) {
			d.Set("config", stateConfig)
		} else {
			d.Set("config", remoteConfig)
		}
	}

	if object.EipMax != nil {
		d.Set("eip_max", tea.IntValue(object.EipMax))
	}
	d.Set("kms_key_id", object.KmsKeyId)
	d.Set("vswitch_ids", []string{object.VSwitchId})

	return nil
}

func resourceAliCloudAlikafkaDeploymentUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	d.Partial(true)
	defer d.Partial(false)

	needWait := false

	if !d.IsNewResource() && d.HasChange("name") {
		name := d.Get("name").(string)
		if name == "" {
			return WrapError(fmt.Errorf("updating name failed: name cannot be empty"))
		}

		req := &ModifyInstanceNameRequest{
			RegionId:     client.RegionId,
			InstanceId:   d.Id(),
			InstanceName: name,
		}

		err = kafkaService.ModifyAlikafkaInstanceName(req)
		if err != nil {
			return WrapError(err)
		}

		addDebug("ModifyAlikafkaInstanceName", "Success", name)
		d.SetPartial("name")
		needWait = true
	}

	if !d.IsNewResource() && d.HasChange("config") {
		configStr := d.Get("config").(string)
		apiConfig := make(map[string]*string)

		if configStr != "" {
			var configMap map[string]interface{}
			if err := json.Unmarshal([]byte(configStr), &configMap); err != nil {
				return WrapError(fmt.Errorf("failed to unmarshal config: %v", err))
			}

			for k, val := range configMap {
				s := fmt.Sprintf("%v", val)
				apiConfig[k] = &s
			}
		}

		err = kafkaService.UpdateInstanceConfig(d.Id(), apiConfig)
		if err != nil {
			return WrapError(err)
		}

		addDebug("UpdateInstanceConfig", "Success", apiConfig)
		d.SetPartial("config")
		needWait = true
	}

	if needWait {
		err = kafkaService.WaitForAliKafkaInstanceUpdating(d.Id(), d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	return resourceAliCloudAlikafkaDeploymentRead(d, meta)
}

func resourceAliCloudAlikafkaDeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Id()

	// Use CWS-Lib-Go API to stop the instance
	err = kafkaService.StopAlikafkaInstance(&StopInstanceRequest{
		InstanceId: instanceId,
	})
	if err != nil {
		return WrapError(err)
	}

	// Optionally wait for instance to stop using the new wait function
	err = kafkaService.WaitForAliKafkaInstanceStopping(instanceId, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		// If wait fails, just log the error but don't fail the delete operation
		log.Printf("[WARN] Failed to wait for instance to stop: %v", err)
	}

	return nil
}
