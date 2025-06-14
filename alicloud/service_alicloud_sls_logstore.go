package alicloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeSlsLogStore <<< Encapsulated get interface for Sls LogStore.

func (s *SlsService) DescribeSlsLogStore(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
		return
	}

	projectName := parts[0]
	logstoreName := parts[1]

	ctx := context.Background()
	logstore, err := s.aliyunSlsAPI.GetLogStore(ctx, projectName, logstoreName)
	if err != nil {
		if strings.Contains(err.Error(), "LogStoreNotExist") {
			return object, WrapErrorf(NotFoundErr("LogStore", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetLogStore", AlibabaCloudSdkGoERROR)
	}

	// Convert aliyunSlsAPI.LogStore to map[string]interface{} for compatibility
	result := make(map[string]interface{})
	result["logstoreName"] = logstore.LogstoreName
	result["ttl"] = logstore.TTL
	result["shardCount"] = logstore.ShardCount
	result["enableWebTracking"] = logstore.EnableWebTracking
	result["autoSplit"] = logstore.AutoSplit
	result["maxSplitShard"] = logstore.MaxSplitShard
	result["appendMeta"] = logstore.AppendMeta
	result["hotTtl"] = logstore.HotTTL
	result["infrequentAccessTtl"] = logstore.InfrequentAccessTTL
	result["mode"] = logstore.Mode
	result["telemetryType"] = logstore.TelemetryType
	result["encryptConf"] = logstore.EncryptConf
	result["productType"] = logstore.ProductType
	result["createTime"] = logstore.CreateTime
	result["lastModifyTime"] = logstore.LastModifyTime

	return result, nil
}

func (s *SlsService) DescribeGetLogStoreMeteringMode(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
		return
	}

	projectName := parts[0]
	logstoreName := parts[1]

	ctx := context.Background()
	meteringMode, err := s.aliyunSlsAPI.GetLogStoreMeteringMode(ctx, projectName, logstoreName)
	if err != nil {
		if strings.Contains(err.Error(), "LogStoreNotExist") {
			return object, WrapErrorf(NotFoundErr("LogStore", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetLogStoreMeteringMode", AlibabaCloudSdkGoERROR)
	}

	// Convert aliyunSlsAPI.LogStoreMeteringMode to map[string]interface{} for compatibility
	result := make(map[string]interface{})
	result["meteringMode"] = meteringMode.MeteringMode

	return result, nil
}

// DescribeSlsLogStoreIndex - Get LogStore index configuration
func (s *SlsService) DescribeSlsLogStoreIndex(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
		return
	}

	projectName := parts[0]
	logstoreName := parts[1]

	ctx := context.Background()
	index, err := s.aliyunSlsAPI.GetLogStoreIndex(ctx, projectName, logstoreName)
	if err != nil {
		if strings.Contains(err.Error(), "IndexConfigNotExist") {
			return object, WrapErrorf(NotFoundErr("LogStoreIndex", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetLogStoreIndex", AlibabaCloudSdkGoERROR)
	}

	// Convert to map for compatibility
	result := make(map[string]interface{})
	result["keys"] = index.Keys
	result["line"] = index.Line
	result["ttl"] = index.TTL

	return result, nil
}

func (s *SlsService) SlsLogStoreStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsLogStore(id)
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
