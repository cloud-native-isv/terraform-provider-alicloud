package alicloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// DescribeSlsProject <<< Encapsulated get interface for Sls Project.

func (s *SlsService) DescribeSlsProject(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	ctx := context.Background()
	project, err := s.aliyunSlsAPI.GetLogProject(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "ProjectNotExist") {
			return object, WrapErrorf(NotFoundErr("Project", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "GetProject", AlibabaCloudSdkGoERROR)
	}

	// Convert aliyunSlsAPI.LogProject to map[string]interface{} for compatibility
	result := make(map[string]interface{})
	result["projectName"] = project.ProjectName
	result["description"] = project.Description
	result["owner"] = project.Owner
	result["region"] = project.Region
	result["status"] = project.Status
	result["createTime"] = project.CreateTime
	result["lastModifyTime"] = project.LastModifyTime
	result["dataRedundancyType"] = project.DataRedundancyType
	result["resourceGroupId"] = project.ResourceGroupId
	result["recycleBinEnabled"] = project.RecycleBinEnabled

	if project.Quota != nil {
		quota := make(map[string]interface{})
		quota["quota"] = project.Quota.Quota
		result["quota"] = quota
	}

	if len(project.Policy) > 0 {
		result["policy"] = project.Policy
	}

	return result, nil
}

func (s *SlsService) DescribeListTagResources(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	ctx := context.Background()
	tagResources, err := s.aliyunSlsAPI.ListTagResources(ctx, "PROJECT", []string{id}, nil)
	if err != nil {
		if strings.Contains(err.Error(), "ProjectNotExist") {
			return object, WrapErrorf(NotFoundErr("Project", id), NotFoundMsg, "")
		}
		return object, WrapErrorf(err, DefaultErrorMsg, id, "ListTagResources", AlibabaCloudSdkGoERROR)
	}

	// Convert aliyunSlsAPI.TagResource slice to map[string]interface{} for compatibility
	result := make(map[string]interface{})
	tagResourcesMap := make([]map[string]interface{}, 0)

	for _, tagResource := range tagResources {
		tagResourceMap := make(map[string]interface{})
		tagResourceMap["resourceId"] = tagResource.ResourceId
		tagResourceMap["resourceType"] = tagResource.ResourceType
		tagResourceMap["tagKey"] = tagResource.TagKey
		tagResourceMap["tagValue"] = tagResource.TagValue
		tagResourcesMap = append(tagResourcesMap, tagResourceMap)
	}

	result["tagResources"] = tagResourcesMap
	return result, nil
}

func (s *SlsService) SlsProjectStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeSlsProject(id)
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

// SetResourceTags <<< Encapsulated tag function for Sls.
func (s *SlsService) SetResourceTags(d *schema.ResourceData, resourceType string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	if d.HasChange("tags") {
		ctx := context.Background()
		added, removed := parsingTags(d)
		removedTagKeys := make([]string, 0)
		for _, v := range removed {
			if !ignoredTags(v, "") {
				removedTagKeys = append(removedTagKeys, v)
			}
		}

		// Remove tags if any
		if len(removedTagKeys) > 0 {
			err := s.aliyunSlsAPI.UntagResources(ctx, resourceType, []string{d.Id()}, removedTagKeys)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UntagResources", AlibabaCloudSdkGoERROR)
			}
		}

		// Add tags if any
		if len(added) > 0 {
			tags := make(map[string]string)
			for key, value := range added {
				tags[key] = value
			}

			tagRequest := &aliyunSlsAPI.TagResourceRequest{
				ResourceType: resourceType,
				ResourceIds:  []string{d.Id()},
				Tags:         tags,
			}

			err := s.aliyunSlsAPI.TagResources(ctx, tagRequest)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "TagResources", AlibabaCloudSdkGoERROR)
			}
		}

		d.SetPartial("tags")
	}

	return nil
}

// SlsLogging functions using aliyunSlsAPI implementation

func (s *SlsService) CreateSlsLogging(projectName string, logging *aliyunSlsAPI.LogProjectLogging) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	ctx := context.Background()
	err := s.aliyunSlsAPI.CreateLogProjectLogging(ctx, projectName, logging)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, projectName, "CreateLogging", AlibabaCloudSdkGoERROR)
	}
	return nil
}

func (s *SlsService) UpdateSlsLogging(projectName string, logging *aliyunSlsAPI.LogProjectLogging) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	ctx := context.Background()
	err := s.aliyunSlsAPI.UpdateLogProjectLogging(ctx, projectName, logging)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, projectName, "UpdateLogging", AlibabaCloudSdkGoERROR)
	}
	return nil
}

func (s *SlsService) DeleteSlsLogging(projectName string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	ctx := context.Background()
	err := s.aliyunSlsAPI.DeleteLogProjectLogging(ctx, projectName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, projectName, "DeleteLogging", AlibabaCloudSdkGoERROR)
	}
	return nil
}

func (s *SlsService) GetSlsLogging(projectName string) (*aliyunSlsAPI.LogProjectLogging, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	ctx := context.Background()
	logging, err := s.aliyunSlsAPI.GetLogProjectLogging(ctx, projectName)
	if err != nil {
		if strings.Contains(err.Error(), "ProjectNotExist") {
			return nil, WrapErrorf(NotFoundErr("LogProjectLogging", projectName), NotFoundMsg, "")
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, projectName, "GetLogging", AlibabaCloudSdkGoERROR)
	}
	return logging, nil
}

// ======== Project Management Functions ========

// CreateProject creates a new SLS project
func (s *SlsService) CreateProject(request map[string]interface{}) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	ctx := context.Background()
	projectName := request["projectName"].(string)

	projectRequest := &aliyunSlsAPI.CreateLogProjectRequest{
		ProjectName: projectName,
	}

	if description, ok := request["description"]; ok {
		projectRequest.Description = description.(string)
	}

	if resourceGroupId, ok := request["resourceGroupId"]; ok {
		projectRequest.ResourceGroupId = resourceGroupId.(string)
	}

	err := s.aliyunSlsAPI.CreateLogProject(ctx, projectRequest)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, projectName, "CreateProject", AlibabaCloudSdkGoERROR)
	}
	return nil
}

// UpdateProject updates an existing SLS project
func (s *SlsService) UpdateProject(request map[string]interface{}) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	ctx := context.Background()
	projectName := request["projectName"].(string)

	updateRequest := &aliyunSlsAPI.UpdateLogProjectRequest{
		ProjectName: projectName,
	}

	if description, ok := request["description"]; ok {
		updateRequest.Description = description.(string)
	}

	err := s.aliyunSlsAPI.UpdateLogProject(ctx, updateRequest)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, projectName, "UpdateProject", AlibabaCloudSdkGoERROR)
	}
	return nil
}

// DeleteProject deletes an SLS project
func (s *SlsService) DeleteProject(projectName string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	ctx := context.Background()
	err := s.aliyunSlsAPI.DeleteLogProject(ctx, projectName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, projectName, "DeleteProject", AlibabaCloudSdkGoERROR)
	}
	return nil
}

// ChangeResourceGroup changes the resource group of a project
func (s *SlsService) ChangeResourceGroup(request map[string]interface{}) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	ctx := context.Background()
	resourceId := request["resourceId"].(string)
	resourceGroupId := request["resourceGroupId"].(string)
	resourceType := request["resourceType"].(string)

	changeRequest := &aliyunSlsAPI.ChangeResourceGroupRequest{
		ResourceId:      resourceId,
		ResourceGroupId: resourceGroupId,
		ResourceType:    resourceType,
	}

	err := s.aliyunSlsAPI.ChangeResourceGroup(ctx, changeRequest)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, resourceId, "ChangeResourceGroup", AlibabaCloudSdkGoERROR)
	}
	return nil
}

// UpdateProjectPolicy updates the policy of an SLS project
func (s *SlsService) UpdateProjectPolicy(projectName, policy string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	ctx := context.Background()

	if policy == "" {
		// Delete policy if empty
		err := s.aliyunSlsAPI.DeleteLogProjectPolicy(ctx, projectName)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, projectName, "DeleteProjectPolicy", AlibabaCloudSdkGoERROR)
		}
	} else {
		// Update policy
		policyRequest := &aliyunSlsAPI.UpdateLogProjectPolicyRequest{
			ProjectName: projectName,
			Policy:      policy,
		}

		err := s.aliyunSlsAPI.UpdateLogProjectPolicy(ctx, policyRequest)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, projectName, "UpdateProjectPolicy", AlibabaCloudSdkGoERROR)
		}
	}
	return nil
}
