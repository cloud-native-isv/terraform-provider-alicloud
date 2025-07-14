package alicloud

import (
	"context"
	"fmt"
	"time"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Secondary Index management functions

func (s *OtsService) CreateOtsSecondaryIndex(d *schema.ResourceData, instanceName, tableName string) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	indexName := d.Get("index_name").(string)

	// Build primary key columns for the index
	primaryKeyList := d.Get("primary_key").([]interface{})
	var primaryKeys []tablestore.PrimaryKeyColumn
	for _, pk := range primaryKeyList {
		pkMap := pk.(map[string]interface{})
		primaryKeys = append(primaryKeys, tablestore.PrimaryKeyColumn{
			Name: pkMap["name"].(string),
			Type: pkMap["type"].(string),
		})
	}

	options := &tablestore.CreateSecondaryIndexOptions{
		TableName: tableName,
		IndexName: indexName,
		IndexMeta: tablestore.IndexMeta{
			IndexName:   indexName,
			PrimaryKeys: primaryKeys,
			IndexType:   tablestore.IT_GLOBAL_INDEX, // Default to global index
		},
	}

	// Set defined columns if provided
	if definedColumns, ok := d.GetOk("defined_column"); ok {
		var columns []string
		for _, col := range definedColumns.([]interface{}) {
			columns = append(columns, col.(string))
		}
		options.IndexMeta.DefinedColumns = columns
	}

	// Set index update mode if provided
	if indexUpdateMode, ok := d.GetOk("index_update_mode"); ok {
		switch indexUpdateMode.(string) {
		case "IUM_ASYNC_INDEX":
			options.IndexMeta.IndexUpdateMode = tablestore.IUM_ASYNC_INDEX
		case "IUM_SYNC_INDEX":
			options.IndexMeta.IndexUpdateMode = tablestore.IUM_SYNC_INDEX
		}
	}

	ctx := context.Background()
	if err := api.CreateSecondaryIndex(ctx, options); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, indexName, "CreateSecondaryIndex", AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", instanceName, tableName, indexName))
	return nil
}

func (s *OtsService) DescribeOtsSecondaryIndex(instanceName, tableName, indexName string) (*tablestore.IndexMeta, error) {
	// Get table information which includes index metadata
	table, err := s.DescribeOtsTable(instanceName, tableName)
	if err != nil {
		return nil, WrapError(err)
	}

	// Find the specific index
	for _, indexMeta := range table.IndexMetas {
		if indexMeta.IndexName == indexName {
			return &indexMeta, nil
		}
	}

	return nil, WrapErrorf(Error("Secondary index not found"), NotFoundMsg, AlibabaCloudSdkGoERROR)
}

func (s *OtsService) DeleteOtsSecondaryIndex(instanceName, tableName, indexName string) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	options := &tablestore.DeleteSecondaryIndexOptions{
		TableName: tableName,
		IndexName: indexName,
	}

	ctx := context.Background()
	if err := api.DeleteSecondaryIndex(ctx, options); err != nil {
		if IsExpectedErrors(err, []string{"NotExist", "InvalidIndexName.NotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, indexName, "DeleteSecondaryIndex", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) WaitForOtsSecondaryIndex(instanceName, tableName, indexName string, status string, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		index, err := s.DescribeOtsSecondaryIndex(instanceName, tableName, indexName)
		if err != nil {
			if NotFoundError(err) {
				if status == string(Deleted) {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}

		if index != nil && string(index.IndexStatus) == status {
			return nil
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, indexName, GetFunc(1), timeout, string(index.IndexStatus), status, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

func (s *OtsService) ListOtsSecondaryIndexes(instanceName, tableName string) ([]tablestore.IndexMeta, error) {
	table, err := s.DescribeOtsTable(instanceName, tableName)
	if err != nil {
		return nil, WrapError(err)
	}

	return table.IndexMetas, nil
}
