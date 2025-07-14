package alicloud

import (
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// RDS Database related operations

func (s *RdsService) DescribeDBDatabase(id string) (object map[string]interface{}, err error) {
	var ds map[string]interface{}
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return ds, WrapError(err)
	}
	dbName := parts[1]
	var response map[string]interface{}
	action := "DescribeDatabases"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": parts[0],
		"DBName":       dbName,
		"SourceIp":     s.client.SourceIp,
	}
	client := s.client
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
		if err != nil {
			if IsExpectedErrors(err, []string{"InternalError", "OperationDenied.DBInstanceStatus"}) {
				return resource.RetryableError(WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR))
			}
			if NotFoundError(err) || IsExpectedErrors(err, []string{"InvalidDBName.NotFound", "InvalidDBInstanceId.NotFoundError"}) {
				return resource.NonRetryableError(WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR))
			}
			return resource.NonRetryableError(WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR))
		}
		addDebug(action, response, request)
		v, err := jsonpath.Get("$.Databases.Database", response)
		if err != nil {
			return resource.NonRetryableError(WrapErrorf(err, FailedGetAttributeMsg, id, "$.Databases.Database", response))
		}
		if len(v.([]interface{})) < 1 {
			return resource.NonRetryableError(WrapErrorf(NotFoundErr("DBDatabase", dbName), NotFoundMsg, ProviderERROR))
		}
		ds = v.([]interface{})[0].(map[string]interface{})
		return nil
	})
	return ds, err
}

func (s *RdsService) WaitForDBDatabase(id string, status Status, timeout int) error {
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return WrapError(err)
	}
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeDBDatabase(id)
		if err != nil {
			if NotFoundError(err) {
				if status == Deleted {
					return nil
				}
			}
			return WrapError(err)
		}
		if object != nil && object["DBName"] == parts[1] {
			break
		}
		time.Sleep(DefaultIntervalShort * time.Second)
		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), timeout, object["DBName"], parts[1], ProviderERROR)
		}
	}
	return nil
}
