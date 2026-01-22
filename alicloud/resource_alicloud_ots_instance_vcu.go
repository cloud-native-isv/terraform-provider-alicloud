package alicloud

import (
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudOtsInstanceVCU() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudOtsInstanceVCUCreate,
		Read:   resourceAliCloudOtsInstanceVCURead,
		Update: resourceAliCloudOtsInstanceVCUUpdate,
		Delete: resourceAliCloudOtsInstanceVCUDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			// Basic Information
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateOTSInstanceName,
				Description:  "The name of the Tablestore instance.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the Tablestore instance.",
			},

			// Network Configuration
			"network_source_acl": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"TRUST_PROXY",
					}, false),
				},
				Description: "The network source ACL of the Tablestore instance. Valid values: TRUST_PROXY.",
			},
			"network_type_acl": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"INTERNET", "VPC", "CLASSIC",
					}, false),
				},
				Description: "The network type ACL of the Tablestore instance. Valid values: INTERNET, VPC, CLASSIC.",
			},

			// Resource Management
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group ID of the Tablestore instance.",
			},

			// Status and Quota
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the Tablestore instance.",
			},
			"table_quota": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The table quota of the Tablestore instance.",
			},
			"vcu_quota": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The VCU quota of the Tablestore instance.",
			},
			"elastic_vcu_upper_limit": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Computed:    true,
				Description: "The elastic VCU upper limit of the Tablestore instance.",
			},

			// Read-only fields
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the Tablestore instance.",
			},
			"user_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user ID of the Tablestore instance.",
			},

			// Tags
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "A mapping of tags to assign to the Tablestore instance.",
			},
		},
	}
}

func resourceAliCloudOtsInstanceVCUCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	instance := &tablestore.TablestoreInstance{
		InstanceName:          d.Get("name").(string),
		InstanceSpecification: "VCU", // Hardcoded for VCU resource
	}

	if v, ok := d.GetOk("description"); ok {
		instance.InstanceDescription = v.(string)
	}
	if v, ok := d.GetOk("resource_group_id"); ok {
		instance.ResourceGroupId = v.(string)
	}
	if v, ok := d.GetOk("elastic_vcu_upper_limit"); ok {
		instance.ElasticVCUUpperLimit = float32(v.(float64))
	}

	// Convert network ACLs
	if v, ok := d.GetOk("network_source_acl"); ok {
		if networkSourceAcl, ok := v.(*schema.Set); ok {
			instance.NetworkSourceACL = convertSetToStringSlice(networkSourceAcl)
		}
	}
	if v, ok := d.GetOk("network_type_acl"); ok {
		if networkTypeAcl, ok := v.(*schema.Set); ok {
			instance.NetworkTypeACL = convertSetToStringSlice(networkTypeAcl)
		}
	}

	// Convert tags
	if v, ok := d.GetOk("tags"); ok {
		if tagsMap, ok := v.(map[string]interface{}); ok {
			instance.Tags = convertMapToTablestoreInstanceTags(tagsMap)
		}
	}

	if err := otsService.CreateOtsInstance(instance); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_ots_instance_vcu", "CreateInstance", AlibabaCloudSdkGoERROR)
	}

	d.SetId(instance.InstanceName)

	if err := otsService.WaitForOtsInstanceCreating(instance.InstanceName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudOtsInstanceVCURead(d, meta)
}

func resourceAliCloudOtsInstanceVCURead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	instance, err := otsService.DescribeOtsInstance(d.Id())
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("name", instance.InstanceName)
	d.Set("description", instance.InstanceDescription)
	d.Set("status", instance.InstanceStatus)
	d.Set("resource_group_id", instance.ResourceGroupId)
	d.Set("table_quota", instance.TableQuota)
	d.Set("vcu_quota", instance.VCUQuota)
	d.Set("elastic_vcu_upper_limit", instance.ElasticVCUUpperLimit)
	d.Set("user_id", instance.UserId)

	if !instance.CreateTime.IsZero() {
		d.Set("create_time", instance.CreateTime.Format("2006-01-02T15:04:05Z"))
	}

	if err := d.Set("network_source_acl", convertStringSliceToSet(instance.NetworkSourceACL)); err != nil {
		return err
	}
	if err := d.Set("network_type_acl", convertStringSliceToSet(instance.NetworkTypeACL)); err != nil {
		return err
	}
	if err := d.Set("tags", convertTablestoreInstanceTagsToMap(instance.Tags)); err != nil {
		return err
	}

	return nil
}

func resourceAliCloudOtsInstanceVCUUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Update ACL
	if d.HasChange("network_source_acl") || d.HasChange("network_type_acl") {
		instance := &tablestore.TablestoreInstance{
			InstanceName: d.Id(),
		}
		if v, ok := d.GetOk("network_source_acl"); ok {
			sourceACL := v.(*schema.Set).List()
			instance.NetworkSourceACL = make([]string, len(sourceACL))
			for i, acl := range sourceACL {
				instance.NetworkSourceACL[i] = acl.(string)
			}
		}
		if v, ok := d.GetOk("network_type_acl"); ok {
			typeACL := v.(*schema.Set).List()
			instance.NetworkTypeACL = make([]string, len(typeACL))
			for i, acl := range typeACL {
				instance.NetworkTypeACL[i] = acl.(string)
			}
		}
		if err := otsService.UpdateOtsInstance(instance); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateInstanceACL", AlibabaCloudSdkGoERROR)
		}
	}

	// Update Basic Info
	if d.HasChange("description") {
		instance := &tablestore.TablestoreInstance{
			InstanceName:        d.Id(),
			InstanceDescription: d.Get("description").(string),
		}
		if err := otsService.UpdateOtsInstance(instance); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateInstanceBasic", AlibabaCloudSdkGoERROR)
		}
	}

	// Update Elastic VCU Limit
	if d.HasChange("elastic_vcu_upper_limit") {
		if err := otsService.UpdateOtsInstanceElasticVCUUpperLimit(d.Id(), float32(d.Get("elastic_vcu_upper_limit").(float64))); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateInstanceElasticVCUUpperLimit", AlibabaCloudSdkGoERROR)
		}
	}

	// Update Tags
	if d.HasChange("tags") {
		old, new := d.GetChange("tags")
		oldTags := convertMapToTablestoreInstanceTags(old.(map[string]interface{}))
		newTags := convertMapToTablestoreInstanceTags(new.(map[string]interface{}))

		if len(oldTags) > 0 {
			var oldTagKeys []string
			for _, tag := range oldTags {
				oldTagKeys = append(oldTagKeys, tag.Key)
			}
			if err := otsService.UntagOtsInstance(d.Id(), oldTagKeys); err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UntagResources", AlibabaCloudSdkGoERROR)
			}
		}

		if len(newTags) > 0 {
			if err := otsService.TagOtsInstance(d.Id(), newTags); err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "TagResources", AlibabaCloudSdkGoERROR)
			}
		}
	}

	return resourceAliCloudOtsInstanceVCURead(d, meta)
}

func resourceAliCloudOtsInstanceVCUDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	if err := otsService.DeleteOtsInstance(d.Id()); err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteInstance", AlibabaCloudSdkGoERROR)
	}

	if err := otsService.WaitForOtsInstanceDeleting(d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
