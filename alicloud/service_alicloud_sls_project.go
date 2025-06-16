package alicloud

import (
	"fmt"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// DescribeLogProject <<< Encapsulated get interface for Sls Project.

func (s *SlsService) DescribeLogProject(id string) (*aliyunSlsAPI.LogProject, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	project, err := s.aliyunSlsAPI.GetLogProject(id)
	if err != nil {
		if strings.Contains(err.Error(), "ProjectNotExist") {
			return nil, WrapErrorf(NotFoundErr("Project", id), NotFoundMsg, "")
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "GetProject", AlibabaCloudSdkGoERROR)
	}

	return project, nil
}

func (s *SlsService) DescribeListTagResources(id string) (object map[string]interface{}, err error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	tagResources, err := s.aliyunSlsAPI.ListTagResources("PROJECT", []string{id}, nil)
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

func (s *SlsService) LogProjectStateRefreshFunc(id string, field string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeLogProject(id)
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
		added, removed := parsingTags(d)
		removedTagKeys := make([]string, 0)
		for _, v := range removed {
			if !ignoredTags(v, "") {
				removedTagKeys = append(removedTagKeys, v)
			}
		}

		// Remove tags if any
		if len(removedTagKeys) > 0 {
			err := s.aliyunSlsAPI.UntagResources(resourceType, []string{d.Id()}, removedTagKeys)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UntagResources", AlibabaCloudSdkGoERROR)
			}
		}

		// Add tags if any
		if len(added) > 0 {
			tags := make([]*aliyunSlsAPI.TagResourceRequestTag, 0)
			for key, value := range added {
				if valueStr, ok := value.(string); ok {
					tag := &aliyunSlsAPI.TagResourceRequestTag{
						Key:   &key,
						Value: &valueStr,
					}
					tags = append(tags, tag)
				}
			}

			if len(tags) > 0 {
				tagRequest := &aliyunSlsAPI.TagResourceRequest{
					ResourceType: &resourceType,
					Tags:         tags,
				}

				err := s.aliyunSlsAPI.TagResources(tagRequest)
				if err != nil {
					return WrapErrorf(err, DefaultErrorMsg, d.Id(), "TagResources", AlibabaCloudSdkGoERROR)
				}
			}
		}

		d.SetPartial("tags")
	}

	return nil
}

// ProjectLogging functions using aliyunSlsAPI implementation

func (s *SlsService) CreateProjectLogging(projectName string, logging *aliyunSlsAPI.LogProjectLogging) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	err := s.aliyunSlsAPI.CreateLogProjectLogging(projectName, logging)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, projectName, "CreateLogging", AlibabaCloudSdkGoERROR)
	}
	return nil
}

func (s *SlsService) UpdateProjectLogging(projectName string, logging *aliyunSlsAPI.LogProjectLogging) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	err := s.aliyunSlsAPI.UpdateLogProjectLogging(projectName, logging)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, projectName, "UpdateLogging", AlibabaCloudSdkGoERROR)
	}
	return nil
}

func (s *SlsService) DeleteProjectLogging(projectName string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	err := s.aliyunSlsAPI.DeleteLogProjectLogging(projectName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, projectName, "DeleteLogging", AlibabaCloudSdkGoERROR)
	}
	return nil
}

func (s *SlsService) GetProjectLogging(projectName string) (*aliyunSlsAPI.LogProjectLogging, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	logging, err := s.aliyunSlsAPI.GetLogProjectLogging(projectName)
	if err != nil {
		if strings.Contains(err.Error(), "LoggingNotExist") {
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

	projectName := request["projectName"].(string)
	project := &aliyunSlsAPI.LogProject{
		ProjectName: projectName,
	}

	if description, ok := request["description"]; ok {
		project.Description = description.(string)
	}

	if resourceGroupId, ok := request["resourceGroupId"]; ok {
		project.ResourceGroupId = resourceGroupId.(string)
	}

	err := s.aliyunSlsAPI.CreateLogProject(project)
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

	projectName := request["projectName"].(string)
	project := &aliyunSlsAPI.LogProject{}

	if description, ok := request["description"]; ok {
		project.Description = description.(string)
	}

	err := s.aliyunSlsAPI.UpdateLogProject(projectName, project)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, projectName, "UpdateProject", AlibabaCloudSdkGoERROR)
	}
	return nil
}

// ChangeResourceGroup changes the resource group of an SLS project
func (s *SlsService) ChangeResourceGroup(request map[string]interface{}) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	resourceId := request["resourceId"].(string)
	resourceGroupId := request["resourceGroupId"].(string)

	err := s.aliyunSlsAPI.ChangeResourceGroup(resourceId, resourceGroupId)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, resourceId, "ChangeResourceGroup", AlibabaCloudSdkGoERROR)
	}
	return nil
}

// UpdateProjectPolicy updates the policy of an SLS project
func (s *SlsService) UpdateProjectPolicy(projectName string, policy string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	err := s.aliyunSlsAPI.UpdateProjectPolicy(projectName, policy)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, projectName, "UpdateProjectPolicy", AlibabaCloudSdkGoERROR)
	}
	return nil
}

// DeleteProject deletes an SLS project
func (s *SlsService) DeleteProject(projectName string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	err := s.aliyunSlsAPI.DeleteLogProject(projectName, false)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, projectName, "DeleteProject", AlibabaCloudSdkGoERROR)
	}
	return nil
}

// ListProjects lists SLS projects - alias for ListLogProjects for compatibility
func (s *SlsService) ListProjects() ([]*aliyunSlsAPI.LogProject, error) {
	if s.aliyunSlsAPI == nil {
		return nil, fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	projects, err := s.aliyunSlsAPI.ListLogProjects("", "")
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "", "ListLogProjects", AlibabaCloudSdkGoERROR)
	}

	return projects, nil
}

// DeleteDashboard deletes an SLS dashboard
func (s *SlsService) DeleteDashboard(projectName, dashboardName string) error {
	if s.aliyunSlsAPI == nil {
		return fmt.Errorf("aliyunSlsAPI client is not initialized")
	}

	err := s.aliyunSlsAPI.DeleteDashboard(projectName, dashboardName)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, dashboardName, "DeleteDashboard", AlibabaCloudSdkGoERROR)
	}
	return nil
}
