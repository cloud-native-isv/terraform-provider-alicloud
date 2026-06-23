package alicloud

import (
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/adbpg"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudAdbpgInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudAdbpgInstanceCreate,
		Read:   resourceAliCloudAdbpgInstanceRead,
		Update: resourceAliCloudAdbpgInstanceUpdate,
		Delete: resourceAliCloudAdbpgInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"engine_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"db_instance_class": {
				Type:     schema.TypeString,
				Required: true,
			},
			"db_instance_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"instance_network_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"vswitch_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"pay_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"seg_node_num": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"storage_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"storage_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"security_ip_list": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"master_node_num": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"resource_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"serverless_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"encryption_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"encryption_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"tags": tagsSchema(),
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_string": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"maintain_start_time": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"maintain_end_time": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAliCloudAdbpgInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	input := &adbpg.AdbpgInstanceCreate{
		RegionId:        client.RegionId,
		ZoneId:          d.Get("zone_id").(string),
		EngineVersion:   d.Get("engine_version").(string),
		DBInstanceClass: d.Get("db_instance_class").(string),
		SegNodeNum:      int32(d.Get("seg_node_num").(int)),
	}

	if v, ok := d.GetOk("db_instance_mode"); ok {
		input.DBInstanceMode = v.(string)
	}
	if v, ok := d.GetOk("instance_network_type"); ok {
		input.InstanceNetworkType = v.(string)
	}
	if v, ok := d.GetOk("vpc_id"); ok {
		input.VPCId = v.(string)
	}
	if v, ok := d.GetOk("vswitch_id"); ok {
		input.VSwitchId = v.(string)
	}
	if v, ok := d.GetOk("pay_type"); ok {
		input.PayType = v.(string)
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = v.(string)
	}
	if v, ok := d.GetOk("security_ip_list"); ok {
		input.SecurityIPList = v.(string)
	}
	if v, ok := d.GetOk("storage_size"); ok {
		input.StorageSize = int64(v.(int))
	}
	if v, ok := d.GetOk("storage_type"); ok {
		input.StorageType = v.(string)
	}
	if v, ok := d.GetOk("master_node_num"); ok {
		input.MasterNodeNum = int32(v.(int))
	}
	if v, ok := d.GetOk("resource_group_id"); ok {
		input.ResourceGroupId = v.(string)
	}
	if v, ok := d.GetOk("serverless_mode"); ok {
		input.ServerlessMode = v.(string)
	}
	if v, ok := d.GetOk("encryption_key"); ok {
		input.EncryptionKey = v.(string)
	}
	if v, ok := d.GetOk("encryption_type"); ok {
		input.EncryptionType = v.(string)
	}

	detail, err := adbpgService.CreateAdbpgInstance(input)
	if err != nil {
		return WrapError(err)
	}

	d.SetId(detail.DBInstanceId)

	if err := adbpgService.WaitForAdbpgInstanceRunning(d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return WrapError(err)
	}

	if v, ok := d.GetOk("tags"); ok {
		tags := tagsFromMap(v.(map[string]interface{}))
		adbpgTags := make([]adbpg.Tag, 0, len(tags))
		for _, t := range tags {
			adbpgTags = append(adbpgTags, adbpg.Tag{Key: t.Key, Value: t.Value})
		}
		if err := adbpgService.TagAdbpgResources(d.Id(), adbpgTags); err != nil {
			return WrapError(err)
		}
	}

	return resourceAliCloudAdbpgInstanceRead(d, meta)
}

func resourceAliCloudAdbpgInstanceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	detail, err := adbpgService.DescribeAdbpgInstance(d.Id())
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[WARN] ADBPG Instance %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("engine_version", detail.EngineVersion)
	d.Set("db_instance_class", detail.DBInstanceClass)
	d.Set("db_instance_mode", detail.DBInstanceMode)
	d.Set("instance_network_type", detail.InstanceNetworkType)
	d.Set("vpc_id", detail.VpcId)
	d.Set("vswitch_id", detail.VSwitchId)
	d.Set("zone_id", detail.ZoneId)
	d.Set("pay_type", detail.PayType)
	d.Set("seg_node_num", int(detail.SegNodeNum))
	d.Set("storage_size", int(detail.StorageSize))
	d.Set("storage_type", detail.StorageType)
	d.Set("description", detail.DBInstanceDescription)
	d.Set("security_ip_list", detail.SecurityIPList)
	d.Set("master_node_num", int(detail.MasterNodeNum))
	d.Set("resource_group_id", detail.ResourceGroupId)
	d.Set("serverless_mode", detail.ServerlessMode)
	d.Set("encryption_key", detail.EncryptionKey)
	d.Set("encryption_type", detail.EncryptionType)
	d.Set("status", detail.DBInstanceStatus)
	d.Set("connection_string", detail.ConnectionString)
	d.Set("port", detail.Port)
	d.Set("maintain_start_time", detail.MaintainStartTime)
	d.Set("maintain_end_time", detail.MaintainEndTime)
	d.Set("creation_time", detail.CreationTime)

	if len(detail.Tags) > 0 {
		tagMap := make(map[string]interface{})
		for _, t := range detail.Tags {
			tagMap[t.Key] = t.Value
		}
		d.Set("tags", tagMap)
	}

	return nil
}

func resourceAliCloudAdbpgInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	if d.HasChange("description") {
		if err := adbpgService.ModifyAdbpgInstanceDescription(d.Id(), d.Get("description").(string)); err != nil {
			return WrapError(err)
		}
	}

	if d.HasChange("maintain_start_time") || d.HasChange("maintain_end_time") {
		startTime := d.Get("maintain_start_time").(string)
		endTime := d.Get("maintain_end_time").(string)
		if err := adbpgService.ModifyAdbpgInstanceMaintainTime(d.Id(), startTime, endTime); err != nil {
			return WrapError(err)
		}
	}

	if d.HasChange("tags") {
		oldRaw, newRaw := d.GetChange("tags")
		oldTags := oldRaw.(map[string]interface{})
		newTags := newRaw.(map[string]interface{})

		removedKeys := make([]string, 0)
		for k := range oldTags {
			if _, ok := newTags[k]; !ok {
				removedKeys = append(removedKeys, k)
			}
		}
		if len(removedKeys) > 0 {
			if err := adbpgService.UntagAdbpgResources(d.Id(), removedKeys); err != nil {
				return WrapError(err)
			}
		}

		if len(newTags) > 0 {
			adbpgTags := make([]adbpg.Tag, 0, len(newTags))
			for k, v := range newTags {
				adbpgTags = append(adbpgTags, adbpg.Tag{Key: k, Value: v.(string)})
			}
			if err := adbpgService.TagAdbpgResources(d.Id(), adbpgTags); err != nil {
				return WrapError(err)
			}
		}
	}

	return resourceAliCloudAdbpgInstanceRead(d, meta)
}

func resourceAliCloudAdbpgInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	adbpgService, err := NewAdbpgService(client)
	if err != nil {
		return WrapError(err)
	}

	if err := adbpgService.DeleteAdbpgInstance(d.Id()); err != nil {
		return WrapError(err)
	}

	return WrapError(adbpgService.WaitForAdbpgInstanceDeleted(d.Id(), d.Timeout(schema.TimeoutDelete)))
}
