package alicloud

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

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
	logstore, err := s.GetAPI().GetLogStore(projectName, logstoreName)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, logstoreName, "GetLogStore", AlibabaCloudSdkGoERROR)
	}

	return logstore, nil
}

// DescribeGetLogStoreMeteringMode returns LogStore metering mode information using structured data
func (s *SlsService) DescribeGetLogStoreMeteringMode(id string) (*aliyunSlsAPI.LogStoreMeteringMode, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return nil, WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
	}

	projectName := parts[0]
	logstoreName := parts[1]

	meteringMode, err := s.GetAPI().GetLogStoreMeteringMode(projectName, logstoreName)
	if err != nil {
		if NotFoundError(err) {
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
			if NotFoundError(err) {
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
			if logstore.HotTtl != nil {
				object["hotTtl"] = *logstore.HotTtl
			}
			if logstore.InfrequentAccessTTL != nil {
				object["infrequentAccessTtl"] = *logstore.InfrequentAccessTTL
			}
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
	return s.GetAPI().ListLogStores(project, logstoreName, mode, telemetryType)
}

// GetLogStore encapsulates the call to aliyunSlsAPI.GetLogStore
func (s *SlsService) GetLogStore(project, logstore string) (*aliyunSlsAPI.LogStore, error) {
	return s.GetAPI().GetLogStore(project, logstore)
}

// CreateLogStore encapsulates the call to aliyunSlsAPI.CreateLogStore
func (s *SlsService) CreateLogStore(project string, logstore *aliyunSlsAPI.LogStore) error {
	return s.GetAPI().CreateLogStore(project, logstore)
}

// CreateLogStoreIfNotExist creates a logstore if it does not exist
func (s *SlsService) CreateLogStoreIfNotExist(projectName string, logstore *aliyunSlsAPI.LogStore) (*aliyunSlsAPI.LogStore, error) {
	if logstore == nil {
		return nil, fmt.Errorf("logstore parameter cannot be nil")
	}

	if logstore.LogstoreName == "" {
		return nil, fmt.Errorf("logstore name cannot be empty")
	}

	// Check if logstore exists
	_, err := s.GetAPI().GetLogStore(projectName, logstore.LogstoreName)
	if err != nil {
		if NotFoundError(err) {
			// Logstore doesn't exist, create it with provided configuration
			if err := s.GetAPI().CreateLogStore(projectName, logstore); err != nil {
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
	return s.GetAPI().UpdateLogStore(project, logstoreName, logstore)
}

// DeleteLogStore encapsulates the call to aliyunSlsAPI.DeleteLogStore
func (s *SlsService) DeleteLogStore(project, logstoreName string) error {
	return s.GetAPI().DeleteLogStore(project, logstoreName)
}

// UpdateLogStoreMeteringMode encapsulates the call to aliyunSlsAPI.UpdateLogStoreMeteringMode
func (s *SlsService) UpdateLogStoreMeteringMode(project, logstoreName string, meteringMode *aliyunSlsAPI.LogStoreMeteringMode) error {
	return s.GetAPI().UpdateLogStoreMeteringMode(project, logstoreName, meteringMode)
}

// GetLogStoreShards encapsulates the call to aliyunSlsAPI.ListLogStoreShards
func (s *SlsService) GetLogStoreShards(project, logstoreName string) ([]*aliyunSlsAPI.LogStoreShard, error) {
	return s.GetAPI().ListLogStoreShards(project, logstoreName)
}

// GetLogStoreShard encapsulates the call to aliyunSlsAPI.GetLogStoreShard
func (s *SlsService) GetLogStoreShard(project, logstoreName string, shardId int32) (*aliyunSlsAPI.LogStoreShard, error) {
	return s.GetAPI().GetLogStoreShard(project, logstoreName, shardId)
}

// SplitLogStoreShard encapsulates the call to aliyunSlsAPI.SplitLogStoreShard
func (s *SlsService) SplitLogStoreShard(project, logstoreName string, shardId int32, splitKey string) ([]*aliyunSlsAPI.LogStoreShard, error) {
	return s.GetAPI().SplitLogStoreShard(project, logstoreName, shardId, splitKey)
}

// MergeLogStoreShards encapsulates the call to aliyunSlsAPI.MergeLogStoreShards
func (s *SlsService) MergeLogStoreShards(project, logstoreName string, shardId int32) ([]*aliyunSlsAPI.LogStoreShard, error) {
	return s.GetAPI().MergeLogStoreShards(project, logstoreName, shardId)
}

// FilterActiveShards filters shards that are in "readwrite" state
func (s *SlsService) FilterActiveShards(shards []*aliyunSlsAPI.LogStoreShard) []*aliyunSlsAPI.LogStoreShard {
	var activeShards []*aliyunSlsAPI.LogStoreShard
	for _, shard := range shards {
		if strings.ToLower(shard.Status) == "readwrite" {
			activeShards = append(activeShards, shard)
		}
	}
	return activeShards
}

// WaitForLogStoreShardCount waits until the number of active shards equals the target count
func (s *SlsService) WaitForLogStoreShardCount(project, logstoreName string, targetCount int, timeout time.Duration) error {
	return resource.Retry(timeout, func() *resource.RetryError {
		shards, err := s.GetLogStoreShards(project, logstoreName)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		activeShards := s.FilterActiveShards(shards)
		if len(activeShards) == targetCount {
			return nil
		}

		return resource.RetryableError(fmt.Errorf("waiting for shard count to be %d, current: %d", targetCount, len(activeShards)))
	})
}

// EnsureLogStoreShardCount adjusts the number of shards to match the target count
func (s *SlsService) EnsureLogStoreShardCount(project, logstoreName string, targetCount int) error {
	for i := 0; i < 500; i++ {
		shards, err := s.GetLogStoreShards(project, logstoreName)
		if err != nil {
			return err
		}
		activeShards := s.FilterActiveShards(shards)
		currentCount := len(activeShards)

		if currentCount == targetCount {
			return nil
		}

		if currentCount < targetCount {
			if err := s.splitOneShard(project, logstoreName, activeShards); err != nil {
				return err
			}
		} else {
			if err := s.mergeOneShard(project, logstoreName, activeShards); err != nil {
				return err
			}
		}

		time.Sleep(3 * time.Second)
	}
	return fmt.Errorf("failed to reach target shard count %d after max attempts", targetCount)
}

func (s *SlsService) splitOneShard(project, logstoreName string, activeShards []*aliyunSlsAPI.LogStoreShard) error {
	if len(activeShards) == 0 {
		return fmt.Errorf("no active shards to split")
	}

	var targetShard *aliyunSlsAPI.LogStoreShard
	var maxRange *big.Int = big.NewInt(0)

	for _, shard := range activeShards {
		begin := new(big.Int)
		end := new(big.Int)
		begin.SetString(shard.InclusiveBeginKey, 16)
		end.SetString(shard.ExclusiveEndKey, 16)

		r := new(big.Int).Sub(end, begin)
		if r.Cmp(maxRange) > 0 {
			maxRange = r
			targetShard = shard
		}
	}

	if targetShard == nil {
		return fmt.Errorf("failed to identify shard to split")
	}

	begin := new(big.Int)
	begin.SetString(targetShard.InclusiveBeginKey, 16)
	halfRange := new(big.Int).Div(maxRange, big.NewInt(2))
	splitKeyBig := new(big.Int).Add(begin, halfRange)

	splitKey := fmt.Sprintf("%032s", splitKeyBig.Text(16))

	_, err := s.SplitLogStoreShard(project, logstoreName, int32(targetShard.ShardId), splitKey)
	return err
}

func (s *SlsService) mergeOneShard(project, logstoreName string, activeShards []*aliyunSlsAPI.LogStoreShard) error {
	if len(activeShards) < 2 {
		return fmt.Errorf("not enough shards to merge")
	}

	sort.Slice(activeShards, func(i, j int) bool {
		bi := new(big.Int)
		bj := new(big.Int)
		bi.SetString(activeShards[i].InclusiveBeginKey, 16)
		bj.SetString(activeShards[j].InclusiveBeginKey, 16)
		return bi.Cmp(bj) < 0
	})

	var targetShardId int32 = -1
	minCombinedRange := new(big.Int)
	first := true

	for i := 0; i < len(activeShards)-1; i++ {
		s1 := activeShards[i]
		s2 := activeShards[i+1]

		if s1.ExclusiveEndKey != s2.InclusiveBeginKey {
			continue
		}

		begin := new(big.Int)
		end := new(big.Int)
		begin.SetString(s1.InclusiveBeginKey, 16)
		end.SetString(s2.ExclusiveEndKey, 16)

		combined := new(big.Int).Sub(end, begin)

		if first || combined.Cmp(minCombinedRange) < 0 {
			minCombinedRange = combined
			targetShardId = int32(s1.ShardId)
			first = false
		}
	}

	if targetShardId == -1 {
		for i := 0; i < len(activeShards)-1; i++ {
			if activeShards[i].ExclusiveEndKey == activeShards[i+1].InclusiveBeginKey {
				targetShardId = int32(activeShards[i].ShardId)
				break
			}
		}
		if targetShardId == -1 {
			return fmt.Errorf("failed to find adjacent shards to merge")
		}
	}

	_, err := s.MergeLogStoreShards(project, logstoreName, targetShardId)
	return err
}
