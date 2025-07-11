package alicloud

import (
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	flinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudFlinkSessionCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkSessionClusterCreate,
		Read:   resourceAliCloudFlinkSessionClusterRead,
		Update: resourceAliCloudFlinkSessionClusterUpdate,
		Delete: resourceAliCloudFlinkSessionClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"namespace_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"engine_version": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"deployment_target_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"basic_resource_setting": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"parallelism": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"jobmanager_resource_setting_spec": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cpu": {
										Type:         schema.TypeFloat,
										Optional:     true,
										ValidateFunc: validation.FloatAtLeast(0.1),
									},
									"memory": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
						"taskmanager_resource_setting_spec": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cpu": {
										Type:         schema.TypeFloat,
										Optional:     true,
										ValidateFunc: validation.FloatAtLeast(0.1),
									},
									"memory": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
					},
				},
			},
			"user_flink_conf": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"logging": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"logging_profile": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Logging profile.",
						},
						"log4j2_configuration_template": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"log4j_loggers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"logger_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"logger_level": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"}, false),
									},
								},
							},
						},
						"log_reserve_policy": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"expiration_days": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"open_history": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
						},
					},
				},
			},
			// Status parameter for controlling session cluster lifecycle
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"STOPPED",
					"RUNNING",
					"FAILED",
				}, false),
				Description: "Target status of the session cluster. Valid values: STOPPED, RUNNING, FAILED. Other statuses (STARTING, UPDATING, STOPPING) are intermediate states managed by the system.",
			},
			// Computed attributes
			"session_cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modified_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creator": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creator_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modifier_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAliCloudFlinkSessionClusterCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)
	sessionClusterName := d.Get("name").(string)

	sessionCluster := &flinkAPI.SessionCluster{
		Name:                 sessionClusterName,
		Namespace:            namespaceName,
		Workspace:            workspaceId,
		EngineVersion:        d.Get("engine_version").(string),
		DeploymentTargetName: d.Get("deployment_target_name").(string),
	}

	// Set basic resource setting
	if v, ok := d.GetOk("basic_resource_setting"); ok {
		sessionCluster.BasicResourceSetting = expandBasicResourceSetting(v.([]interface{}))
	}

	// Set flink configuration
	if v, ok := d.GetOk("user_flink_conf"); ok {
		sessionCluster.FlinkConf = expandFlinkConf(v.(map[string]interface{}))
	}

	// Set labels
	if v, ok := d.GetOk("labels"); ok {
		sessionCluster.Labels = expandLabels(v.(map[string]interface{}))
	}

	// Set logging configuration
	if v, ok := d.GetOk("logging"); ok {
		sessionCluster.Logging = expandLogging(v.([]interface{}))
	}

	_, err = flinkService.CreateSessionCluster(workspaceId, namespaceName, sessionCluster)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_session_cluster", "CreateSessionCluster", AlibabaCloudSdkGoERROR)
	}

	d.SetId(formatSessionClusterId(workspaceId, namespaceName, sessionClusterName))

	// Wait for session cluster to be ready
	stateConf := BuildStateConf([]string{}, []string{"RUNNING", "STOPPED"}, d.Timeout(schema.TimeoutCreate), 30*time.Second, flinkService.SessionClusterStateRefreshFunc(d.Id(), []string{"FAILED", "TERMINATED"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// Handle status parameter - start/stop cluster if needed
	if targetStatus, ok := d.GetOk("status"); ok {
		err := handleSessionClusterStatusChange(d, flinkService, workspaceId, namespaceName, sessionClusterName, targetStatus.(string))
		if err != nil {
			return err
		}
	}

	return resourceAliCloudFlinkSessionClusterRead(d, meta)
}

func resourceAliCloudFlinkSessionClusterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	object, err := flinkService.DescribeSessionCluster(d.Id())
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	workspaceId, namespaceName, sessionClusterName, err := parseSessionClusterId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	d.Set("workspace_id", workspaceId)
	d.Set("namespace_name", namespaceName)
	d.Set("name", sessionClusterName)
	d.Set("session_cluster_id", object.SessionClusterId)
	d.Set("engine_version", object.EngineVersion)
	d.Set("deployment_target_name", object.DeploymentTargetName)
	d.Set("creator", object.Creator)
	d.Set("creator_name", object.CreatorName)
	d.Set("modifier", object.Modifier)
	d.Set("modifier_name", object.ModifierName)

	if object.CreatedAt != 0 {
		d.Set("created_at", formatTimestamp(object.CreatedAt))
	}
	if object.ModifiedAt != 0 {
		d.Set("modified_at", formatTimestamp(object.ModifiedAt))
	}

	if object.Status != nil {
		d.Set("status", object.Status.CurrentSessionClusterStatus)
	}

	if object.BasicResourceSetting != nil {
		d.Set("basic_resource_setting", flattenBasicResourceSetting(object.BasicResourceSetting))
	}

	if object.FlinkConf != nil {
		d.Set("user_flink_conf", flattenFlinkConf(object.FlinkConf))
	}

	if object.Labels != nil {
		d.Set("labels", flattenLabels(object.Labels))
	}

	if object.Logging != nil {
		d.Set("logging", flattenLogging(object.Logging))
	}

	return nil
}

func resourceAliCloudFlinkSessionClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId, namespaceName, sessionClusterName, err := parseSessionClusterId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	// First, retrieve the current complete session cluster to ensure we don't lose existing settings
	existingCluster, err := flinkService.DescribeSessionCluster(d.Id())
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DescribeSessionCluster", AlibabaCloudSdkGoERROR)
	}

	// Initialize updateRequest with the existing complete configuration
	updateRequest := &flinkAPI.SessionCluster{
		Name:                 sessionClusterName,
		Namespace:            namespaceName,
		Workspace:            workspaceId,
		EngineVersion:        existingCluster.EngineVersion,
		DeploymentTargetName: existingCluster.DeploymentTargetName,
		BasicResourceSetting: existingCluster.BasicResourceSetting,
		FlinkConf:            existingCluster.FlinkConf,
		Labels:               existingCluster.Labels,
		Logging:              existingCluster.Logging,
	}

	update := false

	if d.HasChange("engine_version") {
		updateRequest.EngineVersion = d.Get("engine_version").(string)
		update = true
	}

	if d.HasChange("deployment_target_name") {
		updateRequest.DeploymentTargetName = d.Get("deployment_target_name").(string)
		update = true
	}

	if d.HasChange("basic_resource_setting") {
		if v, ok := d.GetOk("basic_resource_setting"); ok {
			updateRequest.BasicResourceSetting = expandBasicResourceSetting(v.([]interface{}))
		}
		update = true
	}

	if d.HasChange("user_flink_conf") {
		if v, ok := d.GetOk("user_flink_conf"); ok {
			updateRequest.FlinkConf = expandFlinkConf(v.(map[string]interface{}))
		}
		update = true
	}

	if d.HasChange("labels") {
		if v, ok := d.GetOk("labels"); ok {
			updateRequest.Labels = expandLabels(v.(map[string]interface{}))
		}
		update = true
	}

	if d.HasChange("logging") {
		if v, ok := d.GetOk("logging"); ok {
			updateRequest.Logging = expandLogging(v.([]interface{}))
		}
		update = true
	}

	if update {
		_, err := flinkService.UpdateSessionCluster(workspaceId, namespaceName, sessionClusterName, updateRequest)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateSessionCluster", AlibabaCloudSdkGoERROR)
		}

		// Wait for update to complete
		stateConf := BuildStateConf([]string{}, []string{"RUNNING", "STOPPED"}, d.Timeout(schema.TimeoutUpdate), 30*time.Second, flinkService.SessionClusterStateRefreshFunc(d.Id(), []string{"FAILED", "TERMINATED"}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	// Handle status parameter change - start/stop cluster if needed
	if d.HasChange("status") {
		if targetStatus, ok := d.GetOk("status"); ok {
			err := handleSessionClusterStatusChange(d, flinkService, workspaceId, namespaceName, sessionClusterName, targetStatus.(string))
			if err != nil {
				return err
			}
		}
	}

	return resourceAliCloudFlinkSessionClusterRead(d, meta)
}

func resourceAliCloudFlinkSessionClusterDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId, namespaceName, sessionClusterName, err := parseSessionClusterId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	_, err = flinkService.DeleteSessionCluster(workspaceId, namespaceName, sessionClusterName)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidSessionCluster.NotFound", "SessionClusterNotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteSessionCluster", AlibabaCloudSdkGoERROR)
	}

	// Wait for deletion
	stateConf := BuildStateConf([]string{}, []string{}, d.Timeout(schema.TimeoutDelete), 30*time.Second, flinkService.SessionClusterStateRefreshFunc(d.Id(), []string{"FAILED"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}

// Helper functions for expanding and flattening schema data

func expandBasicResourceSetting(configured []interface{}) *flinkAPI.BasicResourceSetting {
	if len(configured) == 0 || configured[0] == nil {
		return nil
	}

	raw := configured[0].(map[string]interface{})
	setting := &flinkAPI.BasicResourceSetting{}

	if v, ok := raw["parallelism"]; ok && v.(int) > 0 {
		setting.Parallelism = int64(v.(int))
	}

	if v, ok := raw["jobmanager_resource_setting_spec"]; ok {
		setting.JobManagerResourceSettingSpec = expandBasicResourceSettingSpec(v.([]interface{}))
	}

	if v, ok := raw["taskmanager_resource_setting_spec"]; ok {
		setting.TaskManagerResourceSettingSpec = expandBasicResourceSettingSpec(v.([]interface{}))
	}

	return setting
}

func expandBasicResourceSettingSpec(configured []interface{}) *flinkAPI.BasicResourceSettingSpec {
	if len(configured) == 0 || configured[0] == nil {
		return nil
	}

	raw := configured[0].(map[string]interface{})
	spec := &flinkAPI.BasicResourceSettingSpec{}

	if v, ok := raw["cpu"]; ok && v.(float64) > 0 {
		spec.Cpu = v.(float64)
	}

	if v, ok := raw["memory"]; ok && v.(string) != "" {
		spec.Memory = v.(string)
	}

	return spec
}

func expandFlinkConf(configured map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range configured {
		result[k] = v
	}
	return result
}

func expandLabels(configured map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range configured {
		result[k] = v
	}
	return result
}

func expandLogging(configured []interface{}) *flinkAPI.Logging {
	if len(configured) == 0 || configured[0] == nil {
		return nil
	}

	raw := configured[0].(map[string]interface{})
	logging := &flinkAPI.Logging{}

	if v, ok := raw["logging_profile"]; ok {
		logging.LoggingProfile = v.(string)
	}

	if v, ok := raw["log4j2_configuration_template"]; ok {
		logging.Log4j2ConfigurationTemplate = v.(string)
	}

	if v, ok := raw["log4j_loggers"]; ok {
		logging.Log4jLoggers = expandLog4jLoggers(v.([]interface{}))
	}

	if v, ok := raw["log_reserve_policy"]; ok {
		logging.LogReservePolicy = expandLogReservePolicy(v.([]interface{}))
	}

	return logging
}

func expandLog4jLoggers(configured []interface{}) []flinkAPI.Log4jLogger {
	loggers := make([]flinkAPI.Log4jLogger, len(configured))
	for i, raw := range configured {
		logger := raw.(map[string]interface{})
		loggers[i] = flinkAPI.Log4jLogger{
			LoggerName:  logger["logger_name"].(string),
			LoggerLevel: logger["logger_level"].(string),
		}
	}
	return loggers
}

func expandLogReservePolicy(configured []interface{}) *flinkAPI.LogReservePolicy {
	if len(configured) == 0 || configured[0] == nil {
		return nil
	}

	raw := configured[0].(map[string]interface{})
	return &flinkAPI.LogReservePolicy{
		ExpirationDays: raw["expiration_days"].(int),
		OpenHistory:    raw["open_history"].(bool),
	}
}

func flattenBasicResourceSetting(setting *flinkAPI.BasicResourceSetting) []interface{} {
	if setting == nil {
		return []interface{}{}
	}

	result := map[string]interface{}{
		"parallelism": int(setting.Parallelism),
	}

	if setting.JobManagerResourceSettingSpec != nil {
		result["jobmanager_resource_setting_spec"] = flattenBasicResourceSettingSpec(setting.JobManagerResourceSettingSpec)
	}

	if setting.TaskManagerResourceSettingSpec != nil {
		result["taskmanager_resource_setting_spec"] = flattenBasicResourceSettingSpec(setting.TaskManagerResourceSettingSpec)
	}

	return []interface{}{result}
}

func flattenBasicResourceSettingSpec(spec *flinkAPI.BasicResourceSettingSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"cpu":    spec.Cpu,
			"memory": spec.Memory,
		},
	}
}

func flattenFlinkConf(flinkConf map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range flinkConf {
		result[k] = v
	}
	return result
}

func flattenLabels(labels map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range labels {
		result[k] = v
	}
	return result
}

func flattenLogging(logging *flinkAPI.Logging) []interface{} {
	if logging == nil {
		return []interface{}{}
	}

	result := map[string]interface{}{
		"logging_profile":               logging.LoggingProfile,
		"log4j2_configuration_template": logging.Log4j2ConfigurationTemplate,
	}

	if len(logging.Log4jLoggers) > 0 {
		result["log4j_loggers"] = flattenLog4jLoggers(logging.Log4jLoggers)
	}

	if logging.LogReservePolicy != nil {
		result["log_reserve_policy"] = flattenLogReservePolicy(logging.LogReservePolicy)
	}

	return []interface{}{result}
}

func flattenLog4jLoggers(loggers []flinkAPI.Log4jLogger) []interface{} {
	result := make([]interface{}, len(loggers))
	for i, logger := range loggers {
		result[i] = map[string]interface{}{
			"logger_name":  logger.LoggerName,
			"logger_level": logger.LoggerLevel,
		}
	}
	return result
}

func flattenLogReservePolicy(policy *flinkAPI.LogReservePolicy) []interface{} {
	if policy == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"expiration_days": policy.ExpirationDays,
			"open_history":    policy.OpenHistory,
		},
	}
}

// Helper function to format timestamp
func formatTimestamp(timestamp int64) string {
	return time.Unix(timestamp/1000, 0).Format(time.RFC3339)
}

// handleSessionClusterStatusChange handles starting/stopping session cluster based on target status
func handleSessionClusterStatusChange(d *schema.ResourceData, flinkService *FlinkService, workspaceId, namespaceName, sessionClusterName, targetStatus string) error {
	// Get current status
	currentCluster, err := flinkService.DescribeSessionCluster(d.Id())
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DescribeSessionCluster", AlibabaCloudSdkGoERROR)
	}

	currentStatus := "UNKNOWN"
	if currentCluster != nil && currentCluster.Status != nil {
		currentStatus = currentCluster.Status.CurrentSessionClusterStatus
	}

	// Only handle RUNNING and STOPPED target statuses for start/stop operations
	switch targetStatus {
	case "RUNNING":
		if currentStatus == "STOPPED" {
			// Start the session cluster
			err := flinkService.StartSessionCluster(workspaceId, namespaceName, sessionClusterName)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "StartSessionCluster", AlibabaCloudSdkGoERROR)
			}

			// Wait for cluster to reach RUNNING state
			stateConf := BuildStateConf([]string{"STOPPED", "STARTING"}, []string{"RUNNING"}, d.Timeout(schema.TimeoutUpdate), 30*time.Second, flinkService.SessionClusterStateRefreshFunc(d.Id(), []string{"FAILED", "TERMINATED"}))
			if _, err := stateConf.WaitForState(); err != nil {
				return WrapErrorf(err, IdMsg, d.Id())
			}
		}
	case "STOPPED":
		if currentStatus == "RUNNING" {
			// Stop the session cluster
			err := flinkService.StopSessionCluster(workspaceId, namespaceName, sessionClusterName)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "StopSessionCluster", AlibabaCloudSdkGoERROR)
			}

			// Wait for cluster to reach STOPPED state
			stateConf := BuildStateConf([]string{"RUNNING", "STOPPING"}, []string{"STOPPED"}, d.Timeout(schema.TimeoutUpdate), 30*time.Second, flinkService.SessionClusterStateRefreshFunc(d.Id(), []string{"FAILED", "TERMINATED"}))
			if _, err := stateConf.WaitForState(); err != nil {
				return WrapErrorf(err, IdMsg, d.Id())
			}
		}
	default:
		// For other statuses (STARTING, UPDATING, STOPPING, FAILED), we don't perform any action
		// These are either transitional states or error states that should be handled by the service
		return nil
	}

	return nil
}
