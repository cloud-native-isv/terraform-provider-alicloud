package alicloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

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
			"vswitch_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"disk_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"disk_size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"deploy_type": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: IntInSlice([]int{4, 5}),
			},
			"partition_num": {
				Type:         schema.TypeInt,
				Optional:     true,
				AtLeastOneOf: []string{"partition_num"},
			},
			"io_max": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"io_max_spec": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"io_max", "io_max_spec"},
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"paid_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"spec_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"eip_max": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return d.Get("deploy_type").(int) == 5
				},
			},
			"resource_group_id": {
				Type:     schema.TypeString,
				Optional: true,
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
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"enable_auto_group": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"enable_auto_topic": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: StringInSlice([]string{"enable", "disable"}, false),
			},
			"default_topic_partition_num": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"vswitch_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"selected_zones": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeList,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				Description: "The JSON string of selected zones for the instance. Format: [[\"zone1\", \"zone2\"], [\"zone3\"]]",
			},
			"cross_zone": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Specifies whether to deploy the instance across zones. true: Deploy the instance across zones. false: Do not deploy the instance across zones. Default value: true.",
			},
			"tags": tagsSchema(),
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

	// 1. Create order
	order := &kafka.KafkaOrder{
		RegionId:   client.RegionId,
		DiskSize:   int32(d.Get("disk_size").(int)),
		DiskType:   d.Get("disk_type").(string),
		DeployType: kafka.KafkaDeployType(d.Get("deploy_type").(int)),
	}

	if v, ok := d.GetOk("partition_num"); ok {
		order.PartitionNum = int32(v.(int))
	}

	if v, ok := d.GetOk("io_max"); ok {
		order.IoMax = int32(v.(int))
	}

	if v, ok := d.GetOk("io_max_spec"); ok {
		order.IoMaxSpec = v.(string)
	}

	if v, ok := d.GetOk("spec_type"); ok {
		order.SpecType = kafka.KafkaSpecType(v.(string))
	}

	if v, ok := d.GetOkExists("eip_max"); ok {
		order.EipMax = int32(v.(int))
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		order.ResourceGroupId = v.(string)
	}

	var orderId string
	if v, ok := d.GetOk("paid_type"); ok {
		switch v.(string) {
		case "PostPaid":
			orderId, err = kafkaService.CreatePostPayOrder(order)
			if err != nil {
				return err
			}
			addDebug("CreatePostPayOrder", orderId, order)

		case "PrePaid":
			orderId, err = kafkaService.CreatePrePayOrder(order)
			if err != nil {
				return err
			}
			addDebug("CreatePrePayOrder", orderId, order)
		}
	}

	alikafkaInstanceVO, err := kafkaService.DescribeAlikafkaInstanceByOrderId(orderId, 60)
	if err != nil {
		return WrapError(err)
	}

	d.SetId(fmt.Sprint(alikafkaInstanceVO.InstanceId))

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
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_alikakfa_instance kafkaService.DescribeAlikafkaInstance Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set all schema fields using the correct field access
	d.Set("name", object.Name)
	d.Set("disk_type", object.DiskType)
	d.Set("disk_size", object.DiskSize)
	d.Set("deploy_type", object.DeployType)
	d.Set("io_max", object.IoMax)
	d.Set("io_max_spec", object.IoMaxSpec)
	d.Set("eip_max", object.EipMax)
	d.Set("resource_group_id", object.ResourceGroupId)
	d.Set("vpc_id", object.VpcId)
	d.Set("vswitch_id", object.VSwitchId)
	d.Set("zone_id", object.ZoneId)
	d.Set("paid_type", PostPaid)
	d.Set("spec_type", object.SpecType)
	d.Set("security_group", object.SecurityGroup)
	d.Set("end_point", object.EndPoint)
	d.Set("ssl_endpoint", object.SslEndPoint)
	d.Set("domain_endpoint", object.DomainEndpoint)
	d.Set("ssl_domain_endpoint", object.SslDomainEndpoint)
	d.Set("sasl_domain_endpoint", object.SaslDomainEndpoint)
	d.Set("status", object.ServiceStatus)
	d.Set("config", object.AllConfig)
	d.Set("kms_key_id", object.KmsKeyId)

	if object.PaidType == 0 {
		d.Set("paid_type", PrePaid)
	}

	tags, err := kafkaService.DescribeTags(d.Id(), nil, TagResourceInstance)
	if err != nil {
		return WrapError(err)
	}

	d.Set("tags", kafkaService.tagsToMap(tags))

	return nil
}

func resourceAliCloudAlikafkaInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}
	d.Partial(true)

	if err := kafkaService.setInstanceTags(d, TagResourceInstance); err != nil {
		return WrapError(err)
	}

	// Process change instance name.
	if !d.IsNewResource() && d.HasChange("name") {
		request := &ModifyInstanceNameRequest{
			RegionId:   client.RegionId,
			InstanceId: d.Id(),
		}

		if v, ok := d.GetOk("name"); ok {
			request.InstanceName = v.(string)
		}

		err = kafkaService.ModifyInstanceName(request)
		if err != nil {
			return err
		}
		addDebug("ModifyInstanceName", "Success", request)

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

	if !d.IsNewResource() && d.HasChange("io_max") {
		update = true
		if v, ok := d.GetOk("io_max"); ok {
			upgradeOrder.IoMax = int32(v.(int))
		}
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

	if update {
		var orderId string
		var err error
		if d.Get("paid_type").(string) == string(PrePaid) {
			orderId, err = kafkaService.UpgradePrePayOrder(upgradeOrder)
		} else {
			orderId, err = kafkaService.UpgradePostPayOrder(upgradeOrder)
		}

		if err != nil {
			return err
		}
		addDebug("UpgradeOrder", orderId, upgradeOrder)

		stateConf := BuildStateConf([]string{}, []string{fmt.Sprint(d.Get("disk_size"))}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, kafkaService.AliKafkaInstanceStateRefreshFunc(d.Id(), []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}

		if d.HasChange("io_max") {
			stateConf = BuildStateConf([]string{}, []string{fmt.Sprint(d.Get("io_max"))}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, kafkaService.AliKafkaInstanceStateRefreshFunc(d.Id(), []string{}))
			if _, err := stateConf.WaitForState(); err != nil {
				return WrapErrorf(err, IdMsg, d.Id())
			}
		}

		stateConf = BuildStateConf([]string{}, []string{fmt.Sprint(d.Get("eip_max"))}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, kafkaService.AliKafkaInstanceStateRefreshFunc(d.Id(), []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}

		stateConf = BuildStateConf([]string{}, []string{fmt.Sprint(d.Get("spec_type"))}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, kafkaService.AliKafkaInstanceStateRefreshFunc(d.Id(), []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}

		stateConf = BuildStateConf([]string{}, []string{"5"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, kafkaService.AliKafkaInstanceStateRefreshFunc(d.Id(), []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}

		d.SetPartial("partition_num")
		d.SetPartial("disk_size")
		d.SetPartial("io_max")
		d.SetPartial("io_max_spec")
		d.SetPartial("spec_type")
		d.SetPartial("eip_max")
	}

	if !d.IsNewResource() && d.HasChange("service_version") {
		request := &UpgradeInstanceVersionRequest{
			InstanceId: d.Id(),
			RegionId:   client.RegionId,
		}

		if v, ok := d.GetOk("service_version"); ok {
			request.TargetVersion = v.(string)
		}

		err = kafkaService.UpgradeInstanceVersion(request)
		if err != nil {
			return err
		}
		addDebug("UpgradeInstanceVersion", "Success", request)

		// wait for upgrade task be invoke
		time.Sleep(60 * time.Second)
		// upgrade service may be last a long time
		stateConf := BuildStateConf([]string{}, []string{"5"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, kafkaService.AliKafkaInstanceStateRefreshFunc(d.Id(), []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
		d.SetPartial("service_version")
	}

	if !d.IsNewResource() && d.HasChange("config") {
		request := map[string]interface{}{
			"RegionId":   client.RegionId,
			"InstanceId": d.Id(),
		}

		if v, ok := d.GetOk("config"); ok {
			request["Config"] = v
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

	// Instance delete is not implemented against Kafka API in this provider.
	// Keep a no-op delete to allow Terraform state removal without remote call.
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
