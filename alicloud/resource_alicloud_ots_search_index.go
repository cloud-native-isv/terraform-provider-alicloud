package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore/search"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	tablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudOtsSearchIndex() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudOtsSearchIndexCreate,
		Read:   resourceAliCloudOtsSearchIndexRead,
		Update: resourceAliCloudOtsSearchIndexUpdate,
		Delete: resourceAliCloudOtsSearchIndexDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"instance_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateOTSInstanceName,
				Description:  "The name of the OTS instance.",
			},
			"table_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateOTSTableName,
				Description:  "The name of the table.",
			},
			"index_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateOTSIndexName,
				Description:  "The name of the search index.",
			},
			"source_index_name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of the source index.",
			},
			"time_to_live": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      -1,
				ValidateFunc: validation.Any(validation.IntInSlice([]int{-1}), validation.IntAtLeast(86400)),
				Description:  "The time to live in seconds. -1 means never expire.",
			},
			"field_schemas": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "The field schemas of the search index.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field_name": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The name of the field.",
						},
						"field_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								tablestore.FieldType_LONG.String(),
								tablestore.FieldType_DOUBLE.String(),
								tablestore.FieldType_BOOLEAN.String(),
								tablestore.FieldType_KEYWORD.String(),
								tablestore.FieldType_TEXT.String(),
								tablestore.FieldType_NESTED.String(),
								tablestore.FieldType_GEO_POINT.String(),
								tablestore.FieldType_DATE.String(),
								tablestore.FieldType_VECTOR.String(),
							}, false),
							Description: "The type of the field.",
						},
						"is_array": {
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
							Description: "Whether the field is an array.",
						},
						"index": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							ForceNew:    true,
							Description: "Whether to create an index for the field.",
						},
						"analyzer": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(tablestore.Analyzer_SingleWord),
								string(tablestore.Analyzer_MaxWord),
								string(tablestore.Analyzer_MinWord),
								string(tablestore.Analyzer_Split),
								string(tablestore.Analyzer_Fuzzy),
							}, false),
							Description: "The analyzer for the field.",
						},
						"enable_sort_and_agg": {
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
							Description: "Whether to enable sorting and aggregation for the field.",
						},
						"store": {
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
							Description: "Whether to store the field value.",
						},
					},
				},
			},
			"routing_fields": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The routing fields for the search index.",
			},
			"index_sort": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				MaxItems:    1,
				Description: "The sorting configuration for the search index.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sorters": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"sorter_type": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  string(OtsSearchPrimaryKeySort),
										ForceNew: true,
										ValidateFunc: validation.StringInSlice([]string{
											string(OtsSearchPrimaryKeySort),
											string(OtsSearchFieldSort),
										}, false),
									},
									"order": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  string(OtsSearchSortOrderAsc),
										ForceNew: true,
										ValidateFunc: validation.StringInSlice([]string{
											string(OtsSearchSortOrderAsc),
											string(OtsSearchSortOrderDesc),
										}, false),
									},
									"field_name": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"mode": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.StringInSlice([]string{
											string(OtsSearchModeMin),
											string(OtsSearchModeMax),
											string(OtsSearchModeAvg),
										}, false),
									},
								},
							},
						},
					},
				},
			},
			// Computed fields
			"sync_phase": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The synchronization phase of the search index.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the search index.",
			},
		},
	}
}

func resourceAliCloudOtsSearchIndexCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceName := d.Get("instance_name").(string)
	tableName := d.Get("table_name").(string)
	indexName := d.Get("index_name").(string)

	// Build TablestoreSearchIndex from schema
	index := &tablestoreAPI.TablestoreSearchIndex{
		TableName: tableName,
		IndexName: indexName,
	}

	// Set source index name if provided
	if sourceIndexName, ok := d.GetOk("source_index_name"); ok {
		sourceIndexNameStr := sourceIndexName.(string)
		index.SourceIndexName = &sourceIndexNameStr
	}

	// Set TTL if provided
	if ttl, ok := d.GetOk("time_to_live"); ok {
		ttlInt32 := int32(ttl.(int))
		index.TimeToLive = &ttlInt32
	}

	// Build IndexSchema
	indexSchema, err := buildIndexSchemaFromResource(d)
	if err != nil {
		return WrapError(err)
	}
	index.IndexSchema = indexSchema

	// Create search index
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		if err := otsService.CreateOtsSearchIndex(instanceName, index); err != nil {
			if IsExpectedErrors(err, []string{"ThrottlingException", "ServiceUnavailable"}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_ots_search_index", "CreateSearchIndex", AlibabaCloudSdkGoERROR)
	}

	d.SetId(EncodeOtsSearchIndexId(instanceName, tableName, indexName))

	// Wait for search index to be ready
	if err := otsService.WaitForOtsSearchIndexCreating(instanceName, tableName, indexName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudOtsSearchIndexRead(d, meta)
}

func resourceAliCloudOtsSearchIndexRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceName, tableName, indexName, err := DecodeOtsSearchIndexId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	index, err := otsService.DescribeOtsSearchIndex(instanceName, tableName, indexName)
	if err != nil {
		if !d.IsNewResource() && IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set basic fields
	d.Set("instance_name", instanceName)
	d.Set("table_name", tableName)
	d.Set("index_name", indexName)

	if index.SourceIndexName != nil {
		d.Set("source_index_name", *index.SourceIndexName)
	}

	if index.TimeToLive != nil {
		d.Set("time_to_live", *index.TimeToLive)
	}

	// Set field schemas
	if index.IndexSchema != nil && index.IndexSchema.FieldSchemas != nil {
		fieldSchemas := make([]map[string]interface{}, len(index.IndexSchema.FieldSchemas))
		for i, fieldSchema := range index.IndexSchema.FieldSchemas {
			fieldMap := make(map[string]interface{})
			if fieldSchema.FieldName != nil {
				fieldMap["field_name"] = *fieldSchema.FieldName
			}
			fieldMap["field_type"] = fieldSchema.FieldType.String()
			if fieldSchema.IsArray != nil {
				fieldMap["is_array"] = *fieldSchema.IsArray
			}
			if fieldSchema.Index != nil {
				fieldMap["index"] = *fieldSchema.Index
			}
			if fieldSchema.Analyzer != nil {
				fieldMap["analyzer"] = fieldSchema.Analyzer
			}
			if fieldSchema.EnableSortAndAgg != nil {
				fieldMap["enable_sort_and_agg"] = *fieldSchema.EnableSortAndAgg
			}
			if fieldSchema.Store != nil {
				fieldMap["store"] = *fieldSchema.Store
			}
			fieldSchemas[i] = fieldMap
		}
		d.Set("field_schemas", fieldSchemas)
	}

	// Set routing fields
	if index.IndexSchema != nil && index.IndexSchema.IndexSetting != nil {
		d.Set("routing_fields", index.IndexSchema.IndexSetting.RoutingFields)
	}

	// Set computed fields
	d.Set("sync_phase", index.SyncPhase.String())
	d.Set("create_time", index.CreateTime.Format(time.RFC3339))

	return nil
}

func resourceAliCloudOtsSearchIndexUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceName, tableName, indexName, err := DecodeOtsSearchIndexId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	// Only TTL can be updated
	if d.HasChange("time_to_live") {
		index := &tablestoreAPI.TablestoreSearchIndex{
			TableName: tableName,
			IndexName: indexName,
		}

		if ttl, ok := d.GetOk("time_to_live"); ok {
			ttlInt32 := int32(ttl.(int))
			index.TimeToLive = &ttlInt32
		}

		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			if err := otsService.UpdateOtsSearchIndex(instanceName, index); err != nil {
				if IsExpectedErrors(err, []string{"ThrottlingException", "ServiceUnavailable"}) {
					time.Sleep(5 * time.Second)
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateSearchIndex", AlibabaCloudSdkGoERROR)
		}
	}

	return resourceAliCloudOtsSearchIndexRead(d, meta)
}

func resourceAliCloudOtsSearchIndexDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceName, tableName, indexName, err := DecodeOtsSearchIndexId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	index := &tablestoreAPI.TablestoreSearchIndex{
		TableName: tableName,
		IndexName: indexName,
	}

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		if err := otsService.DeleteOtsSearchIndex(instanceName, index); err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			if IsExpectedErrors(err, []string{"ThrottlingException", "ServiceUnavailable"}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteSearchIndex", AlibabaCloudSdkGoERROR)
	}

	// Wait for deletion to complete
	if err := otsService.WaitForOtsSearchIndexDeleting(instanceName, tableName, indexName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}

// Helper functions

func buildIndexSchemaFromResource(d *schema.ResourceData) (*tablestore.IndexSchema, error) {
	schema := &tablestore.IndexSchema{}

	// Build field schemas
	if fieldSchemasRaw, ok := d.GetOk("field_schemas"); ok {
		fieldSchemasList := fieldSchemasRaw.([]interface{})
		fieldSchemas := make([]*tablestore.FieldSchema, len(fieldSchemasList))

		for i, fieldSchemaRaw := range fieldSchemasList {
			fieldSchemaMap := fieldSchemaRaw.(map[string]interface{})
			fieldSchema := &tablestore.FieldSchema{}

			if fieldName, ok := fieldSchemaMap["field_name"]; ok {
				fieldNameStr := fieldName.(string)
				fieldSchema.FieldName = &fieldNameStr
			}

			if fieldType, ok := fieldSchemaMap["field_type"]; ok {
				fieldTypeEnum, err := tablestore.ToFieldType(fieldType.(string))
				if err != nil {
					return nil, WrapError(err)
				}
				fieldSchema.FieldType = fieldTypeEnum
			}

			if isArray, ok := fieldSchemaMap["is_array"]; ok {
				isArrayBool := isArray.(bool)
				fieldSchema.IsArray = &isArrayBool
			}

			if index, ok := fieldSchemaMap["index"]; ok {
				indexBool := index.(bool)
				fieldSchema.Index = &indexBool
			}

			if analyzer, ok := fieldSchemaMap["analyzer"]; ok && analyzer.(string) != "" {
				analyzerEnum := tablestore.Analyzer(analyzer.(string))
				fieldSchema.Analyzer = &analyzerEnum
			}

			if enableSortAndAgg, ok := fieldSchemaMap["enable_sort_and_agg"]; ok {
				enableSortAndAggBool := enableSortAndAgg.(bool)
				fieldSchema.EnableSortAndAgg = &enableSortAndAggBool
			}

			if store, ok := fieldSchemaMap["store"]; ok {
				storeBool := store.(bool)
				fieldSchema.Store = &storeBool
			}

			fieldSchemas[i] = fieldSchema
		}

		schema.FieldSchemas = fieldSchemas
	}

	// Build index setting
	if routingFieldsRaw, ok := d.GetOk("routing_fields"); ok {
		routingFieldsList := routingFieldsRaw.([]interface{})
		routingFields := make([]string, len(routingFieldsList))
		for i, field := range routingFieldsList {
			routingFields[i] = field.(string)
		}

		schema.IndexSetting = &tablestore.IndexSetting{
			RoutingFields: routingFields,
		}
	}

	// Build index sort
	if indexSortRaw, ok := d.GetOk("index_sort"); ok && len(indexSortRaw.([]interface{})) > 0 {
		indexSortMap := indexSortRaw.([]interface{})[0].(map[string]interface{})

		if sortersRaw, ok := indexSortMap["sorters"]; ok {
			sortersList := sortersRaw.([]interface{})
			sorters := make([]search.Sorter, len(sortersList))

			for i, sorterRaw := range sortersList {
				sorterMap := sorterRaw.(map[string]interface{})

				sorterType := sorterMap["sorter_type"].(string)
				order := sorterMap["order"].(string)

				orderEnum, err := ConvertSearchIndexOrderTypeString(SearchIndexOrderTypeString(order))
				if err != nil {
					return nil, WrapError(err)
				}

				switch sorterType {
				case string(OtsSearchPrimaryKeySort):
					sorters[i] = &search.PrimaryKeySort{Order: &orderEnum}
				case string(OtsSearchFieldSort):
					fieldSort := &search.FieldSort{Order: &orderEnum}
					if fieldName, ok := sorterMap["field_name"]; ok {
						fieldSort.FieldName = fieldName.(string)
					}
					if mode, ok := sorterMap["mode"]; ok && mode.(string) != "" {
						modeEnum, err := ConvertSearchIndexSortModeString(SearchIndexSortModeString(mode.(string)))
						if err != nil {
							return nil, WrapError(err)
						}
						fieldSort.Mode = &modeEnum
					}
					sorters[i] = fieldSort
				}
			}

			schema.IndexSort = &search.Sort{Sorters: sorters}
		}
	}

	return schema, nil
}

// ID encoding/decoding functions
func EncodeOtsSearchIndexId(instanceName, tableName, indexName string) string {
	return fmt.Sprintf("%s:%s:%s", instanceName, tableName, indexName)
}

func DecodeOtsSearchIndexId(id string) (string, string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid search index ID format, expected instanceName:tableName:indexName, got %s", id)
	}
	return parts[0], parts[1], parts[2], nil
}
