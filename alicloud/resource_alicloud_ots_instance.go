package alicloud

import (
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	tablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
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

			// Instance Configuration - Required field
			"instance_specification": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"SSD", "HYBRID",
				}, false),
				Description: "The instance specification type of the Tablestore instance. Valid values: SSD, HYBRID.",
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

			// Reserved CU Configuration
			"is_reserved_cu_instance": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the Tablestore instance is a reserved CU instance.",
			},
			"reserved_read_cu": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The reserved read CU of the Tablestore instance.",
			},
			"reserved_write_cu": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The reserved write CU of the Tablestore instance.",
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
		if IsNotFoundError(err) {
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

	// Handle ACL updates - only NetworkTypeACL and NetworkSourceACL are supported
	if d.HasChange("network_source_acl") || d.HasChange("network_type_acl") {
		instance := convertSchemaToTablestoreInstanceForACLUpdate(d)
		if err := otsService.UpdateOtsInstance(instance); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateInstanceACL", AlibabaCloudSdkGoERROR)
		}
	}

	// Handle other updates
	if d.HasChanges("alias_name", "description", "policy") {
		instance := convertSchemaToTablestoreInstanceForBasicUpdate(d)
		if err := otsService.UpdateOtsInstance(instance); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateInstanceBasic", AlibabaCloudSdkGoERROR)
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
		if IsNotFoundError(err) {
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

// convertSchemaToTablestoreInstanceForACLUpdate creates an instance object for ACL-only updates
func convertSchemaToTablestoreInstanceForACLUpdate(d *schema.ResourceData) *tablestoreAPI.TablestoreInstance {
	instance := &tablestoreAPI.TablestoreInstance{
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

	return instance
}

// convertSchemaToTablestoreInstanceForBasicUpdate creates an instance object for basic field updates
func convertSchemaToTablestoreInstanceForBasicUpdate(d *schema.ResourceData) *tablestoreAPI.TablestoreInstance {
	instance := &tablestoreAPI.TablestoreInstance{
		InstanceName: d.Id(),
	}

	if v, ok := d.GetOk("alias_name"); ok {
		instance.AliasName = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		instance.InstanceDescription = v.(string)
	}

	if v, ok := d.GetOk("policy"); ok {
		instance.Policy = v.(string)
	}

	return instance
}
