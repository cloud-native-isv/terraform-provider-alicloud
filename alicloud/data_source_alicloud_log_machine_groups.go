package alicloud

import (
"regexp"

slsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
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
slsService, err := NewSlsService(client)
if err != nil {
return WrapError(err)
}

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
// NOTE: ListSlsMachineGroups returns []*MachineGroup and handles errors internally
allMachineGroups, err := slsService.ListSlsMachineGroups(projectName)
if err != nil {
return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_log_machine_groups", "ListMachineGroup", AlibabaCloudSdkGoERROR)
}

var filteredGroups []*slsAPI.MachineGroup
for _, group := range allMachineGroups {
if nameRegexp != nil && !nameRegexp.MatchString(group.Name) {
continue
}
filteredGroups = append(filteredGroups, group)
}

var ids []string
var names []string
var groups []map[string]interface{}

for _, machineGroup := range filteredGroups {
id := slsService.BuildMachineGroupId(projectName, machineGroup.Name)

groupMapping := map[string]interface{}{
"id":                    id,
"name":                  machineGroup.Name,
"project_name":          projectName,
"group_type":            machineGroup.Type,
"machine_identify_type": machineGroup.MachineIdType,
"machine_list":          machineGroup.MachineIdList,
"create_time":           int(machineGroup.CreateTime),
"last_modify_time":      int(machineGroup.LastModifyTime),
}

// Set group attributes if available
if machineGroup.Attribute != nil && (machineGroup.Attribute.ExternalName != "" || machineGroup.Attribute.GroupTopic != "") {
groupAttributes := make([]map[string]interface{}, 1)
groupAttributes[0] = map[string]interface{}{
"external_name": machineGroup.Attribute.ExternalName,
"group_topic":   machineGroup.Attribute.GroupTopic,
}
groupMapping["group_attribute"] = groupAttributes
}

ids = append(ids, id)
names = append(names, machineGroup.Name)
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
