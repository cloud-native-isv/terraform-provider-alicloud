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
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"deployment_target_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"basic_resource_setting": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"parallelism": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"jobmanager_resource_setting_spec": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cpu": {
										Type:         schema.TypeFloat,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.FloatAtLeast(0.1),
									},
									"memory": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
						"taskmanager_resource_setting_spec": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cpu": {
										Type:         schema.TypeFloat,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.FloatAtLeast(0.1),
									},
									"memory": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
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
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"logging": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"logging_profile": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"log4j2_configuration_template": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"log4j_loggers": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"logger_name": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"logger_level": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice([]string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"}, false),
									},
								},
							},
						},
						"log_reserve_policy": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"expiration_days": {
										Type:         schema.TypeInt,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"open_history": {
										Type:     schema.TypeBool,
										Optional: true,
										ForceNew: true,
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
				ValidateFunc: validation.StringInSlice(
					flinkAPI.FlinkSessionClusterStatusesToStrings([]flinkAPI.FlinkSessionClusterStatus{
						flinkAPI.FlinkSessionClusterStatusStopped,
						flinkAPI.FlinkSessionClusterStatusRunning,
					}), false),
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
		sessionCluster.BasicResourceSetting = flinkAPI.ExpandBasicResourceSetting(v.([]interface{}))
	}

	// Set flink configuration
	if v, ok := d.GetOk("user_flink_conf"); ok {
		sessionCluster.FlinkConf = flinkAPI.ExpandFlinkConf(v.(map[string]interface{}))
	}

	// Set labels
	if v, ok := d.GetOk("labels"); ok {
		sessionCluster.Labels = flinkAPI.ExpandLabels(v.(map[string]interface{}))
	}

	// Set logging configuration
	if v, ok := d.GetOk("logging"); ok {
		sessionCluster.Logging = flinkAPI.ExpandLogging(v.([]interface{}))
	}

	_, err = flinkService.CreateSessionCluster(workspaceId, namespaceName, sessionCluster)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_session_cluster", "CreateSessionCluster", AlibabaCloudSdkGoERROR)
	}

	d.SetId(formatSessionClusterId(workspaceId, namespaceName, sessionClusterName))

	// Wait for session cluster to be ready
	err = flinkService.WaitForSessionClusterCreating(d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
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
		if IsNotFoundError(err) {
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
		d.Set("created_at", flinkAPI.FormatTimestamp(object.CreatedAt))
	}
	if object.ModifiedAt != 0 {
		d.Set("modified_at", flinkAPI.FormatTimestamp(object.ModifiedAt))
	}

	if object.Status != nil {
		d.Set("status", object.Status.CurrentSessionClusterStatus)
	}

	if object.BasicResourceSetting != nil {
		d.Set("basic_resource_setting", flinkAPI.FlattenBasicResourceSetting(object.BasicResourceSetting))
	}

	if object.FlinkConf != nil {
		d.Set("user_flink_conf", flinkAPI.FlattenFlinkConf(object.FlinkConf))
	}

	if object.Labels != nil {
		d.Set("labels", flinkAPI.FlattenLabels(object.Labels))
	}

	if object.Logging != nil {
		d.Set("logging", flinkAPI.FlattenLogging(object.Logging))
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
			updateRequest.BasicResourceSetting = flinkAPI.ExpandBasicResourceSetting(v.([]interface{}))
		}
		update = true
	}

	if d.HasChange("user_flink_conf") {
		if v, ok := d.GetOk("user_flink_conf"); ok {
			updateRequest.FlinkConf = flinkAPI.ExpandFlinkConf(v.(map[string]interface{}))
		}
		update = true
	}

	if d.HasChange("labels") {
		if v, ok := d.GetOk("labels"); ok {
			updateRequest.Labels = flinkAPI.ExpandLabels(v.(map[string]interface{}))
		}
		update = true
	}

	if d.HasChange("logging") {
		if v, ok := d.GetOk("logging"); ok {
			updateRequest.Logging = flinkAPI.ExpandLogging(v.([]interface{}))
		}
		update = true
	}

	if update {
		_, err := flinkService.UpdateSessionCluster(workspaceId, namespaceName, sessionClusterName, updateRequest)
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateSessionCluster", AlibabaCloudSdkGoERROR)
		}

		// Wait for update to complete
		err = flinkService.WaitForSessionClusterUpdating(d.Id(), d.Timeout(schema.TimeoutUpdate))
		if err != nil {
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

	// First, get the current session cluster to check its status
	currentCluster, err := flinkService.DescribeSessionCluster(d.Id())
	if err != nil {
		if IsNotFoundError(err) {
			return nil
		}
		return WrapError(err)
	}

	// Stop the session cluster if it's running
	currentStatus := "UNKNOWN"
	if currentCluster != nil && currentCluster.Status != nil {
		currentStatus = currentCluster.Status.CurrentSessionClusterStatus
	}

	// Only stop if it's running
	if currentStatus == "RUNNING" {
		err = flinkService.StopSessionCluster(workspaceId, namespaceName, sessionClusterName)
		if err != nil {
			if IsNotFoundError(err) {
				return nil
			}
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "StopSessionCluster", AlibabaCloudSdkGoERROR)
		}

		// Wait for session cluster to stop
		err = flinkService.WaitForSessionClusterStopped(d.Id(), d.Timeout(schema.TimeoutDelete))
		if err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	// Now delete the session cluster
	_, err = flinkService.DeleteSessionCluster(workspaceId, namespaceName, sessionClusterName)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidSessionCluster.NotFound", "SessionClusterNotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteSessionCluster", AlibabaCloudSdkGoERROR)
	}

	// Wait for deletion
	err = flinkService.WaitForSessionClusterDeleting(d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}

// Helper function to handle session cluster status changes
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
	case flinkAPI.FlinkSessionClusterStatusRunning.String():
		if currentStatus == flinkAPI.FlinkSessionClusterStatusStopped.String() {
			// Start the session cluster
			err := flinkService.StartSessionCluster(workspaceId, namespaceName, sessionClusterName)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "StartSessionCluster", AlibabaCloudSdkGoERROR)
			}

			// Wait for cluster to reach RUNNING state
			err = flinkService.WaitForSessionClusterRunning(d.Id(), d.Timeout(schema.TimeoutUpdate))
			if err != nil {
				return WrapErrorf(err, IdMsg, d.Id())
			}
		}
	case flinkAPI.FlinkSessionClusterStatusStopped.String():
		if currentStatus == flinkAPI.FlinkSessionClusterStatusRunning.String() {
			// Stop the session cluster
			err := flinkService.StopSessionCluster(workspaceId, namespaceName, sessionClusterName)
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "StopSessionCluster", AlibabaCloudSdkGoERROR)
			}

			// Wait for cluster to reach STOPPED state
			err = flinkService.WaitForSessionClusterStopped(d.Id(), d.Timeout(schema.TimeoutUpdate))
			if err != nil {
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
