// Package alicloud provides resources for Alibaba Cloud products
package alicloud

import (
	"strings"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAliCloudLogProjectLogging() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudLogProjectLoggingCreate,
		Read:   resourceAliCloudLogProjectLoggingRead,
		Update: resourceAliCloudLogProjectLoggingUpdate,
		Delete: resourceAliCloudLogProjectLoggingDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The project name to which the logging configurations belong.",
			},
			"logging_project": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The project to store the service logs.",
			},
			"logging_details": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The type of service log, such as operation_log, consumer_group, etc.",
						},
						"logstore": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The logstore to store the service logs.",
						},
					},
				},
				Description: "Configuration details for the service logging.",
			},
		},
	}
}

// projectLoggingStateRefreshFunc returns a StateRefreshFunc that checks the status of project logging and its dependencies
func projectLoggingStateRefreshFunc(d *schema.ResourceData, meta interface{}, projectName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		client := meta.(*connectivity.AliyunClient)
		slsService, err := NewSlsService(client)
		if err != nil {
			return nil, "", WrapError(err)
		}

		// Check if the main project exists
		_, err = slsService.DescribeLogProject(projectName)
		if err != nil {
			if NotFoundError(err) {
				return nil, "ProjectNotFound", nil
			}
			return nil, "", WrapError(err)
		}

		// Check project logging configuration
		logging, err := slsService.GetProjectLogging(projectName)
		if err != nil {
			if NotFoundError(err) {
				return nil, "LoggingNotFound", nil
			}
			return nil, "", WrapError(err)
		}

		// Check if logging project exists
		_, err = slsService.DescribeLogProject(logging.LoggingProject)
		if err != nil {
			if NotFoundError(err) {
				return nil, "LoggingProjectNotFound", nil
			}
			return nil, "", WrapError(err)
		}

		// Check if all logstores exist
		for _, detail := range logging.LoggingDetails {
			_, err = slsService.DescribeLogStore(logging.LoggingProject, detail.Logstore)
			if err != nil {
				if NotFoundError(err) {
					return nil, "LogstoreNotFound", nil
				}
				return nil, "", WrapError(err)
			}
		}

		return logging, "Available", nil
	}
}

func resourceAliCloudLogProjectLoggingCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	projectName := d.Get("project_name").(string)

	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	logging := createLoggingFromSchema(d)

	// Check if project logging already exists
	existingLogging, err := slsService.GetProjectLogging(projectName)
	if err != nil && !NotFoundError(err) {
		return WrapError(err)
	}

	if existingLogging != nil {
		d.SetId(projectName)
		return resourceAliCloudLogProjectLoggingRead(d, meta)
	}

	// Try to create the project logging
	err = slsService.CreateProjectLogging(projectName, logging)
	if err != nil {
		// Handle cases where the logstore already exists
		if IsExpectedErrors(err, []string{"LogStoreAlreadyExist"}) ||
			(err != nil && (strings.Contains(err.Error(), "LogStoreAlreadyExist") ||
				strings.Contains(err.Error(), "logstore") && strings.Contains(err.Error(), "already exists"))) {

			d.SetId(projectName)

			// Try to update the existing configuration to match desired state
			updateErr := slsService.UpdateProjectLogging(projectName, logging)
			if updateErr != nil {
				// If update also fails due to already exists, just import the existing state
				if IsExpectedErrors(updateErr, []string{"LogStoreAlreadyExist"}) ||
					(updateErr != nil && (strings.Contains(updateErr.Error(), "LogStoreAlreadyExist") ||
						strings.Contains(updateErr.Error(), "logstore") && strings.Contains(updateErr.Error(), "already exists"))) {

					return resourceAliCloudLogProjectLoggingRead(d, meta)
				}

				return WrapErrorf(updateErr, DefaultErrorMsg, "alicloud_log_project_logging", "UpdateProjectLogging", AlibabaCloudSdkGoERROR)
			}
		} else {
			return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_project_logging", "CreateProjectLogging", AlibabaCloudSdkGoERROR)
		}
	}

	d.SetId(projectName)

	// Wait for the project logging to be available and all dependencies to be ready
	stateConf := BuildStateConf([]string{"LoggingNotFound", "LoggingProjectNotFound", "LogstoreNotFound"}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, projectLoggingStateRefreshFunc(d, meta, projectName, []string{"ProjectNotFound"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return resourceAliCloudLogProjectLoggingRead(d, meta)
}

func resourceAliCloudLogProjectLoggingRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	projectName := d.Id()

	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Check if the main project exists
	_, err = slsService.DescribeLogProject(projectName)
	if err != nil {
		if NotFoundError(err) {
			// Don't remove the resource from the state if the main project is not found
			// This can happen during refresh when the main project is temporarily unavailable
			return nil
		}
		return WrapError(err)
	}

	// Check project logging configuration
	logging, err := slsService.GetProjectLogging(projectName)
	if err != nil {
		if NotFoundError(err) {
			// Don't remove the resource from the state if logging config is not found
			// This can happen during refresh when the logging config is temporarily unavailable
			return nil
		}
		return WrapError(err)
	}

	// Check if logging project exists
	_, err = slsService.DescribeLogProject(logging.LoggingProject)
	if err != nil {
		if NotFoundError(err) {
			// Don't remove the resource from the state if the logging project is not found
			// This can happen during refresh when the logging project is temporarily unavailable
			return nil
		}
		return WrapError(err)
	}

	// Check if all logstores exist
	for _, detail := range logging.LoggingDetails {
		_, err = slsService.DescribeLogStore(logging.LoggingProject, detail.Logstore)
		if err != nil {
			if NotFoundError(err) {
				// Don't remove the resource from the state if a logstore is not found
				// This can happen during refresh when a logstore is temporarily unavailable
				return nil
			}
			return WrapError(err)
		}
	}

	d.Set("project_name", projectName)
	d.Set("logging_project", logging.LoggingProject)

	loggingDetailsSet := make([]map[string]interface{}, 0)
	for _, loggingDetail := range logging.LoggingDetails {
		loggingDetailsSet = append(loggingDetailsSet, map[string]interface{}{
			"type":     loggingDetail.Type,
			"logstore": loggingDetail.Logstore,
		})
	}

	if err := d.Set("logging_details", loggingDetailsSet); err != nil {
		return WrapError(err)
	}

	return nil
}

func resourceAliCloudLogProjectLoggingUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	oldProjectName := d.Id()
	newProjectName := d.Get("project_name").(string)

	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	// Determine if any relevant fields changed
	projectNameChanged := d.HasChange("project_name")
	loggingProjectChanged := d.HasChange("logging_project")
	loggingDetailsChanged := d.HasChange("logging_details")

	if !(projectNameChanged || loggingProjectChanged || loggingDetailsChanged) {
		return resourceAliCloudLogProjectLoggingRead(d, meta)
	}

	logging := createLoggingFromSchema(d)

	if projectNameChanged {
		// Create/Update logging on the new project
		// Try Create first, fall back to Update if already exists
		createErr := slsService.CreateProjectLogging(newProjectName, logging)
		if createErr != nil {
			if NotFoundError(createErr) {
				// New project not ready yet; wait and retry via Update
				// Fall through to Update path
			} else if IsExpectedErrors(createErr, []string{"LogStoreAlreadyExist"}) ||
				(strings.Contains(createErr.Error(), "already exists")) {
				// Exists, go to Update
			} else {
				// For any other error, try Update as well; if that fails, return
			}
			if updErr := slsService.UpdateProjectLogging(newProjectName, logging); updErr != nil {
				return WrapErrorf(updErr, DefaultErrorMsg, newProjectName, "UpdateProjectLogging", AlibabaCloudSdkGoERROR)
			}
		}

		// Wait for new project logging to be available
		stateConfNew := BuildStateConf(
			[]string{"LoggingNotFound", "LoggingProjectNotFound", "LogstoreNotFound"},
			[]string{"Available"},
			d.Timeout(schema.TimeoutUpdate),
			5*time.Second,
			projectLoggingStateRefreshFunc(d, meta, newProjectName, []string{"ProjectNotFound"}),
		)
		if _, err := stateConfNew.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, newProjectName)
		}

		// Best-effort cleanup: delete logging from old project
		if oldProjectName != "" && oldProjectName != newProjectName {
			_ = slsService.DeleteProjectLogging(oldProjectName)
		}

		// Update resource ID to new project
		d.SetId(newProjectName)
	} else {
		// Same project, just update logging config or target logging project
		if err := slsService.UpdateProjectLogging(oldProjectName, logging); err != nil {
			return WrapErrorf(err, DefaultErrorMsg, oldProjectName, "UpdateProjectLogging", AlibabaCloudSdkGoERROR)
		}

		// Wait for the updated project logging to be available
		stateConf := BuildStateConf(
			[]string{"LoggingNotFound", "LoggingProjectNotFound", "LogstoreNotFound"},
			[]string{"Available"},
			d.Timeout(schema.TimeoutUpdate),
			5*time.Second,
			projectLoggingStateRefreshFunc(d, meta, oldProjectName, []string{"ProjectNotFound"}),
		)
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	return resourceAliCloudLogProjectLoggingRead(d, meta)
}

func resourceAliCloudLogProjectLoggingDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	projectName := d.Id()

	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapError(err)
	}

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := slsService.DeleteProjectLogging(projectName)
		if err != nil {
			if NeedRetry(err) {
				wait()
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteProjectLogging", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func createLoggingFromSchema(d *schema.ResourceData) *aliyunSlsAPI.LogProjectLogging {
	loggingDetailsSet := d.Get("logging_details").(*schema.Set).List()
	loggingDetails := make([]aliyunSlsAPI.LogProjectLoggingDetails, 0, len(loggingDetailsSet))

	for _, item := range loggingDetailsSet {
		if m, ok := item.(map[string]interface{}); ok {
			loggingDetails = append(loggingDetails, aliyunSlsAPI.LogProjectLoggingDetails{
				Type:     m["type"].(string),
				Logstore: m["logstore"].(string),
			})
		}
	}

	return &aliyunSlsAPI.LogProjectLogging{
		LoggingProject: d.Get("logging_project").(string),
		LoggingDetails: loggingDetails,
	}
}
