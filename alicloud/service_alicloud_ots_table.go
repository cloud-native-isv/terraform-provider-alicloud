package alicloud

import (
	"fmt"
	"strings"
	"time"

	tablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func parseTableId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid table ID format, expected instanceName:tableName, got %s", id)
	}
	return parts[0], parts[1], nil
}

// Table management functions

func (s *OtsService) CreateOtsTable(d *schema.ResourceData, instanceName string) error {
	// Validate input parameters
	if instanceName == "" {
		return fmt.Errorf("instanceName cannot be empty")
	}

	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	tableName := d.Get("table_name").(string)
	if tableName == "" {
		return fmt.Errorf("table_name cannot be empty")
	}

	// Build primary key schema
	primaryKeyList := d.Get("primary_key").([]interface{})
	if len(primaryKeyList) == 0 {
		return fmt.Errorf("primary_key must be specified")
	}

	var primaryKeys []tablestoreAPI.PrimaryKeyColumn
	for _, pk := range primaryKeyList {
		pkMap := pk.(map[string]interface{})
		primaryKeys = append(primaryKeys, tablestoreAPI.PrimaryKeyColumn{
			Name: pkMap["name"].(string),
			Type: pkMap["type"].(string),
		})
	}

	options := &tablestoreAPI.CreateTableOptions{
		InstanceName: instanceName,
		TableName:    tableName,
		PrimaryKeys:  primaryKeys,
		TableOption: tablestoreAPI.TableOption{
			TimeToLive:                int32(d.Get("time_to_live").(int)),
			MaxVersions:               int32(d.Get("max_version").(int)),
			DeviationCellVersionInSec: int64(d.Get("deviation_cell_version_in_sec").(int)),
		},
		ReservedThroughput: tablestoreAPI.ReservedThroughput{
			ReadCapacity:  int32(d.Get("read_capacity").(int)),
			WriteCapacity: int32(d.Get("write_capacity").(int)),
		},
	}

	// Set defined columns if provided
	if definedColumns, ok := d.GetOk("defined_column"); ok {
		var columns []tablestoreAPI.DefinedColumn
		for _, col := range definedColumns.([]interface{}) {
			colMap := col.(map[string]interface{})
			columns = append(columns, tablestoreAPI.DefinedColumn{
				Name: colMap["name"].(string),
				Type: colMap["type"].(string),
			})
		}
		options.DefinedColumns = columns
	}

	// Set stream specification if provided
	if enableStream, ok := d.GetOk("enable_sse"); ok && enableStream.(bool) {
		options.StreamSpec = &tablestoreAPI.StreamSpecification{
			EnableStream:   true,
			ExpirationTime: int32(d.Get("sse_key_type").(int)), // This mapping might need adjustment
		}
	}

	if err := api.CreateTable(options); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, tableName, "CreateTable", AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceName, tableName))
	return nil
}

func (s *OtsService) DescribeOtsTable(id string) (*tablestoreAPI.TablestoreTable, error) {
	// Validate input parameter
	if id == "" {
		return nil, fmt.Errorf("table ID cannot be empty")
	}

	// Parse table ID to extract instance name and table name
	// Format: instanceName:tableName
	instanceName, tableName, err := parseTableId(id)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "parseTableId", AlibabaCloudSdkGoERROR)
	}

	table, err := s.tablestoreAPI.GetTable(tableName)
	if err != nil {
		if IsExpectedErrors(err, []string{"NotExist", "InvalidTableName.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, tableName, "GetTable", AlibabaCloudSdkGoERROR)
	}

	return table, nil
}

func (s *OtsService) UpdateOtsTable(d *schema.ResourceData, id string) error {
	// Validate input parameter
	if id == "" {
		return fmt.Errorf("table ID cannot be empty")
	}

	// Parse table ID to extract instance name and table name
	instanceName, tableName, err := parseTableId(id)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, id, "parseTableId", AlibabaCloudSdkGoERROR)
	}

	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	options := &tablestoreAPI.UpdateTableOptions{}
	update := false

	if d.HasChange("time_to_live") || d.HasChange("max_version") || d.HasChange("deviation_cell_version_in_sec") {
		options.TableOption = &tablestoreAPI.TableOption{
			TimeToLive:                int32(d.Get("time_to_live").(int)),
			MaxVersions:               int32(d.Get("max_version").(int)),
			DeviationCellVersionInSec: int64(d.Get("deviation_cell_version_in_sec").(int)),
		}
		update = true
	}

	if d.HasChange("read_capacity") || d.HasChange("write_capacity") {
		options.ReservedThroughput = &tablestoreAPI.ReservedThroughput{
			ReadCapacity:  int32(d.Get("read_capacity").(int)),
			WriteCapacity: int32(d.Get("write_capacity").(int)),
		}
		update = true
	}

	if update {
		if err := api.UpdateTable(tableName, options); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, tableName, "UpdateTable", AlibabaCloudSdkGoERROR)
		}
	}

	return nil
}

func (s *OtsService) DeleteOtsTable(id string) error {
	// Validate input parameter
	if id == "" {
		return fmt.Errorf("table ID cannot be empty")
	}

	// Parse table ID to extract instance name and table name
	instanceName, tableName, err := parseTableId(id)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, id, "parseTableId", AlibabaCloudSdkGoERROR)
	}

	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	if err := api.DeleteTable(tableName); err != nil {
		if IsExpectedErrors(err, []string{"NotExist", "InvalidTableName.NotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, tableName, "DeleteTable", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) WaitForOtsTable(id string, status string, timeout time.Duration) error {
	// Validate input parameters
	if id == "" {
		return fmt.Errorf("table ID cannot be empty")
	}
	if status == "" {
		return fmt.Errorf("status cannot be empty")
	}

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

		if table != nil && table.TableStatus == status {
			return nil
		}

		if time.Now().After(deadline) {
			currentStatus := ""
			if table != nil {
				currentStatus = table.TableStatus
			}
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), int(timeout.Seconds()), currentStatus, status, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

func (s *OtsService) ListOtsTable(instanceName string) ([]*tablestoreAPI.TablestoreTable, error) {
	// Validate input parameter
	if instanceName == "" {
		return nil, fmt.Errorf("instanceName cannot be empty")
	}

	tables, err := s.tablestoreAPI.ListTables()
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, instanceName, "ListTables", AlibabaCloudSdkGoERROR)
	}

	return tables, nil
}

func (s *OtsService) ListOtsTables(instanceName string) ([]*tablestoreAPI.TablestoreTable, error) {
	// Validate input parameter
	if instanceName == "" {
		return nil, fmt.Errorf("instanceName cannot be empty")
	}

	tables, err := s.tablestoreAPI.ListTables()
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, instanceName, "ListTables", AlibabaCloudSdkGoERROR)
	}

	return tables, nil
}

func (s *OtsService) OtsTableStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		table, err := s.DescribeOtsTable(id)
		if err != nil {
			if NotFoundError(err) {
				// For deletion scenarios, return nil to indicate resource absence
				// This allows WaitForState to properly handle the "waiting for absence" case
				return nil, "", nil
			}
			return nil, "Failed", WrapErrorf(err, DefaultErrorMsg, id, "DescribeOtsTable", AlibabaCloudSdkGoERROR)
		}

		// If table is nil, it means the resource doesn't exist
		if table == nil {
			// For deletion scenarios, return nil to indicate resource absence
			return nil, "", nil
		}

		return table, table.TableStatus, nil
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
