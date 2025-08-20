package alicloud

import (
	"log"
	"regexp"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/alibabacloud-go/tea/tea"
)

func resourceAliCloudFCFunction() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudFCFunctionCreate,
		Read:   resourceAliCloudFCFunctionRead,
		Update: resourceAliCloudFCFunctionUpdate,
		Delete: resourceAliCloudFCFunctionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			// ========== Basic Function Configuration ==========
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^[0-9a-zA-Z_-]+$"), "The function name. Consists of uppercase and lowercase letters, digits (0 to 9), underscores (_), and dashes (-). It must begin with an English letter (a ~ z), (A ~ Z), or an underscore (_). Case sensitive. The length is 1~128 characters."),
				Description:  "The name of the FC function.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the FC function.",
			},
			"tags": tagsSchema(),

			// ========== Code Configuration ==========
			"code_config": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Code configuration for the FC function.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"oss_bucket_name": {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							Description: "OSS bucket name where the function code is stored.",
						},
						"oss_object_name": {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							Description: "OSS object name where the function code is stored.",
						},
						"zip_file": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "Base64 encoded ZIP file content of the function code.",
						},
						"checksum": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "Checksum of the function code.",
						},
					},
				},
			},

			// ========== Entrypoint Configuration ==========
			"entrypoint_config": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Entrypoint configuration for the FC function.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"handler": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The entry point of the function.",
						},
						"layers": {
							Type:        schema.TypeList,
							Optional:    true,
							Sensitive:   true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of layer ARNs to add to the function's execution environment.",
						},
					},
				},
			},

			// ========== Runtime Configuration ==========
			"runtime_config": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Runtime and resource configuration for the FC function.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"runtime": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"python3.10", "python3.9", "python3", "nodejs20", "nodejs18", "nodejs16", "nodejs14", "java11", "java8", "php7.2", "dotnetcore3.1", "go1", "custom.debian10", "custom", "custom-container"}, false),
							Description:  "The runtime of the function.",
						},
						"timeout": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 86400),
							Description:  "The timeout in seconds for the function execution.",
						},
						"initialization_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 600),
							Description:  "The initialization timeout in seconds.",
						},
						"environment_variables": {
							Type:        schema.TypeMap,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Environment variables for the function.",
						},
						"internet_access": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Whether the function can access the internet.",
						},
						// Resource limits
						"cpu": {
							Type:         schema.TypeFloat,
							Optional:     true,
							Default:      0.1,
							ValidateFunc: validation.FloatBetween(0.05, 16),
							Description:  "The CPU allocation for the function in vCPU.",
						},
						"memory_size": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      128,
							ValidateFunc: validation.IntBetween(64, 32768),
							Description:  "The memory size in MB for the function.",
						},
						"disk_size": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      512,
							ValidateFunc: validation.IntAtLeast(512),
							Description:  "The disk size in MB for the function.",
						},
						"instance_concurrency": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntBetween(1, 200),
							Description:  "The instance concurrency for the function.",
						},
					},
				},
			},

			// ========== Network Configuration ==========
			"network_config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Network configuration for the FC function.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vpc_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The VPC ID for the function.",
						},
						"vswitch_ids": {
							Type:        schema.TypeList,
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of VSwitch IDs for the function.",
						},
						"security_group_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The security group ID for the function.",
						},
						"dns_config": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Custom DNS configuration for the function.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"searches": {
										Type:        schema.TypeList,
										Optional:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Description: "DNS search domains.",
									},
									"dns_options": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "DNS options.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "DNS option name.",
												},
												"value": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "DNS option value.",
												},
											},
										},
									},
									"name_servers": {
										Type:        schema.TypeList,
										Optional:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Description: "List of DNS name servers.",
									},
								},
							},
						},
					},
				},
			},

			// ========== Storage Configuration ==========
			"storage_config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Storage configuration for the FC function.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"nas_config": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "NAS configuration for the function.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mount_points": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "List of NAS mount points.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enable_tls": {
													Type:        schema.TypeBool,
													Optional:    true,
													Default:     false,
													Description: "Whether to enable TLS for the mount.",
												},
												"server_addr": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "NAS server address.",
												},
												"mount_dir": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Mount directory path.",
												},
											},
										},
									},
									"user_id": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "User ID for NAS access.",
									},
									"group_id": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Group ID for NAS access.",
									},
								},
							},
						},
						"oss_mount_config": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "OSS mount configuration for the function.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mount_points": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "List of OSS mount points.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"read_only": {
													Type:        schema.TypeBool,
													Optional:    true,
													Default:     false,
													Description: "Whether the mount is read-only.",
												},
												"bucket_name": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "OSS bucket name.",
												},
												"endpoint": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "OSS endpoint.",
												},
												"bucket_path": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "OSS bucket path.",
												},
												"mount_dir": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Mount directory path.",
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

			// ========== Log Configuration ==========
			"log_config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Logging configuration for the FC function.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"project": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "SLS project name for function logs.",
						},
						"logstore": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "SLS logstore name for function logs.",
						},
						"enable_instance_metrics": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Whether to enable instance metrics.",
						},
						"enable_request_metrics": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Whether to enable request metrics.",
						},
						"log_begin_rule": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"None", "DefaultRegex"}, false),
							Description:  "Log begin rule.",
						},
					},
				},
			},

			// ========== Lifecycle Configuration ==========
			"lifecycle_config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Lifecycle configuration for the FC function.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pre_stop": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Pre-stop hook configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"timeout": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 900),
										Description:  "Timeout in seconds for the pre-stop hook.",
									},
									"handler": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Handler for the pre-stop hook.",
									},
								},
							},
						},
						"initializer": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Initializer configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"timeout": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 600),
										Description:  "Timeout in seconds for the initializer.",
									},
									"handler": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Handler for the initializer.",
									},
								},
							},
						},
					},
				},
			},

			// ========== Container Configuration ==========
			"container_config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Container configuration for custom container runtime.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Container image URI.",
						},
						"resolved_image_uri": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Resolved container image URI.",
						},
						"entrypoint": {
							Type:        schema.TypeList,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Container entrypoint command.",
						},
						"command": {
							Type:        schema.TypeList,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Container command arguments.",
						},
						"args": {
							Type:        schema.TypeList,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Container runtime arguments.",
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 65535),
							Description:  "Container listening port.",
						},
						"health_check_config": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Container health check configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"initial_delay_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(0, 120),
										Description:  "Initial delay in seconds before starting health checks.",
									},
									"timeout_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 60),
										Description:  "Timeout in seconds for each health check.",
									},
									"http_get_url": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "HTTP URL for health check.",
									},
									"period_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 120),
										Description:  "Period in seconds between health checks.",
									},
									"failure_threshold": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 120),
										Description:  "Number of consecutive failures to mark as unhealthy.",
									},
									"success_threshold": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 120),
										Description:  "Number of consecutive successes to mark as healthy.",
									},
								},
							},
						},
						"acr_instance_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Deprecated:  "Field 'acr_instance_id' has been deprecated from provider version 1.228.0. ACR Enterprise version Image Repository ID, which must be entered when using ACR Enterprise version image. (Obsolete)",
							Description: "ACR instance ID for container registry authentication.",
						},
						"acceleration_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Deprecated:   "Field 'acceleration_type' has been deprecated from provider version 1.228.0. Whether to enable Image acceleration. Default: The Default value, indicating that image acceleration is enabled. None: indicates that image acceleration is disabled. (Obsolete)",
							ValidateFunc: validation.StringInSlice([]string{"Default", "None"}, false),
							Description:  "Container image acceleration type.",
						},
						"acceleration_info": {
							Type:        schema.TypeList,
							Computed:    true,
							Deprecated:  "Field 'acceleration_info' has been deprecated from provider version 1.228.0. Image Acceleration Information (Obsolete)",
							MaxItems:    1,
							Description: "Container image acceleration information.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"status": {
										Type:        schema.TypeString,
										Computed:    true,
										Deprecated:  "Field 'status' has been deprecated from provider version 1.228.0. Image Acceleration Status (Deprecated)",
										Description: "Image acceleration status.",
									},
								},
							},
						},
					},
				},
			},

			// ========== GPU Configuration ==========
			"gpu_config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "GPU configuration for the FC function.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gpu_memory_size": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "GPU memory size in MB.",
						},
						"gpu_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "fc.gpu.tesla.1",
							ValidateFunc: validation.StringInSlice([]string{"fc.gpu.tesla.1", "fc.gpu.ampere.1", "fc.gpu.ada.1", "g1"}, false),
							Description:  "GPU device type.",
						},
					},
				},
			},

			// ========== RAM Configuration ==========
			"ram_config": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "RAM role configuration for the FC function.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_arn": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "RAM role ARN for the function execution.",
						},
					},
				},
			},

			// ========== Computed and Read-Only Fields ==========
			"code_size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The size of the function code in bytes.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time when the function was created.",
			},
			"function_arn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ARN of the function.",
			},
			"function_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the function.",
			},
			"last_modified_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time when the function was last modified.",
			},
			"last_update_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the last update.",
			},
			"last_update_status_reason": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The reason for the last update status.",
			},
			"last_update_status_reason_code": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The reason code for the last update status.",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current state of the function.",
			},
			"state_reason": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The reason for the current state.",
			},
			"state_reason_code": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The reason code for the current state.",
			},
			"tracing_config": {
				Type:        schema.TypeList,
				Computed:    true,
				MaxItems:    1,
				Description: "Tracing configuration for the function.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tracing type.",
						},
						"params": {
							Type:        schema.TypeMap,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Tracing parameters.",
						},
					},
				},
			},
		},
	}
}

func resourceAliCloudFCFunctionCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	// Build function name
	var functionName string
	if v, ok := d.GetOk("name"); ok {
		functionName = v.(string)
	} else {
		functionName = resource.PrefixedUniqueId("tf-function-")
	}

	// Build create request from schema
	request := service.BuildCreateFunctionInputFromSchema(d)
	request.FunctionName = tea.String(functionName)

	// Use retry for creation
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := service.CreateFCFunction(request)
		if err != nil {
			if IsExpectedErrors(err, []string{"ServiceUnavailable", "ThrottlingException"}) {
				return resource.RetryableError(err)
			}
			if IsExpectedErrors(err, []string{"FunctionAlreadyExists"}) {
				return resource.NonRetryableError(GetNotFoundErrorFromString(GetNotFoundMessage("FC Function", functionName)))
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_fc_function", "CreateFunction", AlibabaCloudSdkGoERROR)
	}

	// Set resource ID
	d.SetId(functionName)

	// Wait for function to be ready
	err = service.WaitForFCFunctionCreating(functionName, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// Read the function state
	return resourceAliCloudFCFunctionRead(d, meta)
}

func resourceAliCloudFCFunctionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName := d.Id()
	function, err := service.DescribeFCFunction(functionName)
	if err != nil {
		if !d.IsNewResource() && IsNotFoundError(err) {
			log.Printf("[DEBUG] Resource alicloud_fc_function DescribeFunction Failed!!! %s", err)
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// Set basic fields
	d.Set("name", function.FunctionName)
	d.Set("description", function.Description)
	d.Set("function_arn", function.FunctionArn)
	d.Set("function_id", function.FunctionId)
	d.Set("create_time", function.CreatedTime)
	d.Set("last_modified_time", function.LastModifiedTime)
	d.Set("state", function.State)
	d.Set("state_reason", function.StateReason)
	d.Set("state_reason_code", function.StateReasonCode)
	d.Set("last_update_status", function.LastUpdateStatus)
	d.Set("last_update_status_reason", function.LastUpdateStatusReason)
	d.Set("last_update_status_reason_code", function.LastUpdateStatusReasonCode)
	d.Set("code_size", function.CodeSize)

	// Set configuration blocks using the service helper method
	return service.SetSchemaFromFunction(d, function)
}

func resourceAliCloudFCFunctionUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName := d.Id()

	// Build update request from schema changes
	request := service.BuildUpdateFunctionInputFromSchema(d)

	// Use retry for update
	err = resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err := service.UpdateFCFunction(functionName, request)
		if err != nil {
			if IsExpectedErrors(err, []string{"ServiceUnavailable", "ThrottlingException"}) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "UpdateFunction", AlibabaCloudSdkGoERROR)
	}

	// Wait for function to be updated
	err = service.WaitForFCFunctionUpdating(functionName, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	// Read the updated function state
	return resourceAliCloudFCFunctionRead(d, meta)
}

func resourceAliCloudFCFunctionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	service, err := NewFCService(client)
	if err != nil {
		return WrapError(err)
	}

	functionName := d.Id()

	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		err := service.DeleteFCFunction(functionName)
		if err != nil {
			if IsExpectedErrors(err, []string{"ServiceUnavailable", "ThrottlingException"}) {
				return resource.RetryableError(err)
			}
			if IsExpectedErrors(err, []string{"ResourceNotFound", "FunctionNotFound"}) {
				return nil
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteFunction", AlibabaCloudSdkGoERROR)
	}

	// Wait for function to be deleted
	err = service.WaitForFCFunctionDeleting(functionName, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return WrapErrorf(err, IdMsg, d.Id())
	}

	return nil
}
