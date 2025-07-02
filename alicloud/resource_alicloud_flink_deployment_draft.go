package alicloud

import (
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	flinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudFlinkDeploymentDraft() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFlinkDeploymentDraftCreate,
		Read:   resourceAliCloudFlinkDeploymentDraftRead,
		Update: resourceAliCloudFlinkDeploymentDraftUpdate,
		Delete: resourceAliCloudFlinkDeploymentDraftDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"workspace_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the Flink workspace.",
			},
			"namespace_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the namespace.",
			},
			"folder_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the folder where the deployment draft is located.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the deployment draft.",
			},
			"deployment_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Associated deployment ID for this draft.",
			},
			"engine_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "vvr-11.1-jdk11-flink-1.20",
				Description: "The Flink engine version.",
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
				Description: "Streaming resource setting configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_setting_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "BASIC",
							ValidateFunc: validation.StringInSlice([]string{"BASIC", "EXPERT"}, false),
							Description:  "Resource setting mode.",
						},
						"basic_resource_setting": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Basic resource setting.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"parallelism": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Job parallelism.",
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
													Description: "CPU cores.",
												},
												"memory": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Memory size.",
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
													Description: "CPU cores.",
												},
												"memory": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Memory size.",
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
							Description: "Expert resource setting.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_plan": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Resource plan configuration.",
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
													Description: "CPU cores.",
												},
												"memory": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "Memory size.",
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
				Description: "Flink configuration properties.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"logging": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Logging configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"logging_profile": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Logging profile.",
						},
						"log4j2_configuration_template": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Log4j2 configuration template.",
						},
						"log4j_loggers": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Log4j loggers configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"logger_name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Logger name.",
									},
									"logger_level": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Logger level.",
									},
								},
							},
						},
						"log_reserve_policy": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Log reserve policy.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"expiration_days": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "Log expiration days.",
									},
									"open_history": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Whether to open history logs.",
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
				Description: "Resource tags.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			// Draft 特有参数 - 环境变量
			"environment_variables": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Environment variables for the deployment draft.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the deployment draft.",
			},
			"update_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last update time of the deployment draft.",
			},
			"draft_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the deployment draft.",
			},
			"creator": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creator of the deployment draft.",
			},
			"creator_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creator name of the deployment draft.",
			},
			"modifier": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The modifier of the deployment draft.",
			},
			"modifier_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The modifier name of the deployment draft.",
			},
			"referenced_deployment_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The referenced deployment ID.",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceAliCloudFlinkDeploymentDraftCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	flinkService, err := NewFlinkService(client)
	if err != nil {
		return WrapError(err)
	}

	workspaceId := d.Get("workspace_id").(string)
	namespaceName := d.Get("namespace_name").(string)
	name := d.Get("name").(string)

	draft := &flinkAPI.DeploymentDraft{
		Workspace: workspaceId,
		Namespace: namespaceName,
		Name:      name,
	}

	if folderId, ok := d.GetOk("folder_id"); ok {
		draft.ParentId = folderId.(string)
	}

	if engineVersion, ok := d.GetOk("engine_version"); ok {
		draft.EngineVersion = engineVersion.(string)
	}

	if executionMode, ok := d.GetOk("execution_mode"); ok {
		draft.ExecutionMode = executionMode.(string)
	}

	if deploymentId, ok := d.GetOk("deployment_id"); ok {
		draft.ReferencedDeploymentId = deploymentId.(string)
	}

	// Handle artifact configuration
	if artifactUri, ok := d.GetOk("artifact_uri"); ok {
		// Simple artifact URI - create JAR artifact
		draft.Artifact = &aliyunFlinkAPI.Artifact{
			Kind: "JAR",
			JarArtifact: &aliyunFlinkAPI.JarArtifact{
				JarUri: artifactUri.(string),
			},
		}
	} else if artifactConfig, ok := d.GetOk("artifact"); ok {
		// Complex artifact configuration
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
			draft.Artifact = artifact
		}
	}

	// Handle Flink configuration
	if flinkConf, ok := d.GetOk("flink_conf"); ok {
		draft.FlinkConf = make(map[string]string)
		for k, v := range flinkConf.(map[string]interface{}) {
			draft.FlinkConf[k] = v.(string)
		}
	}

	// Handle environment variables (stored as LocalVariables)
	if envVars, ok := d.GetOk("environment_variables"); ok {
		envVarMap := envVars.(map[string]interface{})
		draft.LocalVariables = make([]*aliyunFlinkAPI.LocalVariable, 0, len(envVarMap))
		for k, v := range envVarMap {
			draft.LocalVariables = append(draft.LocalVariables, &aliyunFlinkAPI.LocalVariable{
				Name:  k,
				Value: v.(string),
			})
		}
	}

	// Handle labels/tags
	if tags, ok := d.GetOk("tags"); ok {
		draft.Labels = make(map[string]interface{})
		for k, v := range tags.(map[string]interface{}) {
			draft.Labels[k] = v.(string)
		}
	}

	newDraft, err := flinkService.CreateDeploymentDraft(workspaceId, namespaceName, draft)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_deployment_draft", "CreateDeploymentDraft", AlibabaCloudSdkGoERROR)
	}
	addDebugJson("DeploymentDraft", newDraft)

	d.SetId(fmt.Sprintf("%s:%s", namespaceName, newDraft.Id))
	d.Set("draft_id", newDraft.Id)

	// Wait for deployment draft creation to complete using StateRefreshFunc
	stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, flinkService.FlinkDeploymentDraftStateRefreshFunc(workspaceId, namespaceName, newDraft.Id, []string{}))
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// 最后调用Read同步状态
	return resourceAliCloudFlinkDeploymentDraftRead(d, meta)
}

func resourceAliCloudFlinkDeploymentDraftRead(d *schema.ResourceData, meta interface{}) error {
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
	draftID := parts[1]
	workspaceID := d.Get("workspace_id").(string)

	// Use FlinkService method instead of directly accessing VervericaClient
	deploymentDraft, err := flinkService.GetDeploymentDraft(workspaceID, namespace, draftID)
	if err != nil {
		if IsExpectedErrors(err, []string{"DraftNotFound"}) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	if deploymentDraft == nil {
		d.SetId("")
		return nil
	}

	// Set basic fields
	d.Set("namespace_name", deploymentDraft.Namespace)
	d.Set("workspace_id", deploymentDraft.Workspace)
	d.Set("name", deploymentDraft.Name)
	d.Set("draft_id", deploymentDraft.Id)
	d.Set("engine_version", deploymentDraft.EngineVersion)
	d.Set("execution_mode", deploymentDraft.ExecutionMode)

	// Set folder_id from ParentId
	if deploymentDraft.ParentId != "" {
		d.Set("folder_id", deploymentDraft.ParentId)
	}

	// Set deployment ID if present
	if deploymentDraft.ReferencedDeploymentId != "" {
		d.Set("deployment_id", deploymentDraft.ReferencedDeploymentId)
		d.Set("referenced_deployment_id", deploymentDraft.ReferencedDeploymentId)
	}

	// Set creator information
	d.Set("creator", deploymentDraft.Creator)
	d.Set("creator_name", deploymentDraft.CreatorName)
	d.Set("modifier", deploymentDraft.Modifier)
	d.Set("modifier_name", deploymentDraft.ModifierName)

	// Set timestamps
	if deploymentDraft.CreatedAt > 0 {
		d.Set("create_time", fmt.Sprintf("%d", deploymentDraft.CreatedAt))
	}
	if deploymentDraft.ModifiedAt > 0 {
		d.Set("update_time", fmt.Sprintf("%d", deploymentDraft.ModifiedAt))
	}

	// Set artifact configuration
	if deploymentDraft.Artifact != nil {
		// Handle simple artifact URI case
		if deploymentDraft.Artifact.JarArtifact != nil {
			d.Set("artifact_uri", deploymentDraft.Artifact.JarArtifact.JarUri)
		}

		// Handle complex artifact configuration
		artifactConfig := make([]map[string]interface{}, 0, 1)
		artifactMap := map[string]interface{}{
			"kind": deploymentDraft.Artifact.Kind,
		}

		switch deploymentDraft.Artifact.Kind {
		case "JAR":
			if deploymentDraft.Artifact.JarArtifact != nil {
				jarArtifact := make([]map[string]interface{}, 0, 1)
				jarMap := map[string]interface{}{
					"jar_uri": deploymentDraft.Artifact.JarArtifact.JarUri,
				}
				if deploymentDraft.Artifact.JarArtifact.EntryClass != "" {
					jarMap["entry_class"] = deploymentDraft.Artifact.JarArtifact.EntryClass
				}
				if deploymentDraft.Artifact.JarArtifact.MainArgs != "" {
					jarMap["main_args"] = deploymentDraft.Artifact.JarArtifact.MainArgs
				}
				if len(deploymentDraft.Artifact.JarArtifact.AdditionalDependencies) > 0 {
					jarMap["additional_dependencies"] = deploymentDraft.Artifact.JarArtifact.AdditionalDependencies
				}
				jarArtifact = append(jarArtifact, jarMap)
				artifactMap["jar_artifact"] = jarArtifact
			}
		case "PYTHON":
			if deploymentDraft.Artifact.PythonArtifact != nil {
				pythonArtifact := make([]map[string]interface{}, 0, 1)
				pythonMap := map[string]interface{}{
					"python_artifact_uri": deploymentDraft.Artifact.PythonArtifact.PythonArtifactUri,
				}
				if deploymentDraft.Artifact.PythonArtifact.EntryModule != "" {
					pythonMap["entry_module"] = deploymentDraft.Artifact.PythonArtifact.EntryModule
				}
				if deploymentDraft.Artifact.PythonArtifact.MainArgs != "" {
					pythonMap["main_args"] = deploymentDraft.Artifact.PythonArtifact.MainArgs
				}
				if len(deploymentDraft.Artifact.PythonArtifact.AdditionalDependencies) > 0 {
					pythonMap["additional_dependencies"] = deploymentDraft.Artifact.PythonArtifact.AdditionalDependencies
				}
				if len(deploymentDraft.Artifact.PythonArtifact.AdditionalPythonLibraries) > 0 {
					pythonMap["additional_python_libraries"] = deploymentDraft.Artifact.PythonArtifact.AdditionalPythonLibraries
				}
				if len(deploymentDraft.Artifact.PythonArtifact.AdditionalPythonArchives) > 0 {
					pythonMap["additional_python_archives"] = deploymentDraft.Artifact.PythonArtifact.AdditionalPythonArchives
				}
				pythonArtifact = append(pythonArtifact, pythonMap)
				artifactMap["python_artifact"] = pythonArtifact
			}
		case "SQLSCRIPT":
			if deploymentDraft.Artifact.SqlArtifact != nil {
				sqlArtifact := make([]map[string]interface{}, 0, 1)
				sqlMap := map[string]interface{}{
					"sql_script": deploymentDraft.Artifact.SqlArtifact.SqlScript,
				}
				if len(deploymentDraft.Artifact.SqlArtifact.AdditionalDependencies) > 0 {
					sqlMap["additional_dependencies"] = deploymentDraft.Artifact.SqlArtifact.AdditionalDependencies
				}
				sqlArtifact = append(sqlArtifact, sqlMap)
				artifactMap["sql_artifact"] = sqlArtifact
			}
		}

		artifactConfig = append(artifactConfig, artifactMap)
		d.Set("artifact", artifactConfig)
	}

	// Set Flink configuration
	if deploymentDraft.FlinkConf != nil && len(deploymentDraft.FlinkConf) > 0 {
		flinkConf := make(map[string]interface{})
		for k, v := range deploymentDraft.FlinkConf {
			flinkConf[k] = v
		}
		d.Set("flink_conf", flinkConf)
	}

	// Set environment variables (from LocalVariables)
	if deploymentDraft.LocalVariables != nil && len(deploymentDraft.LocalVariables) > 0 {
		envVars := make(map[string]interface{})
		for _, localVar := range deploymentDraft.LocalVariables {
			envVars[localVar.Name] = localVar.Value
		}
		d.Set("environment_variables", envVars)
	}

	// Set labels/tags
	if deploymentDraft.Labels != nil && len(deploymentDraft.Labels) > 0 {
		tags := make(map[string]interface{})
		for k, v := range deploymentDraft.Labels {
			tags[k] = v
		}
		d.Set("tags", tags)
	}

	return nil
}

func resourceAliCloudFlinkDeploymentDraftUpdate(d *schema.ResourceData, meta interface{}) error {
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
	draftID := parts[1]
	workspaceID := d.Get("workspace_id").(string)

	// Get the current deployment draft
	deploymentDraft, err := flinkService.GetDeploymentDraft(workspaceID, namespace, draftID)
	if err != nil {
		return WrapError(err)
	}

	update := false

	// Update basic fields
	if d.HasChange("name") {
		deploymentDraft.Name = d.Get("name").(string)
		update = true
	}

	if d.HasChange("engine_version") {
		deploymentDraft.EngineVersion = d.Get("engine_version").(string)
		update = true
	}

	if d.HasChange("execution_mode") {
		deploymentDraft.ExecutionMode = d.Get("execution_mode").(string)
		update = true
	}

	if d.HasChange("deployment_id") {
		if deploymentID, ok := d.GetOk("deployment_id"); ok {
			deploymentDraft.ReferencedDeploymentId = deploymentID.(string)
		} else {
			deploymentDraft.ReferencedDeploymentId = ""
		}
		update = true
	}

	// Update artifact configuration
	if d.HasChange("artifact_uri") || d.HasChange("artifact") ||
		d.HasChange("artifact.0.sql_artifact") || d.HasChange("artifact.0.sql_artifact.0.sql_script") ||
		d.HasChange("artifact.0.jar_artifact") || d.HasChange("artifact.0.jar_artifact.0.jar_uri") ||
		d.HasChange("artifact.0.python_artifact") || d.HasChange("artifact.0.python_artifact.0.python_artifact_uri") {

		if artifactUri, ok := d.GetOk("artifact_uri"); ok {
			// Simple artifact URI
			deploymentDraft.Artifact = &aliyunFlinkAPI.Artifact{
				Kind: "JAR",
				JarArtifact: &aliyunFlinkAPI.JarArtifact{
					JarUri: artifactUri.(string),
				},
			}
		} else if artifactConfig, ok := d.GetOk("artifact"); ok {
			// Complex artifact configuration
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
				deploymentDraft.Artifact = artifact
			}
		}
		update = true
	}

	// Update Flink configuration
	if d.HasChange("flink_conf") {
		if flinkConf, ok := d.GetOk("flink_conf"); ok {
			deploymentDraft.FlinkConf = make(map[string]string)
			for k, v := range flinkConf.(map[string]interface{}) {
				deploymentDraft.FlinkConf[k] = v.(string)
			}
		} else {
			deploymentDraft.FlinkConf = nil
		}
		update = true
	}

	// Update environment variables (stored as LocalVariables)
	if d.HasChange("environment_variables") {
		if envVars, ok := d.GetOk("environment_variables"); ok {
			envVarMap := envVars.(map[string]interface{})
			deploymentDraft.LocalVariables = make([]*aliyunFlinkAPI.LocalVariable, 0, len(envVarMap))
			for k, v := range envVarMap {
				deploymentDraft.LocalVariables = append(deploymentDraft.LocalVariables, &aliyunFlinkAPI.LocalVariable{
					Name:  k,
					Value: v.(string),
				})
			}
		} else {
			deploymentDraft.LocalVariables = nil
		}
		update = true
	}

	// Update labels/tags
	if d.HasChange("tags") {
		if tags, ok := d.GetOk("tags"); ok {
			deploymentDraft.Labels = make(map[string]interface{})
			for k, v := range tags.(map[string]interface{}) {
				deploymentDraft.Labels[k] = v.(string)
			}
		} else {
			deploymentDraft.Labels = nil
		}
		update = true
	}

	if update {
		err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err := flinkService.UpdateDeploymentDraft(workspaceID, namespace, deploymentDraft)
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
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateDeploymentDraft", AlibabaCloudSdkGoERROR)
		}

		// Wait for update to complete using StateRefreshFunc
		stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutUpdate), 5*time.Second, flinkService.FlinkDeploymentDraftStateRefreshFunc(workspaceID, namespace, draftID, []string{}))
		if _, err := stateConf.WaitForState(); err != nil {
			return WrapErrorf(err, IdMsg, d.Id())
		}
	}

	// 最后调用Read同步状态
	return resourceAliCloudFlinkDeploymentDraftRead(d, meta)
}

func resourceAliCloudFlinkDeploymentDraftDelete(d *schema.ResourceData, meta interface{}) error {
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
	draftID := parts[1]
	workspaceID := d.Get("workspace_id").(string)

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		// Use service method instead of directly accessing VervericaClient
		err := flinkService.DeleteDeploymentDraft(workspaceID, namespace, draftID)
		if err != nil {
			if IsExpectedErrors(err, []string{"DraftNotFound"}) {
				return nil
			}
			return resource.RetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteDeploymentDraft", AlibabaCloudSdkGoERROR)
	}

	return nil
}
