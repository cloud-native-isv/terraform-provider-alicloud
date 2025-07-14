package alicloud

import (
	"context"
	"fmt"
	"time"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Search Index management functions

func (s *OtsService) CreateOtsSearchIndex(d *schema.ResourceData, instanceName, tableName string) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	indexName := d.Get("index_name").(string)

	// Build field schemas
	var fieldSchemas []tablestore.FieldSchema
	if schemas, ok := d.GetOk("schema"); ok {
		for _, schema := range schemas.([]interface{}) {
			schemaMap := schema.(map[string]interface{})
			fieldSchema := tablestore.FieldSchema{
				FieldName: schemaMap["field_name"].(string),
				FieldType: schemaMap["field_type"].(string),
			}

			if index, exists := schemaMap["index"]; exists {
				fieldSchema.Index = index.(bool)
			}

			if store, exists := schemaMap["store"]; exists {
				fieldSchema.Store = store.(bool)
			}

			if enableSortAndAgg, exists := schemaMap["enable_sort_and_agg"]; exists {
				fieldSchema.EnableSortAndAgg = enableSortAndAgg.(bool)
			}

			if analyzer, exists := schemaMap["analyzer"]; exists && analyzer.(string) != "" {
				fieldSchema.Analyzer = analyzer.(string)
			}

			fieldSchemas = append(fieldSchemas, fieldSchema)
		}
	}

	options := &tablestore.CreateSearchIndexOptions{
		TableName: tableName,
		IndexName: indexName,
		IndexSchema: tablestore.IndexSchema{
			FieldSchemas: fieldSchemas,
		},
	}

	// Set index setting if provided
	if setting, ok := d.GetOk("index_setting"); ok {
		settingMap := setting.(map[string]interface{})
		indexSetting := tablestore.IndexSetting{}

		if routingFields, exists := settingMap["routing_fields"]; exists {
			indexSetting.RoutingFields = expandStringList(routingFields.([]interface{}))
		}

		options.IndexSchema.IndexSetting = &indexSetting
	}

	// Set index sort if provided
	if sorts, ok := d.GetOk("index_sort"); ok {
		var indexSorts []tablestore.IndexSort
		for _, sort := range sorts.([]interface{}) {
			sortMap := sort.(map[string]interface{})
			indexSort := tablestore.IndexSort{
				FieldName: sortMap["field_name"].(string),
				Order:     sortMap["order"].(string),
			}
			indexSorts = append(indexSorts, indexSort)
		}
		options.IndexSchema.IndexSort = indexSorts
	}

	ctx := context.Background()
	if err := api.CreateSearchIndex(ctx, options); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, indexName, "CreateSearchIndex", AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", instanceName, tableName, indexName))
	return nil
}

func (s *OtsService) DescribeOtsSearchIndex(instanceName, tableName, indexName string) (*tablestore.SearchIndexInfo, error) {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	options := &tablestore.DescribeSearchIndexOptions{
		TableName: tableName,
		IndexName: indexName,
	}

	ctx := context.Background()
	index, err := api.DescribeSearchIndex(ctx, options)
	if err != nil {
		if IsExpectedErrors(err, []string{"NotExist", "InvalidIndexName.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, indexName, "DescribeSearchIndex", AlibabaCloudSdkGoERROR)
	}

	return index, nil
}

func (s *OtsService) UpdateOtsSearchIndex(d *schema.ResourceData, instanceName, tableName, indexName string) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	// Build updated field schemas
	var fieldSchemas []tablestore.FieldSchema
	if schemas, ok := d.GetOk("schema"); ok {
		for _, schema := range schemas.([]interface{}) {
			schemaMap := schema.(map[string]interface{})
			fieldSchema := tablestore.FieldSchema{
				FieldName: schemaMap["field_name"].(string),
				FieldType: schemaMap["field_type"].(string),
			}

			if index, exists := schemaMap["index"]; exists {
				fieldSchema.Index = index.(bool)
			}

			if store, exists := schemaMap["store"]; exists {
				fieldSchema.Store = store.(bool)
			}

			if enableSortAndAgg, exists := schemaMap["enable_sort_and_agg"]; exists {
				fieldSchema.EnableSortAndAgg = enableSortAndAgg.(bool)
			}

			if analyzer, exists := schemaMap["analyzer"]; exists && analyzer.(string) != "" {
				fieldSchema.Analyzer = analyzer.(string)
			}

			fieldSchemas = append(fieldSchemas, fieldSchema)
		}
	}

	options := &tablestore.UpdateSearchIndexOptions{
		TableName: tableName,
		IndexName: indexName,
		IndexSchema: tablestore.IndexSchema{
			FieldSchemas: fieldSchemas,
		},
	}

	ctx := context.Background()
	if err := api.UpdateSearchIndex(ctx, options); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, indexName, "UpdateSearchIndex", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) DeleteOtsSearchIndex(instanceName, tableName, indexName string) error {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return WrapError(err)
	}

	options := &tablestore.DeleteSearchIndexOptions{
		TableName: tableName,
		IndexName: indexName,
	}

	ctx := context.Background()
	if err := api.DeleteSearchIndex(ctx, options); err != nil {
		if IsExpectedErrors(err, []string{"NotExist", "InvalidIndexName.NotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, indexName, "DeleteSearchIndex", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func (s *OtsService) WaitForOtsSearchIndex(instanceName, tableName, indexName string, status string, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		index, err := s.DescribeOtsSearchIndex(instanceName, tableName, indexName)
		if err != nil {
			if NotFoundError(err) {
				if status == string(Deleted) {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}

		if index != nil && index.SyncPhase == status {
			return nil
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, indexName, GetFunc(1), timeout, index.SyncPhase, status, ProviderERROR)
		}
		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

func (s *OtsService) ListOtsSearchIndexes(instanceName, tableName string) ([]tablestore.SearchIndexInfo, error) {
	api, err := s.getTablestoreAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	options := &tablestore.ListSearchIndexOptions{
		TableName: tableName,
	}

	ctx := context.Background()
	result, err := api.ListSearchIndex(ctx, options)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, tableName, "ListSearchIndex", AlibabaCloudSdkGoERROR)
	}

	return result.Indexes, nil
}
