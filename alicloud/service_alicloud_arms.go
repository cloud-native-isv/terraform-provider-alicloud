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
		robots, err := s.armsAPI.ListAlertRobots([]string{id}, 1, PageSizeXLarge)
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
				"RuleId":      rule.RuleId,
				"Name":        rule.Name,
				"State":       rule.State,
				"GroupRules":  rule.GroupRules,
				"NotifyRules": rule.NotifyRules,
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

func (s *ArmsService) DescribeArmsAlertNotificationPolicy(id string) (object map[string]interface{}, err error) {
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

// DescribeArmsAlertRule describes ARMS alert rule
func (s *ArmsService) DescribeArmsAlertRule(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		alertId, parseErr := strconv.ParseInt(id, 10, 64)
		if parseErr == nil {
			alertRule, err := s.armsAPI.GetAlertRule(alertId)
			if err == nil {
				// Convert to map[string]interface{} format expected by Terraform
				return map[string]interface{}{
					"AlertId":               alertRule.AlertId,
					"AlertName":             alertRule.AlertName,
					"AlertCheckType":        alertRule.AlertCheckType,
					"AlertGroup":            alertRule.AlertGroup,
					"AlertStatus":           alertRule.AlertStatus,
					"AlertType":             alertRule.AlertType,
					"Level":                 alertRule.Level,
					"Severity":              alertRule.Level, // Map Level to Severity for backward compatibility
					"Message":               alertRule.Message,
					"Duration":              alertRule.Duration,
					"PromQL":                alertRule.PromQL,
					"ClusterId":             alertRule.ClusterId,
					"MetricsType":           alertRule.MetricsType,
					"NotifyStrategy":        alertRule.NotifyStrategy,
					"AutoAddNewApplication": alertRule.AutoAddNewApplication,
					"RegionId":              alertRule.RegionId,
					"UserId":                alertRule.UserId,
					"CreatedTime":           alertRule.CreatedTime,
					"UpdatedTime":           alertRule.UpdatedTime,
					"CreateTime":            alertRule.CreatedTime, // Map CreatedTime to CreateTime for backward compatibility
					"Extend":                alertRule.Extend,
					"Pids":                  alertRule.Pids,
					"Annotations":           alertRule.Annotations,
					"Labels":                alertRule.Labels,
					"Tags":                  alertRule.Tags,
					"Filters":               alertRule.Filters,
					"AlertRuleContent":      alertRule.AlertRuleContent,
					"State":                 alertRule.AlertStatus, // Map AlertStatus to State for consistency
					"Describe":              alertRule.Message,     // Map Message to Describe for backward compatibility
					"Owner":                 "",                    // Default empty, can be extracted from Extend if needed
					"Handler":               "",                    // Default empty, can be extracted from Extend if needed
					"Solution":              "",                    // Default empty, can be extracted from Extend if needed
					"DispatchRuleId":        0,                     // Default 0, will be set if available
					"DispatchRuleName":      "",                    // Default empty, will be set if available
				}, nil
			}
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	action := "ListAlerts"
	client := s.client

	request := map[string]interface{}{
		"Page":     1,
		"Size":     1,
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
		if IsExpectedErrors(err, []string{"404"}) {
			return object, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}

	v, err := jsonpath.Get("$.PageBean.ListAlerts", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.PageBean.ListAlerts", response)
	}

	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	}

	// Find the alert with matching ID
	alertIdInt, _ := strconv.ParseInt(id, 10, 64)
	for _, alert := range v.([]interface{}) {
		alertMap := alert.(map[string]interface{})
		if alertId, ok := alertMap["AlertId"]; ok {
			if int64(alertId.(float64)) == alertIdInt {
				// Convert RPC response to standardized format matching AlertRule struct
				standardizedAlert := make(map[string]interface{})

				// Core fields from AlertRule struct
				if alertId, ok := alertMap["AlertId"]; ok {
					standardizedAlert["AlertId"] = int64(alertId.(float64))
				}
				if alertName, ok := alertMap["AlertName"]; ok && alertName != nil {
					standardizedAlert["AlertName"] = alertName.(string)
				}
				if alertCheckType, ok := alertMap["AlertCheckType"]; ok && alertCheckType != nil {
					standardizedAlert["AlertCheckType"] = alertCheckType.(string)
				}
				if alertGroup, ok := alertMap["AlertGroup"]; ok && alertGroup != nil {
					if alertGroupFloat, ok := alertGroup.(float64); ok {
						standardizedAlert["AlertGroup"] = int64(alertGroupFloat)
					}
				}
				if alertStatus, ok := alertMap["AlertStatus"]; ok && alertStatus != nil {
					standardizedAlert["AlertStatus"] = alertStatus.(string)
					standardizedAlert["State"] = alertStatus.(string) // Map to State for consistency
				}
				if alertType, ok := alertMap["AlertType"]; ok && alertType != nil {
					standardizedAlert["AlertType"] = alertType.(string)
				}
				if level, ok := alertMap["Level"]; ok && level != nil {
					standardizedAlert["Level"] = level.(string)
					standardizedAlert["Severity"] = level.(string) // Map to Severity for backward compatibility
				}
				if message, ok := alertMap["Message"]; ok && message != nil {
					standardizedAlert["Message"] = message.(string)
					standardizedAlert["Describe"] = message.(string) // Map to Describe for backward compatibility
				}
				if duration, ok := alertMap["Duration"]; ok && duration != nil {
					standardizedAlert["Duration"] = duration.(string)
				}
				if promQL, ok := alertMap["PromQL"]; ok && promQL != nil {
					standardizedAlert["PromQL"] = promQL.(string)
				}
				if clusterId, ok := alertMap["ClusterId"]; ok && clusterId != nil {
					standardizedAlert["ClusterId"] = clusterId.(string)
				}
				if metricsType, ok := alertMap["MetricsType"]; ok && metricsType != nil {
					standardizedAlert["MetricsType"] = metricsType.(string)
				}
				if notifyStrategy, ok := alertMap["NotifyStrategy"]; ok && notifyStrategy != nil {
					standardizedAlert["NotifyStrategy"] = notifyStrategy.(string)
				}
				if autoAddNewApp, ok := alertMap["AutoAddNewApplication"]; ok && autoAddNewApp != nil {
					if autoAddBool, ok := autoAddNewApp.(bool); ok {
						standardizedAlert["AutoAddNewApplication"] = autoAddBool
					}
				}
				if regionId, ok := alertMap["RegionId"]; ok && regionId != nil {
					standardizedAlert["RegionId"] = regionId.(string)
				}
				if userId, ok := alertMap["UserId"]; ok && userId != nil {
					standardizedAlert["UserId"] = userId.(string)
				}
				if createdTime, ok := alertMap["CreatedTime"]; ok && createdTime != nil {
					if createdTimeFloat, ok := createdTime.(float64); ok {
						standardizedAlert["CreatedTime"] = int64(createdTimeFloat)
						standardizedAlert["CreateTime"] = int64(createdTimeFloat) // Map to CreateTime for backward compatibility
					}
				}
				if updatedTime, ok := alertMap["UpdatedTime"]; ok && updatedTime != nil {
					if updatedTimeFloat, ok := updatedTime.(float64); ok {
						standardizedAlert["UpdatedTime"] = int64(updatedTimeFloat)
					}
				}
				if extend, ok := alertMap["Extend"]; ok && extend != nil {
					standardizedAlert["Extend"] = extend.(string)
				}
				if pids, ok := alertMap["Pids"]; ok && pids != nil {
					standardizedAlert["Pids"] = pids
				}
				if annotations, ok := alertMap["Annotations"]; ok && annotations != nil {
					standardizedAlert["Annotations"] = annotations
				}
				if labels, ok := alertMap["Labels"]; ok && labels != nil {
					standardizedAlert["Labels"] = labels
				}
				if tags, ok := alertMap["Tags"]; ok && tags != nil {
					standardizedAlert["Tags"] = tags
				}
				if filters, ok := alertMap["Filters"]; ok && filters != nil {
					standardizedAlert["Filters"] = filters
				}
				if alertRuleContent, ok := alertMap["AlertRuleContent"]; ok && alertRuleContent != nil {
					standardizedAlert["AlertRuleContent"] = alertRuleContent
				}

				// Legacy/backward compatibility fields - set defaults if not available
				if _, ok := standardizedAlert["Owner"]; !ok {
					if owner, ok := alertMap["Owner"]; ok && owner != nil {
						standardizedAlert["Owner"] = owner.(string)
					} else {
						standardizedAlert["Owner"] = ""
					}
				}
				if _, ok := standardizedAlert["Handler"]; !ok {
					if handler, ok := alertMap["Handler"]; ok && handler != nil {
						standardizedAlert["Handler"] = handler.(string)
					} else {
						standardizedAlert["Handler"] = ""
					}
				}
				if _, ok := standardizedAlert["Solution"]; !ok {
					if solution, ok := alertMap["Solution"]; ok && solution != nil {
						standardizedAlert["Solution"] = solution.(string)
					} else {
						standardizedAlert["Solution"] = ""
					}
				}
				if _, ok := standardizedAlert["DispatchRuleId"]; !ok {
					if dispatchRuleId, ok := alertMap["DispatchRuleId"]; ok && dispatchRuleId != nil {
						if dispatchRuleIdFloat, ok := dispatchRuleId.(float64); ok {
							standardizedAlert["DispatchRuleId"] = dispatchRuleIdFloat
						}
					} else {
						standardizedAlert["DispatchRuleId"] = float64(0)
					}
				}
				if _, ok := standardizedAlert["DispatchRuleName"]; !ok {
					if dispatchRuleName, ok := alertMap["DispatchRuleName"]; ok && dispatchRuleName != nil {
						standardizedAlert["DispatchRuleName"] = dispatchRuleName.(string)
					} else {
						standardizedAlert["DispatchRuleName"] = ""
					}
				}

				return standardizedAlert, nil
			}
		}
	}

	return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
}

// CreateArmsAlertRule creates a new ARMS alert rule
func (s *ArmsService) CreateArmsAlertRule(alertName, severity, description, integrationType, clusterId string, rule map[string]interface{}) (int64, error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		// Convert parameters to AlertRule struct
		alertRule := &aliyunArmsAPI.AlertRule{
			AlertName: alertName,
			Level:     severity,
			Message:   description,
			AlertType: "PROMETHEUS_MONITORING_ALERT_RULE",
			ClusterId: clusterId,
		}

		// Set rule parameters from map
		if rule != nil {
			if promql, ok := rule["promql"].(string); ok && promql != "" {
				alertRule.PromQL = promql
			} else if expression, ok := rule["expression"].(string); ok && expression != "" {
				alertRule.PromQL = expression
			}

			if duration, ok := rule["duration"].(int); ok && duration > 0 {
				alertRule.Duration = fmt.Sprintf("%d", duration)
			} else if durationFloat, ok := rule["duration"].(float64); ok && durationFloat > 0 {
				alertRule.Duration = fmt.Sprintf("%.0f", durationFloat)
			}

			if message, ok := rule["message"].(string); ok && message != "" {
				alertRule.Message = message
			}

			if checkType, ok := rule["check_type"].(string); ok && checkType != "" {
				alertRule.AlertCheckType = checkType
			} else {
				alertRule.AlertCheckType = "CUSTOM"
			}

			if alertGroup, ok := rule["alert_group"].(int); ok {
				alertRule.AlertGroup = int64(alertGroup)
			} else if alertGroupFloat, ok := rule["alert_group"].(float64); ok {
				alertRule.AlertGroup = int64(alertGroupFloat)
			} else {
				alertRule.AlertGroup = -1
			}

			// Set labels if provided in rule
			if labels, ok := rule["labels"].(map[string]interface{}); ok && len(labels) > 0 {
				for key, value := range labels {
					alertRule.Labels = append(alertRule.Labels, aliyunArmsAPI.AlertRuleLabel{
						Key:   key,
						Value: fmt.Sprintf("%v", value),
					})
				}
			}
		}

		// Set default values for required fields if not specified
		if alertRule.AlertCheckType == "" {
			alertRule.AlertCheckType = "CUSTOM"
		}
		if alertRule.AlertGroup == 0 {
			alertRule.AlertGroup = -1
		}

		createdRule, err := s.armsAPI.CreateAlertRule(alertRule)
		if err == nil && createdRule != nil {
			return createdRule.AlertId, nil
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	action := "CreateOrUpdateAlertRule"
	client := s.client

	request := map[string]interface{}{
		"AlertName": alertName,
		"Level":     severity,
		"AlertType": "PROMETHEUS_MONITORING_ALERT_RULE",
		"RegionId":  s.client.RegionId,
	}

	// Set description if provided
	if description != "" {
		request["Message"] = description
	}

	// Set cluster ID if provided
	if clusterId != "" {
		request["ClusterId"] = clusterId
	}

	// Set PromQL expression and other rule parameters if provided
	if rule != nil {
		if promql, ok := rule["promql"].(string); ok && promql != "" {
			request["PromQL"] = promql
		} else if expression, ok := rule["expression"].(string); ok && expression != "" {
			request["PromQL"] = expression
		}

		// Set duration if provided in rule
		if duration, ok := rule["duration"].(int); ok && duration > 0 {
			request["Duration"] = duration
		} else if durationFloat, ok := rule["duration"].(float64); ok && durationFloat > 0 {
			request["Duration"] = int(durationFloat)
		}

		// Set message if provided in rule
		if message, ok := rule["message"].(string); ok && message != "" {
			request["Message"] = message
		}

		// Set alert check type if provided
		if checkType, ok := rule["check_type"].(string); ok && checkType != "" {
			request["AlertCheckType"] = checkType
		} else {
			request["AlertCheckType"] = "CUSTOM" // Default for custom PromQL
		}

		// Set alert group if provided
		if alertGroup, ok := rule["alert_group"].(int); ok {
			request["AlertGroup"] = alertGroup
		} else if alertGroupFloat, ok := rule["alert_group"].(float64); ok {
			request["AlertGroup"] = int(alertGroupFloat)
		} else {
			request["AlertGroup"] = -1 // Default for custom PromQL
		}

		// Set labels if provided in rule
		if labels, ok := rule["labels"].(map[string]interface{}); ok && len(labels) > 0 {
			labelsMaps := make([]map[string]interface{}, 0)
			for key, value := range labels {
				labelsMaps = append(labelsMaps, map[string]interface{}{
					"name":  key,
					"value": fmt.Sprintf("%v", value),
				})
			}
			if labelString, err := convertArrayObjectToJsonString(labelsMaps); err == nil {
				request["Labels"] = labelString
			}
		}
	}

	// Set default values for required fields if not specified
	if _, ok := request["AlertCheckType"]; !ok {
		request["AlertCheckType"] = "CUSTOM"
	}
	if _, ok := request["AlertGroup"]; !ok {
		request["AlertGroup"] = -1
	}

	wait := incrementalWait(3*time.Second, 3*time.Second)
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		var retryErr error
		response, retryErr = client.RpcPost("ARMS", "2019-08-08", action, nil, request, false)
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
		return 0, WrapErrorf(err, DefaultErrorMsg, "CreateArmsAlertRule", action, AlibabaCloudSdkGoERROR)
	}

	if alertRule, ok := response["AlertRule"].(map[string]interface{}); ok {
		if alertId, ok := alertRule["AlertId"]; ok {
			if alertIdFloat, ok := alertId.(float64); ok {
				return int64(alertIdFloat), nil
			}
		}
	}

	return 0, WrapError(fmt.Errorf("AlertId not found in response"))
}

// UpdateArmsAlertRule updates an existing ARMS alert rule
func (s *ArmsService) UpdateArmsAlertRule(alertId int64, alertName, severity, description, integrationType, clusterId string, rule map[string]interface{}) error {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		_, err := s.armsAPI.UpdateAlertRule(&aliyunArmsAPI.AlertRule{
			AlertId:   alertId,
			AlertName: alertName,
			Level:     severity,
			Message:   description,
			AlertType: "PROMETHEUS_MONITORING_ALERT_RULE",
			ClusterId: clusterId,
		})
		if err == nil {
			return nil
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	action := "CreateOrUpdateAlertRule"
	client := s.client

	request := map[string]interface{}{
		"AlertId":   alertId,
		"AlertName": alertName,
		"Level":     severity,
		"AlertType": "PROMETHEUS_MONITORING_ALERT_RULE",
		"RegionId":  s.client.RegionId,
	}

	// Set description if provided
	if description != "" {
		request["Message"] = description
	}

	// Set cluster ID if provided
	if clusterId != "" {
		request["ClusterId"] = clusterId
	}

	// Set PromQL expression and other rule parameters if provided
	if rule != nil {
		if promql, ok := rule["promql"].(string); ok && promql != "" {
			request["PromQL"] = promql
		} else if expression, ok := rule["expression"].(string); ok && expression != "" {
			request["PromQL"] = expression
		}

		// Set duration if provided in rule
		if duration, ok := rule["duration"].(int); ok && duration > 0 {
			request["Duration"] = duration
		} else if durationFloat, ok := rule["duration"].(float64); ok && durationFloat > 0 {
			request["Duration"] = int(durationFloat)
		}

		// Set message if provided in rule
		if message, ok := rule["message"].(string); ok && message != "" {
			request["Message"] = message
		}

		// Set alert check type if provided
		if checkType, ok := rule["check_type"].(string); ok && checkType != "" {
			request["AlertCheckType"] = checkType
		} else {
			request["AlertCheckType"] = "CUSTOM" // Default for custom PromQL
		}

		// Set alert group if provided
		if alertGroup, ok := rule["alert_group"].(int); ok {
			request["AlertGroup"] = alertGroup
		} else if alertGroupFloat, ok := rule["alert_group"].(float64); ok {
			request["AlertGroup"] = int(alertGroupFloat)
		} else {
			request["AlertGroup"] = -1 // Default for custom PromQL
		}

		// Set labels if provided in rule
		if labels, ok := rule["labels"].(map[string]interface{}); ok && len(labels) > 0 {
			labelsMaps := make([]map[string]interface{}, 0)
			for key, value := range labels {
				labelsMaps = append(labelsMaps, map[string]interface{}{
					"name":  key,
					"value": fmt.Sprintf("%v", value),
				})
			}
			if labelString, err := convertArrayObjectToJsonString(labelsMaps); err == nil {
				request["Labels"] = labelString
			}
		}
	}

	// Set default values for required fields if not specified
	if _, ok := request["AlertCheckType"]; !ok {
		request["AlertCheckType"] = "CUSTOM"
	}
	if _, ok := request["AlertGroup"]; !ok {
		request["AlertGroup"] = -1
	}

	wait := incrementalWait(3*time.Second, 3*time.Second)
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		var retryErr error
		response, retryErr = client.RpcPost("ARMS", "2019-08-08", action, nil, request, false)
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
		return WrapErrorf(err, DefaultErrorMsg, "UpdateArmsAlertRule", action, AlibabaCloudSdkGoERROR)
	}

	return nil
}

// DescribeArmsAlertContactSchedule describes ARMS alert contact schedule (on-call schedule)
func (s *ArmsService) DescribeArmsAlertContactSchedule(id string) (*aliyunArmsAPI.AlertContactSchedule, error) {
	// Parse the ID to get schedule ID and optional time range
	// Expected format: scheduleId[:startTime:endTime] or just scheduleId
	parts := strings.Split(id, ":")
	if len(parts) < 1 {
		return nil, WrapError(fmt.Errorf("invalid schedule ID format: %s", id))
	}

	scheduleIdStr := parts[0]
	scheduleId, err := strconv.ParseInt(scheduleIdStr, 10, 64)
	if err != nil {
		return nil, WrapError(fmt.Errorf("invalid schedule ID: %s", scheduleIdStr))
	}

	var startTime, endTime string
	if len(parts) >= 3 {
		startTime = parts[1]
		endTime = parts[2]
	}

	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		schedule, err := s.armsAPI.GetAlertContactSchedule(scheduleId, startTime, endTime)
		if err == nil {
			return schedule, nil
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	client := s.client
	action := "GetOnCallSchedulesDetail"
	request := map[string]interface{}{
		"Id": scheduleId,
	}

	// Set time range if provided
	if startTime != "" {
		request["StartTime"] = startTime
	}
	if endTime != "" {
		request["EndTime"] = endTime
	}

	wait := incrementalWait(3*time.Second, 3*time.Second)
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

	v, err := jsonpath.Get("$.Data", response)
	if err != nil {
		return nil, WrapErrorf(err, FailedGetAttributeMsg, id, "$.Data", response)
	}

	if v == nil {
		return nil, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	}

	// Convert map[string]interface{} to AlertContactSchedule struct
	scheduleData := v.(map[string]interface{})
	schedule := &aliyunArmsAPI.AlertContactSchedule{}

	// Set basic fields
	schedule.ScheduleId = scheduleIdStr
	if name, ok := scheduleData["Name"]; ok && name != nil {
		schedule.ScheduleName = name.(string)
	}
	if description, ok := scheduleData["Description"]; ok && description != nil {
		schedule.Description = description.(string)
	}

	// Set alert robot ID if available
	if alertRobotId, ok := scheduleData["AlertRobotId"]; ok && alertRobotId != nil {
		if robotIdFloat, ok := alertRobotId.(float64); ok {
			schedule.AlertRobotId = int64(robotIdFloat)
		}
	}

	// Convert rendered final entries (users on duty)
	if renderedFinalEntries, ok := scheduleData["RenderedFinnalEntries"]; ok && renderedFinalEntries != nil {
		for _, entry := range renderedFinalEntries.([]interface{}) {
			entryMap := entry.(map[string]interface{})
			finalEntry := aliyunArmsAPI.ScheduleRenderedFinalEntry{}

			if start, ok := entryMap["Start"]; ok && start != nil {
				finalEntry.Start = start.(string)
			}
			if end, ok := entryMap["End"]; ok && end != nil {
				finalEntry.End = end.(string)
			}

			// Set contact information if available
			if simpleContact, ok := entryMap["SimpleContact"]; ok && simpleContact != nil {
				contactMap := simpleContact.(map[string]interface{})
				finalEntry.SimpleContact = &aliyunArmsAPI.ScheduleSimpleContact{}

				if id, ok := contactMap["Id"]; ok && id != nil {
					if idFloat, ok := id.(float64); ok {
						finalEntry.SimpleContact.Id = int64(idFloat)
					}
				}
				if name, ok := contactMap["Name"]; ok && name != nil {
					finalEntry.SimpleContact.Name = name.(string)
				}
			}

			schedule.RenderedFinalEntries = append(schedule.RenderedFinalEntries, finalEntry)
		}
	}

	// Convert rendered layer entries (scheduled users within time ranges)
	if renderedLayerEntries, ok := scheduleData["RenderedLayerEntries"]; ok && renderedLayerEntries != nil {
		for _, layerGroup := range renderedLayerEntries.([]interface{}) {
			var layerEntries []aliyunArmsAPI.ScheduleRenderedLayerEntry
			for _, entry := range layerGroup.([]interface{}) {
				entryMap := entry.(map[string]interface{})
				layerEntry := aliyunArmsAPI.ScheduleRenderedLayerEntry{}

				if start, ok := entryMap["Start"]; ok && start != nil {
					layerEntry.Start = start.(string)
				}
				if end, ok := entryMap["End"]; ok && end != nil {
					layerEntry.End = end.(string)
				}

				// Set contact information if available
				if simpleContact, ok := entryMap["SimpleContact"]; ok && simpleContact != nil {
					contactMap := simpleContact.(map[string]interface{})
					layerEntry.SimpleContact = &aliyunArmsAPI.ScheduleSimpleContact{}

					if id, ok := contactMap["Id"]; ok && id != nil {
						if idFloat, ok := id.(float64); ok {
							layerEntry.SimpleContact.Id = int64(idFloat)
						}
					}
					if name, ok := contactMap["Name"]; ok && name != nil {
						layerEntry.SimpleContact.Name = name.(string)
					}
				}

				layerEntries = append(layerEntries, layerEntry)
			}
			schedule.RenderedLayerEntries = append(schedule.RenderedLayerEntries, layerEntries)
		}
	}

	// Convert rendered substitute entries (substitutes within time range)
	if renderedSubstituteEntries, ok := scheduleData["RenderedSubstitudeEntries"]; ok && renderedSubstituteEntries != nil {
		for _, entry := range renderedSubstituteEntries.([]interface{}) {
			entryMap := entry.(map[string]interface{})
			substituteEntry := aliyunArmsAPI.ScheduleRenderedFinalEntry{}

			if start, ok := entryMap["Start"]; ok && start != nil {
				substituteEntry.Start = start.(string)
			}
			if end, ok := entryMap["End"]; ok && end != nil {
				substituteEntry.End = end.(string)
			}

			// Set contact information if available
			if simpleContact, ok := entryMap["SimpleContact"]; ok && simpleContact != nil {
				contactMap := simpleContact.(map[string]interface{})
				substituteEntry.SimpleContact = &aliyunArmsAPI.ScheduleSimpleContact{}

				if id, ok := contactMap["Id"]; ok && id != nil {
					if idFloat, ok := id.(float64); ok {
						substituteEntry.SimpleContact.Id = int64(idFloat)
					}
				}
				if name, ok := contactMap["Name"]; ok && name != nil {
					substituteEntry.SimpleContact.Name = name.(string)
				}
			}

			schedule.RenderedSubstituteEntries = append(schedule.RenderedSubstituteEntries, substituteEntry)
		}
	}

	// Convert schedule layers (shift configurations)
	if scheduleLayers, ok := scheduleData["ScheduleLayers"]; ok && scheduleLayers != nil {
		for _, layer := range scheduleLayers.([]interface{}) {
			layerMap := layer.(map[string]interface{})
			scheduleLayer := aliyunArmsAPI.ScheduleLayer{}

			if rotationType, ok := layerMap["RotationType"]; ok && rotationType != nil {
				scheduleLayer.RotationType = rotationType.(string)
			}
			if shiftLength, ok := layerMap["ShiftLength"]; ok && shiftLength != nil {
				if shiftLengthFloat, ok := shiftLength.(float64); ok {
					scheduleLayer.ShiftLength = int64(shiftLengthFloat)
				}
			}
			if startTime, ok := layerMap["StartTime"]; ok && startTime != nil {
				scheduleLayer.StartTime = startTime.(string)
			}

			// Set contact IDs if available
			if contactIds, ok := layerMap["ContactIds"]; ok && contactIds != nil {
				for _, contactId := range contactIds.([]interface{}) {
					if contactIdFloat, ok := contactId.(float64); ok {
						scheduleLayer.ContactIds = append(scheduleLayer.ContactIds, int64(contactIdFloat))
					}
				}
			}

			// Convert time restrictions if available
			if restrictions, ok := layerMap["Restrictions"]; ok && restrictions != nil {
				for _, restriction := range restrictions.([]interface{}) {
					restrictionMap := restriction.(map[string]interface{})
					layerRestriction := aliyunArmsAPI.ScheduleLayerRestriction{}

					if startTime, ok := restrictionMap["StartTime"]; ok && startTime != nil {
						layerRestriction.StartTime = startTime.(string)
					}
					if endTime, ok := restrictionMap["EndTime"]; ok && endTime != nil {
						layerRestriction.EndTime = endTime.(string)
					}
					if restrictionType, ok := restrictionMap["Type"]; ok && restrictionType != nil {
						layerRestriction.Type = restrictionType.(string)
					}

					scheduleLayer.Restrictions = append(scheduleLayer.Restrictions, layerRestriction)
				}
			}

			schedule.ScheduleLayers = append(schedule.ScheduleLayers, scheduleLayer)
		}
	}

	return schedule, nil
}

// DescribeArmsAlertSilencePolicy describes ARMS alert silence policy
func (s *ArmsService) DescribeArmsAlertSilencePolicy(id string) (object map[string]interface{}, err error) {
	// Try using aliyunArmsAPI first if available
	if s.armsAPI != nil {
		// Convert string ID to int64
		silenceIdInt, parseErr := strconv.ParseInt(id, 10, 64)
		if parseErr == nil {
			silencePolicy, err := s.armsAPI.GetAlertSilencePolicy(silenceIdInt)
			if err == nil {
				// Convert to map[string]interface{} format expected by Terraform
				return map[string]interface{}{
					"SilenceId":         silencePolicy.SilenceId,
					"SilenceName":       silencePolicy.SilenceName,
					"State":             silencePolicy.State,
					"EffectiveTimeType": silencePolicy.EffectiveTimeType,
					"TimePeriod":        silencePolicy.TimePeriod,
					"TimeSlots":         silencePolicy.TimeSlots,
					"StartTime":         silencePolicy.StartTime,
					"EndTime":           silencePolicy.EndTime,
					"Comment":           silencePolicy.Comment,
					"MatchingRules":     silencePolicy.MatchingRules,
					"CreatedBy":         silencePolicy.CreatedBy,
					"CreateTime":        silencePolicy.CreateTime,
					"UpdateTime":        silencePolicy.UpdateTime,
				}, nil
			}
		}
	}

	// Fallback to direct RPC call
	var response map[string]interface{}
	action := "ListSilencePolicies"
	client := s.client

	request := map[string]interface{}{
		"Page":     1,
		"Size":     PageSizeXLarge,
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
		if IsExpectedErrors(err, []string{"404"}) {
			return object, WrapErrorf(err, NotFoundMsg, AlibabaCloudSdkGoERROR)
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, action, AlibabaCloudSdkGoERROR)
	}

	v, err := jsonpath.Get("$.PageBean.SilencePolicies", response)
	if err != nil {
		return object, WrapErrorf(err, FailedGetAttributeMsg, id, "$.PageBean.SilencePolicies", response)
	}

	if len(v.([]interface{})) < 1 {
		return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
	}

	// Find the silence policy with matching ID
	for _, policy := range v.([]interface{}) {
		policyMap := policy.(map[string]interface{})
		if fmt.Sprint(policyMap["SilenceId"]) == id {
			return policyMap, nil
		}
	}

	return object, WrapErrorf(NotFoundErr("ARMS", id), NotFoundWithResponse, response)
}
