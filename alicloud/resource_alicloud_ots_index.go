package alicloud

import (
	"fmt"
	"log"
	"time"

	"github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	tablestoreAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/tablestore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudOtsIndex() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudOtsIndexCreate,
		Read:   resourceAliCloudOtsIndexRead,
		Delete: resourceAliCloudOtsIndexDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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

			"index_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateOTSIndexName,
				Description:  "The name of the secondary index.",
			},

			"index_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Global", "Local"}, false),
				Description: "The type of the secondary index. Valid values: Global, Local.",
			},

			"include_base_data": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
				Description: "Whether to include base data in the index.",
			},

			"primary_keys": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "The primary key columns of the index.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MinItems: 1,
				MaxItems: 4,
			},

			"defined_columns": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Description: "The predefined columns of the index.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MaxItems: 32,
			},

			"index_sync_phase": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The synchronization phase of the index. Valid values: FULL, INCR.",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceAliCloudOtsIndexCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceName := d.Get("instance_name").(string)
	tableName := d.Get("table_name").(string)
	indexName := d.Get("index_name").(string)
	indexType := d.Get("index_type").(string)
	indexSyncPhase := d.Get("index_sync_phase").(string)
	includeBaseData := d.Get("include_base_data").(bool)
	primaryKeys := convertToStringSlice(d.Get("primary_keys").([]interface{}))
	definedColumns := convertToStringSlice(d.Get("defined_columns").([]interface{}))

	// Build IndexMeta
	indexMeta := &tablestore.IndexMeta{
		IndexName:      indexName,
		Primarykey:     primaryKeys,
		DefinedColumns: definedColumns,
	}

	// Set index type
	switch indexType {
	case "Global":
		indexMeta.IndexType = tablestore.IT_GLOBAL_INDEX
	case "Local":
		indexMeta.IndexType = tablestore.IT_LOCAL_INDEX
	default:
		return WrapError(fmt.Errorf("invalid index type: %s", indexType))
	}

	// Set index Sync Phase
	switch indexSyncPhase {
	case "FULL":
		phase := tablestore.SyncPhase_FULL
		indexMeta.IndexSyncPhase = &phase
	case "INCR":
		phase := tablestore.SyncPhase_INCR
		indexMeta.IndexSyncPhase = &phase
	default:
		phase := tablestore.SyncPhase_FULL
		indexMeta.IndexSyncPhase = &phase
	}

	// Create TablestoreIndex object
	index := &tablestoreAPI.TablestoreIndex{
		TableName:       tableName,
		IndexName:       indexName,
		IndexMeta:       indexMeta,
		IncludeBaseData: includeBaseData,
	}

	// Create the index
	err = otsService.CreateOtsIndex(instanceName, index)
	if err != nil {
		return WrapError(err)
	}

	// Set resource ID
	d.SetId(EncodeOtsIndexId(instanceName, tableName, indexName))

	// Wait for index creation to complete
	err = otsService.WaitForOtsIndexCreating(instanceName, tableName, indexName, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudOtsIndexRead(d, meta)
}

func resourceAliCloudOtsIndexRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceName, tableName, indexName, err := DecodeOtsIndexId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	object, err := otsService.DescribeOtsIndex(instanceName, tableName, indexName)
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_ots_index DescribeOtsIndex Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set attributes
	d.Set("instance_name", instanceName)
	d.Set("table_name", tableName)
	d.Set("index_name", indexName)
	d.Set("include_base_data", object.IncludeBaseData)

	if object.IndexMeta != nil {
		// Set index type
		switch object.IndexMeta.IndexType {
		case tablestore.IT_GLOBAL_INDEX:
			d.Set("index_type", "Global")
		case tablestore.IT_LOCAL_INDEX:
			d.Set("index_type", "Local")
		}

		d.Set("primary_keys", object.IndexMeta.Primarykey)
		d.Set("defined_columns", object.IndexMeta.DefinedColumns)

		// Set index sync phase if available
		if object.IndexMeta.IndexSyncPhase != nil {
			switch *object.IndexMeta.IndexSyncPhase {
			case tablestore.SyncPhase_FULL:
				d.Set("index_sync_phase", "FULL")
			case tablestore.SyncPhase_INCR:
				d.Set("index_sync_phase", "INCR")
			default:
				d.Set("index_sync_phase", "UNKNOWN")
			}
		}
	}

	return nil
}

func resourceAliCloudOtsIndexDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	otsService, err := NewOtsService(client)
	if err != nil {
		return WrapError(err)
	}

	instanceName, tableName, indexName, err := DecodeOtsIndexId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	// Create TablestoreIndex object for deletion
	index := &tablestoreAPI.TablestoreIndex{
		TableName: tableName,
		IndexName: indexName,
	}

	err = otsService.DeleteOtsIndex(instanceName, index)
	if err != nil {
		if NotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteOtsIndex", AlibabaCloudSdkGoERROR)
	}

	// Wait for index deletion to complete
	err = otsService.WaitForOtsIndexDeleting(instanceName, tableName, indexName, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}

func convertToStringSlice(v []interface{}) []string {
	if v == nil {
		return []string{}
	}
	result := make([]string, len(v))
	for i, val := range v {
		result[i] = val.(string)
	}
	return result
}
