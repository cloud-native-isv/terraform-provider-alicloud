package alicloud

import (
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunNasAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudFileSystems() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudFileSystemsRead,

		Schema: map[string]*schema.Schema{
			"storage_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"Capacity", "Performance", "standard", "advance"}, false),
			},
			"protocol_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"NFS", "SMB"}, false),
			},
			"description_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.ValidateRegexp,
			},
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				ForceNew: true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// Computed values
			"descriptions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"systems": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"create_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"protocol_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"storage_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"metered_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"encrypt_type": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"file_system_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"capacity": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"kms_key_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudFileSystemsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapError(err)
	}

	// Use service layer to get file systems
	fileSystems, err := nasService.ListNasFileSystems()
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_nas_file_systems", "ListNasFileSystems", AlibabaCloudSdkGoERROR)
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
	var filesystemDescriptionRegex *regexp.Regexp
	if v, ok := d.GetOk("description_regex"); ok {
		r, err := regexp.Compile(v.(string))
		if err != nil {
			return WrapError(err)
		}
		filesystemDescriptionRegex = r
	}

	var objects []*aliyunNasAPI.FileSystem
	for _, fileSystem := range fileSystems {
		if filesystemDescriptionRegex != nil {
			if !filesystemDescriptionRegex.MatchString(fileSystem.Description) {
				continue
			}
		}
		if v, ok := d.GetOk("storage_type"); ok && v.(string) != "" && fileSystem.StorageType != v.(string) {
			continue
		}
		if v, ok := d.GetOk("protocol_type"); ok && v.(string) != "" && fileSystem.ProtocolType != v.(string) {
			continue
		}
		if len(idsMap) > 0 {
			if _, ok := idsMap[fileSystem.FileSystemId]; !ok {
				continue
			}
		}
		objects = append(objects, fileSystem)
	}

	ids := make([]string, 0)
	descriptions := make([]interface{}, 0)
	s := make([]map[string]interface{}, 0)
	for _, object := range objects {
		mapping := map[string]interface{}{
			"id":               object.FileSystemId,
			"region_id":        object.RegionId,
			"create_time":      object.CreateTime,
			"description":      object.Description,
			"protocol_type":    object.ProtocolType,
			"storage_type":     object.StorageType,
			"metered_size":     int(object.UsedCapacity), // Use UsedCapacity as metered_size
			"encrypt_type":     int(object.EncryptType),
			"file_system_type": object.FileSystemType,
			"capacity":         int(object.Capacity),
			"kms_key_id":       object.KMSKeyId,
			"zone_id":          object.ZoneId,
		}
		ids = append(ids, object.FileSystemId)
		descriptions = append(descriptions, object.Description)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	if err := d.Set("descriptions", descriptions); err != nil {
		return WrapError(err)
	}

	if err := d.Set("systems", s); err != nil {
		return WrapError(err)
	}
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
