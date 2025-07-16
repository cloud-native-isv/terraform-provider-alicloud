package alicloud

import (
	"regexp"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	tablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudOtsIndexes() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudOtsIndexesRead,

		Schema: map[string]*schema.Schema{
			"instance_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateOTSInstanceName,
				Description:  "The name of the Tablestore instance.",
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
				Description: "A list of index IDs.",
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.ValidateRegexp,
				Description:  "A regex string to filter results by index name.",
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
				Description: "A list of index names.",
			},
			"indexes": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of indexes.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the index. Format: instanceName:tableName:indexName.",
						},
						"instance_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the Tablestore instance.",
						},
						"table_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the table.",
						},
						"index_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the secondary index.",
						},
						"index_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of the secondary index.",
						},
						"include_base_data": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to include base data in the index.",
						},
						"primary_keys": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The primary key columns of the index.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"defined_columns": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The predefined columns of the index.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func parseOtsIndexDataSourceArgs(d *schema.ResourceData) *OtsIndexDataSourceArgs {
	args := &OtsIndexDataSourceArgs{
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

type OtsIndexDataSourceArgs struct {
	instanceName string
	tableName    string
	ids          []string
	nameRegex    string
}

type OtsIndexDataSource struct {
	ids     []string
	names   []string
	indexes []map[string]interface{}
}

func (s *OtsIndexDataSource) export(d *schema.ResourceData) error {
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

func dataSourceAliCloudOtsIndexesRead(d *schema.ResourceData, meta interface{}) error {
	args := parseOtsIndexDataSourceArgs(d)

	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Use the new method that returns TablestoreIndex objects
	indexes, err := otsService.ListOtsIndexes(args.instanceName, args.tableName)
	if err != nil {
		return WrapError(err)
	}

	filteredIndexes := args.doFilters(indexes)
	source, err := genOtsIndexDataSource(filteredIndexes, args)
	if err != nil {
		return WrapError(err)
	}
	if err := source.export(d); err != nil {
		return WrapError(err)
	}
	return nil
}

func genOtsIndexDataSource(filteredIndexes []*tablestoreAPI.TablestoreIndex, args *OtsIndexDataSourceArgs) (*OtsIndexDataSource, error) {
	size := len(filteredIndexes)
	ids := make([]string, 0, size)
	names := make([]string, 0, size)
	indexes := make([]map[string]interface{}, 0, size)

	for _, idx := range filteredIndexes {
		id := EncodeOtsIndexId(args.instanceName, args.tableName, idx.IndexName)

		// Convert index type to string
		var indexType string
		if idx.IndexMeta != nil {
			switch idx.IndexMeta.IndexType {
			case tablestore.IT_GLOBAL_INDEX:
				indexType = "Global"
			case tablestore.IT_LOCAL_INDEX:
				indexType = "Local"
			default:
				indexType = "Unknown"
			}
		}

		index := map[string]interface{}{
			"id":                id,
			"instance_name":     args.instanceName,
			"table_name":        args.tableName,
			"index_name":        idx.IndexName,
			"index_type":        indexType,
			"include_base_data": idx.IncludeBaseData,
		}

		// Set primary keys and defined columns if IndexMeta exists
		if idx.IndexMeta != nil {
			index["primary_keys"] = idx.IndexMeta.Primarykey
			index["defined_columns"] = idx.IndexMeta.DefinedColumns
		} else {
			index["primary_keys"] = []string{}
			index["defined_columns"] = []string{}
		}

		names = append(names, idx.IndexName)
		ids = append(ids, id)
		indexes = append(indexes, index)
	}

	return &OtsIndexDataSource{
		ids:     ids,
		names:   names,
		indexes: indexes,
	}, nil
}

func (args *OtsIndexDataSourceArgs) doFilters(indexes []*tablestoreAPI.TablestoreIndex) []*tablestoreAPI.TablestoreIndex {
	var filteredIndexes []*tablestoreAPI.TablestoreIndex

	for _, idx := range indexes {
		// Filter by IDs if specified
		if args.ids != nil && len(args.ids) > 0 {
			id := EncodeOtsIndexId(args.instanceName, args.tableName, idx.IndexName)
			found := false
			for _, allowedId := range args.ids {
				if id == allowedId {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filter by name regex if specified
		if args.nameRegex != "" {
			matched, err := regexp.MatchString(args.nameRegex, idx.IndexName)
			if err != nil || !matched {
				continue
			}
		}

		filteredIndexes = append(filteredIndexes, idx)
	}

	return filteredIndexes
}
