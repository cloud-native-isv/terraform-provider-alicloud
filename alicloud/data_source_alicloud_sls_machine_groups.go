package alicloud

import (
	"fmt"
	"regexp"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudLogMachineGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudLogMachineGroupsRead,
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the log project.",
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRegexp,
				Description:  "A regex string to filter machine groups by name.",
			},
			"output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "File name where to save data source results (after running `terraform plan`).",
			},
			"names": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "A list of machine group names.",
			},
			"ids": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "A list of machine group IDs.",
			},
			"groups": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of machine groups.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the machine group.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the machine group.",
						},
						"project_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the log project.",
						},
						"group_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of the machine group.",
						},
						"machine_identify_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The machine identify type (ip or userdefined).",
						},
						"machine_list": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "The machine list of the machine group.",
						},
						"group_attribute": {
							Type:        schema.TypeList,
							Computed:    true,
							MaxItems:    1,
							Description: "The attributes of the machine group.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"external_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The external name of the machine group.",
									},
									"group_topic": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The topic of the machine group.",
									},
								},
							},
						},
						"create_time": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The creation time of the machine group (Unix timestamp).",
						},
						"last_modify_time": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The last modification time of the machine group (Unix timestamp).",
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudLogMachineGroupsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	logService := LogService{client}

	projectName := d.Get("project_name").(string)
	nameRegex := d.Get("name_regex").(string)

	var nameRegexp *regexp.Regexp
	if nameRegex != "" {
		var err error
		nameRegexp, err = regexp.Compile(nameRegex)
		if err != nil {
			return WrapError(err)
		}
	}

	// List all machine groups in the project
	var allMachineGroups []string
	var requestInfo *sls.Client
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		raw, err := client.WithLogClient(func(slsClient *sls.Client) (interface{}, error) {
			requestInfo = slsClient
			groups, _, err := slsClient.ListMachineGroup(projectName, 0, 500)
			return groups, err
		})
		if err != nil {
			if IsExpectedErrors(err, []string{"InternalServerError", LogClientTimeout}) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		if debugOn() {
			addDebug("ListMachineGroup", raw, requestInfo, map[string]string{
				"project": projectName,
			})
		}
		// raw should be []string now
		if groups, ok := raw.([]string); ok {
			allMachineGroups = groups
		} else {
			return resource.NonRetryableError(fmt.Errorf("unexpected response type from ListMachineGroup"))
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_log_machine_groups", "ListMachineGroup", AliyunLogGoSdkERROR)
	}

	var filteredGroups []string
	for _, groupName := range allMachineGroups {
		if nameRegexp != nil && !nameRegexp.MatchString(groupName) {
			continue
		}
		filteredGroups = append(filteredGroups, groupName)
	}

	var ids []string
	var names []string
	var groups []map[string]interface{}

	for _, groupName := range filteredGroups {
		id := fmt.Sprintf("%s:%s", projectName, groupName)

		// Get detailed information for each machine group
		machineGroup, err := logService.DescribeLogMachineGroup(id)
		if err != nil {
			if IsNotFoundError(err) {
				continue
			}
			return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_log_machine_groups", "DescribeLogMachineGroup", AliyunLogGoSdkERROR)
		}

		groupMapping := map[string]interface{}{
			"id":                    id,
			"name":                  machineGroup.Name,
			"project_name":          projectName,
			"group_type":            machineGroup.Type,
			"machine_identify_type": machineGroup.MachineIDType,
			"machine_list":          machineGroup.MachineIDList,
			"create_time":           int(machineGroup.CreateTime),
			"last_modify_time":      int(machineGroup.LastModifyTime),
		}

		// Set group attributes if available
		if machineGroup.Attribute.ExternalName != "" || machineGroup.Attribute.TopicName != "" {
			groupAttributes := make([]map[string]interface{}, 1)
			groupAttributes[0] = map[string]interface{}{
				"external_name": machineGroup.Attribute.ExternalName,
				"group_topic":   machineGroup.Attribute.TopicName,
			}
			groupMapping["group_attribute"] = groupAttributes
		}

		ids = append(ids, id)
		names = append(names, groupName)
		groups = append(groups, groupMapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}
	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}
	if err := d.Set("groups", groups); err != nil {
		return WrapError(err)
	}

	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), groups)
	}

	return nil
}
