package alicloud

import (
	"fmt"
	"strconv"
	"time"

	"github.com/PaesslerAG/jsonpath"
	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// CreateArmsIntegration creates a new ARMS alert integration
func (s *ArmsService) CreateArmsIntegration(integration *aliyunArmsAPI.AlertIntegration) (*aliyunArmsAPI.AlertIntegration, error) {
	if s.armsAPI == nil {
		return nil, fmt.Errorf("ARMS API not available")
	}

	result, err := s.GetAPI().CreateIntegration(integration)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// UpdateArmsIntegration updates an existing ARMS alert integration
func (s *ArmsService) UpdateArmsIntegration(integration *aliyunArmsAPI.AlertIntegration) (*aliyunArmsAPI.AlertIntegration, error) {
	if s.armsAPI == nil {
		return nil, fmt.Errorf("ARMS API not available")
	}

	result, err := s.GetAPI().UpdateIntegration(integration)
	if err != nil {
		return nil, WrapError(err)
	}

	return result, nil
}

// DeleteArmsIntegration deletes an ARMS alert integration
func (s *ArmsService) DeleteArmsIntegration(integrationId int64) error {
	if s.armsAPI == nil {
		return fmt.Errorf("ARMS API not available")
	}

	err := s.GetAPI().DeleteIntegration(integrationId)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

func (s *ArmsService) DescribeArmsIntegration(id string) (*aliyunArmsAPI.AlertIntegration, error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		// Convert string ID to int64
		integrationId, parseErr := strconv.ParseInt(id, 10, 64)
		if parseErr == nil {
			integration, err := s.GetAPI().GetIntegrationById(integrationId)
			if err == nil {
				return integration, nil
			}
		}
	}

	// Fallback to direct RPC call - not recommended, return error to encourage API usage
	return nil, fmt.Errorf("ARMS API not available or failed, integration ID: %s", id)
}

func (s *ArmsService) ListArmsAlertIntegrations() ([]*aliyunArmsAPI.AlertIntegration, error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		integrations, err := s.GetAPI().ListAllIntegrations()
		if err == nil {
			return integrations, nil
		}
	}

	// Fallback to direct RPC call - not recommended, return error to encourage API usage
	return nil, fmt.Errorf("ARMS API not available or failed")
}

// DescribeArmsIntegrationExporter describes ARMS integration exporter
func (s *ArmsService) DescribeArmsIntegrationExporter(id string) (object map[string]interface{}, err error) {
	var response map[string]interface{}
	client := s.client
	action := "GetPrometheusIntegration"
	request := map[string]interface{}{
		"InstanceId":      id,
		"IntegrationType": "kafka",
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

// ArmsIntegrationStateRefreshFunc returns state refresh function for ARMS integration
func (s *ArmsService) ArmsIntegrationStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeArmsIntegration(id)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		// Extract current status based on Liveness and State fields
		// Only when Liveness="ready" AND State=true, the integration is considered READY
		var currentStatus string
		if object.Liveness == "ready" && object.State {
			currentStatus = aliyunArmsAPI.AlertIntegrationStatusReady
		} else {
			// All other combinations are considered as CREATING (not ready yet)
			currentStatus = aliyunArmsAPI.AlertIntegrationStatusCreating
		}

		// Check for fail states
		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}

		return object, currentStatus, nil
	}
}

// WaitForArmsIntegrationCreated waits for ARMS integration to be created
func (s *ArmsService) WaitForArmsIntegrationCreated(id string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{aliyunArmsAPI.AlertIntegrationStatusCreating}, // pending states
		[]string{aliyunArmsAPI.AlertIntegrationStatusReady},    // target states
		timeout,
		5*time.Second,
		s.ArmsIntegrationStateRefreshFunc(id, []string{aliyunArmsAPI.AlertIntegrationStatusFailed}),
	)
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}

// WaitForArmsIntegrationDeleted waits for ARMS integration to be deleted
func (s *ArmsService) WaitForArmsIntegrationDeleted(id string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{aliyunArmsAPI.AlertIntegrationStatusReady, aliyunArmsAPI.AlertIntegrationStatusFailed, aliyunArmsAPI.AlertIntegrationStatusCreating}, // pending states
		[]string{}, // target states (empty = wait for resource disappear)
		timeout,
		5*time.Second,
		s.ArmsIntegrationStateRefreshFunc(id, []string{}),
	)
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}
	return nil
}
