package alicloud

import (
	"fmt"

	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// SafeStringValue safely dereferences a string pointer, returning empty string if nil
func SafeStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// DescribeArmsAddonRelease describes ARMS addon release
func (s *ArmsService) DescribeArmsAddonRelease(id string) (object map[string]interface{}, err error) {
	// Using placeholder implementation since the specific addon release API is not yet implemented
	// TODO: Implement when actual ARMS addon release SDK integration is added
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return nil, WrapError(err)
	}

	return map[string]interface{}{
		"EnvironmentId":    parts[0],
		"AddonReleaseName": parts[1],
		"Status":           "Success",
		"AddonVersion":     "1.0.0",
	}, nil
}

// ArmsAddonReleaseStateRefreshFunc returns state refresh function for ARMS addon release
func (s *ArmsService) ArmsAddonReleaseStateRefreshFunc(id string, jsonPath string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeArmsAddonRelease(id)
		if err != nil {
			if IsNotFoundError(err) {
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

// DescribeArmsEnvCustomJob describes ARMS environment custom job
func (s *ArmsService) DescribeArmsEnvCustomJob(id string) (object map[string]interface{}, err error) {
	// Using placeholder implementation since the specific custom job API is not yet implemented
	// TODO: Implement when actual ARMS custom job SDK integration is added
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return nil, WrapError(err)
	}

	return map[string]interface{}{
		"EnvironmentId": parts[0],
		"CustomJobName": parts[1],
		"Status":        "Running",
		"ConfigYaml":    "",
	}, nil
}

// DescribeArmsEnvFeature describes ARMS environment feature
func (s *ArmsService) DescribeArmsEnvFeature(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	// if s.armsAPI != nil {
	// 	parts, err := ParseResourceId(id, 2)
	// 	if err == nil && len(parts) >= 2 {
	// 		feature, err := s.GetAPI().GetEnvFeature(parts[0], parts[1])
	// 		if err == nil {
	// 			// Convert to map[string]interface{} format expected by Terraform
	// 			return map[string]interface{}{
	// 				"EnvironmentId": feature.EnvironmentId,
	// 				"FeatureName":   feature.FeatureName,
	// 				"Status":        feature.Status,
	// 				"Config":        feature.Config,
	// 				"Namespace":     feature.Namespace,
	// 			}, nil
	// 		}
	// 	}
	// }

	// Placeholder implementation for ARMS environment feature description
	// TODO: Implement when actual ARMS SDK integration is added
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return nil, WrapError(err)
	}

	return map[string]interface{}{
		"EnvironmentId": parts[0],
		"FeatureName":   parts[1],
		"Status":        "Success",
		"Config":        "",
	}, nil
}

// ArmsEnvFeatureStateRefreshFunc returns state refresh function for ARMS environment feature
func (s *ArmsService) ArmsEnvFeatureStateRefreshFunc(id string, jsonPath string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeArmsEnvFeature(id)
		if err != nil {
			if IsNotFoundError(err) {
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

// DescribeArmsEnvPodMonitor describes ARMS environment pod monitor
func (s *ArmsService) DescribeArmsEnvPodMonitor(id string) (object map[string]interface{}, err error) {
	// Using placeholder implementation since the specific pod monitor API is not yet implemented
	// TODO: Implement when actual ARMS pod monitor SDK integration is added
	parts, err := ParseResourceId(id, 3)
	if err != nil {
		return nil, WrapError(err)
	}

	return map[string]interface{}{
		"EnvironmentId":  parts[0],
		"Namespace":      parts[1],
		"PodMonitorName": parts[2],
		"Status":         "Success",
		"ConfigYaml":     "",
	}, nil
}

// DescribeArmsEnvServiceMonitor describes ARMS environment service monitor
func (s *ArmsService) DescribeArmsEnvServiceMonitor(id string) (object map[string]interface{}, err error) {
	// Using placeholder implementation since the specific service monitor API is not yet implemented
	// TODO: Implement when actual ARMS service monitor SDK integration is added
	parts, err := ParseResourceId(id, 3)
	if err != nil {
		return nil, WrapError(err)
	}

	return map[string]interface{}{
		"EnvironmentId":      parts[0],
		"Namespace":          parts[1],
		"ServiceMonitorName": parts[2],
		"Status":             "Success",
		"ConfigYaml":         "",
	}, nil
}

// CreateArmsPrometheusRemoteWrite creates an ARMS Prometheus remote write configuration
func (s *ArmsService) CreateArmsPrometheusRemoteWrite(clusterId, remoteWriteYaml string) (*aliyunArmsAPI.PrometheusRemoteWrite, error) {
	if s.armsAPI != nil {
		// Call the API to create remote write
		result, err := s.armsAPI.CreatePrometheusRemoteWrite(clusterId, remoteWriteYaml)
		if err != nil {
			return nil, WrapErrorf(err, DefaultErrorMsg, "ArmsPrometheusRemoteWrite", "CreatePrometheusRemoteWrite", AlibabaCloudSdkGoERROR)
		}
		return result, nil
	}

	// Fallback to placeholder
	return nil, WrapError(Error("ARMS API not initialized"))
}

// UpdateArmsPrometheusRemoteWrite updates an ARMS Prometheus remote write configuration
func (s *ArmsService) UpdateArmsPrometheusRemoteWrite(clusterId, remoteWriteName, remoteWriteYaml string) (*aliyunArmsAPI.PrometheusRemoteWrite, error) {
	if s.armsAPI != nil {
		// Call the API to update remote write
		result, err := s.armsAPI.UpdatePrometheusRemoteWrite(clusterId, remoteWriteName, remoteWriteYaml)
		if err != nil {
			return nil, WrapErrorf(err, DefaultErrorMsg, fmt.Sprintf("%s:%s", clusterId, remoteWriteName), "UpdatePrometheusRemoteWrite", AlibabaCloudSdkGoERROR)
		}
		return result, nil
	}

	// Fallback to placeholder
	return nil, WrapError(Error("ARMS API not initialized"))
}

// DeleteArmsPrometheusRemoteWrite deletes an ARMS Prometheus remote write configuration
func (s *ArmsService) DeleteArmsPrometheusRemoteWrite(clusterId string, remoteWriteNames []string) error {
	if s.armsAPI != nil {
		// Call the API to delete remote write
		err := s.armsAPI.DeletePrometheusRemoteWrite(clusterId, remoteWriteNames)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, fmt.Sprintf("%s:%v", clusterId, remoteWriteNames), "DeletePrometheusRemoteWrite", AlibabaCloudSdkGoERROR)
		}
		return nil
	}

	// Fallback to placeholder
	return WrapError(Error("ARMS API not initialized"))
}

// DescribeArmsRemoteWrite describes ARMS Prometheus remote write configuration
func (s *ArmsService) DescribeArmsRemoteWrite(id string) (object map[string]interface{}, err error) {
	if s.armsAPI != nil {
		// Call the API to get remote write
		remoteWrite, err := s.armsAPI.GetPrometheusRemoteWrite(id)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
			}
			return nil, WrapErrorf(err, DefaultErrorMsg, id, "GetPrometheusRemoteWrite", AlibabaCloudSdkGoERROR)
		}

		// Convert to map[string]interface{} format expected by Terraform
		result := map[string]interface{}{
			"ClusterId":       SafeStringValue(remoteWrite.ClusterId),
			"RemoteWriteName": SafeStringValue(remoteWrite.RemoteWriteName),
			"RemoteWriteYaml": SafeStringValue(remoteWrite.RemoteWriteYaml),
			"Status":          SafeStringValue(remoteWrite.Status),
		}

		return result, nil
	}

	// Fallback placeholder implementation
	return nil, WrapErrorf(Error(GetNotFoundMessage("ArmsRemoteWrite", id)), NotFoundWithResponse, "DescribeArmsRemoteWrite")
}

func (s *ArmsService) DescribeArmsEnvironment(id string) (object map[string]interface{}, err error) {
	// Use the new Environment API
	if s.armsAPI != nil {
		environment, err := s.armsAPI.DescribeEnvironment(id)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
			}
			return nil, WrapErrorf(err, DefaultErrorMsg, id, "DescribeEnvironment", AlibabaCloudSdkGoERROR)
		}

		// Convert to map[string]interface{} format expected by Terraform
		result := map[string]interface{}{
			"EnvironmentId":      SafeStringValue(environment.EnvironmentId),
			"EnvironmentName":    SafeStringValue(environment.EnvironmentName),
			"EnvironmentType":    SafeStringValue(environment.EnvironmentType),
			"EnvironmentSubType": SafeStringValue(environment.EnvironmentSubType),
			"BindResourceId":     SafeStringValue(environment.BindResourceId),
			"UserId":             SafeStringValue(environment.UserId),
			"RegionId":           SafeStringValue(environment.RegionId),
			"Status":             "Success", // Default status for compatibility
		}

		// Set optional fields if they exist
		if environment.BindResourceType != nil {
			result["BindResourceType"] = *environment.BindResourceType
		}
		if environment.BindResourceStatus != nil {
			result["BindResourceStatus"] = *environment.BindResourceStatus
		}
		if environment.BindResourceProfile != nil {
			result["BindResourceProfile"] = *environment.BindResourceProfile
		}
		if environment.CreateTime != nil {
			result["CreateTime"] = *environment.CreateTime
		}
		if environment.PrometheusInstanceId != nil {
			result["PrometheusInstanceId"] = *environment.PrometheusInstanceId
		}
		if environment.GrafanaWorkspaceId != nil {
			result["GrafanaWorkspaceId"] = *environment.GrafanaWorkspaceId
		}
		if environment.ManagedType != nil {
			result["ManagedType"] = *environment.ManagedType
		}
		if environment.FeePackage != nil {
			result["FeePackage"] = *environment.FeePackage
		}
		if environment.ResourceGroupId != nil {
			result["ResourceGroupId"] = *environment.ResourceGroupId
		}

		return result, nil
	}

	// Fallback placeholder implementation
	// TODO: Remove this when all environments use the new API
	return map[string]interface{}{
		"EnvironmentId":   id,
		"EnvironmentName": fmt.Sprintf("Environment-%s", id),
		"EnvironmentType": "ECS",
		"Status":          "Success",
		"RegionId":        s.client.RegionId,
	}, nil
}
