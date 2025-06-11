package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunRdsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/rds"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAlicloudDBInstanceClasses() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudDBInstanceClassesRead,

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"engine": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"MySQL", "SQLServer", "PostgreSQL", "PPAS", "MariaDB"}, false),
			},
			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"sorted_by": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"Price"}, false),
			},
			"instance_charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      PostPaid,
				ValidateFunc: validation.StringInSlice([]string{string(PostPaid), string(PrePaid), string(Serverless)}, false),
			},
			"db_instance_class": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"category": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"Basic", "HighAvailability", "AlwaysOn", "Finance", "serverless_basic", "serverless_standard", "serverless_ha", "cluster"}, false),
			},
			"storage_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"cloud_ssd", "local_ssd", "cloud_essd", "cloud_essd2", "cloud_essd3"}, false),
			},
			"db_instance_storage_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"cloud_ssd", "local_ssd", "cloud_essd", "cloud_essd2", "cloud_essd3"}, false),
			},
			"commodity_code": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"bards", "rds", "rords", "rds_rordspre_public_cn", "bards_intl", "rds_intl", "rords_intl", "rds_rordspre_public_intl", "rds_serverless_public_cn", "rds_serverless_public_intl"}, false),
			},
			"db_instance_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"multi_zone": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"output_file": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ids": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			// Computed values.
			"instance_classes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone_ids": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"sub_zone_ids": {
										Type:     schema.TypeList,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Computed: true,
									},
								},
							},
							Computed: true,
						},
						"instance_class": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"price": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"storage_range": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"min": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"max": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"step": {
										Type:     schema.TypeString,
										Computed: true,
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

func dataSourceAlicloudDBInstanceClassesRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	// Get common parameters from schema
	instanceChargeType := d.Get("instance_charge_type").(string)
	if instanceChargeType == string(PostPaid) {
		instanceChargeType = string(Postpaid)
	} else {
		instanceChargeType = string(Prepaid)
	}

	// Get conditional parameters
	zoneId, zoneIdOk := d.GetOk("zone_id")
	engine, engineOk := d.GetOk("engine")
	engineVersion, engineVersionOk := d.GetOk("engine_version")
	dbInstanceClass, dbInstanceClassOk := d.GetOk("db_instance_class")
	dbInstanceStorageType, dbInstanceStorageTypeOk := d.GetOk("db_instance_storage_type")
	if !dbInstanceStorageTypeOk || dbInstanceStorageType.(string) == "" {
		dbInstanceStorageType, dbInstanceStorageTypeOk = d.GetOk("storage_type")
	}
	category, categoryOk := d.GetOk("category")

	// Create RDS API credentials
	credentials := &aliyunRdsAPI.RDSCredentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	}

	// Create RDS API client
	rdsAPI, err := aliyunRdsAPI.NewRDSAPI(credentials)
	if err != nil {
		return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_db_instance_classes", "NewRDSAPI", AlibabaCloudSdkGoERROR)
	}

	s := make([]map[string]interface{}, 0)
	ids := make([]string, 0)

	// If all required parameters are provided, call ListDBClasses directly
	if zoneIdOk && zoneId.(string) != "" &&
		engineOk && engine.(string) != "" &&
		engineVersionOk && engineVersion.(string) != "" &&
		dbInstanceStorageTypeOk && dbInstanceStorageType.(string) != "" &&
		categoryOk && category.(string) != "" {

		// Set up optional parameters
		dbInstanceId := ""
		orderType := ""
		commodityCode := ""

		if v, ok := d.GetOk("commodity_code"); ok {
			commodityCode = v.(string)
			if v, ok := d.GetOk("db_instance_id"); ok {
				dbInstanceId = v.(string)
			}
		}

		// Call the API with individual parameters
		classes, err := rdsAPI.ListDBClasses(
			client.RegionId,
			zoneId.(string),
			engine.(string),
			engineVersion.(string),
			dbInstanceStorageType.(string),
			category.(string),
			instanceChargeType,
			dbInstanceId,
			orderType,
			commodityCode,
		)
		if err != nil {
			return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_db_instance_classes", "ListDBClasses", AlibabaCloudSdkGoERROR)
		}

		// Process the results
		zoneIds := make([]map[string]interface{}, 0)
		zoneIds = append(zoneIds, map[string]interface{}{
			"id":           zoneId,
			"sub_zone_ids": splitMultiZoneId(zoneId.(string)),
		})

		for _, class := range classes {
			if dbInstanceClassOk && dbInstanceClass != "" && dbInstanceClass != class.DBInstanceClass {
				continue
			}

			mapping := map[string]interface{}{
				"instance_class": class.DBInstanceClass,
				"zone_ids":       zoneIds,
				"storage_range": map[string]interface{}{
					"min":  fmt.Sprint(class.DBInstanceStorageRange.MinValue),
					"max":  fmt.Sprint(class.DBInstanceStorageRange.MaxValue),
					"step": fmt.Sprint(class.DBInstanceStorageRange.Step),
				},
			}

			s = append(s, mapping)
			ids = append(ids, class.DBInstanceClass)
		}
	} else {
		// We need to first get available zones and then query for classes for each zone

		// 1. First, list available zones
		multiZone := false
		if v, ok := d.GetOk("multi_zone"); ok {
			multiZone = v.(bool)
		}

		// Get engines to check
		engines := make([]string, 0)
		if v, ok := d.GetOk("engine"); ok && v.(string) != "" {
			engines = append(engines, v.(string))
		} else {
			engines = []string{"MySQL", "SQLServer", "PostgreSQL", "PPAS", "MariaDB"}
		}

		var targetCategory, targetStorageType string
		if v, ok := d.GetOk("category"); ok && v.(string) != "" {
			targetCategory = v.(string)
		}
		if v, ok := d.GetOk("db_instance_storage_type"); ok && v.(string) != "" {
			targetStorageType = v.(string)
		}

		// Get available zones using standard SDK since the wrapper doesn't have this functionality yet
		availableZones := make([]map[string]interface{}, 0)

		for _, engine := range engines {
			action := "DescribeAvailableZones"
			request := map[string]interface{}{
				"RegionId": client.RegionId,
				"SourceIp": client.SourceIp,
				"Engine":   engine,
			}

			if v, ok := d.GetOk("engine_version"); ok && v.(string) != "" {
				request["EngineVersion"] = v.(string)
			}
			if v, ok := d.GetOk("zone_id"); ok && v.(string) != "" {
				request["ZoneId"] = v.(string)
			}
			if instanceChargeType == string(PostPaid) {
				request["CommodityCode"] = "bards"
			} else {
				request["CommodityCode"] = "rds"
			}
			if v, ok := d.GetOk("commodity_code"); ok {
				request["CommodityCode"] = v.(string)
				if v, ok := d.GetOk("db_instance_id"); ok {
					request["DBInstanceName"] = v.(string)
				}
			}

			var response map[string]interface{}
			wait := incrementalWait(3*time.Second, 3*time.Second)
			err = resource.Retry(5*time.Minute, func() *resource.RetryError {
				response, err = client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
				if err != nil {
					if NeedRetry(err) {
						wait()
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				return nil
			})
			addDebug(action, response, request)
			if err != nil {
				return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_db_zones", action, AlibabaCloudSdkGoERROR)
			}

			resp, err := jsonpath.Get("$.AvailableZones", response)
			if err != nil {
				return WrapErrorf(err, FailedGetAttributeMsg, action, "$.AvailableZones", response)
			}

			for _, r := range resp.([]interface{}) {
				availableZoneItem := r.(map[string]interface{})

				zoneId := fmt.Sprint(availableZoneItem["ZoneId"])

				// Skip zones that don't match multi-zone setting
				if (multiZone && !strings.Contains(zoneId, MULTI_IZ_SYMBOL)) || (!multiZone && strings.Contains(zoneId, MULTI_IZ_SYMBOL)) {
					continue
				}

				if targetCategory == "" && targetStorageType == "" {
					availableZones = append(availableZones, availableZoneItem)
					continue
				}

				// Filter by category and storage type
				for _, r := range availableZoneItem["SupportedEngines"].([]interface{}) {
					supportedEngineItem := r.(map[string]interface{})
					for _, r := range supportedEngineItem["SupportedEngineVersions"].([]interface{}) {
						supportedEngineVersionItem := r.(map[string]interface{})
						for _, r := range supportedEngineVersionItem["SupportedCategorys"].([]interface{}) {
							supportedCategoryItem := r.(map[string]interface{})
							if targetCategory != "" && targetCategory != fmt.Sprint(supportedCategoryItem["Category"]) {
								continue
							}
							if targetStorageType == "" {
								availableZones = append(availableZones, availableZoneItem)
								goto NEXT
							}
							for _, r := range supportedCategoryItem["SupportedStorageTypes"].([]interface{}) {
								supportedStorageTypeItem := r.(map[string]interface{})
								if targetStorageType != fmt.Sprint(supportedStorageTypeItem["StorageType"]) {
									continue
								}
								availableZones = append(availableZones, availableZoneItem)
								goto NEXT
							}
						}
					}
				}
			NEXT:
				continue
			}
		}

		// 2. Query available classes for each zone
		for _, availableZone := range availableZones {
			zoneIds := make([]map[string]interface{}, 0)
			zoneIds = append(zoneIds, map[string]interface{}{
				"id":           fmt.Sprint(availableZone["ZoneId"]),
				"sub_zone_ids": splitMultiZoneId(fmt.Sprint(availableZone["ZoneId"])),
			})

			for _, r := range availableZone["SupportedEngines"].([]interface{}) {
				supportedEngineItem := r.(map[string]interface{})
				currentEngine := fmt.Sprint(supportedEngineItem["Engine"])

				for _, r := range supportedEngineItem["SupportedEngineVersions"].([]interface{}) {
					supportedEngineVersionItem := r.(map[string]interface{})
					currentEngineVersion := fmt.Sprint(supportedEngineVersionItem["Version"])

					for _, r := range supportedEngineVersionItem["SupportedCategorys"].([]interface{}) {
						supportedCategoryItem := r.(map[string]interface{})
						currentCategory := fmt.Sprint(supportedCategoryItem["Category"])

						for _, r := range supportedCategoryItem["SupportedStorageTypes"].([]interface{}) {
							storageTypeItem := r.(map[string]interface{})
							currentStorageType := fmt.Sprint(storageTypeItem["StorageType"])

							// Set up optional parameters
							dbInstanceId := ""
							orderType := ""
							commodityCode := ""

							if v, ok := d.GetOk("commodity_code"); ok {
								commodityCode = v.(string)
								if v, ok := d.GetOk("db_instance_id"); ok {
									dbInstanceId = v.(string)
								}
							}

							// Call the API with individual parameters
							classes, err := rdsAPI.ListDBClasses(
								client.RegionId,
								fmt.Sprint(availableZone["ZoneId"]),
								currentEngine,
								currentEngineVersion,
								currentStorageType,
								currentCategory,
								instanceChargeType,
								dbInstanceId,
								orderType,
								commodityCode,
							)
							if err != nil {
								return WrapErrorf(err, DataDefaultErrorMsg, "alicloud_db_instance_classes", "ListDBClasses", AlibabaCloudSdkGoERROR)
							}

							// Process the results
							for _, class := range classes {
								if dbInstanceClassOk && dbInstanceClass != "" && dbInstanceClass != class.DBInstanceClass {
									continue
								}

								mapping := map[string]interface{}{
									"instance_class": class.DBInstanceClass,
									"zone_ids":       zoneIds,
									"storage_range": map[string]interface{}{
										"min":  fmt.Sprint(class.DBInstanceStorageRange.MinValue),
										"max":  fmt.Sprint(class.DBInstanceStorageRange.MaxValue),
										"step": fmt.Sprint(class.DBInstanceStorageRange.Step),
									},
								}

								s = append(s, mapping)
								ids = append(ids, class.DBInstanceClass)
							}
						}
					}
				}
			}
		}
	}

	d.SetId(dataResourceIdHash(ids))
	err = d.Set("instance_classes", s)
	if err != nil {
		return WrapError(err)
	}
	d.Set("ids", ids)
	if output, ok := d.GetOk("output_file"); ok {
		err = writeToFile(output.(string), s)
		if err != nil {
			return WrapError(err)
		}
	}
	return nil
}
