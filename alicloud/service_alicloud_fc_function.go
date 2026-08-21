package alicloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	fc "github.com/alibabacloud-go/fc-20230330/v4/client"
	aliyunFCAPI "github.com/cloud-native-tools/cws-lib-go/lib/cloud/aliyun/api/fc/v3"
)

// Function methods for FCService

// EncodeFunctionId encodes function name into an ID string
func EncodeFunctionId(functionName string) string {
	return functionName
}

// DecodeFunctionId decodes function ID string to function name
func DecodeFunctionId(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("invalid function ID format, cannot be empty")
	}
	return id, nil
}

// DescribeFCFunction retrieves function information by name
func (s *FCService) DescribeFCFunction(functionName string) (*aliyunFCAPI.Function, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	return s.GetAPI().GetFunction(functionName, nil)
}

// ListFCFunctions lists all functions with optional filters
func (s *FCService) ListFCFunctions(prefix *string, limit *int32, nextToken *string) ([]*aliyunFCAPI.Function, error) {
	request := &fc.ListFunctionsRequest{
		Prefix:    prefix,
		Limit:     limit,
		NextToken: nextToken,
	}
	return s.GetAPI().ListFunctions(request)
}

// CreateFCFunction creates a new FC function
func (s *FCService) CreateFCFunction(function *aliyunFCAPI.Function) (*aliyunFCAPI.Function, error) {
	if function == nil {
		return nil, fmt.Errorf("function cannot be nil")
	}
	return s.GetAPI().CreateFunction(function)
}

// UpdateFCFunction updates an existing FC function
func (s *FCService) UpdateFCFunction(functionName string, function *aliyunFCAPI.Function) (*aliyunFCAPI.Function, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	if function == nil {
		return nil, fmt.Errorf("function cannot be nil")
	}
	return s.GetAPI().UpdateFunction(functionName, function)
}

// DeleteFCFunction deletes an FC function
func (s *FCService) DeleteFCFunction(functionName string) error {
	if functionName == "" {
		return fmt.Errorf("function name cannot be empty")
	}
	return s.GetAPI().DeleteFunction(functionName)
}

// BuildCreateFunctionInputFromSchema builds Function from Terraform schema data.
// The resource schema follows the FC 3.0 layout and keeps the configuration in
// nested blocks (code_config / entrypoint_config / runtime_config / ...), so
// every field must be read from its block path instead of a top-level key.
func (s *FCService) BuildCreateFunctionInputFromSchema(d *schema.ResourceData) *aliyunFCAPI.Function {
	function := &aliyunFCAPI.Function{}

	if v, ok := d.GetOk("name"); ok {
		function.FunctionName = tea.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		function.Description = tea.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		function.Tags = buildFCTagsFromMap(v.(map[string]interface{}))
	}

	// ========== code_config (Required) ==========
	if v, ok := d.GetOk("code_config"); ok {
		function.Code = expandFCCodeConfig(v.([]interface{}))
	}

	// ========== entrypoint_config (Required) ==========
	if v, ok := d.GetOk("entrypoint_config"); ok {
		blocks := v.([]interface{})
		if len(blocks) > 0 && blocks[0] != nil {
			cfg := blocks[0].(map[string]interface{})
			if handler, ok := cfg["handler"].(string); ok && handler != "" {
				function.Handler = tea.String(handler)
			}
			if layers, ok := cfg["layers"].([]interface{}); ok && len(layers) > 0 {
				for _, layer := range layers {
					if arn, ok := layer.(string); ok && arn != "" {
						function.Layers = append(function.Layers, tea.String(arn))
					}
				}
			}
		}
	}

	// ========== runtime_config (Required) ==========
	if v, ok := d.GetOk("runtime_config"); ok {
		blocks := v.([]interface{})
		if len(blocks) > 0 && blocks[0] != nil {
			cfg := blocks[0].(map[string]interface{})
			if runtime, ok := cfg["runtime"].(string); ok && runtime != "" {
				function.Runtime = tea.String(runtime)
			}
			if timeout, ok := cfg["timeout"].(int); ok && timeout > 0 {
				function.Timeout = tea.Int32(int32(timeout))
			}
			if envVars, ok := cfg["environment_variables"].(map[string]interface{}); ok && len(envVars) > 0 {
				function.Environment = make(map[string]*string, len(envVars))
				for k, val := range envVars {
					if strVal, ok := val.(string); ok {
						function.Environment[k] = tea.String(strVal)
					}
				}
			}
			if cpu, ok := cfg["cpu"].(float64); ok && cpu > 0 {
				function.Cpu = tea.Float32(float32(cpu))
			}
			if memorySize, ok := cfg["memory_size"].(int); ok && memorySize > 0 {
				function.MemorySize = tea.Int32(int32(memorySize))
			}
			if diskSize, ok := cfg["disk_size"].(int); ok && diskSize > 0 {
				function.DiskSize = tea.Int32(int32(diskSize))
			}
			if concurrency, ok := cfg["instance_concurrency"].(int); ok && concurrency > 0 {
				function.InstanceConcurrency = tea.Int32(int32(concurrency))
			}
		}
	}
	// internet_access is Optional+Computed without a default: only send it when
	// explicitly configured so the FC-side default stays untouched otherwise.
	if v, ok := d.GetOkExists("runtime_config.0.internet_access"); ok {
		function.InternetAccess = tea.Bool(v.(bool))
	}

	// ========== lifecycle hooks ==========
	// lifecycle_config carries explicit pre_stop/initializer hooks;
	// runtime_config.initialization_timeout maps onto the initializer hook
	// timeout (FC 3.0 keeps the timeout; the initializer handler is deprecated).
	lifecycle := &aliyunFCAPI.InstanceLifecycleConfig{}
	hasLifecycle := false
	if v, ok := d.GetOk("lifecycle_config"); ok {
		blocks := v.([]interface{})
		if len(blocks) > 0 && blocks[0] != nil {
			cfg := blocks[0].(map[string]interface{})
			if hooks, ok := cfg["pre_stop"].([]interface{}); ok && len(hooks) > 0 && hooks[0] != nil {
				lifecycle.PreStop = expandFCLifecycleHook(hooks[0].(map[string]interface{}))
				hasLifecycle = true
			}
			if hooks, ok := cfg["initializer"].([]interface{}); ok && len(hooks) > 0 && hooks[0] != nil {
				lifecycle.Initializer = expandFCLifecycleHook(hooks[0].(map[string]interface{}))
				hasLifecycle = true
			}
		}
	}
	if v, ok := d.GetOk("runtime_config.0.initialization_timeout"); ok {
		if initTimeout, ok := v.(int); ok && initTimeout > 0 {
			if lifecycle.Initializer == nil {
				lifecycle.Initializer = &aliyunFCAPI.InstanceLifecycleHook{}
			}
			lifecycle.Initializer.Timeout = tea.Int32(int32(initTimeout))
			hasLifecycle = true
		}
	}
	if hasLifecycle {
		function.InstanceLifecycleConfig = lifecycle
	}

	// ========== network_config (Optional) ==========
	if v, ok := d.GetOk("network_config"); ok {
		blocks := v.([]interface{})
		if len(blocks) > 0 && blocks[0] != nil {
			cfg := blocks[0].(map[string]interface{})
			vpcConfig := &aliyunFCAPI.VPCConfig{}
			hasVpc := false
			if vpcId, ok := cfg["vpc_id"].(string); ok && vpcId != "" {
				vpcConfig.VpcId = tea.String(vpcId)
				hasVpc = true
			}
			if sgId, ok := cfg["security_group_id"].(string); ok && sgId != "" {
				vpcConfig.SecurityGroupId = tea.String(sgId)
				hasVpc = true
			}
			if vsws, ok := cfg["vswitch_ids"].([]interface{}); ok {
				for _, vsw := range vsws {
					if id, ok := vsw.(string); ok && id != "" {
						vpcConfig.VSwitchIds = append(vpcConfig.VSwitchIds, tea.String(id))
					}
				}
				if len(vpcConfig.VSwitchIds) > 0 {
					hasVpc = true
				}
			}
			if hasVpc {
				function.VpcConfig = vpcConfig
			}
			if dnsConfigs, ok := cfg["dns_config"].([]interface{}); ok && len(dnsConfigs) > 0 && dnsConfigs[0] != nil {
				function.CustomDNS = expandFCCustomDNS(dnsConfigs[0].(map[string]interface{}))
			}
		}
	}

	// ========== storage_config (Optional) ==========
	if v, ok := d.GetOk("storage_config"); ok {
		blocks := v.([]interface{})
		if len(blocks) > 0 && blocks[0] != nil {
			cfg := blocks[0].(map[string]interface{})
			if nasConfigs, ok := cfg["nas_config"].([]interface{}); ok && len(nasConfigs) > 0 && nasConfigs[0] != nil {
				function.NasConfig = expandFCNASConfig(nasConfigs[0].(map[string]interface{}))
			}
			if ossMounts, ok := cfg["oss_mount_config"].([]interface{}); ok && len(ossMounts) > 0 && ossMounts[0] != nil {
				function.OssMountConfig = expandFCOSSMountConfig(ossMounts[0].(map[string]interface{}))
			}
		}
	}

	// ========== log_config (Optional) ==========
	if v, ok := d.GetOk("log_config"); ok {
		blocks := v.([]interface{})
		if len(blocks) > 0 && blocks[0] != nil {
			cfg := blocks[0].(map[string]interface{})
			logConfig := &aliyunFCAPI.LogConfig{}
			if project, ok := cfg["project"].(string); ok && project != "" {
				logConfig.Project = tea.String(project)
			}
			if logstore, ok := cfg["logstore"].(string); ok && logstore != "" {
				logConfig.Logstore = tea.String(logstore)
			}
			if v, ok := cfg["enable_instance_metrics"].(bool); ok {
				logConfig.EnableInstanceMetrics = tea.Bool(v)
			}
			if v, ok := cfg["enable_request_metrics"].(bool); ok {
				logConfig.EnableRequestMetrics = tea.Bool(v)
			}
			if rule, ok := cfg["log_begin_rule"].(string); ok && rule != "" {
				logConfig.LogBeginRule = tea.String(rule)
			}
			function.LogConfig = logConfig
		}
	}

	// ========== container_config (Optional, custom-container runtime) ==========
	if v, ok := d.GetOk("container_config"); ok {
		blocks := v.([]interface{})
		if len(blocks) > 0 && blocks[0] != nil {
			function.CustomContainerConfig = expandFCCustomContainerConfig(blocks[0].(map[string]interface{}))
		}
	}

	// ========== gpu_config (Optional) ==========
	if v, ok := d.GetOk("gpu_config"); ok {
		blocks := v.([]interface{})
		if len(blocks) > 0 && blocks[0] != nil {
			cfg := blocks[0].(map[string]interface{})
			gpuConfig := &aliyunFCAPI.GPUConfig{}
			if memory, ok := cfg["gpu_memory_size"].(int); ok && memory > 0 {
				gpuConfig.GpuMemorySize = tea.Int32(int32(memory))
			}
			if gpuType, ok := cfg["gpu_type"].(string); ok && gpuType != "" {
				gpuConfig.GpuType = tea.String(gpuType)
			}
			function.GpuConfig = gpuConfig
		}
	}

	// ========== ram_config (Optional) ==========
	if v, ok := d.GetOk("ram_config.0.role_arn"); ok {
		function.Role = tea.String(v.(string))
	}

	return function
}

// expandFCCodeConfig turns the code_config block into the write-side code
// payload (OSS location / base64 zip / checksum).
func expandFCCodeConfig(blocks []interface{}) *aliyunFCAPI.InputCodeLocation {
	if len(blocks) == 0 || blocks[0] == nil {
		return nil
	}
	cfg := blocks[0].(map[string]interface{})
	code := &aliyunFCAPI.InputCodeLocation{}
	empty := true
	if v, ok := cfg["oss_bucket_name"].(string); ok && v != "" {
		code.OssBucketName = tea.String(v)
		empty = false
	}
	if v, ok := cfg["oss_object_name"].(string); ok && v != "" {
		code.OssObjectName = tea.String(v)
		empty = false
	}
	if v, ok := cfg["zip_file"].(string); ok && v != "" {
		code.ZipFile = tea.String(v)
		empty = false
	}
	if v, ok := cfg["checksum"].(string); ok && v != "" {
		code.Checksum = tea.String(v)
		empty = false
	}
	if empty {
		return nil
	}
	return code
}

// expandFCLifecycleHook turns a pre_stop/initializer block into a lifecycle hook.
func expandFCLifecycleHook(cfg map[string]interface{}) *aliyunFCAPI.InstanceLifecycleHook {
	hook := &aliyunFCAPI.InstanceLifecycleHook{}
	if v, ok := cfg["handler"].(string); ok && v != "" {
		hook.Handler = tea.String(v)
	}
	if v, ok := cfg["timeout"].(int); ok && v > 0 {
		hook.Timeout = tea.Int32(int32(v))
	}
	return hook
}

// expandFCCustomDNS turns the dns_config block into the CustomDNS payload.
func expandFCCustomDNS(cfg map[string]interface{}) *aliyunFCAPI.CustomDNS {
	dns := &aliyunFCAPI.CustomDNS{}
	empty := true
	if values, ok := cfg["searches"].([]interface{}); ok {
		for _, item := range values {
			if v, ok := item.(string); ok && v != "" {
				dns.Searches = append(dns.Searches, tea.String(v))
			}
		}
	}
	if values, ok := cfg["name_servers"].([]interface{}); ok {
		for _, item := range values {
			if v, ok := item.(string); ok && v != "" {
				dns.NameServers = append(dns.NameServers, tea.String(v))
			}
		}
	}
	if options, ok := cfg["dns_options"].([]interface{}); ok {
		for _, item := range options {
			option, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			dnsOption := &aliyunFCAPI.DNSOption{}
			if v, ok := option["name"].(string); ok && v != "" {
				dnsOption.Name = tea.String(v)
			}
			if v, ok := option["value"].(string); ok && v != "" {
				dnsOption.Value = tea.String(v)
			}
			if dnsOption.Name != nil || dnsOption.Value != nil {
				dns.DnsOptions = append(dns.DnsOptions, dnsOption)
			}
		}
	}
	empty = len(dns.Searches) == 0 && len(dns.NameServers) == 0 && len(dns.DnsOptions) == 0
	if empty {
		return nil
	}
	return dns
}

// expandFCNASConfig turns the nas_config block into the NAS payload.
func expandFCNASConfig(cfg map[string]interface{}) *aliyunFCAPI.NASConfig {
	nasConfig := &aliyunFCAPI.NASConfig{}
	if v, ok := cfg["user_id"].(int); ok {
		nasConfig.UserId = tea.Int32(int32(v))
	}
	if v, ok := cfg["group_id"].(int); ok {
		nasConfig.GroupId = tea.Int32(int32(v))
	}
	if points, ok := cfg["mount_points"].([]interface{}); ok {
		for _, item := range points {
			point, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			mountPoint := &aliyunFCAPI.NASMountPoint{}
			if v, ok := point["server_addr"].(string); ok && v != "" {
				mountPoint.ServerAddr = tea.String(v)
			}
			if v, ok := point["mount_dir"].(string); ok && v != "" {
				mountPoint.MountDir = tea.String(v)
			}
			if v, ok := point["enable_tls"].(bool); ok {
				mountPoint.EnableTLS = tea.Bool(v)
			}
			nasConfig.MountPoints = append(nasConfig.MountPoints, mountPoint)
		}
	}
	return nasConfig
}

// expandFCOSSMountConfig turns the oss_mount_config block into the OSS mount payload.
func expandFCOSSMountConfig(cfg map[string]interface{}) *aliyunFCAPI.OSSMountConfig {
	ossMountConfig := &aliyunFCAPI.OSSMountConfig{}
	if points, ok := cfg["mount_points"].([]interface{}); ok {
		for _, item := range points {
			point, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			mountPoint := &aliyunFCAPI.OSSMountPoint{}
			if v, ok := point["bucket_name"].(string); ok && v != "" {
				mountPoint.BucketName = tea.String(v)
			}
			if v, ok := point["bucket_path"].(string); ok && v != "" {
				mountPoint.BucketPath = tea.String(v)
			}
			if v, ok := point["endpoint"].(string); ok && v != "" {
				mountPoint.Endpoint = tea.String(v)
			}
			if v, ok := point["mount_dir"].(string); ok && v != "" {
				mountPoint.MountDir = tea.String(v)
			}
			if v, ok := point["read_only"].(bool); ok {
				mountPoint.ReadOnly = tea.Bool(v)
			}
			ossMountConfig.MountPoints = append(ossMountConfig.MountPoints, mountPoint)
		}
	}
	return ossMountConfig
}

// expandFCCustomContainerConfig turns the container_config block into the
// custom container payload. The schema keeps a deprecated "args" field which
// has no FC 3.0 API counterpart, so it is intentionally not mapped.
func expandFCCustomContainerConfig(cfg map[string]interface{}) *aliyunFCAPI.CustomContainerConfig {
	containerConfig := &aliyunFCAPI.CustomContainerConfig{}
	if v, ok := cfg["image"].(string); ok && v != "" {
		containerConfig.Image = tea.String(v)
	}
	if values, ok := cfg["entrypoint"].([]interface{}); ok {
		for _, item := range values {
			if v, ok := item.(string); ok && v != "" {
				containerConfig.Entrypoint = append(containerConfig.Entrypoint, tea.String(v))
			}
		}
	}
	if values, ok := cfg["command"].([]interface{}); ok {
		for _, item := range values {
			if v, ok := item.(string); ok && v != "" {
				containerConfig.Command = append(containerConfig.Command, tea.String(v))
			}
		}
	}
	if v, ok := cfg["port"].(int); ok && v > 0 {
		containerConfig.Port = tea.Int32(int32(v))
	}
	if v, ok := cfg["acr_instance_id"].(string); ok && v != "" {
		containerConfig.AcrInstanceId = tea.String(v)
	}
	if v, ok := cfg["acceleration_type"].(string); ok && v != "" {
		containerConfig.AccelerationType = tea.String(v)
	}
	if healthChecks, ok := cfg["health_check_config"].([]interface{}); ok && len(healthChecks) > 0 && healthChecks[0] != nil {
		healthCfg := healthChecks[0].(map[string]interface{})
		healthCheck := &aliyunFCAPI.CustomHealthCheckConfig{}
		if v, ok := healthCfg["http_get_url"].(string); ok && v != "" {
			healthCheck.HttpGetUrl = tea.String(v)
		}
		if v, ok := healthCfg["initial_delay_seconds"].(int); ok && v > 0 {
			healthCheck.InitialDelaySeconds = tea.Int32(int32(v))
		}
		if v, ok := healthCfg["period_seconds"].(int); ok && v > 0 {
			healthCheck.PeriodSeconds = tea.Int32(int32(v))
		}
		if v, ok := healthCfg["timeout_seconds"].(int); ok && v > 0 {
			healthCheck.TimeoutSeconds = tea.Int32(int32(v))
		}
		if v, ok := healthCfg["failure_threshold"].(int); ok && v > 0 {
			healthCheck.FailureThreshold = tea.Int32(int32(v))
		}
		if v, ok := healthCfg["success_threshold"].(int); ok && v > 0 {
			healthCheck.SuccessThreshold = tea.Int32(int32(v))
		}
		containerConfig.HealthCheckConfig = healthCheck
	}
	return containerConfig
}

// buildFCTagsFromMap turns the top-level tags map into wrapper tags.
func buildFCTagsFromMap(tagsMap map[string]interface{}) []*aliyunFCAPI.Tag {
	if len(tagsMap) == 0 {
		return nil
	}
	tags := make([]*aliyunFCAPI.Tag, 0, len(tagsMap))
	for k, v := range tagsMap {
		tag := &aliyunFCAPI.Tag{Key: tea.String(k)}
		if strVal, ok := v.(string); ok {
			tag.Value = tea.String(strVal)
		}
		tags = append(tags, tag)
	}
	return tags
}

// BuildUpdateFunctionInputFromSchema builds Function for update from Terraform
// schema data. It mirrors the create builder but only emits blocks that
// actually changed; nil fields are left untouched by the FC v3 API. Tags are
// intentionally skipped: UpdateFunctionInput has no Tags field (tag changes
// require TagResources/UntagResources).
func (s *FCService) BuildUpdateFunctionInputFromSchema(d *schema.ResourceData) *aliyunFCAPI.Function {
	function := &aliyunFCAPI.Function{}

	if d.HasChange("description") {
		if v, ok := d.GetOk("description"); ok {
			function.Description = tea.String(v.(string))
		} else {
			function.Description = tea.String("")
		}
	}

	// ========== code_config ==========
	if d.HasChange("code_config") {
		if v, ok := d.GetOk("code_config"); ok {
			function.Code = expandFCCodeConfig(v.([]interface{}))
		}
	}

	// ========== entrypoint_config ==========
	if d.HasChange("entrypoint_config") {
		if v, ok := d.GetOk("entrypoint_config"); ok {
			blocks := v.([]interface{})
			if len(blocks) > 0 && blocks[0] != nil {
				cfg := blocks[0].(map[string]interface{})
				if handler, ok := cfg["handler"].(string); ok && handler != "" {
					function.Handler = tea.String(handler)
				}
				if layers, ok := cfg["layers"].([]interface{}); ok {
					function.Layers = make([]*string, 0, len(layers))
					for _, layer := range layers {
						if arn, ok := layer.(string); ok && arn != "" {
							function.Layers = append(function.Layers, tea.String(arn))
						}
					}
				}
			}
		}
	}

	// ========== runtime_config ==========
	if d.HasChange("runtime_config") {
		if v, ok := d.GetOk("runtime_config"); ok {
			blocks := v.([]interface{})
			if len(blocks) > 0 && blocks[0] != nil {
				cfg := blocks[0].(map[string]interface{})
				if runtime, ok := cfg["runtime"].(string); ok && runtime != "" {
					function.Runtime = tea.String(runtime)
				}
				if timeout, ok := cfg["timeout"].(int); ok && timeout > 0 {
					function.Timeout = tea.Int32(int32(timeout))
				}
				if envVars, ok := cfg["environment_variables"].(map[string]interface{}); ok {
					function.Environment = make(map[string]*string, len(envVars))
					for k, val := range envVars {
						if strVal, ok := val.(string); ok {
							function.Environment[k] = tea.String(strVal)
						}
					}
				}
				if cpu, ok := cfg["cpu"].(float64); ok && cpu > 0 {
					function.Cpu = tea.Float32(float32(cpu))
				}
				if memorySize, ok := cfg["memory_size"].(int); ok && memorySize > 0 {
					function.MemorySize = tea.Int32(int32(memorySize))
				}
				if diskSize, ok := cfg["disk_size"].(int); ok && diskSize > 0 {
					function.DiskSize = tea.Int32(int32(diskSize))
				}
				if concurrency, ok := cfg["instance_concurrency"].(int); ok && concurrency > 0 {
					function.InstanceConcurrency = tea.Int32(int32(concurrency))
				}
			}
		}
		if v, ok := d.GetOkExists("runtime_config.0.internet_access"); ok {
			function.InternetAccess = tea.Bool(v.(bool))
		}
	}

	// ========== lifecycle hooks ==========
	if d.HasChange("lifecycle_config") || d.HasChange("runtime_config") {
		lifecycle := &aliyunFCAPI.InstanceLifecycleConfig{}
		hasLifecycle := false
		if v, ok := d.GetOk("lifecycle_config"); ok {
			blocks := v.([]interface{})
			if len(blocks) > 0 && blocks[0] != nil {
				cfg := blocks[0].(map[string]interface{})
				if hooks, ok := cfg["pre_stop"].([]interface{}); ok && len(hooks) > 0 && hooks[0] != nil {
					lifecycle.PreStop = expandFCLifecycleHook(hooks[0].(map[string]interface{}))
					hasLifecycle = true
				}
				if hooks, ok := cfg["initializer"].([]interface{}); ok && len(hooks) > 0 && hooks[0] != nil {
					lifecycle.Initializer = expandFCLifecycleHook(hooks[0].(map[string]interface{}))
					hasLifecycle = true
				}
			}
		}
		if v, ok := d.GetOk("runtime_config.0.initialization_timeout"); ok {
			if initTimeout, ok := v.(int); ok && initTimeout > 0 {
				if lifecycle.Initializer == nil {
					lifecycle.Initializer = &aliyunFCAPI.InstanceLifecycleHook{}
				}
				lifecycle.Initializer.Timeout = tea.Int32(int32(initTimeout))
				hasLifecycle = true
			}
		}
		if hasLifecycle && (d.HasChange("lifecycle_config") || d.HasChange("runtime_config")) {
			function.InstanceLifecycleConfig = lifecycle
		}
	}

	// ========== network_config ==========
	if d.HasChange("network_config") {
		if v, ok := d.GetOk("network_config"); ok {
			blocks := v.([]interface{})
			if len(blocks) > 0 && blocks[0] != nil {
				cfg := blocks[0].(map[string]interface{})
				vpcConfig := &aliyunFCAPI.VPCConfig{}
				hasVpc := false
				if vpcId, ok := cfg["vpc_id"].(string); ok && vpcId != "" {
					vpcConfig.VpcId = tea.String(vpcId)
					hasVpc = true
				}
				if sgId, ok := cfg["security_group_id"].(string); ok && sgId != "" {
					vpcConfig.SecurityGroupId = tea.String(sgId)
					hasVpc = true
				}
				if vsws, ok := cfg["vswitch_ids"].([]interface{}); ok {
					for _, vsw := range vsws {
						if id, ok := vsw.(string); ok && id != "" {
							vpcConfig.VSwitchIds = append(vpcConfig.VSwitchIds, tea.String(id))
						}
					}
					if len(vpcConfig.VSwitchIds) > 0 {
						hasVpc = true
					}
				}
				if hasVpc {
					function.VpcConfig = vpcConfig
				}
				if dnsConfigs, ok := cfg["dns_config"].([]interface{}); ok && len(dnsConfigs) > 0 && dnsConfigs[0] != nil {
					function.CustomDNS = expandFCCustomDNS(dnsConfigs[0].(map[string]interface{}))
				}
			}
		} else {
			// Block removed: send empty VPC config to detach from the VPC.
			function.VpcConfig = &aliyunFCAPI.VPCConfig{}
		}
	}

	// ========== storage_config ==========
	if d.HasChange("storage_config") {
		if v, ok := d.GetOk("storage_config"); ok {
			blocks := v.([]interface{})
			if len(blocks) > 0 && blocks[0] != nil {
				cfg := blocks[0].(map[string]interface{})
				if nasConfigs, ok := cfg["nas_config"].([]interface{}); ok && len(nasConfigs) > 0 && nasConfigs[0] != nil {
					function.NasConfig = expandFCNASConfig(nasConfigs[0].(map[string]interface{}))
				}
				if ossMounts, ok := cfg["oss_mount_config"].([]interface{}); ok && len(ossMounts) > 0 && ossMounts[0] != nil {
					function.OssMountConfig = expandFCOSSMountConfig(ossMounts[0].(map[string]interface{}))
				}
			}
		}
	}

	// ========== log_config ==========
	if d.HasChange("log_config") {
		if v, ok := d.GetOk("log_config"); ok {
			blocks := v.([]interface{})
			if len(blocks) > 0 && blocks[0] != nil {
				cfg := blocks[0].(map[string]interface{})
				logConfig := &aliyunFCAPI.LogConfig{}
				if project, ok := cfg["project"].(string); ok && project != "" {
					logConfig.Project = tea.String(project)
				}
				if logstore, ok := cfg["logstore"].(string); ok && logstore != "" {
					logConfig.Logstore = tea.String(logstore)
				}
				if v, ok := cfg["enable_instance_metrics"].(bool); ok {
					logConfig.EnableInstanceMetrics = tea.Bool(v)
				}
				if v, ok := cfg["enable_request_metrics"].(bool); ok {
					logConfig.EnableRequestMetrics = tea.Bool(v)
				}
				if rule, ok := cfg["log_begin_rule"].(string); ok && rule != "" {
					logConfig.LogBeginRule = tea.String(rule)
				}
				function.LogConfig = logConfig
			}
		}
	}

	// ========== container_config ==========
	if d.HasChange("container_config") {
		if v, ok := d.GetOk("container_config"); ok {
			blocks := v.([]interface{})
			if len(blocks) > 0 && blocks[0] != nil {
				function.CustomContainerConfig = expandFCCustomContainerConfig(blocks[0].(map[string]interface{}))
			}
		}
	}

	// ========== gpu_config ==========
	if d.HasChange("gpu_config") {
		if v, ok := d.GetOk("gpu_config"); ok {
			blocks := v.([]interface{})
			if len(blocks) > 0 && blocks[0] != nil {
				cfg := blocks[0].(map[string]interface{})
				gpuConfig := &aliyunFCAPI.GPUConfig{}
				if memory, ok := cfg["gpu_memory_size"].(int); ok && memory > 0 {
					gpuConfig.GpuMemorySize = tea.Int32(int32(memory))
				}
				if gpuType, ok := cfg["gpu_type"].(string); ok && gpuType != "" {
					gpuConfig.GpuType = tea.String(gpuType)
				}
				function.GpuConfig = gpuConfig
			}
		}
	}

	// ========== ram_config ==========
	if d.HasChange("ram_config") {
		if v, ok := d.GetOk("ram_config.0.role_arn"); ok {
			function.Role = tea.String(v.(string))
		} else {
			function.Role = tea.String("")
		}
	}

	return function
}

// SetSchemaFromFunction sets terraform schema data from Function, writing the
// FC 3.0-style nested blocks. code_config is intentionally left untouched:
// GetFunction does not return the code location (only CodeChecksum/CodeSize),
// so the configured OSS bucket/object must be kept in state as-is.
func (s *FCService) SetSchemaFromFunction(d *schema.ResourceData, function *aliyunFCAPI.Function) error {
	if function == nil {
		return fmt.Errorf("function cannot be nil")
	}

	if function.Description != nil {
		d.Set("description", *function.Description)
	}

	if function.Tags != nil {
		tags := make(map[string]string, len(function.Tags))
		for _, tag := range function.Tags {
			if tag != nil && tag.Key != nil {
				value := ""
				if tag.Value != nil {
					value = *tag.Value
				}
				tags[*tag.Key] = value
			}
		}
		d.Set("tags", tags)
	}

	// ========== entrypoint_config ==========
	if function.Handler != nil || len(function.Layers) > 0 {
		entrypoint := map[string]interface{}{}
		if function.Handler != nil {
			entrypoint["handler"] = *function.Handler
		}
		layers := make([]string, 0, len(function.Layers))
		for _, layer := range function.Layers {
			if layer != nil {
				layers = append(layers, *layer)
			}
		}
		entrypoint["layers"] = layers
		d.Set("entrypoint_config", []interface{}{entrypoint})
	}

	// ========== runtime_config ==========
	runtimeConfig := map[string]interface{}{}
	if function.Runtime != nil {
		runtimeConfig["runtime"] = *function.Runtime
	}
	if function.Timeout != nil {
		runtimeConfig["timeout"] = int(*function.Timeout)
	}
	if function.Environment != nil {
		envVars := make(map[string]string, len(function.Environment))
		for k, v := range function.Environment {
			if v != nil {
				envVars[k] = *v
			}
		}
		runtimeConfig["environment_variables"] = envVars
	}
	if function.InternetAccess != nil {
		runtimeConfig["internet_access"] = *function.InternetAccess
	}
	if function.Cpu != nil {
		runtimeConfig["cpu"] = float64(*function.Cpu)
	}
	if function.MemorySize != nil {
		runtimeConfig["memory_size"] = int(*function.MemorySize)
	}
	if function.DiskSize != nil {
		runtimeConfig["disk_size"] = int(*function.DiskSize)
	}
	if function.InstanceConcurrency != nil {
		runtimeConfig["instance_concurrency"] = int(*function.InstanceConcurrency)
	}
	if function.InstanceLifecycleConfig != nil && function.InstanceLifecycleConfig.Initializer != nil &&
		function.InstanceLifecycleConfig.Initializer.Timeout != nil {
		runtimeConfig["initialization_timeout"] = int(*function.InstanceLifecycleConfig.Initializer.Timeout)
	}
	if len(runtimeConfig) > 0 {
		d.Set("runtime_config", []interface{}{runtimeConfig})
	}

	// ========== network_config ==========
	if function.VpcConfig != nil || function.CustomDNS != nil {
		networkConfig := map[string]interface{}{}
		if function.VpcConfig != nil {
			if function.VpcConfig.VpcId != nil {
				networkConfig["vpc_id"] = *function.VpcConfig.VpcId
			}
			if function.VpcConfig.SecurityGroupId != nil {
				networkConfig["security_group_id"] = *function.VpcConfig.SecurityGroupId
			}
			vswitchIds := make([]string, 0, len(function.VpcConfig.VSwitchIds))
			for _, vsw := range function.VpcConfig.VSwitchIds {
				if vsw != nil {
					vswitchIds = append(vswitchIds, *vsw)
				}
			}
			networkConfig["vswitch_ids"] = vswitchIds
		}
		if function.CustomDNS != nil {
			dnsConfig := map[string]interface{}{}
			searches := make([]string, 0, len(function.CustomDNS.Searches))
			for _, item := range function.CustomDNS.Searches {
				if item != nil {
					searches = append(searches, *item)
				}
			}
			dnsConfig["searches"] = searches
			nameServers := make([]string, 0, len(function.CustomDNS.NameServers))
			for _, item := range function.CustomDNS.NameServers {
				if item != nil {
					nameServers = append(nameServers, *item)
				}
			}
			dnsConfig["name_servers"] = nameServers
			// dns_options is deliberately not written back: the schema block has
			// no Elem type, so round-tripping it would produce unstable state.
			networkConfig["dns_config"] = []interface{}{dnsConfig}
		}
		d.Set("network_config", []interface{}{networkConfig})
	}

	// ========== storage_config ==========
	if function.NasConfig != nil || function.OssMountConfig != nil {
		storageConfig := map[string]interface{}{}
		if function.NasConfig != nil {
			nasConfig := map[string]interface{}{}
			if function.NasConfig.UserId != nil {
				nasConfig["user_id"] = int(*function.NasConfig.UserId)
			}
			if function.NasConfig.GroupId != nil {
				nasConfig["group_id"] = int(*function.NasConfig.GroupId)
			}
			mountPoints := make([]interface{}, 0, len(function.NasConfig.MountPoints))
			for _, point := range function.NasConfig.MountPoints {
				if point == nil {
					continue
				}
				mountPoint := map[string]interface{}{}
				if point.ServerAddr != nil {
					mountPoint["server_addr"] = *point.ServerAddr
				}
				if point.MountDir != nil {
					mountPoint["mount_dir"] = *point.MountDir
				}
				if point.EnableTLS != nil {
					mountPoint["enable_tls"] = *point.EnableTLS
				}
				mountPoints = append(mountPoints, mountPoint)
			}
			nasConfig["mount_points"] = mountPoints
			storageConfig["nas_config"] = []interface{}{nasConfig}
		}
		if function.OssMountConfig != nil {
			mountPoints := make([]interface{}, 0, len(function.OssMountConfig.MountPoints))
			for _, point := range function.OssMountConfig.MountPoints {
				if point == nil {
					continue
				}
				mountPoint := map[string]interface{}{}
				if point.BucketName != nil {
					mountPoint["bucket_name"] = *point.BucketName
				}
				if point.BucketPath != nil {
					mountPoint["bucket_path"] = *point.BucketPath
				}
				if point.Endpoint != nil {
					mountPoint["endpoint"] = *point.Endpoint
				}
				if point.MountDir != nil {
					mountPoint["mount_dir"] = *point.MountDir
				}
				if point.ReadOnly != nil {
					mountPoint["read_only"] = *point.ReadOnly
				}
				mountPoints = append(mountPoints, mountPoint)
			}
			storageConfig["oss_mount_config"] = []interface{}{
				map[string]interface{}{"mount_points": mountPoints},
			}
		}
		d.Set("storage_config", []interface{}{storageConfig})
	}

	// ========== log_config ==========
	if function.LogConfig != nil {
		logConfig := map[string]interface{}{}
		if function.LogConfig.Project != nil {
			logConfig["project"] = *function.LogConfig.Project
		}
		if function.LogConfig.Logstore != nil {
			logConfig["logstore"] = *function.LogConfig.Logstore
		}
		if function.LogConfig.EnableInstanceMetrics != nil {
			logConfig["enable_instance_metrics"] = *function.LogConfig.EnableInstanceMetrics
		}
		if function.LogConfig.EnableRequestMetrics != nil {
			logConfig["enable_request_metrics"] = *function.LogConfig.EnableRequestMetrics
		}
		if function.LogConfig.LogBeginRule != nil {
			logConfig["log_begin_rule"] = *function.LogConfig.LogBeginRule
		}
		d.Set("log_config", []interface{}{logConfig})
	}

	// ========== lifecycle_config ==========
	// The initializer hook is only reflected here when it carries a handler;
	// a timeout-only initializer belongs to runtime_config.initialization_timeout.
	if function.InstanceLifecycleConfig != nil {
		lifecycleConfig := map[string]interface{}{}
		hasLifecycle := false
		if hook := function.InstanceLifecycleConfig.PreStop; hook != nil && hook.Handler != nil {
			preStop := map[string]interface{}{"handler": *hook.Handler}
			if hook.Timeout != nil {
				preStop["timeout"] = int(*hook.Timeout)
			}
			lifecycleConfig["pre_stop"] = []interface{}{preStop}
			hasLifecycle = true
		}
		if hook := function.InstanceLifecycleConfig.Initializer; hook != nil && hook.Handler != nil && *hook.Handler != "" {
			initializer := map[string]interface{}{"handler": *hook.Handler}
			if hook.Timeout != nil {
				initializer["timeout"] = int(*hook.Timeout)
			}
			lifecycleConfig["initializer"] = []interface{}{initializer}
			hasLifecycle = true
		}
		if hasLifecycle {
			d.Set("lifecycle_config", []interface{}{lifecycleConfig})
		}
	}

	// ========== container_config ==========
	if function.CustomContainerConfig != nil {
		containerConfig := map[string]interface{}{}
		cfg := function.CustomContainerConfig
		if cfg.Image != nil {
			containerConfig["image"] = *cfg.Image
		}
		if cfg.ResolvedImageUri != nil {
			containerConfig["resolved_image_uri"] = *cfg.ResolvedImageUri
		}
		entrypoint := make([]string, 0, len(cfg.Entrypoint))
		for _, item := range cfg.Entrypoint {
			if item != nil {
				entrypoint = append(entrypoint, *item)
			}
		}
		containerConfig["entrypoint"] = entrypoint
		command := make([]string, 0, len(cfg.Command))
		for _, item := range cfg.Command {
			if item != nil {
				command = append(command, *item)
			}
		}
		containerConfig["command"] = command
		if cfg.Port != nil {
			containerConfig["port"] = int(*cfg.Port)
		}
		if cfg.AcrInstanceId != nil {
			containerConfig["acr_instance_id"] = *cfg.AcrInstanceId
		}
		if cfg.AccelerationType != nil {
			containerConfig["acceleration_type"] = *cfg.AccelerationType
		}
		if cfg.HealthCheckConfig != nil {
			healthCheck := map[string]interface{}{}
			hc := cfg.HealthCheckConfig
			if hc.HttpGetUrl != nil {
				healthCheck["http_get_url"] = *hc.HttpGetUrl
			}
			if hc.InitialDelaySeconds != nil {
				healthCheck["initial_delay_seconds"] = int(*hc.InitialDelaySeconds)
			}
			if hc.TimeoutSeconds != nil {
				healthCheck["timeout_seconds"] = int(*hc.TimeoutSeconds)
			}
			if hc.PeriodSeconds != nil {
				healthCheck["period_seconds"] = int(*hc.PeriodSeconds)
			}
			if hc.FailureThreshold != nil {
				healthCheck["failure_threshold"] = int(*hc.FailureThreshold)
			}
			if hc.SuccessThreshold != nil {
				healthCheck["success_threshold"] = int(*hc.SuccessThreshold)
			}
			containerConfig["health_check_config"] = []interface{}{healthCheck}
		}
		d.Set("container_config", []interface{}{containerConfig})
	}

	// ========== gpu_config ==========
	if function.GpuConfig != nil {
		gpuConfig := map[string]interface{}{}
		if function.GpuConfig.GpuMemorySize != nil {
			gpuConfig["gpu_memory_size"] = int(*function.GpuConfig.GpuMemorySize)
		}
		if function.GpuConfig.GpuType != nil {
			gpuConfig["gpu_type"] = *function.GpuConfig.GpuType
		}
		d.Set("gpu_config", []interface{}{gpuConfig})
	}

	// ========== ram_config ==========
	if function.Role != nil {
		d.Set("ram_config", []interface{}{map[string]interface{}{"role_arn": *function.Role}})
	}

	// ========== tracing_config (Computed) ==========
	if function.TracingConfig != nil {
		tracingConfig := map[string]interface{}{}
		if function.TracingConfig.Type != nil {
			tracingConfig["type"] = *function.TracingConfig.Type
		}
		if function.TracingConfig.Params != nil {
			params := make(map[string]string, len(function.TracingConfig.Params))
			for k, v := range function.TracingConfig.Params {
				if v != nil {
					params[k] = *v
				}
			}
			tracingConfig["params"] = params
		}
		d.Set("tracing_config", []interface{}{tracingConfig})
	}

	return nil
}

// FCFunctionStateRefreshFunc returns a StateRefreshFunc for FC function operations
func (s *FCService) FCFunctionStateRefreshFunc(functionName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeFCFunction(functionName)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		currentStatus := "Active" // FC functions don't have explicit status, assume Active if retrievable
		if object.State != nil {
			currentStatus = *object.State
		}

		for _, failState := range failStates {
			if currentStatus == failState {
				return object, currentStatus, WrapError(Error(FailedToReachTargetStatus, currentStatus))
			}
		}
		return object, currentStatus, nil
	}
}

// WaitForFCFunctionCreating waits for function creation to complete
func (s *FCService) WaitForFCFunctionCreating(functionName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Creating", "Pending"},
		[]string{"Active"},
		timeout,
		5*time.Second,
		s.FCFunctionStateRefreshFunc(functionName, []string{"Failed"}),
	)

	_, err := stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, functionName)
	}
	return nil
}

// WaitForFCFunctionUpdating waits for function update to complete
func (s *FCService) WaitForFCFunctionUpdating(functionName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Updating"},
		[]string{"Active"},
		timeout,
		5*time.Second,
		s.FCFunctionStateRefreshFunc(functionName, []string{"Failed"}),
	)

	_, err := stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, functionName)
	}
	return nil
}

// WaitForFCFunctionDeleting waits for function deletion to complete
func (s *FCService) WaitForFCFunctionDeleting(functionName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"Deleting"},
		Target:  []string{""},
		Refresh: func() (interface{}, string, error) {
			obj, err := s.DescribeFCFunction(functionName)
			if err != nil {
				if NotFoundError(err) {
					return nil, "", nil
				}
				return nil, "", WrapError(err)
			}
			return obj, "Deleting", nil
		},
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, functionName)
	}
	return nil
}

// EncodeFunctionVersionId encodes function name and version into a resource ID string
// Format: functionName:versionId
func EncodeFunctionVersionId(functionName, versionId string) string {
	return fmt.Sprintf("%s:%s", functionName, versionId)
}

// DecodeFunctionVersionId decodes function version ID string to function name and version
func DecodeFunctionVersionId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid function version ID format, expected functionName:versionId, got %s", id)
	}
	return parts[0], parts[1], nil
}

// DescribeFCFunctionVersion retrieves function version information by ID
func (s *FCService) DescribeFCFunctionVersion(id string) (*aliyunFCAPI.Function, error) {
	functionName, versionId, err := DecodeFunctionVersionId(id)
	if err != nil {
		return nil, err
	}

	// Use GetFunction with version qualifier to get version-specific information
	return s.GetAPI().GetFunction(functionName, &versionId)
}

// FunctionStateRefreshFunc returns a StateRefreshFunc to wait for function status changes
func (s *FCService) FunctionStateRefreshFunc(functionName string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		object, err := s.DescribeFCFunction(functionName)
		if err != nil {
			if NotFoundError(err) {
				return nil, "", nil
			}
			return nil, "", WrapError(err)
		}

		currentState := "Active" // FC v3 functions are typically Active when created
		if object.State != nil && *object.State != "" {
			currentState = *object.State
		}

		for _, failState := range failStates {
			if currentState == failState {
				return object, currentState, WrapError(Error(FailedToReachTargetStatus, currentState))
			}
		}
		return object, currentState, nil
	}
}

// WaitForFunctionCreating waits for function creation to complete
func (s *FCService) WaitForFunctionCreating(functionName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Creating", "Pending"},
		[]string{"Active"},
		timeout,
		5*time.Second,
		s.FunctionStateRefreshFunc(functionName, []string{"Failed", "Error"}),
	)

	_, err := stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, functionName)
	}
	return nil
}

// WaitForFunctionDeleting waits for function deletion to complete
func (s *FCService) WaitForFunctionDeleting(functionName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Deleting", "Active"},
		[]string{""},
		timeout,
		5*time.Second,
		s.FunctionStateRefreshFunc(functionName, []string{"Failed", "Error"}),
	)

	_, err := stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, functionName)
	}
	return nil
}

// WaitForFunctionUpdating waits for function update to complete
func (s *FCService) WaitForFunctionUpdating(functionName string, timeout time.Duration) error {
	stateConf := BuildStateConf(
		[]string{"Updating", "Pending"},
		[]string{"Active"},
		timeout,
		5*time.Second,
		s.FunctionStateRefreshFunc(functionName, []string{"Failed", "Error"}),
	)

	_, err := stateConf.WaitForState()
	if err != nil {
		return WrapErrorf(err, IdMsg, functionName)
	}
	return nil
}
