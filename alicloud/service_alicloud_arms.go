package alicloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type ArmsService struct {
	client    *connectivity.AliyunClient
	aliyunAPI *aliyunArmsAPI.ArmsAPI
}

// NewArmsService creates a new ArmsService instance
func NewArmsService(client *connectivity.AliyunClient) *ArmsService {
	// Initialize the ARMS API client
	armsAPI, err := aliyunArmsAPI.NewARMSClientWithCredentials(&aliyunArmsAPI.ArmsCredentials{
		AccessKey:     client.AccessKey,
		SecretKey:     client.SecretKey,
		RegionId:      client.RegionId,
		SecurityToken: client.SecurityToken,
	})
	if err != nil {
		// Log error but continue with nil API (will fall back to direct RPC calls)
		addDebug("NewArmsService", "Failed to initialize ARMS API client", err)
	}

	return &ArmsService{
		client:        client,
		aliyunArmsAPI: armsAPI,
	}
}

func (s *ArmsService) DescribeArmsAlertContact(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.aliyunArmsAPI != nil {
		contacts, err := s.aliyunArmsAPI.SearchAlertContactByIds(context.Background(), []string{id})
		if err == nil && len(contacts) > 0 {
			// Convert to map[string]interface{} format expected by Terraform
			return map[string]interface{}{
				"ContactId":    contacts[0].ContactId,
				"ContactName":  contacts[0].ContactName,
				"Phone":        contacts[0].Phone,
				"Email":        contacts[0].Email,
				"DingRobotUrl": contacts[0].DingRobotUrl,
				"SystemNoc":    contacts[0].SystemNoc,
			}, nil
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	client := s.client
	action := "SearchAlertContact"
	request := map[string]interface{}{
		"RegionId":   s.client.RegionId,
		"ContactIds": convertListToJsonString([]interface{}{id}),
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
	v, err := jsonpath.Get("$.PageBean.Contacts", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.PageBean.Contacts", response)
	}
	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	} else {
		if fmt.Sprint(v.([]interface{})[0].(map[string]interface{})["ContactId"]) != id {
			return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
		}
	}
	object = v.([]interface{})[0].(map[string]interface{})
	return object, nil
}

func (s *ArmsService) DescribeArmsAlertContactGroup(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.aliyunArmsAPI != nil {
		groups, err := s.aliyunArmsAPI.SearchAlertContactGroupByIds(context.Background(), []string{id}, true)
		if err == nil && len(groups) > 0 {
			// Convert to map[string]interface{} format expected by Terraform
			return map[string]interface{}{
				"ContactGroupId":   groups[0].ContactGroupId,
				"ContactGroupName": groups[0].ContactGroupName,
				"ContactIds":       groups[0].ContactIds,
				"Contacts":         groups[0].Contacts,
			}, nil
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	client := s.client
	action := "SearchAlertContactGroup"
	request := map[string]interface{}{
		"RegionId":        s.client.RegionId,
		"ContactGroupIds": convertListToJsonString([]interface{}{id}),
		"IsDetail":        "true",
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
	v, err := jsonpath.Get("$.ContactGroups", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.ContactGroups", response)
	}
	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	} else {
		if fmt.Sprint(v.([]interface{})[0].(map[string]interface{})["ContactGroupId"]) != id {
			return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
		}
	}
	object = v.([]interface{})[0].(map[string]interface{})
	return object, nil
}

func (s *ArmsService) DescribeArmsAlertRobot(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.aliyunArmsAPI != nil {
		robots, err := s.aliyunArmsAPI.DescribeIMRobotsByIds(context.Background(), []string{id}, 1, PageSizeXLarge)
		if err == nil && len(robots) > 0 {
			// Convert to map[string]interface{} format expected by Terraform
			return map[string]interface{}{
				"RobotId":   robots[0].RobotId,
				"RobotName": robots[0].RobotName,
				"RobotAddr": robots[0].RobotAddr,
				"Type":      robots[0].Type,
				"Token":     robots[0].Token,
			}, nil
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	client := s.client
	action := "DescribeIMRobots"
	request := map[string]interface{}{
		"RobotIds": id,
		"Page":     1,
		"Size":     PageSizeXLarge,
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
	v, err := jsonpath.Get("$.PageBean.AlertIMRobots", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.PageBean.AlertIMRobots", response)
	}
	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	} else {
		if fmt.Sprint(v.([]interface{})[0].(map[string]interface{})["RobotId"]) != id {
			return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
		}
	}
	object = v.([]interface{})[0].(map[string]interface{})
	return object, nil
}

func (s *ArmsService) DescribeArmsDispatchRule(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.aliyunArmsAPI != nil {
		rule, err := s.aliyunArmsAPI.DescribeDispatchRule(context.Background(), id)
		if err == nil {
			// Convert to map[string]interface{} format expected by Terraform
			return map[string]interface{}{
				"RuleId":                   rule.RuleId,
				"Name":                     rule.Name,
				"State":                    rule.State,
				"GroupRules":               rule.GroupRules,
				"LabelMatchExpressionGrid": rule.LabelMatchExpressionGrid,
				"NotifyRules":              rule.NotifyRules,
			}, nil
		}
	}

	// Fallback to direct RPC call
	client := s.client

	request := map[string]interface{}{
		"Id":       id,
		"RegionId": s.client.RegionId,
	}

	var response map[string]interface{}
	action := "DescribeDispatchRule"
	wait := incrementalWait(3*time.Second, 3*time.Second)
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		resp, err := client.RpcPost("ARMS", "2019-08-08", action, nil, request, false)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		response = resp
		addDebug(action, resp, request)
		return nil
	})
	if err != nil {
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	v, err := jsonpath.Get("$.DispatchRule", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.DispatchRule", response)
	}
	return v.(map[string]interface{}), nil
}

func (s *ArmsService) DescribeArmsPrometheusAlertRule(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.aliyunArmsAPI != nil {
		parts, err := ParseResourceId(id, 2)
		if err == nil {
			rule, err := s.aliyunArmsAPI.GetPrometheusAlertRule(context.Background(), parts[0], parts[1])
			if err == nil {
				// Convert to map[string]interface{} format expected by Terraform
				return map[string]interface{}{
					"AlertId":        rule.AlertId,
					"AlertName":      rule.AlertName,
					"ClusterId":      rule.ClusterId,
					"Expression":     rule.Expression,
					"Message":        rule.Message,
					"Duration":       rule.Duration,
					"NotifyType":     rule.NotifyType,
					"Labels":         rule.Labels,
					"Annotations":    rule.Annotations,
					"DispatchRuleId": rule.DispatchRuleId,
					"Status":         rule.Status,
					"Type":           rule.Type,
				}, nil
			}
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	client := s.client
	action := "ListPrometheusAlertRules"
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		err = WrapError(err)
		return
	}
	request := map[string]interface{}{
		"RegionId":  s.client.RegionId,
		"ClusterId": parts[0],
	}
	idExist := false
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
	v, err := jsonpath.Get("$.PrometheusAlertRules", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.PrometheusAlertRules", response)
	}
	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	}
	for _, v := range v.([]interface{}) {
		if fmt.Sprint(v.(map[string]interface{})["AlertId"]) == parts[1] {
			idExist = true
			return v.(map[string]interface{}), nil
		}
	}
	if !idExist {
		return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	}
	return object, nil
}

func (s *ArmsService) ListArmsNotificationPolicies(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.aliyunArmsAPI != nil {
		policy, err := s.aliyunArmsAPI.GetNotificationPolicy(context.Background(), id)
		if err == nil {
			// Convert to map[string]interface{} format expected by Terraform
			return map[string]interface{}{
				"Id":                  policy.Id,
				"Name":                policy.Name,
				"SendRecoverMessage":  policy.SendRecoverMessage,
				"RepeatInterval":      policy.RepeatInterval,
				"EscalationPolicyId":  policy.EscalationPolicyId,
				"GroupRule":           policy.GroupRule,
				"MatchingRules":       policy.MatchingRules,
				"NotifyRule":          policy.NotifyRule,
				"IntegrationExporter": policy.IntegrationExporter,
			}, nil
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	client := s.client
	action := "ListNotificationPolicies"
	request := map[string]interface{}{
		"Page":     1,
		"Size":     PageSizeXLarge,
		"IsDetail": true,
		"RegionId": s.client.RegionId,
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
	v, err := jsonpath.Get("$.PageBean.NotificationPolicies", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.PageBean.NotificationPolicies", response)
	}
	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	}
	for _, v := range v.([]interface{}) {
		if fmt.Sprint(v.(map[string]interface{})["Id"]) == id {
			return v.(map[string]interface{}), nil
		}
	}
	return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
}

func (s *ArmsService) ArmsDispatchRuleStateRefreshFunc(d *schema.ResourceData, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeArmsDispatchRule(d.Id())
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}
		for _, failState := range failStates {
			if fmt.Sprint(object["State"]) == failState {
				return object, fmt.Sprint(object["State"]), WrapError(Error(FailedToReachTargetStatus, fmt.Sprint(object["State"])))
			}
		}
		return object, fmt.Sprint(object["State"]), nil
	}
}

func (s *ArmsService) DescribeArmsPrometheus(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.aliyunArmsAPI != nil {
		instance, err := s.aliyunArmsAPI.GetPrometheusInstance(context.Background(), id)
		if err == nil {
			// Convert to map[string]interface{} format expected by Terraform
			return map[string]interface{}{
				"ClusterId":           instance.ClusterId,
				"ClusterName":         instance.ClusterName,
				"ClusterType":         instance.ClusterType,
				"VpcId":               instance.VpcId,
				"VSwitchId":           instance.VSwitchId,
				"SecurityGroupId":     instance.SecurityGroupId,
				"SubClustersJson":     instance.SubClustersJson,
				"HttpApiInterUrl":     instance.HttpApiInterUrl,
				"HttpApiIntraUrl":     instance.HttpApiIntraUrl,
				"PushGatewayInterUrl": instance.PushGatewayInterUrl,
				"PushGatewayIntraUrl": instance.PushGatewayIntraUrl,
				"RemoteReadInterUrl":  instance.RemoteReadInterUrl,
				"RemoteReadIntraUrl":  instance.RemoteReadIntraUrl,
				"RemoteWriteInterUrl": instance.RemoteWriteInterUrl,
				"RemoteWriteIntraUrl": instance.RemoteWriteIntraUrl,
				"AuthToken":           instance.AuthToken,
				"PaymentType":         instance.PaymentType,
				"GrafanaInstanceId":   instance.GrafanaInstanceId,
				"ResourceGroupId":     instance.ResourceGroupId,
				"ResourceType":        instance.ResourceType,
				"StorageDuration":     instance.StorageDuration,
				"Tags":                instance.Tags,
			}, nil
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	action := "GetPrometheusInstance"

	client := s.client

	request := map[string]interface{}{
		"RegionId":  s.client.RegionId,
		"ClusterId": id,
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
		if IsExpectedErrors(err, []string{"404"}) {
			return object, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}

	v, err := jsonpath.Get("$.Data", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Data", response)
	}

	object = v.(map[string]interface{})

	return object, nil
}

func (s *ArmsService) ListTagResources(id string, resourceType string) (object interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.aliyunArmsAPI != nil {
		tags, _, err := s.aliyunArmsAPI.ListTagResources(context.Background(), resourceType, id, "")
		if err == nil {
			// Convert to the format expected by Terraform
			var result []interface{}
			for _, tag := range tags {
				result = append(result, map[string]interface{}{
					"TagKey":       tag.TagKey,
					"TagValue":     tag.TagValue,
					"ResourceId":   tag.ResourceId,
					"ResourceType": tag.ResourceType,
				})
			}
			return result, nil
		}
	}

	// Fallback to direct RPC call
	client := s.client
	action := "ListTagResources"

	request := map[string]interface{}{
		"RegionId":     s.client.RegionId,
		"ResourceType": resourceType,
	}

	resourceIdNum := strings.Count(id, ":")

	switch resourceIdNum {
	case 0:
		request["ResourceId.1"] = id
	case 1:
		parts, err := ParseResourceId(id, 2)
		if err != nil {
			return object, WrapError(err)
		}
		request["ResourceId.1"] = parts[resourceIdNum]
	}

	tags := make([]interface{}, 0)
	var response map[string]interface{}
	for {
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
			addDebug(action, response, request)
			v, err := jsonpath.Get("$.TagResources", response)
			if err != nil {
				return resource.NonRetryableError(WrapErrorf(err, FailedGetAttributeMsg, id, "$.TagResources", response))
			}

			if v != nil {
				tags = append(tags, v.([]interface{})...)
			}

			return nil
		})
		if err != nil {
			err = WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
			return
		}
		if response["NextToken"] == nil {
			break
		}
		request["NextToken"] = response["NextToken"]
	}

	return tags, nil
}

func (s *ArmsService) SetResourceTags(d *schema.ResourceData, resourceType string) error {
	// Try using aliyunArmsAPI first if available
	if s.aliyunArmsAPI != nil && d.HasChange("tags") {
		added, removed := parsingTags(d)

		resourceIdNum := strings.Count(d.Id(), ":")
		var resourceId string

		switch resourceIdNum {
		case 0:
			resourceId = d.Id()
		case 1:
			parts, err := ParseResourceId(d.Id(), 2)
			if err != nil {
				return WrapError(err)
			}
			resourceId = parts[resourceIdNum]
		}

		// Remove tags
		removedTagKeys := make([]string, 0)
		for _, v := range removed {
			if !ignoredTags(v, "") {
				removedTagKeys = append(removedTagKeys, v)
			}
		}

		if len(removedTagKeys) > 0 {
			err := s.aliyunArmsAPI.UntagResources(context.Background(), resourceType, []string{resourceId}, removedTagKeys)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UntagResources", AlibabaCloudSdkGoERROR)
			}
		}

		// Add tags
		if len(added) > 0 {
			err := s.aliyunArmsAPI.TagResources(context.Background(), resourceType, []string{resourceId}, added)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "TagResources", AlibabaCloudSdkGoERROR)
			}
		}

		d.SetPartial("tags")
		return nil
	}

	// Fallback to direct RPC call
	resourceIdNum := strings.Count(d.Id(), ":")

	if d.HasChange("tags") {
		added, removed := parsingTags(d)
		client := s.client

		removedTagKeys := make([]string, 0)
		for _, v := range removed {
			if !ignoredTags(v, "") {
				removedTagKeys = append(removedTagKeys, v)
			}
		}

		if len(removedTagKeys) > 0 {
			action := "UntagResources"
			request := map[string]interface{}{
				"RegionId":     s.client.RegionId,
				"ResourceType": resourceType,
			}

			switch resourceIdNum {
			case 0:
				request["ResourceId.1"] = d.Id()
			case 1:
				parts, err := ParseResourceId(d.Id(), 2)
				if err != nil {
					return WrapError(err)
				}
				request["ResourceId.1"] = parts[resourceIdNum]
			}

			for i, key := range removedTagKeys {
				request[fmt.Sprintf("TagKey.%d", i+1)] = key
			}
			wait := incrementalWait(2*time.Second, 1*time.Second)
			err := resource.Retry(10*time.Minute, func() *resource.RetryError {
				response, err := client.RpcPost("ARMS", "2019-08-08", action, nil, request, false)
				if err != nil {
					if NeedRetry(err) {
						wait()
						return resource.RetryableError(err)

					}
					return resource.NonRetryableError(err)
				}
				addDebug(action, response, request)
				return nil
			})
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
			}
		}
		if len(added) > 0 {
			action := "TagResources"
			request := map[string]interface{}{
				"RegionId":     s.client.RegionId,
				"ResourceType": resourceType,
			}

			switch resourceIdNum {
			case 0:
				request["ResourceId.1"] = d.Id()
			case 1:
				parts, err := ParseResourceId(d.Id(), 2)
				if err != nil {
					return WrapError(err)
				}
				request["ResourceId.1"] = parts[resourceIdNum]
			}

			count := 1
			for key, value := range added {
				request[fmt.Sprintf("Tag.%d.Key", count)] = key
				request[fmt.Sprintf("Tag.%d.Value", count)] = value
				count++
			}
			wait := incrementalWait(2*time.Second, 1*time.Second)
			err := resource.Retry(10*time.Minute, func() *resource.RetryError {
				response, err := client.RpcPost("ARMS", "2019-08-08", action, nil, request, false)
				if err != nil {
					if NeedRetry(err) {
						wait()
						return resource.RetryableError(err)

					}
					return resource.NonRetryableError(err)
				}
				addDebug(action, response, request)
				return nil
			})
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), action, AlibabaCloudSdkGoERROR)
			}
		}
		d.SetPartial("tags")
	}
	return nil
}

func (s *ArmsService) DescribeArmsIntegrationExporter(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.aliyunArmsAPI != nil {
		parts, err := ParseResourceId(id, 3)
		if err == nil {
			exporter, err := s.aliyunArmsAPI.GetPrometheusIntegration(context.Background(), parts[0], parts[1], parts[2])
			if err == nil {
				// Convert to map[string]interface{} format expected by Terraform
				return map[string]interface{}{
					"ClusterId":       exporter.ClusterId,
					"IntegrationType": exporter.IntegrationType,
					"InstanceId":      exporter.InstanceId,
					"ExporterType":    exporter.ExporterType,
					"Status":          exporter.Status,
					"Target":          exporter.Target,
					"Version":         exporter.Version,
					"Config":          exporter.Config,
				}, nil
			}
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	action := "GetPrometheusIntegration"

	client := s.client

	parts, err := ParseResourceId(id, 3)
	if err != nil {
		return nil, WrapError(err)
	}

	request := map[string]interface{}{
		"RegionId":        s.client.RegionId,
		"ClusterId":       parts[0],
		"IntegrationType": parts[1],
		"InstanceId":      parts[2],
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
		if IsExpectedErrors(err, []string{"404"}) {
			return object, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}

	v, err := jsonpath.Get("$.Data", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Data", response)
	}

	object = v.(map[string]interface{})

	return object, nil
}

func (s *ArmsService) DescribeArmsRemoteWrite(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.aliyunArmsAPI != nil {
		parts, err := ParseResourceId(id, 2)
		if err == nil {
			remoteWrite, err := s.aliyunArmsAPI.GetPrometheusRemoteWrite(context.Background(), parts[0], parts[1])
			if err == nil {
				// Convert to map[string]interface{} format expected by Terraform
				return map[string]interface{}{
					"ClusterId":       remoteWrite.ClusterId,
					"RemoteWriteName": remoteWrite.RemoteWriteName,
					"RemoteWriteYaml": remoteWrite.RemoteWriteYaml,
					"Config":          remoteWrite.Config,
				}, nil
			}
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	action := "GetPrometheusRemoteWrite"

	client := s.client

	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return nil, WrapError(err)
	}

	request := map[string]interface{}{
		"RegionId":        s.client.RegionId,
		"ClusterId":       parts[0],
		"RemoteWriteName": parts[1],
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
		if IsExpectedErrors(err, []string{"404"}) {
			return object, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}

	v, err := jsonpath.Get("$.Data", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Data", response)
	}

	object = v.(map[string]interface{})

	return object, nil
}

func (s *ArmsService) DescribeArmsIntegration(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.aliyunArmsAPI != nil {
		integration, err := s.aliyunArmsAPI.GetIntegration(context.Background(), id)
		if err == nil {
			// Convert to map[string]interface{} format expected by Terraform
			result := map[string]interface{}{
				"IntegrationId":          integration.IntegrationId,
				"IntegrationName":        integration.IntegrationName,
				"IntegrationProductType": integration.IntegrationProductType,
				"ApiEndpoint":            integration.ApiEndpoint,
				"ShortToken":             integration.ShortToken,
				"State":                  integration.State,
				"Liveness":               integration.Liveness,
				"CreateTime":             integration.CreateTime,
			}

			if integration.IntegrationDetail != nil {
				result["IntegrationDetail"] = integration.IntegrationDetail
			}

			return result, nil
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	action := "GetIntegration"
	client := s.client

	request := map[string]interface{}{
		"RegionId":      s.client.RegionId,
		"IntegrationId": id,
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
		if IsExpectedErrors(err, []string{"404"}) {
			return object, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}

	v, err := jsonpath.Get("$.Integration", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Integration", response)
	}

	object = v.(map[string]interface{})
	return object, nil
}

func (s *ArmsService) ListArmsIntegrations() (objects []interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.aliyunArmsAPI != nil {
		integrations, err := s.aliyunArmsAPI.ListAllIntegrations(context.Background(), 1, PageSizeXLarge)
		if err == nil {
			// Convert to the format expected by Terraform
			var result []interface{}
			for _, integration := range integrations {
				integrationMap := map[string]interface{}{
					"IntegrationId":          integration.IntegrationId,
					"IntegrationName":        integration.IntegrationName,
					"IntegrationProductType": integration.IntegrationProductType,
					"ApiEndpoint":            integration.ApiEndpoint,
					"ShortToken":             integration.ShortToken,
					"State":                  integration.State,
					"Liveness":               integration.Liveness,
					"CreateTime":             integration.CreateTime,
				}

				if integration.IntegrationDetail != nil {
					integrationMap["IntegrationDetail"] = integration.IntegrationDetail
				}

				result = append(result, integrationMap)
			}
			return result, nil
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	action := "ListIntegration"
	client := s.client

	request := map[string]interface{}{
		"RegionId": s.client.RegionId,
		"Page":     1,
		"Size":     PageSizeXLarge,
	}

	for {
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
			return objects, WrapErrorf(err, DefaultErrorMsg, "ListIntegration", action, AlibabaCloudSdkGoERROR)
		}

		v, err := jsonpath.Get("$.PageInfo.Integrations", response)
		if err != nil {
			return objects, WrapErrorf(err, FailedGetAttributeMsg, "ListIntegration", "$.PageInfo.Integrations", response)
		}

		if v != nil {
			objects = append(objects, v.([]interface{})...)
		}

		totalCount, err := jsonpath.Get("$.PageInfo.Total", response)
		if err != nil {
			return objects, WrapErrorf(err, FailedGetAttributeMsg, "ListIntegration", "$.PageInfo.Total", response)
		}

		if len(objects) >= int(totalCount.(float64)) {
			break
		}

		if page, err := jsonpath.Get("$.PageInfo.Page", response); err != nil || page == nil {
			break
		} else {
			request["Page"] = int(page.(float64)) + 1
		}
	}

	return objects, nil
}

func (s *ArmsService) ArmsIntegrationStateRefreshFunc(d *schema.ResourceData, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeArmsIntegration(d.Id())
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if fmt.Sprint(object["State"]) == failState {
				return object, fmt.Sprint(object["State"]), WrapError(Error(FailedToReachTargetStatus, fmt.Sprint(object["State"])))
			}
		}

		return object, fmt.Sprint(object["State"]), nil
	}
}
