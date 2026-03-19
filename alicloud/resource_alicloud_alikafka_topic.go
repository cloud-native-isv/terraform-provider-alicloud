package alicloud

import (
	"errors"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudAlikafkaTopic() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudAlikafkaTopicCreate,
		Update: resourceAliCloudAlikafkaTopicUpdate,
		Read:   resourceAliCloudAlikafkaTopicRead,
		Delete: resourceAliCloudAlikafkaTopicDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"topic": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: StringLenBetween(1, 249),
			},
			"local_topic": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"compact_topic": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return !d.Get("local_topic").(bool)
				},
			},
			"replica_num": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      3,
				ValidateFunc: IntBetween(1, 3),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return !d.Get("local_topic").(bool)
				},
			},
			"partition_num": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      12,
				ValidateFunc: IntBetween(0, 360),
			},
			"remark": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: StringLenBetween(1, 64),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAliCloudAlikafkaTopicCreate(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("instance_id").(string)
	topicName := d.Get("topic").(string)
	localTopic := d.Get("local_topic").(bool)
	resourceId := EncodeTopicId(instanceId, topicName)

	newTopic := &kafka.KafkaTopic{
		InstanceId:   instanceId,
		Topic:        topicName,
		PartitionNum: d.Get("partition_num").(int),
		LocalTopic:   localTopic,
	}
	if localTopic {
		newTopic.ReplicaNum = d.Get("replica_num").(int)
		newTopic.CompactTopic = d.Get("compact_topic").(bool)
	}
	if v, ok := d.GetOk("remark"); ok {
		newTopic.Remark = v.(string)
	}

	if err := kafkaService.CreateAlikafkaTopic(newTopic); err != nil {
		if IsAlreadyExistError(err) {
			log.Printf("[INFO] Alikafka topic %s already exists, importing existing resource", resourceId)
			d.SetId(resourceId)
			return resourceAliCloudAlikafkaTopicRead(d, meta)
		}
		return WrapError(err)
	}

	d.SetId(resourceId)

	if err := kafkaService.WaitForAlikafkaTopic(d.Id(), Running, int(d.Timeout(schema.TimeoutCreate).Seconds())); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudAlikafkaTopicUpdate(d, meta)
}

func resourceAliCloudAlikafkaTopicUpdate(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}
	d.Partial(true)
	if err := kafkaService.setInstanceTags(d, TagResourceTopic); err != nil {
		return WrapError(err)
	}
	if d.IsNewResource() {
		d.Partial(false)
		return resourceAliCloudAlikafkaTopicRead(d, meta)
	}

	instanceId := d.Get("instance_id").(string)
	if d.HasChange("remark") {
		remark := d.Get("remark").(string)
		topic := d.Get("topic").(string)
		if err := kafkaService.ModifyAlikafkaTopicRemark(instanceId, topic, remark); err != nil {
			return WrapError(err)
		}
		d.SetPartial("remark")
	}

	if d.HasChange("partition_num") {
		o, n := d.GetChange("partition_num")
		oldPartitionNum := o.(int)
		newPartitionNum := n.(int)

		if newPartitionNum < oldPartitionNum {
			return WrapError(errors.New("partition_num only support adjust to a greater value."))
		} else {
			topic := d.Get("topic").(string)
			if err := kafkaService.ModifyAlikafkaTopicPartitions(instanceId, topic, int32(newPartitionNum-oldPartitionNum)); err != nil {
				return WrapError(err)
			}
			d.SetPartial("partition_num")
		}
	}

	d.Partial(false)
	return resourceAliCloudAlikafkaTopicRead(d, meta)
}

func resourceAliCloudAlikafkaTopicRead(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId, topicName, err := DecodeTopicId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	object, err := kafkaService.DescribeAlikafkaTopic(instanceId, topicName)
	if err != nil {
		// Handle exceptions
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("instance_id", object.InstanceId)
	d.Set("topic", object.Topic)
	d.Set("local_topic", object.LocalTopic)
	if object.LocalTopic {
		d.Set("compact_topic", object.CompactTopic)
		d.Set("replica_num", object.ReplicaNum)
	} else {
		d.Set("compact_topic", false)
		d.Set("replica_num", 3)
	}
	d.Set("partition_num", object.PartitionNum)
	d.Set("remark", object.Remark)

	tags, err := kafkaService.DescribeTags(d.Id(), nil, TagResourceTopic)
	if err != nil {
		return WrapError(err)
	}
	d.Set("tags", kafkaService.tagsToMap(tags))

	return nil
}

func resourceAliCloudAlikafkaTopicDelete(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}
	instanceId := parts[0]
	topic := parts[1]

	if err := kafkaService.DeleteAlikafkaTopic(instanceId, topic); err != nil {
		return WrapError(err)
	}

	return WrapError(kafkaService.WaitForAlikafkaTopic(d.Id(), Deleted, DefaultTimeoutMedium))
}
