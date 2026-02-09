package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudAlikafkaSaslUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudAlikafkaSaslUserCreate,
		Read:   resourceAliCloudAlikafkaSaslUserRead,
		Update: resourceAliCloudAlikafkaSaslUserUpdate,
		Delete: resourceAliCloudAlikafkaSaslUserDelete,
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
				ValidateFunc: StringLenBetween(1, 64),
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: StringInSlice([]string{"plain", "scram"}, false),
			},
			"password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: StringLenBetween(1, 64),
			},
			"kms_encrypted_password": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: kmsDiffSuppressFunc,
			},
			"kms_encryption_context": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return d.Get("kms_encrypted_password").(string) == ""
				},
			},
		},
	}
}

func resourceAliCloudAlikafkaSaslUserCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Get("instance_id").(string)
	username := d.Get("username").(string)
	options := map[string]interface{}{}
	if v, ok := d.GetOk("type"); ok {
		options["type"] = v.(string)
	}

	password := d.Get("password").(string)
	kmsPassword := d.Get("kms_encrypted_password").(string)

	if password == "" && kmsPassword == "" {
		return WrapError(Error("One of the 'password' and 'kms_encrypted_password' should be set."))
	}

	if password != "" {
		// use plain password
	} else {
		kmsService := KmsService{client}
		decryptResp, err := kmsService.Decrypt(kmsPassword, d.Get("kms_encryption_context").(map[string]interface{}))
		if err != nil {
			return WrapError(err)
		}
		password = decryptResp
	}

	if err := kafkaService.CreateAlikafkaSaslUser(instanceId, username, password, options); err != nil {
		return WrapError(err)
	}

	// Server may have cache, sleep a while.
	time.Sleep(2 * time.Second)

	d.SetId(fmt.Sprintf("%v:%v", instanceId, username))

	if err := kafkaService.WaitForAlikafkaSaslUser(d.Id(), Running, int(d.Timeout(schema.TimeoutCreate).Seconds())); err != nil {
		return WrapError(err)
	}

	return resourceAliCloudAlikafkaSaslUserRead(d, meta)
}

func resourceAliCloudAlikafkaSaslUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	object, err := kafkaService.DescribeAlikafkaSaslUser(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_ali_kafka_consumer_group kafkaService.DescribeAlikafkaSaslUser Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	d.Set("instance_id", parts[0])
	d.Set("username", object.Username)
	d.Set("type", object.Type)

	return nil
}

func resourceAliCloudAlikafkaSaslUserUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	options := map[string]interface{}{}
	if v, ok := d.GetOk("type"); ok {
		options["type"] = v.(string)
	}

	if !d.IsNewResource() && (d.HasChange("password") || d.HasChange("kms_encrypted_password")) {
		password := d.Get("password").(string)
		kmsPassword := d.Get("kms_encrypted_password").(string)

		if password == "" && kmsPassword == "" {
			return WrapError(Error("One of the 'password' and 'kms_encrypted_password' should be set."))
		}

		if password != "" {
			// use plain password
		} else {
			kmsService := KmsService{client}
			decryptResp, err := kmsService.Decrypt(kmsPassword, d.Get("kms_encryption_context").(map[string]interface{}))
			if err != nil {
				return WrapError(err)
			}
			password = decryptResp
		}

		if err := kafkaService.CreateAlikafkaSaslUser(parts[0], parts[1], password, options); err != nil {
			return WrapError(err)
		}
		// Server may have cache, sleep a while.
		time.Sleep(2 * time.Second)
	}

	return resourceAliCloudAlikafkaSaslUserRead(d, meta)
}

func resourceAliCloudAlikafkaSaslUserDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	options := map[string]interface{}{}
	if v, ok := d.GetOk("type"); ok {
		options["type"] = v.(string)
	}

	if err := kafkaService.DeleteAlikafkaSaslUser(parts[0], parts[1], options); err != nil {
		return WrapError(err)
	}

	return WrapError(kafkaService.WaitForAlikafkaSaslUser(d.Id(), Deleted, int(d.Timeout(schema.TimeoutDelete).Seconds())))
}
