package alicloud

import (
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudAliKafkaInstanceAllowedIpAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudAliKafkaInstanceAllowedIpAttachmentCreate,
		Read:   resourceAliCloudAliKafkaInstanceAllowedIpAttachmentRead,
		Delete: resourceAliCloudAliKafkaInstanceAllowedIpAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"allowed_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: StringInSlice([]string{"vpc", "internet"}, false),
			},
			"port_range": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: StringInSlice([]string{"9092/9092", "9093/9093", "9094/9094", "9095/9095"}, false),
			},
			"allowed_ip": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAliCloudAliKafkaInstanceAllowedIpAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("instance_id").(string)
	allowedType := d.Get("allowed_type").(string)
	portRange := d.Get("port_range").(string)
	allowedIp := d.Get("allowed_ip").(string)

	if err := kafkaService.AttachAlikafkaAllowedIp(instanceId, allowedType, portRange, allowedIp, ""); err != nil {
		return WrapError(err)
	}

	d.SetId(fmt.Sprintf("%v:%v:%v:%v", instanceId, allowedType, portRange, allowedIp))

	return resourceAliCloudAliKafkaInstanceAllowedIpAttachmentRead(d, meta)
}

func resourceAliCloudAliKafkaInstanceAllowedIpAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	// The DescribeAliKafkaInstanceAllowedIpAttachment method is not available in the current KafkaService implementation
	// object, err := kafkaService.DescribeAliKafkaInstanceAllowedIpAttachment(d.Id())
	// if err != nil {
	// 	if !d.IsNewResource() && NotFoundError(err) {
	// 		log.Printf("[DEBUG] Resource alicloud_alikafka_instance_allowed_ip_attachment kafkaService.DescribeAliKafkaInstanceAllowedIpAttachment Failed!!! %s", err)
	// 		d.SetId("")
	// 		return nil
	// 	}
	// 	return WrapError(err)
	// }

	parts, err := ParseResourceId(d.Id(), 4)
	if err != nil {
		return WrapError(err)
	}

	d.Set("instance_id", parts[0])
	d.Set("allowed_type", parts[1])
	d.Set("port_range", parts[2])
	d.Set("allowed_ip", parts[3])

	return nil
}

func resourceAliCloudAliKafkaInstanceAllowedIpAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 4)
	if err != nil {
		return WrapError(err)
	}

	if err := kafkaService.DetachAlikafkaAllowedIp(parts[0], parts[1], parts[2], parts[3], ""); err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapError(err)
	}

	return nil
}
