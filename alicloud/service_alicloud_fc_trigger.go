package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	aliyunFCAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/fc/v3"
)

// Trigger methods for FCService

// EncodeTriggerResourceId encodes function name and trigger name into a resource ID string
// Format: functionName:triggerName
func EncodeTriggerResourceId(functionName, triggerName string) string {
	return fmt.Sprintf("%s:%s", functionName, triggerName)
}

// DecodeTriggerResourceId decodes trigger resource ID string to function name and trigger name
func DecodeTriggerResourceId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid trigger resource ID format, expected functionName:triggerName, got %s", id)
	}
	return parts[0], parts[1], nil
}

// DescribeFCTriggerByNames retrieves trigger information by function name and trigger name
func (s *FCService) DescribeFCTriggerByNames(functionName, triggerName string) (*aliyunFCAPI.Trigger, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	if triggerName == "" {
		return nil, fmt.Errorf("trigger name cannot be empty")
	}
	return s.GetAPI().GetTrigger(functionName, triggerName)
}

// DescribeFCTrigger retrieves trigger information by resource ID
func (s *FCService) DescribeFCTrigger(id string) (*aliyunFCAPI.Trigger, error) {
	functionName, triggerName, err := DecodeTriggerResourceId(id)
	if err != nil {
		return nil, err
	}
	return s.DescribeFCTriggerByNames(functionName, triggerName)
}

// ListFCTriggers lists all triggers for a function with optional filters
func (s *FCService) ListFCTriggers(functionName string, prefix *string, limit *int32, nextToken *string) ([]*aliyunFCAPI.Trigger, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	result, err := s.GetAPI().ListTriggers(functionName, limit, nextToken, prefix)
	if err != nil {
		return nil, err
	}
	return result.Triggers, nil
}

// CreateFCTrigger creates a new FC trigger
func (s *FCService) CreateFCTrigger(functionName string, trigger *aliyunFCAPI.Trigger) (*aliyunFCAPI.Trigger, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	if trigger == nil {
		return nil, fmt.Errorf("trigger cannot be nil")
	}
	return s.GetAPI().CreateTrigger(functionName, trigger)
}

// UpdateFCTrigger updates an existing FC trigger
func (s *FCService) UpdateFCTrigger(functionName, triggerName string, trigger *aliyunFCAPI.Trigger) (*aliyunFCAPI.Trigger, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	if triggerName == "" {
		return nil, fmt.Errorf("trigger name cannot be empty")
	}
	if trigger == nil {
		return nil, fmt.Errorf("trigger cannot be nil")
	}
	return s.GetAPI().UpdateTrigger(functionName, triggerName, trigger)
}

// DeleteFCTrigger deletes an FC trigger
func (s *FCService) DeleteFCTrigger(functionName, triggerName string) error {
	if functionName == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	if triggerName == "" {
		return fmt.Errorf("trigger name cannot be empty")
	}
	return s.GetAPI().DeleteTrigger(functionName, triggerName)
}

// TriggerStateRefreshFunc returns a StateRefreshFunc to wait for trigger status changes
func (s *FCService) TriggerStateRefreshFunc(functionName, triggerName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeFCTrigger(EncodeTriggerResourceId(functionName, triggerName))
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		currentState := "Active" // FC v3 triggers are typically Active when created
		if object.Status != nil && *object.Status != "" {
			currentState = *object.Status
		}

		for _, failState := range failStates {
			if currentState == failState {
				return object, currentState, WrapError(Error(FailedToReachTargetStatus, currentState))
			}
		}
		return object, currentState, nil
	}
}

// WaitForTriggerCreating waits for trigger creation to complete
func (s *FCService) WaitForTriggerCreating(functionName, triggerName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Creating", "Pending"},
		[]string{"Active"},
		timeout,
		5*time.Second,
		s.TriggerStateRefreshFunc(functionName, triggerName, []string{"Failed", "Error"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, fmt.Sprintf("%s:%s", functionName, triggerName))
}

// WaitForTriggerDeleting waits for trigger deletion to complete
func (s *FCService) WaitForTriggerDeleting(functionName, triggerName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Deleting", "Active"},
		[]string{""},
		timeout,
		5*time.Second,
		s.TriggerStateRefreshFunc(functionName, triggerName, []string{"Failed", "Error"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, fmt.Sprintf("%s:%s", functionName, triggerName))
}

// WaitForTriggerUpdating waits for trigger update to complete
func (s *FCService) WaitForTriggerUpdating(functionName, triggerName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Updating", "Pending"},
		[]string{"Active"},
		timeout,
		5*time.Second,
		s.TriggerStateRefreshFunc(functionName, triggerName, []string{"Failed", "Error"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, fmt.Sprintf("%s:%s", functionName, triggerName))
}

// BuildCreateTriggerInputFromSchema builds Trigger from Terraform schema data
func (s *FCService) BuildCreateTriggerInputFromSchema(d *schema.ResourceData) *aliyunFCAPI.Trigger {
	trigger := &aliyunFCAPI.Trigger{}

	if v, ok := d.GetOk("trigger_name"); ok {
		trigger.TriggerName = tea.String(v.(string))
	}

	if v, ok := d.GetOk("trigger_type"); ok {
		trigger.TriggerType = tea.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		trigger.Description = tea.String(v.(string))
	}

	if v, ok := d.GetOk("qualifier"); ok {
		trigger.Qualifier = tea.String(v.(string))
	}

	if v, ok := d.GetOk("source_arn"); ok {
		trigger.SourceArn = tea.String(v.(string))
	}

	if v, ok := d.GetOk("invocation_role"); ok {
		trigger.InvocationRole = tea.String(v.(string))
	}

	if v, ok := d.GetOk("trigger_config"); ok {
		trigger.TriggerConfig = tea.String(v.(string))
	}

	return trigger
}

// BuildUpdateTriggerInputFromSchema builds Trigger for update from Terraform schema data
func (s *FCService) BuildUpdateTriggerInputFromSchema(d *schema.ResourceData) *aliyunFCAPI.Trigger {
	trigger := &aliyunFCAPI.Trigger{}

	if d.HasChange("description") {
		if v, ok := d.GetOk("description"); ok {
			trigger.Description = tea.String(v.(string))
		}
	}

	if d.HasChange("qualifier") {
		if v, ok := d.GetOk("qualifier"); ok {
			trigger.Qualifier = tea.String(v.(string))
		}
	}

	if d.HasChange("invocation_role") {
		if v, ok := d.GetOk("invocation_role"); ok {
			trigger.InvocationRole = tea.String(v.(string))
		}
	}

	if d.HasChange("trigger_config") {
		if v, ok := d.GetOk("trigger_config"); ok {
			trigger.TriggerConfig = tea.String(v.(string))
		}
	}

	return trigger
}

// SetSchemaFromTrigger sets terraform schema data from Trigger
func (s *FCService) SetSchemaFromTrigger(d *schema.ResourceData, trigger *aliyunFCAPI.Trigger) error {
	if trigger == nil {
		return fmt.Errorf("trigger cannot be nil")
	}

	if trigger.TriggerName != nil {
		d.Set("trigger_name", *trigger.TriggerName)
	}

	if trigger.TriggerType != nil {
		d.Set("trigger_type", *trigger.TriggerType)
	}

	if trigger.Description != nil {
		d.Set("description", *trigger.Description)
	}

	if trigger.Qualifier != nil {
		d.Set("qualifier", *trigger.Qualifier)
	}

	if trigger.SourceArn != nil {
		d.Set("source_arn", *trigger.SourceArn)
	}

	if trigger.TargetArn != nil {
		d.Set("target_arn", *trigger.TargetArn)
	}

	if trigger.InvocationRole != nil {
		d.Set("invocation_role", *trigger.InvocationRole)
	}

	if trigger.Status != nil {
		d.Set("status", *trigger.Status)
	}

	if trigger.TriggerConfig != nil {
		d.Set("trigger_config", *trigger.TriggerConfig)
	}

	if trigger.TriggerId != nil {
		d.Set("trigger_id", *trigger.TriggerId)
	}

	if trigger.CreatedTime != nil {
		d.Set("create_time", *trigger.CreatedTime)
	}

	if trigger.LastModifiedTime != nil {
		d.Set("last_modified_time", *trigger.LastModifiedTime)
	}

	// Set HTTP trigger information if available
	if trigger.HttpTrigger != nil {
		httpTriggerMaps := make([]map[string]interface{}, 0)
		httpTriggerMap := make(map[string]interface{})

		if trigger.HttpTrigger.UrlInternet != nil {
			httpTriggerMap["url_internet"] = *trigger.HttpTrigger.UrlInternet
		}

		if trigger.HttpTrigger.UrlIntranet != nil {
			httpTriggerMap["url_intranet"] = *trigger.HttpTrigger.UrlIntranet
		}

		if len(httpTriggerMap) > 0 {
			httpTriggerMaps = append(httpTriggerMaps, httpTriggerMap)
			d.Set("http_trigger", httpTriggerMaps)
		}
	}

	return nil
}
