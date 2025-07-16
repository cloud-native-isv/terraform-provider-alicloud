package alicloud

import (
	"fmt"
	"strings"
	"time"

	tablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

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

func (s *OtsService) CreateOtsTable(instanceName string, table *tablestoreAPI.TablestoreTable) error {
	// Validate input parameters
	if instanceName == "" {
		return fmt.Errorf("instanceName cannot be empty")
	}
	if table == nil {
		return fmt.Errorf("table cannot be nil")
	}

	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	// Set the instance name for the table
	table.SetInstanceName(instanceName)

	if err := api.CreateTable(instanceName, table); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, table.GetName(), "CreateTable", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) DescribeOtsTable(id string) (*tablestoreAPI.TablestoreTable, error) {
	// Validate input parameter
	if id == "" {
		return nil, fmt.Errorf("table ID cannot be empty")
	}

	// Parse table ID to extract instance name and table name
	instanceName, tableName, err := DecodeOtsTableId(id)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "DecodeOtsTableId", AlibabaCloudSdkGoERROR)
	}

	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	table, err := api.GetTable(instanceName, tableName)
	if err != nil {
		if IsExpectedErrors(err, []string{"NotExist", "InvalidTableName.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, tableName, "GetTable", AlibabaCloudSdkGoERROR)
	}

	return table, nil
}

func (s *OtsService) UpdateOtsTable(instanceName string, table *tablestoreAPI.TablestoreTable) error {
	// Validate input parameters
	if instanceName == "" {
		return fmt.Errorf("instanceName cannot be empty")
	}
	if table == nil {
		return fmt.Errorf("table cannot be nil")
	}

	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	// Set the instance name for the table
	table.SetInstanceName(instanceName)

	if err := api.UpdateTable(instanceName, table); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, table.GetName(), "UpdateTable", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) DeleteOtsTable(id string) error {
	// Validate input parameter
	if id == "" {
		return fmt.Errorf("table ID cannot be empty")
	}

	// Parse table ID to extract instance name and table name
	instanceName, tableName, err := DecodeOtsTableId(id)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, id, "DecodeOtsTableId", AlibabaCloudSdkGoERROR)
	}

	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	if err := api.DeleteTable(instanceName, tableName); err != nil {
		if IsExpectedErrors(err, []string{"NotExist", "InvalidTableName.NotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, tableName, "DeleteTable", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) ListOtsTable(instanceName string) ([]*tablestoreAPI.TablestoreTable, error) {
	// Validate input parameter
	if instanceName == "" {
		return nil, fmt.Errorf("instanceName cannot be empty")
	}

	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	tables, err := api.ListTables(instanceName)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, instanceName, "ListTables", AlibabaCloudSdkGoERROR)
	}

	return tables, nil
}

func (s *OtsService) ListOtsTables(instanceName string) ([]*tablestoreAPI.TablestoreTable, error) {
	return s.ListOtsTable(instanceName)
}

func (s *OtsService) OtsTableStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		table, err := s.DescribeOtsTable(id)
		if err != nil {
			if NotFoundError(err) {
				// For deletion scenarios, return nil to indicate resource absence
				return nil, "", nil
			}
			return nil, "Failed", WrapErrorf(err, DefaultErrorMsg, id, "DescribeOtsTable", AlibabaCloudSdkGoERROR)
		}

		// If table is nil, it means the resource doesn't exist
		if table == nil {
			return nil, "", nil
		}

		// Check for failed states
		for _, failState := range failStates {
			if table.Status == failState {
				return table, table.Status, WrapError(Error(FailedToReachTargetStatus, table.Status))
			}
		}

		return table, table.Status, nil
	}
}

func (s *OtsService) WaitForOtsTableCreating(id string, timeout time.Duration) error {
	createPendingStatus := []string{"Creating", "Updating"}
	createExpectStatus := []string{"Active"}
	createFailedStatus := []string{"Failed"}

	stateConf := BuildStateConf(
		createPendingStatus,
		createExpectStatus,
		timeout,
		5*time.Second,
		s.OtsTableStateRefreshFunc(id, createFailedStatus),
	)

	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}

	return nil
}

func (s *OtsService) WaitForOtsTableDeleting(id string, timeout time.Duration) error {
	deletePendingStatus := []string{"Active", "Deleting"}
	deleteExpectStatus := []string{}
	deleteFailedStatus := []string{"Failed"}

	stateConf := BuildStateConf(
		deletePendingStatus,
		deleteExpectStatus,
		timeout,
		5*time.Second,
		s.OtsTableStateRefreshFunc(id, deleteFailedStatus),
	)

	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}

	return nil
}

// Deprecated: use WaitForOtsTableCreating instead
func (s *OtsService) WaitForOtsTable(id string, status string, timeout time.Duration) error {
	if status == "Active" {
		return s.WaitForOtsTableCreating(id, timeout)
	}
	if status == string(Deleted) {
		return s.WaitForOtsTableDeleting(id, timeout)
	}

	// Legacy implementation for other statuses
	deadline := time.Now().Add(timeout)
	for {
		table, err := s.DescribeOtsTable(id)
		if err != nil {
			if NotFoundError(err) {
				if status == string(Deleted) {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}

		if table != nil && table.Status == status {
			return nil
		}

		if time.Now().After(deadline) {
			currentStatus := ""
			if table != nil {
				currentStatus = table.Status
			}
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), int(timeout.Seconds()), currentStatus, status, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}
}
