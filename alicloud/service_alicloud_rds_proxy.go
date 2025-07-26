package alicloud

import (
	"fmt"
	"log"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeDBProxy describes RDS database proxy
func (s *RdsService) DescribeDBProxy(id string) (object map[string]interface{}, err error) {
	action := "DescribeDBProxy"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": id,
		"SourceIp":     s.client.SourceIp,
	}
	client := s.client
	response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidDBInstanceId.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	addDebug(action, response, request)
	object = make(map[string]interface{})
	v, err := jsonpath.Get("$.DBProxyConnectStringItems", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.DBProxyConnectStringItems", response)
	}
	if dbProxyInstanceType, ok := response["DBProxyInstanceType"]; ok {
		object["DBProxyInstanceType"] = dbProxyInstanceType
	}
	object["DBProxyInstanceStatus"] = response["DBProxyInstanceStatus"]
	if dBProxyInstanceNum, ok := response["DBProxyInstanceNum"]; ok {
		object["DBProxyInstanceNum"] = dBProxyInstanceNum
	}
	if dBProxyPersistentConnectionStatus, ok := response["DBProxyPersistentConnectionStatus"]; ok {
		object["DBProxyPersistentConnectionStatus"] = dBProxyPersistentConnectionStatus
	}
	if dBProxyInstanceCurrentMinorVersion, ok := response["DBProxyInstanceCurrentMinorVersion"]; ok {
		object["DBProxyInstanceCurrentMinorVersion"] = dBProxyInstanceCurrentMinorVersion
	}
	if dBProxyInstanceLatestMinorVersion, ok := response["DBProxyInstanceLatestMinorVersion"]; ok {
		object["DBProxyInstanceLatestMinorVersion"] = dBProxyInstanceLatestMinorVersion
	}
	if dBProxyServiceStatus, ok := response["DBProxyServiceStatus"]; ok {
		object["DBProxyServiceStatus"] = dBProxyServiceStatus
	}
	if dBProxyServiceStatus, ok := response["DBProxyInstanceName"]; ok {
		object["DBProxyInstanceName"] = dBProxyServiceStatus
	}
	if dBProxyConnectStringItems, ok := v.(map[string]interface{})["DBProxyConnectStringItems"].([]interface{}); ok {
		var innerItem, outerItem map[string]interface{}
		for _, item := range dBProxyConnectStringItems {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if dbProxyEndpointAliases, ok := itemMap["DbProxyEndpointAliases"].(string); ok && dbProxyEndpointAliases == "InnerString" {
					innerItem = itemMap
				}
				if dbProxyEndpointAliases, ok := itemMap["DbProxyEndpointAliases"].(string); ok && dbProxyEndpointAliases == "OuterString" {
					outerItem = itemMap
				}
			}
		}

		if innerItem == nil {
			return nil, WrapErrorf(NotFoundErr("DBProxyConnectStringItems", id), NotFoundMsg, ProviderERROR)
		}

		object["DBProxyVpcId"] = innerItem["DBProxyVpcId"]
		object["DBProxyVswitchId"] = innerItem["DBProxyVswitchId"]

		if outerItem != nil {
			object["DBProxyConnectString"] = outerItem["DBProxyConnectString"]
			object["DBProxyConnectStringPort"] = outerItem["DBProxyConnectStringPort"]
		} else {
			log.Printf("[WARN] No OuterString item found for resource %s", id)
		}
	}

	return object, nil
}

// DescribeDBProxyEndpoint describes RDS database proxy endpoint with specific endpoint name
func (s *RdsService) DescribeDBProxyEndpoint(id string, endpointName string) (object map[string]interface{}, err error) {
	action := "DescribeDBProxyEndpoint"
	request := map[string]interface{}{
		"RegionId":          s.client.RegionId,
		"DBInstanceId":      id,
		"DBProxyEndpointId": endpointName,
		"SourceIp":          s.client.SourceIp,
	}
	client := s.client
	response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidDBInstanceId.NotFound", "Endpoint.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	addDebug(action, response, request)
	return response, nil
}

// DescribeRdsProxyEndpoint describes RDS proxy endpoint without specific endpoint name
func (s *RdsService) DescribeRdsProxyEndpoint(id string) (object map[string]interface{}, err error) {
	action := "DescribeDBProxyEndpoint"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": id,
		"SourceIp":     s.client.SourceIp,
	}
	client := s.client
	response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidDBInstanceId.NotFound", "Endpoint.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	addDebug(action, response, request)
	return response, nil
}

// RdsDBProxyStateRefreshFunc returns a StateRefreshFunc for RDS DB proxy status
func (s *RdsService) RdsDBProxyStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeDBProxy(id)
		if err != nil {
			if IsNotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		for _, failState := range failStates {
			if object["DBProxyInstanceStatus"] == failState {
				return object, object["DBProxyInstanceStatus"].(string), WrapError(Error(FailedToReachTargetStatus, object["DBProxyInstanceStatus"].(string)))
			}
		}
		return object, fmt.Sprint(object["DBProxyInstanceStatus"]), nil
	}
}

// GetDbProxyInstanceSsl gets DB proxy instance SSL configuration
func (s *RdsService) GetDbProxyInstanceSsl(id string) (object map[string]interface{}, err error) {
	action := "GetDbProxyInstanceSsl"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": id,
		"SourceIp":     s.client.SourceIp,
	}
	client := s.client
	response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidDBInstanceId.NotFound"}) {
			return object, WrapErrorf(NotFoundErr("RDS", id), NotFoundWithResponse, response)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	addDebug(action, response, request)
	v, err := jsonpath.Get("$.DbProxyCertListItems.DbProxyCertListItems", response)
	if err != nil {
		return object, WrapErrorf(NotFoundErr("RDS", id), NotFoundWithResponse, response)
	}
	if len(v.([]interface{})) < 1 {
		return object, nil
	}
	return v.([]interface{})[0].(map[string]interface{}), nil
}
