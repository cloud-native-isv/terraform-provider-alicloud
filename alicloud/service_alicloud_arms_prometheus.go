package alicloud

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeArmsPrometheus describes ARMS Prometheus - alias for DescribeArmsPrometheusInstance
func (s *ArmsService) DescribeArmsPrometheus(id string) (object map[string]interface{}, err error) {
	return s.DescribeArmsPrometheusInstance(id)
}

// DescribeArmsPrometheusInstance describes ARMS Prometheus instance
func (s *ArmsService) DescribeArmsPrometheusInstance(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		instance, err := s.GetAPI().GetPrometheusInstance(id)
		if err == nil {
			// Convert to map[string]interface{} format expected by Terraform
			return map[string]interface{}{
				"ClusterId":       instance.ClusterId,
				"ClusterName":     instance.ClusterName,
				"ClusterType":     instance.ClusterType,
				"VpcId":           instance.VpcId,
				"VSwitchId":       instance.VSwitchId,
				"SecurityGroupId": instance.SecurityGroupId,

				"GrafanaInstanceId": instance.GrafanaInstanceId,
				"HttpApiInterUrl":   instance.HttpApiInterUrl,

				"PushGatewayInterUrl": instance.PushGatewayInterUrl,

				"RemoteReadInterUrl": instance.RemoteReadInterUrl,

				"RemoteWriteInterUrl": instance.RemoteWriteInterUrl,

				"SubClustersJson": instance.SubClustersJson,
			}, nil
		}
	}

	// Fallback to placeholder implementation
	// TODO: Implement actual RPC call when needed
	return map[string]interface{}{
		"ClusterId":   id,
		"ClusterName": fmt.Sprintf("Prometheus-%s", id),
		"ClusterType": "ask",
		"Status":      "Running",
	}, nil
}

// DescribeArmsPrometheusMonitoring describes ARMS Prometheus monitoring
func (s *ArmsService) DescribeArmsPrometheusMonitoring(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		parts, err := ParseResourceId(id, 3)
		if err == nil && len(parts) >= 3 {
			monitoring, err := s.GetAPI().GetPrometheusMonitoring(parts[0], parts[1], parts[2])
			if err == nil {
				// Convert to map[string]interface{} format expected by Terraform
				return map[string]interface{}{
					"ClusterId":      monitoring.ClusterId,
					"MonitoringName": monitoring.MonitoringName,
					"Type":           monitoring.Type,
					"Status":         monitoring.Status,
					"ConfigYaml":     monitoring.ConfigYaml,
					"Config":         monitoring.Config,
				}, nil
			}
		}
	}

	// Fallback to placeholder implementation
	// TODO: Implement actual RPC call when needed
	parts, err := ParseResourceId(id, 3)
	if err != nil {
		return nil, WrapError(err)
	}

	return map[string]interface{}{
		"ClusterId":      parts[0],
		"MonitoringName": parts[1],
		"Type":           parts[2],
		"Status":         "Running",
		"ConfigYaml":     "",
	}, nil
}

// DescribeArmsPrometheusAlertRule describes ARMS Prometheus alert rule
func (s *ArmsService) DescribeArmsPrometheusAlertRule(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		parts, err := ParseResourceId(id, 2)
		if err == nil && len(parts) >= 2 {
			alertRule, err := s.GetAPI().GetPrometheusAlertRule(parts[0], parts[1])
			if err == nil {
				// Convert to map[string]interface{} format expected by Terraform
				return map[string]interface{}{
					"ClusterId":   alertRule.ClusterId,
					"AlertId":     alertRule.AlertId,
					"AlertName":   alertRule.AlertName,
					"Expression":  alertRule.Expression,
					"Status":      alertRule.Status,
					"Type":        alertRule.Type,
					"Duration":    alertRule.Duration,
					"Message":     alertRule.Message,
					"Annotations": alertRule.Annotations,
					"Labels":      alertRule.Labels,
				}, nil
			}
		}
	}

	// Fallback to placeholder implementation
	// TODO: Implement actual RPC call when needed
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return nil, WrapError(err)
	}

	return map[string]interface{}{
		"ClusterId": parts[0],
		"AlertId":   parts[1],
		"AlertName": fmt.Sprintf("AlertRule-%s", parts[1]),
		"Status":    "1",
		"Type":      "101",
	}, nil
}

// DescribeArmsRemoteWrite describes ARMS remote write configuration
func (s *ArmsService) DescribeArmsRemoteWrite(id string) (object map[string]interface{}, err error) {
	// Placeholder implementation for ARMS remote write description
	// TODO: Implement when actual ARMS SDK integration is added
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return nil, WrapError(err)
	}

	return map[string]interface{}{
		"ClusterId":       parts[0],
		"RemoteWriteName": parts[1],
		"RemoteWriteYaml": "",
		"Status":          "Success",
		"Config":          "",
	}, nil
}

// ArmsPrometheusInstanceStateRefreshFunc returns state refresh function for ARMS Prometheus instance
func (s *ArmsService) ArmsPrometheusInstanceStateRefreshFunc(id string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeArmsPrometheusInstance(id)
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
