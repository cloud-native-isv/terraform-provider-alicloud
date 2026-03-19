package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudLogOssExport() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudLogOssExportCreate,
		Read:   resourceAliCloudLogOssExportRead,
		Update: resourceAliCloudLogOssExportUpdate,
		Delete: resourceAliCloudLogOssExportDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"export_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"logstore": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"from_time": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"to_time": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"sink": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"suffix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"role_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
						"endpoint": {
							Type:     schema.TypeString,
							Required: true,
						},
						"time_zone": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "UTC",
						},
						"content_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"json", "parquet", "csv", "orc"}, false),
						},
						"compression_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"none", "snappy", "lz4", "gzip", "zstd"}, false),
							Default:      "none",
						},
						"path_format": {
							Type:     schema.TypeString,
							Required: true,
						},
						"path_format_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"time"}, false),
							Default:      "time",
						},
						"buffer_interval": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  300,
						},
						"buffer_size": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  256,
						},
						"content_detail": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modify_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAliCloudLogOssExportCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	projectName := d.Get("project_name").(string)
	exportName := d.Get("export_name").(string)

	// Build OSS export configuration
	ossExport := buildOSSExportFromResourceData(d)

	// Create OSS export
	err = slsService.CreateSlsOSSExport(projectName, ossExport)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_oss_export", "CreateOSSExport", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID
	d.SetId(fmt.Sprintf("%s:%s", projectName, exportName))

	// Wait for the OSS export to be created successfully
	stateConf := BuildStateConf([]string{"CREATING"}, []string{"RUNNING", "STOPPED"}, d.Timeout(schema.TimeoutCreate), 10*time.Second, slsService.SlsOSSExportStateRefreshFunc(d.Id(), "status", []string{"FAILED", "ERROR"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudLogOssExportRead(d, meta)
}

func resourceAliCloudLogOssExportRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	object, err := slsService.DescribeSlsOSSExport(d.Id())
	if err != nil {
		if NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_log_oss_export DescribeSlsOSSExport Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}
	d.Set("project_name", parts[0])
	d.Set("export_name", parts[1])

	d.Set("display_name", object["displayName"])
	d.Set("description", object["description"])
	d.Set("status", object["status"])
	d.Set("create_time", object["createTime"])
	d.Set("last_modify_time", object["lastModifyTime"])

	// Set configuration
	if config, ok := object["configuration"].(map[string]interface{}); ok {
		d.Set("logstore", config["logstore"])
		d.Set("role_arn", config["role_arn"])
		if fromTime, ok := config["from_time"].(int64); ok {
			d.Set("from_time", int(fromTime))
		}
		if toTime, ok := config["to_time"].(int64); ok {
			d.Set("to_time", int(toTime))
		}

		// Set sink configuration
		if sinkConfig, ok := config["sink"].(map[string]interface{}); ok {
			sink := make(map[string]interface{})
			sink["bucket"] = sinkConfig["bucket"]
			sink["prefix"] = sinkConfig["prefix"]
			sink["suffix"] = sinkConfig["suffix"]
			sink["role_arn"] = sinkConfig["role_arn"]
			sink["endpoint"] = sinkConfig["endpoint"]
			sink["time_zone"] = sinkConfig["time_zone"]
			sink["content_type"] = sinkConfig["content_type"]
			sink["compression_type"] = sinkConfig["compression_type"]
			sink["path_format"] = sinkConfig["path_format"]
			sink["path_format_type"] = sinkConfig["path_format_type"]
			if bufferInterval, ok := sinkConfig["buffer_interval"].(int64); ok {
				sink["buffer_interval"] = int(bufferInterval)
			}
			if bufferSize, ok := sinkConfig["buffer_size"].(int64); ok {
				sink["buffer_size"] = int(bufferSize)
			}
			sink["content_detail"] = sinkConfig["content_detail"]

			d.Set("sink", []interface{}{sink})
		}
	}

	return nil
}

func resourceAliCloudLogOssExportUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}
	projectName := parts[0]
	exportName := parts[1]

	// Build updated OSS export configuration
	ossExport := buildOSSExportFromResourceData(d)

	// Update OSS export
	err = slsService.UpdateSlsOSSExport(projectName, exportName, ossExport)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateOSSExport", AlibabaCloudSdkGoERROR)
	}

	// Wait for the update to complete
	stateConf := BuildStateConf([]string{"UPDATING"}, []string{"RUNNING", "STOPPED"}, d.Timeout(schema.TimeoutUpdate), 10*time.Second, slsService.SlsOSSExportStateRefreshFunc(d.Id(), "status", []string{"FAILED", "ERROR"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudLogOssExportRead(d, meta)
}

func resourceAliCloudLogOssExportDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}
	projectName := parts[0]
	exportName := parts[1]

	err = slsService.DeleteSlsOSSExport(projectName, exportName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteOSSExport", AlibabaCloudSdkGoERROR)
	}

	// Wait for deletion to complete
	stateConf := BuildStateConf([]string{"RUNNING", "STOPPED", "DELETING"}, []string{}, d.Timeout(schema.TimeoutDelete), 10*time.Second, slsService.SlsOSSExportStateRefreshFunc(d.Id(), "status", []string{"FAILED", "ERROR"}))
	if _, err := stateConf.WaitForState(); err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}

// Helper function to build OSS export from resource data
func buildOSSExportFromResourceData(d *schema.ResourceData) *aliyunSlsAPI.OSSExport {
	ossExport := &aliyunSlsAPI.OSSExport{
		Name:        tea.String(d.Get("export_name").(string)),
		DisplayName: tea.String(d.Get("display_name").(string)),
		Description: tea.String(d.Get("description").(string)),
	}

	// Build configuration
	configuration := &aliyunSlsAPI.OSSExportConfiguration{
		Logstore: tea.String(d.Get("logstore").(string)),
		RoleArn:  tea.String(d.Get("role_arn").(string)),
	}

	if fromTime, ok := d.GetOk("from_time"); ok {
		configuration.FromTime = tea.Int64(int64(fromTime.(int)))
	}
	if toTime, ok := d.GetOk("to_time"); ok {
		configuration.ToTime = tea.Int64(int64(toTime.(int)))
	}

	// Build sink configuration
	if sinkList, ok := d.GetOk("sink"); ok {
		sinks := sinkList.([]interface{})
		if len(sinks) > 0 {
			sinkData := sinks[0].(map[string]interface{})
			sink := &aliyunSlsAPI.OSSExportConfigurationSink{
				Bucket:          tea.String(sinkData["bucket"].(string)),
				RoleArn:         tea.String(sinkData["role_arn"].(string)),
				Endpoint:        tea.String(sinkData["endpoint"].(string)),
				PathFormat:      tea.String(sinkData["path_format"].(string)),
				PathFormatType:  tea.String(sinkData["path_format_type"].(string)),
				TimeZone:        tea.String(sinkData["time_zone"].(string)),
				ContentType:     tea.String(sinkData["content_type"].(string)),
				CompressionType: tea.String(sinkData["compression_type"].(string)),
				BufferInterval:  tea.Int64(int64(sinkData["buffer_interval"].(int))),
				BufferSize:      tea.Int64(int64(sinkData["buffer_size"].(int))),
			}

			if prefix, ok := sinkData["prefix"]; ok && prefix.(string) != "" {
				sink.Prefix = tea.String(prefix.(string))
			}
			if suffix, ok := sinkData["suffix"]; ok && suffix.(string) != "" {
				sink.Suffix = tea.String(suffix.(string))
			}
			if contentDetail, ok := sinkData["content_detail"]; ok {
				sink.ContentDetail = contentDetail.(map[string]interface{})
			}

			configuration.Sink = sink
		}
	}

	ossExport.Configuration = configuration
	return ossExport
}
