package alicloud

import (
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudAlikafkaDeployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudAlikafkaDeploymentCreate,
		Read:   resourceAliCloudAlikafkaDeploymentRead,
		Delete: resourceAliCloudAlikafkaDeploymentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
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
				ForceNew: true,
			},
			"cross_zone": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     true,
				Description: "Specifies whether to deploy the instance across zones. true: Deploy the instance across zones. false: Do not deploy the instance across zones. Default value: true.",
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
				ForceNew: true,
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
		options["selected_zones"] = formatSelectedZonesReq(v.([]interface{}))
		log.Printf("[DEBUG] Resource alicloud_alikafka_deployment SelectedZones=%s", formatSelectedZonesReq(v.([]interface{})))
	}

	if v, ok := d.GetOk("vswitch_ids"); ok {
		vswitchList := expandStringList(v.([]interface{}))
		if len(vswitchList) > 0 {
			options["vswitch_ids"] = vswitchList
		}
	}

	// Use CWS-Lib-Go API to start the instance
	err = kafkaService.kafkaApi.StartInstance(instanceId, client.RegionId, options["vpc_id"].(string), vswitchId, options)
	if err != nil {
		return WrapError(err)
	}

	addDebug("StartInstance", "Success", instanceId)

	d.SetId(instanceId)

	// Wait for deployment to complete using the new wait function
	err = kafkaService.WaitForAliKafkaInstanceCreating(d.Id(), d.Timeout(schema.TimeoutCreate))
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
	d.Set("name", object.Name)
	d.Set("vpc_id", object.VpcId)
	d.Set("vswitch_id", object.VSwitchId)
	d.Set("zone_id", object.ZoneId)
	d.Set("security_group", object.SecurityGroup)
	d.Set("config", object.AllConfig)
	d.Set("kms_key_id", object.KmsKeyId)
	d.Set("vswitch_ids", []string{object.VSwitchId})

	return nil
}

func resourceAliCloudAlikafkaDeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceId := d.Id()

	// Use CWS-Lib-Go API to stop the instance
	err = kafkaService.kafkaApi.StopInstance(instanceId, client.RegionId)
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
