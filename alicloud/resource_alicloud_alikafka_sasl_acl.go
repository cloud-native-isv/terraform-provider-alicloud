package alicloud

import (
	"fmt"
	"log"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudAlikafkaSaslAcl() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudAlikafkaSaslAclCreate,
		Read:   resourceAliCloudAlikafkaSaslAclRead,
		Delete: resourceAliCloudAlikafkaSaslAclDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"username": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"acl_resource_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"Group", "Topic", "Cluster", "TransactionalId"}, false),
			},
			"acl_resource_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"acl_resource_pattern_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"LITERAL", "PREFIXED"}, false),
			},
			"acl_operation_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"Read", "Write"}, false),
			},
			"host": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAliCloudAlikafkaSaslAclCreate(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("instance_id").(string)
	username := d.Get("username").(string)
	aclResourceType := d.Get("acl_resource_type").(string)
	aclResourceName := d.Get("acl_resource_name").(string)
	aclResourcePatternType := d.Get("acl_resource_pattern_type").(string)
	aclOperationType := d.Get("acl_operation_type").(string)

	if err := kafkaService.CreateAlikafkaSaslAcl(instanceId, username, aclResourceType, aclResourceName, aclResourcePatternType, aclOperationType, nil); err != nil {
		return WrapError(err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s:%s:%s:%s", instanceId, username, aclResourceType, aclResourceName, aclResourcePatternType, aclOperationType))

	if err := kafkaService.WaitForAlikafkaSaslAcl(d.Id(), Running, int(d.Timeout(schema.TimeoutCreate).Seconds())); err != nil {
		return WrapError(err)
	}

	return resourceAliCloudAlikafkaSaslAclRead(d, meta)
}

func resourceAliCloudAlikafkaSaslAclRead(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 6)
	if err != nil {
		return WrapError(err)
	}
	object, err := kafkaService.DescribeAlikafkaSaslAcl(d.Id())
	if err != nil {
		// Handle exceptions
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_alikafka_sasl_acl kafkaService.DescribeAlikafkaSaslAcl Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("instance_id", parts[0])
	d.Set("username", object.Username)
	d.Set("acl_resource_type", object.AclResourceType)
	d.Set("acl_resource_name", object.AclResourceName)
	d.Set("acl_resource_pattern_type", object.AclResourcePatternType)
	d.Set("acl_operation_type", object.AclOperationType)
	d.Set("host", object.Host)

	return nil
}

func resourceAliCloudAlikafkaSaslAclDelete(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 6)
	if err != nil {
		return WrapError(err)
	}
	instanceId := parts[0]
	username := parts[1]
	aclResourceType := parts[2]
	aclResourceName := parts[3]
	aclResourcePatternType := parts[4]
	aclOperationType := parts[5]

	if err := kafkaService.DeleteAlikafkaSaslAcl(instanceId, username, aclResourceType, aclResourceName, aclResourcePatternType, aclOperationType, nil); err != nil {
		return WrapError(err)
	}

	return WrapError(kafkaService.WaitForAlikafkaSaslAcl(d.Id(), Deleted, DefaultTimeoutMedium))
}
