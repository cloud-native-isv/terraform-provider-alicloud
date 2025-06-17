package alicloud

import (
	"fmt"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// CreateSlsLogStoreIndex creates a new log store index
func (s *SlsService) CreateSlsLogStoreIndex(projectName string, logstoreName string, index *aliyunSlsAPI.LogStoreIndex) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	err := s.aliyunSlsAPI.CreateLogStoreIndex(projectName, logstoreName, index)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store_index", "CreateLogStoreIndex", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// UpdateSlsLogStoreIndex updates an existing log store index
func (s *SlsService) UpdateSlsLogStoreIndex(projectName string, logstoreName string, index *aliyunSlsAPI.LogStoreIndex) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	err := s.aliyunSlsAPI.UpdateLogStoreIndex(projectName, logstoreName, index)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store_index", "UpdateLogStoreIndex", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DeleteSlsLogStoreIndex deletes a log store index
func (s *SlsService) DeleteSlsLogStoreIndex(projectName string, logstoreName string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	err := s.aliyunSlsAPI.DeleteLogStoreIndex(projectName, logstoreName)
	if err != nil {
		if strings.Contains(err.Error(), "IndexConfigNotExist") {
			return nil // Index already deleted
		}
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_store_index", "DeleteLogStoreIndex", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// LogStoreIndexStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch a log store index
func (s *SlsService) LogStoreIndexStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsLogStoreIndex(id)
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

// GetSlsLogStoreIndex encapsulates the call to aliyunSlsAPI.GetLogStoreIndex
func (s *SlsService) GetSlsLogStoreIndex(projectName, logstoreName string) (*aliyunSlsAPI.LogStoreIndex, error) {
	return s.aliyunSlsAPI.GetLogStoreIndex(projectName, logstoreName)
}
