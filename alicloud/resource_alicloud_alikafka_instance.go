package alicloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudAlikafkaInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudAlikafkaInstanceCreate,
		Read:   resourceAliCloudAlikafkaInstanceRead,
		Update: resourceAliCloudAlikafkaInstanceUpdate,
		Delete: resourceAliCloudAlikafkaInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"deploy_type": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: IntInSlice([]int{4, 5}),
			},
			"disk_size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"disk_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"io_max_spec": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"spec_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "normal",
			},
			"partition_num": {
				Type:         schema.TypeInt,
				Optional:     true,
				AtLeastOneOf: []string{"partition_num"},
			},
			"eip_max": {
				Type:     schema.TypeInt,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return d.Get("deploy_type").(int) == 5
				},
			},
			"paid_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: StringInSlice([]string{"PrePaid", "PostPaid"}, false),
			},
			"duration": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"resource_group_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchema(),

			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_group": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"config": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vswitch_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_auto_group": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enable_auto_topic": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_topic_partition_num": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"cross_zone": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"end_point": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssl_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ssl_domain_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sasl_domain_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"topic_num_of_buy": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"topic_used": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"topic_left": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"partition_used": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"partition_left": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"group_used": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"group_left": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"is_partition_buy": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAliCloudAlikafkaInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	// Create instance directly using CWS-Lib-Go API
	diskTypeInt, _ := strconv.Atoi(d.Get("disk_type").(string))
	diskType := kafka.KafkaDiskType(diskTypeInt)
	deployType := kafka.KafkaDeployType(d.Get("deploy_type").(int))

	instance := &kafka.KafkaInstance{
		RegionId:   client.RegionId,
		DiskSize:   tea.Int(d.Get("disk_size").(int)),
		DiskType:   &diskType,
		DeployType: &deployType,
	}

	paidType := kafka.KafkaPaidTypePostPay
	if v, ok := d.GetOk("paid_type"); ok && v.(string) == "PrePaid" {
		paidType = kafka.KafkaPaidTypePrePay
	}
	instance.PaidType = &paidType

	if v, ok := d.GetOk("partition_num"); ok {
		instance.PartitionNum = tea.Int(v.(int))
	}

	if v, ok := d.GetOk("io_max_spec"); ok {
		instance.IoMaxSpec = tea.String(v.(string))
	}

	if v, ok := d.GetOk("spec_type"); ok {
		instance.SpecType = tea.String(v.(string))
	}

	if v, ok := d.GetOkExists("eip_max"); ok {
		instance.EipMax = tea.Int(v.(int))
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		instance.ResourceGroupId = v.(string)
	}

	if v, ok := d.GetOk("duration"); ok {
		instance.Duration = tea.Int(v.(int))
	}

	if _, ok := d.GetOk("tags"); ok {
		instance.Tags = extractTags(d)
	}

	createdInstance, err := kafkaService.CreateAlikafkaInstance(instance)
	if err != nil {
		return WrapError(err)
	}

	d.SetId(createdInstance.InstanceId)

	// Wait for instance to be in running state (state 5)
	err = kafkaService.WaitForAliKafkaInstanceCreating(d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudAlikafkaInstanceRead(d, meta)
}

func resourceAliCloudAlikafkaInstanceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	object, err := kafkaService.DescribeAlikafkaInstance(d.Id())
	if err != nil {
		// Handle exceptions
		// Note: kafka.NewKafkaError produces error that might not satisfy NotFoundError directly unless unwrapped or checked via strings
		// Assuming NotFoundError handles it or we check message
		if !d.IsNewResource() && (NotFoundError(err) || strings.Contains(err.Error(), "not found")) {
			log.Printf("[DEBUG] Resource alicloud_alikakfa_instance kafkaService.DescribeAlikafkaInstance Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("name", tea.StringValue(object.Name)) // object.Name is *string
	if object.DiskType != nil {
		d.Set("disk_type", fmt.Sprint(*object.DiskType)) // int -> string
	}
	d.Set("disk_size", tea.IntValue(object.DiskSize))
	if object.DeployType != nil {
		d.Set("deploy_type", int(*object.DeployType))
	}
	d.Set("io_max", tea.IntValue(object.IoMax))
	d.Set("io_max_spec", tea.StringValue(object.IoMaxSpec))
	d.Set("eip_max", tea.IntValue(object.EipMax))
	d.Set("resource_group_id", object.ResourceGroupId)
	d.Set("vpc_id", object.VpcId)
	d.Set("vswitch_id", object.VSwitchId)
	d.Set("zone_id", object.ZoneId)
	d.Set("spec_type", tea.StringValue(object.SpecType))
	d.Set("security_group", object.SecurityGroup)
	d.Set("end_point", object.EndPoint)
	d.Set("domain_endpoint", object.DomainEndpoint)
	d.Set("ssl_endpoint", object.SslEndPoint) // Field in VO is SslEndPoint
	d.Set("ssl_domain_endpoint", object.SslDomainEndpoint)
	d.Set("sasl_domain_endpoint", object.SaslDomainEndpoint)
	// d.Set("service_version", object.ServiceVersion) // Missing in VO
	d.Set("config", object.Config)

	d.Set("status", object.ServiceStatus) // ServiceStatus in VO (int)

	d.Set("kms_key_id", object.KmsKeyId)

	// Set service version and other fields if available
	// if object.Version != "" {
	// 	d.Set("service_version", object.Version)
	// }

	tags, err := kafkaService.DescribeTags(d.Id(), nil, TagResourceInstance)
	if err != nil {
		return WrapError(err)
	}

	d.Set("tags", kafkaService.tagsToMap(tags))
	// if err != nil {
	// 	return WrapError(err)
	// }

	// d.Set("tags", kafkaService.tagsToMap(tags))
	d.Set("tags", object.Tags)

	return nil
}

func resourceAliCloudAlikafkaInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}
	d.Partial(true)

	// if err := kafkaService.setInstanceTags(d, TagResourceInstance); err != nil {
	// 	return WrapError(err)
	// }

	// Process change instance name.
	if !d.IsNewResource() && d.HasChange("name") {
		instance := &kafka.KafkaInstance{
			InstanceId: d.Id(),
		}

		if v, ok := d.GetOk("name"); ok {
			instance.Name = tea.String(v.(string))
		}

		err = kafkaService.UpdateAlikafkaInstance(instance)
		if err != nil {
			return WrapError(err)
		}
		addDebug("UpdateAlikafkaInstance", "Success", instance)

		d.SetPartial("name")
	}

	// Process paid type change, note only support change from post to pre pay.
	if !d.IsNewResource() && d.HasChange("paid_type") {
		o, n := d.GetChange("paid_type")
		oldPaidType := o.(string)
		newPaidType := n.(string)
		oldPaidTypeInt := 1
		newPaidTypeInt := 1
		if oldPaidType == string(PrePaid) {
			oldPaidTypeInt = 0
		}
		if newPaidType == string(PrePaid) {
			newPaidTypeInt = 0
		}
		if oldPaidTypeInt == 1 && newPaidTypeInt == 0 {
			return WrapError(errors.New("paid type conversion from post pay to pre pay is not supported in current implementation"))
		} else {
			return WrapError(errors.New("paid type only support change from post pay to pre pay"))
		}
	}

	update := false
	upgradeOrder := &kafka.KafkaOrder{
		InstanceId: d.Id(),
		RegionId:   client.RegionId,
	}

	// updating topic_quota only by updating partition_num
	if !d.IsNewResource() && d.HasChange("partition_num") {
		update = true
	}
	if v, ok := d.GetOk("partition_num"); ok {
		upgradeOrder.PartitionNum = int32(v.(int))
	}

	if !d.IsNewResource() && d.HasChange("disk_size") {
		update = true
	}
	if v, ok := d.GetOk("disk_size"); ok {
		upgradeOrder.DiskSize = int32(v.(int))
	}

	if !d.IsNewResource() && d.HasChange("io_max_spec") {
		update = true
		if v, ok := d.GetOk("io_max_spec"); ok {
			upgradeOrder.IoMaxSpec = v.(string)
		}
	}

	if !d.IsNewResource() && d.HasChange("spec_type") {
		update = true
	}
	if v, ok := d.GetOk("spec_type"); ok {
		upgradeOrder.SpecType = kafka.KafkaSpecType(v.(string))
	}

	if !d.IsNewResource() && d.HasChange("deploy_type") {
		update = true
	}
	if d.Get("deploy_type").(int) == 4 {
		upgradeOrder.EipModel = true
	} else {
		upgradeOrder.EipModel = false
	}

	if !d.IsNewResource() && d.HasChange("eip_max") {
		update = true
	}
	if v, ok := d.GetOk("eip_max"); ok {
		upgradeOrder.EipMax = int32(v.(int))
	}

	if !d.IsNewResource() && d.HasChange("duration") {
		update = true
	}
	if v, ok := d.GetOk("duration"); ok {
		upgradeOrder.Duration = int32(v.(int))
	}

	if update {
		// Update instance directly using CWS-Lib-Go API
		instance := &kafka.KafkaInstance{
			InstanceId: d.Id(),
			RegionId:   client.RegionId,
		}

		if v, ok := d.GetOk("partition_num"); ok {
			instance.PartitionNum = tea.Int(v.(int))
		}

		if v, ok := d.GetOk("disk_size"); ok {
			instance.DiskSize = tea.Int(v.(int))
		}

		if v, ok := d.GetOk("io_max_spec"); ok {
			instance.IoMaxSpec = tea.String(v.(string))
		}

		if v, ok := d.GetOk("spec_type"); ok {
			instance.SpecType = tea.String(v.(string))
		}

		if d.Get("deploy_type").(int) == 4 {
			instance.EipModel = true
		} else {
			instance.EipModel = false
		}

		if v, ok := d.GetOk("eip_max"); ok {
			instance.EipMax = tea.Int(v.(int))
		}

		if v, ok := d.GetOk("duration"); ok {
			instance.Duration = tea.Int(v.(int))
		}

		err = kafkaService.UpgradeAlikafkaInstance(instance)
		if err != nil {
			return WrapError(err)
		}

		addDebug("UpgradeAlikafkaInstance", "Success", instance)

		// Wait for update to complete using the new wait function
		err = kafkaService.WaitForAliKafkaInstanceUpdating(d.Id(), d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}

		d.SetPartial("partition_num")
		d.SetPartial("disk_size")
		d.SetPartial("io_max_spec")
		d.SetPartial("spec_type")
		d.SetPartial("eip_max")
	}

	if !d.IsNewResource() && d.HasChange("service_version") {
		if v, ok := d.GetOk("service_version"); ok {
			serviceVersion := v.(string)

			req := &UpgradeInstanceVersionRequest{
				RegionId:      client.RegionId,
				InstanceId:    d.Id(),
				TargetVersion: serviceVersion,
			}
			err = kafkaService.UpgradeInstanceVersion(req)
			if err != nil {
				return WrapError(err)
			}
			addDebug("UpgradeInstanceVersion", "Success", serviceVersion)

			// wait for upgrade task to be invoked
			time.Sleep(60 * time.Second)

			// Wait for instance to complete upgrade using the new wait function
			err = kafkaService.WaitForAliKafkaInstanceUpdating(d.Id(), d.Timeout(schema.TimeoutUpdate))
			if err != nil {
				return WrapErrorf(err, IdMsg, d.Id())
			}
			d.SetPartial("service_version")
		}
	}

	if !d.IsNewResource() && d.HasChange("config") {
		if v, ok := d.GetOk("config"); ok {
			var configMap map[string]interface{}
			if err := json.Unmarshal([]byte(v.(string)), &configMap); err != nil {
				return WrapError(fmt.Errorf("failed to unmarshal config: %v", err))
			}

			apiConfig := make(map[string]*string)
			for k, val := range configMap {
				s := fmt.Sprintf("%v", val)
				apiConfig[k] = &s
			}

			err = kafkaService.UpdateInstanceConfig(d.Id(), apiConfig)
			if err != nil {
				return err
			}
			addDebug("UpdateInstanceConfig", "Success", apiConfig)
		}
		d.SetPartial("config")
	}

	update = false
	changeResourceGroupReq := map[string]interface{}{
		"RegionId":   client.RegionId,
		"ResourceId": d.Id(),
	}

	if !d.IsNewResource() && d.HasChange("resource_group_id") {
		update = true
	}
	if v, ok := d.GetOk("resource_group_id"); ok {
		changeResourceGroupReq["NewResourceGroupId"] = v
	}

	if update {
		d.SetPartial("resource_group_id")
	}

	update = false
	enableAutoGroupCreationReq := map[string]interface{}{
		"RegionId":   client.RegionId,
		"InstanceId": d.Id(),
	}

	if d.HasChange("enable_auto_group") {
		update = true

		if v, ok := d.GetOkExists("enable_auto_group"); ok {
			enableAutoGroupCreationReq["Enable"] = v
		}
	}

	if update {
		d.SetPartial("enable_auto_group")
	}

	update = false
	enableAutoTopicCreationReq := map[string]interface{}{
		"RegionId":   client.RegionId,
		"InstanceId": d.Id(),
	}

	if d.HasChange("enable_auto_topic") {
		update = true
	}
	if v, ok := d.GetOk("enable_auto_topic"); ok {
		enableAutoTopicCreationReq["Operate"] = v
	}

	if update {
		d.SetPartial("enable_auto_topic")
	}

	d.Partial(false)

	return resourceAliCloudAlikafkaInstanceRead(d, meta)
}

func resourceAliCloudAlikafkaInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	// Pre paid instance can not be release.
	if d.Get("paid_type").(string) == string(PrePaid) {
		return nil
	}

	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	if err := kafkaService.DeleteAlikafkaInstance(d.Id()); err != nil {
		return WrapError(err)
	}

	if err := kafkaService.WaitForAlikafkaInstance(d.Id(), Deleted, int(d.Timeout(schema.TimeoutDelete).Seconds())); err != nil {
		return WrapError(err)
	}

	return nil
}

func formatSelectedZonesReq(configured []interface{}) string {
	if len(configured) == 0 {
		return ""
	}

	var zones [][]string
	for _, item := range configured {
		if innerList, ok := item.([]interface{}); ok {
			var innerZones []string
			for _, z := range innerList {
				if s, ok := z.(string); ok {
					innerZones = append(innerZones, s)
				}
			}
			zones = append(zones, innerZones)
		}
	}

	// 使用json.Marshal进行序列化
	jsonBytes, err := json.Marshal(zones)
	if err != nil {
		// 如果序列化失败，返回空字符串
		return ""
	}

	return string(jsonBytes)
}

func convertAliKafkaAutoCreateTopicEnableResponse(source interface{}) interface{} {
	switch source {
	case true:
		return "enable"
	case false:
		return "disable"
	}

	return source
}

func extractTags(d *schema.ResourceData) map[string]string {
	tags := make(map[string]string)
	if v, ok := d.GetOk("tags"); ok {
		for k, v := range v.(map[string]interface{}) {
			tags[k] = v.(string)
		}
	}
	return tags
}
