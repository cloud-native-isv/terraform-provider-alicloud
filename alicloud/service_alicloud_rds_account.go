package alicloud

import (
	"log"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// RDS Account related operations

func (s *RdsService) DescribeDBAccountPrivilege(id string) (object map[string]interface{}, err error) {
	var ds map[string]interface{}
	parts, err := ParseResourceId(id, 3)
	if err != nil {
		return ds, WrapError(err)
	}
	action := "DescribeAccounts"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": parts[0],
		"AccountName":  parts[1],
		"SourceIp":     s.client.SourceIp,
	}
	client := s.client
	invoker := NewInvoker()
	invoker.AddCatcher(DBInstanceStatusCatcher)
	var response map[string]interface{}
	if err := invoker.Run(func() error {
		response, err = client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
		}
		addDebug(action, response, request)
		return nil
	}); err != nil {
		if IsExpectedErrors(err, []string{"InvalidDBInstanceId.NotFound"}) {
			return ds, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return ds, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	dBInstanceAccounts := response["Accounts"].(map[string]interface{})["DBInstanceAccount"].([]interface{})
	if len(dBInstanceAccounts) < 1 {
		return ds, WrapErrorf(NotFoundErr("DBAccountPrivilege", id), NotFoundMsg, ProviderERROR)
	}
	return dBInstanceAccounts[0].(map[string]interface{}), nil
}

func (s *RdsService) DescribeRdsAccount(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "DescribeAccounts"
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		err = WrapError(err)
		return
	}
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"SourceIp":     s.client.SourceIp,
		"AccountName":  parts[1],
		"DBInstanceId": parts[0],
	}
	response, err = client.RpcPost("Rds", "2014-08-15", action, nil, request, true)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidDBInstanceId.NotFound"}) {
			err = WrapErrorf(NotFoundErr("RdsAccount", id), NotFoundMsg, ProviderERROR)
			return object, err
		}
		err = WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
		return object, err
	}
	addDebug(action, response, request)
	v, err := jsonpath.Get("$.Accounts.DBInstanceAccount", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Accounts.DBInstanceAccount", response)
	}
	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("RDS", id), NotFoundWithResponse, response)
	}
	object = v.([]interface{})[0].(map[string]interface{})
	return object, nil
}

func (s *RdsService) GrantAccountPrivilege(id, dbName string) error {
	parts, err := ParseResourceId(id, 3)
	if err != nil {
		return WrapError(err)
	}
	action := "GrantAccountPrivilege"
	request := map[string]interface{}{
		"RegionId":         s.client.RegionId,
		"DBInstanceId":     parts[0],
		"AccountName":      parts[1],
		"DBName":           dbName,
		"AccountPrivilege": parts[2],
		"SourceIp":         s.client.SourceIp,
	}
	var response map[string]interface{}
	client := s.client
	err = resource.Retry(3*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("Rds", "2014-08-15", action, nil, request, false)
		if err != nil {
			if IsExpectedErrors(err, OperationDeniedDBStatus) || IsExpectedErrors(err, []string{"InvalidDB.NotFound"}) || NeedRetry(err) {
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
	if err := s.WaitForAccountPrivilege(id, dbName, Available, DefaultTimeoutMedium); err != nil {
		return WrapError(err)
	}

	return nil
}

func (s *RdsService) RevokeAccountPrivilege(id, dbName string) error {
	parts, err := ParseResourceId(id, 3)
	if err != nil {
		return WrapError(err)
	}
	action := "RevokeAccountPrivilege"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": parts[0],
		"AccountName":  parts[1],
		"DBName":       dbName,
		"SourceIp":     s.client.SourceIp,
	}
	client := s.client
	err = resource.Retry(3*time.Minute, func() *resource.RetryError {
		response, err := client.RpcPost("Rds", "2014-08-15", action, nil, request, false)
		if err != nil {
			if IsExpectedErrors(err, OperationDeniedDBStatus) || NeedRetry(err) {
				return resource.RetryableError(err)
			} else if IsExpectedErrors(err, []string{"InvalidDB.NotFound"}) {
				log.Printf("[WARN] Resource alicloud_db_account_privilege RevokeAccountPrivilege Failed!!! %s", err)
				return nil
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}

	if err := s.WaitForAccountPrivilegeRevoked(id, dbName, DefaultTimeoutMedium); err != nil {
		return WrapError(err)
	}

	return nil
}

func (s *RdsService) WaitForAccountPrivilege(id, dbName string, status Status, timeout int) error {
	parts, err := ParseResourceId(id, 3)
	if err != nil {
		return WrapError(err)
	}
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeDBDatabase(parts[0] + ":" + dbName)
		if err != nil {
			if IsNotFoundError(err) {
				if status == Deleted {
					return nil
				}
			} else {
				return WrapError(err)
			}
		}
		ready := false
		if object != nil {
			accountPrivilegeInfos := object["Accounts"].(map[string]interface{})["AccountPrivilegeInfo"].([]interface{})
			for _, account := range accountPrivilegeInfos {
				// At present, postgresql response has a bug, DBOwner will be changed to ALL
				// At present, sqlserver response has a bug, DBOwner will be changed to DbOwner
				account := account.(map[string]interface{})
				if account["Account"] == parts[1] && (account["AccountPrivilege"] == parts[2] || (parts[2] == "DBOwner" && (account["AccountPrivilege"] == "ALL" || account["AccountPrivilege"] == "DbOwner"))) {
					ready = true
					break
				}
			}
		}
		if status == Deleted && !ready {
			break
		}
		if ready {
			break
		}
		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), timeout, "", id, ProviderERROR)
		}
	}
	return nil
}

func (s *RdsService) WaitForAccountPrivilegeRevoked(id, dbName string, timeout int) error {
	parts, err := ParseResourceId(id, 3)
	if err != nil {
		return WrapError(err)
	}
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	for {
		object, err := s.DescribeDBDatabase(parts[0] + ":" + dbName)
		if err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			return WrapError(err)
		}

		exist := false
		if object != nil {
			accountPrivilegeInfo := object["Accounts"].(map[string]interface{})["AccountPrivilegeInfo"].([]interface{})
			for _, account := range accountPrivilegeInfo {
				account := account.(map[string]interface{})
				if account["Account"] == parts[1] && (account["AccountPrivilege"] == parts[2] || (parts[2] == "DBOwner" && account["AccountPrivilege"] == "ALL")) {
					exist = true
					break
				}
			}
		}

		if !exist {
			break
		}
		if time.Now().After(deadline) {
			return WrapErrorf(err, WaitTimeoutMsg, id, GetFunc(1), timeout, "", dbName, ProviderERROR)
		}

	}
	return nil
}

func (s *RdsService) RdsAccountStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeRdsAccount(id)
		if err != nil {
			if IsNotFoundError(err) {
				// Set this to nil as if we didn't find anything.
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		for _, failState := range failStates {
			if object["AccountStatus"].(string) == failState {
				return object, object["AccountStatus"].(string), WrapError(Error(FailedToReachTargetStatus, object["AccountStatus"].(string)))
			}
		}
		return object, object["AccountStatus"].(string), nil
	}
}
