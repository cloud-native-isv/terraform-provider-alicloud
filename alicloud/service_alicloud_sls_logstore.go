package alicloud

import (
	"fmt"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeSlsLogStore returns LogStore information using structured data
func (s *SlsService) DescribeSlsLogStore(id string) (*aliyunSlsAPI.LogStore, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return nil, WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
	}

	projectName := parts[0]
	logstoreName := parts[1]

	logstore, err := s.aliyunSlsAPI.GetLogStore(projectName, logstoreName)
	if err != nil {
		if strings.Contains(err.Error(), "LogStoreNotExist") {
			return nil, WrapErrorf(NotFoundErr("LogStore", id), NotFoundMsg, "")
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "GetLogStore", AlibabaCloudSdkGoERROR)
	}

	return logstore, nil
}

// DescribeSlsLogStoreCompat returns LogStore information as map for compatibility with legacy code
func (s *SlsService) DescribeSlsLogStoreCompat(id string) (object map[string]interface{}, err error) {
	logstore, err := s.DescribeSlsLogStore(id)
	if err != nil {
		return nil, err
	}

	// Convert aliyunSlsAPI.LogStore to map[string]interface{} for compatibility
	result := make(map[string]interface{})
	result["logstoreName"] = logstore.LogstoreName
	result["ttl"] = logstore.Ttl
	result["shardCount"] = logstore.ShardCount
	result["enableWebTracking"] = logstore.EnableTracking
	result["autoSplit"] = logstore.AutoSplit
	result["maxSplitShard"] = logstore.MaxSplitShard
	result["appendMeta"] = logstore.AppendMeta
	result["hotTtl"] = logstore.HotTtl
	result["infrequentAccessTtl"] = logstore.InfrequentAccessTTL
	result["mode"] = logstore.Mode
	result["telemetryType"] = logstore.TelemetryType
	result["encryptConf"] = logstore.EncryptConf
	result["productType"] = logstore.ProductType
	result["createTime"] = logstore.CreateTime
	result["lastModifyTime"] = logstore.LastModifyTime

	return result, nil
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

// DescribeSlsLogStoreIndex returns LogStore index configuration using structured data
func (s *SlsService) DescribeSlsLogStoreIndex(id string) (*aliyunSlsAPI.LogStoreIndex, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return nil, WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
	}

	projectName := parts[0]
	logstoreName := parts[1]

	index, err := s.aliyunSlsAPI.GetLogStoreIndex(projectName, logstoreName)
	if err != nil {
		if strings.Contains(err.Error(), "IndexConfigNotExist") {
			return nil, WrapErrorf(NotFoundErr("LogStoreIndex", id), NotFoundMsg, "")
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "GetLogStoreIndex", AlibabaCloudSdkGoERROR)
	}

	return index, nil
}

// DescribeSlsLogStoreIndexCompat returns LogStore index configuration as map for compatibility with legacy code
func (s *SlsService) DescribeSlsLogStoreIndexCompat(id string) (object map[string]interface{}, err error) {
	index, err := s.DescribeSlsLogStoreIndex(id)
	if err != nil {
		return nil, err
	}

	// Convert to map for compatibility
	result := make(map[string]interface{})
	result["keys"] = index.Keys
	result["line"] = index.Line
	result["ttl"] = index.TTL

	return result, nil
}

// SlsLogStoreStateRefreshFunc returns a StateRefreshFunc for LogStore resource state monitoring
func (s *SlsService) SlsLogStoreStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Use the new structured method
		logstore, err := s.DescribeSlsLogStore(id)
		if err != nil {
			if NotFoundError(err) {
				return logstore, "", nil
			}
			return nil, "", WrapError(err)
		}

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

// UpdateLogStore encapsulates the call to aliyunSlsAPI.UpdateLogStore
func (s *SlsService) UpdateLogStore(project, logstoreName string, logstore *aliyunSlsAPI.LogStore) error {
	return s.aliyunSlsAPI.UpdateLogStore(project, logstoreName, logstore)
}

// DeleteLogStore encapsulates the call to aliyunSlsAPI.DeleteLogStore
func (s *SlsService) DeleteLogStore(project, logstoreName string) error {
	return s.aliyunSlsAPI.DeleteLogStore(project, logstoreName)
}

// CreateLogStoreIndex encapsulates the call to aliyunSlsAPI.CreateLogStoreIndex
func (s *SlsService) CreateLogStoreIndex(project, logstoreName string, index *aliyunSlsAPI.LogStoreIndex) error {
	return s.aliyunSlsAPI.CreateLogStoreIndex(project, logstoreName, index)
}

// UpdateLogStoreIndex encapsulates the call to aliyunSlsAPI.UpdateLogStoreIndex
func (s *SlsService) UpdateLogStoreIndex(project, logstoreName string, index *aliyunSlsAPI.LogStoreIndex) error {
	return s.aliyunSlsAPI.UpdateLogStoreIndex(project, logstoreName, index)
}

// DeleteLogStoreIndex encapsulates the call to aliyunSlsAPI.DeleteLogStoreIndex
func (s *SlsService) DeleteLogStoreIndex(project, logstoreName string) error {
	return s.aliyunSlsAPI.DeleteLogStoreIndex(project, logstoreName)
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
