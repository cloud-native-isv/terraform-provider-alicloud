package alicloud

import (
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

	// Build primary keys
	var primaryKeys []string
	if pks, ok := d.GetOk("primary_key"); ok {
		for _, pk := range pks.([]interface{}) {
			primaryKeys = append(primaryKeys, pk.(string))
		}
	}

	// Build defined columns
	var definedColumns []string
	if dcs, ok := d.GetOk("defined_column"); ok {
		for _, dc := range dcs.([]interface{}) {
			definedColumns = append(definedColumns, dc.(string))
		}
	}

	// Create index metadata
	indexMeta := &tablestore.IndexMeta{
		IndexName:      indexName,
		Primarykey:     primaryKeys,
		DefinedColumns: definedColumns,
		IndexType:      tablestore.IT_GLOBAL_INDEX, // Default to global index
	}

	// Create index object
	index := &tablestore.TablestoreIndex{
		TableName:       tableName,
		IndexName:       indexName,
		IndexMeta:       indexMeta,
		IncludeBaseData: d.Get("include_base_data").(bool),
	}

	if err := api.CreateIndex(index); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, indexName, "CreateIndex", AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", instanceName, tableName, indexName))
	return nil
}

func (s *OtsService) DescribeOtsSecondaryIndex(instanceName, tableName, indexName string) (*tablestore.TablestoreIndex, error) {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	index, err := api.GetIndex(tableName, indexName)
	if err != nil {
		if IsExpectedErrors(err, []string{"NotExist", "InvalidIndexName.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, indexName, "DescribeIndex", AlibabaCloudSdkGoERROR)
	}

	return index, nil
}

func (s *OtsService) DeleteOtsSecondaryIndex(instanceName, tableName, indexName string) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	index := &tablestore.TablestoreIndex{
		TableName: tableName,
		IndexName: indexName,
	}

	if err := api.DeleteIndex(index); err != nil {
		if IsExpectedErrors(err, []string{"NotExist", "InvalidIndexName.NotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, indexName, "DeleteIndex", AlibabaCloudSdkGoERROR)
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

		if index != nil {
			// For secondary indexes, we assume they're active if they exist
			if status == "Active" {
				return nil
			}
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, indexName, GetFunc(1), timeout, "Active", status, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

func (s *OtsService) ListOtsSecondaryIndexes(instanceName, tableName string) ([]*tablestore.TablestoreIndex, error) {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	indexes, err := api.ListIndexes(tableName)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, tableName, "ListIndexes", AlibabaCloudSdkGoERROR)
	}

	return indexes, nil
}
