package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	fc "github.com/alibabacloud-go/fc-20230330/v4/client"
	aliyunFCAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/fc/v3"
)

// Function methods for FCService

// EncodeFunctionId encodes function name into an ID string
func EncodeFunctionId(functionName string) string {
	return functionName
}

// DecodeFunctionId decodes function ID string to function name
func DecodeFunctionId(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("invalid function ID format, cannot be empty")
	}
	return id, nil
}

// DescribeFCFunction retrieves function information by name
func (s *FCService) DescribeFCFunction(functionName string) (*aliyunFCAPI.Function, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	return s.GetAPI().GetFunction(functionName, nil)
}

// ListFCFunctions lists all functions with optional filters
func (s *FCService) ListFCFunctions(prefix *string, limit *int32, nextToken *string) ([]*aliyunFCAPI.Function, error) {
	request := &fc.ListFunctionsRequest{
		Prefix:    prefix,
		Limit:     limit,
		NextToken: nextToken,
	}
	return s.GetAPI().ListFunctions(request)
}

// CreateFCFunction creates a new FC function
func (s *FCService) CreateFCFunction(function *aliyunFCAPI.Function) (*aliyunFCAPI.Function, error) {
	if function == nil {
		return nil, fmt.Errorf("function cannot be nil")
	}
	return s.GetAPI().CreateFunction(function)
}

// UpdateFCFunction updates an existing FC function
func (s *FCService) UpdateFCFunction(functionName string, function *aliyunFCAPI.Function) (*aliyunFCAPI.Function, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	if function == nil {
		return nil, fmt.Errorf("function cannot be nil")
	}
	return s.GetAPI().UpdateFunction(functionName, function)
}

// DeleteFCFunction deletes an FC function
func (s *FCService) DeleteFCFunction(functionName string) error {
	if functionName == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	return s.GetAPI().DeleteFunction(functionName)
}

// BuildCreateFunctionInputFromSchema builds Function from Terraform schema data
func (s *FCService) BuildCreateFunctionInputFromSchema(d *schema.ResourceData) *aliyunFCAPI.Function {
	function := &aliyunFCAPI.Function{}

	if v, ok := d.GetOk("function_name"); ok {
		function.FunctionName = tea.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		function.Description = tea.String(v.(string))
	}

	if v, ok := d.GetOk("runtime"); ok {
		function.Runtime = tea.String(v.(string))
	}

	if v, ok := d.GetOk("handler"); ok {
		function.Handler = tea.String(v.(string))
	}

	if v, ok := d.GetOk("timeout"); ok {
		function.Timeout = tea.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("memory_size"); ok {
		function.MemorySize = tea.Int32(int32(v.(int)))
	}

	// Note: Code handling would be done at API level, not exposed in Function
	// The API layer should handle code zip file, OSS bucket, and OSS key conversion

	// Add environment variables
	if v, ok := d.GetOk("environment_variables"); ok {
		envVars := v.(map[string]interface{})
		if len(envVars) > 0 {
			function.Environment = make(map[string]*string)
			for k, val := range envVars {
				function.Environment[k] = tea.String(val.(string))
			}
		}
	}

	return function
}

// BuildUpdateFunctionInputFromSchema builds Function for update from Terraform schema data
func (s *FCService) BuildUpdateFunctionInputFromSchema(d *schema.ResourceData) *aliyunFCAPI.Function {
	function := &aliyunFCAPI.Function{}

	if d.HasChange("description") {
		function.Description = tea.String(d.Get("description").(string))
	}

	if d.HasChange("handler") {
		function.Handler = tea.String(d.Get("handler").(string))
	}

	if d.HasChange("timeout") {
		function.Timeout = tea.Int32(int32(d.Get("timeout").(int)))
	}

	if d.HasChange("memory_size") {
		function.MemorySize = tea.Int32(int32(d.Get("memory_size").(int)))
	}

	if d.HasChange("environment_variables") {
		if v, ok := d.GetOk("environment_variables"); ok {
			envVars := v.(map[string]interface{})
			if len(envVars) > 0 {
				function.Environment = make(map[string]*string)
				for k, val := range envVars {
					function.Environment[k] = tea.String(val.(string))
				}
			}
		} else {
			function.Environment = make(map[string]*string)
		}
	}

	// Handle code changes - Note: Code handling would be done at API level
	// The API layer should handle code zip file, OSS bucket, and OSS key conversion
	// if d.HasChange("code_zip_file") || d.HasChange("oss_bucket") || d.HasChange("oss_key") {
	//     // Code handling would be managed by API layer
	// }

	return function
} // SetSchemaFromFunction sets terraform schema data from Function
func (s *FCService) SetSchemaFromFunction(d *schema.ResourceData, function *aliyunFCAPI.Function) error {
	if function == nil {
		return fmt.Errorf("function cannot be nil")
	}

	if function.FunctionName != nil {
		d.Set("function_name", *function.FunctionName)
	}

	if function.Description != nil {
		d.Set("description", *function.Description)
	}

	if function.Runtime != nil {
		d.Set("runtime", *function.Runtime)
	}

	if function.Handler != nil {
		d.Set("handler", *function.Handler)
	}

	if function.Timeout != nil {
		d.Set("timeout", *function.Timeout)
	}

	if function.MemorySize != nil {
		d.Set("memory_size", *function.MemorySize)
	}

	if function.Environment != nil {
		envVars := make(map[string]string)
		for k, v := range function.Environment {
			if v != nil {
				envVars[k] = *v
			}
		}
		d.Set("environment_variables", envVars)
	}

	if function.CreatedTime != nil {
		d.Set("creation_time", *function.CreatedTime)
	}

	if function.LastModifiedTime != nil {
		d.Set("last_modification_time", *function.LastModifiedTime)
	}

	return nil
}

// FCFunctionStateRefreshFunc returns a StateRefreshFunc for FC function operations
func (s *FCService) FCFunctionStateRefreshFunc(functionName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeFCFunction(functionName)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		currentStatus := "Active" // FC functions don't have explicit status, assume Active if retrievable
		if object.State != nil {
			currentStatus = *object.State
		}

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

// WaitForFCFunctionCreating waits for function creation to complete
func (s *FCService) WaitForFCFunctionCreating(functionName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Creating", "Pending"},
		[]string{"Active"},
		timeout,
		5*time.Second,
		s.FCFunctionStateRefreshFunc(functionName, []string{"Failed"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, functionName)
}

// WaitForFCFunctionUpdating waits for function update to complete
func (s *FCService) WaitForFCFunctionUpdating(functionName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Updating"},
		[]string{"Active"},
		timeout,
		5*time.Second,
		s.FCFunctionStateRefreshFunc(functionName, []string{"Failed"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, functionName)
}

// WaitForFCFunctionDeleting waits for function deletion to complete
func (s *FCService) WaitForFCFunctionDeleting(functionName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"Deleting"},
		Target:  []string{""},
		Refresh: func() (interface{}, string, error) {
			obj, err := s.DescribeFCFunction(functionName)
			if err != nil {
				if NotFoundError(err) {
					return nil, "", nil
				}
				return nil, "", WrapError(err)
			}
			return obj, "Deleting", nil
		},
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, functionName)
}

// EncodeFunctionVersionId encodes function name and version into a resource ID string
// Format: functionName:versionId
func EncodeFunctionVersionId(functionName, versionId string) string {
	return fmt.Sprintf("%s:%s", functionName, versionId)
}

// DecodeFunctionVersionId decodes function version ID string to function name and version
func DecodeFunctionVersionId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid function version ID format, expected functionName:versionId, got %s", id)
	}
	return parts[0], parts[1], nil
}

// DescribeFCFunctionVersion retrieves function version information by ID
func (s *FCService) DescribeFCFunctionVersion(id string) (*aliyunFCAPI.Function, error) {
	functionName, versionId, err := DecodeFunctionVersionId(id)
	if err != nil {
		return nil, err
	}

	// Use GetFunction with version qualifier to get version-specific information
	return s.GetAPI().GetFunction(functionName, &versionId)
}

// FunctionStateRefreshFunc returns a StateRefreshFunc to wait for function status changes
func (s *FCService) FunctionStateRefreshFunc(functionName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeFCFunction(functionName)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		currentState := "Active" // FC v3 functions are typically Active when created
		if object.State != nil && *object.State != "" {
			currentState = *object.State
		}

		for _, failState := range failStates {
			if currentState == failState {
				return object, currentState, WrapError(Error(FailedToReachTargetStatus, currentState))
			}
		}
		return object, currentState, nil
	}
}

// WaitForFunctionCreating waits for function creation to complete
func (s *FCService) WaitForFunctionCreating(functionName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Creating", "Pending"},
		[]string{"Active"},
		timeout,
		5*time.Second,
		s.FunctionStateRefreshFunc(functionName, []string{"Failed", "Error"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, functionName)
}

// WaitForFunctionDeleting waits for function deletion to complete
func (s *FCService) WaitForFunctionDeleting(functionName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Deleting", "Active"},
		[]string{""},
		timeout,
		5*time.Second,
		s.FunctionStateRefreshFunc(functionName, []string{"Failed", "Error"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, functionName)
}

// WaitForFunctionUpdating waits for function update to complete
func (s *FCService) WaitForFunctionUpdating(functionName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Updating", "Pending"},
		[]string{"Active"},
		timeout,
		5*time.Second,
		s.FunctionStateRefreshFunc(functionName, []string{"Failed", "Error"}),
	)

	_, err := stateConf.WaitForState()
	return WrapErrorf(err, IdMsg, functionName)
}
