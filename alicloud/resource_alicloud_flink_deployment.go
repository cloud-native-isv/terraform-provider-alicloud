package alicloud

import (
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"execution_mode": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"STREAMING", "BATCH"}, false),
			},
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
						"mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "PER_JOB",
							ValidateFunc: validation.StringInSlice([]string{"PER_JOB", "SESSION"}, false),
						},
					},
				},
			},
			"artifact": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kind": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{"JAR", "PYTHON", "SQLSCRIPT"}, false),
						},
						"jar_artifact": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"jar_uri": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"entry_class": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"main_args": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"additional_dependencies": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"python_artifact": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"python_artifact_uri": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"entry_module": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"main_args": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"additional_dependencies": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"additional_python_libraries": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"additional_python_archives": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"sql_artifact": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"sql_script": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"additional_dependencies": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
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
			"flink_conf": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"user_flink_conf": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
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
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_id": {
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
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
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

	deployment := &aliyunFlinkAPI.Deployment{
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
			deployment.DeploymentTarget = &aliyunFlinkAPI.DeploymentTarget{
				Name: targetMap["name"].(string),
			}
			if mode, ok := targetMap["mode"]; ok {
				deployment.DeploymentTarget.Mode = mode.(string)
			}
		}
	}

	if artifactConfig, ok := d.GetOk("artifact"); ok {
		artifactList := artifactConfig.([]interface{})
		if len(artifactList) > 0 {
			artifactMap := artifactList[0].(map[string]interface{})
			artifact := &aliyunFlinkAPI.Artifact{
				Kind: artifactMap["kind"].(string),
			}

			switch artifact.Kind {
			case "JAR":
				if jarArtifactList, ok := artifactMap["jar_artifact"].([]interface{}); ok && len(jarArtifactList) > 0 {
					jarArtifactMap := jarArtifactList[0].(map[string]interface{})
					jarArtifact := &aliyunFlinkAPI.JarArtifact{
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
					pythonArtifact := &aliyunFlinkAPI.PythonArtifact{
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
					sqlArtifact := &aliyunFlinkAPI.SqlArtifact{
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

	streamingResourceSetting := resourceAliCloudFlinkDeploymentExpandStreamingResourceSetting(d)
	if streamingResourceSetting != nil {
		deployment.StreamingResourceSetting = streamingResourceSetting
	}

	// Only use user_flink_conf for FlinkConf in Create operation
	if userFlinkConf, ok := d.GetOk("user_flink_conf"); ok {
		deployment.FlinkConf = make(map[string]string)
		for k, v := range userFlinkConf.(map[string]interface{}) {
			deployment.FlinkConf[k] = v.(string)
		}
	}

	logging := aliyunFlinkAPI.ExpandLogging(d.Get("logging").([]interface{}))
	if logging != nil {
		deployment.Logging = logging
	}

	if tags, ok := d.GetOk("tags"); ok {
		deployment.Labels = make(map[string]string)
		for k, v := range tags.(map[string]interface{}) {
			deployment.Labels[k] = v.(string)
		}
	}

	tempId := fmt.Sprintf("%s:%s:temp", workspaceId, namespaceName)
	newDeployment, err := flinkService.CreateDeployment(tempId, deployment)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_deployment", "CreateDeployment", AlibabaCloudSdkGoERROR)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", workspaceId, namespaceName, newDeployment.DeploymentId))
	d.Set("deployment_id", newDeployment.DeploymentId)

	stateConf := BuildStateConf([]string{"NotFound"}, []string{"CREATED"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, flinkService.FlinkDeploymentStateRefreshFunc(d.Id(), []string{"FAILED"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

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
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	if deployment == nil {
		d.SetId("")
		return nil
	}

	d.Set("namespace_name", deployment.Namespace)
	d.Set("workspace_id", deployment.Workspace)
	d.Set("name", deployment.Name)
	d.Set("deployment_id", deployment.DeploymentId)
	d.Set("description", deployment.Description)
	d.Set("engine_version", deployment.EngineVersion)
	d.Set("execution_mode", deployment.ExecutionMode)
	d.Set("status", deployment.Status)
	d.Set("creator", deployment.Creator)
	d.Set("creator_name", deployment.CreatorName)
	d.Set("modifier", deployment.Modifier)
	d.Set("modifier_name", deployment.ModifierName)

	if deployment.CreatedAt > 0 {
		d.Set("create_time", fmt.Sprintf("%d", deployment.CreatedAt))
	}
	if deployment.ModifiedAt > 0 {
		d.Set("update_time", fmt.Sprintf("%d", deployment.ModifiedAt))
	}
	if deployment.ReferencedDeploymentDraftId != "" {
		d.Set("referenced_deployment_draft_id", deployment.ReferencedDeploymentDraftId)
	}

	if deployment.DeploymentTarget != nil {
		deploymentTargetMap := map[string]interface{}{
			"name": deployment.DeploymentTarget.Name,
		}
		if deployment.DeploymentTarget.Mode != "" {
			deploymentTargetMap["mode"] = deployment.DeploymentTarget.Mode
		}
		d.Set("deployment_target", []interface{}{deploymentTargetMap})
	}

	if deployment.Artifact != nil {
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

	if deployment.StreamingResourceSetting != nil {
		d.Set("streaming_resource_setting", resourceAliCloudFlinkDeploymentFlattenStreamingResourceSetting(deployment.StreamingResourceSetting))
	}

	if deployment.FlinkConf != nil && len(deployment.FlinkConf) > 0 {
		flinkConf := make(map[string]interface{})
		for k, v := range deployment.FlinkConf {
			flinkConf[k] = v
		}
		d.Set("flink_conf", flinkConf)
	}

	// Only set user_flink_conf if it already has a value in state
	// This preserves the user's configuration without overwriting it
	if _, ok := d.GetOk("user_flink_conf"); ok {
		// We don't update user_flink_conf from API response since it's user-provided
		// and should be preserved as is in the state
	}

	if deployment.Logging != nil {
		d.Set("logging", aliyunFlinkAPI.FlattenLogging(deployment.Logging))
	}

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

	existingDeployment, err := flinkService.GetDeployment(d.Id())
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "GetDeployment", AlibabaCloudSdkGoERROR)
	}

	updateRequest := &aliyunFlinkAPI.Deployment{
		DeploymentId:             existingDeployment.DeploymentId,
		Workspace:                existingDeployment.Workspace,
		Namespace:                existingDeployment.Namespace,
		Name:                     existingDeployment.Name,
		Description:              existingDeployment.Description,
		EngineVersion:            existingDeployment.EngineVersion,
		ExecutionMode:            existingDeployment.ExecutionMode,
		DeploymentTarget:         existingDeployment.DeploymentTarget,
		Artifact:                 existingDeployment.Artifact,
		StreamingResourceSetting: existingDeployment.StreamingResourceSetting,
		FlinkConf:                existingDeployment.FlinkConf,
		Logging:                  existingDeployment.Logging,
		Labels:                   existingDeployment.Labels,
	}

	update := false

	if d.HasChange("name") {
		updateRequest.Name = d.Get("name").(string)
		update = true
	}

	if d.HasChange("description") {
		if description, ok := d.GetOk("description"); ok {
			updateRequest.Description = description.(string)
		} else {
			updateRequest.Description = ""
		}
		update = true
	}

	if d.HasChange("engine_version") {
		updateRequest.EngineVersion = d.Get("engine_version").(string)
		update = true
	}

	if d.HasChange("execution_mode") {
		updateRequest.ExecutionMode = d.Get("execution_mode").(string)
		update = true
	}

	if d.HasChange("deployment_target") {
		if deploymentTargetList, ok := d.GetOk("deployment_target"); ok {
			targets := deploymentTargetList.([]interface{})
			if len(targets) > 0 {
				targetMap := targets[0].(map[string]interface{})
				updateRequest.DeploymentTarget = &aliyunFlinkAPI.DeploymentTarget{
					Name: targetMap["name"].(string),
				}
				if mode, ok := targetMap["mode"]; ok {
					updateRequest.DeploymentTarget.Mode = mode.(string)
				}
			} else {
				updateRequest.DeploymentTarget = nil
			}
		} else {
			updateRequest.DeploymentTarget = nil
		}
		update = true
	}

	if d.HasChange("artifact") || d.HasChange("artifact.0.kind") || d.HasChange("artifact.0.jar_artifact") || d.HasChange("artifact.0.python_artifact") || d.HasChange("artifact.0.sql_artifact") || d.HasChange("artifact.0.jar_artifact.0.jar_uri") || d.HasChange("artifact.0.jar_artifact.0.entry_class") || d.HasChange("artifact.0.jar_artifact.0.main_args") || d.HasChange("artifact.0.jar_artifact.0.additional_dependencies") || d.HasChange("artifact.0.python_artifact.0.python_artifact_uri") || d.HasChange("artifact.0.python_artifact.0.entry_module") || d.HasChange("artifact.0.python_artifact.0.main_args") || d.HasChange("artifact.0.python_artifact.0.additional_dependencies") || d.HasChange("artifact.0.python_artifact.0.additional_python_libraries") || d.HasChange("artifact.0.python_artifact.0.additional_python_archives") || d.HasChange("artifact.0.sql_artifact.0.sql_script") || d.HasChange("artifact.0.sql_artifact.0.additional_dependencies") {
		if artifactConfig, ok := d.GetOk("artifact"); ok {
			artifactList := artifactConfig.([]interface{})
			if len(artifactList) > 0 {
				artifactMap := artifactList[0].(map[string]interface{})
				artifact := &aliyunFlinkAPI.Artifact{
					Kind: artifactMap["kind"].(string),
				}

				switch artifact.Kind {
				case "JAR":
					if jarArtifactList, ok := artifactMap["jar_artifact"].([]interface{}); ok && len(jarArtifactList) > 0 {
						jarArtifactMap := jarArtifactList[0].(map[string]interface{})
						jarArtifact := &aliyunFlinkAPI.JarArtifact{
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
						pythonArtifact := &aliyunFlinkAPI.PythonArtifact{
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
						sqlArtifact := &aliyunFlinkAPI.SqlArtifact{
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
				updateRequest.Artifact = artifact
			} else {
				updateRequest.Artifact = nil
			}
		} else {
			updateRequest.Artifact = nil
		}
		update = true
	}

	if d.HasChange("streaming_resource_setting") {
		streamingResourceSetting := resourceAliCloudFlinkDeploymentExpandStreamingResourceSetting(d)
		if streamingResourceSetting != nil {
			updateRequest.StreamingResourceSetting = streamingResourceSetting
		} else {
			updateRequest.StreamingResourceSetting = nil
		}
		update = true
	}

	if d.HasChange("user_flink_conf") {
		// Initialize FlinkConf map
		updateRequest.FlinkConf = make(map[string]string)

		// First, get the existing flink_conf from state
		if flinkConfState, ok := d.GetOk("flink_conf"); ok {
			for k, v := range flinkConfState.(map[string]interface{}) {
				updateRequest.FlinkConf[k] = v.(string)
			}
		}

		// Then, add or override with user_flink_conf
		if userFlinkConf, ok := d.GetOk("user_flink_conf"); ok {
			for k, v := range userFlinkConf.(map[string]interface{}) {
				updateRequest.FlinkConf[k] = v.(string)
			}
		}
		update = true
	}

	if d.HasChange("logging") {
		logging := aliyunFlinkAPI.ExpandLogging(d.Get("logging").([]interface{}))
		if logging != nil {
			updateRequest.Logging = logging
		} else {
			updateRequest.Logging = nil
		}
		update = true
	}

	if d.HasChange("tags") {
		if tags, ok := d.GetOk("tags"); ok {
			updateRequest.Labels = make(map[string]string)
			for k, v := range tags.(map[string]interface{}) {
				updateRequest.Labels[k] = v.(string)
			}
		} else {
			updateRequest.Labels = nil
		}
		update = true
	}

	if update {
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err := flinkService.UpdateDeployment(d.Id(), updateRequest)
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
	}

	return resourceAliCloudFlinkDeploymentRead(d, meta)
}

func resourceAliCloudFlinkDeploymentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	// 在删除部署之前，先等待相关的Job进入终端状态
	// 这可以解决"Deleting a deployment which has job is not terminal status is not allowed"的错误
	err = flinkService.WaitForDeploymentJobsTerminal(d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "WaitForDeploymentJobsTerminal", AlibabaCloudSdkGoERROR)
	}

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := flinkService.DeleteDeployment(d.Id())
		if err != nil {
			if NotFoundError(err) {
				return nil
			}
			return resource.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteDeployment", AlibabaCloudSdkGoERROR)
	}

	stateConf := BuildStateConf([]string{"CREATED"}, []string{}, d.Timeout(schema.TimeoutDelete), 5*time.Second, flinkService.FlinkDeploymentStateRefreshFunc(d.Id(), []string{"Failed"}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}

func resourceAliCloudFlinkDeploymentExpandArtifact(d *schema.ResourceData) (*aliyunFlinkAPI.Artifact, error) {
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
	return nil, nil
}

func resourceAliCloudFlinkDeploymentExpandStreamingResourceSetting(d *schema.ResourceData) *aliyunFlinkAPI.StreamingResourceSetting {
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
	return nil
}

func resourceAliCloudFlinkDeploymentExpandLogging(d *schema.ResourceData) *aliyunFlinkAPI.Logging {
	if loggingList, ok := d.GetOk("logging"); ok {
		loggings := loggingList.([]interface{})
		if len(loggings) > 0 {
			loggingMap := loggings[0].(map[string]interface{})
			logging := &aliyunFlinkAPI.Logging{}

			if profile, ok := loggingMap["logging_profile"]; ok && profile != nil {
				if profileStr, ok := profile.(string); ok && profileStr != "" {
					logging.LoggingProfile = profileStr
				}
			}

			if template, ok := loggingMap["log4j2_configuration_template"]; ok && template != nil {
				if templateStr, ok := template.(string); ok && templateStr != "" {
					logging.Log4j2ConfigurationTemplate = templateStr
				}
			}

			if loggersList, ok := loggingMap["log4j_loggers"]; ok && loggersList != nil {
				if loggersSlice, ok := loggersList.([]interface{}); ok && len(loggersSlice) > 0 {
					logging.Log4jLoggers = make([]aliyunFlinkAPI.Log4jLogger, len(loggersSlice))
					for i, loggerItem := range loggersSlice {
						if loggerItem != nil {
							loggerMap := loggerItem.(map[string]interface{})
							logging.Log4jLoggers[i] = aliyunFlinkAPI.Log4jLogger{
								LoggerName:  loggerMap["logger_name"].(string),
								LoggerLevel: loggerMap["logger_level"].(string),
							}
						}
					}
				}
			}

			if reservePolicyList, ok := loggingMap["log_reserve_policy"]; ok && reservePolicyList != nil {
				if policySlice, ok := reservePolicyList.([]interface{}); ok && len(policySlice) > 0 && policySlice[0] != nil {
					policyMap := policySlice[0].(map[string]interface{})
					logging.LogReservePolicy = &aliyunFlinkAPI.LogReservePolicy{
						ExpirationDays: policyMap["expiration_days"].(int),
						OpenHistory:    policyMap["open_history"].(bool),
					}
				}
			}

			return logging
		}
	}
	return nil
}

func resourceAliCloudFlinkDeploymentFlattenArtifact(artifact *aliyunFlinkAPI.Artifact) []interface{} {
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

func resourceAliCloudFlinkDeploymentFlattenStreamingResourceSetting(setting *aliyunFlinkAPI.StreamingResourceSetting) []interface{} {
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

func resourceAliCloudFlinkDeploymentflattenLogging(logging *aliyunFlinkAPI.Logging) []interface{} {
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
