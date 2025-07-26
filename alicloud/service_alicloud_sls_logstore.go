package alicloud

import (
	"fmt"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeLogStore returns LogStore information using structured data
func (s *SlsService) DescribeLogStoreById(id string) (*aliyunSlsAPI.LogStore, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return nil, WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
	}

	projectName := parts[0]
	logstoreName := parts[1]

	return s.DescribeLogStore(projectName, logstoreName)
}

// DescribeLogStore returns LogStore information using structured data
func (s *SlsService) DescribeLogStore(projectName, logstoreName string) (*aliyunSlsAPI.LogStore, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	logstore, err := s.aliyunSlsAPI.GetLogStore(projectName, logstoreName)
	if err != nil {
		if strings.Contains(err.Error(), "LogStoreNotExist") {
			return nil, WrapErrorf(NotFoundErr("LogStore", projectName, logstoreName), NotFoundMsg, "")
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, logstoreName, "GetLogStore", AlibabaCloudSdkGoERROR)
	}

	return logstore, nil
}

// DescribeGetLogStoreMeteringMode returns LogStore metering mode information using structured data
func (s *SlsService) DescribeGetLogStoreMeteringMode(id string) (*aliyunSlsAPI.LogStoreMeteringMode, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return nil, WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
	}

	projectName := parts[0]
	logstoreName := parts[1]

	meteringMode, err := s.aliyunSlsAPI.GetLogStoreMeteringMode(projectName, logstoreName)
	if err != nil {
		if strings.Contains(err.Error(), "LogStoreNotExist") {
			return nil, WrapErrorf(NotFoundErr("LogStore", id), NotFoundMsg, "")
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "GetLogStoreMeteringMode", AlibabaCloudSdkGoERROR)
	}

	return meteringMode, nil
}

// DescribeGetLogStoreMeteringModeCompat returns LogStore metering mode as map for compatibility with legacy code
func (s *SlsService) DescribeGetLogStoreMeteringModeCompat(id string) (object map[string]interface{}, err error) {
	meteringMode, err := s.DescribeGetLogStoreMeteringMode(id)
	if err != nil {
		return nil, err
	}

	// Convert aliyunSlsAPI.LogStoreMeteringMode to map[string]interface{} for compatibility
	result := make(map[string]interface{})
	result["meteringMode"] = meteringMode.MeteringMode

	return result, nil
}

// LogStoreStateRefreshFunc returns a StateRefreshFunc for LogStore resource state monitoring
func (s *SlsService) LogStoreStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Use the new structured method
		logstore, err := s.DescribeLogStoreById(id)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, "deleted", nil
			}
			return nil, "", WrapError(err)
		}

		// If logstore exists, it's available
		if logstore != nil {
			// Convert to map for jsonpath compatibility
			object := make(map[string]interface{})
			object["logstoreName"] = logstore.LogstoreName
			object["ttl"] = logstore.Ttl
			object["shardCount"] = logstore.ShardCount
			object["enableWebTracking"] = logstore.EnableTracking
			object["autoSplit"] = logstore.AutoSplit
			object["maxSplitShard"] = logstore.MaxSplitShard
			object["appendMeta"] = logstore.AppendMeta
			object["hotTtl"] = logstore.HotTtl
			object["infrequentAccessTtl"] = logstore.InfrequentAccessTTL
			object["mode"] = logstore.Mode
			object["telemetryType"] = logstore.TelemetryType
			object["encryptConf"] = logstore.EncryptConf
			object["productType"] = logstore.ProductType
			object["createTime"] = logstore.CreateTime
			object["lastModifyTime"] = logstore.LastModifyTime
			object["status"] = "available" // Set explicit status

			// For LogStore resources, if we can successfully retrieve the logstore, it's available
			// The field parameter is typically used to specify which field to check for state
			// but for LogStore, existence means available state
			currentStatus := "available"

			// Only try to get specific field if it's not the default "logstoreName" field
			// The issue was that "logstoreName" was being used as the state which returned the actual name
			if field != "" && field != "logstoreName" {
				if v, err := jsonpath.Get(field, object); err == nil && v != nil {
					fieldValue := fmt.Sprint(v)
					// Only use field value as status if it looks like a status (not a name)
					if fieldValue != "" && fieldValue != "<nil>" && fieldValue != logstore.LogstoreName {
						currentStatus = fieldValue
					}
				}
			}

			for _, failState := range failStates {
				if currentStatus == failState {
					return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
				}
			}

			return object, currentStatus, nil
		}

		// This should not happen, but handle it gracefully
		return nil, "unknown", nil
	}
}

// ListLogStores encapsulates the call to aliyunSlsAPI.ListLogStores
func (s *SlsService) ListLogStores(project, logstoreName, mode, telemetryType string) ([]*aliyunSlsAPI.LogStore, error) {
	return s.aliyunSlsAPI.ListLogStores(project, logstoreName, mode, telemetryType)
}

// GetLogStore encapsulates the call to aliyunSlsAPI.GetLogStore
func (s *SlsService) GetLogStore(project, logstore string) (*aliyunSlsAPI.LogStore, error) {
	return s.aliyunSlsAPI.GetLogStore(project, logstore)
}

// CreateLogStore encapsulates the call to aliyunSlsAPI.CreateLogStore
func (s *SlsService) CreateLogStore(project string, logstore *aliyunSlsAPI.LogStore) error {
	return s.aliyunSlsAPI.CreateLogStore(project, logstore)
}

// CreateLogStoreIfNotExist creates a logstore if it does not exist
func (s *SlsService) CreateLogStoreIfNotExist(projectName string, logstore *aliyunSlsAPI.LogStore) (*aliyunSlsAPI.LogStore, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	if logstore == nil {
		return nil, fmt.Errorf("logstore parameter cannot be nil")
	}

	if logstore.LogstoreName == "" {
		return nil, fmt.Errorf("logstore name cannot be empty")
	}

	// Check if logstore exists
	_, err := s.aliyunSlsAPI.GetLogStore(projectName, logstore.LogstoreName)
	if err != nil {
		if strings.Contains(err.Error(), "LogStoreNotExist") || strings.Contains(err.Error(), "does not exist") {
			// Logstore doesn't exist, create it with provided configuration
			if err := s.aliyunSlsAPI.CreateLogStore(projectName, logstore); err != nil {
				return nil, WrapErrorf(err, DefaultErrorMsg, logstore.LogstoreName, "CreateLogStore", AlibabaCloudSdkGoERROR)
			}
			// Return the created logstore
			return logstore, nil
		} else {
			// Other error occurred
			return nil, WrapErrorf(err, DefaultErrorMsg, logstore.LogstoreName, "GetLogStore", AlibabaCloudSdkGoERROR)
		}
	}
	// Logstore already exists, return nil, nil
	return nil, nil
}

// UpdateLogStore encapsulates the call to aliyunSlsAPI.UpdateLogStore
func (s *SlsService) UpdateLogStore(project, logstoreName string, logstore *aliyunSlsAPI.LogStore) error {
	return s.aliyunSlsAPI.UpdateLogStore(project, logstoreName, logstore)
}

// DeleteLogStore encapsulates the call to aliyunSlsAPI.DeleteLogStore
func (s *SlsService) DeleteLogStore(project, logstoreName string) error {
	return s.aliyunSlsAPI.DeleteLogStore(project, logstoreName)
}

// UpdateLogStoreMeteringMode encapsulates the call to aliyunSlsAPI.UpdateLogStoreMeteringMode
func (s *SlsService) UpdateLogStoreMeteringMode(project, logstoreName string, meteringMode *aliyunSlsAPI.LogStoreMeteringMode) error {
	return s.aliyunSlsAPI.UpdateLogStoreMeteringMode(project, logstoreName, meteringMode)
}

// GetLogStoreShards encapsulates the call to aliyunSlsAPI.ListLogStoreShards
func (s *SlsService) GetLogStoreShards(project, logstoreName string) ([]*aliyunSlsAPI.LogStoreShard, error) {
	return s.aliyunSlsAPI.ListLogStoreShards(project, logstoreName)
}

// GetLogStoreShard encapsulates the call to aliyunSlsAPI.GetLogStoreShard
func (s *SlsService) GetLogStoreShard(project, logstoreName string, shardId int32) (*aliyunSlsAPI.LogStoreShard, error) {
	return s.aliyunSlsAPI.GetLogStoreShard(project, logstoreName, shardId)
}

// SplitLogStoreShard encapsulates the call to aliyunSlsAPI.SplitLogStoreShard
func (s *SlsService) SplitLogStoreShard(project, logstoreName string, shardId int32, splitKey string) ([]*aliyunSlsAPI.LogStoreShard, error) {
	return s.aliyunSlsAPI.SplitLogStoreShard(project, logstoreName, shardId, splitKey)
}

// MergeLogStoreShards encapsulates the call to aliyunSlsAPI.MergeLogStoreShards
func (s *SlsService) MergeLogStoreShards(project, logstoreName string, shardId int32) ([]*aliyunSlsAPI.LogStoreShard, error) {
	return s.aliyunSlsAPI.MergeLogStoreShards(project, logstoreName, shardId)
}
