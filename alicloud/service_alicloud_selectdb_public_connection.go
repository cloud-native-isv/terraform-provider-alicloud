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

	result, err := s.GetAPI().AllocatePublicConnection(connection)
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

	err := s.GetAPI().ReleasePublicConnection(connection)
	if err != nil {
		if selectdb.IsNotFoundError(err) {
			return nil // Connection already released
		}
		return WrapError(err)
	}

	return nil
}

// DescribeSelectDBPublicConnection retrieves the public connection information for a SelectDB instance
func (s *SelectDBService) DescribeSelectDBPublicConnection(instanceId string) (*selectdb.PublicConnection, error) {
	if instanceId == "" {
		return nil, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}

	// Get network information to check for public connection
	networkInfo, err := s.GetAPI().GetNetworkInfo(instanceId)
	if err != nil {
		if NotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapError(err)
	}

	// Use the GetPublicConnection method to extract public connection
	publicConnection := networkInfo.GetPublicConnection()
	if publicConnection == nil {
		return nil, WrapErrorf(Error(GetNotFoundMessage("SelectDB Public Connection", instanceId)), NotFoundMsg, ProviderERROR)
	}

	// Set the instance ID in the returned connection object
	publicConnection.DBInstanceId = instanceId
	publicConnection.RegionId = s.GetRegionId()

	return publicConnection, nil
}

// GetSelectDBPublicConnections retrieves all public connection information for a SelectDB instance
func (s *SelectDBService) GetSelectDBPublicConnections(instanceId string) ([]*selectdb.PublicConnection, error) {
	if instanceId == "" {
		return nil, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}

	// Get network information to check for public connections
	networkInfo, err := s.GetAPI().GetNetworkInfo(instanceId)
	if err != nil {
		if NotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapError(err)
	}

	// Use the GetPublicConnections method to extract all public connections
	publicConnections := networkInfo.GetPublicConnections()
	if len(publicConnections) == 0 {
		return nil, WrapErrorf(Error(GetNotFoundMessage("SelectDB Public Connection", instanceId)), NotFoundMsg, ProviderERROR)
	}

	// Set the instance ID and region in all returned connection objects
	for _, conn := range publicConnections {
		conn.DBInstanceId = instanceId
		conn.RegionId = s.GetRegionId()
	}

	return publicConnections, nil
}

// GetSelectDBMySQLConnection retrieves the MySQL public connection for a SelectDB instance
func (s *SelectDBService) GetSelectDBMySQLConnection(instanceId string) (*selectdb.PublicConnection, error) {
	connections, err := s.GetSelectDBPublicConnections(instanceId)
	if err != nil {
		return nil, err
	}

	// Find MySQL connection
	for _, conn := range connections {
		if conn.Protocol == "MySQLPort" {
			return conn, nil
		}
	}

	return nil, WrapErrorf(Error(GetNotFoundMessage("SelectDB MySQL Public Connection", instanceId)), NotFoundMsg, ProviderERROR)
}

// GetSelectDBHTTPConnection retrieves the HTTP public connection for a SelectDB instance
func (s *SelectDBService) GetSelectDBHTTPConnection(instanceId string) (*selectdb.PublicConnection, error) {
	connections, err := s.GetSelectDBPublicConnections(instanceId)
	if err != nil {
		return nil, err
	}

	// Find HTTP connection
	for _, conn := range connections {
		if conn.Protocol == "HttpPort" {
			return conn, nil
		}
	}

	return nil, WrapErrorf(Error(GetNotFoundMessage("SelectDB HTTP Public Connection", instanceId)), NotFoundMsg, ProviderERROR)
}

// GetSelectDBAllConnections retrieves all connection information (VPC and Public) for a SelectDB instance
func (s *SelectDBService) GetSelectDBAllConnections(instanceId string) ([]*selectdb.PublicConnection, error) {
	if instanceId == "" {
		return nil, WrapError(fmt.Errorf("instance ID cannot be empty"))
	}

	// Get network information to check for all connections
	networkInfo, err := s.GetAPI().GetNetworkInfo(instanceId)
	if err != nil {
		if NotFoundError(err) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapError(err)
	}

	// Use the GetAllConnections method to extract all connections
	allConnections := networkInfo.GetAllConnections()
	if len(allConnections) == 0 {
		return nil, WrapErrorf(Error(GetNotFoundMessage("SelectDB Connections", instanceId)), NotFoundMsg, ProviderERROR)
	}

	// Set the instance ID and region in all returned connection objects
	for _, conn := range allConnections {
		conn.DBInstanceId = instanceId
		conn.RegionId = s.GetRegionId()
	}

	return allConnections, nil
}

// State Management and Refresh Functions

// SelectDBPublicConnectionStateRefreshFunc returns a ResourceStateRefreshFunc for SelectDB public connection
func (s *SelectDBService) SelectDBPublicConnectionStateRefreshFunc(instanceId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		publicConnection, err := s.DescribeSelectDBPublicConnection(instanceId)
		if err != nil {
			if NotFoundError(err) {
				return nil, "NotFound", WrapErrorf(Error(GetNotFoundMessage("SelectDB Public Connection", instanceId)), NotFoundMsg, ProviderERROR)
			}
			return nil, "", WrapError(err)
		}

		if publicConnection != nil {
			return publicConnection, "Available", nil
		}

		return nil, "NotAvailable", nil
	}
}

// WaitForSelectDBPublicConnection waits for SelectDB public connection to reach expected status
func (s *SelectDBService) WaitForSelectDBPublicConnection(instanceId string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)

	for {
		publicConnection, err := s.DescribeSelectDBPublicConnection(instanceId)
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
		if publicConnection != nil {
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
func ConvertToPublicConnection(d *schema.ResourceData, service *SelectDBService) *selectdb.PublicConnection {
	connection := &selectdb.PublicConnection{}

	// Map instance_id from schema to DBInstanceId in API
	if v, ok := d.GetOk("instance_id"); ok {
		connection.DBInstanceId = v.(string)
	}

	if v, ok := d.GetOk("connection_string_prefix"); ok {
		connection.ConnectionStringPrefix = v.(string)
	}

	// Set fixed network type for public connection
	connection.NetType = "Public"

	// Use service's GetRegionId() for region
	connection.RegionId = service.GetRegionId()

	return connection
}
