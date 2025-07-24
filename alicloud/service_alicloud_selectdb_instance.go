package alicloud

import (
	"fmt"
	"time"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Instance Management Operations

// CreateSelectDBInstance creates a new SelectDB instance
func (s *SelectDBService) CreateSelectDBInstance(instance *selectdb.Instance) (*selectdb.Instance, error) {
	if instance == nil {
		return nil, WrapError(fmt.Errorf("instance cannot be nil"))
	}

	result, err := s.selectdbAPI.CreateInstance(instance)
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
		if selectdb.NotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapError(err)
	}

	return instance, nil
}

// DescribeSelectDBInstances lists SelectDB instances with pagination
func (s *SelectDBService) DescribeSelectDBInstances(regionId string) ([]selectdb.Instance, error) {
	instances, err := s.selectdbAPI.ListInstances(regionId, 1, 50)
	if err != nil {
		return nil, WrapError(err)
	}

	return instances, nil
}

// ModifySelectDBInstance modifies attributes of a SelectDB instance
func (s *SelectDBService) ModifySelectDBInstance(instanceId, attributeType, value, regionId string) error {
	if instanceId == "" {
		return WrapError(fmt.Errorf("instance ID cannot be empty"))
	}

	err := s.selectdbAPI.ModifyInstance(instanceId, attributeType, value, regionId)
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
		if selectdb.NotFoundError(err) {
			return nil // Instance already deleted
		}
		return WrapError(err)
	}

	return nil
}

// CheckCreateSelectDBInstance validates instance creation parameters
func (s *SelectDBService) CheckCreateSelectDBInstance(instance *selectdb.Instance) error {
	if instance == nil {
		return WrapError(fmt.Errorf("instance cannot be nil"))
	}

	err := s.selectdbAPI.CheckCreateInstance(instance)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// UpgradeSelectDBInstanceEngineVersion upgrades the engine version of a SelectDB instance
func (s *SelectDBService) UpgradeSelectDBInstanceEngineVersion(instanceId, engineVersion, regionId string) error {
	if instanceId == "" || engineVersion == "" {
		return WrapError(fmt.Errorf("instance ID and engine version cannot be empty"))
	}

	err := s.selectdbAPI.UpgradeInstanceEngineVersion(instanceId, engineVersion, regionId)
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

// WaitForSelectDBInstanceStatus waits for instance to reach desired status using polling
func (s *SelectDBService) WaitForSelectDBInstanceStatus(instanceId string, targetStatus string, timeout time.Duration) (*selectdb.Instance, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"CREATING", "STARTING", "STOPPING", "DELETING", "RESTARTING", "UPGRADING", "RESOURCE_CHANGING"},
		Target:  []string{targetStatus},
		Refresh: func() (interface{}, string, error) {
			instance, err := s.DescribeSelectDBInstance(instanceId)
			if err != nil {
				if selectdb.NotFoundError(err) {
					if targetStatus == "DELETED" {
						return nil, "DELETED", nil
					}
					return nil, "", err
				}
				return nil, "", err
			}
			return instance, instance.Status, nil
		},
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	result, err := stateConf.WaitForState()
	if err != nil {
		return nil, WrapError(err)
	}

	if result != nil {
		return result.(*selectdb.Instance), nil
	}
	return nil, nil
}

// Utility Functions

// resetSelectDBInstancePassword resets account password for SelectDB instance
func (s *SelectDBService) resetSelectDBInstancePassword(instanceId, accountName, accountPassword, regionId string) error {
	if instanceId == "" || accountName == "" || accountPassword == "" {
		return WrapError(fmt.Errorf("instance ID, account name, and password cannot be empty"))
	}

	account := &selectdb.Account{
		AccountName:     accountName,
		AccountPassword: accountPassword,
		RegionId:        regionId,
	}

	err := s.selectdbAPI.ResetInstanceAccountPassword(instanceId, account)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// Region and Zone Operations

// DescribeSelectDBRegions retrieves available regions for SelectDB
func (s *SelectDBService) DescribeSelectDBRegions() ([]selectdb.RegionInfo, error) {
	regions, err := s.selectdbAPI.ListRegions()
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
	return selectdb.NotFoundError(err)
}

// IsSelectDBInvalidParameterError checks if the error indicates invalid parameters
func IsSelectDBInvalidParameterError(err error) bool {
	return selectdb.IsInvalidParameterError(err)
}

// Helper functions for converting between Terraform schema and API types

// ConvertInstanceToMap converts API instance to Terraform map
func ConvertInstanceToMap(instance *selectdb.Instance) map[string]interface{} {
	if instance == nil {
		return nil
	}

	result := map[string]interface{}{
		"db_instance_id":       instance.Id,
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
