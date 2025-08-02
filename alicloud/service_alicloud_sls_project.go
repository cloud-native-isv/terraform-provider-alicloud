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
	project, err := s.GetAPI().GetLogProject(id)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, id, "GetProject", AlibabaCloudSdkGoERROR)
	}

	return project, nil
}

func (s *SlsService) DescribeListTagResources(id string) (object map[string]interface{}, err error) {
	tagResources, err := s.GetAPI().ListTagResources("PROJECT", []string{id}, nil)
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
			if IsNotFoundError(err) {
				return nil, "deleted", nil
			}
			return nil, "", WrapError(err)
		}

		// If project exists, it's in Normal status
		if object != nil {
			// Convert to map for jsonpath compatibility
			projectMap := make(map[string]interface{})
			projectMap["projectName"] = object.ProjectName
			projectMap["description"] = object.Description
			projectMap["resourceGroupId"] = object.ResourceGroupId
			projectMap["status"] = object.Status
			projectMap["owner"] = object.Owner
			projectMap["region"] = object.Region
			projectMap["location"] = object.Location
			projectMap["createTime"] = object.CreateTime
			projectMap["lastModifyTime"] = object.LastModifyTime
			projectMap["dataRedundancyType"] = object.DataRedundancyType
			projectMap["recycleBinEnabled"] = object.RecycleBinEnabled
			projectMap["quota"] = object.Quota

			// Try to get the requested field, fallback to "Normal" if field doesn't exist
			var currentStatus string
			if field != "" {
				if v, err := jsonpath.Get(field, projectMap); err == nil {
					currentStatus = fmt.Sprint(v)
				} else {
					// If the field doesn't exist or there's an error, default to "Normal"
					currentStatus = "Normal"
				}
			} else {
				currentStatus = "Normal"
			}

			// Handle empty status - default to Normal if object exists
			if currentStatus == "" || currentStatus == "<nil>" {
				currentStatus = "Normal"
			}

			for _, failState := range failStates {
				if currentStatus == failState {
					return projectMap, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
				}
			}

			return projectMap, currentStatus, nil
		}

		// This should not happen, but handle it gracefully
		return nil, "unknown", nil
	}
}

// SetResourceTags <<< Encapsulated tag function for Sls.
func (s *SlsService) SetResourceTags(d *schema.ResourceData, resourceType string) error {
	return nil
}

// ProjectLogging functions using aliyunSlsAPI implementation

// CreateProjectIfNotExist creates a project if it does not exist
func (s *SlsService) CreateProjectIfNotExist(project *aliyunSlsAPI.LogProject) (*aliyunSlsAPI.LogProject, error) {
	if project == nil {
		return nil, fmt.Errorf("project parameter cannot be nil")
	}

	if project.ProjectName == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}

	// Check if project exists
	_, err := s.DescribeLogProject(project.ProjectName)
	if err != nil {
		if IsNotFoundError(err) {
			// Project doesn't exist, create it with provided configuration
			if err := s.CreateProject(project); err != nil {
				return nil, WrapErrorf(err, DefaultErrorMsg, project.ProjectName, "CreateProject", AlibabaCloudSdkGoERROR)
			}
			// Return the created project
			return project, nil
		} else {
			// Other error occurred
			return nil, WrapErrorf(err, DefaultErrorMsg, project.ProjectName, "DescribeLogProject", AlibabaCloudSdkGoERROR)
		}
	}
	// Project already exists, return nil, nil
	return nil, nil
}

// ======== Project Management Functions ========

// CreateProject creates a new SLS project
func (s *SlsService) CreateProject(project *aliyunSlsAPI.LogProject) error {
	err := s.GetAPI().CreateLogProject(project)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, project.ProjectName, "CreateProject", AlibabaCloudSdkGoERROR)
	}
	return nil
}

// UpdateProject updates an existing SLS project
func (s *SlsService) UpdateProject(request map[string]interface{}) error {
	projectName := request["projectName"].(string)
	project := &aliyunSlsAPI.LogProject{}

	if description, ok := request["description"]; ok {
		project.Description = description.(string)
	}

	err := s.GetAPI().UpdateLogProject(projectName, project)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, projectName, "UpdateProject", AlibabaCloudSdkGoERROR)
	}
	return nil
}

// ChangeResourceGroup changes the resource group of an SLS project
func (s *SlsService) ChangeResourceGroup(request map[string]interface{}) error {
	resourceId := request["resourceId"].(string)
	resourceGroupId := request["resourceGroupId"].(string)

	err := s.GetAPI().ChangeResourceGroup(resourceId, resourceGroupId)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, resourceId, "ChangeResourceGroup", AlibabaCloudSdkGoERROR)
	}
	return nil
}

// UpdateProjectPolicy updates the policy of an SLS project
func (s *SlsService) UpdateProjectPolicy(projectName string, policy string) error {
	err := s.GetAPI().UpdateProjectPolicy(projectName, policy)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, projectName, "UpdateProjectPolicy", AlibabaCloudSdkGoERROR)
	}
	return nil
}

// DeleteProject deletes an SLS project
func (s *SlsService) DeleteProject(projectName string) error {
	err := s.GetAPI().DeleteLogProject(projectName, false)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, projectName, "DeleteProject", AlibabaCloudSdkGoERROR)
	}
	return nil
}

// ListProjects lists SLS projects - alias for ListLogProjects for compatibility
func (s *SlsService) ListProjects() ([]*aliyunSlsAPI.LogProject, error) {
	projects, err := s.GetAPI().ListLogProjects("", "")
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "", "ListLogProjects", AlibabaCloudSdkGoERROR)
	}

	return projects, nil
}
