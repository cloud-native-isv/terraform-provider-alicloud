package alicloud

import (
	"fmt"
	"regexp"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	tablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAliCloudOtsTables() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAliCloudOtsTablesRead,

		Schema: map[string]*schema.Schema{
			"instance_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateOTSInstanceName,
			},
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				ForceNew: true,
				MinItems: 1,
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

			// Computed values
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tables": {
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
						"primary_key": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
							MaxItems: 4,
						},
						"defined_column": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
							MaxItems: 32,
						},
						"time_to_live": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"max_version": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAliCloudOtsTablesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}
	instanceName := d.Get("instance_name").(string)

	tableInfos, err := otsService.ListOtsTable(instanceName)
	if err != nil {
		return WrapError(err)
	}

	// Extract table names from strongly typed OtsTableInfo slice
	var tables []string
	for _, tableInfo := range tableInfos {
		tables = append(tables, tableInfo.GetName())
	}

	idsMap := make(map[string]bool)
	if v, ok := d.GetOk("ids"); ok && len(v.([]interface{})) > 0 {
		for _, x := range v.([]interface{}) {
			if x == nil {
				continue
			}
			idsMap[x.(string)] = true
		}
	}

	var nameReg *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok && v.(string) != "" {
		nameReg = regexp.MustCompile(v.(string))
	}

	var filteredTableNames []string
	for _, tableName := range tables {
		//name_regex mismatch
		if nameReg != nil && !nameReg.MatchString(tableName) {
			continue
		}
		// ids mismatch
		if len(idsMap) != 0 {
			id := fmt.Sprintf("%s%s%s", instanceName, COLON_SEPARATED, tableName)
			if _, ok := idsMap[id]; !ok {
				continue
			}
		}
		filteredTableNames = append(filteredTableNames, tableName)
	}

	// get full table info via DescribeTable
	var allTableInfos []*tablestoreAPI.TablestoreTable
	for _, tableName := range filteredTableNames {
		object, err := otsService.DescribeOtsTable(instanceName, tableName)
		if err != nil {
			return WrapError(err)
		}
		allTableInfos = append(allTableInfos, object)
	}

	return otsTablesDescriptionAttributes(d, allTableInfos, meta)
}

func otsTablesDescriptionAttributes(d *schema.ResourceData, tableInfos []*tablestoreAPI.TablestoreTable, meta interface{}) error {
	var ids []string
	var names []string
	var s []map[string]interface{}
	for _, table := range tableInfos {
		id := fmt.Sprintf("%s:%s", table.GetInstanceName(), table.GetName())
		mapping := map[string]interface{}{
			"id":            id,
			"instance_name": table.GetInstanceName(),
			"table_name":    table.GetName(),
			"time_to_live":  table.GetTimeToAlive(),
			"max_version":   table.GetMaxVersion(),
		}
		var primaryKey []map[string]interface{}
		for _, pk := range table.GetPrimaryKeys() {
			pkColumn := make(map[string]interface{})
			pkColumn["name"] = pk.Name
			pkColumn["type"] = pk.Type
			primaryKey = append(primaryKey, pkColumn)
		}
		mapping["primary_key"] = primaryKey

		var definedColumn []map[string]interface{}
		for _, col := range table.GetDefinedColumns() {
			viewCol := map[string]interface{}{
				"name": col.Name,
				"type": col.ColumnType,
			}
			definedColumn = append(definedColumn, viewCol)
		}
		mapping["defined_column"] = definedColumn

		names = append(names, table.GetTableName())
		ids = append(ids, id)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("tables", s); err != nil {
		return WrapError(err)
	}

	if err := d.Set("names", names); err != nil {
		return WrapError(err)
	}

	if err := d.Set("ids", ids); err != nil {
		return WrapError(err)
	}

	// create a json file in current directory and write data source to it.
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}
	return nil
}
