package alicloud

import (
	"fmt"
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	timeUtils "github.com/cloud-native-tools/cws-lib-go/lib/common/time"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudLogQuery() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudLogQueryRead,
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_-]{3,63}$`),
					"The project name must be 3-63 characters long and can contain letters, digits, hyphens, and underscores."),
			},
			"logstore_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_-]{3,63}$`),
					"The logstore name must be 3-63 characters long and can contain letters, digits, hyphens, and underscores."),
			},
			"query": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "*",
				Description: "Query string for log search. Default is '*' which matches all logs.",
			},
			"from_time": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Query start time. Supports Unix timestamp or relative time like 'now', '-1h', '-30m', '-1d'.",
			},
			"to_time": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Query end time. Supports Unix timestamp or relative time like 'now', '-1h', '-30m', '-1d'.",
			},
			"line_count": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      100,
				ValidateFunc: validation.IntBetween(1, 50000),
				Description:  "Maximum number of log lines to return. Range: 1-50000.",
			},
			"output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "File name where to save data source results (after running `terraform plan`).",
			},
			// Query result outputs
			"logs": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of log entries returned from the query.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content": {
							Type:        schema.TypeMap,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Log content as key-value pairs.",
						},
					},
				},
			},
			"query_meta": {
				Type:        schema.TypeList,
				Computed:    true,
				MaxItems:    1,
				Description: "Query execution metadata.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of log entries returned.",
						},
						"processed_rows": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of rows processed during query execution.",
						},
						"processed_bytes": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of bytes processed during query execution.",
						},
						"elapsed_millisecond": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Query execution time in milliseconds.",
						},
						"is_accurate": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the query result is accurate.",
						},
						"progress": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Query execution progress status.",
						},
						"has_sql": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the query contains SQL syntax.",
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudLogQueryRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Get input parameters
	projectName := d.Get("project_name").(string)
	logstoreName := d.Get("logstore_name").(string)
	query := d.Get("query").(string)
	fromTimeStr := d.Get("from_time").(string)
	toTimeStr := d.Get("to_time").(string)
	lineCount := int64(d.Get("line_count").(int))

	// Parse time parameters
	fromTime, err := timeUtils.ParseTimeParam(fromTimeStr)
	if err != nil {
		return WrapError(fmt.Errorf("invalid from_time: %w", err))
	}

	toTime, err := timeUtils.ParseTimeParam(toTimeStr)
	if err != nil {
		return WrapError(fmt.Errorf("invalid to_time: %w", err))
	}

	// Convert to int32 for API call
	fromTime32 := int32(fromTime)
	toTime32 := int32(toTime)

	// Validate time range
	if fromTime32 >= toTime32 {
		return WrapError(fmt.Errorf("from_time must be less than to_time"))
	}

	// Execute log query
	result, err := slsService.QuerySlsLogs(projectName, logstoreName, fromTime32, toTime32, query, lineCount)
	if err != nil {
		return WrapError(err)
	}

	// Set unique ID for the data source
	d.SetId(fmt.Sprintf("%s:%s:%d:%d", projectName, logstoreName, fromTime32, toTime32))

	// Process and set log data
	logs := make([]map[string]interface{}, 0)
	if result.Data != nil {
		for _, logEntry := range result.Data {
			logMap := map[string]interface{}{
				"content": logEntry,
			}
			logs = append(logs, logMap)
		}
	}
	if err := d.Set("logs", logs); err != nil {
		return WrapError(err)
	}

	// Process and set query metadata
	queryMeta := make([]map[string]interface{}, 0)
	if result.Meta != nil {
		metaMap := map[string]interface{}{
			"count":               int(result.Meta.Count),
			"processed_rows":      int(result.Meta.ProcessedRows),
			"processed_bytes":     int(result.Meta.ProcessedBytes),
			"elapsed_millisecond": int(result.Meta.ElapsedMillisecond),
			"is_accurate":         result.Meta.IsAccurate,
			"progress":            result.Meta.Progress,
			"has_sql":             result.Meta.HasSQL,
		}
		queryMeta = append(queryMeta, metaMap)
	}
	if err := d.Set("query_meta", queryMeta); err != nil {
		return WrapError(err)
	}

	// Write to output file if specified
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		if err := writeToFile(output.(string), logs); err != nil {
			return WrapError(err)
		}
	}

	return nil
}
