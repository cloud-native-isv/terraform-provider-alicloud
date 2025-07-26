package alicloud

import (
	"fmt"
	"time"

	"github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/selectdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Public Connection Management Operations

// AllocateSelectDBPublicConnection allocates public connection for a SelectDB instance
func (s *SelectDBService) AllocateSelectDBPublicConnection(connection *selectdb.PublicConnection) (*selectdb.PublicConnection, error) {
	if connection == nil {
		return nil, WrapError(fmt.Errorf("public connection cannot be nil"))
	}

	result, err := s.selectdbAPI.AllocatePublicConnection(connection)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// ReleaseSelectDBPublicConnection releases public connection for a SelectDB instance
func (s *SelectDBService) ReleaseSelectDBPublicConnection(connection *selectdb.PublicConnection) error {
	if connection == nil {
		return WrapError(fmt.Errorf("release connection cannot be nil"))
	}

	err := s.selectdbAPI.ReleasePublicConnection(connection)
	if err != nil {
		if selectdb.NotFoundError(err) {
			return nil // Connection already released
		}
		return WrapError(err)
	}

	return nil
}

// DescribeSelectDBPublicConnection checks if public connection exists for a SelectDB instance
func (s *SelectDBService) DescribeSelectDBPublicConnection(instanceId string) (bool, error) {
	if instanceId == "" {
		return false, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}

	// Get instance information to check for public connection
	instance, err := s.DescribeSelectDBInstance(instanceId)
	if err != nil {
		if NotFoundError(err) {
			return false, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return false, WrapError(err)
	}

	// Check if instance has a public connection string
	if instance != nil && instance.ConnectionString != "" {
		return true, nil
	}

	return false, nil
}

// State Management and Refresh Functions

// SelectDBPublicConnectionStateRefreshFunc returns a ResourceStateRefreshFunc for SelectDB public connection
func (s *SelectDBService) SelectDBPublicConnectionStateRefreshFunc(instanceId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		hasConnection, err := s.DescribeSelectDBPublicConnection(instanceId)
		if err != nil {
			if NotFoundError(err) {
				return nil, "NotFound", WrapErrorf(Error(GetNotFoundMessage("SelectDB Public Connection", instanceId)), NotFoundMsg, ProviderERROR)
			}
			return nil, "", WrapError(err)
		}

		if hasConnection {
			return hasConnection, "Available", nil
		}

		return hasConnection, "NotAvailable", nil
	}
}

// WaitForSelectDBPublicConnection waits for SelectDB public connection to reach expected status
func (s *SelectDBService) WaitForSelectDBPublicConnection(instanceId string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)

	for {
		hasConnection, err := s.DescribeSelectDBPublicConnection(instanceId)
		if err != nil {
			if NotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}

		currentStatus := "NotAvailable"
		if hasConnection {
			currentStatus = "Available"
		}

		if currentStatus == string(status) {
			return nil
		}

		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, instanceId, GetFunc(1), timeout, currentStatus, string(status), ProviderERROR)
		}

		time.Sleep(DefaultIntervalShort * time.Second)
	}
}

// Helper functions for converting between Terraform schema and API types

// ConvertToPublicConnection converts schema data to API public connection
func ConvertToPublicConnection(d *schema.ResourceData) *selectdb.PublicConnection {
	connection := &selectdb.PublicConnection{}

	if v, ok := d.GetOk("db_instance_id"); ok {
		connection.DBInstanceId = v.(string)
	}
	if v, ok := d.GetOk("connection_string_prefix"); ok {
		connection.ConnectionStringPrefix = v.(string)
	}
	if v, ok := d.GetOk("connection_string"); ok {
		connection.ConnectionString = v.(string)
	}
	if v, ok := d.GetOk("net_type"); ok {
		connection.NetType = v.(string)
	}
	if v, ok := d.GetOk("region_id"); ok {
		connection.RegionId = v.(string)
	}

	return connection
}

// ConvertPublicConnectionToMap converts API public connection to Terraform map
func ConvertPublicConnectionToMap(connection *selectdb.PublicConnection) map[string]interface{} {
	if connection == nil {
		return nil
	}

	return map[string]interface{}{
		"db_instance_id":           connection.DBInstanceId,
		"connection_string":        connection.ConnectionString,
		"connection_string_prefix": connection.ConnectionStringPrefix,
		"net_type":                 connection.NetType,
		"region_id":                connection.RegionId,
		"instance_name":            connection.InstanceName,
		"task_id":                  connection.TaskId,
	}
}
