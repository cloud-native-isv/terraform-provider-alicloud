package alicloud

import (
	"fmt"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	aliyunFlinkAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/flink"
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
			"job_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the deployment draft.",
			},
			// Draft 特有参数 - 关联的部署ID
			"deployment_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Associated deployment ID for this draft.",
			},
			// 基本配置 - 与 deployment 对齐
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the deployment draft.",
			},
			"engine_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "vvr-8.0.5-flink-1.17",
				Description: "The Flink engine version.",
			},
			"execution_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "STREAMING",
				ValidateFunc: validation.StringInSlice([]string{"STREAMING", "BATCH"}, false),
				Description:  "The execution mode for the Flink job.",
			},
			// 部署目标配置 - 与 deployment 对齐
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
			// Artifact 配置 - 支持两种方式：简单的 URI 和复杂的结构
			"artifact_uri": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The URI of the deployment artifact (e.g., JAR file in OSS). Use this for simple JAR artifacts.",
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
			// 资源配置 - 保持旧的简单结构以及新的复杂结构
			"parallelism": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The parallelism of the job.",
			},
			"job_manager_resource_spec": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "JobManager resource specification (legacy, use streaming_resource_setting instead).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     1.0,
							Description: "CPU cores for JobManager.",
						},
						"memory": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "1g",
							Description: "Memory for JobManager (e.g., \"1g\").",
						},
					},
				},
			},
			"task_manager_resource_spec": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "TaskManager resource specification (legacy, use streaming_resource_setting instead).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:        schema.TypeFloat,
							Optional:    true,
							Default:     2.0,
							Description: "CPU cores for TaskManager.",
						},
						"memory": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "2g",
							Description: "Memory for TaskManager (e.g., \"2g\").",
						},
					},
				},
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
			// Flink 配置 - 统一命名
			"flink_conf": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Flink configuration properties.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			// 日志配置 - 与 deployment 对齐
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
			// 标签配置 - 与 deployment 对齐
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
			// 保持向后兼容的旧参数名 - 标记为 Deprecated
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Deprecated:  "Use job_name instead",
				Description: "Deprecated: Use job_name instead. The name of the deployment draft.",
			},
			"flink_configuration": {
				Type:        schema.TypeMap,
				Optional:    true,
				Deprecated:  "Use flink_conf instead",
				Description: "Deprecated: Use flink_conf instead. Map of Flink configuration properties.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			// Computed 字段
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
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the deployment draft.",
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
	artifactURI := d.Get("artifact_uri").(string)

	// Create deployment draft using cws-lib-go service
	draft := &aliyunFlinkAPI.DeploymentDraft{
		Workspace: workspaceId,
		Namespace: namespaceName,
		Name:      name,
	}

	// Set artifact
	draft.Artifact = &aliyunFlinkAPI.Artifact{
		Kind: "JAR",
		JarArtifact: &aliyunFlinkAPI.JarArtifact{
			JarUri: artifactURI,
		},
	}

	// Set deployment ID if provided
	if deploymentId, ok := d.GetOk("deployment_id"); ok {
		draft.ReferencedDeploymentId = deploymentId.(string)
	}

	// Handle Flink configuration
	if flinkConf, ok := d.GetOk("flink_configuration"); ok {
		draft.FlinkConf = make(map[string]string)
		for k, v := range flinkConf.(map[string]interface{}) {
			draft.FlinkConf[k] = v.(string)
		}
	}

	// Note: DeploymentDraft doesn't support resource specifications in the API
	// Resource specs are handled during deployment, not in draft creation

	// Create the deployment draft
	var response *aliyunFlinkAPI.DeploymentDraft
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		resp, err := flinkService.CreateDeploymentDraft(workspaceId, namespaceName, draft)
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
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_flink_deployment_draft", "CreateDeploymentDraft", AlibabaCloudSdkGoERROR)
	}

	if response == nil || response.DeploymentDraftId == "" {
		return WrapError(Error("Failed to get deployment draft ID from response"))
	}

	d.SetId(fmt.Sprintf("%s:%s", namespaceName, response.DeploymentDraftId))
	d.Set("draft_id", response.DeploymentDraftId)

	// Wait for deployment draft creation to complete using StateRefreshFunc
	stateConf := BuildStateConf([]string{}, []string{"Available"}, d.Timeout(schema.TimeoutCreate), 5*time.Second, flinkService.FlinkDeploymentDraftStateRefreshFunc(workspaceId, namespaceName, response.DeploymentDraftId, []string{}))
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

	// Update state with values from the response
	d.Set("namespace_name", deploymentDraft.Namespace)
	d.Set("workspace_id", deploymentDraft.Workspace)
	d.Set("name", deploymentDraft.Name)
	d.Set("draft_id", deploymentDraft.DeploymentDraftId)

	if deploymentDraft.ReferencedDeploymentId != "" {
		d.Set("deployment_id", deploymentDraft.ReferencedDeploymentId)
	}

	if deploymentDraft.CreatedAt > 0 {
		// Convert from int64 to string if needed
		createdAtStr := fmt.Sprintf("%d", deploymentDraft.CreatedAt)
		d.Set("create_time", createdAtStr)
	}

	if deploymentDraft.ModifiedAt > 0 {
		// Convert from int64 to string if needed
		modifiedAtStr := fmt.Sprintf("%d", deploymentDraft.ModifiedAt)
		d.Set("update_time", modifiedAtStr)
	}

	// Set artifact URI
	if deploymentDraft.Artifact != nil {
		if deploymentDraft.Artifact.JarArtifact != nil {
			d.Set("artifact_uri", deploymentDraft.Artifact.JarArtifact.JarUri)
		}
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

	// Use service methods instead of directly accessing Ververica client
	// First, get the current deployment draft
	deploymentDraft, err := flinkService.GetDeploymentDraft(workspaceID, namespace, draftID)
	if err != nil {
		return WrapError(err)
	}

	// Only update fields that have changed
	update := false

	if d.HasChange("name") {
		deploymentDraft.Name = d.Get("name").(string)
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

	if d.HasChange("artifact_uri") {
		if deploymentDraft.Artifact == nil {
			deploymentDraft.Artifact = &aliyunFlinkAPI.Artifact{
				Kind:        "JAR",
				JarArtifact: &aliyunFlinkAPI.JarArtifact{},
			}
		}

		if deploymentDraft.Artifact.JarArtifact == nil {
			deploymentDraft.Artifact.JarArtifact = &aliyunFlinkAPI.JarArtifact{}
		}

		deploymentDraft.Artifact.JarArtifact.JarUri = d.Get("artifact_uri").(string)
		update = true
	}

	// Handle Flink configuration if changed
	if d.HasChange("flink_configuration") {
		flinkConf := d.Get("flink_configuration").(map[string]interface{})
		deploymentDraft.FlinkConf = make(map[string]string)
		for k, v := range flinkConf {
			deploymentDraft.FlinkConf[k] = v.(string)
		}
		update = true
	}

	// Handle resource specifications if changed
	// Note: DeploymentDraft API doesn't support resource specifications
	// These are handled during deployment, not in draft
	if d.HasChange("job_manager_resource_spec") || d.HasChange("task_manager_resource_spec") {
		// Log a warning that resource specs are ignored for drafts
		// In practice, these would be applied when the draft is deployed
		update = true
	}

	if update {
		// Call the service method to update the deployment draft
		_, err = flinkService.UpdateDeploymentDraft(workspaceID, namespace, deploymentDraft)
		if err != nil {
			return WrapError(err)
		}

		// Wait a moment for the update to be processed
		time.Sleep(5 * time.Second)
	}

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
