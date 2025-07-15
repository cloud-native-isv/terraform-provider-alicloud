package alicloud

import (
	"encoding/json"
	"fmt"
	"regexp"

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
			},
			"table_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateOTSTableName,
			},
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				ForceNew: true,
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.ValidateRegexp,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"indexes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"table_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"index_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"create_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"time_to_live": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"sync_phase": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"current_sync_timestamp": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"schema": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"storage_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"row_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"reserved_read_cu": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"metering_last_update_time": {
							Type:     schema.TypeInt,
							Computed: true,
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
	totalIndexes, err := otsService.ListOtsSearchIndex(args.instanceName, args.tableName)
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
		id := fmt.Sprintf("%s:%s:%s:SearchIndex", args.instanceName, args.tableName, indexName)

		indexResp, err := otsService.DescribeOtsSearchIndex(args.instanceName, args.tableName, indexName)
		if err != nil {
			return nil, WrapError(err)
		}

		var phase string
		var currentSyncTimestamp int64
		var schema string

		if indexResp.IndexSchema != nil {
			b, err := json.MarshalIndent(indexResp.IndexSchema, "", "  ")
			if err != nil {
				return nil, WrapError(err)
			}
			schema = string(b)
		}

		index := map[string]interface{}{
			"id":            id,
			"instance_name": args.instanceName,
			"table_name":    args.tableName,
			"index_name":    indexName,

			"create_time":            indexResp.CreateTime,
			"time_to_live":           indexResp.TimeToLive,
			"sync_phase":             phase,
			"current_sync_timestamp": currentSyncTimestamp,
			"schema":                 schema,
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
			id := fmt.Sprintf("%s:%s:%s:SearchIndex", args.instanceName, args.tableName, indexInfo.IndexName)
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
