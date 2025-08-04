package alicloud

import (
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudSelectDBInstanceClasses() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudSelectDBInstanceClassesRead,

		Schema: map[string]*schema.Schema{
			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The region ID.",
			},
			"ids": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "A list of SelectDB instance class codes.",
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
				ForceNew:     true,
				Description:  "A regex string to filter results by instance class code.",
			},
			"output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "File name where to save data source results (after running `terraform plan`).",
			},
			"classes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of SelectDB instance classes.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"class_code": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The class code of the instance.",
						},
						"cpu_cores": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of CPU cores.",
						},
						"memory_in_gb": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The memory size in GB.",
						},
						"default_storage_in_gb": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The default storage size in GB.",
						},
						"max_storage_in_gb": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The maximum storage size in GB.",
						},
						"min_storage_in_gb": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The minimum storage size in GB.",
						},
						"step_storage_in_gb": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The storage size increment step in GB.",
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudSelectDBInstanceClassesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	selectDBService, err := NewSelectDBService(client)
	if err != nil {
		return WrapError(err)
	}

	// Get all instance classes
	instanceClasses, err := selectDBService.DescribeSelectDBInstanceClasses()
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_selectdb_instance_classes", "DescribeSelectDBInstanceClasses", AlibabaCloudSdkGoERROR)
	}

	// Apply filters
	var filteredClasses []map[string]interface{}
	var ids []string

	// Get filter parameters
	var idsFilter []string
	if v, ok := d.GetOk("ids"); ok {
		for _, vv := range v.([]interface{}) {
			if vv != nil {
				idsFilter = append(idsFilter, vv.(string))
			}
		}
	}

	nameRegex, nameRegexOk := d.GetOk("name_regex")
	var nameRegexPattern *regexp.Regexp
	if nameRegexOk {
		nameRegexPattern, err = regexp.Compile(nameRegex.(string))
		if err != nil {
			return WrapError(err)
		}
	}

	for _, instanceClass := range instanceClasses {
		// Apply ID filter if specified
		if len(idsFilter) > 0 {
			found := false
			for _, filterId := range idsFilter {
				if instanceClass.ClassCode == filterId {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Apply name regex filter if specified
		if nameRegexOk {
			if !nameRegexPattern.MatchString(instanceClass.ClassCode) {
				continue
			}
		}

		mapping := map[string]interface{}{
			"class_code":            instanceClass.ClassCode,
			"cpu_cores":             instanceClass.CpuCores,
			"memory_in_gb":          instanceClass.MemoryInGB,
			"default_storage_in_gb": instanceClass.DefaultStorageInGB,
			"max_storage_in_gb":     instanceClass.MaxStorageInGB,
			"min_storage_in_gb":     instanceClass.MinStorageInGB,
			"step_storage_in_gb":    instanceClass.StepStorageInGB,
		}

		ids = append(ids, instanceClass.ClassCode)
		filteredClasses = append(filteredClasses, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	if err := d.Set("classes", filteredClasses); err != nil {
		return WrapError(err)
	}

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), filteredClasses)
	}

	return nil
}
