package alicloud

import (
	"fmt"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeArmsGrafanaWorkspace describes ARMS Grafana workspace
func (s *ArmsService) DescribeArmsGrafanaWorkspace(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "GetGrafanaWorkspace"
	request := map[string]interface{}{
		"GrafanaWorkspaceId": id,
	}
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		response, err = client.RpcPost("ARMS", "2019-08-08", action, nil, request, true)
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
	v, err := jsonpath.Get("$.Data", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Data", response)
	}
	object = v.(map[string]interface{})
	return object, nil
}

// ArmsGrafanaWorkspaceStateRefreshFunc returns state refresh function for ARMS Grafana workspace
func (s *ArmsService) ArmsGrafanaWorkspaceStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeArmsGrafanaWorkspace(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if fmt.Sprint(object["Status"]) == failState {
				return object, fmt.Sprint(object["Status"]), WrapError(Error(FailedToReachTargetStatus, fmt.Sprint(object["Status"])))
			}
		}

		return object, fmt.Sprint(object["Status"]), nil
	}
}

// WaitForArmsGrafanaWorkspaceCreated waits for ARMS Grafana workspace to be created
func (s *ArmsService) WaitForArmsGrafanaWorkspaceCreated(id string, timeout time.Duration) error {
	stateConf := BuildStateConf([]string{"CREATING"}, []string{"RUNNING"}, timeout, 5*time.Second, s.ArmsGrafanaWorkspaceStateRefreshFunc(id, []string{"CREATE_FAILED"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}

// WaitForArmsGrafanaWorkspaceDeleted waits for ARMS Grafana workspace to be deleted
func (s *ArmsService) WaitForArmsGrafanaWorkspaceDeleted(id string, timeout time.Duration) error {
	stateConf := BuildStateConf([]string{"RUNNING", "STOPPING"}, []string{}, timeout, 5*time.Second, s.ArmsGrafanaWorkspaceStateRefreshFunc(id, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}
