package alicloud

import (
"encoding/json"
"fmt"
"regexp"

slsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudLogLogtailConfigs() *schema.Resource {
return &schema.Resource{
Read: dataSourceAliCloudLogLogtailConfigsRead,
Schema: map[string]*schema.Schema{
"project_name": {
Type:        schema.TypeString,
Required:    true,
Description: "The name of the log project.",
},
"logstore_name": {
Type:        schema.TypeString,
Optional:    true,
Description: "Filter configs by logstore name.",
},
"name_regex": {
Type:         schema.TypeString,
Optional:     true,
ValidateFunc: validation.ValidateRegexp,
Description:  "A regex string to filter logtail configs by name.",
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
Description: "A list of logtail config names.",
},
"ids": {
Type:        schema.TypeList,
Elem:        &schema.Schema{Type: schema.TypeString},
Computed:    true,
Description: "A list of logtail config IDs.",
},
"configs": {
Type:        schema.TypeList,
Computed:    true,
Description: "A list of logtail configurations.",
Elem: &schema.Resource{
Schema: map[string]*schema.Schema{
"id": {
Type:        schema.TypeString,
Computed:    true,
Description: "The ID of the logtail config.",
},
"name": {
Type:        schema.TypeString,
Computed:    true,
Description: "The name of the logtail config.",
},
"project_name": {
Type:        schema.TypeString,
Computed:    true,
Description: "The name of the log project.",
},
"logstore_name": {
Type:        schema.TypeString,
Computed:    true,
Description: "The target logstore name.",
},
"input_type": {
Type:        schema.TypeString,
Computed:    true,
Description: "The input type of the logtail config.",
},
"output_type": {
Type:        schema.TypeString,
Computed:    true,
Description: "The output type of the logtail config.",
},
"log_sample": {
Type:        schema.TypeString,
Computed:    true,
Description: "The log sample of the logtail config.",
},
"create_time": {
Type:        schema.TypeInt,
Computed:    true,
Description: "The creation time of the logtail config (Unix timestamp).",
},
"last_modify_time": {
Type:        schema.TypeInt,
Computed:    true,
Description: "The last modification time of the logtail config (Unix timestamp).",
},
"input_detail": {
Type:        schema.TypeString,
Computed:    true,
Description: "The input detail configuration of the logtail config (JSON string).",
},
"output_detail": {
Type:        schema.TypeList,
Computed:    true,
MaxItems:    1,
Description: "The output detail configuration of the logtail config.",
Elem: &schema.Resource{
Schema: map[string]*schema.Schema{
"project_name": {
Type:        schema.TypeString,
Computed:    true,
Description: "The output project name.",
},
"logstore_name": {
Type:        schema.TypeString,
Computed:    true,
Description: "The output logstore name.",
},
"compress_type": {
Type:        schema.TypeString,
Computed:    true,
Description: "The compress type.",
},
},
},
},
},
},
},
},
}
}

func dataSourceAliCloudLogLogtailConfigsRead(d *schema.ResourceData, meta interface{}) error {
client := meta.(*connectivity.AliyunClient)
slsService, err := NewSlsService(client)
if err != nil {
return WrapError(err)
}

projectName := d.Get("project_name").(string)
logstoreName := d.Get("logstore_name").(string)
nameRegex := d.Get("name_regex").(string)

var filteredConfigs []*slsAPI.LogtailConfig

// Mode 1: Fetch specific names if provided and NO regex is present
if v, ok := d.GetOk("names"); ok && nameRegex == "" {
names := v.([]interface{})
for _, item := range names {
configName := item.(string)
// Construct 3-part ID for Service call: project:config:name
id := fmt.Sprintf("%s:config:%s", projectName, configName)
config, err := slsService.DescribeSlsLogtailConfig(id)
if err != nil {
if NotFoundError(err) {
continue
}
return WrapError(err)
}
filteredConfigs = append(filteredConfigs, config)
}
} else {
// Mode 2: List all and filter
// Use empty string filter to get all, then filter by regex
allConfigs, err := slsService.ListSlsLogtailConfigs(projectName, "")
if err != nil {
return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_log_logtail_configs", "ListConfig", AlibabaCloudSdkGoERROR)
}

var nameRegexp *regexp.Regexp
if nameRegex != "" {
nameRegexp, err = regexp.Compile(nameRegex)
if err != nil {
return WrapError(err)
}
}

for _, config := range allConfigs {
if nameRegexp != nil && !nameRegexp.MatchString(config.ConfigName) {
continue
}
filteredConfigs = append(filteredConfigs, config)
}
}

// Filter by logstore_name if provided
var finalConfigs []*slsAPI.LogtailConfig
if logstoreName != "" {
for _, config := range filteredConfigs {
if config.OutputDetail != nil && config.OutputDetail.LogstoreName == logstoreName {
finalConfigs = append(finalConfigs, config)
}
}
} else {
finalConfigs = filteredConfigs
}

var ids []string
var names []string
var configs []map[string]interface{}

for _, logtailConfig := range finalConfigs {
// Use Legacy ID format: projectName:configName
id := fmt.Sprintf("%s:%s", projectName, logtailConfig.ConfigName)

var inputDetailStr string
if logtailConfig.InputDetail != nil {
if inputDetailBytes, err := json.Marshal(logtailConfig.InputDetail); err == nil {
inputDetailStr = string(inputDetailBytes)
}
}

// Note: CompressorType/CompressType is not present in CWS-Lib-Go version of LogtailConfigOutputDetail
// ProjectName is also not in OutputDetail struct
outputDetailList := make([]map[string]interface{}, 1)
outputDetailList[0] = map[string]interface{}{
"project_name":  projectName,
"logstore_name": "",
"compress_type": "",
}
if logtailConfig.OutputDetail != nil {
outputDetailList[0]["logstore_name"] = logtailConfig.OutputDetail.LogstoreName
}

configMapping := map[string]interface{}{
"id":               id,
"name":             logtailConfig.ConfigName,
"project_name":     projectName,
"logstore_name":    "",
"input_type":       logtailConfig.InputType,
"output_type":      logtailConfig.OutputType,
"log_sample":       logtailConfig.LogSample,
"create_time":      int(logtailConfig.CreateTime),
"last_modify_time": int(logtailConfig.LastModifyTime),
"input_detail":     inputDetailStr,
"output_detail":    outputDetailList,
}
if logtailConfig.OutputDetail != nil {
configMapping["logstore_name"] = logtailConfig.OutputDetail.LogstoreName
}

ids = append(ids, id)
names = append(names, logtailConfig.ConfigName)
configs = append(configs, configMapping)
}

d.SetId(dataResourceIdHash(ids))
if err := d.Set("ids", ids); err != nil {
return WrapError(err)
}
if err := d.Set("names", names); err != nil {
return WrapError(err)
}
if err := d.Set("configs", configs); err != nil {
return WrapError(err)
}

if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
writeToFile(output.(string), configs)
}

return nil
}
