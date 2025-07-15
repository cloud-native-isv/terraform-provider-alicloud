package alicloud

import (
	"fmt"
	"time"

	tablestoreSDK "github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	tablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)


// Table management functions

func (s *OtsService) CreateOtsTable(d *schema.ResourceData, instanceName string) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	tableName := d.Get("table_name").(string)

	// Build primary key schema
	primaryKeyList := d.Get("primary_key").([]interface{})
	var primaryKeys []tablestore.PrimaryKeyColumn
	for _, pk := range primaryKeyList {
		pkMap := pk.(map[string]interface{})
		primaryKeys = append(primaryKeys, tablestore.PrimaryKeyColumn{
			Name: pkMap["name"].(string),
			Type: pkMap["type"].(string),
		})
	}

	options := &tablestore.CreateTableOptions{
		InstanceName: instanceName,
		TableName:    tableName,
		PrimaryKeys:  primaryKeys,
		TableOption: tablestore.TableOption{
			TimeToLive:                int32(d.Get("time_to_live").(int)),
			MaxVersions:               int32(d.Get("max_version").(int)),
			DeviationCellVersionInSec: int64(d.Get("deviation_cell_version_in_sec").(int)),
		},
		ReservedThroughput: tablestore.ReservedThroughput{
			ReadCapacity:  int32(d.Get("read_capacity").(int)),
			WriteCapacity: int32(d.Get("write_capacity").(int)),
		},
	}

	// Set defined columns if provided
	if definedColumns, ok := d.GetOk("defined_column"); ok {
		var columns []tablestore.DefinedColumn
		for _, col := range definedColumns.([]interface{}) {
			colMap := col.(map[string]interface{})
			columns = append(columns, tablestore.DefinedColumn{
				Name: colMap["name"].(string),
				Type: colMap["type"].(string),
			})
		}
		options.DefinedColumns = columns
	}

	// Set stream specification if provided
	if enableStream, ok := d.GetOk("enable_sse"); ok && enableStream.(bool) {
		options.StreamSpec = &tablestore.StreamSpecification{
			EnableStream:   true,
			ExpirationTime: int32(d.Get("sse_key_type").(int)), // This mapping might need adjustment
		}
	}

	if err := s.tablestoreAPI.CreateTable(options); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, tableName, "CreateTable", AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceName, tableName))
	return nil
}

func (s *OtsService) DescribeOtsTable(instanceName, tableName string) (*tablestoreAPI.TablestoreTable, error) {
	table, err := s.tablestoreAPI.GetTable(tableName)
	if err != nil {
		if IsExpectedErrors(err, []string{"NotExist", "InvalidTableName.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, tableName, "GetTable", AlibabaCloudSdkGoERROR)
	}

	return table, nil
}

func (s *OtsService) UpdateOtsTable(d *schema.ResourceData, instanceName, tableName string) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	options := &tablestore.UpdateTableOptions{}
	update := false

	if d.HasChange("time_to_live") || d.HasChange("max_version") || d.HasChange("deviation_cell_version_in_sec") {
		options.TableOption = &tablestore.TableOption{
			TimeToLive:                int32(d.Get("time_to_live").(int)),
			MaxVersions:               int32(d.Get("max_version").(int)),
			DeviationCellVersionInSec: int64(d.Get("deviation_cell_version_in_sec").(int)),
		}
		update = true
	}

	if d.HasChange("read_capacity") || d.HasChange("write_capacity") {
		options.ReservedThroughput = &tablestore.ReservedThroughput{
			ReadCapacity:  int32(d.Get("read_capacity").(int)),
			WriteCapacity: int32(d.Get("write_capacity").(int)),
		}
		update = true
	}

	if update {
		if err := s.tablestoreAPI.UpdateTable(tableName, options); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, tableName, "UpdateTable", AlibabaCloudSdkGoERROR)
		}
	}

	return nil
}

func (s *OtsService) DeleteOtsTable(instanceName, tableName string) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	if err := s.tablestoreAPI.DeleteTable(tableName); err != nil {
		if IsExpectedErrors(err, []string{"NotExist", "InvalidTableName.NotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, tableName, "DeleteTable", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) WaitForOtsTable(instanceName, tableName string, status string, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		table, err := s.DescribeOtsTable(instanceName, tableName)
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
			return WrapErrorf(err, WaitTimeoutMsg, tableName, GetFunc(1), timeout, table.TableStatus, status, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

func (s *OtsService) ListOtsTable(instanceName string) ([]*tablestoreAPI.TablestoreTable, error) {
	tables, err := s.tablestoreAPI.ListTables()
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, instanceName, "ListTables", AlibabaCloudSdkGoERROR)
	}

	return tables, nil
}

func (s *OtsService) ListOtsTables(instanceName string) ([]string, error) {
	tables, err := s.tablestoreAPI.ListTables()
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, instanceName, "ListTables", AlibabaCloudSdkGoERROR)
	}

	return tables, nil
}
