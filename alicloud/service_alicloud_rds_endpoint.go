package alicloud

import (
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeRdsNode describes RDS node information
func (s *RdsService) DescribeRdsNode(id string) (object map[string]interface{}, err error) {
	client := s.client
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return nil, WrapError(err)
	}
	action := "DescribeDBInstanceAttribute"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": parts[0],
		"SourceIp":     s.client.SourceIp,
	}
	var response map[string]interface{}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
		if err != nil {
			if NeedRetry(err) || IsExpectedErrors(err, []string{"InvalidParameter"}) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidDBInstanceId.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	v, err := jsonpath.Get("$.Items.DBInstanceAttribute", response)
	if err != nil {
		return nil, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Items.DBInstanceAttribute", response)
	}
	if len(v.([]interface{})) < 1 {
		return nil, WrapErrorf(NotFoundErr("DBAccount", id), NotFoundMsg, ProviderERROR)
	}

	dbNodeList := v.([]interface{})[0].(map[string]interface{})
	if dbNodesList, ok := dbNodeList["DBClusterNodes"]; ok && dbNodesList != nil {
		if nodeList, ok := dbNodesList.(map[string]interface{})["DBClusterNode"].([]interface{}); ok {
			if len(nodeList) < 3 {
				return object, WrapErrorf(NotFoundErr("DBNode", id), NotFoundMsg, ProviderERROR)
			}
			for _, node := range nodeList {
				node := node.(map[string]interface{})
				if node["NodeId"].(string) == parts[1] {
					object = node
					break
				}
			}
		}
	}
	return object, nil
}

// DescribeDBInstanceEndpoints describes RDS DB instance endpoints
func (s *RdsService) DescribeDBInstanceEndpoints(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return nil, WrapError(err)
	}
	_, err = s.DescribeDBInstance(parts[0])
	if err != nil {
		return nil, WrapError(err)
	}
	action := "DescribeDBInstanceEndpoints"
	request := map[string]interface{}{
		"SourceIp":             s.client.SourceIp,
		"DBInstanceId":         parts[0],
		"DBInstanceEndpointId": parts[1],
		"RegionId":             s.client.RegionId,
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(3*time.Minute, func() *resource.RetryError {
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
		if IsExpectedErrors(err, []string{"InvalidDBInstance.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	v, err := jsonpath.Get("$.Data.DBInstanceEndpoints", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Data.DBInstanceEndpoints", response)
	}
	if endpoints, ok := v.(map[string]interface{})["DBInstanceEndpoint"].([]interface{}); ok {
		if len(endpoints) < 1 {
			return nil, WrapErrorf(NotFoundErr("DBInstanceEndpoint", id), NotFoundMsg, ProviderERROR)
		}
		endpoint := endpoints[0].(map[string]interface{})
		object = make(map[string]interface{})
		object["EndpointDescription"] = endpoint["EndpointDescription"]
		object["EndpointType"] = endpoint["EndpointType"]
		object["DBInstanceId"] = parts[0]
		object["DBInstanceEndpointId"] = parts[1]
		nodeList := endpoint["NodeItems"]
		nodeItems := nodeList.(map[string]interface{})["NodeItem"].([]interface{})
		dbNodesMaps := make([]map[string]interface{}, 0)
		if nodeItems != nil && len(nodeItems) > 0 {
			for _, nodeItem := range nodeItems {
				nodeItem := nodeItem.(map[string]interface{})
				dbNodesMap := map[string]interface{}{
					"NodeId": nodeItem["NodeId"],
					"Weight": nodeItem["Weight"],
				}
				dbNodesMaps = append(dbNodesMaps, dbNodesMap)
			}
			object["NodeItems"] = dbNodesMaps
		}
		addressList := endpoint["AddressItems"]
		addressItems := addressList.(map[string]interface{})["AddressItem"].([]interface{})
		if addressItems != nil && len(addressItems) > 0 {
			for _, addressItem := range addressItems {
				addressItem := addressItem.(map[string]interface{})
				object["ConnectionString"] = addressItem["ConnectionString"]
				object["IpAddress"] = addressItem["IpAddress"]
				object["IpType"] = addressItem["IpType"]
				object["Port"] = addressItem["Port"]
				object["VPCId"] = addressItem["VPCId"]
				object["VSwitchId"] = addressItem["VSwitchId"]
			}
		}
	}
	return object, nil
}

// DescribeDBInstanceEndpointPublicAddress describes RDS DB instance endpoint public address
func (s *RdsService) DescribeDBInstanceEndpointPublicAddress(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return nil, WrapError(err)
	}
	action := "DescribeDBInstanceEndpoints"
	request := map[string]interface{}{
		"SourceIp":             s.client.SourceIp,
		"DBInstanceId":         parts[0],
		"DBInstanceEndpointId": parts[1],
		"RegionId":             s.client.RegionId,
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(3*time.Minute, func() *resource.RetryError {
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
		if IsExpectedErrors(err, []string{"InvalidDBInstance.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	v, err := jsonpath.Get("$.Data.DBInstanceEndpoints", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Data.DBInstanceEndpoints", response)
	}
	if endpoints, ok := v.(map[string]interface{})["DBInstanceEndpoint"].([]interface{}); ok {
		if len(endpoints) < 1 {
			return nil, WrapErrorf(NotFoundErr("DBInstanceEndpoint", id), NotFoundMsg, ProviderERROR)
		}
		endpoint := endpoints[0].(map[string]interface{})
		object = make(map[string]interface{})
		object["DBInstanceId"] = parts[0]
		object["DBInstanceEndpointId"] = parts[1]
		addressList := endpoint["AddressItems"]
		addressItems := addressList.(map[string]interface{})["AddressItem"].([]interface{})
		if addressItems != nil && len(addressItems) > 0 {
			for _, addressItem := range addressItems {
				addressItem := addressItem.(map[string]interface{})
				if addressItem["IpType"] == "Public" {
					object["ConnectionString"] = addressItem["ConnectionString"]
					object["IpAddress"] = addressItem["IpAddress"]
					object["IpType"] = addressItem["IpType"]
					object["Port"] = addressItem["Port"]
					object["VPCId"] = addressItem["VPCId"]
					object["VSwitchId"] = addressItem["VSwitchId"]
					break
				}
			}
		}
		if _, ok := object["IpType"]; !ok {
			return nil, WrapErrorf(NotFoundErr("DBInstanceEndpointPublicAddress", id), NotFoundMsg, ProviderERROR)
		}
	}
	return object, nil
}
