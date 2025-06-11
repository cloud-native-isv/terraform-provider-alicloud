package alicloud

import (
	"log"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudFlinkDeployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkDeploymentCreate,
		Read:   resourceAliCloudFlinkDeploymentRead,
		Update: resourceAliCloudFlinkDeploymentUpdate,
		Delete: resourceAliCloudFlinkDeploymentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"namespace_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"job_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			// Basic deployment configuration
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "vvr-8.0.5-flink-1.17",
			},
			"execution_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "STREAMING",
				ValidateFunc: validation.StringInSlice([]string{"STREAMING", "BATCH"}, false),
			},
			// Deployment target configuration
			"deployment_target": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "session-cluster",
						},
					},
				},
			},
			// Artifact configuration (existing fields for backward compatibility)
			"entry_class": {
				Type:     schema.TypeString,
				Required: true,
			},
			"jar_uri": {
				Type:     schema.TypeString,
				Required: true,
			},
			"jar_artifact_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"program_args": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// Extended artifact configuration
			"artifact": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kind": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"JAR", "PYTHON", "SQLSCRIPT"}, false),
						},
						"jar_artifact": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"jar_uri": {
										Type:     schema.TypeString,
										Required: true,
									},
									"entry_class": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"main_args": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"additional_dependencies": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"python_artifact": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"python_artifact_uri": {
										Type:     schema.TypeString,
										Required: true,
									},
									"entry_module": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"main_args": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"additional_dependencies": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"additional_python_libraries": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"additional_python_archives": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"sql_artifact": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"sql_script": {
										Type:     schema.TypeString,
										Required: true,
									},
									"additional_dependencies": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			// Resource settings
			"parallelism": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"streaming_resource_setting": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_setting_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "BASIC",
							ValidateFunc: validation.StringInSlice([]string{"BASIC", "EXPERT"}, false),
						},
						"basic_resource_setting": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"parallelism": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"jobmanager_resource_setting_spec": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cpu": {
													Type:     schema.TypeFloat,
													Optional: true,
												},
												"memory": {
													Type:     schema.TypeString,
													Optional: true,
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
													Type:     schema.TypeFloat,
													Optional: true,
												},
												"memory": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"expert_resource_setting": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_plan": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"jobmanager_resource_setting_spec": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cpu": {
													Type:     schema.TypeFloat,
													Optional: true,
												},
												"memory": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			// Flink configuration
			"flink_conf": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			// Logging configuration
			"logging": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"logging_profile": {
							Type:     schema.TypeString,
							Optional: true,
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
										Type:     schema.TypeString,
										Required: true,
									},
									"logger_level": {
										Type:     schema.TypeString,
										Required: true,
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
										Type:     schema.TypeInt,
										Optional: true,
									},
									"open_history": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			// Tags configuration
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			// Computed fields
			"deployment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"update_time": {
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
			"referenced_deployment_draft_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceAliCloudFlinkDeploymentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// Create deployment request using cws-lib-go types
	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)

	request := &aliyunFlinkAPI.Deployment{
		Workspace: workspaceId,
		Namespace: namespaceName,
		Name:      d.Get("job_name").(string),
	}

	// Set basic deployment properties
	if description, ok := d.GetOk("description"); ok {
		request.Description = description.(string)
	}

	if engineVersion, ok := d.GetOk("engine_version"); ok {
		request.EngineVersion = engineVersion.(string)
	}

	if executionMode, ok := d.GetOk("execution_mode"); ok {
		request.ExecutionMode = executionMode.(string)
	}

	// Handle deployment target
	if deploymentTargetList, ok := d.GetOk("deployment_target"); ok {
		targets := deploymentTargetList.([]interface{})
		if len(targets) > 0 {
			targetMap := targets[0].(map[string]interface{})
			request.DeploymentTarget = &aliyunFlinkAPI.DeploymentTarget{
				Name: targetMap["name"].(string),
			}
		}
	}

	// Handle artifact configuration using helper function
	artifact, err := expandArtifact(d)
	if err != nil {
		return WrapError(err)
	}
	if artifact != nil {
		request.Artifact = artifact
	}

	// Handle streaming resource setting using helper function
	streamingResourceSetting := expandStreamingResourceSetting(d)
	if streamingResourceSetting != nil {
		request.StreamingResourceSetting = streamingResourceSetting
	}

	// Handle Flink configuration
	if flinkConf := d.Get("flink_conf").(map[string]interface{}); len(flinkConf) > 0 {
		request.FlinkConf = make(map[string]string)
		for k, v := range flinkConf {
			request.FlinkConf[k] = v.(string)
		}
	}

	// Handle logging configuration using helper function
	logging := expandLogging(d)
	if logging != nil {
		// Convert Logging to LoggingProfile for deployment
		request.Logging = &aliyunFlinkAPI.LoggingProfile{
			Template: logging.LoggingProfile,
		}
	}

	// Handle tags
	if tags := d.Get("tags").(map[string]interface{}); len(tags) > 0 {
		request.Labels = make(map[string]string)
		for k, v := range tags {
			request.Labels[k] = v.(string)
		}
	}

	// Create deployment
	var response *aliyunFlinkAPI.Deployment
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		resp, err := flinkService.CreateDeployment(&namespaceName, request)
		if err != nil {
			if IsExpectedErrors(err, []string{"ThrottlingException", "OperationConflict"}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		response = resp
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_deployment", "CreateDeployment", AlibabaCloudSdkGoERROR)
	}

	if response == nil || response.DeploymentId == "" {
		return WrapError(Error("Failed to get deployment ID from response"))
	}

	// Set composite ID: namespace:deploymentId
	d.SetId(namespaceName + ":" + response.DeploymentId)

	// Wait for deployment creation to complete using StateRefreshFunc
	stateConf := BuildStateConf([]string{"CREATING", "STARTING"}, []string{"CREATED", "RUNNING"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, flinkService.FlinkDeploymentStateRefreshFunc(d.Id(), []string{"FAILED"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// 最后调用Read同步状态
	return resourceAliCloudFlinkDeploymentRead(d, meta)
}

func resourceAliCloudFlinkDeploymentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	deployment, err := flinkService.GetDeployment(d.Id())
	if err != nil {
		if !d.IsNewResource() && NotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_flink_deployment GetDeployment Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Parse namespace and deployment ID from composite ID
	namespaceName, deploymentId, err := parseDeploymentId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	// Set basic attributes
	d.Set("workspace_id", deployment.Workspace)
	d.Set("namespace_name", namespaceName)
	d.Set("deployment_id", deploymentId)
	d.Set("job_name", deployment.Name)
	d.Set("description", deployment.Description)
	d.Set("engine_version", deployment.EngineVersion)
	d.Set("execution_mode", deployment.ExecutionMode)
	d.Set("create_time", deployment.CreatedAt)
	d.Set("update_time", deployment.ModifiedAt)
	d.Set("creator", deployment.Creator)
	d.Set("creator_name", deployment.CreatorName)
	d.Set("modifier", deployment.Modifier)
	d.Set("modifier_name", deployment.ModifierName)
	d.Set("referenced_deployment_draft_id", deployment.ReferencedDeploymentDraftId)

	// Set deployment target
	if deployment.DeploymentTarget != nil {
		deploymentTargetMap := map[string]interface{}{
			"name": deployment.DeploymentTarget.Name,
		}
		d.Set("deployment_target", []interface{}{deploymentTargetMap})
	}

	// Set artifact information using both new and legacy fields for backward compatibility
	if deployment.Artifact != nil {
		// Set new artifact structure
		d.Set("artifact", flattenArtifact(deployment.Artifact))

		// Set legacy fields for backward compatibility
		if deployment.Artifact.JarArtifact != nil {
			d.Set("jar_uri", deployment.Artifact.JarArtifact.JarUri)
			d.Set("entry_class", deployment.Artifact.JarArtifact.EntryClass)
			d.Set("program_args", deployment.Artifact.JarArtifact.MainArgs)
		}
	}

	// Set streaming resource setting using both new and legacy fields
	if deployment.StreamingResourceSetting != nil {
		// Set new streaming resource setting structure
		d.Set("streaming_resource_setting", flattenStreamingResourceSetting(deployment.StreamingResourceSetting))

		// Set legacy parallelism field for backward compatibility
		if deployment.StreamingResourceSetting.BasicResourceSetting != nil {
			d.Set("parallelism", deployment.StreamingResourceSetting.BasicResourceSetting.Parallelism)
		}
	}

	// Set Flink configuration
	if deployment.FlinkConf != nil {
		flinkConf := make(map[string]interface{})
		for k, v := range deployment.FlinkConf {
			flinkConf[k] = v
		}
		d.Set("flink_conf", flinkConf)
	}

	// Set logging configuration
	if deployment.Logging != nil {
		// Convert LoggingProfile to Logging for flattening
		logging := &aliyunFlinkAPI.Logging{
			LoggingProfile: deployment.Logging.Template,
		}
		d.Set("logging", flattenLogging(logging))
	}

	// Set tags
	if deployment.Labels != nil {
		d.Set("tags", deployment.Labels)
	}

	// Set job status from job summary
	if deployment.JobSummary != nil {
		d.Set("status", deployment.GetStatus())
	}

	return nil
}

func resourceAliCloudFlinkDeploymentUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// Get current deployment
	deployment, err := flinkService.GetDeployment(d.Id())
	if err != nil {
		return WrapError(err)
	}

	hasChanged := false

	// Check for changes in basic deployment properties
	if d.HasChange("job_name") {
		deployment.Name = d.Get("job_name").(string)
		hasChanged = true
	}

	if d.HasChange("description") {
		deployment.Description = d.Get("description").(string)
		hasChanged = true
	}

	if d.HasChange("engine_version") {
		deployment.EngineVersion = d.Get("engine_version").(string)
		hasChanged = true
	}

	if d.HasChange("execution_mode") {
		deployment.ExecutionMode = d.Get("execution_mode").(string)
		hasChanged = true
	}

	// Check for changes in deployment target
	if d.HasChange("deployment_target") {
		if deploymentTargetList, ok := d.GetOk("deployment_target"); ok {
			targets := deploymentTargetList.([]interface{})
			if len(targets) > 0 {
				targetMap := targets[0].(map[string]interface{})
				deployment.DeploymentTarget = &aliyunFlinkAPI.DeploymentTarget{
					Name: targetMap["name"].(string),
				}
			}
		}
		hasChanged = true
	}

	// Check for changes in artifact configuration (new or legacy)
	if d.HasChange("artifact") || d.HasChange("jar_uri") || d.HasChange("entry_class") || d.HasChange("program_args") {
		artifact, err := expandArtifact(d)
		if err != nil {
			return WrapError(err)
		}
		if artifact != nil {
			deployment.Artifact = artifact
		}
		hasChanged = true
	}

	// Check for changes in streaming resource setting
	if d.HasChange("streaming_resource_setting") || d.HasChange("parallelism") {
		streamingResourceSetting := expandStreamingResourceSetting(d)
		if streamingResourceSetting != nil {
			deployment.StreamingResourceSetting = streamingResourceSetting
		}
		hasChanged = true
	}

	// Check for changes in Flink configuration
	if d.HasChange("flink_conf") {
		if flinkConf := d.Get("flink_conf").(map[string]interface{}); len(flinkConf) > 0 {
			deployment.FlinkConf = make(map[string]string)
			for k, v := range flinkConf {
				deployment.FlinkConf[k] = v.(string)
			}
		} else {
			deployment.FlinkConf = nil
		}
		hasChanged = true
	}

	// Check for changes in logging configuration
	if d.HasChange("logging") {
		logging := expandLogging(d)
		if logging != nil {
			// Convert Logging to LoggingProfile for deployment
			deployment.Logging = &aliyunFlinkAPI.LoggingProfile{
				Template: logging.LoggingProfile,
			}
		} else {
			deployment.Logging = nil
		}
		hasChanged = true
	}

	// Check for changes in tags
	if d.HasChange("tags") {
		if tags := d.Get("tags").(map[string]interface{}); len(tags) > 0 {
			deployment.Labels = make(map[string]string)
			for k, v := range tags {
				deployment.Labels[k] = v.(string)
			}
		} else {
			deployment.Labels = nil
		}
		hasChanged = true
	}

	if hasChanged {
		// Update deployment
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err := flinkService.UpdateDeployment(deployment)
			if err != nil {
				if IsExpectedErrors(err, []string{"ThrottlingException", "OperationConflict"}) {
					time.Sleep(5 * time.Second)
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateDeployment", AlibabaCloudSdkGoERROR)
		}

		// Wait for update to complete
		stateConf := resource.StateChangeConf{
			Pending:    []string{"UPDATING"},
			Target:     []string{"RUNNING", "CREATED"},
			Refresh:    flinkService.FlinkDeploymentStateRefreshFunc(d.Id(), []string{"FAILED"}),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      5 * time.Second,
			MinTimeout: 3 * time.Second,
		}
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	return resourceAliCloudFlinkDeploymentRead(d, meta)
}

func resourceAliCloudFlinkDeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// Parse namespace and deployment ID from composite ID
	namespaceName, deploymentId, err := parseDeploymentId(d.Id())
	if err != nil {
		return WrapError(err)
	}

	err = flinkService.DeleteDeployment(namespaceName, deploymentId)
	if err != nil {
		if IsExpectedErrors(err, []string{"InvalidDeployment.NotFound"}) {
			return nil
		}
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteDeployment", AlibabaCloudSdkGoERROR)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"DELETING"},
		Target:     []string{},
		Refresh:    flinkService.FlinkDeploymentStateRefreshFunc(d.Id(), []string{}),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}

// Helper functions for schema conversion

// expandArtifact converts schema artifact to API artifact
func expandArtifact(d *schema.ResourceData) (*aliyunFlinkAPI.Artifact, error) {
	// Check for new artifact structure first
	if artifactList, ok := d.GetOk("artifact"); ok {
		artifacts := artifactList.([]interface{})
		if len(artifacts) > 0 {
			artifactMap := artifacts[0].(map[string]interface{})
			artifact := &aliyunFlinkAPI.Artifact{
				Kind: artifactMap["kind"].(string),
			}

			switch artifact.Kind {
			case "JAR":
				if jarArtifactList, ok := artifactMap["jar_artifact"].([]interface{}); ok && len(jarArtifactList) > 0 {
					jarMap := jarArtifactList[0].(map[string]interface{})
					artifact.JarArtifact = &aliyunFlinkAPI.JarArtifact{
						JarUri:     jarMap["jar_uri"].(string),
						EntryClass: jarMap["entry_class"].(string),
						MainArgs:   jarMap["main_args"].(string),
					}
					if deps, ok := jarMap["additional_dependencies"].([]interface{}); ok {
						artifact.JarArtifact.AdditionalDependencies = make([]string, len(deps))
						for i, dep := range deps {
							artifact.JarArtifact.AdditionalDependencies[i] = dep.(string)
						}
					}
				}
			case "PYTHON":
				if pythonArtifactList, ok := artifactMap["python_artifact"].([]interface{}); ok && len(pythonArtifactList) > 0 {
					pythonMap := pythonArtifactList[0].(map[string]interface{})
					artifact.PythonArtifact = &aliyunFlinkAPI.PythonArtifact{
						PythonArtifactUri: pythonMap["python_artifact_uri"].(string),
						EntryModule:       pythonMap["entry_module"].(string),
						MainArgs:          pythonMap["main_args"].(string),
					}
					if deps, ok := pythonMap["additional_dependencies"].([]interface{}); ok {
						artifact.PythonArtifact.AdditionalDependencies = make([]string, len(deps))
						for i, dep := range deps {
							artifact.PythonArtifact.AdditionalDependencies[i] = dep.(string)
						}
					}
					if libs, ok := pythonMap["additional_python_libraries"].([]interface{}); ok {
						artifact.PythonArtifact.AdditionalPythonLibraries = make([]string, len(libs))
						for i, lib := range libs {
							artifact.PythonArtifact.AdditionalPythonLibraries[i] = lib.(string)
						}
					}
					if archives, ok := pythonMap["additional_python_archives"].([]interface{}); ok {
						artifact.PythonArtifact.AdditionalPythonArchives = make([]string, len(archives))
						for i, archive := range archives {
							artifact.PythonArtifact.AdditionalPythonArchives[i] = archive.(string)
						}
					}
				}
			case "SQLSCRIPT":
				if sqlArtifactList, ok := artifactMap["sql_artifact"].([]interface{}); ok && len(sqlArtifactList) > 0 {
					sqlMap := sqlArtifactList[0].(map[string]interface{})
					artifact.SqlArtifact = &aliyunFlinkAPI.SqlArtifact{
						SqlScript: sqlMap["sql_script"].(string),
					}
					if deps, ok := sqlMap["additional_dependencies"].([]interface{}); ok {
						artifact.SqlArtifact.AdditionalDependencies = make([]string, len(deps))
						for i, dep := range deps {
							artifact.SqlArtifact.AdditionalDependencies[i] = dep.(string)
						}
					}
				}
			}
			return artifact, nil
		}
	}

	// Fallback to legacy artifact configuration for backward compatibility
	if jarUri, ok := d.GetOk("jar_uri"); ok {
		artifact := &aliyunFlinkAPI.Artifact{
			Kind: "JAR",
			JarArtifact: &aliyunFlinkAPI.JarArtifact{
				JarUri: jarUri.(string),
			},
		}
		if entryClass, ok := d.GetOk("entry_class"); ok {
			artifact.JarArtifact.EntryClass = entryClass.(string)
		}
		if mainArgs, ok := d.GetOk("program_args"); ok {
			artifact.JarArtifact.MainArgs = mainArgs.(string)
		}
		return artifact, nil
	}

	return nil, nil
}

// expandStreamingResourceSetting converts schema streaming resource setting to API format
func expandStreamingResourceSetting(d *schema.ResourceData) *aliyunFlinkAPI.StreamingResourceSetting {
	if streamingResourceList, ok := d.GetOk("streaming_resource_setting"); ok {
		settings := streamingResourceList.([]interface{})
		if len(settings) > 0 {
			settingMap := settings[0].(map[string]interface{})
			resourceSetting := &aliyunFlinkAPI.StreamingResourceSetting{
				ResourceSettingMode: settingMap["resource_setting_mode"].(string),
			}

			if basicList, ok := settingMap["basic_resource_setting"].([]interface{}); ok && len(basicList) > 0 {
				basicMap := basicList[0].(map[string]interface{})
				resourceSetting.BasicResourceSetting = &aliyunFlinkAPI.BasicResourceSetting{
					Parallelism: int64(basicMap["parallelism"].(int)),
				}

				if jmList, ok := basicMap["jobmanager_resource_setting_spec"].([]interface{}); ok && len(jmList) > 0 {
					jmMap := jmList[0].(map[string]interface{})
					resourceSetting.BasicResourceSetting.JobManagerResourceSettingSpec = &aliyunFlinkAPI.BasicResourceSettingSpec{
						Cpu:    jmMap["cpu"].(float64),
						Memory: jmMap["memory"].(string),
					}
				}

				if tmList, ok := basicMap["taskmanager_resource_setting_spec"].([]interface{}); ok && len(tmList) > 0 {
					tmMap := tmList[0].(map[string]interface{})
					resourceSetting.BasicResourceSetting.TaskManagerResourceSettingSpec = &aliyunFlinkAPI.BasicResourceSettingSpec{
						Cpu:    tmMap["cpu"].(float64),
						Memory: tmMap["memory"].(string),
					}
				}
			}

			if expertList, ok := settingMap["expert_resource_setting"].([]interface{}); ok && len(expertList) > 0 {
				expertMap := expertList[0].(map[string]interface{})
				resourceSetting.ExpertResourceSetting = &aliyunFlinkAPI.ExpertResourceSetting{
					ResourcePlan: expertMap["resource_plan"].(string),
				}

				if jmList, ok := expertMap["jobmanager_resource_setting_spec"].([]interface{}); ok && len(jmList) > 0 {
					jmMap := jmList[0].(map[string]interface{})
					resourceSetting.ExpertResourceSetting.JobManagerResourceSettingSpec = &aliyunFlinkAPI.BasicResourceSettingSpec{
						Cpu:    jmMap["cpu"].(float64),
						Memory: jmMap["memory"].(string),
					}
				}
			}

			return resourceSetting
		}
	}

	// Fallback to simple parallelism setting for backward compatibility
	if parallelism, ok := d.GetOk("parallelism"); ok {
		return &aliyunFlinkAPI.StreamingResourceSetting{
			ResourceSettingMode: "BASIC",
			BasicResourceSetting: &aliyunFlinkAPI.BasicResourceSetting{
				Parallelism: int64(parallelism.(int)),
			},
		}
	}

	return nil
}

// expandLogging converts schema logging configuration to API format
func expandLogging(d *schema.ResourceData) *aliyunFlinkAPI.Logging {
	if loggingList, ok := d.GetOk("logging"); ok {
		loggings := loggingList.([]interface{})
		if len(loggings) > 0 {
			loggingMap := loggings[0].(map[string]interface{})
			logging := &aliyunFlinkAPI.Logging{}

			if profile, ok := loggingMap["logging_profile"].(string); ok {
				logging.LoggingProfile = profile
			}

			if template, ok := loggingMap["log4j2_configuration_template"].(string); ok {
				logging.Log4j2ConfigurationTemplate = template
			}

			if loggersList, ok := loggingMap["log4j_loggers"].([]interface{}); ok {
				logging.Log4jLoggers = make([]aliyunFlinkAPI.Log4jLogger, len(loggersList))
				for i, loggerItem := range loggersList {
					loggerMap := loggerItem.(map[string]interface{})
					logging.Log4jLoggers[i] = aliyunFlinkAPI.Log4jLogger{
						LoggerName:  loggerMap["logger_name"].(string),
						LoggerLevel: loggerMap["logger_level"].(string),
					}
				}
			}

			if reservePolicyList, ok := loggingMap["log_reserve_policy"].([]interface{}); ok && len(reservePolicyList) > 0 {
				policyMap := reservePolicyList[0].(map[string]interface{})
				logging.LogReservePolicy = &aliyunFlinkAPI.LogReservePolicy{
					ExpirationDays: policyMap["expiration_days"].(int),
					OpenHistory:    policyMap["open_history"].(bool),
				}
			}

			return logging
		}
	}
	return nil
}

// flattenArtifact converts API artifact to schema format
func flattenArtifact(artifact *aliyunFlinkAPI.Artifact) []interface{} {
	if artifact == nil {
		return []interface{}{}
	}

	artifactMap := map[string]interface{}{
		"kind": artifact.Kind,
	}

	switch artifact.Kind {
	case "JAR":
		if artifact.JarArtifact != nil {
			jarMap := map[string]interface{}{
				"jar_uri":     artifact.JarArtifact.JarUri,
				"entry_class": artifact.JarArtifact.EntryClass,
				"main_args":   artifact.JarArtifact.MainArgs,
			}
			if artifact.JarArtifact.AdditionalDependencies != nil {
				jarMap["additional_dependencies"] = artifact.JarArtifact.AdditionalDependencies
			}
			artifactMap["jar_artifact"] = []interface{}{jarMap}
		}
	case "PYTHON":
		if artifact.PythonArtifact != nil {
			pythonMap := map[string]interface{}{
				"python_artifact_uri": artifact.PythonArtifact.PythonArtifactUri,
				"entry_module":        artifact.PythonArtifact.EntryModule,
				"main_args":           artifact.PythonArtifact.MainArgs,
			}
			if artifact.PythonArtifact.AdditionalDependencies != nil {
				pythonMap["additional_dependencies"] = artifact.PythonArtifact.AdditionalDependencies
			}
			if artifact.PythonArtifact.AdditionalPythonLibraries != nil {
				pythonMap["additional_python_libraries"] = artifact.PythonArtifact.AdditionalPythonLibraries
			}
			if artifact.PythonArtifact.AdditionalPythonArchives != nil {
				pythonMap["additional_python_archives"] = artifact.PythonArtifact.AdditionalPythonArchives
			}
			artifactMap["python_artifact"] = []interface{}{pythonMap}
		}
	case "SQLSCRIPT":
		if artifact.SqlArtifact != nil {
			sqlMap := map[string]interface{}{
				"sql_script": artifact.SqlArtifact.SqlScript,
			}
			if artifact.SqlArtifact.AdditionalDependencies != nil {
				sqlMap["additional_dependencies"] = artifact.SqlArtifact.AdditionalDependencies
			}
			artifactMap["sql_artifact"] = []interface{}{sqlMap}
		}
	}

	return []interface{}{artifactMap}
}

// flattenStreamingResourceSetting converts API streaming resource setting to schema format
func flattenStreamingResourceSetting(setting *aliyunFlinkAPI.StreamingResourceSetting) []interface{} {
	if setting == nil {
		return []interface{}{}
	}

	settingMap := map[string]interface{}{
		"resource_setting_mode": setting.ResourceSettingMode,
	}

	if setting.BasicResourceSetting != nil {
		basicMap := map[string]interface{}{
			"parallelism": setting.BasicResourceSetting.Parallelism,
		}

		if setting.BasicResourceSetting.JobManagerResourceSettingSpec != nil {
			jmMap := map[string]interface{}{
				"cpu":    setting.BasicResourceSetting.JobManagerResourceSettingSpec.Cpu,
				"memory": setting.BasicResourceSetting.JobManagerResourceSettingSpec.Memory,
			}
			basicMap["jobmanager_resource_setting_spec"] = []interface{}{jmMap}
		}

		if setting.BasicResourceSetting.TaskManagerResourceSettingSpec != nil {
			tmMap := map[string]interface{}{
				"cpu":    setting.BasicResourceSetting.TaskManagerResourceSettingSpec.Cpu,
				"memory": setting.BasicResourceSetting.TaskManagerResourceSettingSpec.Memory,
			}
			basicMap["taskmanager_resource_setting_spec"] = []interface{}{tmMap}
		}

		settingMap["basic_resource_setting"] = []interface{}{basicMap}
	}

	if setting.ExpertResourceSetting != nil {
		expertMap := map[string]interface{}{
			"resource_plan": setting.ExpertResourceSetting.ResourcePlan,
		}

		if setting.ExpertResourceSetting.JobManagerResourceSettingSpec != nil {
			jmMap := map[string]interface{}{
				"cpu":    setting.ExpertResourceSetting.JobManagerResourceSettingSpec.Cpu,
				"memory": setting.ExpertResourceSetting.JobManagerResourceSettingSpec.Memory,
			}
			expertMap["jobmanager_resource_setting_spec"] = []interface{}{jmMap}
		}

		settingMap["expert_resource_setting"] = []interface{}{expertMap}
	}

	return []interface{}{settingMap}
}

// flattenLogging converts API logging configuration to schema format
func flattenLogging(logging *aliyunFlinkAPI.Logging) []interface{} {
	if logging == nil {
		return []interface{}{}
	}

	loggingMap := map[string]interface{}{
		"logging_profile":               logging.LoggingProfile,
		"log4j2_configuration_template": logging.Log4j2ConfigurationTemplate,
	}

	if logging.Log4jLoggers != nil {
		loggers := make([]interface{}, len(logging.Log4jLoggers))
		for i, logger := range logging.Log4jLoggers {
			loggers[i] = map[string]interface{}{
				"logger_name":  logger.LoggerName,
				"logger_level": logger.LoggerLevel,
			}
		}
		loggingMap["log4j_loggers"] = loggers
	}

	if logging.LogReservePolicy != nil {
		policyMap := map[string]interface{}{
			"expiration_days": logging.LogReservePolicy.ExpirationDays,
			"open_history":    logging.LogReservePolicy.OpenHistory,
		}
		loggingMap["log_reserve_policy"] = []interface{}{policyMap}
	}

	return []interface{}{loggingMap}
}
