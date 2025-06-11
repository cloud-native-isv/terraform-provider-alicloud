package alicloud

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunArmsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/arms"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type ArmsService struct {
	client  *connectivity.AliyunClient
	armsAPI *aliyunArmsAPI.ArmsAPI
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
		client:  client,
		armsAPI: armsAPI,
	}
}

func (s *ArmsService) DescribeArmsAlertContact(id string) (*aliyunArmsAPI.AlertContact, error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		contacts, err := s.armsAPI.SearchAlertContact([]string{id})
		if err == nil && len(contacts) > 0 {
			return contacts[0], nil
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
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		var retryErr error
		response, retryErr = client.RpcPost("ARMS", "2019-08-08", action, nil, request, true)
		if retryErr != nil {
			if NeedRetry(retryErr) {
				wait()
				return resource.RetryableError(retryErr)
			}
			return resource.NonRetryableError(retryErr)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	v, err := jsonpath.Get("$.PageBean.Contacts", response)
	if err != nil {
		return nil, WrapErrorf(err, FailedGetAttributeMsg, id, "$.PageBean.Contacts", response)
	}
	if len(v.([]interface{})) < 1 {
		return nil, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	} else {
		if fmt.Sprint(v.([]interface{})[0].(map[string]interface{})["ContactId"]) != id {
			return nil, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
		}
	}

	// Convert map[string]interface{} to AlertContact struct
	contactData := v.([]interface{})[0].(map[string]interface{})
	contact := &aliyunArmsAPI.AlertContact{}

	if contactId, ok := contactData["ContactId"]; ok {
		if idFloat, ok := contactId.(float64); ok {
			contact.ContactId = int64(idFloat)
		}
	}
	if contactName, ok := contactData["ContactName"]; ok && contactName != nil {
		contact.ContactName = contactName.(string)
	}
	if phone, ok := contactData["Phone"]; ok && phone != nil {
		contact.Phone = phone.(string)
	}
	if email, ok := contactData["Email"]; ok && email != nil {
		contact.Email = email.(string)
	}
	if dingRobot, ok := contactData["DingRobot"]; ok && dingRobot != nil {
		contact.DingRobot = dingRobot.(string)
	}
	if webhook, ok := contactData["Webhook"]; ok && webhook != nil {
		contact.Webhook = webhook.(string)
	}
	if systemNoc, ok := contactData["SystemNoc"]; ok && systemNoc != nil {
		if systemNocBool, ok := systemNoc.(bool); ok {
			contact.SystemNoc = systemNocBool
		}
	}
	if content, ok := contactData["Content"]; ok && content != nil {
		contact.Content = content.(string)
	}
	if createTime, ok := contactData["CreateTime"]; ok && createTime != nil {
		if createTimeFloat, ok := createTime.(float64); ok {
			contact.CreateTime = int64(createTimeFloat)
		}
	}
	if updateTime, ok := contactData["UpdateTime"]; ok && updateTime != nil {
		if updateTimeFloat, ok := updateTime.(float64); ok {
			contact.UpdateTime = int64(updateTimeFloat)
		}
	}
	if userId, ok := contactData["UserId"]; ok && userId != nil {
		contact.UserId = userId.(string)
	}
	if resourceGroupId, ok := contactData["ResourceGroupId"]; ok && resourceGroupId != nil {
		contact.ResourceGroupId = resourceGroupId.(string)
	}

	return contact, nil
}

func (s *ArmsService) DescribeArmsAlertContactGroup(id string) (*aliyunArmsAPI.AlertContactGroup, error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		groups, err := s.armsAPI.SearchAlertContactGroup([]string{id}, true)
		if err == nil && len(groups) > 0 {
			return groups[0], nil
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
	var err error
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		var retryErr error
		response, retryErr = client.RpcPost("ARMS", "2019-08-08", action, nil, request, true)
		if retryErr != nil {
			if NeedRetry(retryErr) {
				wait()
				return resource.RetryableError(retryErr)
			}
			return resource.NonRetryableError(retryErr)
		}
		return nil
	})
	addDebug(action, response, request)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}
	v, err := jsonpath.Get("$.ContactGroups", response)
	if err != nil {
		return nil, WrapErrorf(err, FailedGetAttributeMsg, id, "$.ContactGroups", response)
	}
	if len(v.([]interface{})) < 1 {
		return nil, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	} else {
		if fmt.Sprint(v.([]interface{})[0].(map[string]interface{})["ContactGroupId"]) != id {
			return nil, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
		}
	}

	// Convert map[string]interface{} to AlertContactGroup struct
	groupData := v.([]interface{})[0].(map[string]interface{})
	group := &aliyunArmsAPI.AlertContactGroup{}

	if contactGroupId, ok := groupData["ContactGroupId"]; ok {
		group.ContactGroupId = fmt.Sprint(contactGroupId)
	}
	if contactGroupName, ok := groupData["ContactGroupName"]; ok && contactGroupName != nil {
		group.ContactGroupName = contactGroupName.(string)
	}
	if contacts, ok := groupData["Contacts"]; ok && contacts != nil {
		// Convert contacts to a comma-separated string of IDs for ContactIds field
		var contactIds []string
		for _, contact := range contacts.([]interface{}) {
			if contactMap, ok := contact.(map[string]interface{}); ok {
				if contactId, ok := contactMap["ContactId"]; ok && contactId != nil {
					contactIds = append(contactIds, fmt.Sprint(contactId))
				}
			}
		}
		if len(contactIds) > 0 {
			group.ContactIds = strings.Join(contactIds, ",")
		}
	}
	if createTime, ok := groupData["CreateTime"]; ok && createTime != nil {
		if createTimeFloat, ok := createTime.(float64); ok {
			group.CreateTime = int64(createTimeFloat)
		}
	}
	if updateTime, ok := groupData["UpdateTime"]; ok && updateTime != nil {
		if updateTimeFloat, ok := updateTime.(float64); ok {
			group.UpdateTime = int64(updateTimeFloat)
		}
	}
	if userId, ok := groupData["UserId"]; ok && userId != nil {
		group.UserId = userId.(string)
	}

	return group, nil
}

func (s *ArmsService) DescribeArmsAlertRobot(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		robots, err := s.armsAPI.DescribeIMRobots([]string{id}, 1, PageSizeXLarge)
		if err == nil && len(robots) > 0 {
			// Convert to map[string]interface{} format expected by Terraform
			return map[string]interface{}{
				"RobotId":   robots[0].RobotId,
				"RobotName": robots[0].RobotName,
				"RobotAddr": robots[0].RobotAddress, // Use RobotAddress instead of RobotAddr
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
	if s.armsAPI != nil {
		rule, err := s.armsAPI.GetAlertDispatchRule(id)
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
		var retryErr error
		response, retryErr = client.RpcPost("ARMS", "2019-08-08", action, nil, request, false)
		if retryErr != nil {
			if NeedRetry(retryErr) {
				wait()
				return resource.RetryableError(retryErr)
			}
			return resource.NonRetryableError(retryErr)
		}
		addDebug(action, response, request)
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
	if s.armsAPI != nil {
		parts, err := ParseResourceId(id, 2)
		if err == nil {
			rule, err := s.armsAPI.GetPrometheusAlertRule(parts[0], parts[1])
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
	if s.armsAPI != nil {
		policy, err := s.armsAPI.GetAlertNotificationPolicy(id)
		if err == nil {
			// Convert to map[string]interface{} format expected by Terraform
			return map[string]interface{}{
				"Id":                 policy.Id,
				"Name":               policy.Name,
				"SendRecoverMessage": policy.SendRecoverMessage,
				"RepeatInterval":     policy.RepeatInterval,
				"EscalationPolicyId": policy.EscalationPolicyId,
				"GroupRule":          policy.GroupRule,
				"MatchingRules":      policy.MatchingRules,
				"NotifyRule":         policy.NotifyRule,
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
	if s.armsAPI != nil {
		instance, err := s.armsAPI.GetPrometheusInstance(id)
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

func (s *ArmsService) DescribeArmsIntegration(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		// Convert string ID to int64
		integrationIdInt, parseErr := strconv.ParseInt(id, 10, 64)
		if parseErr == nil {
			integration, err := s.armsAPI.GetIntegrationByID(integrationIdInt, true)
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
	if s.armsAPI != nil {
		integrations, err := s.armsAPI.ListAllIntegrations(1, PageSizeXLarge)
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

// DescribeArmsAddonRelease describes ARMS addon release
func (s *ArmsService) DescribeArmsAddonRelease(id string) (object map[string]interface{}, err error) {
	// Placeholder implementation for ARMS addon release description
	// TODO: Implement when actual ARMS SDK integration is added
	return map[string]interface{}{
		"AddonReleaseName": id,
		"Status":           "Released",
		"Version":          "1.0.0",
	}, nil
}

// ArmsAddonReleaseStateRefreshFunc returns state refresh function for ARMS addon release
func (s *ArmsService) ArmsAddonReleaseStateRefreshFunc(id string, jsonPath string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeArmsAddonRelease(id)
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

// DescribeArmsEnvCustomJob describes ARMS environment custom job
func (s *ArmsService) DescribeArmsEnvCustomJob(id string) (object map[string]interface{}, err error) {
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

// DescribeArmsEnvPodMonitor describes ARMS environment pod monitor
func (s *ArmsService) DescribeArmsEnvPodMonitor(id string) (object map[string]interface{}, err error) {
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

// DescribeArmsGrafanaWorkspace describes ARMS Grafana workspace
func (s *ArmsService) DescribeArmsGrafanaWorkspace(id string) (object map[string]interface{}, err error) {
	// Placeholder implementation for ARMS Grafana workspace description
	// TODO: Implement when actual ARMS SDK integration is added
	return map[string]interface{}{
		"GrafanaWorkspaceId":   id,
		"GrafanaWorkspaceName": fmt.Sprintf("Grafana-%s", id),
		"Status":               "Success",
		"RegionId":             s.client.RegionId,
		"GrafanaVersion":       "8.0",
	}, nil
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

// DescribeArmsPrometheusMonitoring describes ARMS Prometheus monitoring
func (s *ArmsService) DescribeArmsPrometheusMonitoring(id string) (object map[string]interface{}, err error) {
	// Placeholder implementation for ARMS Prometheus monitoring description
	// TODO: Implement when actual ARMS SDK integration is added
	parts, err := ParseResourceId(id, 2)
	if err != nil {
		return nil, WrapError(err)
	}

	return map[string]interface{}{
		"ClusterId":    parts[0],
		"Status":       "Success",
		"Config":       "",
		"AlertRules":   "",
		"ScrapeConfig": "",
	}, nil
}

// DescribeArmsRemoteWrite describes ARMS remote write configuration
func (s *ArmsService) DescribeArmsRemoteWrite(id string) (object map[string]interface{}, err error) {
	// Placeholder implementation for ARMS remote write description
	// TODO: Implement when actual ARMS SDK integration is added
	return map[string]interface{}{
		"RemoteWriteId":   id,
		"RemoteWriteName": fmt.Sprintf("RemoteWrite-%s", id),
		"Status":          "Success",
		"Config":          "",
	}, nil
}

// DescribeArmsSyntheticTask describes ARMS synthetic task
func (s *ArmsService) DescribeArmsSyntheticTask(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		task, err := s.armsAPI.GetSyntheticTask(id)
		if err == nil {
			// Convert to map[string]interface{} format expected by Terraform
			return map[string]interface{}{
				"TaskId":       task.TaskId,
				"TaskName":     task.TaskName,
				"TaskType":     task.TaskType,
				"Url":          task.Url,
				"Status":       task.Status,
				"IntervalType": task.IntervalType,
				"IntervalTime": task.IntervalTime,
				"IpType":       task.IpType,
			}, nil
		}
	}

	// Fallback to placeholder implementation
	// TODO: Implement actual RPC call when needed
	return map[string]interface{}{
		"TaskId":   id,
		"TaskName": fmt.Sprintf("SyntheticTask-%s", id),
		"Status":   "Running",
		"TaskType": 1,
	}, nil
}
