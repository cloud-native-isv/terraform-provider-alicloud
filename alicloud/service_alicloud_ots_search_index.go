package alicloud

import (
	"time"

	tablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Search Index management functions

func (s *OtsService) CreateOtsSearchIndex(instanceName string, index *tablestoreAPI.TablestoreSearchIndex) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	if err := api.CreateSearchIndex(instanceName, index); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, index.IndexName, "CreateSearchIndex", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) DescribeOtsSearchIndex(instanceName, tableName, indexName string) (*tablestoreAPI.TablestoreSearchIndex, error) {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	index, err := api.GetSearchIndex(instanceName, tableName, indexName)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, indexName, "DescribeSearchIndex", AlibabaCloudSdkGoERROR)
	}

	return index, nil
}

func (s *OtsService) UpdateOtsSearchIndex(instanceName string, index *tablestoreAPI.TablestoreSearchIndex) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	if err := api.UpdateSearchIndex(instanceName, index); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, index.IndexName, "UpdateSearchIndex", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) DeleteOtsSearchIndex(instanceName string, index *tablestoreAPI.TablestoreSearchIndex) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	if err := api.DeleteSearchIndex(instanceName, index); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, index.IndexName, "DeleteSearchIndex", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) ListOtsSearchIndexes(instanceName, tableName string) ([]*tablestoreAPI.TablestoreSearchIndex, error) {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	indexes, err := api.ListSearchIndexes(instanceName, tableName)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, tableName, "ListSearchIndex", AlibabaCloudSdkGoERROR)
	}

	return indexes, nil
}

// State refresh function for search index
func (s *OtsService) OtsSearchIndexStateRefreshFunc(instanceName, tableName, indexName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOtsSearchIndex(instanceName, tableName, indexName)
		if err != nil {
			if IsNotFoundError(err) {
				// 资源已删除，返回 nil, "", nil 表示目标达成
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// Use the new status type
		currentStatus := string(tablestoreAPI.TablestoreSearchIndexStatusExisting)

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

// Wait for search index creation
func (s *OtsService) WaitForOtsSearchIndexCreating(instanceName, tableName, indexName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{string(tablestoreAPI.TablestoreSearchIndexStatusNotFound)}, // pending states
		[]string{string(tablestoreAPI.TablestoreSearchIndexStatusExisting)}, // target states
		timeout,
		5*time.Second,
		s.OtsSearchIndexStateRefreshFunc(instanceName, tableName, indexName, []string{string(tablestoreAPI.TablestoreSearchIndexStatusFailed)}),
	)
	_, err := stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, indexName)
	}
	return nil
}

// Wait for search index deletion
func (s *OtsService) WaitForOtsSearchIndexDeleting(instanceName, tableName, indexName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{string(tablestoreAPI.TablestoreSearchIndexStatusExisting)}, // pending states
		[]string{}, // target states (empty = wait for resource disappear)
		timeout,
		5*time.Second,
		s.OtsSearchIndexStateRefreshFunc(instanceName, tableName, indexName, []string{string(tablestoreAPI.TablestoreSearchIndexStatusFailed)}),
	)

	_, err := stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, indexName)
	}
	return nil
}
