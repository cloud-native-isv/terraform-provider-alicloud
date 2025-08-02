package alicloud

import (
	"fmt"
	"strings"
	"time"

	tablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Table management functions

func (s *OtsService) CreateOtsTable(instanceName string, table *tablestoreAPI.TablestoreTable) error {
	api := s.GetAPI()

	if err := api.CreateTable(instanceName, table); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, table.GetName(), "CreateTable", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) DescribeOtsTable(instanceName, tableName string) (*tablestoreAPI.TablestoreTable, error) {
	api := s.GetAPI()

	table, err := api.GetTable(instanceName, tableName)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, tableName, "GetTable", AlibabaCloudSdkGoERROR)
	}

	return table, nil
}

func (s *OtsService) UpdateOtsTable(instanceName string, table *tablestoreAPI.TablestoreTable) error {
	api := s.GetAPI()

	if err := api.UpdateTable(instanceName, table); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, table.GetName(), "UpdateTable", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) DeleteOtsTable(instanceName, tableName string) error {
	api := s.GetAPI()

	if err := api.DeleteTable(instanceName, tableName); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, tableName, "DeleteTable", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) ListOtsTables(instanceName string) ([]*tablestoreAPI.TablestoreTable, error) {
	api := s.GetAPI()

	tables, err := api.ListTables(instanceName)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, instanceName, "ListTables", AlibabaCloudSdkGoERROR)
	}

	return tables, nil
}

// State refresh function for table
func (s *OtsService) OtsTableStateRefreshFunc(instanceName, tableName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeOtsTable(instanceName, tableName)
		if err != nil {
			if IsNotFoundError(err) {
				// 资源已删除，返回 nil, "", nil 表示目标达成
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// Use the new status type
		currentStatus := string(tablestoreAPI.TablestoreTableStatusExisting)

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

// Wait for table creation
func (s *OtsService) WaitForOtsTableCreating(instanceName, tableName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{string(tablestoreAPI.TablestoreTableStatusNotFound)}, // pending states
		[]string{string(tablestoreAPI.TablestoreTableStatusExisting)}, // target states
		timeout,
		5*time.Second,
		s.OtsTableStateRefreshFunc(instanceName, tableName, []string{string(tablestoreAPI.TablestoreTableStatusFailed)}),
	)

	_, err := stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, tableName)
	}
	return nil
}

// Wait for table deletion
func (s *OtsService) WaitForOtsTableDeleting(instanceName, tableName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{string(tablestoreAPI.TablestoreTableStatusExisting)}, // pending states
		[]string{}, // target states (empty = wait for resource disappear)
		timeout,
		5*time.Second,
		s.OtsTableStateRefreshFunc(instanceName, tableName, []string{string(tablestoreAPI.TablestoreTableStatusFailed)}),
	)

	_, err := stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, tableName)
	}
	return nil
}

// EncodeOtsTableId encodes instance name and table name into a single ID string
// Format: instanceName:tableName
func EncodeOtsTableId(instanceName, tableName string) string {
	return fmt.Sprintf("%s:%s", instanceName, tableName)
}

// DecodeOtsTableId parses table ID string into instance name and table name components
func DecodeOtsTableId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid table ID format, expected instanceName:tableName, got %s", id)
	}
	return parts[0], parts[1], nil
}
