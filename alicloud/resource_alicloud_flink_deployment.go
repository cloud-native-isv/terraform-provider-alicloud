package alicloud

import (
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	flinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The workspace ID of the Flink workspace.",
			},
			"namespace_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the namespace within the workspace.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Flink deployment.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the Flink deployment.",
			},
			"engine_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "vvr-11.1-jdk11-flink-1.20",
				Description: "The version of Flink engine to use for the deployment.",
			},
			"execution_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "STREAMING",
				ValidateFunc: validation.StringInSlice([]string{"STREAMING", "BATCH"}, false),
				Description:  "The execution mode for the Flink job.",
			},
			"deployment_target": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Deployment target configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "session-cluster",
							Description: "The name of the deployment target.",
						},
					},
				},
			},
			"artifact": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Extended artifact configuration. If specified, takes precedence over artifact_uri.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kind": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"JAR", "PYTHON", "SQLSCRIPT"}, false),
							Description:  "The type of artifact.",
						},
						"jar_artifact": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "JAR artifact configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"jar_uri": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The URI of the JAR file.",
									},
									"entry_class": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The main class of the JAR.",
									},
									"main_args": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Main method arguments.",
									},
									"additional_dependencies": {
										Type:        schema.TypeList,
										Optional:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Description: "Additional dependencies.",
									},
								},
							},
						},
						"python_artifact": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Python artifact configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"python_artifact_uri": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The URI of the Python artifact.",
									},
									"entry_module": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The entry module.",
									},
									"main_args": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Main method arguments.",
									},
									"additional_dependencies": {
										Type:        schema.TypeList,
										Optional:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Description: "Additional dependencies.",
									},
									"additional_python_libraries": {
										Type:        schema.TypeList,
										Optional:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Description: "Additional Python libraries.",
									},
									"additional_python_archives": {
										Type:        schema.TypeList,
										Optional:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Description: "Additional Python archives.",
									},
								},
							},
						},
						"sql_artifact": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "SQL artifact configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"sql_script": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The SQL script content.",
									},
									"additional_dependencies": {
										Type:        schema.TypeList,
										Optional:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Description: "Additional dependencies.",
									},
								},
							},
						},
					},
				},
			},
			"parallelism": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The parallelism of the job.",
			},
			"streaming_resource_setting": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Resource configuration for streaming jobs.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_setting_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "BASIC",
							ValidateFunc: validation.StringInSlice([]string{"BASIC", "EXPERT"}, false),
							Description:  "The resource setting mode.",
						},
						"basic_resource_setting": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Basic resource setting configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"parallelism": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "The parallelism for basic resource setting.",
									},
									"jobmanager_resource_setting_spec": {
										Type:        schema.TypeList,
										Optional:    true,
										MaxItems:    1,
										Description: "JobManager resource specification.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cpu": {
													Type:        schema.TypeFloat,
													Optional:    true,
													Description: "CPU cores for JobManager.",
												},
												"memory": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Memory allocation for JobManager.",
												},
											},
										},
									},
									"taskmanager_resource_setting_spec": {
										Type:        schema.TypeList,
										Optional:    true,
										MaxItems:    1,
										Description: "TaskManager resource specification.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cpu": {
													Type:        schema.TypeFloat,
													Optional:    true,
													Description: "CPU cores for TaskManager.",
												},
												"memory": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Memory allocation for TaskManager.",
												},
											},
										},
									},
								},
							},
						},
						"expert_resource_setting": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Expert resource setting configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_plan": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The resource plan for expert mode.",
									},
									"jobmanager_resource_setting_spec": {
										Type:        schema.TypeList,
										Optional:    true,
										MaxItems:    1,
										Description: "JobManager resource specification for expert mode.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cpu": {
													Type:        schema.TypeFloat,
													Optional:    true,
													Description: "CPU cores for JobManager.",
												},
												"memory": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Memory allocation for JobManager.",
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
			"flink_conf": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Flink configuration key-value pairs.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"logging": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Logging configuration for the deployment.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"logging_profile": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The logging profile to use.",
						},
						"log4j2_configuration_template": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The Log4j2 configuration template.",
						},
						"log4j_loggers": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of Log4j loggers configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"logger_name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The name of the logger.",
									},
									"logger_level": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The log level for the logger.",
									},
								},
							},
						},
						"log_reserve_policy": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Log retention policy configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"expiration_days": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Number of days to retain logs.",
									},
									"open_history": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Whether to enable history logs.",
									},
								},
							},
						},
					},
				},
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Tags to assign to the deployment.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the deployment.",
			},
			"update_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last update time of the deployment.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current status of the deployment.",
			},
			"deployment_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique ID of the deployment.",
			},
			"creator": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creator of the deployment.",
			},
			"creator_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the deployment creator.",
			},
			"modifier": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last modifier of the deployment.",
			},
			"modifier_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the last modifier.",
			},
			"referenced_deployment_draft_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the referenced deployment draft.",
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

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)
	name := d.Get("name").(string)

	deployment := &flinkAPI.Deployment{
		Workspace: workspaceId,
		Namespace: namespaceName,
		Name:      name,
	}

	if description, ok := d.GetOk("description"); ok {
		deployment.Description = description.(string)
	}

	if engineVersion, ok := d.GetOk("engine_version"); ok {
		deployment.EngineVersion = engineVersion.(string)
	}

	if executionMode, ok := d.GetOk("execution_mode"); ok {
		deployment.ExecutionMode = executionMode.(string)
	}

	if deploymentTargetList, ok := d.GetOk("deployment_target"); ok {
		targets := deploymentTargetList.([]interface{})
		if len(targets) > 0 {
			targetMap := targets[0].(map[string]interface{})
			deployment.DeploymentTarget = &flinkAPI.DeploymentTarget{
				Name: targetMap["name"].(string),
			}
		}
	}

	// Handle artifact configuration
	if artifactConfig, ok := d.GetOk("artifact"); ok {
		artifactList := artifactConfig.([]interface{})
		if len(artifactList) > 0 {
			artifactMap := artifactList[0].(map[string]interface{})
			artifact := &flinkAPI.Artifact{
				Kind: artifactMap["kind"].(string),
			}

			switch artifact.Kind {
			case "JAR":
				if jarArtifactList, ok := artifactMap["jar_artifact"].([]interface{}); ok && len(jarArtifactList) > 0 {
					jarArtifactMap := jarArtifactList[0].(map[string]interface{})
					jarArtifact := &flinkAPI.JarArtifact{
						JarUri: jarArtifactMap["jar_uri"].(string),
					}
					if entryClass, ok := jarArtifactMap["entry_class"]; ok {
						jarArtifact.EntryClass = entryClass.(string)
					}
					if mainArgs, ok := jarArtifactMap["main_args"]; ok {
						jarArtifact.MainArgs = mainArgs.(string)
					}
					if additionalDeps, ok := jarArtifactMap["additional_dependencies"]; ok {
						deps := additionalDeps.([]interface{})
						jarArtifact.AdditionalDependencies = make([]string, len(deps))
						for i, dep := range deps {
							jarArtifact.AdditionalDependencies[i] = dep.(string)
						}
					}
					artifact.JarArtifact = jarArtifact
				}
			case "PYTHON":
				if pythonArtifactList, ok := artifactMap["python_artifact"].([]interface{}); ok && len(pythonArtifactList) > 0 {
					pythonArtifactMap := pythonArtifactList[0].(map[string]interface{})
					pythonArtifact := &flinkAPI.PythonArtifact{
						PythonArtifactUri: pythonArtifactMap["python_artifact_uri"].(string),
					}
					if entryModule, ok := pythonArtifactMap["entry_module"]; ok {
						pythonArtifact.EntryModule = entryModule.(string)
					}
					if mainArgs, ok := pythonArtifactMap["main_args"]; ok {
						pythonArtifact.MainArgs = mainArgs.(string)
					}
					if additionalDeps, ok := pythonArtifactMap["additional_dependencies"]; ok {
						deps := additionalDeps.([]interface{})
						pythonArtifact.AdditionalDependencies = make([]string, len(deps))
						for i, dep := range deps {
							pythonArtifact.AdditionalDependencies[i] = dep.(string)
						}
					}
					if additionalLibs, ok := pythonArtifactMap["additional_python_libraries"]; ok {
						libs := additionalLibs.([]interface{})
						pythonArtifact.AdditionalPythonLibraries = make([]string, len(libs))
						for i, lib := range libs {
							pythonArtifact.AdditionalPythonLibraries[i] = lib.(string)
						}
					}
					if additionalArchives, ok := pythonArtifactMap["additional_python_archives"]; ok {
						archives := additionalArchives.([]interface{})
						pythonArtifact.AdditionalPythonArchives = make([]string, len(archives))
						for i, archive := range archives {
							pythonArtifact.AdditionalPythonArchives[i] = archive.(string)
						}
					}
					artifact.PythonArtifact = pythonArtifact
				}
			case "SQLSCRIPT":
				if sqlArtifactList, ok := artifactMap["sql_artifact"].([]interface{}); ok && len(sqlArtifactList) > 0 {
					sqlArtifactMap := sqlArtifactList[0].(map[string]interface{})
					sqlArtifact := &flinkAPI.SqlArtifact{
						SqlScript: sqlArtifactMap["sql_script"].(string),
					}
					if additionalDeps, ok := sqlArtifactMap["additional_dependencies"]; ok {
						deps := additionalDeps.([]interface{})
						sqlArtifact.AdditionalDependencies = make([]string, len(deps))
						for i, dep := range deps {
							sqlArtifact.AdditionalDependencies[i] = dep.(string)
						}
					}
					artifact.SqlArtifact = sqlArtifact
				}
			}
			deployment.Artifact = artifact
		}
	}

	// Handle streaming resource setting
	streamingResourceSetting := expandStreamingResourceSetting(d)
	if streamingResourceSetting != nil {
		deployment.StreamingResourceSetting = streamingResourceSetting
	}

	// Handle Flink configuration
	if flinkConf, ok := d.GetOk("flink_conf"); ok {
		deployment.FlinkConf = make(map[string]string)
		for k, v := range flinkConf.(map[string]interface{}) {
			deployment.FlinkConf[k] = v.(string)
		}
	}

	// Handle logging configuration
	logging := expandLogging(d)
	if logging != nil {
		// Convert Logging to LoggingProfile for deployment
		deployment.Logging = &flinkAPI.LoggingProfile{
			Template: logging.LoggingProfile,
		}
	}

	// Handle labels/tags
	if tags, ok := d.GetOk("tags"); ok {
		deployment.Labels = make(map[string]string)
		for k, v := range tags.(map[string]interface{}) {
			deployment.Labels[k] = v.(string)
		}
	}

	newDeployment, err := flinkService.CreateDeployment(&namespaceName, deployment)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_deployment", "CreateDeployment", AlibabaCloudSdkGoERROR)
	}
	addDebugJson("Deployment", newDeployment)

	d.SetId(fmt.Sprintf("%s:%s", namespaceName, newDeployment.DeploymentId))
	d.Set("deployment_id", newDeployment.DeploymentId)

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
		if IsExpectedErrors(err, []string{"DeploymentNotFound"}) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	if deployment == nil {
		d.SetId("")
		return nil
	}

	// Set basic fields
	d.Set("namespace_name", deployment.Namespace)
	d.Set("workspace_id", deployment.Workspace)
	d.Set("name", deployment.Name)
	d.Set("deployment_id", deployment.DeploymentId)
	d.Set("description", deployment.Description)
	d.Set("engine_version", deployment.EngineVersion)
	d.Set("execution_mode", deployment.ExecutionMode)
	d.Set("status", deployment.Status)

	// Set creator information
	d.Set("creator", deployment.Creator)
	d.Set("creator_name", deployment.CreatorName)
	d.Set("modifier", deployment.Modifier)
	d.Set("modifier_name", deployment.ModifierName)

	// Set timestamps
	if deployment.CreatedAt > 0 {
		d.Set("create_time", fmt.Sprintf("%d", deployment.CreatedAt))
	}
	if deployment.ModifiedAt > 0 {
		d.Set("update_time", fmt.Sprintf("%d", deployment.ModifiedAt))
	}

	// Set referenced deployment draft ID if present
	if deployment.ReferencedDeploymentDraftId != "" {
		d.Set("referenced_deployment_draft_id", deployment.ReferencedDeploymentDraftId)
	}

	// Set deployment target
	if deployment.DeploymentTarget != nil {
		deploymentTargetMap := map[string]interface{}{
			"name": deployment.DeploymentTarget.Name,
		}
		d.Set("deployment_target", []interface{}{deploymentTargetMap})
	}

	// Set artifact configuration
	if deployment.Artifact != nil {
		// Handle complex artifact configuration
		artifactConfig := make([]map[string]interface{}, 0, 1)
		artifactMap := map[string]interface{}{
			"kind": deployment.Artifact.Kind,
		}

		switch deployment.Artifact.Kind {
		case "JAR":
			if deployment.Artifact.JarArtifact != nil {
				jarArtifact := make([]map[string]interface{}, 0, 1)
				jarMap := map[string]interface{}{
					"jar_uri": deployment.Artifact.JarArtifact.JarUri,
				}
				if deployment.Artifact.JarArtifact.EntryClass != "" {
					jarMap["entry_class"] = deployment.Artifact.JarArtifact.EntryClass
				}
				if deployment.Artifact.JarArtifact.MainArgs != "" {
					jarMap["main_args"] = deployment.Artifact.JarArtifact.MainArgs
				}
				if len(deployment.Artifact.JarArtifact.AdditionalDependencies) > 0 {
					jarMap["additional_dependencies"] = deployment.Artifact.JarArtifact.AdditionalDependencies
				}
				jarArtifact = append(jarArtifact, jarMap)
				artifactMap["jar_artifact"] = jarArtifact
			}
		case "PYTHON":
			if deployment.Artifact.PythonArtifact != nil {
				pythonArtifact := make([]map[string]interface{}, 0, 1)
				pythonMap := map[string]interface{}{
					"python_artifact_uri": deployment.Artifact.PythonArtifact.PythonArtifactUri,
				}
				if deployment.Artifact.PythonArtifact.EntryModule != "" {
					pythonMap["entry_module"] = deployment.Artifact.PythonArtifact.EntryModule
				}
				if deployment.Artifact.PythonArtifact.MainArgs != "" {
					pythonMap["main_args"] = deployment.Artifact.PythonArtifact.MainArgs
				}
				if len(deployment.Artifact.PythonArtifact.AdditionalDependencies) > 0 {
					pythonMap["additional_dependencies"] = deployment.Artifact.PythonArtifact.AdditionalDependencies
				}
				if len(deployment.Artifact.PythonArtifact.AdditionalPythonLibraries) > 0 {
					pythonMap["additional_python_libraries"] = deployment.Artifact.PythonArtifact.AdditionalPythonLibraries
				}
				if len(deployment.Artifact.PythonArtifact.AdditionalPythonArchives) > 0 {
					pythonMap["additional_python_archives"] = deployment.Artifact.PythonArtifact.AdditionalPythonArchives
				}
				pythonArtifact = append(pythonArtifact, pythonMap)
				artifactMap["python_artifact"] = pythonArtifact
			}
		case "SQLSCRIPT":
			if deployment.Artifact.SqlArtifact != nil {
				sqlArtifact := make([]map[string]interface{}, 0, 1)
				sqlMap := map[string]interface{}{
					"sql_script": deployment.Artifact.SqlArtifact.SqlScript,
				}
				if len(deployment.Artifact.SqlArtifact.AdditionalDependencies) > 0 {
					sqlMap["additional_dependencies"] = deployment.Artifact.SqlArtifact.AdditionalDependencies
				}
				sqlArtifact = append(sqlArtifact, sqlMap)
				artifactMap["sql_artifact"] = sqlArtifact
			}
		}

		artifactConfig = append(artifactConfig, artifactMap)
		d.Set("artifact", artifactConfig)
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
	if deployment.FlinkConf != nil && len(deployment.FlinkConf) > 0 {
		flinkConf := make(map[string]interface{})
		for k, v := range deployment.FlinkConf {
			flinkConf[k] = v
		}
		d.Set("flink_conf", flinkConf)
	}

	// Set logging configuration
	if deployment.Logging != nil {
		// Convert LoggingProfile to Logging for flattening
		logging := &flinkAPI.Logging{
			LoggingProfile: deployment.Logging.Template,
		}
		d.Set("logging", flattenLogging(logging))
	}

	// Set labels/tags
	if deployment.Labels != nil && len(deployment.Labels) > 0 {
		tags := make(map[string]interface{})
		for k, v := range deployment.Labels {
			tags[k] = v
		}
		d.Set("tags", tags)
	}

	return nil
}

func resourceAliCloudFlinkDeploymentUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	deployment, err := flinkService.GetDeployment(d.Id())
	if err != nil {
		return WrapError(err)
	}

	update := false

	if d.HasChange("name") {
		deployment.Name = d.Get("name").(string)
		update = true
	}

	if d.HasChange("description") {
		deployment.Description = d.Get("description").(string)
		update = true
	}

	if d.HasChange("engine_version") {
		deployment.EngineVersion = d.Get("engine_version").(string)
		update = true
	}

	if d.HasChange("execution_mode") {
		deployment.ExecutionMode = d.Get("execution_mode").(string)
		update = true
	}

	// Update deployment target
	if d.HasChange("deployment_target") {
		if deploymentTargetList, ok := d.GetOk("deployment_target"); ok {
			targets := deploymentTargetList.([]interface{})
			if len(targets) > 0 {
				targetMap := targets[0].(map[string]interface{})
				deployment.DeploymentTarget = &flinkAPI.DeploymentTarget{
					Name: targetMap["name"].(string),
				}
			}
		} else {
			deployment.DeploymentTarget = nil
		}
		update = true
	}

	// Update artifact configuration
	if d.HasChange("artifact") {
		if artifactConfig, ok := d.GetOk("artifact"); ok {
			artifactList := artifactConfig.([]interface{})
			if len(artifactList) > 0 {
				artifactMap := artifactList[0].(map[string]interface{})
				artifact := &flinkAPI.Artifact{
					Kind: artifactMap["kind"].(string),
				}

				switch artifact.Kind {
				case "JAR":
					if jarArtifactList, ok := artifactMap["jar_artifact"].([]interface{}); ok && len(jarArtifactList) > 0 {
						jarArtifactMap := jarArtifactList[0].(map[string]interface{})
						jarArtifact := &flinkAPI.JarArtifact{
							JarUri: jarArtifactMap["jar_uri"].(string),
						}
						if entryClass, ok := jarArtifactMap["entry_class"]; ok {
							jarArtifact.EntryClass = entryClass.(string)
						}
						if mainArgs, ok := jarArtifactMap["main_args"]; ok {
							jarArtifact.MainArgs = mainArgs.(string)
						}
						if additionalDeps, ok := jarArtifactMap["additional_dependencies"]; ok {
							deps := additionalDeps.([]interface{})
							jarArtifact.AdditionalDependencies = make([]string, len(deps))
							for i, dep := range deps {
								jarArtifact.AdditionalDependencies[i] = dep.(string)
							}
						}
						artifact.JarArtifact = jarArtifact
					}
				case "PYTHON":
					if pythonArtifactList, ok := artifactMap["python_artifact"].([]interface{}); ok && len(pythonArtifactList) > 0 {
						pythonArtifactMap := pythonArtifactList[0].(map[string]interface{})
						pythonArtifact := &flinkAPI.PythonArtifact{
							PythonArtifactUri: pythonArtifactMap["python_artifact_uri"].(string),
						}
						if entryModule, ok := pythonArtifactMap["entry_module"]; ok {
							pythonArtifact.EntryModule = entryModule.(string)
						}
						if mainArgs, ok := pythonArtifactMap["main_args"]; ok {
							pythonArtifact.MainArgs = mainArgs.(string)
						}
						if additionalDeps, ok := pythonArtifactMap["additional_dependencies"]; ok {
							deps := additionalDeps.([]interface{})
							pythonArtifact.AdditionalDependencies = make([]string, len(deps))
							for i, dep := range deps {
								pythonArtifact.AdditionalDependencies[i] = dep.(string)
							}
						}
						if additionalLibs, ok := pythonArtifactMap["additional_python_libraries"]; ok {
							libs := additionalLibs.([]interface{})
							pythonArtifact.AdditionalPythonLibraries = make([]string, len(libs))
							for i, lib := range libs {
								pythonArtifact.AdditionalPythonLibraries[i] = lib.(string)
							}
						}
						if additionalArchives, ok := pythonArtifactMap["additional_python_archives"]; ok {
							archives := additionalArchives.([]interface{})
							pythonArtifact.AdditionalPythonArchives = make([]string, len(archives))
							for i, archive := range archives {
								pythonArtifact.AdditionalPythonArchives[i] = archive.(string)
							}
						}
						artifact.PythonArtifact = pythonArtifact
					}
				case "SQLSCRIPT":
					if sqlArtifactList, ok := artifactMap["sql_artifact"].([]interface{}); ok && len(sqlArtifactList) > 0 {
						sqlArtifactMap := sqlArtifactList[0].(map[string]interface{})
						sqlArtifact := &flinkAPI.SqlArtifact{
							SqlScript: sqlArtifactMap["sql_script"].(string),
						}
						if additionalDeps, ok := sqlArtifactMap["additional_dependencies"]; ok {
							deps := additionalDeps.([]interface{})
							sqlArtifact.AdditionalDependencies = make([]string, len(deps))
							for i, dep := range deps {
								sqlArtifact.AdditionalDependencies[i] = dep.(string)
							}
						}
						artifact.SqlArtifact = sqlArtifact
					}
				}
				deployment.Artifact = artifact
			}
		} else {
			deployment.Artifact = nil
		}
		update = true
	}

	// Update streaming resource setting
	if d.HasChange("streaming_resource_setting") || d.HasChange("parallelism") {
		streamingResourceSetting := expandStreamingResourceSetting(d)
		if streamingResourceSetting != nil {
			deployment.StreamingResourceSetting = streamingResourceSetting
		} else {
			deployment.StreamingResourceSetting = nil
		}
		update = true
	}

	// Update Flink configuration
	if d.HasChange("flink_conf") {
		if flinkConf, ok := d.GetOk("flink_conf"); ok {
			deployment.FlinkConf = make(map[string]string)
			for k, v := range flinkConf.(map[string]interface{}) {
				deployment.FlinkConf[k] = v.(string)
			}
		} else {
			deployment.FlinkConf = nil
		}
		update = true
	}

	// Update logging configuration
	if d.HasChange("logging") {
		logging := expandLogging(d)
		if logging != nil {
			// Convert Logging to LoggingProfile for deployment
			deployment.Logging = &flinkAPI.LoggingProfile{
				Template: logging.LoggingProfile,
			}
		} else {
			deployment.Logging = nil
		}
		update = true
	}

	// Update labels/tags
	if d.HasChange("tags") {
		if tags, ok := d.GetOk("tags"); ok {
			deployment.Labels = make(map[string]string)
			for k, v := range tags.(map[string]interface{}) {
				deployment.Labels[k] = v.(string)
			}
		} else {
			deployment.Labels = nil
		}
		update = true
	}

	if update {
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

		// Wait for update to complete using StateRefreshFunc
		stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, flinkService.FlinkDeploymentStateRefreshFunc(d.Id(), []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	// 最后调用Read同步状态
	return resourceAliCloudFlinkDeploymentRead(d, meta)
}

func resourceAliCloudFlinkDeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	parts, err := ParseResourceId(d.Id(), 2)
	if err != nil {
		return WrapError(err)
	}

	namespace := parts[0]
	deploymentID := parts[1]

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		// Use service method with correct parameters
		err := flinkService.DeleteDeployment(namespace, deploymentID)
		if err != nil {
			if IsExpectedErrors(err, []string{"DeploymentNotFound"}) {
				return nil
			}
			return resource.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteDeployment", AlibabaCloudSdkGoERROR)
	}

	return nil
}

// Helper functions for schema conversion

// expandArtifact converts schema artifact to API artifact
func expandArtifact(d *schema.ResourceData) (*flinkAPI.Artifact, error) {
	// Check for new artifact structure
	if artifactList, ok := d.GetOk("artifact"); ok {
		artifacts := artifactList.([]interface{})
		if len(artifacts) > 0 {
			artifactMap := artifacts[0].(map[string]interface{})
			artifact := &flinkAPI.Artifact{
				Kind: artifactMap["kind"].(string),
			}

			switch artifact.Kind {
			case "JAR":
				if jarArtifactList, ok := artifactMap["jar_artifact"].([]interface{}); ok && len(jarArtifactList) > 0 {
					jarMap := jarArtifactList[0].(map[string]interface{})
					artifact.JarArtifact = &flinkAPI.JarArtifact{
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
					artifact.PythonArtifact = &flinkAPI.PythonArtifact{
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
					artifact.SqlArtifact = &flinkAPI.SqlArtifact{
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

	return nil, nil
}

// expandStreamingResourceSetting converts schema streaming resource setting to API format
func expandStreamingResourceSetting(d *schema.ResourceData) *flinkAPI.StreamingResourceSetting {
	if streamingResourceList, ok := d.GetOk("streaming_resource_setting"); ok {
		settings := streamingResourceList.([]interface{})
		if len(settings) > 0 {
			settingMap := settings[0].(map[string]interface{})
			resourceSetting := &flinkAPI.StreamingResourceSetting{
				ResourceSettingMode: settingMap["resource_setting_mode"].(string),
			}

			if basicList, ok := settingMap["basic_resource_setting"].([]interface{}); ok && len(basicList) > 0 {
				basicMap := basicList[0].(map[string]interface{})
				resourceSetting.BasicResourceSetting = &flinkAPI.BasicResourceSetting{
					Parallelism: int64(basicMap["parallelism"].(int)),
				}

				if jmList, ok := basicMap["jobmanager_resource_setting_spec"].([]interface{}); ok && len(jmList) > 0 {
					jmMap := jmList[0].(map[string]interface{})
					resourceSetting.BasicResourceSetting.JobManagerResourceSettingSpec = &flinkAPI.BasicResourceSettingSpec{
						Cpu:    jmMap["cpu"].(float64),
						Memory: jmMap["memory"].(string),
					}
				}

				if tmList, ok := basicMap["taskmanager_resource_setting_spec"].([]interface{}); ok && len(tmList) > 0 {
					tmMap := tmList[0].(map[string]interface{})
					resourceSetting.BasicResourceSetting.TaskManagerResourceSettingSpec = &flinkAPI.BasicResourceSettingSpec{
						Cpu:    tmMap["cpu"].(float64),
						Memory: tmMap["memory"].(string),
					}
				}
			}

			if expertList, ok := settingMap["expert_resource_setting"].([]interface{}); ok && len(expertList) > 0 {
				expertMap := expertList[0].(map[string]interface{})
				resourceSetting.ExpertResourceSetting = &flinkAPI.ExpertResourceSetting{
					ResourcePlan: expertMap["resource_plan"].(string),
				}

				if jmList, ok := expertMap["jobmanager_resource_setting_spec"].([]interface{}); ok && len(jmList) > 0 {
					jmMap := jmList[0].(map[string]interface{})
					resourceSetting.ExpertResourceSetting.JobManagerResourceSettingSpec = &flinkAPI.BasicResourceSettingSpec{
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
		return &flinkAPI.StreamingResourceSetting{
			ResourceSettingMode: "BASIC",
			BasicResourceSetting: &flinkAPI.BasicResourceSetting{
				Parallelism: int64(parallelism.(int)),
			},
		}
	}

	return nil
}

// expandLogging converts schema logging configuration to API format
func expandLogging(d *schema.ResourceData) *flinkAPI.Logging {
	if loggingList, ok := d.GetOk("logging"); ok {
		loggings := loggingList.([]interface{})
		if len(loggings) > 0 {
			loggingMap := loggings[0].(map[string]interface{})
			logging := &flinkAPI.Logging{}

			if profile, ok := loggingMap["logging_profile"].(string); ok {
				logging.LoggingProfile = profile
			}

			if template, ok := loggingMap["log4j2_configuration_template"].(string); ok {
				logging.Log4j2ConfigurationTemplate = template
			}

			if loggersList, ok := loggingMap["log4j_loggers"].([]interface{}); ok {
				logging.Log4jLoggers = make([]flinkAPI.Log4jLogger, len(loggersList))
				for i, loggerItem := range loggersList {
					loggerMap := loggerItem.(map[string]interface{})
					logging.Log4jLoggers[i] = flinkAPI.Log4jLogger{
						LoggerName:  loggerMap["logger_name"].(string),
						LoggerLevel: loggerMap["logger_level"].(string),
					}
				}
			}

			if reservePolicyList, ok := loggingMap["log_reserve_policy"].([]interface{}); ok && len(reservePolicyList) > 0 {
				policyMap := reservePolicyList[0].(map[string]interface{})
				logging.LogReservePolicy = &flinkAPI.LogReservePolicy{
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
func flattenArtifact(artifact *flinkAPI.Artifact) []interface{} {
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
func flattenStreamingResourceSetting(setting *flinkAPI.StreamingResourceSetting) []interface{} {
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
func flattenLogging(logging *flinkAPI.Logging) []interface{} {
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
