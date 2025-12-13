package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudLogOssIngestion() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudLogOssIngestionCreate,
		Read:   resourceAliCloudLogOssIngestionRead,
		Update: resourceAliCloudLogOssIngestionUpdate,
		Delete: resourceAliCloudLogOssIngestionDelete,
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
			"ingestion_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"logstore": {
				Type:     schema.TypeString,
				Required: true,
			},
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"pattern": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"encoding": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "utf-8",
				ValidateFunc: validation.StringInSlice([]string{"utf-8", "gbk"}, false),
			},
			"compression_codec": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "none",
				ValidateFunc: validation.StringInSlice([]string{"none", "gzip", "snappy", "lz4", "zstd"}, false),
			},
			"format_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"json", "delimited_text", "regex", "single_line_text"}, false),
			},
			"time_field": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"time_format": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"time_pattern": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"time_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "+0800",
			},
			"interval": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "300s",
			},
			"start_time": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"end_time": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"use_meta_index": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"restore_object_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"schedule_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Resident",
				ValidateFunc: validation.StringInSlice([]string{"Resident", "Trigger"}, false),
			},
			"schedule_interval": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"run_immediately": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"last_modified_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAliCloudLogOssIngestionCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	projectName := d.Get("project_name").(string)
	ingestionName := d.Get("ingestion_name").(string)

	// Build source configuration
	sourceConfig := make(map[string]interface{})
	sourceConfig["bucket"] = d.Get("bucket").(string)
	sourceConfig["encoding"] = d.Get("encoding").(string)
	sourceConfig["time_zone"] = d.Get("time_zone").(string)
	sourceConfig["interval"] = d.Get("interval").(string)
	sourceConfig["use_meta_index"] = d.Get("use_meta_index").(bool)
	sourceConfig["restore_object_enabled"] = d.Get("restore_object_enabled").(bool)

	if v, ok := d.GetOk("endpoint"); ok {
		sourceConfig["endpoint"] = v.(string)
	}
	if v, ok := d.GetOk("role_arn"); ok {
		sourceConfig["role_arn"] = v.(string)
	}
	if v, ok := d.GetOk("prefix"); ok {
		sourceConfig["prefix"] = v.(string)
	}
	if v, ok := d.GetOk("pattern"); ok {
		sourceConfig["pattern"] = v.(string)
	}
	if v, ok := d.GetOk("compression_codec"); ok {
		sourceConfig["compression_codec"] = v.(string)
	}
	if v, ok := d.GetOk("time_field"); ok {
		sourceConfig["time_field"] = v.(string)
	}
	if v, ok := d.GetOk("time_format"); ok {
		sourceConfig["time_format"] = v.(string)
	}
	if v, ok := d.GetOk("time_pattern"); ok {
		sourceConfig["time_pattern"] = v.(string)
	}
	if v, ok := d.GetOk("start_time"); ok {
		sourceConfig["start_time"] = int64(v.(int))
	}
	if v, ok := d.GetOk("end_time"); ok {
		sourceConfig["end_time"] = int64(v.(int))
	}

	// Set format based on format_type
	formatType := d.Get("format_type").(string)
	formatConfig := make(map[string]interface{})
	formatConfig["type"] = formatType
	sourceConfig["format"] = formatConfig

	// Build schedule
	schedule := &aliyunSlsAPI.Schedule{
		Type:           d.Get("schedule_type").(string),
		RunImmediately: d.Get("run_immediately").(bool),
	}

	if v, ok := d.GetOk("schedule_interval"); ok {
		schedule.Interval = v.(string)
	}

	// Build configuration
	config := aliyunSlsAPI.BuildOSSIngestionConfiguration(
		d.Get("logstore").(string),
		d.Get("bucket").(string),
		d.Get("endpoint").(string),
		d.Get("role_arn").(string),
		sourceConfig,
	)

	// Build ingestion object
	ingestion := &aliyunSlsAPI.Ingestion{
		ScheduledJob: aliyunSlsAPI.ScheduledJob{
			BaseJob: aliyunSlsAPI.BaseJob{
				Name:        ingestionName,
				DisplayName: d.Get("display_name").(string),
				Description: d.Get("description").(string),
				Type:        aliyunSlsAPI.INGESTION_JOB,
			},
			Schedule: schedule,
		},
		IngestionConfiguration: config,
	}

	_, err = slsService.CreateSlsOSSIngestion(projectName, ingestion)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_oss_ingestion", "CreateOSSIngestion", AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprintf("%s:%s", projectName, ingestionName))

	// Wait for ingestion to be ready
	stateConf := BuildStateConf([]string{}, []string{"RUNNING", "STOPPING"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, slsService.SlsOSSIngestionStateRefreshFunc(d.Id(), "status", []string{"FAILED"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudLogOssIngestionRead(d, meta)
}

func resourceAliCloudLogOssIngestionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	ingestion, err := slsService.DescribeSlsOSSIngestion(d.Id())
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return WrapError(fmt.Errorf("invalid resource ID format"))
	}

	d.Set("project_name", parts[0])
	d.Set("ingestion_name", parts[1])

	// Set basic fields
	d.Set("ingestion_name", ingestion.ScheduledJob.BaseJob.Name)
	d.Set("display_name", ingestion.ScheduledJob.BaseJob.DisplayName)
	d.Set("description", ingestion.ScheduledJob.BaseJob.Description)
	d.Set("status", ingestion.ScheduledJob.BaseJob.Status)

	// Set configuration
	if ingestion.IngestionConfiguration != nil {
		d.Set("logstore", ingestion.IngestionConfiguration.LogStore)

		if ingestion.IngestionConfiguration.Source != nil {
			source := ingestion.IngestionConfiguration.Source
			if bucket, ok := source["bucket"].(string); ok {
				d.Set("bucket", bucket)
			}
			if endpoint, ok := source["endpoint"].(string); ok {
				d.Set("endpoint", endpoint)
			}
			if roleARN, ok := source["roleARN"].(string); ok {
				d.Set("role_arn", roleARN)
			}
			if prefix, ok := source["prefix"].(string); ok {
				d.Set("prefix", prefix)
			}
			if pattern, ok := source["pattern"].(string); ok {
				d.Set("pattern", pattern)
			}
			if encoding, ok := source["encoding"].(string); ok {
				d.Set("encoding", encoding)
			}
			if compressionCodec, ok := source["compressionCodec"].(string); ok {
				d.Set("compression_codec", compressionCodec)
			}
			if timeField, ok := source["timeField"].(string); ok {
				d.Set("time_field", timeField)
			}
			if timeFormat, ok := source["timeFormat"].(string); ok {
				d.Set("time_format", timeFormat)
			}
			if timePattern, ok := source["timePattern"].(string); ok {
				d.Set("time_pattern", timePattern)
			}
			if timeZone, ok := source["timeZone"].(string); ok {
				d.Set("time_zone", timeZone)
			}
			if interval, ok := source["interval"].(string); ok {
				d.Set("interval", interval)
			}
			if useMetaIndex, ok := source["useMetaIndex"].(bool); ok {
				d.Set("use_meta_index", useMetaIndex)
			}
			if restoreObjectEnabled, ok := source["restoreObjectEnabled"].(bool); ok {
				d.Set("restore_object_enabled", restoreObjectEnabled)
			}

			if startTime, ok := source["startTime"].(int64); ok {
				d.Set("start_time", int(startTime))
			}
			if endTime, ok := source["endTime"].(int64); ok {
				d.Set("end_time", int(endTime))
			}

			// Set format type
			if format, ok := source["format"].(map[string]interface{}); ok {
				if formatType, ok := format["type"].(string); ok {
					d.Set("format_type", formatType)
				}
			}
		}
	}

	// Set schedule
	if ingestion.ScheduledJob.Schedule != nil {
		d.Set("schedule_type", ingestion.ScheduledJob.Schedule.Type)
		d.Set("schedule_interval", ingestion.ScheduledJob.Schedule.Interval)
		d.Set("run_immediately", ingestion.ScheduledJob.Schedule.RunImmediately)
	}

	return nil
}

func resourceAliCloudLogOssIngestionUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return WrapError(fmt.Errorf("invalid resource ID format"))
	}

	projectName := parts[0]
	ingestionName := parts[1]

	// Build source configuration
	sourceConfig := make(map[string]interface{})
	sourceConfig["bucket"] = d.Get("bucket").(string)
	sourceConfig["encoding"] = d.Get("encoding").(string)
	sourceConfig["time_zone"] = d.Get("time_zone").(string)
	sourceConfig["interval"] = d.Get("interval").(string)
	sourceConfig["use_meta_index"] = d.Get("use_meta_index").(bool)
	sourceConfig["restore_object_enabled"] = d.Get("restore_object_enabled").(bool)

	if v, ok := d.GetOk("endpoint"); ok {
		sourceConfig["endpoint"] = v.(string)
	}
	if v, ok := d.GetOk("role_arn"); ok {
		sourceConfig["role_arn"] = v.(string)
	}
	if v, ok := d.GetOk("prefix"); ok {
		sourceConfig["prefix"] = v.(string)
	}
	if v, ok := d.GetOk("pattern"); ok {
		sourceConfig["pattern"] = v.(string)
	}
	if v, ok := d.GetOk("compression_codec"); ok {
		sourceConfig["compression_codec"] = v.(string)
	}
	if v, ok := d.GetOk("time_field"); ok {
		sourceConfig["time_field"] = v.(string)
	}
	if v, ok := d.GetOk("time_format"); ok {
		sourceConfig["time_format"] = v.(string)
	}
	if v, ok := d.GetOk("time_pattern"); ok {
		sourceConfig["time_pattern"] = v.(string)
	}
	if v, ok := d.GetOk("start_time"); ok {
		sourceConfig["start_time"] = int64(v.(int))
	}
	if v, ok := d.GetOk("end_time"); ok {
		sourceConfig["end_time"] = int64(v.(int))
	}

	// Set format based on format_type
	formatType := d.Get("format_type").(string)
	formatConfig := make(map[string]interface{})
	formatConfig["type"] = formatType
	sourceConfig["format"] = formatConfig

	// Build schedule
	schedule := &aliyunSlsAPI.Schedule{
		Type:           d.Get("schedule_type").(string),
		RunImmediately: d.Get("run_immediately").(bool),
	}

	if v, ok := d.GetOk("schedule_interval"); ok {
		schedule.Interval = v.(string)
	}

	// Build configuration
	config := aliyunSlsAPI.BuildOSSIngestionConfiguration(
		d.Get("logstore").(string),
		d.Get("bucket").(string),
		d.Get("endpoint").(string),
		d.Get("role_arn").(string),
		sourceConfig,
	)

	// Build ingestion object
	ingestion := &aliyunSlsAPI.Ingestion{
		ScheduledJob: aliyunSlsAPI.ScheduledJob{
			BaseJob: aliyunSlsAPI.BaseJob{
				Name:        ingestionName,
				DisplayName: d.Get("display_name").(string),
				Description: d.Get("description").(string),
				Type:        aliyunSlsAPI.INGESTION_JOB,
			},
			Schedule: schedule,
		},
		IngestionConfiguration: config,
	}

	_, err = slsService.UpdateSlsOSSIngestion(projectName, ingestionName, ingestion)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateOSSIngestion", AlibabaCloudSdkGoERROR)
	}

	return resourceAliCloudLogOssIngestionRead(d, meta)
}

func resourceAliCloudLogOssIngestionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return WrapError(fmt.Errorf("invalid resource ID format"))
	}

	projectName := parts[0]
	ingestionName := parts[1]

	err = slsService.DeleteSlsOSSIngestion(projectName, ingestionName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteOSSIngestion", AlibabaCloudSdkGoERROR)
	}

	return nil
}
