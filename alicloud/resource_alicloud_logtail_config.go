package alicloud

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	slsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAlicloudLogtailConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudLogtailConfigCreate,
		Read:   resourceAlicloudLogtailConfigRead,
		Update: resourceAlicloudLogtailConfigUpdate,
		Delete: resourceAlicloudLogtailConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"input_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"file", "plugin"}, false),
			},
			"log_sample": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"create_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"last_modify_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"logstore": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"output_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"LogService"}, false),
			},
			"input_detail": {
				Type:     schema.TypeString,
				Required: true,
				StateFunc: func(v interface{}) string {
					jsonString, _ := normalizeJsonString(v)
					return jsonString
				},
				ValidateFunc: validation.StringIsJSON,
			},
			"output_detail": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceAlicloudLogtailConfigCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_config", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	projectName := d.Get("project").(string)
	configName := d.Get("name").(string)
	logstoreName := d.Get("logstore").(string)

	// Parse input detail JSON directly to strongly typed structure
	inputDetailStr := d.Get("input_detail").(string)
	inputDetail := &slsAPI.LogtailConfigInputDetail{}
	if err := json.Unmarshal([]byte(inputDetailStr), inputDetail); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_config", "ParseInputDetail", AlibabaCloudSdkGoERROR)
	}

	// Build output detail - use a default SLS endpoint
	endpoint := fmt.Sprintf("%s.%s.log.aliyuncs.com", projectName, client.RegionId)
	outputDetail := &slsAPI.LogtailConfigOutputDetail{
		Endpoint:     endpoint,
		LogstoreName: logstoreName,
	}

	// Create LogtailConfig object with strongly typed InputDetail
	config := &slsAPI.LogtailConfig{
		ConfigName:   configName,
		InputType:    d.Get("input_type").(string),
		InputDetail:  inputDetail,
		OutputType:   d.Get("output_type").(string),
		OutputDetail: outputDetail,
	}

	if v, ok := d.GetOk("log_sample"); ok {
		config.LogSample = v.(string)
	}

	// Validate config
	if err := slsService.ValidateSlsLogtailConfig(config); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_config", "ValidateConfig", AlibabaCloudSdkGoERROR)
	}

	// Create logtail config with retry logic to handle ConfigAlreadyExist error
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		err := slsService.CreateSlsLogtailConfig(projectName, config)
		if err != nil {
			// Handle ConfigAlreadyExist error by importing existing resource
			if IsExpectedErrors(err, []string{"ConfigAlreadyExist"}) {
				log.Printf("[INFO] LogtailConfig %s already exists, importing existing resource", configName)
				d.SetId(fmt.Sprintf("%s%s%s%s%s", projectName, COLON_SEPARATED, "config", COLON_SEPARATED, configName))
				return nil
			}
			if IsExpectedErrors(err, []string{"InternalServerError", LogClientTimeout}) {
				time.Sleep(10 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_config", "CreateLogtailConfig", AlibabaCloudSdkGoERROR)
	}

	// Set ID if not already set (for new resources)
	if d.Id() == "" {
		d.SetId(fmt.Sprintf("%s%s%s%s%s", projectName, COLON_SEPARATED, "config", COLON_SEPARATED, configName))
	}

	return resourceAlicloudLogtailConfigRead(d, meta)
}

func resourceAlicloudLogtailConfigRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_config", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	object, err := slsService.DescribeSlsLogtailConfig(d.Id())
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}

	d.Set("project", parts[0])
	d.Set("name", object["configName"])
	d.Set("input_type", object["inputType"])
	d.Set("output_type", object["outputType"])
	d.Set("log_sample", object["logSample"])
	d.Set("create_time", object["createTime"])
	d.Set("last_modify_time", object["lastModifyTime"])

	// Set logstore from output detail
	if v, ok := object["logstoreName"]; ok {
		d.Set("logstore", v)
	}

	// Convert input detail to JSON string
	if inputDetail, ok := object["inputDetail"]; ok {
		inputDetailBytes, err := json.Marshal(inputDetail)
		if err != nil {
			return WrapError(err)
		}
		d.Set("input_detail", string(inputDetailBytes))
	}

	// Set output detail
	if outputDetail, ok := object["outputDetail"]; ok {
		d.Set("output_detail", outputDetail)
	}

	return nil
}

func resourceAlicloudLogtailConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_config", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}

	projectName := parts[0]
	configName := parts[2]

	// Parse input detail JSON directly to strongly typed structure
	inputDetailStr := d.Get("input_detail").(string)
	inputDetail := &slsAPI.LogtailConfigInputDetail{}
	if err := json.Unmarshal([]byte(inputDetailStr), inputDetail); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_config", "ParseInputDetail", AlibabaCloudSdkGoERROR)
	}

	// Build output detail
	endpoint := fmt.Sprintf("%s.%s.log.aliyuncs.com", projectName, client.RegionId)
	outputDetail := &slsAPI.LogtailConfigOutputDetail{
		Endpoint:     endpoint,
		LogstoreName: d.Get("logstore").(string),
	}

	// Create updated LogtailConfig object
	config := &slsAPI.LogtailConfig{
		ConfigName:   configName,
		InputType:    d.Get("input_type").(string),
		InputDetail:  inputDetail,
		OutputType:   d.Get("output_type").(string),
		OutputDetail: outputDetail,
	}

	if v, ok := d.GetOk("log_sample"); ok {
		config.LogSample = v.(string)
	}

	// Validate config
	if err := slsService.ValidateSlsLogtailConfig(config); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_config", "ValidateConfig", AlibabaCloudSdkGoERROR)
	}

	// Update logtail config
	if err := slsService.UpdateSlsLogtailConfig(projectName, configName, config); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_config", "UpdateLogtailConfig", AlibabaCloudSdkGoERROR)
	}

	return resourceAlicloudLogtailConfigRead(d, meta)
}

func resourceAlicloudLogtailConfigDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_config", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}

	projectName := parts[0]
	configName := parts[2]

	// Delete logtail config
	if err := slsService.DeleteSlsLogtailConfig(projectName, configName); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_config", "DeleteLogtailConfig", AlibabaCloudSdkGoERROR)
	}

	return nil
}
