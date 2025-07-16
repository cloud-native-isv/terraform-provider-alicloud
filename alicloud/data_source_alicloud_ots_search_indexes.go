package alicloud

import (
	"regexp"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore/search"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	tablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudOtsSearchIndexes() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudOtsSearchIndexesRead,

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
			"ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				ForceNew:    true,
				Description: "A list of search index IDs.",
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.ValidateRegexp,
				Description:  "A regex string to filter results by search index name.",
			},
			"output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "File name where to save data source results (after running `terraform plan`).",
			},

			"names": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "A list of search index names.",
			},
			"indexes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of search indexes.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the search index.",
						},
						"instance_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the OTS instance.",
						},
						"table_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the table.",
						},
						"index_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the search index.",
						},
						"source_index_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the source index.",
						},
						"time_to_live": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The time to live in seconds.",
						},
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
						"field_schemas": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The field schemas of the search index.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"field_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The name of the field.",
									},
									"field_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The type of the field.",
									},
									"is_array": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether the field is an array.",
									},
									"index": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether an index is created for the field.",
									},
									"analyzer": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The analyzer for the field.",
									},
									"enable_sort_and_agg": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether sorting and aggregation are enabled for the field.",
									},
									"store": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Whether the field value is stored.",
									},
								},
							},
						},
						"routing_fields": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "The routing fields for the search index.",
						},
						"index_sort": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The sorting configuration for the search index.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"sorters": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "The sorters for the index.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"sorter_type": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The type of the sorter.",
												},
												"order": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The sort order.",
												},
												"field_name": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The field name for field sort.",
												},
												"mode": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The sort mode for field sort.",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func parseSearchIndexDataSourceArgs(d *schema.ResourceData) *SearchIndexDataSourceArgs {
	args := &SearchIndexDataSourceArgs{
		instanceName: d.Get("instance_name").(string),
		tableName:    d.Get("table_name").(string),
	}

	if ids, ok := d.GetOk("ids"); ok && len(ids.([]interface{})) > 0 {
		args.ids = Interface2StrSlice(ids.([]interface{}))
	}
	if regx, ok := d.GetOk("name_regex"); ok && regx.(string) != "" {
		args.nameRegex = regx.(string)
	}
	return args
}

type SearchIndexDataSourceArgs struct {
	instanceName string
	tableName    string
	ids          []string
	nameRegex    string
}

type SearchIndexDataSource struct {
	ids     []string
	names   []string
	indexes []map[string]interface{}
}

func (s *SearchIndexDataSource) export(d *schema.ResourceData) error {
	d.SetId(dataResourceIdHash(s.ids))
	if err := d.Set("indexes", s.indexes); err != nil {
		return WrapError(err)
	}

	if err := d.Set("names", s.names); err != nil {
		return WrapError(err)
	}

	if err := d.Set("ids", s.ids); err != nil {
		return WrapError(err)
	}

	// create a json file in current directory and write data source to it.
	if filepath, ok := d.GetOk("output_file"); ok && filepath.(string) != "" {
		err := writeToFile(filepath.(string), s.indexes)
		if err != nil {
			return err
		}
	}
	return nil
}

func dataSourceAliCloudOtsSearchIndexesRead(d *schema.ResourceData, meta interface{}) error {
	args := parseSearchIndexDataSourceArgs(d)

	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	totalIndexes, err := otsService.ListOtsSearchIndexes(args.instanceName, args.tableName)
	if err != nil {
		return WrapError(err)
	}

	filteredIndexes := args.doFilters(totalIndexes)
	source, err := genSearchIndexDataSource(otsService, filteredIndexes, args)
	if err != nil {
		return WrapError(err)
	}
	if err := source.export(d); err != nil {
		return WrapError(err)
	}
	return nil
}

func genSearchIndexDataSource(otsService *OtsService, filteredIndexes []*tablestoreAPI.TablestoreSearchIndex, args *SearchIndexDataSourceArgs) (*SearchIndexDataSource, error) {
	size := len(filteredIndexes)
	ids := make([]string, 0, size)
	names := make([]string, 0, size)
	indexes := make([]map[string]interface{}, 0, size)

	for _, indexInfo := range filteredIndexes {
		indexName := indexInfo.IndexName
		id := EncodeOtsSearchIndexId(args.instanceName, args.tableName, indexName)

		// Get detailed information for each index
		indexResp, err := otsService.DescribeOtsSearchIndex(args.instanceName, args.tableName, indexName)
		if err != nil {
			return nil, WrapError(err)
		}

		index := map[string]interface{}{
			"id":            id,
			"instance_name": args.instanceName,
			"table_name":    args.tableName,
			"index_name":    indexName,
			"sync_phase":    indexResp.SyncPhase.String(),
			"create_time":   indexResp.CreateTime.Format("2006-01-02T15:04:05Z"),
		}

		// Set optional fields
		if indexResp.SourceIndexName != nil {
			index["source_index_name"] = *indexResp.SourceIndexName
		}

		if indexResp.TimeToLive != nil {
			index["time_to_live"] = int(*indexResp.TimeToLive)
		}

		// Set field schemas
		if indexResp.IndexSchema != nil && indexResp.IndexSchema.FieldSchemas != nil {
			fieldSchemas := make([]map[string]interface{}, len(indexResp.IndexSchema.FieldSchemas))
			for i, fieldSchema := range indexResp.IndexSchema.FieldSchemas {
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
			index["field_schemas"] = fieldSchemas
		}

		// Set routing fields
		if indexResp.IndexSchema != nil && indexResp.IndexSchema.IndexSetting != nil {
			index["routing_fields"] = indexResp.IndexSchema.IndexSetting.RoutingFields
		}

		// Set index sort if available
		if indexResp.IndexSchema != nil && indexResp.IndexSchema.IndexSort != nil {
			indexSortList := make([]map[string]interface{}, 1)
			sortersList := make([]map[string]interface{}, len(indexResp.IndexSchema.IndexSort.Sorters))

			for i, sorter := range indexResp.IndexSchema.IndexSort.Sorters {
				sorterMap := make(map[string]interface{})

				// Check sorter type and extract information
				switch s := sorter.(type) {
				case *search.PrimaryKeySort:
					sorterMap["sorter_type"] = "PrimaryKeySort"
					if s.Order != nil {
						sorterMap["order"] = s.Order.String()
					}
				case *search.FieldSort:
					sorterMap["sorter_type"] = "FieldSort"
					sorterMap["field_name"] = s.FieldName
					if s.Order != nil {
						sorterMap["order"] = s.Order.String()
					}
					if s.Mode != nil {
						sorterMap["mode"] = s.Mode.String()
					}
				}

				sortersList[i] = sorterMap
			}

			indexSortList[0] = map[string]interface{}{
				"sorters": sortersList,
			}
			index["index_sort"] = indexSortList
		}

		names = append(names, indexName)
		ids = append(ids, id)
		indexes = append(indexes, index)
	}

	return &SearchIndexDataSource{
		ids:     ids,
		names:   names,
		indexes: indexes,
	}, nil
}

func (args *SearchIndexDataSourceArgs) doFilters(total []*tablestoreAPI.TablestoreSearchIndex) []*tablestoreAPI.TablestoreSearchIndex {
	var result []*tablestoreAPI.TablestoreSearchIndex

	// Filter by IDs if specified
	idsMap := make(map[string]bool)
	if args.ids != nil {
		for _, id := range args.ids {
			idsMap[id] = true
		}
	}

	// Compile regex if specified
	var nameRegex *regexp.Regexp
	if args.nameRegex != "" {
		nameRegex = regexp.MustCompile(args.nameRegex)
	}

	for _, indexInfo := range total {
		// Apply ID filter
		if len(idsMap) > 0 {
			id := EncodeOtsSearchIndexId(args.instanceName, args.tableName, indexInfo.IndexName)
			if !idsMap[id] {
				continue
			}
		}

		// Apply name regex filter
		if nameRegex != nil && !nameRegex.MatchString(indexInfo.IndexName) {
			continue
		}

		result = append(result, indexInfo)
	}

	return result
}
