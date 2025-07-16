package alicloud

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"strconv"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	tablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudOtsTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliyunOtsTableCreate,
		Read:   resourceAliyunOtsTableRead,
		Update: resourceAliyunOtsTableUpdate,
		Delete: resourceAliyunOtsTableDelete,
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
			"primary_key": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				MaxItems:    4,
				Description: "The primary key schema of the table.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The name of the primary key column.",
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(IntegerType), string(BinaryType), string(StringType)}, false),
							Description: "The type of the primary key column.",
						},
					},
				},
			},
			"defined_column": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    32,
				Description: "The defined column schema of the table.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the defined column.",
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(DefinedColumnInteger), string(DefinedColumnString),
								string(DefinedColumnBinary), string(DefinedColumnDouble),
								string(DefinedColumnBoolean)}, false),
							Description: "The type of the defined column.",
						},
					},
				},
			},
			"time_to_live": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(-1, INT_MAX),
				Description:  "The time to live of the table data in seconds.",
			},
			"max_version": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The maximum number of versions for each column.",
			},
			"allow_update": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to allow update operations on the table.",
			},
			"deviation_cell_version_in_sec": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateStringConvertInt64(),
				Default:      "86400",
				Description:  "The maximum deviation of the cell version in seconds.",
			},
			"enable_sse": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to enable server-side encryption.",
			},
			"sse_key_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(SseKMSService), string(SseByOk)}, false),
				Description: "The type of the server-side encryption key.",
			},
			"sse_key_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the server-side encryption key.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return string(SseByOk) != d.Get("sse_key_type").(string)
				},
			},
			"sse_role_arn": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ARN of the server-side encryption role.",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return string(SseByOk) != d.Get("sse_key_type").(string)
				},
			},
			"enable_local_txn": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to enable local transaction.",
			},
			// Computed fields
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
		},
	}
}

func resourceAliyunOtsTableCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceName := d.Get("instance_name").(string)
	tableName := d.Get("table_name").(string)

	// Check if instance exists
	if err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, e := otsService.DescribeOtsInstance(instanceName)
		if e != nil {
			if NotFoundError(e) {
				return resource.RetryableError(e)
			}
			return resource.NonRetryableError(e)
		}
		return nil
	}); err != nil {
		return WrapError(err)
	}

	// Build table object
	table := &tablestoreAPI.TablestoreTable{
		InstanceName: instanceName,
	}
	table.SetName(tableName)

	// Set table options
	table.SetTimeToAlive(d.Get("time_to_live").(int))
	table.SetMaxVersion(d.Get("max_version").(int))

	// Create table using service
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		err := otsService.CreateOtsTable(instanceName, table)
		if err != nil {
			if IsExpectedErrors(err, []string{"ThrottlingException", "ServiceUnavailable"}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_ots_table", "CreateOtsTable", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID
	d.SetId(EncodeOtsTableId(instanceName, tableName))

	// Wait for table to be ready
	err = otsService.WaitForOtsTableCreating(d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliyunOtsTableRead(d, meta)
}

func resourceAliyunOtsTableRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	table, err := otsService.DescribeOtsTable(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	if table == nil {
		if !d.IsNewResource() {
			d.SetId("")
			return nil
		}
		return WrapError(Error("table not found"))
	}

	// Set basic information
	d.Set("instance_name", table.GetInstanceName())
	d.Set("table_name", table.GetName())
	d.Set("time_to_live", table.GetTimeToAlive())
	d.Set("max_version", table.GetMaxVersion())
	d.Set("status", table.Status)
	if !table.CreateTime.IsZero() {
		d.Set("create_time", table.CreateTime.Format(time.RFC3339))
	}

	// Set primary keys
	var primaryKeys []map[string]interface{}
	for _, pk := range table.GetPrimaryKeys() {
		if pk != nil && pk.Name != nil {
			primaryKeys = append(primaryKeys, map[string]interface{}{
				"name": *pk.Name,
				"type": *pk.Type,
			})
		}
	}
	d.Set("primary_key", primaryKeys)

	// Set defined columns
	var definedColumns []map[string]interface{}
	for _, col := range table.GetDefinedColumns() {
		if col != nil {
			columnType, err := ConvertDefinedColumnType(col.ColumnType)
			if err != nil {
				return WrapError(err)
			}
			definedColumns = append(definedColumns, map[string]interface{}{
				"name": col.Name,
				"type": columnType,
			})
		}
	}
	d.Set("defined_column", definedColumns)

	// Set table options if available
	if table.TableOption != nil {
		if table.TableOption.AllowUpdate != nil {
			d.Set("allow_update", *table.TableOption.AllowUpdate)
		}
		d.Set("deviation_cell_version_in_sec", strconv.FormatInt(table.TableOption.DeviationCellVersionInSec, 10))
	}

	// Set SSE information if available
	if table.SSEDetails != nil && table.SSEDetails.Enable {
		d.Set("enable_sse", table.SSEDetails.Enable)
		d.Set("sse_key_type", table.SSEDetails.KeyType.String())
		d.Set("sse_key_id", table.SSEDetails.KeyId)
		d.Set("sse_role_arn", table.SSEDetails.RoleArn)
	}

	// Set local transaction if available
	if table.EnableLocalTxn != nil {
		d.Set("enable_local_txn", *table.EnableLocalTxn)
	}

	return nil
}

func resourceAliyunOtsTableUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceName, tableName, err := DecodeOtsTableId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	// Check if table options need to be updated
	if d.HasChange("time_to_live") || d.HasChange("max_version") ||
		d.HasChange("deviation_cell_version_in_sec") || d.HasChange("allow_update") {

		// Build updated table object
		table := &tablestoreAPI.TablestoreTable{
			InstanceName: instanceName,
		}
		table.SetName(tableName)
		table.SetTimeToAlive(d.Get("time_to_live").(int))
		table.SetMaxVersion(d.Get("max_version").(int))

		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			err := otsService.UpdateOtsTable(instanceName, table)
			if err != nil {
				if IsExpectedErrors(err, []string{"ThrottlingException", "ServiceUnavailable"}) {
					time.Sleep(5 * time.Second)
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateOtsTable", AlibabaCloudSdkGoERROR)
		}
	}

	return resourceAliyunOtsTableRead(d, meta)
}

func resourceAliyunOtsTableDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := otsService.DeleteOtsTable(d.Id())
		if err != nil {
			if NotFoundError(err) {
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
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteOtsTable", AlibabaCloudSdkGoERROR)
	}

	// Wait for table deletion
	err = otsService.WaitForOtsTableDeleting(d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
