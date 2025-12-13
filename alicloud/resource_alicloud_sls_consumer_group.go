package alicloud

import (
	"log"
	"regexp"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudSlsConsumerGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudSlsConsumerGroupCreate,
		Read:   resourceAliCloudSlsConsumerGroupRead,
		Update: resourceAliCloudSlsConsumerGroupUpdate,
		Delete: resourceAliCloudSlsConsumerGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"project": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{1,127}$`), "invalid project name"),
				Description:  "The name of the SLS project.",
			},
			"logstore": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{1,127}$`), "invalid logstore name"),
				Description:  "The name of the SLS logstore.",
			},
			"consumer_group": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{1,127}$`), "invalid consumer group name"),
				Description:  "The name of the SLS consumer group.",
			},
			"timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      60,
				ValidateFunc: validation.IntBetween(1, 86400),
				Description:  "The heartbeat timeout (seconds) for the consumer group.",
			},
			"order": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the consumer group processes data in order.",
			},
			"checkpoints": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Shard checkpoints associated with this consumer group.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"shard_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The shard ID.",
						},
						"checkpoint": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The checkpoint cursor.",
						},
						"update_time": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The unix timestamp when checkpoint was updated.",
						},
						"consumer": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The consumer name that updated the checkpoint.",
						},
					},
				},
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceAliCloudSlsConsumerGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	project := d.Get("project").(string)
	logstore := d.Get("logstore").(string)
	group := d.Get("consumer_group").(string)

	req := &aliyunSlsAPI.LogConsumerGroup{
		ProjectName:   project,
		LogstoreName:  logstore,
		ConsumerGroup: group,
		Timeout:       int32(d.Get("timeout").(int)),
		Order:         d.Get("order").(bool),
	}

	// Create with retry and adopt+converge behavior
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		e := service.CreateOrAdoptSlsConsumerGroup(project, logstore, req)
		if e != nil {
			if NeedRetry(e) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(e)
			}
			return resource.NonRetryableError(e)
		}
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_sls_consumer_group", "CreateOrAdoptSlsConsumerGroup", AlibabaCloudSdkGoERROR)
	}

	d.SetId(EncodeSlsConsumerGroupId(project, logstore, group))

	if werr := service.WaitForSlsConsumerGroupCreating(d.Id(), d.Timeout(schema.TimeoutCreate)); werr != nil {
		return WrapErrorf(werr, IdMsg, d.Id())
	}

	return resourceAliCloudSlsConsumerGroupRead(d, meta)
}

func resourceAliCloudSlsConsumerGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	project, logstore, group, err := DecodeSlsConsumerGroupId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	obj, err := service.DescribeSlsConsumerGroup(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_sls_consumer_group DescribeSlsConsumerGroup NotFound: %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("project", project)
	d.Set("logstore", logstore)
	d.Set("consumer_group", group)
	if obj != nil {
		d.Set("timeout", int(obj.Timeout))
		d.Set("order", obj.Order)
		// Populate checkpoints (best-effort, ignore errors)
		cps, cpErr := service.GetAPI().GetCheckPoint(project, logstore, group, nil)
		if cpErr == nil && cps != nil {
			items := make([]map[string]interface{}, 0, len(cps))
			for _, cp := range cps {
				items = append(items, map[string]interface{}{
					"shard_id":    int(cp.ShardId),
					"checkpoint":  cp.Checkpoint,
					"update_time": int(cp.UpdateTime),
					"consumer":    cp.Consumer,
				})
			}
			if err := d.Set("checkpoints", items); err != nil {
				return WrapError(err)
			}
		}
	}

	return nil
}

func resourceAliCloudSlsConsumerGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	project, logstore, group, err := DecodeSlsConsumerGroupId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	if d.HasChange("timeout") || d.HasChange("order") {
		req := &aliyunSlsAPI.LogConsumerGroup{
			Timeout: int32(d.Get("timeout").(int)),
			Order:   d.Get("order").(bool),
		}
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			e := service.UpdateSlsConsumerGroup(project, logstore, group, req)
			if e != nil {
				if NeedRetry(e) {
					time.Sleep(5 * time.Second)
					return resource.RetryableError(e)
				}
				return resource.NonRetryableError(e)
			}
			return nil
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateSlsConsumerGroup", AlibabaCloudSdkGoERROR)
		}
	}

	return resourceAliCloudSlsConsumerGroupRead(d, meta)
}

func resourceAliCloudSlsConsumerGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	project, logstore, group, err := DecodeSlsConsumerGroupId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		e := service.DeleteSlsConsumerGroup(project, logstore, group)
		if e != nil {
			if NotFoundError(e) {
				return nil
			}
			if NeedRetry(e) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(e)
			}
			return resource.NonRetryableError(e)
		}
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteSlsConsumerGroup", AlibabaCloudSdkGoERROR)
	}

	if werr := service.WaitForSlsConsumerGroupDeleting(d.Id(), d.Timeout(schema.TimeoutDelete)); werr != nil {
		// if already 404, treat as deleted
		if NotFoundError(werr) {
			return nil
		}
		return WrapErrorf(werr, IdMsg, d.Id())
	}

	return nil
}
