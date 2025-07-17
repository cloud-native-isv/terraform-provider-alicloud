package alicloud

import (
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
				Description:  "The name of the OTS instance.",
			},
			"ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				ForceNew:    true,
				MinItems:    1,
				Description: "A list of table IDs.",
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.ValidateRegexp,
				Description:  "A regex string to filter results by table name.",
			},
			"output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "File name where to save data source results (after running `terraform plan`).",
			},

			// Computed values
			"names": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "A list of table names.",
			},
			"tables": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of tables.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The resource ID in terraform of table.",
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
						"primary_key": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The primary key schema of the table.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The name of the primary key column.",
									},
									"type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The type of the primary key column.",
									},
								},
							},
							MaxItems: 4,
						},
						"defined_column": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The defined column schema of the table.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The name of the defined column.",
									},
									"type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The type of the defined column.",
									},
								},
							},
							MaxItems: 32,
						},
						"time_to_live": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The time to live of the table data in seconds.",
						},
						"max_version": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The maximum number of versions for each column.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of the table.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The creation time of the table.",
						},
						"enable_sse": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether server-side encryption is enabled.",
						},
						"sse_key_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of the server-side encryption key.",
						},
						"enable_local_txn": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether local transaction is enabled.",
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

	// Get list of tables from service
	tableInfos, err := otsService.ListOtsTables(instanceName)
	if err != nil {
		return WrapError(err)
	}

	// Extract table names for filtering
	var tables []string
	for _, tableInfo := range tableInfos {
		tables = append(tables, tableInfo.GetName())
	}

	// Apply ID filtering
	idsMap := make(map[string]bool)
	if v, ok := d.GetOk("ids"); ok && len(v.([]interface{})) > 0 {
		for _, x := range v.([]interface{}) {
			if x == nil {
				continue
			}
			idsMap[x.(string)] = true
		}
	}

	// Apply name regex filtering
	var nameReg *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok && v.(string) != "" {
		nameReg = regexp.MustCompile(v.(string))
	}

	// Filter table names
	var filteredTableNames []string
	for _, tableName := range tables {
		// name_regex mismatch
		if nameReg != nil && !nameReg.MatchString(tableName) {
			continue
		}
		// ids mismatch
		if len(idsMap) != 0 {
			id := EncodeOtsTableId(instanceName, tableName)
			if _, ok := idsMap[id]; !ok {
				continue
			}
		}
		filteredTableNames = append(filteredTableNames, tableName)
	}

	// Get full table info via DescribeTable for filtered tables
	var allTableInfos []*tablestoreAPI.TablestoreTable
	for _, tableName := range filteredTableNames {
		object, err := otsService.DescribeOtsTable(instanceName, tableName)
		if err != nil {
			if NotFoundError(err) {
				// Skip tables that are not found (may have been deleted)
				continue
			}
			return WrapError(err)
		}
		if object != nil {
			allTableInfos = append(allTableInfos, object)
		}
	}

	return otsTablesDescriptionAttributes(d, allTableInfos, meta)
}

func otsTablesDescriptionAttributes(d *schema.ResourceData, tableInfos []*tablestoreAPI.TablestoreTable, meta interface{}) error {
	var ids []string
	var names []string
	var s []map[string]interface{}

	for _, table := range tableInfos {
		id := EncodeOtsTableId(table.GetInstanceName(), table.GetName())
		mapping := map[string]interface{}{
			"id":            id,
			"instance_name": table.GetInstanceName(),
			"table_name":    table.GetName(),
			"time_to_live":  table.GetTimeToAlive(),
			"max_version":   table.GetMaxVersion(),
			"status":        table.Status,
		}

		// Set creation time if available
		if !table.CreateTime.IsZero() {
			mapping["create_time"] = table.CreateTime.Format("2006-01-02T15:04:05Z")
		}

		// Set primary keys
		var primaryKey []map[string]interface{}
		for _, pk := range table.GetPrimaryKeys() {
			if pk != nil && pk.Name != nil {
				pkColumn := map[string]interface{}{
					"name": *pk.Name,
					"type": *pk.Type,
				}
				primaryKey = append(primaryKey, pkColumn)
			}
		}
		mapping["primary_key"] = primaryKey

		// Set defined columns
		var definedColumn []map[string]interface{}
		for _, col := range table.GetDefinedColumns() {
			if col != nil {
				columnType, err := ConvertDefinedColumnType(col.ColumnType)
				if err != nil {
					return WrapError(err)
				}
				viewCol := map[string]interface{}{
					"name": col.Name,
					"type": columnType,
				}
				definedColumn = append(definedColumn, viewCol)
			}
		}
		mapping["defined_column"] = definedColumn

		// Set SSE information if available
		if table.SSEDetails != nil {
			mapping["enable_sse"] = table.SSEDetails.Enable
			if table.SSEDetails.Enable {
				mapping["sse_key_type"] = table.SSEDetails.KeyType.String()
			}
		} else {
			mapping["enable_sse"] = false
		}

		// Set local transaction if available
		if table.EnableLocalTxn != nil {
			mapping["enable_local_txn"] = *table.EnableLocalTxn
		} else {
			mapping["enable_local_txn"] = false
		}

		names = append(names, table.GetName())
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
