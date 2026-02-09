package alicloud

import (
	"fmt"
	"log"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudAlikafkaConsumerGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudAlikafkaConsumerGroupCreate,
		Update: resourceAliCloudAlikafkaConsumerGroupUpdate,
		Read:   resourceAliCloudAlikafkaConsumerGroupRead,
		Delete: resourceAliCloudAlikafkaConsumerGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"consumer_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: StringLenBetween(1, 64),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAliCloudAlikafkaConsumerGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("instance_id").(string)
	consumerId := d.Get("consumer_id").(string)
	remark := ""
	if v, ok := d.GetOk("description"); ok {
		remark = v.(string)
	}

	if _, err := kafkaService.CreateAlikafkaConsumerGroup(instanceId, consumerId, remark, extractTags(d)); err != nil {
		return WrapError(err)
	}

	d.SetId(fmt.Sprint(instanceId, ":", consumerId))

	if err := kafkaService.WaitForAlikafkaConsumerGroup(d.Id(), Running, int(d.Timeout(schema.TimeoutCreate).Seconds())); err != nil {
		return WrapError(err)
	}

	return resourceAliCloudAlikafkaConsumerGroupUpdate(d, meta)
}

func resourceAliCloudAlikafkaConsumerGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}
	object, err := kafkaService.DescribeAlikafkaConsumerGroup(d.Id())
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_ali_kafka_consumer_group kafkaService.DescribeAlikafkaConsumerGroup Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}
	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}
	d.Set("consumer_id", parts[1])
	d.Set("instance_id", parts[0])
	d.Set("description", object.Remark)
	if object.Tags.TagVO != nil {
		d.Set("tags", kafkaService.tagVOTagsToMap(object.Tags.TagVO))
	}

	return nil
}

func resourceAliCloudAlikafkaConsumerGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}
	if d.HasChange("tags") {
		if err := kafkaService.SetResourceTags(d, "CONSUMERGROUP"); err != nil {
			return WrapError(err)
		}
		d.SetPartial("tags")
	}
	return resourceAliCloudAlikafkaConsumerGroupRead(d, meta)
}

func resourceAliCloudAlikafkaConsumerGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}
	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}
	if err := kafkaService.DeleteAlikafkaConsumerGroup(parts[0], parts[1]); err != nil {
		return WrapError(err)
	}
	return WrapError(kafkaService.WaitForAlikafkaConsumerGroup(d.Id(), Deleted, DefaultTimeoutMedium))
}
