package alicloud

import (
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
			"instance_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      AliKafkaInstanceTypeReserved,
				ValidateFunc: StringInSlice([]string{AliKafkaInstanceTypeReserved, AliKafkaInstanceTypeServerless}, false),
			},
			"deploy_type": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: IntInSlice([]int{4, 5}),
			},
			"disk_size": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"disk_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"io_max_spec": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"spec_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "normal",
			},
			"partition_num": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				AtLeastOneOf: []string{"partition_num"},
			},
			"eip_max": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return d.Get("deploy_type").(int) == 5
				},
			},
			"paid_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      AliKafkaBillingTypePostPaid,
				ValidateFunc: StringInSlice([]string{"PrePaid", "PostPaid"}, false),
			},
			"duration": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"resource_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"tags": tagsSchemaForceNew(),

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
				ForceNew: true,
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
	instanceTypeInput := ""
	if v, ok := d.GetOk("instance_type"); ok {
		instanceTypeInput = v.(string)
	}
	paidTypeInput := ""
	if v, ok := d.GetOk("paid_type"); ok {
		paidTypeInput = v.(string)
	}

	instanceType, billingType, err := resolveAliKafkaInstanceBilling(instanceTypeInput, paidTypeInput)
	if err != nil {
		return WrapError(err)
	}

	if instanceType == kafka.InstanceSeriesReserved {
		if _, ok := d.GetOk("disk_size"); !ok {
			return WrapError(fmt.Errorf("disk_size is required for reserved instances"))
		}
		if _, ok := d.GetOk("disk_type"); !ok {
			return WrapError(fmt.Errorf("disk_type is required for reserved instances"))
		}
		if _, ok := d.GetOk("deploy_type"); !ok {
			return WrapError(fmt.Errorf("deploy_type is required for reserved instances"))
		}
	} else {
		if _, ok := d.GetOk("disk_size"); ok {
			return WrapError(fmt.Errorf("disk_size is not supported for serverless instances"))
		}
		if _, ok := d.GetOk("disk_type"); ok {
			return WrapError(fmt.Errorf("disk_type is not supported for serverless instances"))
		}
		if _, ok := d.GetOk("partition_num"); ok {
			return WrapError(fmt.Errorf("partition_num is not supported for serverless instances"))
		}
		if _, ok := d.GetOk("io_max_spec"); ok {
			return WrapError(fmt.Errorf("io_max_spec is not supported for serverless instances"))
		}
		if _, ok := d.GetOk("deploy_type"); ok {
			return WrapError(fmt.Errorf("deploy_type is not supported for serverless instances"))
		}
		if _, ok := d.GetOk("eip_max"); ok {
			return WrapError(fmt.Errorf("eip_max is not supported for serverless instances"))
		}
	}

	instance := &kafka.KafkaInstance{
		RegionId: client.RegionId,
	}

	if instanceType == kafka.InstanceSeriesReserved {
		if v, ok := d.GetOk("disk_type"); ok {
			diskTypeInt, _ := strconv.Atoi(v.(string))
			diskType := kafka.KafkaDiskType(diskTypeInt)
			instance.DiskType = &diskType
		}
		if v, ok := d.GetOk("disk_size"); ok {
			instance.DiskSize = tea.Int(v.(int))
		}
		if v, ok := d.GetOk("deploy_type"); ok {
			deployType := kafka.KafkaDeployType(v.(int))
			instance.DeployType = &deployType
		}
	}

	paidType := kafka.KafkaPaidTypePostPay
	if billingType == kafka.BillingTypePrePay {
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

	createConfig := buildAliKafkaInstanceCreationConfig(instance, instanceType, billingType)
	createResult, err := kafkaService.CreateAlikafkaInstance(createConfig)
	if err != nil {
		return WrapError(err)
	}

	if createResult == nil {
		return WrapError(fmt.Errorf("create instance result is nil"))
	}

	instanceId := createResult.InstanceId
	if instanceId == "" {
		return WrapError(fmt.Errorf("instance id is empty after creation"))
	}

	d.SetId(instanceId)

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
		diskType := *object.DiskType
		d.Set("disk_type", fmt.Sprint(diskType))
	}
	d.Set("disk_size", tea.IntValue(object.DiskSize))
	if object.DeployType != nil {
		d.Set("deploy_type", int(*object.DeployType))
	}
	d.Set("io_max", tea.IntValue(object.IoMax))
	d.Set("io_max_spec", tea.StringValue(object.IoMaxSpec))
	if v := tea.IntValue(object.EipMax); v != 0 {
		d.Set("eip_max", v)
	}
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
	d.Set("config", object.Config)

	d.Set("status", object.ServiceStatus) // ServiceStatus in VO (int)

	if billingType := FormatAliKafkaPaidType(object.PaidType); billingType != "" {
		d.Set("paid_type", billingType)
	}
	if instanceType := inferAliKafkaInstanceType(object); instanceType != "" {
		d.Set("instance_type", instanceType)
	}

	d.Set("kms_key_id", object.KmsKeyId)

	tags, err := kafkaService.DescribeTags(d.Id(), nil, TagResourceInstance)
	if err != nil {
		return WrapError(err)
	}

	d.Set("tags", kafkaService.tagsToMap(tags))

	return nil
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

	if err := kafkaService.WaitForAliKafkaInstanceDeleting(d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return WrapError(err)
	}

	return nil
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

func resolveAliKafkaInstanceBilling(instanceTypeInput, paidTypeInput string) (kafka.KafkaInstanceSeries, kafka.KafkaBillingType, error) {
	instanceType, err := ResolveAliKafkaInstanceType(instanceTypeInput)
	if err != nil {
		return "", "", err
	}
	billingType, err := ResolveAliKafkaPaidType(paidTypeInput)
	if err != nil {
		return "", "", err
	}
	if err := validateAliKafkaBillingCombination(instanceType, billingType); err != nil {
		return "", "", err
	}
	return instanceType, billingType, nil
}

func inferAliKafkaInstanceType(object *kafka.KafkaInstance) string {
	if object == nil {
		return ""
	}
	if object.DeployType != nil || object.DiskSize != nil || object.DiskType != nil || object.PartitionNum != nil || object.IoMaxSpec != nil {
		return AliKafkaInstanceTypeReserved
	}
	if object.SpecType != nil && *object.SpecType != "" {
		return AliKafkaInstanceTypeServerless
	}
	return AliKafkaInstanceTypeReserved
}
