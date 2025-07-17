package alicloud

import (
	"fmt"
	"strings"
	"time"

	tablestoreSDK "github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	tablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Secondary Index management functions

func (s *OtsService) CreateOtsIndex(instanceName string, index *tablestoreAPI.TablestoreIndex) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	if err := api.CreateIndex(instanceName, index); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, index.IndexName, "CreateIndex", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) DescribeOtsIndex(instanceName, tableName, indexName string) (*tablestoreAPI.TablestoreIndex, error) {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	index, err := api.GetIndex(instanceName, tableName, indexName)
	if err != nil {
		if IsExpectedErrors(err, []string{"NotExist", "InvalidIndexName.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, indexName, "DescribeIndex", AlibabaCloudSdkGoERROR)
	}

	return index, nil
}

func (s *OtsService) DeleteOtsIndex(instanceName string, index *tablestoreAPI.TablestoreIndex) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	if err := api.DeleteIndex(instanceName, index); err != nil {
		if IsExpectedErrors(err, []string{"NotExist", "InvalidIndexName.NotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, index.IndexName, "DeleteIndex", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) WaitForOtsIndexCreating(instanceName, tableName, indexName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"CREATING", "INITIALIZING"}, // pending states
		[]string{"ACTIVE"},                   // target states
		timeout,
		5*time.Second,
		s.OtsIndexStateRefreshFunc(instanceName, tableName, indexName, []string{"FAILED"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, fmt.Sprintf("%s:%s:%s", instanceName, tableName, indexName))
}

func (s *OtsService) WaitForOtsIndexDeleting(instanceName, tableName, indexName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{string(tablestoreAPI.TablestoreIndexStatusExisting)}, // pending states
		[]string{string(tablestoreAPI.TablestoreIndexStatusNotFound)}, // target states
		timeout,
		5*time.Second,
		s.OtsIndexStateRefreshFunc(instanceName, tableName, indexName, []string{"FAILED"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, fmt.Sprintf("%s:%s:%s", instanceName, tableName, indexName))
}

func (s *OtsService) OtsIndexStateRefreshFunc(instanceName, tableName, indexName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOtsIndex(instanceName, tableName, indexName)
		if err != nil {
			if NotFoundError(err) {
				return nil, string(tablestoreAPI.TablestoreIndexStatusNotFound), nil
			}
			return nil, "", WrapError(err)
		}

		// For secondary indexes, we assume they're active if they exist
		currentStatus := string(tablestoreAPI.TablestoreIndexStatusExisting)

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

func (s *OtsService) ListOtsIndex(instanceName, tableName string) ([]*tablestoreSDK.IndexMeta, error) {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	indexes, err := api.ListIndexes(instanceName, tableName)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, tableName, "ListIndexes", AlibabaCloudSdkGoERROR)
	}

	// Convert to IndexMeta slice using the correct SDK type
	var result []*tablestoreSDK.IndexMeta
	for _, index := range indexes {
		if index.IndexMeta != nil {
			result = append(result, index.IndexMeta)
		}
	}

	return result, nil
}

func (s *OtsService) ListOtsIndexes(instanceName, tableName string) ([]*tablestoreAPI.TablestoreIndex, error) {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	indexes, err := api.ListIndexes(instanceName, tableName)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, tableName, "ListIndexes", AlibabaCloudSdkGoERROR)
	}

	return indexes, nil
}

// EncodeOtsIndexId encodes instance name, table name, and index name into a single ID string
// Format: instanceName:tableName:indexName
func EncodeOtsIndexId(instanceName, tableName, indexName string) string {
	return fmt.Sprintf("%s:%s:%s", instanceName, tableName, indexName)
}

// DecodeOtsIndexId parses index ID string into instance name, table name, and index name components
func DecodeOtsIndexId(id string) (string, string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid index ID format, expected instanceName:tableName:indexName, got %s", id)
	}
	return parts[0], parts[1], parts[2], nil
}
