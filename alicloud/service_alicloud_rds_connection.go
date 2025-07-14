package alicloud

import (
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// RDS Network and Connection related operations

func (s *RdsService) DescribeDBInstanceRwNetInfoByMssql(id string) ([]interface{}, error) {
	action := "DescribeDBInstanceNetInfo"
	request := map[string]interface{}{
		"RegionId":                 s.client.RegionId,
		"DBInstanceId":             id,
		"SourceIp":                 s.client.SourceIp,
		"DBInstanceNetRWSplitType": "ReadWriteSplitting",
	}
	client := s.client
	var response map[string]interface{}
	var err error
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
		if err != nil {
			if IsExpectedErrors(err, []string{"InternalError"}) {
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
	return response["DBInstanceNetInfos"].(map[string]interface{})["DBInstanceNetInfo"].([]interface{}), nil
}

func (s *RdsService) DescribeDBInstanceNetInfo(id string) ([]interface{}, error) {
	action := "DescribeDBInstanceNetInfo"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": id,
		"SourceIp":     s.client.SourceIp,
	}
	client := s.client
	var response map[string]interface{}
	var err error
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
		if err != nil {
			if IsExpectedErrors(err, []string{"InternalError"}) {
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
	return response["DBInstanceNetInfos"].(map[string]interface{})["DBInstanceNetInfo"].([]interface{}), nil
}

func (s *RdsService) DescribeDBConnection(id string) (map[string]interface{}, error) {
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return nil, WrapError(err)
	}
	object, err := s.DescribeDBInstanceNetInfo(parts[0])

	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidCurrentConnectionString.NotFound"}) {
			return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return nil, WrapError(err)
	}

	if object != nil {
		for _, o := range object {
			o := o.(map[string]interface{})
			if strings.HasPrefix(o["ConnectionString"].(string), parts[1]) {
				return o, nil
			}
		}
	}

	return nil, WrapErrorf(NotFoundErr("DBConnection", id), NotFoundMsg, ProviderERROR)
}

func (s *RdsService) DescribeDBReadWriteSplittingConnection(id string) (map[string]interface{}, error) {
	object, err := s.DescribeDBInstanceRwNetInfoByMssql(id)
	if err != nil && !NotFoundError(err) {
		return nil, err
	}

	if object != nil {
		for _, conn := range object {
			conn := conn.(map[string]interface{})
			if conn["ConnectionStringType"] != "ReadWriteSplitting" {
				continue
			}
			if _, ok := conn["MaxDelayTime"]; ok {
				if conn["MaxDelayTime"] == nil {
					continue
				}
				if _, err := strconv.Atoi(conn["MaxDelayTime"].(string)); err != nil {
					return nil, err
				}
			}
			return conn, nil
		}
	}

	return nil, WrapErrorf(NotFoundErr("ReadWriteSplittingConnection", id), NotFoundMsg, ProviderERROR)
}

func (s *RdsService) ReleaseDBPublicConnection(instanceId, connection string) error {
	action := "ReleaseInstancePublicConnection"
	request := map[string]interface{}{
		"RegionId":                s.client.RegionId,
		"DBInstanceId":            instanceId,
		"CurrentConnectionString": connection,
		"SourceIp":                s.client.SourceIp,
	}
	client := s.client
	response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, false)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, instanceId, action, AlibabaCloudSdkGoERROR)
	}
	addDebug(action, response, request)
	return nil
}

func (s *RdsService) WaitForDBConnection(id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeDBConnection(id)
		if err != nil {
			if NotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}
		if object != nil && object["ConnectionString"] != "" {
			return nil
		}
		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), timeout, object["ConnectionString"], id, ProviderERROR)
		}
	}
}

func (s *RdsService) WaitForDBReadWriteSplitting(id string, status Status, timeout int) error {
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeDBReadWriteSplittingConnection(id)
		if err != nil {
			if NotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}
		if err == nil {
			break
		}
		time.Sleep(DefaultIntervalShort * time.Second)
		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), timeout, object["ConnectionString"], id, ProviderERROR)
		}
	}
	return nil
}
