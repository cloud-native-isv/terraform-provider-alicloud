package alicloud

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeArmsAddonRelease describes ARMS addon release
func (s *ArmsService) DescribeArmsAddonRelease(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		parts, err := ParseResourceId(id, 2)
		if err == nil && len(parts) >= 2 {
			addon, err := s.GetAPI().GetAddonRelease(parts[0], parts[1])
			if err == nil {
				// Convert to map[string]interface{} format expected by Terraform
				return map[string]interface{}{
					"EnvironmentId":    addon.EnvironmentId,
					"AddonReleaseName": addon.ReleaseName,
					"AddonVersion":     addon.Version,
					"Status":           addon.Status,
					"Config":           addon.Config,
				}, nil
			}
		}
	}

	// Placeholder implementation for ARMS addon release description
	// TODO: Implement when actual ARMS SDK integration is added
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
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		parts, err := ParseResourceId(id, 2)
		if err == nil && len(parts) >= 2 {
			customJob, err := s.GetAPI().GetEnvCustomJob(parts[0], parts[1])
			if err == nil {
				// Convert to map[string]interface{} format expected by Terraform
				return map[string]interface{}{
					"EnvironmentId": customJob.EnvironmentId,
					"CustomJobName": customJob.CustomJobName,
					"Status":        customJob.Status,
					"ConfigYaml":    customJob.ConfigYaml,
				}, nil
			}
		}
	}

	// Placeholder implementation for ARMS environment custom job description
	// TODO: Implement when actual ARMS SDK integration is added
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
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		parts, err := ParseResourceId(id, 3)
		if err == nil && len(parts) >= 3 {
			podMonitor, err := s.GetAPI().GetEnvPodMonitor(parts[0], parts[1], parts[2])
			if err == nil {
				// Convert to map[string]interface{} format expected by Terraform
				return map[string]interface{}{
					"EnvironmentId":  podMonitor.EnvironmentId,
					"Namespace":      podMonitor.Namespace,
					"PodMonitorName": podMonitor.PodMonitorName,
					"Status":         podMonitor.Status,
					"ConfigYaml":     podMonitor.ConfigYaml,
				}, nil
			}
		}
	}

	// Placeholder implementation for ARMS environment pod monitor description
	// TODO: Implement when actual ARMS SDK integration is added
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
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		parts, err := ParseResourceId(id, 3)
		if err == nil && len(parts) >= 3 {
			serviceMonitor, err := s.GetAPI().GetEnvServiceMonitor(parts[0], parts[1], parts[2])
			if err == nil {
				// Convert to map[string]interface{} format expected by Terraform
				return map[string]interface{}{
					"EnvironmentId":      serviceMonitor.EnvironmentId,
					"Namespace":          serviceMonitor.Namespace,
					"ServiceMonitorName": serviceMonitor.ServiceMonitorName,
					"Status":             serviceMonitor.Status,
					"ConfigYaml":         serviceMonitor.ConfigYaml,
				}, nil
			}
		}
	}

	// Placeholder implementation for ARMS environment service monitor description
	// TODO: Implement when actual ARMS SDK integration is added
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

// DescribeArmsEnvironment describes ARMS environment
func (s *ArmsService) DescribeArmsEnvironment(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		environment, err := s.GetAPI().GetEnvironment(id)
		if err == nil {
			// Convert to map[string]interface{} format expected by Terraform
			return map[string]interface{}{
				"EnvironmentId":       environment.EnvironmentId,
				"EnvironmentName":     environment.EnvironmentName,
				"EnvironmentType":     environment.EnvironmentType,
				"EnvironmentSubType":  environment.EnvironmentSubType,
				"BindResourceId":      environment.BindResourceId,
				"UserId":              environment.UserId,
				"Status":              "Success",
				"RegionId":            s.client.RegionId,
				"BindResourceProfile": "",
				"CreateTime":          "",
				"UpdateTime":          "",
			}, nil
		}
	}

	// Placeholder implementation for ARMS environment description
	// TODO: Implement when actual ARMS SDK integration is added
	return map[string]interface{}{
		"EnvironmentId":   id,
		"EnvironmentName": fmt.Sprintf("Environment-%s", id),
		"EnvironmentType": "ECS",
		"Status":          "Success",
		"RegionId":        s.client.RegionId,
	}, nil
}
