package alicloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeSlsCollectionPolicy <<< Encapsulated get interface for Sls CollectionPolicy.

func (s *SlsService) DescribeSlsCollectionPolicy(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	ctx := context.Background()
	collectionPolicy, err := s.aliyunSlsAPI.GetCollectionPolicy(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "PolicyNotExist") {
			return object, WrapErrorf(NotFoundErr("CollectionPolicy", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetCollectionPolicy", AlibabaCloudSdkGoERROR)
	}

	// Convert aliyunSlsAPI.CollectionPolicy to map[string]interface{} for compatibility
	result := make(map[string]interface{})
	result["policyName"] = collectionPolicy.PolicyName
	result["policyConfig"] = collectionPolicy.PolicyConfig
	result["productCode"] = collectionPolicy.ProductCode
	result["policyScript"] = collectionPolicy.PolicyScript
	result["resourceDirectory"] = collectionPolicy.ResourceDirectory
	result["enabled"] = collectionPolicy.Enabled
	result["dataCode"] = collectionPolicy.DataCode
	result["dataConfig"] = collectionPolicy.DataConfig
	result["centralizeEnabled"] = collectionPolicy.CentralizeEnabled
	result["centralizeConfig"] = collectionPolicy.CentralizeConfig
	result["createTime"] = collectionPolicy.CreateTime
	result["lastModifyTime"] = collectionPolicy.LastModifyTime

	return result, nil
}

func (s *SlsService) SlsCollectionPolicyStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsCollectionPolicy(id)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		currentStatus := fmt.Sprint(v)

		if field == "#policyName" {
			if currentStatus != "" {
				currentStatus = "#CHECKSET"
			}
		}

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

// DescribeSlsMachineGroup - Get machine group configuration
func (s *SlsService) DescribeSlsMachineGroup(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
		return
	}

	projectName := parts[0]
	machineGroupName := parts[1]

	ctx := context.Background()
	machineGroup, err := s.aliyunSlsAPI.GetMachineGroup(ctx, projectName, machineGroupName)
	if err != nil {
		if strings.Contains(err.Error(), "MachineGroupNotExist") {
			return object, WrapErrorf(NotFoundErr("MachineGroup", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetMachineGroup", AlibabaCloudSdkGoERROR)
	}

	// Convert to map for compatibility
	result := make(map[string]interface{})
	result["name"] = machineGroup.Name
	result["machineIdentifyType"] = machineGroup.MachineIdentifyType
	result["machineList"] = machineGroup.MachineList
	result["attribute"] = machineGroup.Attribute
	result["createTime"] = machineGroup.CreateTime
	result["lastModifyTime"] = machineGroup.LastModifyTime

	return result, nil
}

func (s *SlsService) SlsMachineGroupStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsMachineGroup(id)
		if err != nil {
			if NotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		currentStatus := fmt.Sprint(v)

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

// DescribeSlsLogtailConfig - Get logtail configuration
func (s *SlsService) DescribeSlsLogtailConfig(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 3, len(parts)))
		return
	}

	projectName := parts[0]
	configName := parts[2]

	ctx := context.Background()
	config, err := s.aliyunSlsAPI.GetLogtailConfig(ctx, projectName, configName)
	if err != nil {
		if strings.Contains(err.Error(), "ConfigNotExist") {
			return object, WrapErrorf(NotFoundErr("LogtailConfig", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetLogtailConfig", AlibabaCloudSdkGoERROR)
	}

	// Convert to map for compatibility
	result := make(map[string]interface{})
	result["configName"] = config.ConfigName
	result["logstoreName"] = config.LogstoreName
	result["inputType"] = config.InputType
	result["inputDetail"] = config.InputDetail
	result["outputType"] = config.OutputType
	result["outputDetail"] = config.OutputDetail
	result["createTime"] = config.CreateTime
	result["lastModifyTime"] = config.LastModifyTime

	return result, nil
}

func (s *SlsService) SlsLogtailConfigStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsLogtailConfig(id)
		if err != nil {
			if NotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		currentStatus := fmt.Sprint(v)

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

// DescribeSlsLogtailAttachment - Get logtail attachment configuration
func (s *SlsService) DescribeSlsLogtailAttachment(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 3, len(parts)))
		return
	}

	projectName := parts[0]
	configName := parts[1]
	machineGroupName := parts[2]

	ctx := context.Background()
	attachments, err := s.aliyunSlsAPI.GetLogtailConfigMachineGroups(ctx, projectName, configName)
	if err != nil {
		if strings.Contains(err.Error(), "ConfigNotExist") {
			return object, WrapErrorf(NotFoundErr("LogtailAttachment", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetLogtailConfigMachineGroups", AlibabaCloudSdkGoERROR)
	}

	// Check if the machine group is attached
	attached := false
	for _, groupName := range attachments {
		if groupName == machineGroupName {
			attached = true
			break
		}
	}

	if !attached {
		return object, WrapErrorf(NotFoundErr("LogtailAttachment", id), NotFoundMsg, "")
	}

	// Convert to map for compatibility
	result := make(map[string]interface{})
	result["projectName"] = projectName
	result["configName"] = configName
	result["machineGroupName"] = machineGroupName
	result["attached"] = attached

	return result, nil
}

func (s *SlsService) SlsLogtailAttachmentStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsLogtailAttachment(id)
		if err != nil {
			if NotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		currentStatus := fmt.Sprint(v)

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

// DescribeSlsLogAlertResource - Get log alert resource configuration
func (s *SlsService) DescribeSlsLogAlertResource(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 3, len(parts)))
		return
	}

	resourceType := parts[1]
	resourceId := parts[2]

	result := make(map[string]interface{})

	ctx := context.Background()
	switch resourceType {
	case "user":
		// Check if user alert resource exists
		alertConfig, err := s.aliyunSlsAPI.GetAlertGlobalConfig(ctx)
		if err != nil {
			return result, WrapErrorf(err, DefaultErrorMsg, id, "GetAlertGlobalConfig", AlibabaCloudSdkGoERROR)
		}

		result["type"] = resourceType
		result["project"] = ""
		result["lang"] = resourceId
		result["config"] = alertConfig

	case "project":
		// Check if project alert resource exists
		projectName := resourceId
		_, err := s.aliyunSlsAPI.GetLogStore(ctx, projectName, "internal-alert-history")
		if err != nil {
			if strings.Contains(err.Error(), "LogStoreNotExist") {
				return result, WrapErrorf(NotFoundErr("LogAlertResource", id), NotFoundMsg, "")
			}
			return result, WrapErrorf(err, DefaultErrorMsg, id, "GetLogStore", AlibabaCloudSdkGoERROR)
		}

		result["type"] = resourceType
		result["project"] = projectName
		result["lang"] = ""

	default:
		return result, WrapError(fmt.Errorf("unsupported resource type: %s", resourceType))
	}

	return result, nil
}

// DescribeSlsDashboard - Get dashboard configuration
func (s *SlsService) DescribeSlsDashboard(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
		return
	}

	projectName := parts[0]
	dashboardName := parts[1]

	ctx := context.Background()
	dashboard, err := s.aliyunSlsAPI.GetDashboard(ctx, projectName, dashboardName)
	if err != nil {
		if strings.Contains(err.Error(), "DashboardNotExist") {
			return object, WrapErrorf(NotFoundErr("Dashboard", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetDashboard", AlibabaCloudSdkGoERROR)
	}

	// Convert to map for compatibility
	result := make(map[string]interface{})
	result["dashboardName"] = dashboard.DashboardName
	result["description"] = dashboard.Description
	result["charts"] = dashboard.Charts

	return result, nil
}

func (s *SlsService) SlsDashboardStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsDashboard(id)
		if err != nil {
			if NotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		currentStatus := fmt.Sprint(v)

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

// DescribeSlsOssShipper - Get OSS shipper configuration
func (s *SlsService) DescribeSlsOssShipper(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 3, len(parts)))
		return
	}

	projectName := parts[0]
	logstoreName := parts[1]
	shipperName := parts[2]

	ctx := context.Background()
	shipper, err := s.aliyunSlsAPI.GetOSSShipper(ctx, projectName, logstoreName, shipperName)
	if err != nil {
		if strings.Contains(err.Error(), "ShipperNotExist") {
			return object, WrapErrorf(NotFoundErr("OssShipper", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetOSSShipper", AlibabaCloudSdkGoERROR)
	}

	// Convert to map for compatibility
	result := make(map[string]interface{})
	result["shipperName"] = shipper.ShipperName
	result["targetType"] = shipper.TargetType
	result["targetConfiguration"] = shipper.TargetConfiguration

	return result, nil
}

// DescribeSlsResource - Get resource configuration
func (s *SlsService) DescribeSlsResource(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	ctx := context.Background()
	resource, err := s.aliyunSlsAPI.GetResource(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "ResourceNotExist") {
			return object, WrapErrorf(NotFoundErr("Resource", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetResource", AlibabaCloudSdkGoERROR)
	}

	// Convert to map for compatibility
	result := make(map[string]interface{})
	result["name"] = resource.Name
	result["type"] = resource.Type
	result["schema"] = resource.Schema
	result["description"] = resource.Description

	return result, nil
}

func (s *SlsService) SlsResourceStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsResource(id)
		if err != nil {
			if NotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		currentStatus := fmt.Sprint(v)

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

// DescribeSlsResourceRecord - Get resource record configuration
func (s *SlsService) DescribeSlsResourceRecord(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = WrapError(fmt.Errorf("invalid Resource Id %s. Expected parts' length %d, got %d", id, 2, len(parts)))
		return
	}

	resourceName := parts[0]
	recordId := parts[1]

	ctx := context.Background()
	record, err := s.aliyunSlsAPI.GetResourceRecord(ctx, resourceName, recordId)
	if err != nil {
		if strings.Contains(err.Error(), "ResourceRecordNotExist") {
			return object, WrapErrorf(NotFoundErr("ResourceRecord", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetResourceRecord", AlibabaCloudSdkGoERROR)
	}

	// Convert to map for compatibility
	result := make(map[string]interface{})
	result["id"] = record.Id
	result["tag"] = record.Tag
	result["value"] = record.Value

	return result, nil
}

func (s *SlsService) SlsResourceRecordStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsResourceRecord(id)
		if err != nil {
			if NotFoundError(err) {
				return object, "", nil
			}
			return nil, "", WrapError(err)
		}

		v, err := jsonpath.Get(field, object)
		currentStatus := fmt.Sprint(v)

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}
