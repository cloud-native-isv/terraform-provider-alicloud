package alicloud

import (
	"fmt"
	"strings"
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
	var lastErr error
	var lastObserved interface{}
	for {
		object, err := s.DescribeDBDatabase(id)
		if err != nil {
			if NotFoundError(err) {
				if status == Deleted {
					return nil
				}
			}
			// Record the last error and return immediately as it's non-retryable here
			lastErr = err
			return WrapError(err)
		}
		if object != nil && object["DBName"] == parts[1] {
			break
		}
		// keep track of last observed object name (if any)
		if object != nil {
			lastObserved = object["DBName"]
		}
		time.Sleep(DefaultIntervalShort * time.Second)
		if time.Now().After(deadline) {
			// Avoid wrapping a nil cause; provide a meaningful timeout error instead
			cause := lastErr
			if cause == nil {
				cause = Error(FailedToReachTargetStatus, "timeout while waiting DB to reach target state")
			}
			return WrapErrorf(cause, WaitTimeoutMsg, id, GetFunc(1), timeout, lastObserved, parts[1], ProviderERROR)
		}
	}
	return nil
}

// EncodeDBId encodes instanceId and dbName to a single identifier: instanceId:dbName
func EncodeDBId(instanceId, dbName string) string {
	return fmt.Sprintf("%s:%s", instanceId, dbName)
}

// DecodeDBId decodes identifier to instanceId and dbName
func DecodeDBId(id string) (string, string, error) {
	parts := strings.Split(id, COLON_SEPARATED)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid DB id format, expected instanceId:dbName, got %s", id)
	}
	return parts[0], parts[1], nil
}

// CreateDBDatabase creates a database in the specified RDS instance.
// Note: This implementation uses underlying RpcPost for now to align with existing patterns.
func (s *RdsService) CreateDBDatabase(instanceId, name, characterSet, description string) error {
	action := "CreateDatabase"
	request := map[string]interface{}{
		"RegionId":         s.client.RegionId,
		"DBInstanceId":     instanceId,
		"DBName":           name,
		"CharacterSetName": characterSet,
		"SourceIp":         s.client.SourceIp,
	}
	if description != "" {
		request["DBDescription"] = description
	}

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err := s.client.RpcPost("Rds", "2014-08-15", action, nil, request, false)
		if err != nil {
			// Retry on common transient errors
			if IsExpectedErrors(err, []string{"ServiceUnavailable", "ThrottlingException", "InternalError", "Throttling", "SystemBusy", "OperationConflict"}) ||
				IsExpectedErrors(err, OperationDeniedDBStatus) || NeedRetry(err) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		addDebug(action, response, request)
		return nil
	})
}

// ModifyDBDatabaseDescription updates database description.
func (s *RdsService) ModifyDBDatabaseDescription(id, desc string) error {
	instanceId, dbName, err := func() (string, string, error) {
		i, n, e := DecodeDBId(id)
		if e != nil {
			// fallback to ParseResourceId used widely in repo
			parts, e2 := ParseResourceId(id, 2)
			if e2 != nil {
				return "", "", e
			}
			return parts[0], parts[1], nil
		}
		return i, n, nil
	}()
	if err != nil {
		return WrapError(err)
	}

	action := "ModifyDBDescription"
	request := map[string]interface{}{
		"RegionId":      s.client.RegionId,
		"DBInstanceId":  instanceId,
		"DBName":        dbName,
		"DBDescription": desc,
		"SourceIp":      s.client.SourceIp,
	}
	response, err := s.client.RpcPost("Rds", "2014-08-15", action, nil, request, false)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	addDebug(action, response, request)
	return nil
}

// DeleteDBDatabase deletes a database; caller should ensure instance is running if needed.
func (s *RdsService) DeleteDBDatabase(id string) error {
	instanceId, dbName, err := func() (string, string, error) {
		i, n, e := DecodeDBId(id)
		if e != nil {
			parts, e2 := ParseResourceId(id, 2)
			if e2 != nil {
				return "", "", e
			}
			return parts[0], parts[1], nil
		}
		return i, n, nil
	}()
	if err != nil {
		return WrapError(err)
	}
	action := "DeleteDatabase"
	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"DBInstanceId": instanceId,
		"DBName":       dbName,
		"SourceIp":     s.client.SourceIp,
	}
	response, err := s.client.RpcPost("Rds", "2014-08-15", action, nil, request, false)
	if err != nil {
		if NotFoundError(err) || IsExpectedErrors(err, []string{"InvalidDBName.NotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	addDebug(action, response, request)
	return nil
}

// DBDatabaseStateRefreshFunc returns a refresh func that checks if DB exists; used for waits.
func (s *RdsService) DBDatabaseStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		obj, err := s.DescribeDBDatabase(id)
		if err != nil {
			if NotFoundError(err) {
				// Not existing
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		// No specific status for DB; we use "Exists" as synthetic status
		for _, fail := range failStates {
			if fail == "FAILED" { // placeholder for extensibility
				return obj, fail, WrapError(Error(FailedToReachTargetStatus, fail))
			}
		}
		return obj, "Exists", nil
	}
}

// WaitForDBDatabaseCreating waits until DB is observable via Describe.
func (s *RdsService) WaitForDBDatabaseCreating(id string, timeout time.Duration) error {
	stateConf := BuildStateConf([]string{""}, []string{"Exists"}, timeout, 5*time.Second, s.DBDatabaseStateRefreshFunc(id, []string{}))
	_, err := stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}

// WaitForDBDatabaseDeleted waits until DB is not found.
func (s *RdsService) WaitForDBDatabaseDeleted(id string, timeout time.Duration) error {
	refresh := func() (interface{}, string, error) {
		_, err := s.DescribeDBDatabase(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "Exists", WrapError(err)
		}
		return nil, "Exists", nil
	}
	stateConf := BuildStateConf([]string{"Exists"}, []string{""}, timeout, 5*time.Second, refresh)
	_, err := stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}

// IsPermissionDenied determines whether the error indicates insufficient permission
func (s *RdsService) IsPermissionDenied(err error) bool {
	if err == nil {
		return false
	}
	// Check common Forbidden patterns
	if IsExpectedErrors(err, []string{"Forbidden", "Forbidden.", "Forbidden.ResourceNotFound", "Forbidden.InstanceNotFound", "Forbidden.Unauthorized"}) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "forbidden") || strings.Contains(msg, "access denied") || strings.Contains(msg, "unauthorized")
}
