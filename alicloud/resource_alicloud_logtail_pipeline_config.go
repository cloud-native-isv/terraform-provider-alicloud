package alicloud

import (
	"fmt"
	"regexp"
	"time"

	"github.com/aliyun/terraform-provider-alicloud/alicloud/connectivity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAliCloudLogtailPipelineConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliCloudLogtailPipelineConfigCreate,
		Read:   resourceAliCloudLogtailPipelineConfigRead,
		Update: resourceAliCloudLogtailPipelineConfigUpdate,
		Delete: resourceAliCloudLogtailPipelineConfigDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAliCloudLogtailPipelineConfigImport,
		},

		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^[a-z0-9-_]+$`),
					"name must match ^[a-z0-9-_]+$",
				),
			},
			"inputs":      pluginSchema("inputs", true),
			"processors":  pluginSchema("processors", false),
			"flushers":    pluginSchema("flushers", true),
			"aggregators": pluginSchema("aggregators", false),

			"global_json": {
				Type:     schema.TypeString,
				Optional: true,
				StateFunc: func(v interface{}) string {
					s, _ := NormalizeLogtailConfigJson(v.(string))
					return s
				},
				ValidateFunc: validateLogtailConfigJsonObjectString,
			},
			"task_json": {
				Type:     schema.TypeString,
				Optional: true,
				StateFunc: func(v interface{}) string {
					s, _ := NormalizeLogtailConfigJson(v.(string))
					return s
				},
				ValidateFunc: validateLogtailConfigJsonObjectString,
			},
			"log_sample": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"create_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"last_modify_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func pluginSchema(name string, required bool) *schema.Schema {
	s := &schema.Schema{
		Type:     schema.TypeList,
		Optional: !required,
		Required: required,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": {
					Type:     schema.TypeString,
					Required: true,
				},
				"config_json": {
					Type:     schema.TypeString,
					Optional: true,
					StateFunc: func(v interface{}) string {
						s, _ := NormalizeLogtailConfigJson(v.(string))
						return s
					},
					ValidateFunc: validation.StringIsJSON,
				},
			},
		},
	}
	if required {
		s.MinItems = 1
	}
	return s
}

func resourceAliCloudLogtailPipelineConfigCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_pipeline_config", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	domainConfig, err := expandSlsLogtailPipelineConfig(d)
	if err != nil {
		return WrapError(err)
	}

	libConfig, err := domainConfig.ToLibConfig()
	if err != nil {
		return WrapError(err)
	}

	projectName := domainConfig.Project
	configName := domainConfig.Name

	// Retry logic for creation
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		err := slsService.CreateSlsLogtailPipelineConfig(projectName, libConfig)
		if err != nil {
			if IsExpectedErrors(err, []string{"ConfigAlreadyExist"}) {
				return resource.NonRetryableError(fmt.Errorf("Logtail pipeline config %s already exists in project %s", configName, projectName))
			}
			if IsExpectedErrors(err, []string{"InternalServerError", LogClientTimeout}) {
				time.Sleep(5 * time.Second)
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_pipeline_config", "CreateLogtailPipelineConfig", AlibabaCloudSdkGoERROR)
	}

	resourceId := fmt.Sprintf("%s:%s:%s", projectName, "config", configName)
	d.SetId(resourceId)

	// Wait for state
	stateConf := BuildStateConf(
		[]string{""},       // Pending (doesn't apply strictly here)
		[]string{"active"}, // Target
		d.Timeout(schema.TimeoutCreate),
		5*time.Second,
		slsService.LogtailPipelineConfigStateRefreshFunc(resourceId, []string{}),
	)
	if _, err := stateConf.WaitForState(); err != nil {
		return WrapErrorf(err, IdMsg, resourceId)
	}

	return resourceAliCloudLogtailPipelineConfigRead(d, meta)
}

func resourceAliCloudLogtailPipelineConfigRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_pipeline_config", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	// 1. Get from API
	libConfig, err := slsService.DescribeSlsLogtailPipelineConfig(d.Id())
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	// 2. Convert to Domain
	// Need project name from ID because Describe returns config obj which might not have Project field populated (depends on SDK response)
	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}
	projectName := parts[0]

	domainConfig := FromLibConfig(libConfig, projectName)

	// 3. Flatten to State
	if err := flattenSlsLogtailPipelineConfig(d, domainConfig); err != nil {
		return WrapError(err)
	}

	return nil
}

func resourceAliCloudLogtailPipelineConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_pipeline_config", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	domainConfig, err := expandSlsLogtailPipelineConfig(d)
	if err != nil {
		return WrapError(err)
	}

	libConfig, err := domainConfig.ToLibConfig()
	if err != nil {
		return WrapError(err)
	}

	projectName := domainConfig.Project

	// Update
	if err := slsService.UpdateSlsLogtailPipelineConfig(projectName, libConfig); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_pipeline_config", "UpdateLogtailPipelineConfig", AlibabaCloudSdkGoERROR)
	}

	// Wait for update (read back consistency)
	// Usually strict consistency is not guaranteed immediately, but Refresh should eventually succeed with new values.
	// For update, we rely on Read to sync state.

	return resourceAliCloudLogtailPipelineConfigRead(d, meta)
}

func resourceAliCloudLogtailPipelineConfigDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_pipeline_config", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	parts, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return WrapError(err)
	}
	projectName := parts[0]
	configName := parts[2]

	if err := slsService.DeleteSlsLogtailPipelineConfig(projectName, configName); err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_pipeline_config", "DeleteLogtailPipelineConfig", AlibabaCloudSdkGoERROR)
	}

	return nil
}

func resourceAliCloudLogtailPipelineConfigImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// ID format: project:config:name
	_, err := ParseResourceId(d.Id(), 3)
	if err != nil {
		return nil, WrapError(err)
	}
	// Verify it exists
	client := meta.(*connectivity.AliyunClient)
	slsService, err := NewSlsService(client)
	if err != nil {
		return nil, WrapErrorf(err, DefaultErrorMsg, "alicloud_logtail_pipeline_config", "NewSlsService", AlibabaCloudSdkGoERROR)
	}

	_, err = slsService.DescribeSlsLogtailPipelineConfig(d.Id())
	if err != nil {
		if NotFoundError(err) {
			return nil, fmt.Errorf("Logtail Pipeline Config not found with ID %s", d.Id())
		}
		return nil, WrapError(err)
	}

	return []*schema.ResourceData{d}, nil
}

// Helpers

func expandSlsLogtailPipelineConfig(d *schema.ResourceData) (*SlsLogtailPipelineConfig, error) {
	c := &SlsLogtailPipelineConfig{
		Project:    d.Get("project").(string),
		Name:       d.Get("name").(string),
		LogSample:  d.Get("log_sample").(string),
		GlobalJson: d.Get("global_json").(string),
		TaskJson:   d.Get("task_json").(string),
	}

	var err error
	c.Inputs, err = expandPlugins(d.Get("inputs").([]interface{}))
	if err != nil {
		return nil, err
	}

	c.Processors, err = expandPlugins(d.Get("processors").([]interface{}))
	if err != nil {
		return nil, err
	}

	c.Flushers, err = expandPlugins(d.Get("flushers").([]interface{}))
	if err != nil {
		return nil, err
	}

	c.Aggregators, err = expandPlugins(d.Get("aggregators").([]interface{}))
	if err != nil {
		return nil, err
	}

	return c, nil
}

func expandPlugins(list []interface{}) ([]SlsLogtailPipelineConfigPlugin, error) {
	if len(list) == 0 {
		return nil, nil
	}
	plugins := make([]SlsLogtailPipelineConfigPlugin, 0, len(list))
	for _, item := range list {
		if item == nil {
			continue
		}
		m := item.(map[string]interface{})

		p := SlsLogtailPipelineConfigPlugin{
			Type: m["type"].(string),
		}
		if v, ok := m["config_json"]; ok {
			p.ConfigJson = v.(string)
		}
		plugins = append(plugins, p)
	}
	return plugins, nil
}

func flattenSlsLogtailPipelineConfig(d *schema.ResourceData, c *SlsLogtailPipelineConfig) error {
	d.Set("project", c.Project)
	d.Set("name", c.Name)
	d.Set("log_sample", c.LogSample)
	d.Set("create_time", c.CreateTime)
	d.Set("last_modify_time", c.LastModifyTime)

	// JSON fields - should be normalized before setting to avoid unnecessary diffs if API returns different format?
	// But api returns map, we converted to json string in FromLibConfig using helper (which normalizes).
	// So c.GlobalJson etc are already normalized.
	d.Set("global_json", c.GlobalJson)
	d.Set("task_json", c.TaskJson)

	if err := d.Set("inputs", flattenPlugins(c.Inputs)); err != nil {
		return err
	}
	if err := d.Set("processors", flattenPlugins(c.Processors)); err != nil {
		return err
	}
	if err := d.Set("flushers", flattenPlugins(c.Flushers)); err != nil {
		return err
	}
	if err := d.Set("aggregators", flattenPlugins(c.Aggregators)); err != nil {
		return err
	}

	return nil
}

func flattenPlugins(plugins []SlsLogtailPipelineConfigPlugin) []interface{} {
	if len(plugins) == 0 {
		return nil
	}
	list := make([]interface{}, 0, len(plugins))
	for _, p := range plugins {
		m := make(map[string]interface{})
		m["type"] = p.Type
		m["config_json"] = p.ConfigJson
		list = append(list, m)
	}
	return list
}
