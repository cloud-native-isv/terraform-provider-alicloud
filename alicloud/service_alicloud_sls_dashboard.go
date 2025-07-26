package alicloud

import (
	"fmt"
	"strings"
	"time"

	common "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/common"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DescribeSlsDashboard retrieves a dashboard by project and dashboard name
func (s *SlsService) DescribeSlsDashboard(projectName, dashboardName string) (*aliyunSlsAPI.Dashboard, error) {
	slsAPI, err := s.getSlsAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	dashboard, err := slsAPI.GetDashboard(projectName, dashboardName)
	if err != nil {
		if common.IsNotFoundError(err) {
			return nil, WrapErrorf(NotFoundErr("SlsDashboard", fmt.Sprintf("%s:%s", projectName, dashboardName)), NotFoundMsg, ProviderERROR)
		}
		return nil, WrapErrorf(err, DefaultErrorMsg, fmt.Sprintf("%s:%s", projectName, dashboardName), "GetDashboard", AlibabaCloudSdkGoERROR)
	}

	return dashboard, nil
}

// CreateSlsDashboard creates a new dashboard in the specified project
func (s *SlsService) CreateSlsDashboard(projectName string, dashboard *aliyunSlsAPI.Dashboard) error {
	slsAPI, err := s.getSlsAPI()
	if err != nil {
		return WrapError(err)
	}

	err = slsAPI.CreateDashboard(projectName, dashboard)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, dashboard.Name, "CreateDashboard", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// UpdateSlsDashboard updates an existing dashboard in the specified project
func (s *SlsService) UpdateSlsDashboard(projectName, dashboardName string, dashboard *aliyunSlsAPI.Dashboard) error {
	slsAPI, err := s.getSlsAPI()
	if err != nil {
		return WrapError(err)
	}

	err = slsAPI.UpdateDashboard(projectName, dashboardName, dashboard)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, dashboardName, "UpdateDashboard", AlibabaCloudSdkGoERROR)
	}

	return nil
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

// DeleteSlsDashboard deletes a dashboard from the specified project
func (s *SlsService) DeleteSlsDashboard(projectName, dashboardName string) error {
	slsAPI, err := s.getSlsAPI()
	if err != nil {
		return WrapError(err)
	}

	err = slsAPI.DeleteDashboard(projectName, dashboardName)
	if err != nil {
		if common.IsNotFoundError(err) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, dashboardName, "DeleteDashboard", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// ListSlsDashboards lists dashboards in the specified project with pagination
func (s *SlsService) ListSlsDashboards(projectName string, offset, size int32) ([]*aliyunSlsAPI.Dashboard, error) {
	slsAPI, err := s.getSlsAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	dashboards, err := slsAPI.ListDashboard(projectName, offset, size)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, projectName, "ListDashboard", AlibabaCloudSdkGoERROR)
	}

	return dashboards, nil
}

// ListAllSlsDashboards lists all dashboards in the specified project (handles pagination automatically)
func (s *SlsService) ListAllSlsDashboards(projectName string) ([]*aliyunSlsAPI.Dashboard, error) {
	slsAPI, err := s.getSlsAPI()
	if err != nil {
		return nil, WrapError(err)
	}

	dashboards, err := slsAPI.ListAllDashboards(projectName)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, projectName, "ListAllDashboards", AlibabaCloudSdkGoERROR)
	}

	return dashboards, nil
}

// SlsDashboardExists checks if a dashboard exists in the specified project
func (s *SlsService) SlsDashboardExists(projectName, dashboardName string) (bool, error) {
	slsAPI, err := s.getSlsAPI()
	if err != nil {
		return false, WrapError(err)
	}

	exists, err := slsAPI.DashboardExists(projectName, dashboardName)
	if err != nil {
		return false, WrapErrorf(err, DefaultErrorMsg, fmt.Sprintf("%s:%s", projectName, dashboardName), "DashboardExists", AlibabaCloudSdkGoERROR)
	}

	return exists, nil
}

// SlsDashboardStateRefreshFunc returns a StateRefreshFunc that monitors dashboard state changes
func (s *SlsService) SlsDashboardStateRefreshFunc(projectName, dashboardName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		dashboard, err := s.DescribeSlsDashboard(projectName, dashboardName)
		if err != nil {
			if IsNotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		for _, failState := range failStates {
			if dashboard.Name == failState {
				return dashboard, failState, WrapError(Error(FailedToReachTargetStatus, failState))
			}
		}

		return dashboard, "Available", nil
	}
}

// WaitForSlsDashboard waits for a dashboard to reach the target state
func (s *SlsService) WaitForSlsDashboard(id string, status Status, timeout time.Duration) error {
	// Parse the ID to get project name and dashboard name
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return WrapErrorf(Error("Invalid SLS Dashboard ID format"), IdMsg, id)
	}
	projectName, dashboardName := parts[0], parts[1]

	// Handle different target states
	var targets []string
	var pending []string
	var failStates []string

	switch status {
	case Deleted:
		// For deletion, we expect the resource to not be found
		targets = []string{}
		pending = []string{"Available"}
		failStates = []string{}
	case Available:
		targets = []string{"Available"}
		pending = []string{"Creating", "Updating"}
		failStates = []string{"Failed"}
	default:
		targets = []string{string(status)}
		pending = []string{"Available"}
		failStates = []string{"Failed"}
	}

	stateConf := BuildStateConf(pending, targets, timeout, 5*time.Second, s.SlsDashboardStateRefreshFunc(projectName, dashboardName, failStates))
	stateConf.Delay = 5 * time.Second
	stateConf.MinTimeout = 3 * time.Second

	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, id)
	}

	return nil
}

// Helper functions for converting between Terraform and SLS API types

// ConvertToSlsChart converts a Terraform chart configuration to SLS Chart
func ConvertToSlsChart(terraformChart map[string]interface{}) *aliyunSlsAPI.Chart {
	chart := &aliyunSlsAPI.Chart{}

	if title, ok := terraformChart["title"].(string); ok {
		chart.Title = title
	}

	if chartType, ok := terraformChart["type"].(string); ok {
		chart.Type = chartType
	}

	// Convert search configuration
	if searchData, ok := terraformChart["search"].([]interface{}); ok && len(searchData) > 0 {
		if searchMap, ok := searchData[0].(map[string]interface{}); ok {
			chart.Search = &aliyunSlsAPI.ChartSearch{}

			if logstore, ok := searchMap["logstore"].(string); ok {
				chart.Search.Logstore = logstore
			}
			if topic, ok := searchMap["topic"].(string); ok {
				chart.Search.Topic = topic
			}
			if query, ok := searchMap["query"].(string); ok {
				chart.Search.Query = query
			}
			if start, ok := searchMap["start"].(string); ok {
				chart.Search.Start = start
			}
			if end, ok := searchMap["end"].(string); ok {
				chart.Search.End = end
			}
			if timeSpanType, ok := searchMap["time_span_type"].(string); ok {
				chart.Search.TimeSpanType = timeSpanType
			}
		}
	}

	// Convert display configuration
	if displayData, ok := terraformChart["display"].([]interface{}); ok && len(displayData) > 0 {
		if displayMap, ok := displayData[0].(map[string]interface{}); ok {
			chart.Display = &aliyunSlsAPI.ChartDisplay{}

			if xPos, ok := displayMap["x_pos"].(int); ok {
				chart.Display.XPos = xPos
			}
			if yPos, ok := displayMap["y_pos"].(int); ok {
				chart.Display.YPos = yPos
			}
			if width, ok := displayMap["width"].(int); ok {
				chart.Display.Width = width
			}
			if height, ok := displayMap["height"].(int); ok {
				chart.Display.Height = height
			}
			if displayName, ok := displayMap["display_name"].(string); ok {
				chart.Display.DisplayName = displayName
			}
		}
	}

	return chart
}

// ConvertFromSlsChart converts an SLS Chart to Terraform chart configuration
func ConvertFromSlsChart(chart *aliyunSlsAPI.Chart) map[string]interface{} {
	result := make(map[string]interface{})

	if chart.Title != "" {
		result["title"] = chart.Title
	}
	if chart.Type != "" {
		result["type"] = chart.Type
	}

	// Convert search configuration
	if chart.Search != nil {
		search := make(map[string]interface{})
		if chart.Search.Logstore != "" {
			search["logstore"] = chart.Search.Logstore
		}
		if chart.Search.Topic != "" {
			search["topic"] = chart.Search.Topic
		}
		if chart.Search.Query != "" {
			search["query"] = chart.Search.Query
		}
		if chart.Search.Start != "" {
			search["start"] = chart.Search.Start
		}
		if chart.Search.End != "" {
			search["end"] = chart.Search.End
		}
		if chart.Search.TimeSpanType != "" {
			search["time_span_type"] = chart.Search.TimeSpanType
		}
		result["search"] = []interface{}{search}
	}

	// Convert display configuration
	if chart.Display != nil {
		display := make(map[string]interface{})
		display["x_pos"] = chart.Display.XPos
		display["y_pos"] = chart.Display.YPos
		display["width"] = chart.Display.Width
		display["height"] = chart.Display.Height
		if chart.Display.DisplayName != "" {
			display["display_name"] = chart.Display.DisplayName
		}
		result["display"] = []interface{}{display}
	}

	return result
}

// ConvertToSlsDashboard converts Terraform dashboard configuration to SLS Dashboard
func ConvertToSlsDashboard(d map[string]interface{}) *aliyunSlsAPI.Dashboard {
	dashboard := &aliyunSlsAPI.Dashboard{}

	if name, ok := d["dashboard_name"].(string); ok {
		dashboard.Name = name
	}
	if displayName, ok := d["display_name"].(string); ok {
		dashboard.DisplayName = displayName
	}
	if description, ok := d["description"].(string); ok {
		dashboard.Description = description
	}

	// Convert attributes
	if attributes, ok := d["attribute"].(map[string]interface{}); ok {
		dashboard.Attribute = make(map[string]string)
		for k, v := range attributes {
			if strVal, ok := v.(string); ok {
				dashboard.Attribute[k] = strVal
			}
		}
	}

	// Convert charts - Fixed typo from "char_list" to "chart_list"
	if chartsData, ok := d["chart_list"].([]interface{}); ok {
		dashboard.ChartList = make([]*aliyunSlsAPI.Chart, len(chartsData))
		for i, chartData := range chartsData {
			if chartMap, ok := chartData.(map[string]interface{}); ok {
				dashboard.ChartList[i] = ConvertToSlsChart(chartMap)
			}
		}
	}

	return dashboard
}

// ConvertFromSlsDashboard converts an SLS Dashboard to Terraform dashboard configuration
func ConvertFromSlsDashboard(dashboard *aliyunSlsAPI.Dashboard) map[string]interface{} {
	result := make(map[string]interface{})

	if dashboard.Name != "" {
		result["dashboard_name"] = dashboard.Name
	}
	if dashboard.DisplayName != "" {
		result["display_name"] = dashboard.DisplayName
	}
	if dashboard.Description != "" {
		result["description"] = dashboard.Description
	}

	// Convert attributes
	if dashboard.Attribute != nil && len(dashboard.Attribute) > 0 {
		result["attribute"] = dashboard.Attribute
	}

	// Convert charts - Fixed typo from "char_list" to "chart_list"
	if dashboard.ChartList != nil && len(dashboard.ChartList) > 0 {
		charts := make([]interface{}, len(dashboard.ChartList))
		for i, chart := range dashboard.ChartList {
			charts[i] = ConvertFromSlsChart(chart)
		}
		result["chart_list"] = charts
	}

	return result
}

// GetSlsDashboardName extracts dashboard name from a compound ID
func GetSlsDashboardName(id string) (projectName, dashboardName string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid dashboard ID format, expected project:dashboard, got: %s", id)
	}
	return parts[0], parts[1], nil
}

// BuildSlsDashboardId constructs a compound ID from project and dashboard names
func BuildSlsDashboardId(projectName, dashboardName string) string {
	return fmt.Sprintf("%s:%s", projectName, dashboardName)
}
