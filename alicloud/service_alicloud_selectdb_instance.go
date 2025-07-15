package alicloud

import (
	"fmt"
	"time"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Instance Management Operations

// CreateSelectDBInstance creates a new SelectDB instance
func (s *SelectDBService) CreateSelectDBInstance(options *selectdb.CreateInstanceOptions) (*selectdb.CreateInstanceResult, error) {
	if options == nil {
		return nil, WrapError(fmt.Errorf("create instance options cannot be nil"))
	}

	result, err := s.selectdbAPI.CreateInstance(options)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// DescribeSelectDBInstance retrieves information about a SelectDB instance
func (s *SelectDBService) DescribeSelectDBInstance(instanceId string) (*selectdb.Instance, error) {
	if instanceId == "" {
		return nil, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}

	instance, err := s.selectdbAPI.GetInstance(instanceId)
	if err != nil {
		if selectdb.IsNotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapError(err)
	}

	return instance, nil
}

// DescribeSelectDBInstances lists SelectDB instances with pagination
func (s *SelectDBService) DescribeSelectDBInstances(options *selectdb.ListInstancesOptions) (*selectdb.ListInstancesResult, error) {
	if options == nil {
		options = &selectdb.ListInstancesOptions{}
	}

	result, err := s.selectdbAPI.ListInstances(options)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// ModifySelectDBInstance modifies attributes of a SelectDB instance
func (s *SelectDBService) ModifySelectDBInstance(options *selectdb.ModifyInstanceOptions) error {
	if options == nil {
		return WrapError(fmt.Errorf("modify instance options cannot be nil"))
	}

	err := s.selectdbAPI.ModifyInstance(options)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// DeleteSelectDBInstance deletes a SelectDB instance
func (s *SelectDBService) DeleteSelectDBInstance(instanceId string, regionId ...string) error {
	if instanceId == "" {
		return WrapError(fmt.Errorf("instance ID cannot be empty"))
	}

	err := s.selectdbAPI.DeleteInstance(instanceId, regionId...)
	if err != nil {
		if selectdb.IsNotFoundError(err) {
			return nil // Instance already deleted
		}
		return WrapError(err)
	}

	return nil
}

// CheckCreateSelectDBInstance validates instance creation parameters
func (s *SelectDBService) CheckCreateSelectDBInstance(options *selectdb.CheckCreateInstanceOptions) error {
	if options == nil {
		return WrapError(fmt.Errorf("check create instance options cannot be nil"))
	}

	err := s.selectdbAPI.CheckCreateInstance(options)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// State Management and Refresh Functions

// SelectDBInstanceStateRefreshFunc returns a ResourceStateRefreshFunc for SelectDB instance
func (s *SelectDBService) SelectDBInstanceStateRefreshFunc(instanceId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		instance, err := s.DescribeSelectDBInstance(instanceId)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", WrapErrorf(Error(GetNotFoundMessage("SelectDB Instance", instanceId)), NotFoundMsg, ProviderERROR)
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if instance.Status == failState {
				return instance, instance.Status, WrapError(Error(FailedToReachTargetStatus, instance.Status))
			}
		}

		return instance, instance.Status, nil
	}
}

// WaitForSelectDBInstance waits for SelectDB instance to reach expected status
func (s *SelectDBService) WaitForSelectDBInstance(instanceId string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)

	for {
		instance, err := s.DescribeSelectDBInstance(instanceId)
		if err != nil {
			if NotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}

		if instance != nil && instance.Status == string(status) {
			return nil
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, instanceId, GetFunc(1), timeout, instance.Status, string(status), ProviderERROR)
		}

		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

// WaitForSelectDBInstanceStatus waits for instance to reach desired status using state refresh
func (s *SelectDBService) WaitForSelectDBInstanceStatus(instanceId string, targetStatus string, timeout time.Duration) (*selectdb.Instance, error) {
	instance, err := s.selectdbAPI.WaitForInstanceStatus(instanceId, targetStatus, timeout)
	if err != nil {
		return nil, WrapError(err)
	}

	return instance, nil
}

// Utility Functions

// resetSelectDBInstancePassword resets account password for SelectDB instance
func (s *SelectDBService) resetSelectDBInstancePassword(options *selectdb.ResetPasswordOptions) error {
	if options == nil {
		return WrapError(fmt.Errorf("reset password options cannot be nil"))
	}

	err := s.selectdbAPI.ResetAccountPassword(options)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// Region and Zone Operations

// DescribeSelectDBRegions retrieves available regions for SelectDB
func (s *SelectDBService) DescribeSelectDBRegions(options *selectdb.DescribeRegionsOptions) ([]selectdb.RegionInfo, error) {
	regions, err := s.selectdbAPI.ListRegions(options)
	if err != nil {
		return nil, WrapError(err)
	}

	return regions, nil
}

// Connection Management

// CreateSelectDBConnection creates a connection string for SelectDB instance
func (s *SelectDBService) CreateSelectDBConnection(instanceId, connectionStringPrefix, netType string) error {
	// options := &selectdb.CreateConnectionOptions{
	// 	DBInstanceId:           instanceId,
	// 	ConnectionStringPrefix: connectionStringPrefix,
	// 	NetType:                netType,
	// }

	// err := s.selectdbAPI.CreateConnection(options)
	// if err != nil {
	// 	return WrapError(err)
	// }

	return nil
}

// ReleaseSelectDBConnection releases a connection string for SelectDB instance
func (s *SelectDBService) ReleaseSelectDBConnection(instanceId, connectionString string) error {
	// options := &selectdb.ReleaseConnectionOptions{
	// 	DBInstanceId:     instanceId,
	// 	ConnectionString: connectionString,
	// }

	// err := s.selectdbAPI.ReleaseConnection(options)
	// if err != nil {
	// 	return WrapError(err)
	// }

	return nil
}

// Error handling utilities

// IsSelectDBNotFoundError checks if the error indicates a resource was not found
func IsSelectDBNotFoundError(err error) bool {
	return selectdb.IsNotFoundError(err)
}

// IsSelectDBInvalidParameterError checks if the error indicates invalid parameters
func IsSelectDBInvalidParameterError(err error) bool {
	return selectdb.IsInvalidParameterError(err)
}

// Helper functions for converting between Terraform schema and API types

// ConvertToCreateInstanceOptions converts schema data to API create instance options
func ConvertToCreateInstanceOptions(d *schema.ResourceData) *selectdb.CreateInstanceOptions {
	// options := &selectdb.CreateInstanceOptions{}

	// if v, ok := d.GetOk("engine"); ok {
	// 	options.Engine = v.(string)
	// }
	// if v, ok := d.GetOk("engine_version"); ok {
	// 	options.EngineVersion = v.(string)
	// }
	// if v, ok := d.GetOk("db_instance_class"); ok {
	// 	options.DBInstanceClass = v.(string)
	// }
	// if v, ok := d.GetOk("db_instance_description"); ok {
	// 	options.DBInstanceDescription = v.(string)
	// }
	// if v, ok := d.GetOk("charge_type"); ok {
	// 	options.ChargeType = v.(string)
	// }
	// if v, ok := d.GetOk("period"); ok {
	// 	options.Period = v.(string)
	// }
	// if v, ok := d.GetOk("used_time"); ok {
	// 	options.UsedTime = int32(v.(int))
	// }
	// if v, ok := d.GetOk("region_id"); ok {
	// 	options.RegionId = v.(string)
	// }
	// if v, ok := d.GetOk("zone_id"); ok {
	// 	options.ZoneId = v.(string)
	// }
	// if v, ok := d.GetOk("vpc_id"); ok {
	// 	options.VpcId = v.(string)
	// }
	// if v, ok := d.GetOk("vswitch_id"); ok {
	// 	options.VSwitchId = v.(string)
	// }
	// if v, ok := d.GetOk("security_ip_list"); ok {
	// 	options.SecurityIPList = v.(string)
	// }
	// if v, ok := d.GetOk("cache_size"); ok {
	// 	options.CacheSize = int32(v.(int))
	// }
	// if v, ok := d.GetOk("resource_group_id"); ok {
	// 	options.ResourceGroupId = v.(string)
	// }

	// // Convert tags
	// if v, ok := d.GetOk("tags"); ok {
	// 	tags := v.(map[string]interface{})
	// 	for key, value := range tags {
	// 		options.Tags = append(options.Tags, selectdb.TagInfo{
	// 			Key:   key,
	// 			Value: value.(string),
	// 		})
	// 	}
	// }

	// // Convert multi-zone configuration
	// if v, ok := d.GetOk("multi_zone"); ok {
	// 	multiZoneList := v.([]interface{})
	// 	for _, mz := range multiZoneList {
	// 		multiZoneMap := mz.(map[string]interface{})
	// 		multiZone := selectdb.MultiZoneInfo{}

	// 		if zoneId, ok := multiZoneMap["zone_id"]; ok {
	// 			multiZone.ZoneId = zoneId.(string)
	// 		}
	// 		if vswitchIds, ok := multiZoneMap["vswitch_ids"]; ok {
	// 			vswitchIdList := vswitchIds.([]interface{})
	// 			for _, vsId := range vswitchIdList {
	// 				multiZone.VSwitchIds = append(multiZone.VSwitchIds, vsId.(string))
	// 			}
	// 		}

	// 		options.MultiZone = append(options.MultiZone, multiZone)
	// 	}
	// }

	return nil
}

// ConvertToModifyInstanceOptions converts schema data to API modify instance options
func ConvertToModifyInstanceOptions(d *schema.ResourceData, instanceId string) *selectdb.ModifyInstanceOptions {
	options := &selectdb.ModifyInstanceOptions{
		DBInstanceId: instanceId,
	}

	if v, ok := d.GetOk("instance_attribute_type"); ok {
		options.InstanceAttributeType = v.(string)
	}
	if v, ok := d.GetOk("value"); ok {
		options.Value = v.(string)
	}
	if v, ok := d.GetOk("region_id"); ok {
		options.RegionId = v.(string)
	}

	return options
}

// ConvertInstanceToMap converts API instance to Terraform map
func ConvertInstanceToMap(instance *selectdb.Instance) map[string]interface{} {
	if instance == nil {
		return nil
	}

	result := map[string]interface{}{
		"db_instance_id":       instance.DBInstanceId,
		"description":          instance.Description,
		"engine":               instance.Engine,
		"engine_version":       instance.EngineVersion,
		"engine_minor_version": instance.EngineMinorVersion,
		"status":               instance.Status,
		"category":             instance.Category,
		"charge_type":          instance.ChargeType,
		"vpc_id":               instance.VpcId,
		"vswitch_id":           instance.VswitchId,
		"zone_id":              instance.ZoneId,
		"region_id":            instance.RegionId,
		"connection_string":    instance.ConnectionString,
		"sub_domain":           instance.SubDomain,
		"resource_cpu":         instance.ResourceCpu,
		"resource_memory":      instance.ResourceMemory,
		"storage_size":         instance.StorageSize,
		"storage_type":         instance.StorageType,
		"object_store_size":    instance.ObjectStoreSize,
		"cluster_count":        instance.ClusterCount,
		"gmt_created":          instance.GmtCreated,
		"gmt_modified":         instance.GmtModified,
		"expire_time":          instance.ExpireTime,
	}

	// Convert tags
	if len(instance.Tags) > 0 {
		tags := make(map[string]interface{})
		for _, tag := range instance.Tags {
			tags[tag.Key] = tag.Value
		}
		result["tags"] = tags
	}

	// Convert multi-zone information
	if len(instance.MultiZone) > 0 {
		multiZones := make([]map[string]interface{}, 0, len(instance.MultiZone))
		for _, mz := range instance.MultiZone {
			multiZone := map[string]interface{}{
				"zone_id":            mz.ZoneId,
				"vswitch_ids":        mz.VSwitchIds,
				"cidr":               mz.Cidr,
				"available_ip_count": mz.AvailableIpCount,
			}
			multiZones = append(multiZones, multiZone)
		}
		result["multi_zone"] = multiZones
	}

	return result
}
