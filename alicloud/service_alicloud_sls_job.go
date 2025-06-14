package alicloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeSlsAlert <<< Encapsulated get interface for Sls Alert.

func (s *SlsService) DescribeSlsAlert(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
		return
	}

	projectName := parts[0]
	alertName := parts[1]

	ctx := context.Background()
	alert, err := s.aliyunSlsAPI.GetAlert(ctx, projectName, alertName)
	if err != nil {
		if strings.Contains(err.Error(), "JobNotExist") {
			return object, WrapErrorf(NotFoundErr("Alert", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetAlert", AlibabaCloudSdkGoERROR)
	}

	// Convert aliyunSlsAPI.Alert to map[string]interface{} for compatibility
	result := make(map[string]interface{})
	result["name"] = alert.Name
	result["displayName"] = alert.DisplayName
	result["description"] = alert.Description
	result["state"] = alert.State
	result["status"] = alert.Status
	result["configuration"] = alert.Configuration
	result["schedule"] = alert.Schedule
	result["createTime"] = alert.CreateTime
	result["lastModifiedTime"] = alert.LastModifiedTime

	return result, nil
}

func (s *SlsService) SlsAlertStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsAlert(id)
		if err != nil {
			if NotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		currentStatus := fmt.Sprint(v)

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

// DescribeSlsScheduledSQL <<< Encapsulated get interface for Sls ScheduledSQL.

func (s *SlsService) DescribeSlsScheduledSQL(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
		return
	}

	projectName := parts[0]
	scheduledSQLName := parts[1]

	ctx := context.Background()
	scheduledSQL, err := s.aliyunSlsAPI.GetScheduledSQL(ctx, projectName, scheduledSQLName)
	if err != nil {
		if strings.Contains(err.Error(), "JobNotExist") {
			return object, WrapErrorf(NotFoundErr("ScheduledSQL", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetScheduledSQL", AlibabaCloudSdkGoERROR)
	}

	// Convert aliyunSlsAPI.ScheduledSQL to map[string]interface{} for compatibility
	result := make(map[string]interface{})
	result["name"] = scheduledSQL.Name
	result["displayName"] = scheduledSQL.DisplayName
	result["description"] = scheduledSQL.Description
	result["status"] = scheduledSQL.Status
	result["configuration"] = scheduledSQL.Configuration
	result["schedule"] = scheduledSQL.Schedule
	result["createTime"] = scheduledSQL.CreateTime
	result["lastModifiedTime"] = scheduledSQL.LastModifiedTime

	return result, nil
}

func (s *SlsService) SlsScheduledSQLStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsScheduledSQL(id)
		if err != nil {
			if NotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		currentStatus := fmt.Sprint(v)

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

// DescribeSlsEtl <<< Encapsulated get interface for Sls Etl.

func (s *SlsService) DescribeSlsEtl(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
		return
	}

	projectName := parts[0]
	etlName := parts[1]

	ctx := context.Background()
	etl, err := s.aliyunSlsAPI.GetETL(ctx, projectName, etlName)
	if err != nil {
		if strings.Contains(err.Error(), "JobNotExist") {
			return object, WrapErrorf(NotFoundErr("Etl", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetETL", AlibabaCloudSdkGoERROR)
	}

	// Convert aliyunSlsAPI.ETL to map[string]interface{} for compatibility
	result := make(map[string]interface{})
	result["name"] = etl.Name
	result["displayName"] = etl.DisplayName
	result["description"] = etl.Description
	result["status"] = etl.Status
	result["configuration"] = etl.Configuration
	result["schedule"] = etl.Schedule
	result["createTime"] = etl.CreateTime
	result["lastModifiedTime"] = etl.LastModifiedTime

	return result, nil
}

func (s *SlsService) SlsEtlStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsEtl(id)
		if err != nil {
			if NotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		currentStatus := fmt.Sprint(v)

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

// DescribeSlsOssExportSink <<< Encapsulated get interface for Sls OssExportSink.

func (s *SlsService) DescribeSlsOssExportSink(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
		return
	}

	projectName := parts[0]
	ossExportName := parts[1]

	ctx := context.Background()
	ossExport, err := s.aliyunSlsAPI.GetOSSExport(ctx, projectName, ossExportName)
	if err != nil {
		if strings.Contains(err.Error(), "JobNotExist") {
			return object, WrapErrorf(NotFoundErr("OssExportSink", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetOSSExport", AlibabaCloudSdkGoERROR)
	}

	// Convert aliyunSlsAPI.OSSExport to map[string]interface{} for compatibility
	result := make(map[string]interface{})
	result["name"] = ossExport.Name
	result["displayName"] = ossExport.DisplayName
	result["description"] = ossExport.Description
	result["status"] = ossExport.Status
	result["configuration"] = ossExport.Configuration
	result["schedule"] = ossExport.Schedule
	result["createTime"] = ossExport.CreateTime
	result["lastModifiedTime"] = ossExport.LastModifiedTime

	return result, nil
}

func (s *SlsService) SlsOssExportSinkStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsOssExportSink(id)
		if err != nil {
			if NotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		currentStatus := fmt.Sprint(v)

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

// DescribeSlsIngestion - Get ingestion job configuration
func (s *SlsService) DescribeSlsIngestion(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 3, len(parts)))
		return
	}

	projectName := parts[0]
	ingestionName := parts[2]

	ctx := context.Background()
	ingestion, err := s.aliyunSlsAPI.GetIngestion(ctx, projectName, ingestionName)
	if err != nil {
		if strings.Contains(err.Error(), "JobNotExist") {
			return object, WrapErrorf(NotFoundErr("Ingestion", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetIngestion", AlibabaCloudSdkGoERROR)
	}

	// Convert to map for compatibility
	result := make(map[string]interface{})
	result["name"] = ingestion.Name
	result["displayName"] = ingestion.DisplayName
	result["description"] = ingestion.Description
	result["status"] = ingestion.Status
	result["configuration"] = ingestion.Configuration
	result["schedule"] = ingestion.Schedule
	result["createTime"] = ingestion.CreateTime
	result["lastModifiedTime"] = ingestion.LastModifiedTime

	return result, nil
}

func (s *SlsService) SlsIngestionStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsIngestion(id)
		if err != nil {
			if NotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		currentStatus := fmt.Sprint(v)

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}
