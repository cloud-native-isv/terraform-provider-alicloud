package alicloud

import (
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudOtsInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudOtsInstanceCreate,
		Read:   resourceAliCloudOtsInstanceRead,
		Update: resourceAliCloudOtsInstanceUpdate,
		Delete: resourceAliCloudOtsInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
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
			"alias_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The alias name of the Tablestore instance.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the Tablestore instance.",
			},

			// Instance Configuration
			"cluster_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"SSD", "HYBRID",
				}, false),
				Description: "The cluster type of the Tablestore instance. Valid values: SSD, HYBRID.",
			},
			"storage_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The storage type of the Tablestore instance.",
			},

			// Network Configuration
			"network": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Vpc", "Vpc_CONSOLE", "NORMAL",
				}, false),
				Description: "The network type of the Tablestore instance. Valid values: Vpc, Vpc_CONSOLE, NORMAL.",
			},
			"network_source_acl": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The network source ACL of the Tablestore instance.",
			},
			"network_type_acl": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The network type ACL of the Tablestore instance.",
			},

			// Resource Management
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group ID of the Tablestore instance.",
			},
			"payment_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The payment type of the Tablestore instance.",
			},
			"is_multi_az": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the Tablestore instance is multi-AZ.",
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
				Computed:    true,
				Description: "The elastic VCU upper limit of the Tablestore instance.",
			},

			// Policy and Security
			"policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The policy of the Tablestore instance.",
			},
			"policy_version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The policy version of the Tablestore instance.",
			},

			// Read-only fields
			"region_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The region ID of the Tablestore instance.",
			},
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

func resourceAliCloudOtsInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Convert schema to TablestoreInstance
	instance := convertSchemaToTablestoreInstance(d)

	// Create instance
	if err := otsService.CreateOtsInstance(instance); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_ots_instance", "CreateInstance", AlibabaCloudSdkGoERROR)
	}

	d.SetId(instance.InstanceName)

	// Wait for instance to be ready
	if err := otsService.WaitForOtsInstanceCreating(instance.InstanceName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudOtsInstanceRead(d, meta)
}

func resourceAliCloudOtsInstanceRead(d *schema.ResourceData, meta interface{}) error {
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

	// Convert TablestoreInstance to schema
	return convertTablestoreInstanceToSchema(d, instance)
}

func resourceAliCloudOtsInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Check if there are changes that require update
	if d.HasChanges("alias_name", "description", "network", "network_source_acl", "network_type_acl", "policy") {
		// Convert schema to TablestoreInstance
		instance := convertSchemaToTablestoreInstance(d)

		// Update instance
		if err := otsService.UpdateOtsInstance(instance); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateInstance", AlibabaCloudSdkGoERROR)
		}
	}

	// Handle tags separately
	if d.HasChange("tags") {
		old, new := d.GetChange("tags")
		oldTags := convertMapToTablestoreInstanceTags(old.(map[string]interface{}))
		newTags := convertMapToTablestoreInstanceTags(new.(map[string]interface{}))

		// Remove old tags
		if len(oldTags) > 0 {
			var oldTagKeys []string
			for _, tag := range oldTags {
				oldTagKeys = append(oldTagKeys, tag.Key)
			}
			if err := otsService.UntagOtsInstance(d.Id(), oldTagKeys); err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UntagResources", AlibabaCloudSdkGoERROR)
			}
		}

		// Add new tags
		if len(newTags) > 0 {
			if err := otsService.TagOtsInstance(d.Id(), newTags); err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "TagResources", AlibabaCloudSdkGoERROR)
			}
		}
	}

	return resourceAliCloudOtsInstanceRead(d, meta)
}

func resourceAliCloudOtsInstanceDelete(d *schema.ResourceData, meta interface{}) error {
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

	// Wait for instance to be deleted
	if err := otsService.WaitForOtsInstanceDeleting(d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
