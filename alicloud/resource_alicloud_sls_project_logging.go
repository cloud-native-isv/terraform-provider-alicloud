// Package alicloud provides resources for Alibaba Cloud products
package alicloud

import (
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunSlsAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/sls"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAlicloudLogProjectLogging() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudLogProjectLoggingCreate,
		Read:   resourceAlicloudLogProjectLoggingRead,
		Update: resourceAlicloudLogProjectLoggingUpdate,
		Delete: resourceAlicloudLogProjectLoggingDelete,
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
				ForceNew:    true,
				Description: "The project name to which the logging configurations belong.",
			},
			"logging_project": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				Description: "The project to store the service logs.",
			},
			"logging_details": {
				Type:     schema.TypeSet,
				Required: true,
				Optional: false,
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
			addDebug("projectLoggingStateRefreshFunc", err, "NewSlsService error for project: "+projectName)
			return nil, "", WrapError(err)
		}

		addDebug("projectLoggingStateRefreshFunc", "Starting state check for project: "+projectName, nil)

		// Check if the main project exists
		_, err = slsService.DescribeLogProject(projectName)
		if err != nil {
			if NotFoundError(err) {
				addDebug("projectLoggingStateRefreshFunc", "Main project not found: "+projectName, err)
				return nil, "ProjectNotFound", nil
			}
			addDebug("projectLoggingStateRefreshFunc", "Error checking main project: "+projectName, err)
			return nil, "", WrapError(err)
		}

		// Check project logging configuration
		logging, err := slsService.GetProjectLogging(projectName)
		if err != nil {
			if NotFoundError(err) {
				addDebug("projectLoggingStateRefreshFunc", "Project logging configuration not found for: "+projectName, err)
				return nil, "LoggingNotFound", nil
			}
			addDebug("projectLoggingStateRefreshFunc", "Error getting project logging for: "+projectName, err)
			return nil, "", WrapError(err)
		}

		// Check if logging project exists
		_, err = slsService.DescribeLogProject(logging.LoggingProject)
		if err != nil {
			if NotFoundError(err) {
				addDebug("projectLoggingStateRefreshFunc", "Logging project not found: "+logging.LoggingProject, err)
				return nil, "LoggingProjectNotFound", nil
			}
			addDebug("projectLoggingStateRefreshFunc", "Error checking logging project: "+logging.LoggingProject, err)
			return nil, "", WrapError(err)
		}

		// Check if all logstores exist
		for _, detail := range logging.LoggingDetails {
			_, err = slsService.DescribeLogStore(logging.LoggingProject, detail.Logstore)
			if err != nil {
				if NotFoundError(err) {
					addDebug("projectLoggingStateRefreshFunc", "Logstore not found: "+detail.Logstore+" in project: "+logging.LoggingProject, err)
					return nil, "LogstoreNotFound", nil
				}
				addDebug("projectLoggingStateRefreshFunc", "Error checking logstore: "+detail.Logstore+" in project: "+logging.LoggingProject, err)
				return nil, "", WrapError(err)
			}
		}

		addDebug("projectLoggingStateRefreshFunc", "All dependencies available for project: "+projectName, nil)
		return logging, "Available", nil
	}
}

func resourceAlicloudLogProjectLoggingCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	projectName := d.Get("project_name").(string)

	addDebug("resourceAlicloudLogProjectLoggingCreate", "Starting create operation for project: "+projectName, nil)

	slsService, err := NewSlsService(client)
	if err != nil {
		addDebug("resourceAlicloudLogProjectLoggingCreate", "Failed to create SLS service for project: "+projectName, err)
		return WrapError(err)
	}

	logging := createLoggingFromSchema(d)
	addDebug("resourceAlicloudLogProjectLoggingCreate", "Created logging configuration from schema for project: "+projectName, logging)

	_, err = slsService.GetProjectLogging(projectName)
	if err != nil {
		if !NotFoundError(err) {
			addDebug("resourceAlicloudLogProjectLoggingCreate", "Error checking existing logging for project: "+projectName, err)
			return WrapError(err)
		}
		addDebug("resourceAlicloudLogProjectLoggingCreate", "Creating new project logging for: "+projectName, nil)
		err = slsService.CreateProjectLogging(projectName, logging)
	} else {
		addDebug("resourceAlicloudLogProjectLoggingCreate", "Updating existing project logging for: "+projectName, nil)
		err = slsService.UpdateProjectLogging(projectName, logging)
	}
	if err != nil {
		addDebug("resourceAlicloudLogProjectLoggingCreate", "Failed to create/update project logging for: "+projectName, err)
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_log_project_logging", "CreateOrUpdateProjectLogging", AlibabaCloudSdkGoERROR)
	}

	d.SetId(projectName)
	addDebug("resourceAlicloudLogProjectLoggingCreate", "Set resource ID to: "+projectName, nil)

	// Wait for the project logging to be available and all dependencies to be ready
	addDebug("resourceAlicloudLogProjectLoggingCreate", "Starting state refresh wait for project: "+projectName, nil)
	stateConf := BuildStateConf([]string{"LoggingNotFound", "LoggingProjectNotFound", "LogstoreNotFound"}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, projectLoggingStateRefreshFunc(d, meta, projectName, []string{"ProjectNotFound"}))
	if _, err := stateConf.WaitForState(); err != nil {
		addDebug("resourceAlicloudLogProjectLoggingCreate", "State refresh wait failed for project: "+projectName, err)
		return WrapErrorf(err, IdMsg, d.Id())
	}

	addDebug("resourceAlicloudLogProjectLoggingCreate", "Create operation completed successfully for project: "+projectName, nil)
	return resourceAlicloudLogProjectLoggingRead(d, meta)
}

func resourceAlicloudLogProjectLoggingRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	projectName := d.Id()

	addDebug("resourceAlicloudLogProjectLoggingRead", "Starting read operation for project: "+projectName, nil)

	slsService, err := NewSlsService(client)
	if err != nil {
		addDebug("resourceAlicloudLogProjectLoggingRead", "Failed to create SLS service for project: "+projectName, err)
		return WrapError(err)
	}

	// Check if the main project exists
	addDebug("resourceAlicloudLogProjectLoggingRead", "Checking main project existence: "+projectName, nil)
	_, err = slsService.DescribeLogProject(projectName)
	if err != nil {
		if NotFoundError(err) {
			addDebug("resourceAlicloudLogProjectLoggingRead", "Main project not found, removing from state: "+projectName, err)
			d.SetId("")
			return nil
		}
		addDebug("resourceAlicloudLogProjectLoggingRead", "Error checking main project: "+projectName, err)
		return WrapError(err)
	}

	// Check project logging configuration
	addDebug("resourceAlicloudLogProjectLoggingRead", "Getting project logging configuration for: "+projectName, nil)
	logging, err := slsService.GetProjectLogging(projectName)
	if err != nil {
		if NotFoundError(err) {
			addDebug("resourceAlicloudLogProjectLoggingRead", "Project logging not found, removing from state: "+projectName, err)
			d.SetId("")
			return nil
		}
		addDebug("resourceAlicloudLogProjectLoggingRead", "Error getting project logging: "+projectName, err)
		return WrapError(err)
	}

	// Check if logging project exists - if not, mark for recreation
	addDebug("resourceAlicloudLogProjectLoggingRead", "Checking logging project existence: "+logging.LoggingProject, nil)
	_, err = slsService.DescribeLogProject(logging.LoggingProject)
	if err != nil {
		if NotFoundError(err) {
			addDebug("resourceAlicloudLogProjectLoggingRead", "Logging project not found, removing from state: "+logging.LoggingProject, err)
			d.SetId("")
			return nil
		}
		addDebug("resourceAlicloudLogProjectLoggingRead", "Error checking logging project: "+logging.LoggingProject, err)
		return WrapError(err)
	}

	// Check if all logstores exist - if any missing, mark for recreation
	addDebug("resourceAlicloudLogProjectLoggingRead", "Checking logstores existence in project: "+logging.LoggingProject, nil)
	for _, detail := range logging.LoggingDetails {
		_, err = slsService.DescribeLogStore(logging.LoggingProject, detail.Logstore)
		if err != nil {
			if NotFoundError(err) {
				addDebug("resourceAlicloudLogProjectLoggingRead", "Logstore not found, removing from state: "+detail.Logstore+" in project: "+logging.LoggingProject, err)
				d.SetId("")
				return nil
			}
			addDebug("resourceAlicloudLogProjectLoggingRead", "Error checking logstore: "+detail.Logstore+" in project: "+logging.LoggingProject, err)
			return WrapError(err)
		}
	}

	addDebug("resourceAlicloudLogProjectLoggingRead", "Setting state attributes for project: "+projectName, nil)
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
		addDebug("resourceAlicloudLogProjectLoggingRead", "Error setting logging_details for project: "+projectName, err)
		return WrapError(err)
	}

	addDebug("resourceAlicloudLogProjectLoggingRead", "Read operation completed successfully for project: "+projectName, nil)
	return nil
}

func resourceAlicloudLogProjectLoggingUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	projectName := d.Id()

	addDebug("resourceAlicloudLogProjectLoggingUpdate", "Starting update operation for project: "+projectName, nil)

	slsService, err := NewSlsService(client)
	if err != nil {
		addDebug("resourceAlicloudLogProjectLoggingUpdate", "Failed to create SLS service for project: "+projectName, err)
		return WrapError(err)
	}

	if d.HasChange("logging_details") {
		addDebug("resourceAlicloudLogProjectLoggingUpdate", "Detected changes in logging_details for project: "+projectName, nil)

		logging := createLoggingFromSchema(d)
		addDebug("resourceAlicloudLogProjectLoggingUpdate", "Created updated logging configuration for project: "+projectName, logging)

		err := slsService.UpdateProjectLogging(projectName, logging)
		if err != nil {
			addDebug("resourceAlicloudLogProjectLoggingUpdate", "Failed to update project logging for: "+projectName, err)
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateProjectLogging", AlibabaCloudSdkGoERROR)
		}

		// Wait for the updated project logging to be available
		addDebug("resourceAlicloudLogProjectLoggingUpdate", "Starting state refresh wait after update for project: "+projectName, nil)
		stateConf := BuildStateConf([]string{"LoggingNotFound", "LoggingProjectNotFound", "LogstoreNotFound"}, []string{"Available"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, projectLoggingStateRefreshFunc(d, meta, projectName, []string{"ProjectNotFound"}))
		if _, err := stateConf.WaitForState(); err != nil {
			addDebug("resourceAlicloudLogProjectLoggingUpdate", "State refresh wait failed after update for project: "+projectName, err)
			return WrapErrorf(err, IdMsg, d.Id())
		}

		addDebug("resourceAlicloudLogProjectLoggingUpdate", "Successfully updated logging configuration for project: "+projectName, nil)
	} else {
		addDebug("resourceAlicloudLogProjectLoggingUpdate", "No changes detected in logging_details for project: "+projectName, nil)
	}

	addDebug("resourceAlicloudLogProjectLoggingUpdate", "Update operation completed for project: "+projectName, nil)
	return resourceAlicloudLogProjectLoggingRead(d, meta)
}

func resourceAlicloudLogProjectLoggingDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	projectName := d.Id()

	addDebug("resourceAlicloudLogProjectLoggingDelete", "Starting delete operation for project: "+projectName, nil)

	slsService, err := NewSlsService(client)
	if err != nil {
		addDebug("resourceAlicloudLogProjectLoggingDelete", "Failed to create SLS service for project: "+projectName, err)
		return WrapError(err)
	}

	wait := incrementalWait(3*time.Second, 5*time.Second)
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		addDebug("resourceAlicloudLogProjectLoggingDelete", "Attempting to delete project logging for: "+projectName, nil)
		err := slsService.DeleteProjectLogging(projectName)
		if err != nil {
			if NeedRetry(err) {
				addDebug("resourceAlicloudLogProjectLoggingDelete", "Retryable error during delete for project: "+projectName, err)
				wait()
				return resource.RetryableError(err)
			}
			addDebug("resourceAlicloudLogProjectLoggingDelete", "Non-retryable error during delete for project: "+projectName, err)
			return resource.NonRetryableError(err)
		}
		addDebug("resourceAlicloudLogProjectLoggingDelete", "Successfully called delete API for project: "+projectName, nil)
		return nil
	})

	if err != nil {
		addDebug("resourceAlicloudLogProjectLoggingDelete", "Failed to delete project logging for: "+projectName, err)
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteProjectLogging", AlibabaCloudSdkGoERROR)
	}

	addDebug("resourceAlicloudLogProjectLoggingDelete", "Delete operation completed successfully for project: "+projectName, nil)
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
