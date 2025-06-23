package alicloud

import (
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunNasAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAlicloudNasMountTargets() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudNasMountTargetsRead,
		Schema: map[string]*schema.Schema{
			"access_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"mount_target_domain": {
				Type:       schema.TypeString,
				Optional:   true,
				ForceNew:   true,
				Deprecated: "Field 'mount_target_domain' has been deprecated from provider version 1.53.0. New field 'ids' replaces it.",
			},
			"type": {
				Type:       schema.TypeString,
				Optional:   true,
				ForceNew:   true,
				Deprecated: "Field 'type' has been deprecated from provider version 1.95.0. New field 'network_type' replaces it.",
			},
			"network_type": {
				Type:     schema.TypeString,
				Optional: true,
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
			"file_system_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"Active", "Inactive", "Pending"}, false),
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"targets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"mount_target_domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"network_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vswitch_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAlicloudNasMountTargetsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapError(err)
	}

	fileSystemId := d.Get("file_system_id").(string)

	// Use service layer to get mount targets list
	mountTargets, err := nasService.ListNasMountTargets(fileSystemId)
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_nas_mount_targets", "ListNasMountTargets", AlibabaCloudSdkGoERROR)
	}

	idsMap := make(map[string]string)
	if v, ok := d.GetOk("ids"); ok {
		for _, vv := range v.([]interface{}) {
			if vv == nil {
				continue
			}
			idsMap[vv.(string)] = vv.(string)
		}
	}
	status, statusOk := d.GetOk("status")

	var objects []*aliyunNasAPI.MountTarget
	for _, mountTarget := range mountTargets {
		if v, ok := d.GetOk("access_group_name"); ok && v.(string) != "" && mountTarget.AccessGroupName != v.(string) {
			continue
		}
		if v, ok := d.GetOk("mount_target_domain"); ok && v.(string) != "" && mountTarget.MountTargetDomain != v.(string) {
			continue
		}
		if v, ok := d.GetOk("type"); ok && v.(string) != "" && mountTarget.NetworkType != v.(string) {
			continue
		}
		if v, ok := d.GetOk("network_type"); ok && v.(string) != "" && mountTarget.NetworkType != v.(string) {
			continue
		}
		if v, ok := d.GetOk("vpc_id"); ok && v.(string) != "" && mountTarget.VpcId != v.(string) {
			continue
		}
		if v, ok := d.GetOk("vswitch_id"); ok && v.(string) != "" && mountTarget.VSwitchId != v.(string) {
			continue
		}
		if len(idsMap) > 0 {
			if _, ok := idsMap[mountTarget.MountTargetDomain]; !ok {
				continue
			}
		}
		if statusOk && status.(string) != "" && status.(string) != mountTarget.Status {
			continue
		}
		objects = append(objects, &mountTarget)
	}

	ids := make([]string, 0)
	s := make([]map[string]interface{}, 0)
	for _, object := range objects {
		mapping := map[string]interface{}{
			"access_group_name":   object.AccessGroupName,
			"id":                  object.MountTargetDomain,
			"mount_target_domain": object.MountTargetDomain,
			"network_type":        object.NetworkType,
			"type":                object.NetworkType,
			"status":              object.Status,
			"vpc_id":              object.VpcId,
			"vswitch_id":          object.VSwitchId,
		}
		ids = append(ids, object.MountTargetDomain)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	if err := d.Set("targets", s); err != nil {
		return WrapError(err)
	}
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
