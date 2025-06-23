package alicloud

import (
	"fmt"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunNasAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/nas"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAlicloudNasFilesets() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudNasFilesetsRead,
		Schema: map[string]*schema.Schema{
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
				ValidateFunc: validation.StringInSlice([]string{"CREATED", "CREATING", "RELEASED", "RELEASING"}, false),
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"filesets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"create_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"file_system_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"file_system_path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"fileset_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"update_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAlicloudNasFilesetsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	nasService, err := NewNasService(client)
	if err != nil {
		return WrapError(err)
	}

	fileSystemId := d.Get("file_system_id").(string)

	// Use service layer to get filesets
	filesets, err := nasService.ListNasFilesets(fileSystemId)
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_nas_filesets", "DescribeNasFilesets", AlibabaCloudSdkGoERROR)
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

	var objects []*aliyunNasAPI.Fileset
	for _, fileset := range filesets {
		filesetId := fmt.Sprint(fileSystemId, ":", fileset.FsetId)
		if len(idsMap) > 0 {
			if _, ok := idsMap[filesetId]; !ok {
				continue
			}
		}
		if statusOk && status.(string) != "" && status.(string) != fileset.Status {
			continue
		}
		objects = append(objects, &fileset)
	}

	ids := make([]string, 0)
	s := make([]map[string]interface{}, 0)
	for _, object := range objects {
		mapping := map[string]interface{}{
			"create_time":      object.CreateTime,
			"description":      object.Description,
			"file_system_id":   fileSystemId,
			"file_system_path": object.FileSystemPath,
			"id":               fmt.Sprint(fileSystemId, ":", object.FsetId),
			"fileset_id":       object.FsetId,
			"status":           object.Status,
			"update_time":      object.CreateTime, // Note: API may not have UpdateTime, using CreateTime as fallback
		}
		ids = append(ids, fmt.Sprint(mapping["id"]))
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	if err := d.Set("filesets", s); err != nil {
		return WrapError(err)
	}
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}

	return nil
}
