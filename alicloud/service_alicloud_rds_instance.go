package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// RDS Instance related operations

func (s *RdsService) DescribeDBInstance(id string) (object map[string]interface{}, err error) {
	client := s.client
	action := "DescribeDBInstanceAttribute"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": id,
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
		if IsExpectedErrors(err, []string{"InvalidDBInstanceId.NotFound", "InvalidDBInstanceName.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	v, err := jsonpath.Get("$.Items.DBInstanceAttribute", response)
	if err != nil {
		return nil, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Items.DBInstanceAttribute", response)
	}
	if len(v.([]interface{})) < 1 {
		return nil, WrapErrorf(NotFoundErr("DBInstance", id), NotFoundMsg, ProviderERROR)
	}
	return v.([]interface{})[0].(map[string]interface{}), nil
}

func (s *RdsService) DescribeDBReadonlyInstance(id string) (object map[string]interface{}, err error) {
	action := "DescribeDBInstanceAttribute"
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
	dBInstanceAttributes := response["Items"].(map[string]interface{})["DBInstanceAttribute"].([]interface{})
	if len(dBInstanceAttributes) < 1 {
		return nil, WrapErrorf(NotFoundErr("DBInstance", id), NotFoundMsg, ProviderERROR)
	}

	return dBInstanceAttributes[0].(map[string]interface{}), nil
}

func (s *RdsService) DescribeRdsCloneDbInstance(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "DescribeDBInstanceAttribute"
	request := map[string]interface{}{
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
	v, err := jsonpath.Get("$.Items.DBInstanceAttribute", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Items.DBInstanceAttribute", response)
	}
	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("RDS", id), NotFoundWithResponse, response)
	} else {
		if fmt.Sprint(v.([]interface{})[0].(map[string]interface{})["DBInstanceId"]) != id {
			return object, WrapErrorf(NotFoundErr("RDS", id), NotFoundWithResponse, response)
		}
	}
	object = v.([]interface{})[0].(map[string]interface{})
	return object, nil
}

func (s *RdsService) DescribeTasks(id string) (object map[string]interface{}, err error) {
	action := "DescribeTasks"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": id,
		"SourceIp":     s.client.SourceIp,
	}
	client := s.client
	response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidDBInstanceId.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, ProviderERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	addDebug(action, response, request)
	return response, nil
}

func (s *RdsService) DescribeDBInstanceHAConfig(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "DescribeDBInstanceHAConfig"
	request := map[string]interface{}{
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

// WaitForInstance waits for instance to given status
func (s *RdsService) WaitForDBInstance(id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeDBInstance(id)
		if err != nil {
			if NotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}
		if object != nil && strings.ToLower(fmt.Sprint(object["DBInstanceStatus"])) == strings.ToLower(string(status)) {
			break
		}
		time.Sleep(DefaultIntervalShort * time.Second)
		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), timeout, object["DBInstanceStatus"], status, ProviderERROR)
		}
	}
	return nil
}

func (s *RdsService) RdsDBInstanceStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeDBInstance(id)
		if err != nil {
			if NotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if object["DBInstanceStatus"] == failState {
				return object, fmt.Sprint(object["DBInstanceStatus"]), WrapError(Error(FailedToReachTargetStatus, object["DBInstanceStatus"]))
			}
		}
		return object, fmt.Sprint(object["DBInstanceStatus"]), nil
	}
}

func (s *RdsService) RdsDBInstanceNodeIdRefreshFunc(id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		describeDBInstanceHAConfigObject, err := s.DescribeDBInstanceHAConfig(id)
		if err != nil {
			if NotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		hostInstanceInfos := describeDBInstanceHAConfigObject["HostInstanceInfos"].(map[string]interface{})["NodeInfo"].([]interface{})
		var nodeId string
		for _, val := range hostInstanceInfos {
			item := val.(map[string]interface{})
			nodeType := item["NodeType"].(string)
			if nodeType == "Master" {
				nodeId = item["NodeId"].(string)
				break // 停止遍历
			}
		}
		return describeDBInstanceHAConfigObject, fmt.Sprint(nodeId), nil
	}
}

func (s *RdsService) RdsTaskStateRefreshFunc(id string, taskAction string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeTasks(id)
		if err != nil {
			if NotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		taskProgressInfos := object["Items"].(map[string]interface{})["TaskProgressInfo"].([]interface{})
		for _, t := range taskProgressInfos {
			t := t.(map[string]interface{})
			if t["TaskAction"] == taskAction {
				return object, t["Status"].(string), nil
			}
		}

		return object, "Pending", nil
	}
}

func (s *RdsService) ModifyDBInstanceDeletionProtection(d *schema.ResourceData, attribute string) error {
	client := s.client
	var err error
	action := "ModifyDBInstanceDeletionProtection"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": d.Id(),
		"SourceIp":     s.client.SourceIp,
	}
	request["DeletionProtection"] = d.Get("deletion_protection")
	var response map[string]interface{}
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
