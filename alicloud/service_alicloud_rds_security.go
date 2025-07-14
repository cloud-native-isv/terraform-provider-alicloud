package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// RDS Security and Configuration related operations

func (s *RdsService) DescribeDBSecurityIps(instanceId string) ([]interface{}, error) {
	action := "DescribeDBInstanceIPArrayList"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": instanceId,
		"SourceIp":     s.client.SourceIp,
	}
	client := s.client
	var response map[string]interface{}
	var err error
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
		addDebug(action, response, request)
		return nil
	})
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, instanceId, action, AlibabaCloudSdkGoERROR)
	}
	return response["Items"].(map[string]interface{})["DBInstanceIPArray"].([]interface{}), nil
}

func (s *RdsService) DescribeSecurityGroupConfiguration(id string) ([]string, error) {
	action := "DescribeSecurityGroupConfiguration"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": id,
		"SourceIp":     s.client.SourceIp,
	}
	client := s.client
	var response map[string]interface{}
	var err error
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
		addDebug(action, response, request)
		return nil
	})
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	groupIds := make([]string, 0)
	ecsSecurityGroupRelations := response["Items"].(map[string]interface{})["EcsSecurityGroupRelation"].([]interface{})
	for _, v := range ecsSecurityGroupRelations {
		v := v.(map[string]interface{})
		groupIds = append(groupIds, v["SecurityGroupId"].(string))
	}
	return groupIds, nil
}

func (s *RdsService) DescribeDBInstanceSSL(id string) (object map[string]interface{}, err error) {
	action := "DescribeDBInstanceSSL"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": id,
		"SourceIp":     s.client.SourceIp,
	}
	client := s.client
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
		addDebug(action, response, request)
		return nil
	})
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	return response, nil
}

func (s *RdsService) DescribeDBInstanceEncryptionKey(id string) (object map[string]interface{}, err error) {
	action := "DescribeDBInstanceEncryptionKey"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": id,
		"SourceIp":     s.client.SourceIp,
	}
	client := s.client
	response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	addDebug(action, response, request)
	return response, nil
}

func (s *RdsService) DescribeRdsTDEInfo(id string) (object map[string]interface{}, err error) {
	action := "DescribeDBInstanceTDE"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": id,
		"SourceIp":     s.client.SourceIp,
	}
	statErr := s.WaitForDBInstance(id, Running, DefaultLongTimeout)
	if statErr != nil {
		return nil, WrapError(statErr)
	}
	client := s.client
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
		addDebug(action, response, request)
		return nil
	})
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	return response, nil
}

func (s *RdsService) DescribeHASwitchConfig(id string) (object map[string]interface{}, err error) {
	action := "DescribeHASwitchConfig"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": id,
		"SourceIp":     s.client.SourceIp,
	}
	client := s.client
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
		addDebug(action, response, request)
		return nil
	})
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	return response, nil
}

func (s *RdsService) DescribeHADiagnoseConfig(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "DescribeHADiagnoseConfig"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"SourceIp":     s.client.SourceIp,
		"DBInstanceId": id,
	}
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
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	v, err := jsonpath.Get("$", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$", response)
	}
	object = v.(map[string]interface{})
	return object, nil
}

func (s *RdsService) ModifyDBSecurityIps(instanceId, ips string) error {
	client := s.client
	action := "ModifySecurityIps"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": instanceId,
		"SecurityIps":  ips,
		"SourceIp":     s.client.SourceIp,
	}
	response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, false)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, instanceId, action, AlibabaCloudSdkGoERROR)
	}
	addDebug(action, response, request)

	if err := s.WaitForDBInstance(instanceId, Running, DefaultTimeoutMedium); err != nil {
		return WrapError(err)
	}
	return nil
}

func (s *RdsService) ModifySecurityGroupConfiguration(id string, groupid string) error {
	client := s.client
	action := "ModifySecurityGroupConfiguration"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": id,
		"SourceIp":     s.client.SourceIp,
	}
	//openapi required that input "Empty" if groupid is ""
	if len(groupid) == 0 {
		groupid = "Empty"
	}
	request["SecurityGroupId"] = groupid
	var response map[string]interface{}
	var err error
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("Rds", "2014-08-15", action, nil, request, false)
		if err != nil {
			if NeedRetry(err) || IsExpectedErrors(err, []string{"ServiceUnavailable"}) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	addDebug(action, response, request)
	return nil
}

func (s *RdsService) ModifyHADiagnoseConfig(d *schema.ResourceData, attribute string) error {
	client := s.client
	action := "ModifyHADiagnoseConfig"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": d.Id(),
		"SourceIp":     s.client.SourceIp,
	}
	request["TcpConnectionType"] = d.Get("tcp_connection_type")
	var response map[string]interface{}
	var err error
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("Rds", "2014-08-15", action, nil, request, false)
		if err != nil {
			if IsExpectedErrors(err, []string{"InternalError"}) || NeedRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})
	if err != nil {
		return WrapError(err)
	}
	if err := s.WaitForDBInstance(d.Id(), Running, DefaultLongTimeout); err != nil {
		return WrapError(err)
	}
	d.SetPartial(attribute)
	return nil
}

func (s *RdsService) GetSecurityIps(instanceId string, dbInstanceIpArrayName string) ([]string, error) {
	object, err := s.DescribeDBSecurityIps(instanceId)
	if err != nil {
		return nil, WrapError(err)
	}

	var ips, separator string
	ipsMap := make(map[string]string)
	for _, ip := range object {
		ip := ip.(map[string]interface{})
		if ip["DBInstanceIPArrayName"].(string) == dbInstanceIpArrayName {
			ips += separator + ip["SecurityIPList"].(string)
			separator = COMMA_SEPARATED
		}
	}

	for _, ip := range strings.Split(ips, COMMA_SEPARATED) {
		ipsMap[ip] = ip
	}

	var finalIps []string
	if len(ipsMap) > 0 {
		for key := range ipsMap {
			finalIps = append(finalIps, key)
		}
	}

	return finalIps, nil
}

func (s *RdsService) flattenDBSecurityIPs(list []interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, i := range list {
		i := i.(map[string]interface{})
		l := map[string]interface{}{
			"security_ips": i["SecurityIPList"],
		}
		result = append(result, l)
	}
	return result
}

func (s *RdsService) FindKmsRoleArnDdr(k string) (string, error) {
	action := "DescribeKey"
	var response map[string]interface{}
	var err error
	client := s.client
	request := make(map[string]interface{})
	request["KeyId"] = k
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("Kms", "2016-01-20", action, nil, request, true)
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
		return "", WrapErrorf(err, DataDefaultErrorMsg, k, action, AlibabaCloudSdkGoERROR)
	}
	resp, err := jsonpath.Get("$.KeyMetadata.Creator", response)
	if err != nil {
		return "", WrapErrorf(err, FailedGetAttributeMsg, action, "$.VersionIds.VersionId", response)
	}
	return strings.Join([]string{"acs:ram::", fmt.Sprint(resp), ":role/aliyunrdsinstanceencryptiondefaultrole"}, ""), nil
}
