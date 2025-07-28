package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/alibabacloud-go/tea/tea"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// OSS Export Service Methods

// DescribeSlsOSSExport - Get OSS export job configuration
func (s *SlsService) DescribeSlsOSSExport(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
		return
	}

	projectName := parts[0]
	exportName := parts[1]

	ossExport, err := s.aliyunSlsAPI.GetOSSExport(projectName, exportName)
	if err != nil {
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetOSSExport", AlibabaCloudSdkGoERROR)
	}

	// Convert CWS-Lib-Go OSSExport to map for compatibility
	result := make(map[string]interface{})
	if ossExport.Name != nil {
		result["name"] = *ossExport.Name
	}
	if ossExport.DisplayName != nil {
		result["displayName"] = *ossExport.DisplayName
	}
	if ossExport.Description != nil {
		result["description"] = *ossExport.Description
	}
	if ossExport.Status != nil {
		result["status"] = *ossExport.Status
	}
	if ossExport.Configuration != nil {
		result["configuration"] = convertOSSExportConfiguration(ossExport.Configuration)
	}
	if ossExport.CreateTime != nil {
		result["createTime"] = *ossExport.CreateTime
	}
	if ossExport.LastModifyTime != nil {
		result["lastModifyTime"] = *ossExport.LastModifyTime
	}

	return result, nil
}

// CreateSlsOSSExport creates a new OSS export
func (s *SlsService) CreateSlsOSSExport(projectName string, ossExport *aliyunSlsAPI.OSSExport) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI is not initialized")
	}

	err := s.aliyunSlsAPI.CreateOSSExport(projectName, ossExport)
	if err != nil {
		exportName := "unknown"
		if ossExport.Name != nil {
			exportName = *ossExport.Name
		}
		return WrapErrorf(err, DefaultErrorMsg, exportName, "CreateOSSExport", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// UpdateSlsOSSExport updates an existing OSS export
func (s *SlsService) UpdateSlsOSSExport(projectName, exportName string, ossExport *aliyunSlsAPI.OSSExport) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI is not initialized")
	}

	err := s.aliyunSlsAPI.UpdateOSSExport(projectName, exportName, ossExport)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, exportName, "UpdateOSSExport", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DeleteSlsOSSExport deletes an OSS export
func (s *SlsService) DeleteSlsOSSExport(projectName, exportName string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI is not initialized")
	}

	err := s.aliyunSlsAPI.DeleteOSSExport(projectName, exportName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, exportName, "DeleteOSSExport", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// StartSlsOSSExport starts an OSS export
func (s *SlsService) StartSlsOSSExport(projectName, exportName string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI is not initialized")
	}

	err := s.aliyunSlsAPI.StartOSSExport(projectName, exportName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, exportName, "StartOSSExport", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// StopSlsOSSExport stops an OSS export
func (s *SlsService) StopSlsOSSExport(projectName, exportName string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI is not initialized")
	}

	err := s.aliyunSlsAPI.StopOSSExport(projectName, exportName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, exportName, "StopOSSExport", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// ListSlsOSSExports lists OSS exports in a project
func (s *SlsService) ListSlsOSSExports(projectName, ossExportName, logstore string) ([]*aliyunSlsAPI.OSSExport, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI is not initialized")
	}

	exports, err := s.aliyunSlsAPI.ListOSSExports(projectName, ossExportName, logstore)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, projectName, "ListOSSExports", AlibabaCloudSdkGoERROR)
	}

	return exports, nil
}

// SlsOSSExportStateRefreshFunc returns a StateRefreshFunc for OSS export status monitoring
func (s *SlsService) SlsOSSExportStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsOSSExport(id)
		if err != nil {
			if IsNotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		if err != nil {
			return nil, "", WrapErrorf(err, FailedGetAttributeMsg, id, field)
		}
		currentStatus := fmt.Sprint(v)

		// Handle special field prefix for existence check
		if strings.HasPrefix(field, "#") {
			v, _ := jsonpath.Get(strings.TrimPrefix(field, "#"), object)
			if v != nil {
				currentStatus = "#CHECKSET"
			}
		}

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

// WaitForSlsOSSExportStatus waits for OSS export to reach target status
func (s *SlsService) WaitForSlsOSSExportStatus(projectName, exportName, targetStatus string, timeout time.Duration) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI is not initialized")
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"STARTING", "STOPPING", "CREATING", "UPDATING"},
		Target:     []string{targetStatus},
		Refresh:    s.SlsOSSExportStateRefreshFunc(fmt.Sprintf("%s:%s", projectName, exportName), "status", []string{"FAILED", "ERROR"}),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, fmt.Sprintf("%s:%s", projectName, exportName))
}

// Helper functions for data conversion

// convertOSSExportConfiguration converts CWS-Lib-Go OSSExportConfiguration to map
func convertOSSExportConfiguration(config *aliyunSlsAPI.OSSExportConfiguration) map[string]interface{} {
	if config == nil {
		return nil
	}

	result := make(map[string]interface{})

	if config.Logstore != nil {
		result["logstore"] = *config.Logstore
	}
	if config.RoleArn != nil {
		result["role_arn"] = *config.RoleArn
	}
	if config.FromTime != nil {
		result["from_time"] = *config.FromTime
	}
	if config.ToTime != nil {
		result["to_time"] = *config.ToTime
	}

	// Convert Sink configuration
	if config.Sink != nil {
		sink := make(map[string]interface{})

		if config.Sink.Bucket != nil {
			sink["bucket"] = *config.Sink.Bucket
		}
		if config.Sink.Prefix != nil {
			sink["prefix"] = *config.Sink.Prefix
		}
		if config.Sink.Suffix != nil {
			sink["suffix"] = *config.Sink.Suffix
		}
		if config.Sink.RoleArn != nil {
			sink["role_arn"] = *config.Sink.RoleArn
		}
		if config.Sink.Endpoint != nil {
			sink["endpoint"] = *config.Sink.Endpoint
		}
		if config.Sink.TimeZone != nil {
			sink["time_zone"] = *config.Sink.TimeZone
		}
		if config.Sink.ContentType != nil {
			sink["content_type"] = *config.Sink.ContentType
		}
		if config.Sink.CompressionType != nil {
			sink["compression_type"] = *config.Sink.CompressionType
		}
		if config.Sink.ContentDetail != nil {
			sink["content_detail"] = config.Sink.ContentDetail
		}
		if config.Sink.BufferInterval != nil {
			sink["buffer_interval"] = *config.Sink.BufferInterval
		}
		if config.Sink.BufferSize != nil {
			sink["buffer_size"] = *config.Sink.BufferSize
		}
		if config.Sink.PathFormat != nil {
			sink["path_format"] = *config.Sink.PathFormat
		}
		if config.Sink.PathFormatType != nil {
			sink["path_format_type"] = *config.Sink.PathFormatType
		}

		result["sink"] = sink
	}

	return result
}

// buildOSSExportConfigurationFromMap builds OSS export configuration from Terraform map
func buildOSSExportConfigurationFromMap(configMap map[string]interface{}) *aliyunSlsAPI.OSSExportConfiguration {
	config := &aliyunSlsAPI.OSSExportConfiguration{}

	if logstore, ok := configMap["logstore"].(string); ok {
		config.Logstore = tea.String(logstore)
	}
	if roleArn, ok := configMap["role_arn"].(string); ok {
		config.RoleArn = tea.String(roleArn)
	}
	if fromTime, ok := configMap["from_time"].(int64); ok {
		config.FromTime = tea.Int64(fromTime)
	}
	if toTime, ok := configMap["to_time"].(int64); ok {
		config.ToTime = tea.Int64(toTime)
	}

	if sinkMap, ok := configMap["sink"].(map[string]interface{}); ok {
		sink := &aliyunSlsAPI.OSSExportConfigurationSink{}

		if bucket, ok := sinkMap["bucket"].(string); ok {
			sink.Bucket = tea.String(bucket)
		}
		if prefix, ok := sinkMap["prefix"].(string); ok {
			sink.Prefix = tea.String(prefix)
		}
		if suffix, ok := sinkMap["suffix"].(string); ok {
			sink.Suffix = tea.String(suffix)
		}
		if roleArn, ok := sinkMap["role_arn"].(string); ok {
			sink.RoleArn = tea.String(roleArn)
		}
		if endpoint, ok := sinkMap["endpoint"].(string); ok {
			sink.Endpoint = tea.String(endpoint)
		}
		if timeZone, ok := sinkMap["time_zone"].(string); ok {
			sink.TimeZone = tea.String(timeZone)
		}
		if contentType, ok := sinkMap["content_type"].(string); ok {
			sink.ContentType = tea.String(contentType)
		}
		if compressionType, ok := sinkMap["compression_type"].(string); ok {
			sink.CompressionType = tea.String(compressionType)
		}
		if contentDetail, ok := sinkMap["content_detail"].(map[string]interface{}); ok {
			sink.ContentDetail = contentDetail
		}
		if bufferInterval, ok := sinkMap["buffer_interval"].(int64); ok {
			sink.BufferInterval = tea.Int64(bufferInterval)
		}
		if bufferSize, ok := sinkMap["buffer_size"].(int64); ok {
			sink.BufferSize = tea.Int64(bufferSize)
		}
		if pathFormat, ok := sinkMap["path_format"].(string); ok {
			sink.PathFormat = tea.String(pathFormat)
		}
		if pathFormatType, ok := sinkMap["path_format_type"].(string); ok {
			sink.PathFormatType = tea.String(pathFormatType)
		}

		config.Sink = sink
	}

	return config
}

// BuildOSSExportFromMap builds OSS export from Terraform configuration map
func (s *SlsService) BuildOSSExportFromMap(d map[string]interface{}) *aliyunSlsAPI.OSSExport {
	ossExport := &aliyunSlsAPI.OSSExport{}

	if name, ok := d["name"].(string); ok {
		ossExport.Name = tea.String(name)
	}
	if displayName, ok := d["display_name"].(string); ok {
		ossExport.DisplayName = tea.String(displayName)
	}
	if description, ok := d["description"].(string); ok {
		ossExport.Description = tea.String(description)
	}

	if configMap, ok := d["configuration"].(map[string]interface{}); ok {
		ossExport.Configuration = buildOSSExportConfigurationFromMap(configMap)
	}

	return ossExport
}
