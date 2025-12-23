package alicloud

import (
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
			"vswitch_id": {
				Type:     schema.TypeString,
				Required: true,
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
			"security_group": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"service_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"config": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: alikafkaInstanceConfigDiffSuppressFunc,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: StringLenBetween(3, 64),
			},
			"vswitch_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"eip_max": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"status": {
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
	vpcService := VpcService{client}

	startInstanceReq := &StartInstanceRequest{
		RegionId:   client.RegionId,
		InstanceId: d.Get("instance_id").(string),
		VSwitchId:  d.Get("vswitch_id").(string),
	}

	if v, ok := d.GetOk("vpc_id"); ok {
		startInstanceReq.VpcId = v.(string)
	}

	if v, ok := d.GetOk("zone_id"); ok {
		startInstanceReq.ZoneId = v.(string)
	}

	if startInstanceReq.VpcId == "" {
		vsw, err := vpcService.DescribeVswitch(startInstanceReq.VSwitchId)
		if err != nil {
			return WrapError(err)
		}

		if startInstanceReq.VpcId == "" {
			if vpcId, ok := vsw["VpcId"].(string); ok {
				startInstanceReq.VpcId = vpcId
			}
		}
	}

	if v, ok := d.GetOk("vswitch_ids"); ok {
		startInstanceReq.VSwitchIds = expandStringList(v.([]interface{}))
	}

	if _, ok := d.GetOkExists("eip_max"); ok {
		startInstanceReq.DeployModule = "eip"
		startInstanceReq.IsEipInner = true
	}

	if v, ok := d.GetOk("name"); ok {
		startInstanceReq.Name = v.(string)
	}

	if v, ok := d.GetOk("security_group"); ok {
		startInstanceReq.SecurityGroup = v.(string)
	}

	if v, ok := d.GetOk("service_version"); ok {
		startInstanceReq.ServiceVersion = v.(string)
	}

	if v, ok := d.GetOk("config"); ok {
		startInstanceReq.Config = v.(string)
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		startInstanceReq.KMSKeyId = v.(string)
	}

	if v, ok := d.GetOk("selected_zones"); ok {
		startInstanceReq.SelectedZones = formatSelectedZonesReq(v.([]interface{}))
		log.Printf("[DEBUG] Resource alicloud_alikafka_deployment SelectedZones=%s", startInstanceReq.SelectedZones)
	}

	startInstanceReq.CrossZone = d.Get("cross_zone").(bool)

	err = kafkaService.StartInstance(startInstanceReq)
	if err != nil {
		return err
	}
	addDebug("StartInstance", "Success", startInstanceReq)

	d.SetId(startInstanceReq.InstanceId)

	// wait until running
	stateConf := BuildStateConf([]string{}, []string{"5"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, kafkaService.AliKafkaInstanceStateRefreshFunc(d.Id(), []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
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
	d.Set("status", object.ServiceStatus)
	d.Set("config", object.AllConfig)
	d.Set("kms_key_id", object.KmsKeyId)

	return nil
}

func resourceAliCloudAlikafkaDeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	kafkaService, err := NewKafkaService(client)
	if err != nil {
		return WrapError(err)
	}

	request := &StopInstanceRequest{
		RegionId:   client.RegionId,
		InstanceId: d.Id(),
	}

	err = kafkaService.StopInstance(request)
	if err != nil {
		return WrapError(err)
	}

	return nil
}
